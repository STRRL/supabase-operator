package v1alpha1

// Default image versions for Supabase components.
//
// This file is the single source of truth for component image defaults.
// The kubebuilder default markers in supabaseproject_types.go must carry
// the same values so that the generated CRD schema matches the webhook
// defaults. The constants in the block below are updated together with
// the markers by hack/sync-upstream-images.sh, which reads the upstream
// compose file:
// https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml
//
// DefaultKongImage and DefaultPostgresImage further down are pinned by
// hand and are not touched by the sync script.
const (
	DefaultAuthImage       = "supabase/gotrue:v2.189.0"
	DefaultPostgRESTImage  = "postgrest/postgrest:v14.12"
	DefaultRealtimeImage   = "supabase/realtime:v2.102.3"
	DefaultStorageAPIImage = "supabase/storage-api:v1.60.4"
	DefaultMetaImage       = "supabase/postgres-meta:v0.96.6"
	DefaultStudioImage     = "supabase/studio:2026.08.03-sha-022b374"
)

// DefaultKongImage is pinned by hand and not synced from upstream.
// Upstream removed Kong from the default compose file (supabase/supabase
// PR #48153, Envoy is the default gateway now). Kong still exists upstream
// as an optional override in docker-compose.kong.yml (currently
// kong/kong:3.9.3), but this operator keeps its own pinned version.
const DefaultKongImage = "kong/kong:3.9.1"

// DefaultPostgresImage is used by the database init job only. The operator
// requires a user provided PostgreSQL, so this image is intentionally not
// synced from upstream.
const DefaultPostgresImage = "postgres:15-alpine"
