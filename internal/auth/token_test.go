package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Hour

	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		t.Fatalf("expected valid token claims, got invalid")
	}

	if claims.Issuer != "chirpy" {
		t.Errorf("expected issuer 'chirpy', got %v", claims.Issuer)
	}

	if claims.Subject != userID.String() {
		t.Errorf("expected subject %v, got %v", userID.String(), claims.Subject)
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Errorf("expected expiration time in the future, got %v", claims.ExpiresAt.Time)
	}
}
func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Hour

	// Create a valid token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test valid token
	validatedUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("expected userID %v, got %v", userID, validatedUserID)
	}

	// Test invalid token secret
	_, err = ValidateJWT(tokenString, "wrongsecret")
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	// Test expired token
	expiredTokenString, err := MakeJWT(userID, tokenSecret, -time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = ValidateJWT(expiredTokenString, tokenSecret)
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	// Test malformed token
	_, err = ValidateJWT("malformed.token.string", tokenSecret)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}
func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer validtoken123",
			expectedToken: "validtoken123",
			expectError:   false,
		},
		{
			name:        "missing authorization header",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "missing bearer token",
			authHeader:  "Bearer ",
			expectError: true,
		},
		{
			name:        "invalid authorization header format",
			authHeader:  "InvalidHeader",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Set("Authorization", tt.authHeader)

			token, err := GetBearerToken(headers)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token != tt.expectedToken {
					t.Errorf("expected token %v, got %v", tt.expectedToken, token)
				}
			}
		})
	}
}

