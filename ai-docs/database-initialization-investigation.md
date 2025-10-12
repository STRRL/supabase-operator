# Database Initialization Guide for Supabase Operator

## Overview

This guide explains the database initialization strategy for the Supabase Operator, including implementation details and the solution to schema permission issues.

**Last Updated**: 2025-10-11 - Updated with GoTrue v2.177+ migration bug fix and clean image recommendation

## Recommended Approach: Use supabase/postgres Image

The operator requires users to provide PostgreSQL using the **`supabase/postgres` image**. This approach ensures 100% compatibility with Supabase's official stack and simplifies the initialization process.

### Why This Approach?

1. **Upstream Compatibility** - Uses the same image and initialization scripts as Supabase's official docker-compose setup
2. **Pre-configured Roles** - The image comes with all required PostgreSQL roles and extensions
3. **Proven Solution** - Well-tested by the Supabase community
4. **Simplified Implementation** - No need to handle multiple PostgreSQL versions or cloud provider limitations

## PostgreSQL Requirements

### Required Image

Users must deploy PostgreSQL using:
- **Recommended Image**: `supabase/postgres:15.14.1.021` or later
- **Why this version?**: Older versions (e.g., 15.8.1.085) pre-populate auth/storage schemas with partial migration state, causing conflicts with GoTrue v2.177+ migration system. Version 15.14.1.021+ provides a clean slate.
- **Source**: https://hub.docker.com/r/supabase/postgres

### Pre-created Roles

The `supabase/postgres` image includes these roles out of the box:

```sql
-- Admin roles
supabase_admin                    (superuser, createdb, createrole, replication, bypassrls)
supabase_replication_admin        (login, replication)
supabase_etl_admin                (login, replication)
supabase_read_only_user           (login, bypassrls)

-- API roles
authenticator                     (noinherit)
anon                              (nologin, noinherit)
authenticated                     (nologin, noinherit)
service_role                      (nologin, noinherit, bypassrls)

-- Service-specific admin roles
supabase_auth_admin               (login)
supabase_storage_admin            (login)
supabase_functions_admin          (login, create role)
```

## Database Connection Configuration

Users provide database credentials via Kubernetes Secret with these fields:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
stringData:
  host: postgres.namespace.svc.cluster.local
  port: "5432"
  database: postgres                    # Must be "postgres"
  username: postgres                    # MUST be "postgres" (see permission fix below)
  password: your-secure-password
```

**Important Notes:**
- `database` must be `postgres` (matching docker-compose behavior)
- `username` should be `postgres` for simplicity
- All Supabase services use the same database user for connections
- The operator creates empty schemas that services populate via their own migrations

## Initialization Process

### Phase Flow

The operator implements an `InitializingDatabase` phase that executes SQL initialization scripts:

```
PhaseValidatingDependencies
    ↓
PhaseInitializingDatabase        # Execute SQL scripts via Job
    ↓
PhaseDeployingSecrets            # Generate JWT secrets
    ↓
PhaseDeployingComponents         # Deploy services
    ↓
PhaseRunning
```

### SQL Scripts Execution

The initialization Job runs 7 SQL files in sequence (synced from upstream docker-compose):

1. **00-initial-schema.sql** - Creates `_supabase` database, auth/storage/realtime schemas, and critical enum types
2. **01-roles.sql** - Sets passwords for service roles
3. **02-jwt.sql** - Configures JWT settings in database
4. **03-logs.sql** - Creates `_analytics` schema in `_supabase` database
5. **04-webhooks.sql** - Sets up webhooks and functions schemas
6. **05-realtime.sql** - Configures realtime publication schemas
7. **06-pooler.sql** - Sets up connection pooler configuration

**CRITICAL: 00-initial-schema.sql GoTrue Migration Bug Fix**

The script pre-creates essential PostgreSQL enum types to work around a migration bug in GoTrue v2.177+:

```sql
-- GoTrue v2.177+ skips early migrations and jumps directly to 2024-2025 migrations
-- These migrations expect certain enum types to exist, causing fatal errors if missing
-- See: https://github.com/supabase/auth/issues/1729

