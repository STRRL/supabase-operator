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
kubectl apply -f https://raw.githubusercontent.com/strrl/supabase-operator/main/config/install.yaml
```

Or install from source:

```bash
git clone https://github.com/strrl/supabase-operator
cd supabase-operator
make install
make deploy
```

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

### 3. Deploy SupabaseProject

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
```

Apply the manifest:

```bash
kubectl apply -f my-supabase.yaml
```

### 4. Check Status

```bash
kubectl get supabaseproject my-supabase -o yaml
```

Check component status:

```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.components}'
```

### 5. Access Services

The operator creates services for each component:

```bash
kubectl get services -l app.kubernetes.io/part-of=supabase
```

Access Kong API Gateway:

```bash
kubectl port-forward svc/my-supabase-kong 8000:8000
```

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
make install  # Install CRDs
make run      # Run controller locally
```

### Generate Manifests

```bash
make manifests  # Generate CRD and RBAC
make generate   # Generate deepcopy code
```

## Architecture Decisions

See [future-considerations.md](specs/001-selfhost-supabase-operator/future-considerations.md) for deferred features and architectural flexibility.

## Contributing

Contributions welcome! Please read the [design documents](specs/001-selfhost-supabase-operator/) for context.

## License

MIT

## Roadmap

- [x] v1alpha1: Core operator with basic deployment
- [ ] v1beta1: Advanced features (HA, backup, monitoring)
- [ ] v1: Production-ready with stability guarantees

See [future-considerations.md](specs/001-selfhost-supabase-operator/future-considerations.md) for planned features.
