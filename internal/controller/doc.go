// Package controller implements the Kubernetes controller for SupabaseProject resources.
//
// The controller follows the standard Kubernetes operator pattern with idempotent
// reconciliation. It watches SupabaseProject custom resources and ensures the actual
// cluster state matches the desired state specified in the CRD.
//
// Reconciliation Flow:
//  1. Fetch SupabaseProject from API
//  2. Handle deletion (run finalizer cleanup if needed)
//  3. Add finalizer if not present
//  4. Validate external dependencies (PostgreSQL, S3)
//  5. Initialize database (schemas, extensions, roles)
//  6. Generate/reconcile JWT secrets
//  7. Deploy components in order (Kong → Auth → PostgREST → Realtime → Storage → Meta → Studio)
//  8. Update status with component health
//  9. Update CRD status subresource
//  10. Requeue if needed
//
// Phase Management:
//
// The controller progresses through well-defined phases:
//   - Pending: Initial state
//   - ValidatingDependencies: Checking PostgreSQL and S3
//   - InitializingDatabase: Creating schemas, roles, extensions
//   - DeployingSecrets: Generating JWT secrets
//   - DeployingComponents: Creating deployments
//   - Running: All components healthy
//   - Failed: Reconciliation error (no automatic rollback)
//   - Updating: Spec change detected
//
// Error Handling:
//
// The controller does not perform automatic rollbacks on failures. Failed states
// are preserved for investigation, following Kubernetes best practices. Users
// must manually fix issues and the controller will retry reconciliation.
//
// Status Management:
//
// The controller maintains 15+ granular conditions following Kubernetes API conventions:
//   - Standard: Ready, Progressing, Available, Degraded
//   - Component: KongReady, AuthReady, RealtimeReady, etc.
//   - Dependency: PostgreSQLConnected, S3Connected
//   - Infrastructure: NetworkReady, SecretsReady
//
// Per-component status includes phase, readiness, version, and replica counts.
//
// RBAC Permissions:
//
// The controller requires permissions defined via kubebuilder markers:
//   - SupabaseProject: full CRUD + status + finalizers
//   - Deployments, Services, ConfigMaps, Secrets: full CRUD
//   - Jobs: full CRUD (for database initialization)
//   - Events: create, patch (for event recording)
package controller
