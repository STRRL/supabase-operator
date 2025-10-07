# Feature Specification: Supabase Operator for Self-Hosted Deployments

**Feature Branch**: `001-selfhost-supabase-operator`
**Created**: 2025-10-03
**Status**: Draft
**Input**: User description: "selfhost supabase operator, when we create a supabase project CRD in kube apiserver, we should provider the entire stack of supabase's component, but we do not take care about the postgresql and s3 part in this controller;"


## Clarifications

### Session 2025-10-03
- Q: How should the operator handle PostgreSQL and S3 dependencies? → A: External only, no managed mode
- Q: What are the default resource limits for components? → A: Based on production usage (see FR-012)
- Q: What information should the SupabaseProject status display? → A: Detailed status (component status + versions + endpoints + observed generation + last reconcile time)
- Q: Which Kubernetes-standard conditions should be included? → A: Ready + Progressing + Available + Degraded (partial failures)
- Q: How should the operator handle updates to specifications? → A: Rolling updates (one by one with health checks)
- Q: Should we add component-specific conditions? → A: Full granular (Component + Dependency + NetworkReady + SecretsReady)
- Q: What phases should SupabaseProject go through? → A: Component-aware with dependency validation


## User Scenarios & Testing *(mandatory)*

### Primary User Story
As a platform administrator, I want to deploy multiple isolated Supabase instances on my Kubernetes cluster by creating custom resources, so that different teams can have their own dedicated Supabase projects without manual configuration of individual components.

### Acceptance Scenarios
1. **Given** a Kubernetes cluster with the Supabase operator installed, **When** I create a SupabaseProject custom resource with minimal configuration, **Then** all Supabase components (Auth, Realtime, Storage API, PostgREST, Meta, Kong) are deployed and configured automatically
2. **Given** an existing SupabaseProject deployment, **When** I update the custom resource specification with new configuration values, **Then** the operator reconciles and updates the deployed components to match the desired state
3. **Given** external PostgreSQL and S3 storage services are available, **When** I create a SupabaseProject with connection details to these services, **Then** the Supabase stack connects to and uses the external dependencies
4. **Given** a running SupabaseProject instance, **When** I delete the custom resource, **Then** all managed Supabase components are properly cleaned up while preserving external dependencies

### Edge Cases
- When external PostgreSQL becomes unavailable, the operator marks PostgreSQLConnected condition as False and enters reconciliation retry with exponential backoff
- System handles partial deployment failures by maintaining per-component status tracking, allowing healthy components to continue running while failed components retry
- When required secrets or credentials are missing, the operator blocks deployment and sets SecretsReady condition to False with descriptive error message
- Operator handles version upgrades through rolling updates: updating one component at a time with health checks between updates (manual version changes in CRD spec)
- When resource limits are exceeded during deployment, Kubernetes OOMKills the pod, operator detects via pod status and reports in component condition
- System detects degraded state through health check endpoints, pod restart counts, and readiness probes, updating Degraded condition accordingly

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: System MUST deploy all core Supabase components (Kong API Gateway, GoTrue Auth, PostgREST, Realtime, Storage API, Meta) when a SupabaseProject resource is created
- **FR-002**: System MUST accept connection details for external PostgreSQL database (host, port, database, username, password/secret reference) and use it as the backing store for all Supabase services
- **FR-003**: System MUST accept configuration for external S3-compatible object storage (endpoint, region, bucket name, access key, secret key) and use it for the Storage API component
- **FR-004**: System MUST generate and manage JWT secrets and API keys (ANON_KEY, SERVICE_ROLE_KEY) for each SupabaseProject instance
- **FR-005**: System MUST report detailed deployment status through the SupabaseProject resource status including: individual component health (Ready/NotReady), deployed versions, service endpoints, observed generation for drift detection, and last successful reconciliation timestamp
- **FR-006**: System MUST support updating deployed components when SupabaseProject specification changes using rolling update strategy with health checks between component updates
- **FR-007**: System MUST clean up all managed resources when a SupabaseProject is deleted, while preserving external dependencies
- **FR-008**: System MUST support deploying multiple isolated SupabaseProject instances in different namespaces
- **FR-009**: System MUST expose service endpoints for client applications to connect to each Supabase instance
- **FR-010**: System MUST validate that required external dependencies (PostgreSQL, S3) are accessible before deploying Supabase components
- **FR-011**: System MUST handle component failures gracefully and attempt reconciliation to restore desired state
- **FR-012**: System MUST support configuration of resource limits for each Supabase component with defaults: Kong (2.5GB memory, 500m CPU), Realtime (256MB memory, 200m CPU), Auth/GoTrue (128MB memory, 100m CPU), Storage API (128MB memory, 100m CPU), PostgREST (256MB memory, 200m CPU), Meta (128MB memory, 100m CPU)
- **FR-013**: System MUST provide observability through logs and events for deployment operations and errors
- **FR-015**: System MUST include granular Kubernetes conditions in status: standard conditions (Ready, Progressing, Available, Degraded), component-specific conditions (KongReady, AuthReady, RealtimeReady, StorageReady, PostgRESTReady, MetaReady), dependency conditions (PostgreSQLConnected, S3Connected), and infrastructure conditions (NetworkReady, SecretsReady)
- **FR-014**: System MUST support external dependencies only - users provide PostgreSQL connection details (host, port, credentials) and S3-compatible storage configuration (endpoint, bucket, access keys)
- **FR-016**: System MUST track deployment phase per component (e.g., Kong:ValidatingDeps, Auth:Deploying, Realtime:Running) with phases including: ValidatingDependencies, Deploying, Configuring, Running, Failed, and Updating
- **FR-017**: System MUST handle PostgreSQL database initialization including creating required schemas (auth, storage, realtime), roles (authenticator, anon, service_role), and extensions (pgcrypto, pgjwt, uuid-ossp, pg_stat_statements) for Supabase components

### Non-Functional Requirements
- **NFR-001 (Performance)**: Reconciliation loop MUST complete within 5 seconds for warm cache scenarios (existing resources, no image pulls required)
- **NFR-002 (Availability)**: Control plane operations MUST maintain 99.9% uptime, allowing for at most 8.76 hours downtime per year
- **NFR-003 (Scalability)**: Single operator instance MUST support managing 100+ SupabaseProject resources across multiple namespaces without performance degradation
- **NFR-004 (Security)**: All secrets MUST be encrypted at rest using Kubernetes native encryption, all inter-component communication MUST use TLS 1.2+
- **NFR-005 (Reliability)**: System MUST handle transient failures (network timeouts, pod restarts) through exponential backoff retry with maximum 5 retries
- **NFR-006 (Resource Efficiency)**: Operator pod MUST consume less than 256MB memory and 100m CPU under normal operation with 100 managed instances
- **NFR-007 (Observability)**: All reconciliation operations MUST emit structured logs with correlation IDs, trace IDs, and operation duration metrics
- **NFR-008 (Compatibility)**: Operator MUST support Kubernetes versions 1.25+ and be tested against latest 3 stable releases

### Key Entities *(include if feature involves data)*
- **SupabaseProject**: Represents a complete Supabase instance with all its components, configuration, and status
- **Component Status**: Per-component tracking including phase (ValidatingDeps/Deploying/Running/Failed), health state, version, endpoints, and individual conditions
- **Connection Configuration**: External dependency details for PostgreSQL and S3 storage
- **Secrets**: Generated credentials and keys required for Supabase operation

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

