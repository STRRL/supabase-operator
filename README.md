# Supabase Operator

A Kubernetes operator for deploying and managing self-hosted Supabase instances.

## Overview

The Supabase Operator enables you to deploy complete Supabase instances on Kubernetes using a single Custom Resource Definition (CRD). It manages all Supabase components including Kong API Gateway, GoTrue authentication, PostgREST, Realtime, Storage API, and Meta.

**Features:**
- 🚀 Deploy full Supabase stack with a single manifest
- 🔐 Automatic JWT secret generation
- 📊 Granular status reporting with per-component tracking
- 🔄 Rolling updates with health checks
- 🏗️ Multi-tenant: Multiple Supabase projects per namespace
- ☸️ Kubernetes-native with standard types and patterns

## Architecture

The operator manages:
- **Kong**: API Gateway (v2.8.1)
- **Auth**: GoTrue authentication service (v2.177.0)
- **PostgREST**: Automatic REST API (v12.2.12)
- **Realtime**: WebSocket server (v2.34.47)
- **Storage API**: File storage service (v1.25.7)
- **Meta**: PostgreSQL metadata service (v0.91.0)
- **Studio**: Supabase management UI (2025.10.01-sha-8460121)

**External dependencies** (user-provided):
- PostgreSQL database
- S3-compatible storage

## Installation

### Prerequisites
- Kubernetes 1.33+
- kubectl configured
- External PostgreSQL database
- S3-compatible storage (MinIO, AWS S3, etc.)

### Deploy the Operator

```bash
helm upgrade --install supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system \
  --create-namespace \
  --wait
```

Or install from source:

```bash
git clone https://github.com/strrl/supabase-operator
cd supabase-operator
helm upgrade --install supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system \
  --create-namespace \
  --wait
```

### Build Container Image

