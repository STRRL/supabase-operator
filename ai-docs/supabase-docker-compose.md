# Supabase Docker Compose Architecture Analysis

> Source: https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml

## Overview

The official Supabase Docker Compose configuration defines a complete self-hosted Supabase stack, containing 13 core services.

## Service Architecture

### 1. Studio (Management Interface)
**Image**: `supabase/studio:2025.10.01-sha-8460121`
**Port**: Internal 3000
**Purpose**: Web UI management interface for database management, user management, storage management, etc.

**Key Environment Variables**:
- `STUDIO_PG_META_URL`: Connection to Meta service
- `SUPABASE_URL`: Kong API Gateway address
- `LOGFLARE_URL`: Analytics service address
- `NEXT_ANALYTICS_BACKEND_PROVIDER`: Analytics backend (postgres/bigquery)

**Dependencies**: analytics

---

### 2. Kong (API Gateway)
**Image**: `kong:2.8.1`
**Ports**:
- 8000 (HTTP)
- 8443 (HTTPS)

**Purpose**: API Gateway, unified routing and authentication for all Supabase services

**Key Configuration**:
- `KONG_DATABASE`: "off" (DB-less mode)
- `KONG_DECLARATIVE_CONFIG`: Uses declarative configuration file
- `KONG_PLUGINS`: request-transformer,cors,key-auth,acl,basic-auth
- `KONG_DNS_ORDER`: LAST,A,CNAME (resolve DNS issues)
- `KONG_NGINX_PROXY_PROXY_BUFFER_SIZE`: 160k (large request support)

**Special Startup Command**:
```bash
bash -c 'eval "echo \"$$(cat ~/temp.yml)\"" > ~/kong.yml && /docker-entrypoint.sh kong docker-start'
```
> Dynamically replaces environment variables in config file at startup

**Dependencies**: analytics

---

### 3. Auth (Authentication Service)
**Image**: `supabase/gotrue:v2.180.0`
**Port**: 9999
**Purpose**: User authentication, authorization, JWT generation

**Key Environment Variables**:
- `GOTRUE_DB_DATABASE_URL`: Uses dedicated user `supabase_auth_admin`
- `GOTRUE_JWT_SECRET`: JWT signing key
- `GOTRUE_JWT_EXP`: JWT expiration time
- `GOTRUE_SITE_URL`: Frontend application URL
- `GOTRUE_EXTERNAL_EMAIL_ENABLED`: Email registration toggle
- `GOTRUE_EXTERNAL_PHONE_ENABLED`: Phone registration toggle
- `GOTRUE_MAILER_AUTOCONFIRM`: Auto-confirm email

**SMTP Configuration** (optional):
```yaml
GOTRUE_SMTP_HOST
GOTRUE_SMTP_PORT
GOTRUE_SMTP_USER
GOTRUE_SMTP_PASS
```

**Authentication Hooks** (optional):
- `GOTRUE_HOOK_CUSTOM_ACCESS_TOKEN`: Custom access token
- `GOTRUE_HOOK_MFA_VERIFICATION_ATTEMPT`: MFA verification hook
- `GOTRUE_HOOK_PASSWORD_VERIFICATION_ATTEMPT`: Password verification hook
- `GOTRUE_HOOK_SEND_SMS/EMAIL`: Custom SMS/email sending

**Dependencies**: db, analytics

---

### 4. REST (PostgREST)
**Image**: `postgrest/postgrest:v13.0.7`
**Port**: 3000
**Purpose**: Automatically converts PostgreSQL tables to RESTful API

**Key Environment Variables**:
- `PGRST_DB_URI`: Connects using `authenticator` user
- `PGRST_DB_SCHEMAS`: Exposed database schemas
- `PGRST_DB_ANON_ROLE`: Anonymous user role
- `PGRST_JWT_SECRET`: JWT verification key
- `PGRST_DB_USE_LEGACY_GUCS`: "false" (use new configuration method)

**Dependencies**: db, analytics

---

### 5. Realtime (Real-time Subscriptions)
**Image**: `supabase/realtime:v2.51.11`
**Port**: 4000
**Container Name**: `realtime-dev.supabase-realtime` (special naming for tenant ID parsing)

**Purpose**: Database change real-time push, Presence, Broadcast

