package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hashedPassword == "" {
		t.Fatalf("expected hashed password to be non-empty")
	}

	if err := ComparePassword(hashedPassword, password); err != nil {
		t.Fatalf("expected passwords to match, got %v", err)
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	password := ""
	_, err := HashPassword(password)
	if err == nil {
		t.Fatalf("expected error for empty password, got nil")
	}
}
