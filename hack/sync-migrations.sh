#!/bin/bash
set -e

# Sync database initialization SQL files from upstream Supabase
# Usage: ./hack/sync-migrations.sh [VERSION]

UPSTREAM_REPO="https://raw.githubusercontent.com/supabase/supabase"
VERSION="${1:-master}"  # Default to master, or specify a tag like v1.2.3
SQL_DIR="internal/database/migrations/sql"

echo "Syncing SQL migrations from supabase/supabase@${VERSION}..."

# Create directory structure
mkdir -p "${SQL_DIR}"

# List of SQL files from docker/volumes/db/ in execution order
# Based on: https://github.com/supabase/supabase/tree/master/docker/volumes/db
# Format: "upstream_path:local_name"
SQL_FILES=(
    "_supabase.sql:00-initial-schema.sql"
    "roles.sql:01-roles.sql"
    "jwt.sql:02-jwt.sql"
    "logs.sql:03-logs.sql"
    "webhooks.sql:04-webhooks.sql"
    "realtime.sql:05-realtime.sql"
    "pooler.sql:06-pooler.sql"
)

# Download each file
for entry in "${SQL_FILES[@]}"; do
    IFS=':' read -r upstream_path local_name <<< "$entry"
    url="${UPSTREAM_REPO}/${VERSION}/docker/volumes/db/${upstream_path}"
    output="${SQL_DIR}/${local_name}"

    echo "Downloading ${upstream_path} -> ${local_name}..."
    if curl -fsSL "$url" -o "$output"; then
        echo "  ✓ Downloaded"
    else
        echo "  ✗ Failed (file might not exist in this version)"
    fi
done

echo ""
echo "Migration sync complete!"
echo "Files are in: ${SQL_DIR}"
echo ""
echo "Next steps:"
echo "1. Review the downloaded SQL files"
echo "2. Run: git diff ${SQL_DIR}"
echo "3. Update internal/database/migrations/migrations.go if needed"
echo "4. Commit the changes"
