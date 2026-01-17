# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kubernetes operator for deploying self-hosted Supabase instances using Kubebuilder. A single `SupabaseProject` CRD deploys and manages Kong, Auth/GoTrue, PostgREST, Realtime, Storage API, Meta, and Studio components.

## Build & Development Commands

```bash
make build              # Compile binary to bin/supabase-operator
make run                # Run controller locally (requires k8s cluster)
make check              # Run all code generation, formatting, vetting, and linting
make manifests          # Generate CRDs into helm/supabase-operator/crds/
make generate           # Generate DeepCopy code for API types
make image              # Build Docker image (ghcr.io/strrl/supabase-operator:<commit>)
```

## Testing

```bash
make test               # Unit tests with envtest (uses Ginkgo/Gomega)
make test-e2e           # E2E tests on Minikube with Chrome screenshot capture
```

Run a single test file:
```bash
go test ./internal/controller/... -v -run TestReconcile
```

Run specific Ginkgo specs:
```bash
ginkgo -v --focus="should create kong" ./internal/controller/...
```

## Linting

```bash
make lint               # Run golangci-lint
make lint-fix           # Auto-fix linting issues
```

## Code Architecture

### CRD and API Types
- `api/v1alpha1/supabaseproject_types.go` - SupabaseProject spec and status definitions
- Spec requires `projectId`, `database.secretRef`, and `storage.secretRef`
- Status tracks phase (Pending → ValidatingDependencies → DeployingSecrets → DeployingComponents → Running → Failed) and per-component status

### Controller Logic
- `internal/controller/supabaseproject_controller.go` - Main reconciler
- `internal/controller/reconciler/component.go` - Component reconciliation logic
- Controller watches SupabaseProject and manages Deployments, Services, Secrets, ConfigMaps, Jobs

### Component Builders
Each Supabase component has a dedicated builder in `internal/component/`:
- `kong.go`, `auth.go`, `postgrest.go`, `realtime.go`, `storage_api.go`, `meta.go`, `studio.go`
- `database_init.go` - Creates Job for PostgreSQL initialization

### Database Migrations
- `internal/database/migrations/sql/` - Embedded SQL scripts (00-initial-schema through 06-pooler)
- Files embedded in binary via `//go:embed`
- Sync with upstream Supabase: `./hack/sync-migrations.sh [tag]`

### Secrets Management
- `internal/secrets/` - JWT and API key generation
- Operator generates `<project>-jwt` secret with anon-key, service-role-key

## Key Patterns

- **Idempotent reconciliation**: Controllers repeatedly reconcile desired vs actual state
- **Finalizers**: Clean up resources when SupabaseProject is deleted
- **Component builder pattern**: Modular resource creation per component file
- **Status conditions**: Kubernetes-standard conditions for granular status tracking

## External Dependencies

User must provide:
- PostgreSQL database (referenced via secret with host, port, database, username, password)
- S3-compatible storage (referenced via secret with endpoint, region, bucket, accessKeyId, secretAccessKey)

## Helm Chart

Located in `helm/supabase-operator/`. Install:
```bash
helm upgrade --install supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system --create-namespace
```

## Code Style

- No end-of-line comments
- No Chinese in code comments