CREATE TYPE auth.factor_type AS ENUM('totp', 'webauthn');
CREATE TYPE auth.code_challenge_method AS ENUM('s256', 'plain');
```

Without these pre-created types, Auth service fails with errors like:
- `ERROR: type "auth.factor_type" does not exist (SQLSTATE 42704)`
- `ERROR: type "auth.code_challenge_method" does not exist (SQLSTATE 42704)`

### Implementation Details

**Job Configuration:**
- Uses `postgres:15-alpine` image with psql
- SQL scripts provided via ConfigMap
- Connects using user-provided credentials (postgres user)
- Scripts use `\c` to switch databases (psql reuses connection parameters)

**Script Location:**
- Source: `internal/database/migrations/sql/*.sql`
- Embedded: `internal/database/migrations/migrations.go`
- Upstream: https://github.com/supabase/supabase/tree/master/docker/volumes/db (adapted for operator use)

## Database Structure

After initialization, the database structure matches docker-compose:

```
PostgreSQL Server
├── postgres database                # Main database (all services connect here)
│   ├── public schema
│   ├── auth schema                  # Created by Auth service on startup
│   ├── storage schema               # Created by Storage service on startup
│   ├── realtime schema              # Created by Realtime service on startup
│   ├── _realtime schema             # Created by init scripts
│   └── supabase_functions schema    # Created by init scripts
│
└── _supabase database               # Analytics database
    └── _analytics schema            # Created by init scripts
```

## Service Database User Strategy

**Unified User Approach:**
All Supabase services use the **`postgres`** user for database connections.

| Service | Database User | Schema Managed |
|---------|---------------|----------------|
| Auth | `postgres` | `auth` schema |
| Storage | `postgres` | `storage` schema |
| PostgREST | `postgres` | Queries via `authenticator` role context |
| Realtime | `postgres` | `realtime` and `_realtime` schemas |
| Meta | `postgres` | All schemas (introspection) |
| Kong | `postgres` | N/A (API gateway) |

**Why Use postgres User?**

Using a single postgres user simplifies the setup:
- All schemas are created with `postgres` as owner
- No complex role membership grants needed
- Services can run their migrations without permission issues
- Clean separation from supabase/postgres image's internal roles

## GoTrue v2.177+ Migration Bug and Workaround

### Problem Discovery

**Issue Reference:** https://github.com/supabase/auth/issues/1729

During testing with GoTrue v2.177+, we discovered a critical migration bug:

**Error Symptoms:**
- Auth pod in CrashLoopBackOff
- Error: `ERROR: type "auth.factor_type" does not exist (SQLSTATE 42704)` during migration
- Error: `ERROR: type "auth.code_challenge_method" does not exist (SQLSTATE 42704)` during migration
- Errors occur even on completely fresh, empty auth schema

**Root Cause Analysis:**

GoTrue v2.177+ has a migration system bug where:
1. It skips early migrations (2017-2018) that create base enum types
2. Jumps directly to 2024-2025 migrations
3. These newer migrations assume enum types already exist (from skipped migrations)
4. Result: Fatal errors when trying to alter non-existent types

This affects migrations:
- `20240729123726_add_mfa_phone_config.up.sql` - expects `auth.factor_type`
- `20250804100000_add_oauth_authorizations_consents.up.sql` - expects `auth.code_challenge_method`

### Solution: Pre-create Required Enum Types

**Strategy:**
Pre-create the enum types that GoTrue's migration system expects in our `00-initial-schema.sql`.

**Implementation in `internal/database/migrations/sql/00-initial-schema.sql`:**

```sql
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
```

**Why This Works:**

- These types match exactly what the skipped early migrations would have created
- `auth.factor_type` is used for MFA (multi-factor authentication) with TOTP/WebAuthn
- `auth.code_challenge_method` is used for PKCE (Proof Key for Code Exchange) in OAuth flows
- Creating them idempotently prevents errors if GoTrue's migration system is ever fixed

**Verification:**

Check that types exist before Auth service starts:

```bash
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "\dT auth.*"
```

Expected output:
```
                    List of data types
 Schema |          Name           | Internal name | Size | Elements
--------+-------------------------+---------------+------+----------
 auth   | code_challenge_method   | ...           | ...  | s256+plain
 auth   | factor_type             | ...           | ...  | totp+webauthn
```

## Key Considerations

### Why Not Support External Databases (RDS/Cloud SQL)?

External managed databases have limitations:
- No `CREATE DATABASE` privilege
- Restricted `CREATE EXTENSION` permissions
- May not allow `CREATE ROLE` operations
- Complex role/permission management

**Future Extension (V2):**
If external database support is needed:
- Add `spec.database.mode: managed | external`
- Provide separate SQL scripts for external mode
- Document feature limitations
- Require users to manually create certain roles

### Why Services Don't Auto-Initialize Everything?

While Supabase services (Auth, Storage, Realtime) do run migrations and create their own schemas, they require:
- Certain base schemas to exist (`_realtime`)
- JWT configuration in database settings
- Proper role setup and permissions
- The `_supabase` database for analytics

The initialization Job ensures these prerequisites are in place before services start.

## Troubleshooting

### Common Issues

**Issue: Job fails with "role does not exist"**
- Cause: Not using `supabase/postgres` image
- Solution: Deploy PostgreSQL using the required image

**Issue: Services fail with "permission denied for schema auth/storage"** ⚠️ CRITICAL
- Symptoms: Auth/Storage pods in CrashLoopBackOff, logs show "ERROR: permission denied for schema"
- Cause: `postgres` user doesn't have admin role memberships
- Root cause: Database initialized without running operator's 01-roles.sql properly
- Solution:
  1. Check if role grants succeeded: `kubectl logs job/project-db-init | grep "GRANT ROLE"`
  2. Verify postgres user has admin roles: `kubectl exec deployment/postgres -- psql -U postgres -c "\du postgres"`
  3. If missing, ensure operator's 01-roles.sql is using supabase_admin user (check database_init.go)

**Issue: Auth service reports "type auth.factor_type does not exist"** ⚠️ CRITICAL
- **Symptoms**: Auth pod in CrashLoopBackOff, logs show `ERROR: type "auth.factor_type" does not exist (SQLSTATE 42704)` during migration `20240729123726_add_mfa_phone_config.up.sql`
- **Root Cause**: GoTrue v2.177+ has a migration bug where it skips early migrations that create enum types and jumps directly to 2024-2025 migrations. See: https://github.com/supabase/auth/issues/1729
- **Solution**: Pre-create the enum types in `00-initial-schema.sql`:
  ```sql
  CREATE TYPE auth.factor_type AS ENUM('totp', 'webauthn');
  CREATE TYPE auth.code_challenge_method AS ENUM('s256', 'plain');
  ```
- **Prevention**: Use supabase/postgres:15.14.1.021+ which provides a clean database without pre-populated migration state

**Issue: Auth service reports "type auth.code_challenge_method does not exist"**
- **Symptoms**: Similar to above, but fails on migration `20250804100000_add_oauth_authorizations_consents.up.sql`
- **Root Cause**: Same GoTrue v2.177+ migration bug
- **Solution**: Same as above - pre-create enum types in init scripts

**Issue: "supabase_auth_admin role memberships are reserved"**
- Cause: Trying to grant reserved roles using non-superuser account
- Solution: Execute 01-roles.sql with supabase_admin user (has superuser privileges)
- Check: database_init.go should use ADMIN_DATABASE_URL for 01-roles.sql

**Issue: Services fail to start after initialization**
- Cause: JWT secret not created before init Job
- Solution: Verify operator creates JWT secret in `DeployingSecrets` phase (after init)

**Issue: Auth service reports "schema auth does not exist"**
- Cause: Database init Job failed silently
- Solution: Check Job logs with `kubectl logs job/project-name-db-init`

### Validation Commands

```bash
# Verify roles exist
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "\du" | grep -E "authenticator|supabase_admin"

# CRITICAL: Verify postgres user has admin role memberships
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "\du postgres"
# Expected: Should show supabase_auth_admin and supabase_storage_admin in "Member of" column

# Verify postgres user can create objects in auth schema
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "SELECT has_schema_privilege('postgres', 'auth', 'CREATE');"
# Expected: t (true)

# Verify databases created
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "\l" | grep _supabase

# Verify schemas created and their owners
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "\dn+" | grep -E "auth|storage|realtime"
# Check owners match expected values

# Verify JWT config
kubectl exec -n namespace deployment/postgres -- \
  psql -U postgres -c "SHOW app.settings.jwt_secret;"

# Check database initialization job logs
kubectl logs -n namespace job/project-name-db-init | grep "01-roles.sql" -A 5
# Should show "GRANT ROLE" messages without errors
```

## References

- [Supabase Docker Compose Configuration](https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml)
- [Supabase PostgreSQL Image](https://hub.docker.com/r/supabase/postgres)
- [Database Initialization Scripts](https://github.com/supabase/supabase/tree/master/docker/volumes/db)
- [GoTrue Migration Issue #1729](https://github.com/supabase/auth/issues/1729) - auth.factor_type does not exist
- [Manual MVP Testing Guide](./manually-mvp-testing.md)

## Implementation Checklist

When implementing database initialization:

- [ ] User provides `supabase/postgres:15.14.1.021+` image deployment
- [ ] Operator validates PostgreSQL connection in `ValidatingDependencies` phase
- [ ] Operator creates database init Job with SQL scripts
- [ ] Job executes all 7 SQL files successfully
  - [ ] 00-initial-schema.sql creates schemas and GoTrue enum types workaround
  - [ ] 01-roles.sql sets passwords for service roles
  - [ ] JWT configuration applied via 02-jwt.sql
- [ ] Verify enum types exist before services start
  - [ ] `auth.factor_type` created (workaround for GoTrue bug)
  - [ ] `auth.code_challenge_method` created (workaround for GoTrue bug)
- [ ] Operator waits for Job completion before deploying services
- [ ] Services start and run their own migrations
  - [ ] Auth service migrations complete without enum type errors
  - [ ] Storage service can create objects in storage schema
  - [ ] Realtime service can access realtime schemas
- [ ] All 6 services running (Kong, Auth, PostgREST, Realtime, Storage, Meta)
- [ ] No CrashLoopBackOff due to migration or permission errors
