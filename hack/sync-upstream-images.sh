#!/bin/bash
set -euo pipefail

# Sync default component image versions from the upstream Supabase compose file
# https://github.com/supabase/supabase/blob/master/docker/docker-compose.yml
#
# Updates two files that must stay consistent:
#   api/v1alpha1/wellknown_images.go          Go constants used by the webhook
#   api/v1alpha1/supabaseproject_types.go     kubebuilder default markers
# Run "make manifests" afterwards to regenerate the CRD schema.
#
# Usage: ./hack/sync-upstream-images.sh [REF]
# REF defaults to master, a tag or commit of supabase/supabase also works.

UPSTREAM_REPO="https://raw.githubusercontent.com/supabase/supabase"
REF="${1:-master}"
COMPOSE_URL="${UPSTREAM_REPO}/${REF}/docker/docker-compose.yml"

CONSTANTS_FILE="api/v1alpha1/wellknown_images.go"
TYPES_FILE="api/v1alpha1/supabaseproject_types.go"

# Format: "compose_service constant_name struct_name"
# The db service is intentionally not listed, the operator requires a user
# provided PostgreSQL so DefaultPostgresImage is never synced from upstream.
# The kong service is also not listed. Upstream removed Kong from the default
# compose file (supabase/supabase PR #48153, Envoy is the default gateway
# now), so DefaultKongImage is pinned by hand like DefaultPostgresImage.
MAPPINGS=(
    "auth DefaultAuthImage AuthConfig"
    "rest DefaultPostgRESTImage PostgRESTConfig"
    "realtime DefaultRealtimeImage RealtimeConfig"
    "storage DefaultStorageAPIImage StorageAPIConfig"
    "meta DefaultMetaImage MetaConfig"
    "studio DefaultStudioImage StudioConfig"
)

compose_file="$(mktemp)"
trap 'rm -f "$compose_file"' EXIT

echo "Fetching ${COMPOSE_URL}..."
curl -fsSL "$COMPOSE_URL" -o "$compose_file"

lookup_image() {
    awk -v svc="$1" '
        /^  [a-zA-Z0-9_-]+:$/ { current = substr($1, 1, length($1) - 1) }
        current == svc && $1 == "image:" { print $2; exit }
    ' "$compose_file"
}

current_constant() {
    awk -v name="$1" '$1 == name { gsub(/"/, "", $3); print $3; exit }' "$CONSTANTS_FILE"
}

update_constant() {
    local name="$1" image="$2"
    sed -E -i.bak "s|(${name}[[:space:]]*=[[:space:]]*)\"[^\"]*\"|\1\"${image}\"|" "$CONSTANTS_FILE"
    rm -f "${CONSTANTS_FILE}.bak"
}

update_marker() {
    local struct="$1" image="$2"
    awk -v st="$struct" -v img="$image" '
        $0 ~ "^type " st " struct" { in_struct = 1 }
        in_struct && /^}/ { in_struct = 0 }
        in_struct && /\+kubebuilder:default="/ && !done {
            sub(/\+kubebuilder:default="[^"]*"/, "+kubebuilder:default=\"" img "\"")
            done = 1
        }
        { print }
    ' "$TYPES_FILE" > "${TYPES_FILE}.tmp"
    mv "${TYPES_FILE}.tmp" "$TYPES_FILE"
}

changed=0
missing=()
for entry in "${MAPPINGS[@]}"; do
    read -r service constant struct <<< "$entry"

    new_image="$(lookup_image "$service")"
    if [ -z "$new_image" ]; then
        echo "WARNING: service '${service}' not found in upstream compose file, skipping" >&2
        missing+=("$service")
        continue
    fi

    old_image="$(current_constant "$constant")"
    if [ "$old_image" = "$new_image" ]; then
        echo "  = ${service}: ${old_image} (up to date)"
        continue
    fi

    echo "  * ${service}: ${old_image} to ${new_image}"
    update_constant "$constant" "$new_image"
    update_marker "$struct" "$new_image"
    changed=1
done

if [ "$changed" -eq 0 ]; then
    echo "All component images are up to date."
else
    gofmt -w "$CONSTANTS_FILE" "$TYPES_FILE"
    echo "Done. Run 'make manifests' to regenerate the CRD schema."
fi

# A missing service means the upstream compose structure changed. The other
# components are still synced above, but the run must fail so the workflow
# raises an alert.
if [ ${#missing[@]} -gt 0 ]; then
    echo "ERROR: missing upstream services: ${missing[*]}. Upstream structure changed, needs a human." >&2
    exit 1
fi
