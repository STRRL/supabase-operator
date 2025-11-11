// Package status provides status management utilities for SupabaseProject resources.
//
// This package implements comprehensive status tracking following Kubernetes API
// conventions and patterns from the Rook operator. It provides:
//   - Condition management (creation, updates, transitions)
//   - Phase progression state machine
//   - Component status tracking
//   - Helper functions for common status operations
//
// Condition Types:
//
// The package defines 15+ granular conditions:
//
// Standard Conditions (Kubernetes API conventions):
//   - Ready: Overall readiness (True when all components healthy)
//   - Progressing: Reconciliation in progress
//   - Available: Endpoints accessible
//   - Degraded: Some components unhealthy
//
// Component-Specific Conditions:
//   - KongReady, AuthReady, RealtimeReady, StorageReady, PostgRESTReady, MetaReady, StudioReady
//
// Dependency Conditions:
//   - PostgreSQLConnected: Database connectivity verified
//   - S3Connected: Storage connectivity verified
//
// Infrastructure Conditions:
//   - NetworkReady: Services and networking configured
//   - SecretsReady: JWT secrets generated
//
// Phase Management:
//
// The state machine progresses through well-defined phases:
//
//	Pending → ValidatingDependencies → InitializingDatabase →
//	DeployingSecrets → DeployingComponents → Running
//
// Error states:
//   - Failed: Reconciliation error (no automatic rollback)
//   - Updating: Spec change detected
//
// Each phase transition includes:
//   - Human-readable message
//   - Timestamp tracking
//   - Condition updates
//
// Example usage:
//
//	project.Status.Phase = status.PhaseRunning
//	project.Status.Message = status.GetPhaseMessage(status.PhaseRunning)
//
//	project.Status.Conditions = status.SetCondition(
//	    project.Status.Conditions,
//	    status.NewReadyCondition(metav1.ConditionTrue, "AllComponentsReady", ""),
//	)
//
// Component Status:
//
// Per-component status tracking includes:
//   - Phase: Component lifecycle phase
//   - Ready: Boolean readiness flag
//   - Version: Deployed container image
//   - Replicas: Total and ready replica counts
//   - Conditions: Component-specific conditions
//   - LastUpdateTime: Timestamp of last status change
//
// Example usage:
//
//	componentStatus := status.NewComponentStatus(
//	    "Running",
//	    true,
//	    "kong:2.8.1",
//	    1,
//	    1,
//	)
//
// Best Practices:
//
// - Always update conditions atomically using SetCondition
// - Include descriptive reasons in condition transitions
// - Update observedGeneration after successful reconciliation
// - Set lastReconcileTime on every reconciliation
// - Use phase messages for human-readable status
package status
