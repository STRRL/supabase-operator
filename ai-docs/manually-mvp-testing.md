# Manual MVP Testing Guide

**Purpose:** Step-by-step guide for manually testing the Supabase Operator MVP to validate core functionality.

**Last Updated:** 2025-10-11

---

## Prerequisites

### Required Tools

- [x] `kubectl` (v1.25+)
- [x] `docker` or compatible container runtime
- [x] `minikube` (v1.30+) or similar Kubernetes cluster
- [x] `make` (for building)
- [x] `curl` (for API testing)

### Environment Setup

```bash
# Start fresh minikube cluster (recommended for clean testing)
minikube delete
minikube start --cpus=4 --memory=8192

# Verify cluster is ready
kubectl cluster-info
kubectl get nodes
```

---

## Test Phases

### Phase 1: Deploy Dependencies

**Objective:** Deploy external dependencies (PostgreSQL and MinIO) that Supabase requires.

#### Step 1.1: Deploy PostgreSQL (supabase/postgres)

```bash
# Deploy PostgreSQL
kubectl apply -f dev-postgres.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n dev-deps --timeout=180s

# Verify PostgreSQL is using supabase/postgres image
kubectl get pod -n dev-deps -l app=postgres -o jsonpath='{.items[0].spec.containers[0].image}'
# Expected: supabase/postgres:15.8.1.085
```

**Success Criteria:**
- PostgreSQL pod is Running
- Image is `supabase/postgres:15.8.1.085` or later
- Pod logs show "database system is ready to accept connections"

**Verification Commands:**

```bash
# Check pod status
kubectl get pods -n dev-deps -l app=postgres

# Check logs for successful initialization
kubectl logs -n dev-deps -l app=postgres --tail=50

# Verify Supabase roles exist
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "\du" | grep -E "supabase_admin|authenticator|anon|authenticated|service_role"
```

**Expected Roles:**

```
authenticator              | Cannot login
anon                       | Cannot login
authenticated              | Cannot login
service_role               | Cannot login, Bypass RLS
supabase_admin             | Superuser, Create role, Create DB, Replication, Bypass RLS
supabase_auth_admin        | Login
supabase_functions_admin   | Login, Create role
supabase_storage_admin     | Login
```

**Common Issues:**

| Issue | Cause | Solution |
|-------|-------|----------|
| Roles missing | Data persisted from previous run | Delete minikube and restart fresh |
| Pod CrashLoopBackOff | Resource constraints | Increase minikube memory/CPU |
| Init skipped | emptyDir has stale data | Delete namespace and recreate |

#### Step 1.2: Deploy MinIO (S3 Compatible Storage)

```bash
# Deploy MinIO
kubectl apply -f dev-minio.yaml

# Wait for MinIO to be ready
kubectl wait --for=condition=ready pod -l app=minio -n dev-deps --timeout=120s

# Wait for minio-setup job to complete
kubectl wait --for=condition=complete job/minio-setup -n dev-deps --timeout=120s
```

**Success Criteria:**
- MinIO pod is Running
- minio-setup job is Completed (creates the bucket)
- Bucket "supabase" exists

**Verification Commands:**

```bash
# Check MinIO pod
kubectl get pods -n dev-deps -l app=minio

# Check setup job completed successfully
kubectl get job minio-setup -n dev-deps
# Expected: COMPLETIONS: 1/1

# Verify bucket creation
kubectl logs -n dev-deps job/minio-setup
# Expected: "Bucket created successfully" or "Bucket 'supabase' already exists"
```

---

### Phase 2: Deploy Supabase Operator

**Objective:** Build and deploy the operator to the Kubernetes cluster.

#### Step 2.1: Build Operator Image

```bash
# Build the operator image
docker build -t supabase-operator:test .

# Load image into minikube
minikube image load supabase-operator:test

# Verify image is loaded
minikube image ls | grep supabase-operator
```

**Success Criteria:**
- Image builds without errors
- Image is loaded into minikube

#### Step 2.2: Deploy Operator

```bash
# Generate and apply manifests
make deploy IMG=supabase-operator:test

# Wait for operator to be ready
kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n supabase-operator-system --timeout=120s
```

**Success Criteria:**
- Operator pod is Running
- CRD is installed
- RBAC roles are created

**Verification Commands:**

