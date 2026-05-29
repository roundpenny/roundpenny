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
	"github.com/roundup-platform/services/roundup-engine/internal/engine"
	"github.com/roundup-platform/services/roundup-engine/internal/repository"
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

	tp, err := tracing.InitTracing("roundup-engine")
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

	consumer, err := kafka.NewConsumer("roundup-engine", []string{"tx.settled"})
	if err != nil {
		slog.Error("consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	repo := repository.NewRoundUpRepository(pool)
	eng := engine.NewRoundUpEngine(repo, producer)

	metrics := monitoring.New("roundup_engine")

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9097"
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

	go func() {
		slog.Info("round-up engine started, consuming from tx.settled")
		if err := consumer.Consume(context.Background(), eng.HandleTransaction); err != nil {
			slog.Error("consume", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down round-up engine")
}
