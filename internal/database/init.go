/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type InitConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

func InitializeDatabase(ctx context.Context, config InitConfig) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(ctx); closeErr != nil {
			// Log the error but don't override the return value
			// since this is cleanup code
			_ = closeErr
		}
	}()

	if err := createExtensions(ctx, conn); err != nil {
		return fmt.Errorf("failed to create extensions: %w", err)
	}

	if err := createSchemas(ctx, conn); err != nil {
		return fmt.Errorf("failed to create schemas: %w", err)
	}

	if err := createRoles(ctx, conn); err != nil {
		return fmt.Errorf("failed to create roles: %w", err)
	}

	return nil
}

func createExtensions(ctx context.Context, conn *pgx.Conn) error {
	extensions := []string{
		"pgcrypto",
		"pgjwt",
		"uuid-ossp",
		"pg_stat_statements",
	}

	for _, ext := range extensions {
		sql := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext)
		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to create extension %s: %w", ext, err)
		}
	}

	return nil
}

func createSchemas(ctx context.Context, conn *pgx.Conn) error {
	schemas := []string{
		"auth",
		"storage",
		"realtime",
	}

	for _, schema := range schemas {
		sql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)
		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to create schema %s: %w", schema, err)
		}
	}

	return nil
}

func createRoles(ctx context.Context, conn *pgx.Conn) error {
	roles := map[string]string{
		"authenticator": "NOLOGIN",
		"anon":          "NOLOGIN",
		"service_role":  "NOLOGIN BYPASSRLS",
	}

	for role, attrs := range roles {
		sql := fmt.Sprintf("DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN CREATE ROLE %s %s; END IF; END $$", role, role, attrs)
		if _, err := conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("failed to create role %s: %w", role, err)
		}
	}

	grantSQL := `
		GRANT USAGE ON SCHEMA auth TO authenticator;
		GRANT ALL ON ALL TABLES IN SCHEMA auth TO authenticator;
		GRANT ALL ON ALL SEQUENCES IN SCHEMA auth TO authenticator;
		GRANT ALL ON ALL FUNCTIONS IN SCHEMA auth TO authenticator;

		GRANT USAGE ON SCHEMA storage TO authenticator;
		GRANT ALL ON ALL TABLES IN SCHEMA storage TO authenticator;
		GRANT ALL ON ALL SEQUENCES IN SCHEMA storage TO authenticator;
		GRANT ALL ON ALL FUNCTIONS IN SCHEMA storage TO authenticator;

		GRANT USAGE ON SCHEMA realtime TO authenticator;
		GRANT ALL ON ALL TABLES IN SCHEMA realtime TO authenticator;
		GRANT ALL ON ALL SEQUENCES IN SCHEMA realtime TO authenticator;
		GRANT ALL ON ALL FUNCTIONS IN SCHEMA realtime TO authenticator;

		GRANT authenticator TO anon;
		GRANT authenticator TO service_role;
	`

	if _, err := conn.Exec(ctx, grantSQL); err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}

	return nil
}

func GetInitSQL() []string {
	return []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE EXTENSION IF NOT EXISTS "pgjwt"`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS "pg_stat_statements"`,
		`CREATE SCHEMA IF NOT EXISTS auth`,
		`CREATE SCHEMA IF NOT EXISTS storage`,
		`CREATE SCHEMA IF NOT EXISTS realtime`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'authenticator') THEN CREATE ROLE authenticator NOLOGIN; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'anon') THEN CREATE ROLE anon NOLOGIN; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'service_role') THEN CREATE ROLE service_role NOLOGIN BYPASSRLS; END IF; END $$`,
		`GRANT USAGE ON SCHEMA auth TO authenticator`,
		`GRANT ALL ON ALL TABLES IN SCHEMA auth TO authenticator`,
		`GRANT ALL ON ALL SEQUENCES IN SCHEMA auth TO authenticator`,
		`GRANT ALL ON ALL FUNCTIONS IN SCHEMA auth TO authenticator`,
		`GRANT USAGE ON SCHEMA storage TO authenticator`,
		`GRANT ALL ON ALL TABLES IN SCHEMA storage TO authenticator`,
		`GRANT ALL ON ALL SEQUENCES IN SCHEMA storage TO authenticator`,
		`GRANT ALL ON ALL FUNCTIONS IN SCHEMA storage TO authenticator`,
		`GRANT USAGE ON SCHEMA realtime TO authenticator`,
		`GRANT ALL ON ALL TABLES IN SCHEMA realtime TO authenticator`,
		`GRANT ALL ON ALL SEQUENCES IN SCHEMA realtime TO authenticator`,
		`GRANT ALL ON ALL FUNCTIONS IN SCHEMA realtime TO authenticator`,
		`GRANT authenticator TO anon`,
		`GRANT authenticator TO service_role`,
	}
}