```bash
# Check operator pod
kubectl get pods -n supabase-operator-system

# Check operator logs (should show leader election success)
kubectl logs -n supabase-operator-system deployment/supabase-operator-controller-manager --tail=20

# Verify CRD is installed
kubectl get crd supabaseprojects.supabase.strrl.dev
```

---

### Phase 3: Deploy SupabaseProject

**Objective:** Create a SupabaseProject CR and verify all components are deployed.

#### Step 3.1: Create Secrets

```bash
# Create namespace and secrets
kubectl apply -f dev-secrets.yaml

# Verify secrets exist
kubectl get secrets -n test-supabase
```

**Expected Secrets:**
- `test-db-creds` - PostgreSQL credentials
- `test-storage-creds` - MinIO credentials

**Verify Secret Contents:**

```bash
# Check database secret has required keys
kubectl get secret test-db-creds -n test-supabase -o jsonpath='{.data}' | jq 'keys'
# Expected: ["database", "host", "password", "port", "username"]

# Check storage secret has required keys
kubectl get secret test-storage-creds -n test-supabase -o jsonpath='{.data}' | jq 'keys'
# Expected: ["accessKeyId", "bucket", "endpoint", "region", "secretAccessKey"]
```

#### Step 3.2: Create SupabaseProject

```bash
# Apply SupabaseProject CR
kubectl apply -f test-project.yaml

# Watch the reconciliation process
kubectl get supabaseproject test-project -n test-supabase -w
```

**Expected Phase Progression:**

```
PHASE                       DURATION    STATUS
Pending                     ~5s         Initial state
ValidatingDependencies      ~3s         Checking database/storage secrets
DeployingSecrets            ~2s         Generating JWT secrets
InitializingDatabase        ~30-60s     Running database init Job
DeployingComponents         ~30-60s     Creating deployments/services
Running                     Steady      All components ready
```

#### Step 3.3: Monitor Database Initialization

**This is a critical step - database initialization must succeed for services to work.**

```bash
# Wait for database init job to complete
kubectl wait --for=condition=complete job/test-project-db-init -n test-supabase --timeout=300s

# Check job status
kubectl get job test-project-db-init -n test-supabase

# View initialization logs
kubectl logs -n test-supabase job/test-project-db-init
```

**Expected Log Output:**

```
Starting Supabase database initialization...
Database: postgres at postgres.dev-deps.svc.cluster.local:5432

Executing: 00-initial-schema.sql...
CREATE DATABASE
  ✓ 00-initial-schema.sql completed

Executing: 01-roles.sql...
ALTER ROLE
[... may show some errors for roles that don't need password changes ...]
  ✓ 01-roles.sql completed

Executing: 02-jwt.sql...
ALTER DATABASE
ALTER DATABASE
  ✓ 02-jwt.sql completed

Executing: 03-logs.sql...
You are now connected to database "_supabase" as user "postgres".
CREATE SCHEMA
  ✓ 03-logs.sql completed

Executing: 04-webhooks.sql...
[... may show some errors if pg_net extension not available ...]
  ✓ 04-webhooks.sql completed

Executing: 05-realtime.sql...
CREATE SCHEMA
  ✓ 05-realtime.sql completed

Executing: 06-pooler.sql...
CREATE SCHEMA
  ✓ 06-pooler.sql completed

Database initialization complete!
```

**Success Criteria:**
- Job status is "Complete" (1/1)
- All 7 SQL files executed
- No fatal errors (some warnings are OK)
- Key schemas created: `_realtime`, `_analytics`, `pgbouncer`

**Verification:**

```bash
# Check schemas were created
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "\dn"
# Expected schemas: public, _realtime, _analytics, pgbouncer, storage

# Check JWT configuration was written
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "SELECT name, setting FROM pg_settings WHERE name LIKE 'app.settings.%';"
# Expected: app.settings.jwt_secret and app.settings.jwt_exp
```

**Common Issues:**

| Issue | Cause | Solution |
|-------|-------|----------|
| Job fails with "schema does not exist" | Roles not created during PostgreSQL init | Restart with fresh PostgreSQL |
| Job fails with "secret not found" | JWT secret creation order wrong | Check operator logs, verify fix is deployed |
| Job stuck in pending | Image pull issues | Check: `kubectl describe job test-project-db-init -n test-supabase` |

#### Step 3.4: Verify Component Deployments

```bash
# Check all pods
kubectl get pods -n test-supabase

# Check JWT secrets were generated
kubectl get secret test-project-jwt -n test-supabase -o jsonpath='{.data}' | jq 'keys'
# Expected: ["anon-key", "jwt-secret", "service-role-key"]
```

