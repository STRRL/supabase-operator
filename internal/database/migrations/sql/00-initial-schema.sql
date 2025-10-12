\set pguser `echo "$POSTGRES_USER"`

-- Create _supabase database if it doesn't exist
SELECT 'CREATE DATABASE _supabase WITH OWNER ' || :'pguser'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '_supabase')\gexec

-- Create required schemas for Supabase services
-- These schemas will be populated by services themselves on first startup
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION :pguser;
CREATE SCHEMA IF NOT EXISTS storage AUTHORIZATION :pguser;
CREATE SCHEMA IF NOT EXISTS realtime AUTHORIZATION :pguser;

-- CRITICAL FIX: Create auth enum types that GoTrue v2.177+ expects
-- GoTrue v2.177+ has a bug where it skips early migrations and directly executes
-- 2024 migrations that expect these types to already exist
-- See: https://github.com/supabase/auth/issues/1729
DO $$ BEGIN
  CREATE TYPE auth.factor_type AS ENUM('totp', 'webauthn');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE auth.code_challenge_method AS ENUM('s256', 'plain');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