**Key Environment Variables**:
- `DB_USER`: supabase_admin
- `DB_AFTER_CONNECT_QUERY`: 'SET search_path TO _realtime'
- `DB_ENC_KEY`: supabaserealtime
- `SECRET_KEY_BASE`: Phoenix application key
- `RLIMIT_NOFILE`: "10000" (file handle limit)
- `APP_NAME`: realtime
- `SEED_SELF_HOST`: true (self-hosted mode initialization)
- `RUN_JANITOR`: true (clean expired connections)

**Dependencies**: db, analytics

---

### 6. Storage (Object Storage)
**Image**: `supabase/storage-api:v1.28.0`
**Port**: 5000
**Purpose**: File upload, download, management

**Two Backend Modes**:
1. **File System Mode** (default):
   ```yaml
   STORAGE_BACKEND: file
   FILE_STORAGE_BACKEND_PATH: /var/lib/storage
   ```
2. **S3 Mode** (enabled via docker-compose.s3.yml)

**Key Environment Variables**:
- `DATABASE_URL`: Uses `supabase_storage_admin` user
- `POSTGREST_URL`: PostgREST service address
- `PGRST_JWT_SECRET`: JWT verification
- `FILE_SIZE_LIMIT`: 52428800 (50MB)
- `ENABLE_IMAGE_TRANSFORMATION`: "true"
- `IMGPROXY_URL`: Image processing service address

**Volume Mount**:
```yaml
- ./volumes/storage:/var/lib/storage:z
```

**Dependencies**: db, rest, imgproxy

---

### 7. ImgProxy (Image Processing)
**Image**: `darthsim/imgproxy:v3.8.0`
**Port**: 5001
**Purpose**: Real-time image conversion, scaling, format conversion

**Key Environment Variables**:
- `IMGPROXY_BIND`: ":5001"
- `IMGPROXY_LOCAL_FILESYSTEM_ROOT`: /
- `IMGPROXY_USE_ETAG`: "true"
- `IMGPROXY_ENABLE_WEBP_DETECTION`: Auto WebP detection

**Volume Mount** (shared with Storage):
```yaml
- ./volumes/storage:/var/lib/storage:z
```

---

### 8. Meta (Database Metadata)
**Image**: `supabase/postgres-meta:v0.91.6`
**Port**: 8080
**Purpose**: PostgreSQL metadata API, used by Studio to manage database structure

**Key Environment Variables**:
- `PG_META_DB_USER`: supabase_admin
- `CRYPTO_KEY`: Key for encrypting sensitive information

**Dependencies**: db, analytics

---

### 9. Functions (Edge Functions)
**Image**: `supabase/edge-runtime:v1.69.6`
**Port**: Internal use
**Purpose**: Serverless function runtime (based on Deno)

**Key Environment Variables**:
- `SUPABASE_URL`: Kong gateway address
- `SUPABASE_DB_URL`: Direct database connection
- `VERIFY_JWT`: Whether to verify JWT

**Volume Mount**:
```yaml
- ./volumes/functions:/home/deno/functions:Z
```

**Startup Command**:
```yaml
command: ["start", "--main-service", "/home/deno/functions/main"]
```

**Dependencies**: analytics

---

### 10. Analytics (Log Analysis)
**Image**: `supabase/logflare:1.22.4`
**Port**: 4000
**Purpose**: Log collection, analysis, querying

**Two Backend Modes**:
1. **PostgreSQL Backend** (default):
   ```yaml
   POSTGRES_BACKEND_URL: postgresql://...
   LOGFLARE_FEATURE_FLAG_OVERRIDE: multibackend=true
   ```
2. **BigQuery Backend** (optional)

**Key Environment Variables**:
- `DB_SCHEMA`: _analytics
- `LOGFLARE_SINGLE_TENANT`: true
- `LOGFLARE_SUPABASE_MODE`: true
- `LOGFLARE_MIN_CLUSTER_SIZE`: 1

**Dependencies**: db

---

### 11. DB (PostgreSQL Database)
**Image**: `supabase/postgres:15.8.1.085`
**Port**: 5432
**Purpose**: Core database, includes all Supabase extensions

