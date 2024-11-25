package auth

import (
	"errors"
	"fmt"
	"net/http"
)

func GetAPIKey(header http.Header) (string, error) {
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("API key is missing")
	}

	if len(authHeader) < 7 || authHeader[:7] != "ApiKey " {
		return "", fmt.Errorf("invalid authorization header format")
	}

	apiKey := authHeader[7:]
	if apiKey == "" {
		return "", errors.New("API key is missing")
	}

	return apiKey, nil
}
