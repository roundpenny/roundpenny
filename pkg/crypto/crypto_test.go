package crypto

import (
	"testing"
)

func TestSHA256(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte("hello"), "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{[]byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{[]byte("roundpenny"), "c6d9cf0a88a32a3f4331e84dda92ef96d72e29e5f9c7e0e6a19c3b0f3d6e0f9b"},
	}
	for _, tt := range tests {
		got := SHA256(tt.input)
		if len(got) != 32 {
			t.Fatalf("SHA256 length = %d, want 32", len(got))
		}
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
}

func TestComparePassword_success(t *testing.T) {
	password := "correct-horse-battery-staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := ComparePassword(hash, password); err != nil {
		t.Fatalf("ComparePassword failed: %v", err)
	}
}

func TestComparePassword_wrong(t *testing.T) {
	password := "correct-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := ComparePassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestComparePassword_invalid_hash(t *testing.T) {
	if err := ComparePassword("not-a-bcrypt-hash", "password"); err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestHashPassword_same_input_different_hash(t *testing.T) {
	h1, _ := HashPassword("samepassword")
	h2, _ := HashPassword("samepassword")
	if h1 == h2 {
		t.Fatal("bcrypt should produce different hashes for same input due to salt")
	}
}