The repo provides an opinionated build that mirrors the pattern used in [`STRRL/cloudflare-tunnel-ingress-controller`](https://github.com/STRRL/cloudflare-tunnel-ingress-controller). Build it with BuildKit enabled so module downloads are cached automatically:

```bash
make image
```

The helper tags the result as `ghcr.io/strrl/supabase-operator:<commit>`, `<commit>-linux-amd64`, and `latest`. Provide an extra tag if needed:

```bash
IMAGE_TAG=v0.1.0 make image
```

### Deploy with Helm

A self-contained Helm chart lives in `helm/supabase-operator`, including the CRD. Render or install it directly:

```bash
# Render the manifests for review
helm template supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system

# Or install into a cluster
helm install supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system \
  --create-namespace \
  --set image.repository=ghcr.io/strrl/supabase-operator \
  --set image.tag=$(git rev-parse --short HEAD)
```

Pass additional overrides (for example `.Values.extraArgs` or `.Values.resources`) to tailor the deployment.

## Quick Start

### 1. Create Database Secret

```bash
kubectl create secret generic postgres-config \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=database=supabase \
  --from-literal=username=postgres \
  --from-literal=password=your-secure-password
```

### 2. Create Storage Secret

```bash
kubectl create secret generic s3-config \
  --from-literal=endpoint=https://s3.example.com \
  --from-literal=region=us-east-1 \
  --from-literal=bucket=supabase-storage \
  --from-literal=accessKeyId=your-access-key \
  --from-literal=secretAccessKey=your-secret-key
```

> **Note:** Use camelCase for secret keys: `accessKeyId` and `secretAccessKey` (not kebab-case)

### 3. Create Studio Dashboard Secret

```bash
kubectl create secret generic studio-dashboard-creds \
  --from-literal=username=supabase \
  --from-literal=password='choose-a-strong-password'
```

> Change these credentials before exposing Kong publicly.

### 4. Deploy SupabaseProject

```yaml
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: my-supabase
  namespace: default
spec:
  projectId: my-supabase-project

  database:
    secretRef:
      name: postgres-config
    sslMode: require
    maxConnections: 50

  storage:
    secretRef:
      name: s3-config
    forcePathStyle: true
  studio:
    dashboardBasicAuthSecretRef:
      name: studio-dashboard-creds
```

Apply the manifest:

```bash
kubectl apply -f my-supabase.yaml
```

### 5. Check Status

```bash
kubectl get supabaseproject my-supabase -o yaml
```

Check component status:

```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.components}'
```

### 6. Access Services

The operator creates services for each component:

```bash
kubectl get services -l app.kubernetes.io/part-of=supabase
```

Access Kong API Gateway:

```bash
kubectl port-forward svc/my-supabase-kong 8000:8000
```

Requests to `http://localhost:8000/` will answer `401 Unauthorized` until you supply the username/password stored in `studio-dashboard-creds`.

### 7. Retrieve Supabase API Keys

The operator generates API keys and stores them in a secret named `<project>-jwt` within the same namespace as your `SupabaseProject`.

```bash
# Get the public ANON key
ANON_KEY=$(kubectl get secret my-supabase-jwt \
  -o jsonpath='{.data.anon-key}' | base64 -d)

# Get the Service Role key
SERVICE_ROLE_KEY=$(kubectl get secret my-supabase-jwt \
  -o jsonpath='{.data.service-role-key}' | base64 -d)

# Optional: discover the API endpoint
API_URL=$(kubectl get supabaseproject my-supabase \
  -o jsonpath='{.status.endpoints.api}')
```

Use `$ANON_KEY` for client-side requests and `$SERVICE_ROLE_KEY` for trusted backend workflows.

### 8. Connect to Your Database

Supabase components use the external PostgreSQL database you referenced via `postgres-config`. You can reuse the same credentials to connect with tools like `psql`.

```bash
POSTGRES_HOST=$(kubectl get secret postgres-config -o jsonpath='{.data.host}' | base64 -d)
POSTGRES_PORT=$(kubectl get secret postgres-config -o jsonpath='{.data.port}' | base64 -d)
POSTGRES_DB=$(kubectl get secret postgres-config -o jsonpath='{.data.database}' | base64 -d)
POSTGRES_USER=$(kubectl get secret postgres-config -o jsonpath='{.data.username}' | base64 -d)
POSTGRES_PASSWORD=$(kubectl get secret postgres-config -o jsonpath='{.data.password}' | base64 -d)

psql "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"
```

If the database is only reachable inside the cluster (for example, over a private network), run `kubectl run` with a temporary pod or establish a VPN/tunnel that matches your deployment topology.

## Configuration

### Component Customization

Override default images and resources:

```yaml
spec:
  kong:
    image: kong:3.0.0
    replicas: 2
    resources:
      limits:
        memory: "4Gi"
        cpu: "1000m"
      requests:
        memory: "2Gi"
        cpu: "500m"
    extraEnv:
      - name: KONG_LOG_LEVEL
        value: "debug"
```

### Default Resource Limits

| Component | Memory Limit | CPU Limit | Memory Request | CPU Request |
|-----------|--------------|-----------|----------------|-------------|
| Kong      | 2.5Gi        | 500m      | 1Gi            | 250m        |
| Auth      | 128Mi        | 100m      | 64Mi           | 50m         |
| PostgREST | 256Mi        | 200m      | 128Mi          | 100m        |
| Realtime  | 256Mi        | 200m      | 128Mi          | 100m        |
| Storage   | 128Mi        | 100m      | 64Mi           | 50m         |
| Meta      | 128Mi        | 100m      | 64Mi           | 50m         |

## Database Initialization

On first deployment, the operator automatically initializes your PostgreSQL database with:

**Extensions:**
- `pgcrypto` - Cryptographic functions
- `uuid-ossp` - UUID generation
- `pg_stat_statements` - Query statistics

**Schemas:**
- `auth` - Authentication data
- `storage` - File metadata
- `realtime` - Real-time subscriptions

**Roles:**
- `authenticator` - API request authenticator role
- `anon` - Anonymous access role
- `service_role` - Service-level access role with RLS bypass

All initialization operations are idempotent and safe to re-run.

## Status Tracking

The operator provides granular status reporting:

```yaml
status:
  phase: Running
  message: "All components running"
  conditions:
    - type: Ready
      status: "True"
      reason: AllComponentsReady
    - type: Progressing
      status: "False"
      reason: ReconciliationComplete
  components:
    kong:
      phase: Running
      ready: true
      version: kong:2.8.1
      replicas: 1
      readyReplicas: 1
    auth:
      phase: Running
      ready: true
      version: supabase/gotrue:v2.177.0
```

**Phases:**
- `Pending`: Initial state
- `ValidatingDependencies`: Checking PostgreSQL and S3
- `DeployingSecrets`: Generating JWT secrets
- `DeployingComponents`: Creating deployments
- `Running`: All components healthy
- `Failed`: Reconciliation error

## Monitoring

The operator exposes Prometheus metrics at `:8443/metrics`:

```bash
kubectl port-forward -n supabase-operator-system \
  svc/supabase-operator-controller-manager-metrics-service 8443:8443
```

**Key metrics:**
- `controller_runtime_reconcile_total` - Total reconciliations
- `controller_runtime_reconcile_errors_total` - Reconciliation errors
- `controller_runtime_reconcile_time_seconds` - Reconciliation duration
- `workqueue_depth` - Controller work queue depth
- `workqueue_adds_total` - Items added to work queue

### ServiceMonitor

If you run Prometheus Operator, create a `ServiceMonitor` that points at the metrics service
(`app.kubernetes.io/name=supabase-operator`, port `8443`). The Helm chart does not yet ship one;
add your own manifest or a Helm subchart to scrape the metrics endpoint.

## Troubleshooting

### SupabaseProject stuck in "ValidatingDependencies"

**Check secrets exist:**
```bash
kubectl get secret postgres-config s3-config
```

**Verify secret keys:**
```bash
kubectl get secret postgres-config -o jsonpath='{.data}' | jq
```

Required database secret keys: `host`, `port`, `database`, `username`, `password`
Required storage secret keys: `endpoint`, `region`, `bucket`, `accessKeyId`, `secretAccessKey`

### SupabaseProject in "Failed" phase

**Check status message:**
```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.message}'
```

**Check conditions:**
```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.conditions}' | jq
```

**View controller logs:**
```bash
kubectl logs -n supabase-operator-system \
  -l control-plane=controller-manager \
  --tail=100
```

### Database initialization fails

**Check database connectivity:**
```bash
kubectl run -it --rm debug --image=postgres:16 --restart=Never -- \
  psql -h postgres.example.com -U postgres -d supabase
```

**Verify database permissions:**
The database user must have `CREATEDB` or superuser privileges to create extensions and schemas.

### Components not starting

**Check pod status:**
```bash
kubectl get pods -l app.kubernetes.io/part-of=supabase
```

**Check pod logs:**
```bash
kubectl logs my-supabase-kong-xxx
```

**Verify resource limits:**
Ensure your cluster has sufficient resources for all components.

## Development

### Prerequisites
- Go 1.22+
- Kubebuilder 4.0+
- Docker

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Locally

```bash
kubectl apply -f helm/supabase-operator/crds/supabase.strrl.dev_supabaseprojects.yaml  # Install CRD
make run      # Run controller locally
```

### Generate Manifests

```bash
make manifests  # Generate CRD and RBAC
make generate   # Generate deepcopy code
```

## Documentation

- **[Architecture Guide](docs/architecture.md)**: Detailed architecture documentation covering system design, controller patterns, component deployment, status management, and design decisions
- **[API Reference](docs/api-reference.md)**: Complete API reference for the SupabaseProject CRD with field descriptions, examples, and validation rules
- **[Database Initialization](docs/database-initialization.md)**: PostgreSQL setup requirements and initialization details
- **[Quick Start](docs/quick-start.md)**: Getting started guide with step-by-step instructions
- **[Future Considerations](specs/001-selfhost-supabase-operator/future-considerations.md)**: Deferred features and architectural flexibility

## Contributing

Contributions welcome! Please read the [design documents](specs/001-selfhost-supabase-operator/) for context.

## License

MIT

## Roadmap

- [x] v1alpha1: Core operator with basic deployment
- [ ] v1beta1: Advanced features (HA, backup, monitoring)
- [ ] v1: Production-ready with stability guarantees

See [future-considerations.md](specs/001-selfhost-supabase-operator/future-considerations.md) for planned features.

## End-to-End Tests

The e2e suite spins up a temporary [Minikube](https://minikube.sigs.k8s.io/) profile, deploys the operator, and exercises a SupabaseProject end-to-end (including capturing a Kong Studio screenshot via headless Chrome).

- Install Minikube (`minikube version` should succeed) and ensure Docker is available.
- Install Google Chrome or Chromium locally, or set `E2E_CHROME_PATH` to a compatible executable. If Chrome is missing, the screenshot spec is skipped.
- Run:

  ```bash
  MINIKUBE_START_ARGS="--driver=docker --cpus=4 --memory=8192 --wait=all" make test-e2e
  ```

  This command creates the `supabase-operator-test-e2e` Minikube profile, runs `go test -tags=e2e`, saves screenshots to `.artifacts/screenshots/`, and tears the profile down afterwards.
