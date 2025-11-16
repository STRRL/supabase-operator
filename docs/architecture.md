# Architecture Documentation

## Overview

The Supabase Operator is a Kubernetes controller that automates the deployment and lifecycle management of self-hosted Supabase instances. Built on the Kubebuilder framework, it follows the Kubernetes operator pattern to provide declarative, API-driven management of complex Supabase deployments.

**Core Responsibilities:**
- Deploy and manage all Supabase components (Kong, Auth, PostgREST, Realtime, Storage, Meta, Studio)
- Integrate with external dependencies (PostgreSQL, S3-compatible storage)
- Provide granular status reporting and observability
- Handle configuration updates with rolling deployments
- Manage secrets and JWT token generation

## System Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                           │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │            Supabase Operator Controller                  │   │
│  │  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │   │
│  │  │ Reconciler  │  │ Status Mgr   │  │ Component    │  │   │
│  │  │   Loop      │→│              │→│   Builder    │  │   │
│  │  └─────────────┘  └──────────────┘  └──────────────┘  │   │
│  │         ↓                  ↓                 ↓          │   │
│  │  ┌──────────────────────────────────────────────────┐ │   │
│  │  │       Kubernetes API (Watch/Create/Update)       │ │   │
│  │  └──────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
│                               ↓                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              SupabaseProject Resources                    │  │
│  │                                                            │  │
│  │  ┌────────┐ ┌────────┐ ┌──────────┐ ┌─────────┐        │  │
│  │  │  Kong  │ │  Auth  │ │PostgREST │ │Realtime │ ...    │  │
│  │  │  Pod   │ │  Pod   │ │   Pod    │ │  Pod    │        │  │
│  │  └────────┘ └────────┘ └──────────┘ └─────────┘        │  │
│  │                                                            │  │
│  │  ┌─────────┐ ┌─────────┐ ┌──────────────┐              │  │
│  │  │Services │ │ConfigMap│ │JWT Secrets   │              │  │
│  │  └─────────┘ └─────────┘ └──────────────┘              │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
                               ↓
                   ┌──────────────────────┐
                   │  External Dependencies│
                   │  - PostgreSQL         │
                   │  - S3 Storage         │
                   └──────────────────────┘
```

### Component Hierarchy

**API Layer** (`api/v1alpha1/`)
- `SupabaseProject`: Primary CRD defining the desired state
- Validation webhooks for admission control
- Type definitions and kubebuilder markers

**Controller Layer** (`internal/controller/`)
- Main reconciliation loop
- Phase-based deployment orchestration
- Dependency validation
- Finalizer handling for cleanup

**Component Layer** (`internal/component/`)
- Per-component builders (Kong, Auth, PostgREST, etc.)
- Resource specification (Deployment, Service, ConfigMap)
- Environment variable injection
- Default resource limits

**Status Management** (`internal/status/`)
- Condition management following Kubernetes API conventions
- Phase progression state machine
- Per-component status tracking

**Secrets Management** (`internal/secrets/`)
- JWT secret generation (cryptographically secure)
- API key generation (ANON_KEY, SERVICE_ROLE_KEY)
- Secret validation

**Database Initialization** (`internal/database/`)
- Schema creation (auth, storage, realtime)
- Extension setup (pgcrypto, uuid-ossp, pg_stat_statements)
- Role creation (authenticator, anon, service_role)

**Webhook Layer** (`internal/webhook/`)
- Validating webhook for CRD admission
- Mutating webhook for defaults
- Secret reference validation

## Controller Pattern

### Reconciliation Loop

The operator follows the standard Kubernetes controller pattern with idempotent reconciliation:

```go
1. Fetch SupabaseProject from API
2. Check for deletion (run finalizer cleanup if needed)
3. Add finalizer if not present
4. Validate external dependencies (PostgreSQL, S3)
5. Initialize database (schemas, extensions, roles)
6. Generate/reconcile JWT secrets
7. Deploy components in order:
   - Kong (API Gateway)
   - Auth (GoTrue)
   - PostgREST
   - Realtime
   - Storage API
   - Meta
   - Studio (optional)
8. Update status with component health
9. Update CRD status subresource
10. Requeue if needed
```

**Key Characteristics:**
- Idempotent operations: Safe to run multiple times
- Error handling: No automatic rollback, maintain failed state
- Retry mechanism: Controller-runtime's exponential backoff
- Watch-based: Triggered by resource changes

### Phase Management

The reconciler progresses through well-defined phases:

```
Pending
  ↓
