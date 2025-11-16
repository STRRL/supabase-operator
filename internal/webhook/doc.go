// Package webhook implements admission webhooks for SupabaseProject validation.
//
// This package provides both validating and mutating admission webhooks for the
// SupabaseProject custom resource. The webhooks run as part of the Kubernetes
// admission control flow before resources are persisted to etcd.
//
// Webhook Types:
//
//  1. Validating Webhook:
//     Validates the SupabaseProject resource for correctness.
//     Rejects invalid resources with descriptive error messages.
//
//  2. Mutating Webhook:
//     Applies default values to optional fields.
//     Sets resource defaults when not specified.
//
// Validation Rules:
//
// The validating webhook enforces the following rules:
//
// Secret Validation:
//   - Referenced secrets must exist in the same namespace
//   - Database secret must contain keys: host, port, database, username, password
//   - Storage secret must contain keys: endpoint, region, bucket, accessKeyId, secretAccessKey
//   - All secret values must be non-empty
//
// Field Constraints:
//   - projectId must match DNS-1123 pattern: ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
//   - database.maxConnections must be between 1 and 100
//   - Component replicas must be between 0 and 10
//
// Cross-Field Validation:
//   - Resource limits must be greater than or equal to requests
//   - Image references must be valid container image URIs
//
// Mutating Logic:
//
// The mutating webhook applies defaults:
//   - Component images (e.g., kong:2.8.1)
//   - Resource requirements (memory, CPU)
//   - Replica counts (defaults to 1)
//   - SSL mode for database (defaults to "require")
//   - Path style for S3 (defaults to true)
//
// Example webhook registration:
//
//	err = ctrl.NewWebhookManagedBy(mgr).
//	    For(&supabasev1alpha1.SupabaseProject{}).
//	    WithValidator(&webhook.SupabaseProjectWebhook{Client: mgr.GetClient()}).
//	    WithDefaulter(&webhook.SupabaseProjectWebhook{Client: mgr.GetClient()}).
//	    Complete()
//
// Error Responses:
//
// Validation failures return descriptive errors:
//   - "database secret 'postgres-config' not found in namespace 'default'"
//   - "database secret missing required key 'host'"
//   - "component replicas must be between 0 and 10, got: 15"
//
// These errors are returned directly to the user via kubectl/API.
//
// Webhook Configuration:
//
// The webhook is configured via kubebuilder markers in the API types:
//
//	//+kubebuilder:webhook:path=/validate-supabase-strrl-dev-v1alpha1-supabaseproject
//	//+kubebuilder:webhook:path=/mutate-supabase-strrl-dev-v1alpha1-supabaseproject
//
// The operator must have a valid TLS certificate for webhook serving.
// This is typically provided by cert-manager in production deployments.
//
// Security Considerations:
//
// - Webhooks prevent invalid configurations from reaching the controller
// - Early validation improves user experience (fail fast)
// - Secret content is never exposed in error messages
// - Webhooks must complete quickly (<1s) to avoid timeouts
package webhook
