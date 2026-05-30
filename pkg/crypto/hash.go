// Copyright (c) 2026 RoundPenny. All rights reserved.

package crypto

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func SHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate hash: %w", err)
	}
	return string(bytes), nil
}

func ComparePassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}
	return nil
}
