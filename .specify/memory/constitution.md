<!--
Sync Impact Report:
Version: 1.0.0 (initial constitution)
Modified Principles: None (new constitution)
Added Sections:
  - Core Principles (I-VII)
  - Kubernetes Best Practices
  - Development Workflow
  - Governance
Templates Requiring Updates:
  ✅ plan-template.md (aligned with Kubebuilder principles)
  ✅ spec-template.md (aligned with operator requirements)
  ✅ tasks-template.md (aligned with TDD and controller patterns)
Follow-up TODOs: None
-->

# Supabase Operator Constitution

## Core Principles

### I. Controller Reconciliation Pattern (NON-NEGOTIABLE)

All business logic MUST be implemented in reconciliation loops following the Kubernetes controller pattern. Controllers MUST be idempotent, edge-triggered, and level-driven. Each reconcile function MUST handle both creation and updates, returning appropriate results for requeue behavior.

**Rationale**: This is the fundamental pattern for Kubernetes operators. Idempotent reconciliation ensures the system self-heals and converges to desired state regardless of transient failures or partial updates.

### II. Custom Resource Definitions First

All operator functionality MUST be exposed through well-designed CRDs with comprehensive validation, defaulting, and versioning. CRDs MUST have detailed OpenAPI schemas with field descriptions, validation rules, and examples. Status subresources MUST be used for all CRDs.

**Rationale**: CRDs are the API contract. Well-designed CRDs with proper validation prevent invalid configurations and reduce error handling complexity. Status subresources enable proper status reporting without triggering reconciliation loops.

### III. Test-First Development (NON-NEGOTIABLE)

TDD is mandatory: Unit tests for reconciliation logic → User approved → Tests fail → Implement reconciliation logic. Integration tests MUST use envtest (controller-runtime test environment). Tests MUST verify both happy paths and error scenarios including resource conflicts, API errors, and dependency failures.

**Rationale**: Kubernetes controllers have complex state management. Tests prevent regressions and ensure reconciliation logic handles all edge cases, especially failure scenarios that are hard to reproduce manually.

### IV. Structured Status Reporting

Every custom resource MUST have a comprehensive status subresource with conditions following Kubernetes conventions. Conditions MUST use standard types (Ready, Available, Progressing) and include ObservedGeneration for change detection. Status updates MUST use separate API calls and MUST NOT trigger reconciliation.

**Rationale**: Users and other controllers need visibility into operator state. Standard conditions enable kubectl wait and consistent tooling. Proper status management prevents reconciliation loops and enables debugging.

### V. Dependency Integration via Composition

Supabase Operator MUST NOT reimplement functionality provided by existing operators. PostgreSQL MUST use CloudNativePG, Zalando, or Crunchy operators. Object storage MUST use MinIO Operator or cloud provider integrations. The operator MUST coordinate external resources through owner references and label selectors.

**Rationale**: Kubernetes ecosystem thrives on composition. Reusing battle-tested operators reduces code, improves reliability, and provides users with best-in-class implementations for each component.

### VI. Observability and Operations

All controllers MUST emit structured logs using controller-runtime logging. Metrics MUST be exposed via Prometheus endpoints for reconciliation duration, queue depth, error rates, and API call latency. Events MUST be recorded for significant state transitions and user-actionable errors.

**Rationale**: Operators run autonomously and must be observable. Metrics enable SLO tracking and alerting. Events provide users with real-time feedback without log access. Structured logs enable troubleshooting in production.

### VII. Security and RBAC

Operator MUST follow principle of least privilege. Service accounts MUST have minimal RBAC permissions required for reconciliation. Secrets MUST be generated using cryptographically secure methods. TLS certificates MUST be managed via cert-manager integration. Sensitive configuration MUST be stored in Kubernetes Secrets, never in CRD spec fields.

**Rationale**: Operators have elevated privileges and manage sensitive data. Minimal RBAC reduces blast radius. Proper secret management prevents credential leaks. Cert-manager integration ensures certificate rotation without operator code.

## Kubernetes Best Practices

### API Versioning

CRDs MUST start at v1alpha1 and follow Kubernetes API versioning conventions. Breaking changes MUST increment API version. Conversion webhooks MUST be implemented before promoting to v1beta1. All versions MUST support bidirectional conversion.

**Rationale**: Users need stable APIs. Version progression signals maturity. Conversion webhooks enable gradual migration and backward compatibility.

### Finalizers and Cleanup

Resources that create external dependencies MUST use finalizers for proper cleanup. Finalizer logic MUST be idempotent and handle partial cleanup. Finalizers MUST be removed only after successful cleanup to prevent resource leaks.

**Rationale**: Kubernetes garbage collection alone cannot clean external resources. Finalizers ensure PostgreSQL databases, S3 buckets, and service accounts are deleted when Supabase instances are removed.

### Admission Webhooks

Validating webhooks MUST be implemented for complex cross-field validation that cannot be expressed in OpenAPI schemas. Mutating webhooks MUST be used for defaulting complex configurations. Webhook implementations MUST be fast (<1s) and handle failures gracefully with appropriate failure policies.

**Rationale**: Webhooks provide runtime validation beyond static schemas. Defaulting webhooks improve user experience by auto-configuring sensible defaults.

### Performance and Scalability

Controllers MUST use caching and avoid direct API calls in hot paths. List operations MUST use label selectors to minimize data transfer. Watch predicates MUST filter irrelevant events. Reconciliation MUST be bounded by timeouts to prevent queue blocking.

**Rationale**: Operators must scale to hundreds of resources. Unbounded reconciliation and excessive API calls degrade cluster performance and violate Kubernetes API request limits.

## Development Workflow

### Code Organization

Follow Kubebuilder project layout: `api/` for CRD types, `controllers/` for reconciliation logic, `internal/` for shared utilities. Each controller MUST be in a separate file. Reconciliation logic MUST be extracted into testable service functions.

**Rationale**: Standard layout improves code navigation and maintainability. Testable service functions enable unit testing without Kubernetes API machinery.

### Documentation Requirements

All CRD fields MUST have kubebuilder markers with descriptions. Controllers MUST have top-level package documentation explaining reconciliation strategy. Design decisions MUST be documented in `docs/design/` with ADR (Architecture Decision Record) format.

**Rationale**: Kubebuilder generates CRD OpenAPI schemas from markers. Package docs help maintainers understand control flow. ADRs preserve context for future changes.

### Integration Testing

Integration tests MUST use envtest with real etcd and API server. Tests MUST verify reconciliation correctness, status updates, finalizer cleanup, and error handling. CI MUST run integration tests on every PR.

**Rationale**: Unit tests alone cannot verify controller-runtime integration. Envtest provides lightweight Kubernetes environment for comprehensive testing without full cluster setup.

## Governance

This constitution supersedes all other development practices. All design decisions MUST verify compliance with these principles. Deviations MUST be documented with justification in design docs and approved explicitly.

**Amendment Process**: Constitutional changes require documented rationale, impact analysis on existing code, and migration plan if breaking changes are introduced.

**Compliance Review**: All PRs MUST pass constitution checks. Complex features MUST document constitution compliance in design phase. Use `.specify/memory/constitution.md` for governance and `.specify/templates/agent-file-template.md` for runtime development guidance.

**Version**: 1.0.0 | **Ratified**: 2025-10-02 | **Last Amended**: 2025-10-02