ValidatingDependencies (check PostgreSQL, S3)
  ↓
InitializingDatabase (schemas, roles, extensions)
  ↓
DeployingSecrets (JWT generation)
  ↓
DeployingComponents (ordered deployment)
  ↓
Running (all healthy)

Error states:
- Failed (reconciliation error, no rollback)
- Updating (spec change detected)
```

Each phase transition is reflected in status conditions and logged for observability.

## Component Deployment Strategy

### Deployment Order

Components are deployed in a specific order to handle dependencies:

1. **Kong**: API Gateway must be first (routes traffic to all services)
2. **Auth**: Authentication service (needed by other components)
3. **PostgREST**: REST API layer
4. **Realtime**: WebSocket subscriptions
5. **Storage API**: File storage management
6. **Meta**: PostgreSQL metadata service
7. **Studio**: Management UI (optional)

### Resource Specifications

Each component builder creates:
- **Deployment**: Pod specification with containers, resources, env vars
- **Service**: ClusterIP service for internal communication
- **ConfigMap**: Configuration data (if needed)

Resource defaults are applied in the component builder when not specified:

| Component | Memory Limit | CPU Limit | Replicas |
|-----------|--------------|-----------|----------|
| Kong      | 2.5Gi        | 500m      | 1        |
| Auth      | 128Mi        | 100m      | 1        |
| PostgREST | 256Mi        | 200m      | 1        |
| Realtime  | 256Mi        | 200m      | 1        |
| Storage   | 128Mi        | 100m      | 1        |
| Meta      | 128Mi        | 100m      | 1        |
| Studio    | 256Mi        | 100m      | 1        |

### Health Checks

Each deployment includes:
- **Readiness probe**: Determines when pod is ready to receive traffic
- **Liveness probe**: Detects crashed/hung containers
- **Startup probe**: Allows slow-starting containers extra time

Component status is derived from Deployment readiness.

## Status Management

### Condition Types

The operator implements 15+ granular conditions following Kubernetes API conventions:

**Standard Conditions:**
- `Ready`: Overall readiness (all components healthy)
- `Progressing`: Reconciliation in progress
- `Available`: Endpoints accessible
- `Degraded`: Some components unhealthy

**Component Conditions:**
- `KongReady`, `AuthReady`, `RealtimeReady`, `StorageReady`, `PostgRESTReady`, `MetaReady`, `StudioReady`

**Dependency Conditions:**
- `PostgreSQLConnected`: Database connectivity
- `S3Connected`: Storage connectivity

**Infrastructure Conditions:**
- `NetworkReady`: Services and networking configured
- `SecretsReady`: JWT secrets generated

### Status Structure

```yaml
status:
  phase: Running
  message: "All components running"
  observedGeneration: 5
  lastReconcileTime: "2025-11-11T12:00:00Z"

  conditions:
    - type: Ready
      status: "True"
      reason: AllComponentsReady
      lastTransitionTime: "2025-11-11T12:00:00Z"

  components:
    kong:
      phase: Running
      ready: true
      version: "kong:2.8.1"
      replicas: 1
      readyReplicas: 1
    auth:
      phase: Running
      ready: true
      version: "supabase/gotrue:v2.177.0"
      replicas: 1
      readyReplicas: 1

  dependencies:
    postgresql:
      connected: true
      lastConnectedTime: "2025-11-11T12:00:00Z"
      latencyMs: 5
    s3:
      connected: true
      lastConnectedTime: "2025-11-11T12:00:00Z"
      latencyMs: 12

  endpoints:
    api: "http://my-supabase-kong:8000"
    auth: "http://my-supabase-auth:9999"
