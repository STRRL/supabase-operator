package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func GenerateAnonKey(jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("jwt secret cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode jwt secret: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"role": "anon",
		"iss":  "supabase",
		"iat":  now.Unix(),
		"exp":  now.Add(10 * 365 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(decoded)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func GenerateServiceRoleKey(jwtSecret string) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("jwt secret cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode jwt secret: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"role": "service_role",
		"iss":  "supabase",
		"iat":  now.Unix(),
		"exp":  now.Add(10 * 365 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(decoded)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
