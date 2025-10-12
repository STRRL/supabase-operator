-- NOTE: change to your own passwords for production environments
\set pgpass `echo "$POSTGRES_PASSWORD"`

-- Create required Supabase roles
-- supabase/postgres:15.14.1.021+ does not pre-create these roles
CREATE ROLE anon                 NOLOGIN NOINHERIT;
CREATE ROLE authenticated        NOLOGIN NOINHERIT;
CREATE ROLE service_role         NOLOGIN NOINHERIT BYPASSRLS;
CREATE ROLE supabase_auth_admin  NOLOGIN CREATEROLE;
CREATE ROLE supabase_storage_admin NOLOGIN CREATEROLE;
CREATE ROLE supabase_functions_admin NOLOGIN CREATEROLE CREATE ROLE;
CREATE ROLE authenticator        NOINHERIT LOGIN PASSWORD :'pgpass';
CREATE ROLE pgbouncer            LOGIN PASSWORD :'pgpass';
CREATE ROLE supabase_admin       SUPERUSER CREATEDB CREATEROLE REPLICATION BYPASSRLS LOGIN PASSWORD :'pgpass';

-- Grant API roles to authenticator so it can switch to them via JWTs
GRANT anon              TO authenticator;
GRANT authenticated     TO authenticator;
GRANT service_role      TO authenticator;