**Expected Pods:**

| Pod | Status | Restarts | Notes |
|-----|--------|----------|-------|
| test-project-kong-* | Running | 0 | API Gateway |
| test-project-auth-* | Running | 0-2 | May restart while waiting for DB init |
| test-project-postgrest-* | Running | 0 | REST API |
| test-project-realtime-* | Running | 0-2 | May restart while waiting for DB init |
| test-project-storage-* | Running | 0-2 | May restart while waiting for DB init |
| test-project-meta-* | Running | 0 | Database management |
| test-project-db-init-* | Completed | 0 | Init job |

**Success Criteria:**
- All service pods are Running or Completed
- No CrashLoopBackOff (except during initial DB setup)
- JWT secret exists with 3 keys

**Detailed Pod Checks:**

```bash
# Check each component's readiness
for component in kong auth postgrest realtime storage meta; do
  echo "=== $component ==="
  kubectl get pods -n test-supabase -l "app.kubernetes.io/name=$component"
  kubectl logs -n test-supabase deployment/test-project-$component --tail=5
  echo ""
done
```

---

### Phase 4: Functional Testing

**Objective:** Verify Supabase services are functional through API calls.

#### Step 4.1: Get API Keys

```bash
# Extract API keys
export ANON_KEY=$(kubectl get secret test-project-jwt -n test-supabase -o jsonpath='{.data.anon-key}' | base64 -d)
export SERVICE_ROLE_KEY=$(kubectl get secret test-project-jwt -n test-supabase -o jsonpath='{.data.service-role-key}' | base64 -d)

echo "ANON_KEY: $ANON_KEY"
echo "SERVICE_ROLE_KEY: $SERVICE_ROLE_KEY"
```

#### Step 4.2: Setup Port Forwarding

```bash
# Forward Kong (API Gateway) port
kubectl port-forward -n test-supabase service/test-project-kong 8000:8000 &
export PF_PID=$!

# Wait for port forward to be ready
sleep 3
```

#### Step 4.3: Test Health Endpoints

```bash
# Test Kong health
curl -i http://localhost:8000/

# Expected: HTTP 404 (Kong is running, no route matched)
# Should NOT get connection refused
```

#### Step 4.4: Test Auth Service (GoTrue)

```bash
# Test Auth health endpoint
curl -i http://localhost:8000/auth/v1/health

# Expected: HTTP 200
# Body: {"version":"...", "name":"GoTrue"}
```

**Test User Signup:**

```bash
# Sign up a test user
curl -X POST http://localhost:8000/auth/v1/signup \
  -H "apikey: $ANON_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "test123456"
  }'

# Expected: HTTP 200
# Response includes: {"id":"...", "email":"test@example.com", ...}
```

**Success Criteria:**
- Health endpoint returns 200
- User signup returns 200 with user object
- User is created in `auth.users` table

**Verification:**

```bash
# Check user was created in database
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "SELECT id, email, created_at FROM auth.users;"
```

#### Step 4.5: Test PostgREST (REST API)

```bash
# Test PostgREST endpoint
curl -i http://localhost:8000/rest/v1/

# Expected: HTTP 200 with OpenAPI spec
```

**Create a test table and query it:**

```bash
# Create a test table using service role key
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "
CREATE TABLE IF NOT EXISTS public.test_table (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE public.test_table ENABLE ROW LEVEL SECURITY;

CREATE POLICY \"Allow all\" ON public.test_table FOR ALL USING (true);
"

# Insert test data via REST API
curl -X POST http://localhost:8000/rest/v1/test_table \
  -H "apikey: $SERVICE_ROLE_KEY" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{"name": "Test Item"}'

# Expected: HTTP 201 with created row

# Query the data
curl http://localhost:8000/rest/v1/test_table \
  -H "apikey: $ANON_KEY" \
  -H "Authorization: Bearer $ANON_KEY"

# Expected: HTTP 200 with JSON array containing the test item
```

**Success Criteria:**
- PostgREST root endpoint returns 200
- Can create data via POST
- Can query data via GET
- RLS policies are enforced

#### Step 4.6: Test Storage API