**Initialization Script Mounts** (in execution order):
```yaml
volumes:
  # Migration scripts (migrations/)
  - ./volumes/db/_supabase.sql:/docker-entrypoint-initdb.d/migrations/97-_supabase.sql
  - ./volumes/db/realtime.sql:/docker-entrypoint-initdb.d/migrations/99-realtime.sql
  - ./volumes/db/logs.sql:/docker-entrypoint-initdb.d/migrations/99-logs.sql
  - ./volumes/db/pooler.sql:/docker-entrypoint-initdb.d/migrations/99-pooler.sql

  # Initialization scripts (init-scripts/)
  - ./volumes/db/webhooks.sql:/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql
  - ./volumes/db/roles.sql:/docker-entrypoint-initdb.d/init-scripts/99-roles.sql
  - ./volumes/db/jwt.sql:/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql

  # Data persistence
  - ./volumes/db/data:/var/lib/postgresql/data:Z

  # pgsodium key persistence
  - db-config:/etc/postgresql-custom
```

**Startup Command**:
```yaml
command:
  - postgres
  - -c
  - config_file=/etc/postgresql/postgresql.conf
  - -c
  - log_min_messages=fatal  # Avoid Realtime polling logs
```

**Dependencies**: vector

---

### 12. Vector (Log Collection)
**Image**: `timberio/vector:0.28.1-alpine`
**Port**: 9001 (health check)
**Purpose**: Collect logs from Docker containers and send to Analytics

**Volume Mount**:
```yaml
- ./volumes/logs/vector.yml:/etc/vector/vector.yml:ro
- ${DOCKER_SOCKET_LOCATION}:/var/run/docker.sock:ro
```

**Configuration File**: `./volumes/logs/vector.yml`

---

### 13. Supavisor (Connection Pool)
**Image**: `supabase/supavisor:2.7.0`
**Ports**:
- 5432 (Postgres protocol)
- 6543 (Transaction mode proxy)
- 4000 (API/health check)

**Purpose**: PostgreSQL connection pool, supports Transaction and Session modes

**Key Environment Variables**:
- `DATABASE_URL`: Connects to `_supabase` database
- `POOLER_TENANT_ID`: Tenant ID
- `POOLER_DEFAULT_POOL_SIZE`: Default connection pool size
- `POOLER_MAX_CLIENT_CONN`: Maximum client connections
- `POOLER_POOL_MODE`: transaction (connection pool mode)
- `CLUSTER_POSTGRES`: true (cluster mode)
- `VAULT_ENC_KEY`: Encryption key

**Startup Command**:
```bash
/app/bin/migrate && \
/app/bin/supavisor eval "$(cat /etc/pooler/pooler.exs)" && \
/app/bin/server
```

**Dependencies**: db, analytics

---

## Service Dependency Graph

```
vector
  ↓
db ←─────────────────────┐
  ↓                      │
analytics                │
  ↓                      │
├─ studio                │
├─ kong                  │
├─ auth ─────────────────┤
├─ rest ─────────────────┤
│   ↓                    │
├─ storage ──→ imgproxy  │
├─ meta ─────────────────┤
├─ functions ────────────┤
└─ supavisor ────────────┘
```

**Key Dependency Chain**:
1. `vector` starts
2. `db` waits for vector health
3. `analytics` waits for db health
4. All other services wait for analytics health
5. `storage` additionally depends on `rest` and `imgproxy`
6. `supavisor` depends on `db` and `analytics`

---

## Database Users and Permissions

From the configuration, we can see Supabase uses multiple dedicated database users:

| User | Purpose | Used By Services |
|------|------|----------|
| `supabase_auth_admin` | Auth management | Auth (GoTrue) |
| `authenticator` | REST API execution | PostgREST |
| `supabase_storage_admin` | Storage management | Storage API |
| `supabase_admin` | General management | Realtime, Meta, Analytics, Supavisor |
| `postgres` | Superuser | Initialization scripts |
| `anon` | Anonymous user role | PostgREST (switched by authenticator) |

---

## Environment Variable Configuration

### Core Keys
```bash
JWT_SECRET              # JWT signing key
JWT_EXPIRY             # JWT expiration time
POSTGRES_PASSWORD      # Database password
SECRET_KEY_BASE        # Phoenix/Elixir application key
PG_META_CRYPTO_KEY     # Meta service encryption key
VAULT_ENC_KEY          # Supavisor encryption key
```

