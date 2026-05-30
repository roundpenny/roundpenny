// Copyright (c) 2026 RoundPenny. All rights reserved.

package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"
)

func generateTestCert(t *testing.T) (certFile, keyFile string, cleanup func()) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.roundpenny.local"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		IsCA:         true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	certFile = "test_cert.pem"
	keyFile = "test_key.pem"

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		t.Fatalf("write cert: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(keyFile, keyPEM, 0644); err != nil {
		os.Remove(certFile)
		t.Fatalf("write key: %v", err)
	}

	cleanup = func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}
	return
}

func TestLoadTLSCert(t *testing.T) {
	certFile, keyFile, cleanup := generateTestCert(t)
	defer cleanup()

	tlsCfg, err := LoadTLSCert(certFile, keyFile)
	if err != nil {
		t.Fatalf("LoadTLSCert failed: %v", err)
	}

	if len(tlsCfg.Certificates) != 1 {
		t.Fatalf("got %d certs, want 1", len(tlsCfg.Certificates))
	}

	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("got %d, want %d", tlsCfg.MinVersion, tls.VersionTLS12)
	}
}

func TestLoadTLSCert_missing_cert(t *testing.T) {
	_, err := LoadTLSCert("nonexistent.pem", "nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for missing cert")
	}
}

func TestLoadTLSCert_missing_key(t *testing.T) {
	certFile, _, cleanup := generateTestCert(t)
	defer cleanup()

	_, err := LoadTLSCert(certFile, "nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestTLSEnabled_both_set(t *testing.T) {
	os.Setenv("TLS_CERT_FILE", "/path/to/cert")
	os.Setenv("TLS_KEY_FILE", "/path/to/key")
	defer func() {
		os.Unsetenv("TLS_CERT_FILE")
		os.Unsetenv("TLS_KEY_FILE")
	}()

	if !TLSEnabled() {
		t.Fatal("expected TLS enabled")
	}
}

func TestTLSEnabled_missing_cert(t *testing.T) {
	os.Unsetenv("TLS_CERT_FILE")
	os.Unsetenv("TLS_KEY_FILE")
	os.Setenv("TLS_KEY_FILE", "/path/to/key")
	defer os.Unsetenv("TLS_KEY_FILE")

	if TLSEnabled() {
		t.Fatal("expected TLS disabled without cert")
	}
}

func TestTLSEnabled_missing_key(t *testing.T) {
	os.Unsetenv("TLS_CERT_FILE")
	os.Unsetenv("TLS_KEY_FILE")
	os.Setenv("TLS_CERT_FILE", "/path/to/cert")
	defer os.Unsetenv("TLS_CERT_FILE")

	if TLSEnabled() {
		t.Fatal("expected TLS disabled without key")
	}
}

func TestTLSEnabled_both_missing(t *testing.T) {
	os.Unsetenv("TLS_CERT_FILE")
	os.Unsetenv("TLS_KEY_FILE")

	if TLSEnabled() {
		t.Fatal("expected TLS disabled")
	}
}