```bash
# Test Storage health
curl -i http://localhost:8000/storage/v1/

# Create a test bucket
curl -X POST http://localhost:8000/storage/v1/bucket \
  -H "apikey: $SERVICE_ROLE_KEY" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-bucket",
    "name": "test-bucket",
    "public": false
  }'

# Expected: HTTP 200 with bucket details

# List buckets
curl http://localhost:8000/storage/v1/bucket \
  -H "apikey: $SERVICE_ROLE_KEY" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY"

# Expected: HTTP 200 with array including "test-bucket"
```

**Success Criteria:**
- Storage root endpoint returns 200
- Can create buckets
- Can list buckets
- MinIO integration working

#### Step 4.7: Test Realtime Service

```bash
# Test Realtime health
curl -i http://localhost:8000/realtime/v1/

# Expected: HTTP 403 or 404 (WebSocket endpoint, HTTP not supported)
# As long as it's not connection refused, Realtime is running
```

**For full WebSocket testing, use a WebSocket client:**

```bash
# Install wscat if needed: npm install -g wscat

# Connect to Realtime (test connection only)
wscat -c "ws://localhost:8000/realtime/v1/websocket?apikey=$ANON_KEY&vsn=1.0.0"

# Expected: WebSocket connection established
# Type: {"topic":"test","event":"phx_join","payload":{},"ref":"1"}
# Should receive a response
```

#### Step 4.8: Test Meta (Database Management)

```bash
# Test Meta health
curl -i http://localhost:8000/pg/ \
  -H "apikey: $SERVICE_ROLE_KEY" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY"

# Expected: HTTP 200 with API information

# List tables
curl http://localhost:8000/pg/tables \
  -H "apikey: $SERVICE_ROLE_KEY" \
  -H "Authorization: Bearer $SERVICE_ROLE_KEY"

# Expected: HTTP 200 with array of tables
```

**Success Criteria:**
- Meta API returns 200
- Can list tables
- Database introspection working

#### Step 4.9: Cleanup Port Forward

```bash
# Kill port forward process
kill $PF_PID
```

---

## Phase 5: Status and Observability

**Objective:** Verify status reporting and monitoring capabilities.

### Step 5.1: Check SupabaseProject Status

```bash
# Get detailed status
kubectl get supabaseproject test-project -n test-supabase -o yaml | grep -A 50 status:
```

**Expected Status Fields:**

```yaml
status:
  phase: Running
  message: All components are running
  observedGeneration: 1
  lastReconcileTime: "2025-10-11T..."

  components:
    auth:
      phase: Running
      ready: true
      version: "supabase/gotrue:v2.177.0"
      readyReplicas: 1
      replicas: 1

    kong:
      phase: Running
      ready: true
      version: "kong:2.8.1"
      readyReplicas: 1
      replicas: 1

    # ... (similar for postgrest, realtime, storage, meta)

  conditions:
  - type: Ready
    status: "True"
    reason: AllComponentsReady
  - type: Progressing
    status: "False"
    reason: ReconciliationComplete

  dependencies:
    postgresql:
      connected: true
      lastConnectedTime: "2025-10-11T..."

    s3:
      connected: true
      lastConnectedTime: "2025-10-11T..."
```

**Success Criteria:**
- phase: Running
- All components have ready: true
- Ready condition is True
- No error conditions

### Step 5.2: Check Operator Logs

```bash
# Check operator logs for errors
kubectl logs -n supabase-operator-system deployment/supabase-operator-controller-manager --tail=100 | grep -i error

# Should have no ERROR or WARN messages (except initial reconciliation)
```

### Step 5.3: Check Resource Usage

```bash
# Check CPU and memory usage
kubectl top pods -n test-supabase

# Typical usage (may vary):
# kong: 10-50m CPU, 50-100Mi memory
# auth: 10-30m CPU, 50-100Mi memory
# postgrest: 10-30m CPU, 50-100Mi memory
# realtime: 20-50m CPU, 100-200Mi memory
# storage: 10-30m CPU, 50-100Mi memory
# meta: 10-20m CPU, 50-100Mi memory
```

---

## Overall Success Criteria

### MVP is Successful If:

1. **Infrastructure:**
   - [ ] PostgreSQL with Supabase roles is running
   - [ ] MinIO is running with "supabase" bucket created
   - [ ] Operator is deployed and reconciling

2. **Database Initialization:**
   - [ ] Database init Job completed successfully
   - [ ] All schemas created (_realtime, _analytics, pgbouncer)
   - [ ] JWT configuration written to database

