// Copyright (c) 2026 RoundPenny. All rights reserved.

package secrets

import (
	"context"
	"errors"
	"os"
)

var ErrNotFound = errors.New("secret not found")

type Manager interface {
	Get(ctx context.Context, key string) (string, error)
	Close() error
}

type EnvManager struct{}

func NewEnvManager() *EnvManager {
	return &EnvManager{}
}

func (m *EnvManager) Get(ctx context.Context, key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", ErrNotFound
	}
	return val, nil
}

func (m *EnvManager) Close() error { return nil }
