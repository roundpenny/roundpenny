// Copyright (c) 2026 RoundPenny. All rights reserved.

package secrets

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
)

type VaultManager struct {
	client *vault.Client
	path   string
}

func NewVaultManager(ctx context.Context, addr, token, path string) (*VaultManager, error) {
	config := vault.DefaultConfig()
	config.Address = addr
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	client.SetToken(token)
	return &VaultManager{client: client, path: path}, nil
}

func NewVaultManagerFromEnv(ctx context.Context) (*VaultManager, error) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	path := os.Getenv("VAULT_SECRET_PATH")
	if path == "" {
		path = "secret"
	}
	return NewVaultManager(ctx, addr, token, path)
}

func (m *VaultManager) Get(ctx context.Context, key string) (string, error) {
	secret, err := m.client.KVv2(m.path).Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("vault get %s: %w", key, err)
	}
	if secret == nil || secret.Data == nil {
		return "", ErrNotFound
	}
	val, ok := secret.Data["value"].(string)
	if !ok {
		return "", ErrNotFound
	}
	return val, nil
}

func (m *VaultManager) Close() error { return nil }
