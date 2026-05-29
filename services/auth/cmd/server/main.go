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
	"github.com/roundup-platform/services/auth/internal/handler"
	"github.com/roundup-platform/services/auth/internal/middleware"
	"github.com/roundup-platform/services/auth/internal/repository"
	"github.com/roundup-platform/services/auth/internal/service"
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

	tp, err := tracing.InitTracing("auth")
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

	metrics := monitoring.New("auth")
	repo := repository.NewAuthRepository(pool)
	svc := service.NewAuthService(repo, jwtSecret)
	h := handler.NewAuthHandler(svc)

	loginLimit := 10
	if v := os.Getenv("RATE_LIMIT_LOGIN"); v != "" {
		if n, _ := fmt.Sscanf(v, "%d", &loginLimit); n != 1 {
			slog.Warn("invalid RATE_LIMIT_LOGIN, using default", "value", v)
		}
	}
	registerLimit := 5
	if v := os.Getenv("RATE_LIMIT_REGISTER"); v != "" {
		if n, _ := fmt.Sscanf(v, "%d", &registerLimit); n != 1 {
			slog.Warn("invalid RATE_LIMIT_REGISTER, using default", "value", v)
		}
	}
	slog.Info("rate limits", "login", loginLimit, "register", registerLimit, "login_env", os.Getenv("RATE_LIMIT_LOGIN"), "register_env", os.Getenv("RATE_LIMIT_REGISTER"))
	loginLimiter := middleware.NewRateLimiter(loginLimit, time.Minute)
	registerLimiter := middleware.NewRateLimiter(registerLimit, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("POST /v1/auth/register", registerLimiter.Middleware(http.HandlerFunc(h.Register)))
	mux.Handle("POST /v1/auth/login", loginLimiter.Middleware(http.HandlerFunc(h.Login)))
	mux.HandleFunc("POST /v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /v1/auth/me", h.Me)
	mux.HandleFunc("POST /v1/auth/oauth/{provider}", h.OAuth)
	mux.HandleFunc("POST /v1/auth/mfa/verify", h.VerifyMFA)
	mux.HandleFunc("POST /v1/auth/mfa/setup", h.SetupMFA)
	mux.HandleFunc("POST /v1/auth/mfa/enable", h.EnableMFA)
	mux.HandleFunc("POST /v1/auth/mfa/disable", h.DisableMFA)
	mux.HandleFunc("POST /v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /v1/auth/verify-email/confirm", h.ConfirmEmailVerification)
	mux.HandleFunc("POST /v1/auth/kyc", h.SubmitKYC)
	mux.HandleFunc("GET /v1/auth/kyc/status", h.GetKYCStatus)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "error", err)
		os.Exit(1)
	}
}
