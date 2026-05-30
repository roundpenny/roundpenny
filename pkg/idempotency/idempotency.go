// Copyright (c) 2026 RoundPenny. All rights reserved.

package idempotency

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyMissing    = errors.New("idempotency key is required")
	ErrKeyReplayed   = errors.New("idempotency key already used with different request")
	ErrConflict      = errors.New("idempotency key already exists")
	DefaultTTL       = 24 * time.Hour
	HeaderKey        = "Idempotency-Key"
)

type StoredResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

type Client struct {
	rdb  *redis.Client
	mock bool
	data map[string]*StoredResponse
}

func NewClient() *Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c := &Client{rdb: rdb}

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis not available, using mock idempotency storage", "error", err)
		c.mock = true
		c.data = make(map[string]*StoredResponse)
	}

	return c
}

func (c *Client) Get(ctx context.Context, key string) (*StoredResponse, error) {
	if c.mock {
		resp, ok := c.data[key]
		if !ok {
			return nil, redis.Nil
		}
		return resp, nil
	}
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var sr StoredResponse
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

func (c *Client) Set(ctx context.Context, key string, resp *StoredResponse) error {
	if c.mock {
		c.data[key] = resp
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, DefaultTTL).Err()
}

func (c *Client) TryLock(ctx context.Context, key string) (bool, error) {
	if c.mock {
		if _, ok := c.data[key]; ok {
			return false, nil
		}
		c.data[key] = nil
		return true, nil
	}
	ok, err := c.rdb.SetNX(ctx, key, "locked", DefaultTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if c.mock {
		delete(c.data, key)
		return nil
	}
	return c.rdb.Del(ctx, key).Err()
}

func GenerateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate idempotency key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

func Middleware(client *Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get(HeaderKey)
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

			stored, err := client.Get(r.Context(), key)
		if err == nil && stored != nil {
			for k, v := range stored.Headers {
				w.Header()[k] = []string{v}
			}
			w.WriteHeader(stored.StatusCode)
			w.Write(stored.Body)
			return
		}

		locked, err := client.TryLock(r.Context(), key)
		if err != nil {
			slog.Error("idempotency lock failed", "key", key, "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !locked {
			http.Error(w, "conflict: idempotency key already in progress", http.StatusConflict)
			return
		}

		rw := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		if rw.statusCode >= 200 && rw.statusCode < 500 {
			sr := &StoredResponse{
				StatusCode: rw.statusCode,
				Headers:    make(map[string]string),
				Body:       rw.body,
			}
			for k := range w.Header() {
				sr.Headers[k] = w.Header().Get(k)
			}

			if err := client.Set(r.Context(), key, sr); err != nil {
				slog.Error("idempotency store failed", "key", key, "error", err)
			}
		}
	})
}
