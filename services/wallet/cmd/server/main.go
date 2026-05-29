package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roundup-platform/pkg/config"
	"github.com/roundup-platform/pkg/cors"
	"github.com/roundup-platform/pkg/db"
	"github.com/roundup-platform/pkg/monitoring"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/pkg/tls"
	"github.com/roundup-platform/pkg/tracing"
	"github.com/roundup-platform/services/wallet/internal/consumer"
	"github.com/roundup-platform/services/wallet/internal/handler"
	"github.com/roundup-platform/services/wallet/internal/repository"
	"github.com/roundup-platform/services/wallet/internal/service"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrations(ctx, pool.Pool, migrationsDir); err != nil {
		slog.Warn("migrations warning", "error", err)
	}

	tp, err := tracing.InitTracing("wallet")
	if err != nil {
		slog.Warn("tracing", "error", err)
	}
	defer func() {
		if tp != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tp.Shutdown(ctx)
		}
	}()

	_, err = config.LoadJWTSecret()
	if err != nil {
		slog.Error("jwt secret", "error", err)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer()
	if err != nil {
		slog.Error("producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	repo := repository.NewWalletRepository(pool)
	svc := service.NewWalletService(repo, producer)
	metrics := monitoring.New("wallet")
	h := handler.NewWalletHandler(svc)

	cns, err := kafka.NewConsumer("wallet-service", []string{"roundup.calculated"})
	if err != nil {
		slog.Error("consumer", "error", err)
		os.Exit(1)
	}
	defer cns.Close()

	rd := consumer.NewRoundUpConsumer(svc)
	go func() {
		slog.Info("wallet consumer started")
		if err := cns.Consume(context.Background(), rd.HandleRoundUp); err != nil {
			slog.Error("consume", "error", err)
			os.Exit(1)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /v1/wallets/me", h.GetWallet)
	mux.HandleFunc("GET /v1/wallets/me/transactions", h.GetTransactions)
	mux.HandleFunc("POST /v1/wallets/me/withdraw", h.Withdraw)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      cors.Middleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if certFile != "" && keyFile != "" {
		tlsCfg, err := tls.LoadTLSCert(certFile, keyFile)
		if err != nil {
			slog.Error("tls config", "error", err)
			os.Exit(1)
		}
		srv.TLSConfig = tlsCfg
	}

	go func() {
		if srv.TLSConfig != nil {
			slog.Info("listening with TLS", "port", port)
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("server", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("listening", "port", port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server", "error", err)
				os.Exit(1)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
