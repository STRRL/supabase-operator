package v1alpha1

// Default image versions for Supabase components.
//
// This file is the single source of truth for component image defaults.
// The kubebuilder default markers in supabaseproject_types.go must carry
// the same values so that the generated CRD schema matches the webhook
// defaults. Both files are updated together by hack/sync-upstream-images.sh,
// which reads the upstream compose file:
// https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml
const (
	DefaultKongImage       = "kong:2.8.1"
	DefaultAuthImage       = "supabase/gotrue:v2.180.0"
	DefaultPostgRESTImage  = "postgrest/postgrest:v13.0.7"
	DefaultRealtimeImage   = "supabase/realtime:v2.51.11"
	DefaultStorageAPIImage = "supabase/storage-api:v1.32.0"
	DefaultMetaImage       = "supabase/postgres-meta:v0.93.1"
	DefaultStudioImage     = "supabase/studio:2025.10.01-sha-8460121"
)

// DefaultPostgresImage is used by the database init job only. The operator
// requires a user provided PostgreSQL, so this image is intentionally not
// synced from upstream.
const DefaultPostgresImage = "postgres:15-alpine"
