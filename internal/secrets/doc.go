// Package secrets provides secure secret generation and validation for Supabase.
//
// This package handles two primary responsibilities:
//  1. JWT secret and API key generation
//  2. User-provided secret validation
//
// JWT Generation:
//
// The package generates cryptographically secure JWT secrets and API keys:
//   - JWT Secret: 256-bit random value (base64-encoded)
//   - ANON_KEY: JWT token with 'anon' role claim (public API key)
//   - SERVICE_ROLE_KEY: JWT token with 'service_role' role claim (admin API key)
//
// All JWT tokens are signed using HS256 (HMAC-SHA256) and include:
//   - iss: Issuer claim (supabase)
//   - iat: Issued at timestamp
//   - exp: Expiration timestamp (10 years from issuance)
//   - role: Role claim (anon or service_role)
//
// Example usage:
//
//	jwtSecret, err := secrets.GenerateJWTSecret()
//	anonKey, err := secrets.GenerateAnonKey(jwtSecret)
//	serviceKey, err := secrets.GenerateServiceRoleKey(jwtSecret)
//
// Secret Validation:
//
// The package validates user-provided secrets contain required keys:
//
// Database secret must contain:
//   - host: PostgreSQL hostname
//   - port: PostgreSQL port
//   - database: Database name
//   - username: PostgreSQL username
//   - password: PostgreSQL password
//
// Storage secret must contain:
//   - endpoint: S3-compatible endpoint URL
//   - region: Storage region
//   - bucket: Bucket name
//   - accessKeyId: S3 access key ID
//   - secretAccessKey: S3 secret access key
//
// Example usage:
//
//	err := secrets.ValidateDatabaseSecret(secret)
//	err := secrets.ValidateStorageSecret(secret)
//
// Security Considerations:
//
// - JWT secrets use crypto/rand for cryptographic security
// - Generated secrets are never logged
// - Validation ensures no empty values
// - All operations are idempotent
package secrets
