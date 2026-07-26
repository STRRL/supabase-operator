# Context: Supabase Operator

Glossary of domain terms for this repo. Behavior language only, no implementation detail.

## Terms

### SupabaseProject
The single CRD of this operator. Namespace scoped. Describes one Supabase instance: which external database and object storage to use, and per component settings. The operator deploys and manages all Supabase components (Kong, Auth, PostgREST, Realtime, Storage API, Meta, Studio) from it.

### External dependency
A service the operator needs but never creates or manages: PostgreSQL and S3 compatible object storage. The user provides both and hands the operator connection credentials through Secrets. The operator only validates and consumes them.

### Secret contract
The exact key names a credentials Secret must contain so the operator can consume it.
- Database Secret keys: `host`, `port`, `database`, `username`, `password`
- Storage Secret keys: `endpoint`, `region`, `bucket`, `accessKeyId`, `secretAccessKey`

### Quickstart
The documented golden path for a first time user: install the operator, install external dependencies with well known Helm charts (CloudNativePG for PostgreSQL, MinIO for object storage), create the credential Secrets, apply one SupabaseProject, reach a working Studio via port-forward. Target is a working Studio on a fresh cluster in minutes.

### Database initialization
The one time setup the operator runs against the external PostgreSQL after a project is created: creating the `_supabase` database, Supabase schemas, service roles, and extensions. Requires a superuser grade connection and the `supabase/postgres` image, because Supabase specific extensions must be present.

### dev scaffolding
The manifests under `dev/` used by maintainers for local development. Not user facing, not part of the quickstart.
