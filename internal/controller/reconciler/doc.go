// Package reconciler provides component reconciliation utilities for the controller.
//
// This package contains helper functions for reconciling individual Supabase
// components as part of the main controller reconciliation loop. It encapsulates
// the logic for creating or updating component Deployments and Services.
//
// Component Reconciliation:
//
// Each component reconciliation function:
//   - Builds the desired Deployment and Service from the CRD spec
//   - Checks if the resources already exist in the cluster
//   - Creates new resources if they don't exist
//   - Updates existing resources if the spec has changed
//   - Sets owner references for garbage collection
//   - Returns errors for the controller to handle
//
// Example usage:
//
//	err := reconciler.ReconcileKong(ctx, r.Client, project, jwtSecret, dbSecret)
//	if err != nil {
//	    return ctrl.Result{}, err
//	}
//
// Reconciliation Strategy:
//
// The package uses Kubernetes strategic merge for updates:
//   - Only specified fields are updated
//   - Kubernetes handles rolling updates for Deployments
//   - Services are updated in-place
//   - Owner references ensure proper garbage collection
//
// Error Handling:
//
// Reconciliation errors are returned to the caller:
//   - Creation errors: Resource could not be created
//   - Update errors: Resource could not be updated
//   - API errors: Communication with Kubernetes API failed
//
// The controller will retry on errors with exponential backoff.
//
// Idempotency:
//
// All reconciliation functions are idempotent:
//   - Safe to call multiple times with same input
//   - Produce deterministic output
//   - No side effects on repeated calls
//
// This enables reliable reconciliation in the controller loop.
package reconciler
