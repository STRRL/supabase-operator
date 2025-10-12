# Supabase Database Migrations

This directory contains SQL migration scripts for initializing a PostgreSQL database for Supabase.

## Architecture

The operator uses **embedded SQL scripts** that are compiled into the Go binary. This approach:

- ✅ Provides a single binary distribution (no external dependencies)
- ✅ Version locks SQL scripts with operator releases
- ✅ Simplifies deployment (no ConfigMap/Volume management needed by users)

## Files

### `sql/00-*.sql` to `sql/06-*.sql`
These are the **original upstream SQL files** from [supabase/supabase](https://github.com/supabase/supabase/tree/master/docker/volumes/db).

**Important:** We use these files **exactly as-is** from upstream, maintaining 100% compatibility with docker-compose behavior:

- `00-initial-schema.sql` - Creates `_supabase` database
- `01-roles.sql` - Sets passwords for roles
- `02-jwt.sql` - JWT configuration
- `03-logs.sql` - Logging setup
- `04-webhooks.sql` - Webhook configuration
- `05-realtime.sql` - Realtime schema setup
- `06-pooler.sql` - Connection pooler setup

### How It Works

The operator executes these files **in order** using a shell script (`run-migrations.sh`) that:

1. Sets environment variables that psql scripts expect (`POSTGRES_USER`, `POSTGRES_PASSWORD`)
2. Executes each SQL file sequentially
3. Continues on errors (some operations may fail in Kubernetes context, which is OK)

This approach ensures we stay **100% aligned** with upstream Supabase initialization behavior.

## Syncing with Upstream

### Automated Sync

Run the sync script to download the latest SQL files from upstream:

```bash
# Sync from master branch
./hack/sync-migrations.sh

# Sync from a specific version
./hack/sync-migrations.sh v1.2.3
```

### Manual Review Process

After syncing:

1. **Review changes:**
   ```bash
   git diff internal/database/migrations/sql/
   ```

2. **Verify execution order:**
   - Check if new files were added to upstream
   - Update `migrations.go` and `database_init.go` if the file list changed

3. **Test the changes:**
   ```bash
   make test
   make test-e2e
   ```

4. **Commit:**
   ```bash
   git add internal/database/migrations/
   git commit -m "chore: sync database migrations from supabase@v1.2.3"
   ```

**Note:** We use upstream files **exactly as-is** without modification. Any incompatibilities are handled at runtime by the Job's error handling.

## How It Works

### At Build Time

1. All upstream SQL files are embedded into the Go binary using `//go:embed`
2. A shell script (`run-migrations.sh`) is generated to execute them in order
3. The binary contains: 7 SQL files + 1 execution script

### At Runtime (Kubernetes)

When a `SupabaseProject` is created:

1. **Controller creates a ConfigMap** with:
   - All 7 upstream SQL files (00-initial-schema.sql through 06-pooler.sql)
   - The execution script (run-migrations.sh)

2. **Controller creates a Job** that:
   - Uses `postgres:15-alpine` image (includes psql client)
   - Mounts the ConfigMap containing all scripts
   - Sets environment variables (`POSTGRES_USER`, `POSTGRES_PASSWORD`)
   - Runs: `sh /scripts/run-migrations.sh`

3. **The script executes SQL files sequentially:**
   - Each file runs with `psql $DATABASE_URL -f /scripts/XX-*.sql`
   - Continues on errors (some operations may fail, which is OK)
   - Logs progress for each file

4. **Job completes** (or fails with retry up to 3 times)
5. **Controller proceeds** to deploy Supabase services
6. **Services run their own migrations** (Auth, Storage, Realtime create schemas/tables)

## Failure Handling

The initialization Job uses **ON_ERROR_STOP=0** (continue on errors) because:

- **Database creation** (`_supabase`) may fail (user provides existing DB) → OK, continue
- **Role password changes** may fail (roles don't exist yet) → OK, continue
- **Extensions** may fail (insufficient privileges) → OK, services still work
- **Schemas** may already exist → OK, skip creation

This tolerance for errors allows:
- ✅ Work with various database permission levels
- ✅ Re-run initialization safely (idempotent)
- ✅ Match docker-compose behavior (it also ignores some errors)

The Job only fails if **all retries** (3 attempts) are exhausted or if there's a connection issue.

## Comparison with docker-compose

| Feature | docker-compose | Kubernetes Operator |
|---------|---------------|---------------------|
| **SQL Files** | Files in `docker/volumes/db/` | **Same files, embedded in binary** |
| **Execution Order** | Sequential via shell | **Same order via shell script** |
| **psql Variables** | Set from env vars | **Same mechanism** |
| **Error Handling** | Continue on some errors | **Same (ON_ERROR_STOP=0)** |
| **Database** | Created by postgres container | User-provided (may already exist) |
| **Execution Method** | Volume mount | Kubernetes ConfigMap mount |
| **Runtime** | Docker container | Kubernetes Job |

**Key Point:** We use the **exact same SQL files** from upstream, executed in the **exact same order**, with the **same error handling**. The only differences are infrastructure-related (ConfigMap vs volume, Job vs container).

## Future Improvements

- [ ] Support custom SQL scripts via ConfigMap override
- [ ] Add Helm chart option for SQL distribution
- [ ] Add migration version tracking
- [ ] Support for database upgrades/migrations
