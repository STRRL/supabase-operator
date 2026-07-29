# Research: Supabase Operator Implementation

## Kubernetes Operator Patterns

### Decision: Kubebuilder Framework
**Rationale**: Industry standard for Go operators, provides scaffolding, CRD generation, webhook support, and envtest integration
**Alternatives considered**:
- Operator SDK: More complex, better for multi-language
- Manual controller-runtime: More work, less standardization
- Kopf (Python): Slower performance, less mature ecosystem

### Decision: Reconciliation Pattern
**Rationale**: Follow Kubernetes controller pattern with idempotent reconciliation loops
**Key Patterns from Rook**:
- Component-aware status tracking
- Granular condition management
- Phase progression per component
- Ordered deployment with health checks

## Supabase Component Dependencies

### Decision: Component Deployment Order
**Rationale**: Based on service dependencies and initialization requirements
**Order**:
1. Validate external dependencies (PostgreSQL, S3)
2. Generate and store secrets (JWT, API keys)
3. Deploy Kong (API gateway - entry point)
4. Deploy Auth/GoTrue (needs Kong routes)
5. Deploy PostgREST (needs auth)
6. Deploy Realtime (needs auth)
7. Deploy Storage API (needs auth, S3)
8. Deploy Meta (metadata service)

### Decision: Container Images
**Sources**: Official Supabase Docker images with specific versions from upstream docker-compose.yml, https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml
- `kong:2.8.1`
- `supabase/gotrue:v2.177.0`
- `postgrest/postgrest:v12.2.12`
- `supabase/realtime:v2.34.47`
- `supabase/storage-api:v1.25.7`
- `supabase/postgres-meta:v0.91.0`

**Note**: Image tags are configurable per component in the CRD, with defaults fetched from upstream

## Configuration Management

### Decision: Secret Generation Strategy
**Rationale**: Generate once, store in Kubernetes Secret, never regenerate unless explicitly requested
**Implementation**:
- Use crypto/rand for secure generation
- Base64 encode JWT secret (32 bytes)
- Generate ANON_KEY and SERVICE_ROLE_KEY from JWT secret
- Store in namespace-scoped Secret

### Decision: Configuration Injection
**Rationale**: Use environment variables for service configuration
**Method**:
- ConfigMaps for non-sensitive config
- Secrets for credentials and keys
- Volume mounts for complex configs (Kong routes)

## External Dependencies

### Decision: PostgreSQL Integration
**Rationale**: Users provide connection details, no managed mode initially
**Required fields**:
- Host, port, database name
- Username (via Secret reference)
- Password (via Secret reference)
- SSL mode configuration

### Decision: S3 Integration
**Rationale**: Support any S3-compatible storage
**Required fields**:
- Endpoint URL
- Region
- Bucket name
- Access key ID (via Secret)
- Secret access key (via Secret)

## Status Management

### Decision: Condition Types
**Rationale**: Follow Kubernetes conventions with Rook-inspired granularity
**Standard Conditions**:
- Ready: All components healthy
- Progressing: Reconciliation in progress
- Available: Services accessible
- Degraded: Partial failures detected

**Component Conditions** (15 total):
- PostgreSQLConnected, S3Connected
- SecretsReady, NetworkReady
- KongReady, AuthReady, RealtimeReady, StorageReady, PostgRESTReady, MetaReady

### Decision: Phase Tracking
**Rationale**: Per-component phases provide granular visibility
**Phases**:
- ValidatingDependencies
- Deploying
- Configuring
- Running
- Failed
- Updating

## Resource Management

### Decision: Default Resource Limits
**Rationale**: Based on production usage patterns and Docker Compose observations
**Implementation**: Controller applies these defaults when Resources field is nil
**Defaults** (hardcoded in controller):
- Kong: 2.5GB memory, 500m CPU (handles routing for all services)
- Realtime: 256MB memory, 200m CPU (WebSocket connections)
- PostgREST: 256MB memory, 200m CPU (database queries)
- Auth/GoTrue: 128MB memory, 100m CPU (authentication)
- Storage API: 128MB memory, 100m CPU (file operations)
- Meta: 128MB memory, 100m CPU (metadata)

### Decision: Update Strategy
**Rationale**: Rolling updates minimize downtime while maintaining Kubernetes reconciliation patterns
**Implementation**:
- Update one component at a time
- Health check before proceeding
- Kong updated last (API gateway)
- Keep failed state for manual investigation (no automatic rollback)
- Let reconciliation loop handle retries

## Testing Strategy

### Decision: Test Framework
**Rationale**: Standard Go testing with Kubernetes-specific tools
**Tools**:
- Ginkgo/Gomega for BDD-style tests
- Envtest for controller testing
- Fake client for unit tests
- Kind for e2e testing

### Decision: Test Coverage Requirements
**Rationale**: High coverage for critical paths
**Requirements**:
- 80% unit test coverage
- Integration tests for all reconciliation paths
- E2E tests for create/update/delete scenarios
- Failure injection tests

## Observability

### Decision: Logging Strategy
**Rationale**: Structured logging for production debugging
**Implementation**:
- controller-runtime logger (zap-based)
- Correlation IDs for request tracking
- Log levels: Debug, Info, Warn, Error

### Decision: Metrics Strategy
**Rationale**: Prometheus metrics for monitoring
**Metrics**:
- Reconciliation duration histogram
- Component health gauges
- Error rate counters
- Resource count gauges

### Decision: Events Strategy
**Rationale**: Kubernetes events for user visibility
**Events**:
- Component deployment started/completed
- Dependency validation success/failure
- Configuration updates
- Error conditions

## Security Considerations

### Decision: RBAC Strategy
**Rationale**: Least privilege principle
**Permissions**:
- Create/update/delete: Deployments, Services, ConfigMaps, Secrets
- Get/list/watch: SupabaseProjects, Pods
- Update: SupabaseProject status

### Decision: Network Policies
**Rationale**: Secure by default
**Implementation**:
- Ingress rules for service communication
- Egress rules for external dependencies
- Deny all by default

## Migration and Upgrades

### Decision: CRD Versioning
**Rationale**: Follow Kubernetes API conventions
**Strategy**:
- Start with v1alpha1
- Implement conversion webhooks for v1beta1
- Maintain backward compatibility

### Decision: Component Upgrade Strategy
**Rationale**: Coordinated upgrades with validation following Kubernetes patterns
**Process**:
1. Validate new versions compatibility
2. Update components in dependency order
3. Verify health after each update
4. Keep failed state for investigation (no automatic rollback)
5. Allow manual intervention or reconciliation retry
