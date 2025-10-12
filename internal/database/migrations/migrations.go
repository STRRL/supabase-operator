/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package migrations

import (
	"embed"
)

// SQL files are embedded at compile time
// These are synced from upstream Supabase repo using hack/sync-migrations.sh
// Source: https://github.com/supabase/supabase/tree/master/docker/volumes/db

//go:embed sql/*.sql
var sqlFS embed.FS

// SQLFiles defines the execution order of SQL files, matching docker-compose behavior
// These are executed sequentially in the database initialization Job
var SQLFiles = []string{
	"sql/00-initial-schema.sql",
	"sql/01-roles.sql",
	"sql/02-jwt.sql",
	"sql/03-logs.sql",
	"sql/04-webhooks.sql",
	"sql/05-realtime.sql",
	"sql/06-pooler.sql",
}

// GetSQLContent reads a SQL file from the embedded filesystem
func GetSQLContent(filename string) (string, error) {
	data, err := sqlFS.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetAllSQLFiles returns a map of filename -> content for all SQL files
// The returned map uses just the filename (without path) as keys to be compatible with ConfigMap
func GetAllSQLFiles() (map[string]string, error) {
	result := make(map[string]string)
	for _, filename := range SQLFiles {
		content, err := GetSQLContent(filename)
		if err != nil {
			return nil, err
		}
		// Extract just the filename without the sql/ prefix
		// e.g., "sql/00-initial-schema.sql" -> "00-initial-schema.sql"
		key := filename[len("sql/"):]
		result[key] = content
	}
	return result, nil
}
