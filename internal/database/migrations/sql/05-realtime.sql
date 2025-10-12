\set pguser `echo "$POSTGRES_USER"`

create schema if not exists _realtime;
alter schema _realtime owner to :pguser;

-- Pre-create schema_migrations tables with correct structure for Realtime
-- Realtime's Ecto migrator expects this table with inserted_at column
-- If we don't create it, Ecto will create it without inserted_at, causing failures
-- Realtime uses public.schema_migrations by default

-- Drop and recreate public.schema_migrations to ensure correct structure
-- (PostgREST or other services may have created it without inserted_at)
DROP TABLE IF EXISTS public.schema_migrations CASCADE;
CREATE TABLE public.schema_migrations (
  version BIGINT NOT NULL PRIMARY KEY,
  inserted_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Also create in realtime schema for consistency
CREATE TABLE IF NOT EXISTS realtime.schema_migrations (
  version BIGINT NOT NULL PRIMARY KEY,
  inserted_at TIMESTAMP NOT NULL DEFAULT NOW()
);