3. **Component Deployment:**
   - [ ] All 6 service pods are Running
   - [ ] JWT secrets generated
   - [ ] All services have 0-2 restarts max

4. **Functional Tests:**
   - [ ] Auth: User signup works
   - [ ] PostgREST: Can CRUD data
   - [ ] Storage: Can create/list buckets
   - [ ] Realtime: Service is accessible
   - [ ] Meta: Can introspect database

5. **Status Reporting:**
   - [ ] SupabaseProject phase is "Running"
   - [ ] All components report as ready
   - [ ] Conditions show healthy state

---

## Cleanup

```bash
# Delete SupabaseProject
kubectl delete -f test-project.yaml

# Wait for finalizers to complete
kubectl wait --for=delete supabaseproject/test-project -n test-supabase --timeout=60s

# Delete namespaces
kubectl delete namespace test-supabase dev-deps

# Stop operator
kubectl delete -k config/default

# (Optional) Delete minikube cluster
minikube delete
```

---

## Troubleshooting Guide

### Issue: Database Init Job Fails

**Symptoms:**
- Job status shows "Failed"
- Pods show "Error" or "CrashLoopBackOff"

**Diagnosis:**

```bash
# Check job logs
kubectl logs -n test-supabase job/test-project-db-init

# Check pod events
kubectl describe job test-project-db-init -n test-supabase
```

**Common Causes:**

1. **Missing JWT Secret**
   - Error: `secret "test-project-jwt" not found`
   - Solution: Verify operator created JWT secret before Job

2. **PostgreSQL Roles Missing**
   - Error: `role "authenticator" does not exist`
   - Solution: Restart with fresh PostgreSQL (delete minikube)

3. **Connection Issues**
   - Error: `could not connect to server`
   - Solution: Verify PostgreSQL service is accessible

### Issue: Auth Service Crashes

**Symptoms:**
- auth pod in CrashLoopBackOff
- Logs show "schema auth does not exist"

**Diagnosis:**

```bash
# Check auth logs
kubectl logs -n test-supabase deployment/test-project-auth

# Check if auth schema exists
kubectl exec -n dev-deps deployment/postgres -- psql -U postgres -c "\dn" | grep auth
```

**Solution:**

If schema doesn't exist, database init failed. Check database init job logs.

### Issue: Storage API Can't Connect to MinIO

**Symptoms:**
- storage pod logs show S3 connection errors

**Diagnosis:**

```bash
# Check storage logs
kubectl logs -n test-supabase deployment/test-project-storage

# Verify MinIO is accessible
kubectl exec -n test-supabase deployment/test-project-storage -- curl -v http://minio.dev-deps.svc.cluster.local:9000/minio/health/live
```

**Solution:**

Verify MinIO service exists and bucket was created.

---

## Testing Checklist

Print and use this checklist during manual testing:

```
[ ] Phase 1: Dependencies
    [ ] PostgreSQL deployed and ready
    [ ] Supabase roles verified
    [ ] MinIO deployed and ready
    [ ] Bucket "supabase" created

[ ] Phase 2: Operator
    [ ] Image built and loaded
    [ ] Operator deployed and ready
    [ ] CRD installed

[ ] Phase 3: SupabaseProject
    [ ] Secrets created
    [ ] SupabaseProject CR applied
    [ ] Database init Job completed (CRITICAL)
    [ ] All pods Running
    [ ] JWT secret generated

[ ] Phase 4: Functional Tests
    [ ] Auth signup works
    [ ] PostgREST CRUD works
    [ ] Storage bucket operations work
    [ ] Realtime accessible
    [ ] Meta API works

[ ] Phase 5: Status
    [ ] Phase is "Running"
    [ ] All components ready
    [ ] Conditions healthy

[ ] MVP SUCCESS
```

---

## Next Steps After MVP Success

1. **Performance Testing**
   - Load test each service
   - Measure reconciliation time
   - Test scaling

2. **Reliability Testing**
   - Test pod restarts
   - Test node failures
   - Test operator restarts

3. **Security Testing**
   - Test RLS policies
   - Test JWT validation
   - Test service-to-service auth

4. **Documentation**
   - Update README with tested examples
   - Create user guides
   - Document known limitations

---

## References

- [Supabase Self-Hosting Docs](https://supabase.com/docs/guides/self-hosting)
- [Database Initialization Investigation](./database-initialization-investigation.md)
- [Operator Reconciliation Logic](../internal/controller/supabaseproject_controller.go)
