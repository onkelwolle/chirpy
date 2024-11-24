package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("cannot generate refresh token: %w", err)
	}

	refreshToken := hex.EncodeToString(tokenBytes)
	return refreshToken, nil
}
