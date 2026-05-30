// Copyright (c) 2026 RoundPenny. All rights reserved.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/roundup-platform/pkg/config"
	"github.com/roundup-platform/pkg/cors"
	"github.com/roundup-platform/pkg/db"
	"github.com/roundup-platform/pkg/monitoring"
	"github.com/roundup-platform/pkg/tls"
	"github.com/roundup-platform/pkg/tracing"
	"github.com/roundup-platform/services/admin/internal/handler"
	"github.com/roundup-platform/services/admin/internal/repository"
	"github.com/roundup-platform/services/admin/internal/service"
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

	tp, err := tracing.InitTracing("admin")
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

	jwtSecret, err := config.LoadJWTSecret()
	if err != nil {
		slog.Error("jwt secret", "error", err)
		os.Exit(1)
	}

	metrics := monitoring.New("admin")
	repo := repository.NewAdminRepository(pool)
	svc := service.NewAdminService(repo, []byte(jwtSecret))
	h := handler.NewAdminHandler(svc)

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth
	mux.HandleFunc("POST /v1/admin/login", h.Login)
	mux.HandleFunc("POST /v1/admin/logout", h.Logout)

	// Dashboard stats
	mux.HandleFunc("GET /v1/admin/stats", h.AuthMiddleware(h.GetStats))
	mux.HandleFunc("GET /v1/admin/users", h.AuthMiddleware(h.ListUsers))
	mux.HandleFunc("GET /v1/admin/users/{id}", h.AuthMiddleware(h.GetUser))
	mux.HandleFunc("PUT /v1/admin/users/{id}/status", h.AuthMiddleware(h.UpdateUserStatus))
	mux.HandleFunc("GET /v1/admin/merchants", h.AuthMiddleware(h.ListMerchants))
	mux.HandleFunc("GET /v1/admin/merchants/{id}", h.AuthMiddleware(h.GetMerchant))
	mux.HandleFunc("GET /v1/admin/transactions", h.AuthMiddleware(h.ListTransactions))
	mux.HandleFunc("GET /v1/admin/transactions/{id}", h.AuthMiddleware(h.GetTransaction))
	mux.HandleFunc("GET /v1/admin/fraud-alerts", h.AuthMiddleware(h.ListFraudAlerts))
	mux.HandleFunc("POST /v1/admin/fraud-alerts/{id}/review", h.AuthMiddleware(h.ReviewFraudAlert))
	mux.HandleFunc("GET /v1/admin/kyc-submissions", h.AuthMiddleware(h.ListKYCSubmissions))
	mux.HandleFunc("POST /v1/admin/kyc-submissions/{id}/review", h.AuthMiddleware(h.ReviewKYCSubmission))

	// Serve dashboard UI
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}
	staticDir := filepath.Join(webDir, "static")
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /admin/", http.StripPrefix("/admin/", fs))
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/admin/", http.StatusFound)
				return
			}
			http.NotFound(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
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