```

## Security Architecture

### Secret Management

**User-Provided Secrets:**
- Database credentials (PostgreSQL connection)
- S3 credentials (storage backend)
- SMTP credentials (optional, for Auth emails)
- OAuth provider credentials (optional)

**Operator-Generated Secrets:**
- JWT secret (256-bit cryptographically secure)
- ANON_KEY (JWT with 'anon' role claim)
- SERVICE_ROLE_KEY (JWT with 'service_role' role claim)
- PG_META_CRYPTO_KEY (encryption key for Meta service)

All secrets are stored in Kubernetes Secret resources with proper RBAC controls.

### Secret Validation

The admission webhook validates:
- Secret references exist in the same namespace
- Required keys are present in secrets
- Secret values are not empty

This prevents runtime failures due to misconfiguration.

### Database Initialization

On first reconciliation, the operator initializes PostgreSQL with:

**Extensions:**
- `pgcrypto`: Cryptographic functions for Supabase
- `uuid-ossp`: UUID generation
- `pg_stat_statements`: Query performance tracking

**Schemas:**
- `auth`: Authentication data (GoTrue)
- `storage`: File metadata (Storage API)
- `realtime`: Subscription tracking (Realtime)

**Roles:**
- `authenticator`: Request authenticator (used by Kong/PostgREST)
- `anon`: Anonymous access role
- `service_role`: Service-level access (bypasses RLS)

All operations are idempotent and safe to re-run.

## Configuration Design

### Configuration Sources

The operator follows a clear hierarchy for configuration:

1. **CRD Spec** (`SupabaseProject.spec`):
   - Component overrides (image, replicas, resources)
   - Feature flags
   - Network configuration

2. **Kubernetes Secrets**:
   - Database connection details
   - S3 credentials
   - SMTP configuration
   - OAuth provider settings

3. **Hardcoded Defaults**:
   - Resource limits (in component builders)
   - Image versions (in kubebuilder markers)
   - Configuration constants

**Design Rationale:**
- Sensitive data never appears in CRD spec (GitOps-safe)
- Defaults provide good out-of-box experience
- Explicit overrides for customization

### Configuration Updates

When `SupabaseProject.spec` changes:
1. Controller detects `metadata.generation` change
2. Status phase transitions to `Updating`
3. Affected resources are updated via strategic merge
4. Kubernetes handles rolling update of Deployments
5. Status reflects new generation in `observedGeneration`

No automatic rollback occurs on failures (investigative state preservation).

## Data Flow

### Request Flow (Runtime)

```
Client Request
    ↓
Kong API Gateway (Port 8000)
    ↓
