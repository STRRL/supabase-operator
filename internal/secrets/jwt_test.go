package secrets

import (
	"encoding/base64"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTSecret(t *testing.T) {
	secret, err := GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret() error = %v", err)
	}

	if len(secret) == 0 {
		t.Error("GenerateJWTSecret() returned empty secret")
	}

	decoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Errorf("GenerateJWTSecret() returned invalid base64: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("GenerateJWTSecret() decoded length = %d, want 32", len(decoded))
	}

	secret2, err := GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret() second call error = %v", err)
	}

	if secret == secret2 {
		t.Error("GenerateJWTSecret() should generate different secrets each time")
	}
}

func TestGeneratePGMetaCryptoKey(t *testing.T) {
	key, err := GeneratePGMetaCryptoKey()
	if err != nil {
		t.Fatalf("GeneratePGMetaCryptoKey() error = %v", err)
	}

	if len(key) == 0 {
		t.Error("GeneratePGMetaCryptoKey() returned empty key")
	}

	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Errorf("GeneratePGMetaCryptoKey() returned invalid base64: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("GeneratePGMetaCryptoKey() decoded length = %d, want 32", len(decoded))
	}

	key2, err := GeneratePGMetaCryptoKey()
	if err != nil {
		t.Fatalf("GeneratePGMetaCryptoKey() second call error = %v", err)
	}

	if key == key2 {
		t.Error("GeneratePGMetaCryptoKey() should generate different secrets each time")
	}
}

func TestGenerateAnonKey(t *testing.T) {
	jwtSecret := "dGVzdC1zZWNyZXQtdGhhdC1pcy1sb25nLWVub3VnaAAAAAAAAAAAAAAAAAAAAA=="

	anonKey, err := GenerateAnonKey(jwtSecret)
	if err != nil {
		t.Fatalf("GenerateAnonKey() error = %v", err)
	}

	if len(anonKey) == 0 {
		t.Error("GenerateAnonKey() returned empty key")
	}

	decoded, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to decode JWT secret: %v", err)
	}

	token, err := jwt.Parse(anonKey, func(token *jwt.Token) (interface{}, error) {
		return decoded, nil
	})

	if err != nil {
		t.Errorf("GenerateAnonKey() generated invalid JWT: %v", err)
	}

	if !token.Valid {
		t.Error("GenerateAnonKey() generated invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("GenerateAnonKey() token claims not MapClaims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		t.Error("GenerateAnonKey() token missing 'role' claim")
	}

	if role != "anon" {
		t.Errorf("GenerateAnonKey() role = %s, want 'anon'", role)
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		t.Error("GenerateAnonKey() token missing 'iss' claim")
	}

	if iss != "supabase" {
		t.Errorf("GenerateAnonKey() iss = %s, want 'supabase'", iss)
	}
}

func TestGenerateServiceRoleKey(t *testing.T) {
	jwtSecret := "dGVzdC1zZWNyZXQtdGhhdC1pcy1sb25nLWVub3VnaAAAAAAAAAAAAAAAAAAAAA=="

	serviceKey, err := GenerateServiceRoleKey(jwtSecret)
	if err != nil {
		t.Fatalf("GenerateServiceRoleKey() error = %v", err)
	}

	if len(serviceKey) == 0 {
		t.Error("GenerateServiceRoleKey() returned empty key")
	}

	decoded, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to decode JWT secret: %v", err)
	}

	token, err := jwt.Parse(serviceKey, func(token *jwt.Token) (interface{}, error) {
		return decoded, nil
	})

	if err != nil {
		t.Errorf("GenerateServiceRoleKey() generated invalid JWT: %v", err)
	}

	if !token.Valid {
		t.Error("GenerateServiceRoleKey() generated invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("GenerateServiceRoleKey() token claims not MapClaims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		t.Error("GenerateServiceRoleKey() token missing 'role' claim")
	}

	if role != "service_role" {
		t.Errorf("GenerateServiceRoleKey() role = %s, want 'service_role'", role)
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		t.Error("GenerateServiceRoleKey() token missing 'iss' claim")
	}

	if iss != "supabase" {
		t.Errorf("GenerateServiceRoleKey() iss = %s, want 'supabase'", iss)
	}
}

func TestGenerateAnonKey_InvalidSecret(t *testing.T) {
	_, err := GenerateAnonKey("invalid-base64!")
	if err == nil {
		t.Error("GenerateAnonKey() with invalid secret should return error")
	}
}

func TestGenerateServiceRoleKey_InvalidSecret(t *testing.T) {
	_, err := GenerateServiceRoleKey("invalid-base64!")
	if err == nil {
		t.Error("GenerateServiceRoleKey() with invalid secret should return error")
	}
}

func TestGenerateAnonKey_EmptySecret(t *testing.T) {
	_, err := GenerateAnonKey("")
	if err == nil {
		t.Error("GenerateAnonKey() with empty secret should return error")
	}
}

func TestGenerateServiceRoleKey_EmptySecret(t *testing.T) {
	_, err := GenerateServiceRoleKey("")
	if err == nil {
		t.Error("GenerateServiceRoleKey() with empty secret should return error")
	}
}