### API Keys
```bash
ANON_KEY               # Public API key (client)
SERVICE_ROLE_KEY       # Service role key (backend)
```

### Database Connection
```bash
POSTGRES_HOST          # Database host
POSTGRES_PORT          # Database port (default 5432)
POSTGRES_DB            # Database name
```

### Kong Gateway
```bash
KONG_HTTP_PORT         # HTTP port (default 8000)
KONG_HTTPS_PORT        # HTTPS port (default 8443)
```

### Authentication Configuration
```bash
SITE_URL               # Frontend application URL
API_EXTERNAL_URL       # API external access URL
SUPABASE_PUBLIC_URL    # Supabase public URL
DISABLE_SIGNUP         # Disable registration
ENABLE_EMAIL_SIGNUP    # Enable email registration
ENABLE_PHONE_SIGNUP    # Enable phone registration
ENABLE_ANONYMOUS_USERS # Enable anonymous users
```

### Log Analysis
```bash
LOGFLARE_PUBLIC_ACCESS_TOKEN
LOGFLARE_PRIVATE_ACCESS_TOKEN
```

---

## Volume Persistence

### Key Data Directories
```yaml
volumes:
  # Database data
  ./volumes/db/data:/var/lib/postgresql/data

  # File storage
  ./volumes/storage:/var/lib/storage

  # Edge Functions code
  ./volumes/functions:/home/deno/functions

  # Configuration files
  ./volumes/api/kong.yml
  ./volumes/logs/vector.yml
  ./volumes/pooler/pooler.exs

  # Named volumes (pgsodium keys)
  db-config:/etc/postgresql-custom
```

---

## Health Checks

All critical services are configured with health checks:

| Service | Check Method | Interval | Timeout | Retries |
|------|----------|------|------|------|
| studio | HTTP API | 5s | 10s | 3 |
| auth | wget /health | 5s | 5s | 3 |
| realtime | curl with auth | 5s | 5s | 3 |
| storage | wget /status | 5s | 5s | 3 |
| imgproxy | imgproxy health | 5s | 5s | 3 |
| db | pg_isready | 5s | 5s | 10 |
| vector | wget /health | 5s | 5s | 3 |
| supavisor | curl /api/health | 10s | 5s | 5 |
| analytics | curl /health | 5s | 5s | 10 |

---

## Extensions and Customization

### Using External Database
Comment out `db` service related configuration and update `POSTGRES_HOST` environment variables.

### Using S3 Storage
```bash
docker compose -f docker-compose.yml -f docker-compose.s3.yml up
```

### Using BigQuery Analytics
Configure in Analytics service:
```yaml
GOOGLE_PROJECT_ID
GOOGLE_PROJECT_NUMBER
# Mount gcloud.json credentials
```

### Development Mode
```bash
docker compose -f docker-compose.yml -f ./dev/docker-compose.dev.yml up
```

---

## Kubernetes Migration Considerations

Comparing Docker Compose and our Kubernetes Operator:

### 1. Database Initialization
**Docker Compose**: Uses `docker-entrypoint-initdb.d` to mount SQL scripts
**Kubernetes**: Requires Init Container or Job to execute initialization

### 2. Configuration Management
**Docker Compose**: Template files + environment variable substitution
**Kubernetes**: ConfigMap + Secret + native `$(VAR)` syntax

### 3. Service Discovery
**Docker Compose**: Service names resolve directly (like `http://auth:9999`)
**Kubernetes**: Requires creating Service resources (`http://<service-name>.<namespace>.svc.cluster.local`)

### 4. Volume Management
**Docker Compose**: Local directory mounts
**Kubernetes**: PersistentVolumeClaim (PVC) + StorageClass

### 5. Health Checks
**Docker Compose**: Built-in healthcheck
**Kubernetes**: livenessProbe + readinessProbe

### 6. Dependency Management
**Docker Compose**: `depends_on` + `condition: service_healthy`
**Kubernetes**: Init Containers + readinessProbe waiting

### 7. Log Collection
**Docker Compose**: Vector monitors Docker socket
**Kubernetes**: Sidecar container or cluster-level DaemonSet

---

## Reference Resources

- Official Documentation: https://supabase.com/docs/guides/self-hosting/docker
- GitHub Repository: https://github.com/supabase/supabase/tree/master/docker
- Image Versions: Regularly updated in docker-compose.yml
