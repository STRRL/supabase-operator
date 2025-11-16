// Package database provides PostgreSQL database initialization for Supabase.
//
// This package handles the initial setup of a PostgreSQL database for use with
// Supabase. It creates the necessary schemas, roles, and extensions required
// for all Supabase components to function correctly.
//
// Initialization Steps:
//
// The package performs the following operations (all idempotent):
//
//  1. Create Extensions:
//     - pgcrypto: Cryptographic functions for Supabase
//     - uuid-ossp: UUID generation
//     - pg_stat_statements: Query performance tracking
//
//  2. Create Schemas:
//     - auth: Authentication data (used by GoTrue)
//     - storage: File metadata (used by Storage API)
//     - realtime: Subscription tracking (used by Realtime)
//
//  3. Create Roles:
//     - authenticator: Request authenticator role (used by Kong/PostgREST)
//     - anon: Anonymous access role
//     - service_role: Service-level access role (bypasses RLS)
//
//  4. Grant Permissions:
//     - Schema usage permissions
//     - Role membership configuration
//     - Default privileges for future objects
//
// Database Requirements:
//
// The database user must have the following PostgreSQL privileges:
//   - CREATEDB or SUPERUSER
//   - Ability to create roles
//   - Ability to create extensions
//   - Ability to create event triggers (superuser required)
//
// Recommended users from supabase/postgres image:
//   - postgres (superuser)
//   - supabase_admin (has required privileges)
//
// Example usage:
//
//	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
//	    username, password, host, port, database, sslMode)
//	err := database.InitializeDatabase(ctx, connStr)
//
// Idempotency:
//
// All operations use "IF NOT EXISTS" clauses and are safe to run multiple times.
// The function will not fail if extensions, schemas, or roles already exist.
// This allows the operator to safely retry initialization on failures.
//
// Error Handling:
//
// Initialization errors are returned to the caller for proper reconciliation handling.
// Common errors include:
//   - Connection failures (network, credentials)
//   - Permission denied (insufficient privileges)
//   - Extension not available (package not installed)
//
// The controller will set the phase to "Failed" and include the error in status
// for investigation.
//
// Security Considerations:
//
// - Connection strings should never be logged
// - Use SSL mode "require" or higher in production
// - Rotate database credentials regularly
// - Limit permissions of the database user to minimum required
package database
