// Package component provides builders for Supabase component Kubernetes resources.
//
// This package contains functions to build Deployments, Services, and ConfigMaps
// for each Supabase component: Kong, Auth, PostgREST, Realtime, Storage API, Meta, and Studio.
//
// Each component builder:
//   - Creates a Deployment with proper container specifications
//   - Creates a ClusterIP Service for internal communication
//   - Applies default resource limits if not specified in the CRD
//   - Injects environment variables from secrets and configuration
//   - Sets proper labels and owner references for garbage collection
//
// Example usage:
//
//	deployment := component.BuildKongDeployment(project, jwtSecret, dbSecret)
//	service := component.BuildKongService(project)
//
// Component deployment order is important and handled by the controller:
//  1. Kong (API Gateway must be first)
//  2. Auth (Authentication service)
//  3. PostgREST (REST API layer)
//  4. Realtime (WebSocket subscriptions)
//  5. Storage API (File storage management)
//  6. Meta (PostgreSQL metadata service)
//  7. Studio (Management UI, optional)
//
// All builders are idempotent and produce deterministic output for the same input,
// enabling proper reconciliation in the controller loop.
package component
