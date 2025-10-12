# Database Initialization

## Overview

The Supabase Operator initializes PostgreSQL databases using a **Kubernetes Job-based approach** instead of direct database connections from the controller. This design provides better isolation, retry logic, and debugging capabilities.

## Architecture

```
SupabaseProject CR
    ↓
Controller (InitializingDatabase phase)
    ↓
Creates ConfigMap (embedded SQL)
    ↓
Creates Job (postgres:15-alpine + psql)
    ↓
Job executes SQL against user's database
    ↓
Controller monitors Job status
    ↓
Continues to next phase
```

## Components

### 1. SQL Scripts (Embedded in Binary)

**Location:** `internal/database/migrations/sql/`

We use the **exact same SQL files from upstream Supabase** ([docker/volumes/db/](https://github.com/supabase/supabase/tree/master/docker/volumes/db)):

- `00-initial-schema.sql` - Creates `_supabase` database
- `01-roles.sql` - Sets passwords for roles
- `02-jwt.sql` - JWT configuration
- `03-logs.sql` - Logging setup
- `04-webhooks.sql` - Webhook configuration
- `05-realtime.sql` - Realtime schema setup
- `06-pooler.sql` - Connection pooler setup

These files are:
- ✅ **Identical to docker-compose** - No modifications, 100% upstream compatibility
- ✅ **Synced automatically** - Use `./hack/sync-migrations.sh` to update
- ✅ **Embedded at build time** - Single binary distribution

### 2. ConfigMap

**Builder:** `resources.BuildDatabaseInitConfigMap()`

Contains:
- All 7 upstream SQL files (embedded from binary)
- A shell script (`run-migrations.sh`) that executes them in order

The ConfigMap is mounted into the Job pod at `/scripts/`.

### 3. Kubernetes Job

**Builder:** `resources.BuildDatabaseInitJob()`

Specifications:
- **Image**: `postgres:15-alpine` (includes psql client)
- **Command**: `sh /scripts/run-migrations.sh`
- **Environment**: Sets `POSTGRES_USER`, `POSTGRES_PASSWORD` (for psql variable substitution)
- **Retry**: Up to 3 attempts (BackoffLimit: 3)
- **Cleanup**: Auto-delete after 10 minutes (TTLSecondsAfterFinished: 600)
- **Credentials**: From user's database Secret

The shell script:
1. Executes each SQL file in order: `00-initial-schema.sql` → `01-roles.sql` → ... → `06-pooler.sql`
2. Uses `ON_ERROR_STOP=0` (continue on errors) to handle various database states
3. Logs progress for each file

### 4. Controller Logic

**Method:** `ensureDatabaseInitJob()`

Flow:
1. Create ConfigMap with SQL scripts (if not exists)
2. Create Job (if not exists)
3. Monitor Job status:
   - **Succeeded**: Continue to next phase
   - **Failed**: Check retry count, requeue or fail
   - **Running**: Requeue and check later (every 5s)

## Reconciliation Flow

```
PhasePending
    ↓
PhaseValidatingDependencies (validate secrets)
    ↓
PhaseInitializingDatabase (run Job)
    ↓ (Job succeeded)
PhaseDeployingSecrets (JWT secrets)
    ↓
PhaseDeployingComponents (Kong, Auth, etc.)
    ↓
PhaseRunning
```

## Error Handling

### Partial Failures (Tolerated)

The SQL script uses `IF NOT EXISTS` checks, so:
- ✅ Extensions already exist → Skip
- ✅ Roles already exist → Skip
- ✅ Insufficient privileges for extensions → Warning (services may still work)

### Complete Failures

If Job fails after 3 retries:
- Status → `PhaseFailed`
- Error message includes job failure details
- User can check Job logs: `kubectl logs job/<project-name>-db-init`

## Syncing with Upstream

The SQL scripts are synced from [supabase/supabase](https://github.com/supabase/supabase/tree/master/docker/volumes/db):

```bash
# Sync latest from upstream
./hack/sync-migrations.sh

# Sync from specific version
./hack/sync-migrations.sh v1.2.3

# Review changes
git diff internal/database/migrations/sql/

# Update operator-init.sql if needed
# Then commit
```

## Debugging

### Check Job Status

```bash
# View job
kubectl get job <project-name>-db-init -n <namespace>

# View job logs
kubectl logs job/<project-name>-db-init -n <namespace>

# View job details
kubectl describe job <project-name>-db-init -n <namespace>
```

### Check ConfigMap

```bash
# View SQL script
kubectl get configmap <project-name>-db-init -o yaml -n <namespace>
```

### Manual Execution

For testing, you can run the SQL manually:

```bash
# Get the SQL
kubectl get configmap <project-name>-db-init -o jsonpath='{.data.init\.sql}' > init.sql

# Run against your database
psql "postgresql://user:pass@host:port/db?sslmode=require" -f init.sql
```

## Comparison with docker-compose

| Aspect | docker-compose | Kubernetes Operator |
|--------|---------------|---------------------|
| **SQL Files** | `docker/volumes/db/*.sql` | **Same files (embedded)** |
| **Execution Order** | Sequential shell script | **Same order, same script logic** |
| **psql Variables** | `POSTGRES_USER`, `POSTGRES_PASSWORD` | **Same variables** |
| **Error Handling** | Continue on errors | **Same (ON_ERROR_STOP=0)** |
| **Database** | Created by postgres container | User-provided (may exist) |
| **Mounting** | Docker volume | Kubernetes ConfigMap |
| **Runtime** | Docker container | Kubernetes Job |
| **Retry** | No automatic retry | 3 retries via Job |
| **Logs** | `docker logs` | `kubectl logs` |

**Key Insight:** We achieve **100% compatibility** by using the exact same SQL files and execution logic. The only differences are infrastructure-related (how files are mounted and where they run).

## Future Enhancements

- [ ] Optional: Skip initialization if database already initialized
- [ ] Support custom SQL scripts via ConfigMap override
- [ ] Add migration version tracking
- [ ] Support for database schema upgrades
- [ ] Health checks before proceeding to next phase
