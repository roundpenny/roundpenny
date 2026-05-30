// Copyright (c) 2026 RoundPenny. All rights reserved.

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/roundup-platform/pkg/cache"
)

type RateLimiter struct {
	redis  *cache.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration, redisClient *cache.Client) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}

		key := "ratelimit:" + r.URL.Path + ":" + ip

		if !rl.allow(r.Context(), key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "too many requests"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ctx context.Context, key string) bool {
	windowSec := int64(rl.window.Seconds())

	count, err := rl.redis.Incr(ctx, key)
	if err != nil {
		return true
	}

	if count == 1 {
		rl.redis.Expire(ctx, key, rl.window)
	}

	if count > int64(rl.limit) {
		remaining := time.Duration(windowSec) * time.Second
		if err == redis.Nil {
			rl.redis.Expire(ctx, key, remaining)
		}
		return false
	}

	return true
}