Route Decision (based on path)
    ↓
    ├─→ /auth/*     → Auth Service (GoTrue)
    ├─→ /rest/*     → PostgREST Service
    ├─→ /realtime/* → Realtime Service
    ├─→ /storage/*  → Storage API Service
    └─→ /pg/*       → Meta Service
         ↓
    PostgreSQL Database
```

### Control Flow (Reconciliation)

```
User applies SupabaseProject CR
    ↓
Kubernetes API persists CR
    ↓
Controller watch triggers
    ↓
Reconcile function executes
    ↓
Validate dependencies
    ↓
Initialize database
    ↓
Generate/update secrets
    ↓
Deploy/update components (ordered)
    ↓
Query component health
    ↓
Update status subresource
    ↓
Reconciliation complete (requeue if needed)
```

## Design Decisions

### 1. External Dependencies Only

**Decision:** Operator does not deploy PostgreSQL or S3-compatible storage.

**Rationale:**
- Supabase requires persistent, highly-available database
- Production PostgreSQL needs specialized operators (Zalando, Crunchy)
- Storage backends have diverse deployment requirements
- Separation of concerns: focus on Supabase components

**Trade-off:** Users must provision infrastructure separately.

### 2. No Automatic Rollback

**Decision:** Failed deployments remain in failed state for investigation.

**Rationale:**
- Preserves "crime scene" for debugging
- Follows Kubernetes controller best practices
- Automatic rollback can mask underlying issues
- Users have full control via spec updates

**Implementation:** Controller sets `phase: Failed` and stops reconciliation.

### 3. Secrets in Kubernetes Only

**Decision:** All sensitive configuration stored in Kubernetes Secrets.

**Rationale:**
- GitOps-safe: CRD specs contain no secrets
- Integrates with external secret managers (via Secret Store CSI)
- Follows Kubernetes security best practices
- RBAC controls access

**Trade-off:** Less convenient for quick debugging (must view secrets separately).

### 4. Ordered Component Deployment

**Decision:** Components deploy in fixed order (Kong → Auth → PostgREST → ...).

**Rationale:**
- Kong must exist before services (routes traffic)
- Auth should be available early (other components may depend)
- Predictable startup sequence simplifies debugging

**Implementation:** Controller deploys components sequentially in `reconcileComponents()`.

### 5. Status-Driven Reconciliation

**Decision:** Comprehensive status reporting with 15+ conditions.

**Rationale:**
- Inspired by Rook operator (proven pattern)
- Enables detailed observability
- Supports GitOps workflows (status as source of truth)
- Facilitates debugging

**Trade-off:** More complex status management code.

### 6. Kubebuilder Framework

**Decision:** Built on Kubebuilder v4.0+ scaffolding.

**Rationale:**
- Industry-standard operator framework
- Automatic RBAC generation from markers
- Integrated testing with envtest
- CRD generation from Go types
- Well-documented patterns

**Benefit:** Reduced boilerplate, focus on business logic.

## Extensibility

### Adding New Components

To add a new Supabase component:

1. Add configuration type in `api/v1alpha1/supabaseproject_types.go`
2. Create component builder in `internal/component/<name>.go`
3. Add condition type in `internal/status/conditions.go`
4. Update reconciliation loop in `internal/controller/supabaseproject_controller.go`
5. Add tests in `internal/component/<name>_test.go`
6. Update CRD with `make manifests`

### Custom Resource Overrides

Users can override any component setting via CRD spec:

```yaml
spec:
  kong:
    image: kong:3.0.0
    replicas: 3
    resources:
      limits:
        memory: "4Gi"
    extraEnv:
      - name: KONG_LOG_LEVEL
        value: "debug"
```

The component builder merges user overrides with defaults.

### Webhook Extensions

Validation logic is centralized in `internal/webhook/supabaseproject_webhook.go`:

```go
func (v *SupabaseProjectWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error
func (v *SupabaseProjectWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error
func (v *SupabaseProjectWebhook) Default(ctx context.Context, obj runtime.Object) error
```

Additional validation rules can be added without changing the controller.

## Testing Strategy

### Test Levels

1. **Unit Tests** (`*_test.go`):
   - Component builders return correct specs
   - Status condition helpers work correctly
   - JWT generation produces valid tokens

2. **Integration Tests** (envtest):
   - Controller reconciliation with fake Kubernetes API
   - Dependency validation logic
   - Status updates propagate correctly
   - Finalizer cleanup works

3. **E2E Tests** (`test/e2e/`):
   - Real Kubernetes cluster (Minikube)
   - Full deployment lifecycle
   - Component health checks
   - Spec updates trigger rolling updates

### Test-Driven Development

The project follows TDD:
- Tests written before implementation
- Red-Green-Refactor cycle
- envtest for controller testing (no real cluster needed)

## Observability

### Logging

Structured logging via controller-runtime:

```go
logger := log.FromContext(ctx)
logger.Info("Reconciling SupabaseProject", "name", project.Name)
logger.Error(err, "Failed to create deployment", "component", "kong")
```

Logs include request context for correlation.

### Metrics

Prometheus metrics exposed at `:8443/metrics`:

- `controller_runtime_reconcile_total`: Total reconciliations
- `controller_runtime_reconcile_errors_total`: Error count
- `controller_runtime_reconcile_time_seconds`: Duration histogram
- `workqueue_depth`: Queue length
- `workqueue_adds_total`: Items enqueued

ServiceMonitor available for Prometheus Operator integration.

### Events

Kubernetes events emitted for significant state changes:

```go
r.Recorder.Event(project, "Normal", "Created", "Created Kong deployment")
r.Recorder.Event(project, "Warning", "Failed", "PostgreSQL connection failed")
```

Events visible via `kubectl describe supabaseproject`.

## Performance Considerations

### Reconciliation Efficiency

- **Caching**: controller-runtime caches watched resources
- **Selective Updates**: Only update resources that changed
- **Status Subresource**: Avoids unnecessary reconciliations
- **Predicates**: Filter watch events (ignore status-only updates)

**Target:** <5s reconciliation time for spec updates.

### Resource Limits

Operator itself has minimal resource requirements:
- Memory: <256MB
- CPU: <100m

Designed to manage 100+ SupabaseProject instances per cluster.

### Scalability

Multi-tenant design:
- Namespace-scoped CRD
- Multiple projects per namespace supported
- No shared state between projects
- Component isolation via labels

## Future Considerations

See [future-considerations.md](../specs/001-selfhost-supabase-operator/future-considerations.md) for planned enhancements:

- High availability (multi-replica deployments)
- Backup and restore integration
- Advanced networking (Ingress, TLS)
- Database connection pooling (PgBouncer)
- Horizontal Pod Autoscaling
- Custom resource hooks
- Multi-cluster federation

## References

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Rook Operator](https://github.com/rook/rook) (status management patterns)
- [Supabase Documentation](https://supabase.com/docs)
- [Controller Runtime](https://github.com/kubernetes-sigs/controller-runtime)
