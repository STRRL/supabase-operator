package webhook

// Default image versions for Supabase components
// Upstream reference: https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml
const (
	DefaultKongImage       = "kong:2.8.1"
	DefaultAuthImage       = "supabase/gotrue:v2.180.0"
	DefaultPostgRESTImage  = "postgrest/postgrest:v13.0.7"
	DefaultRealtimeImage   = "supabase/realtime:v2.51.11"
	DefaultStorageAPIImage = "supabase/storage-api:v1.28.0"
	DefaultMetaImage       = "supabase/postgres-meta:v0.93.1"
	DefaultStudioImage     = "supabase/studio:2025.10.01-sha-8460121"
	DefaultPostgresImage   = "postgres:15-alpine"
)
