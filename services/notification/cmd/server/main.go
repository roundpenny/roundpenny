// Copyright (c) 2026 RoundPenny. All rights reserved.

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
	"github.com/roundup-platform/pkg/tls"
	"github.com/roundup-platform/pkg/tracing"
	"github.com/roundup-platform/pkg/email"
	"github.com/roundup-platform/services/notification/internal/consumer"
	"github.com/roundup-platform/services/notification/internal/handler"
	"github.com/roundup-platform/services/notification/internal/repository"
	"github.com/roundup-platform/services/notification/internal/service"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx)
	if err != nil {
		slog.Error("database connection", "error", err)
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

	tp, err := tracing.InitTracing("notification")
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

	metrics := monitoring.New("notification")

	webhookRepo := repository.NewWebhookRepository(pool)
	webhookSvc := service.NewWebhookService(webhookRepo)
	webhookH := handler.NewWebhookHandler(webhookSvc)

	emailRepo := repository.NewEmailRepository(pool.Pool)
	emailClient := email.NewClient()
	emailSvc := service.NewEmailService(emailRepo, emailClient)
	emailH := handler.NewEmailHandler(emailSvc)

	webhookConsumer := consumer.NewWebhookConsumer(webhookSvc)
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}
	if err := webhookConsumer.Start(ctx, brokers, "notification-service"); err != nil {
		slog.Warn("kafka consumer warning", "error", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /v1/webhooks", webhookH.CreateWebhook)
	mux.HandleFunc("GET /v1/webhooks/{id}", webhookH.GetWebhook)
	mux.HandleFunc("GET /v1/webhooks", webhookH.ListUserWebhooks)
	mux.HandleFunc("PUT /v1/webhooks/{id}", webhookH.UpdateWebhook)
	mux.HandleFunc("DELETE /v1/webhooks/{id}", webhookH.DeleteWebhook)

	mux.HandleFunc("POST /v1/emails/send", emailH.SendEmail)
	mux.HandleFunc("GET /v1/emails", emailH.ListEmails)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      cors.Middleware(monitoring.MetricsMiddleware(metrics, mux)),
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

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "error", err)
		os.Exit(1)
	}
}
