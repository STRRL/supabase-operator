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

package database

import (
	"strings"
	"testing"
)

func TestGetInitSQL(t *testing.T) {
	sql := GetInitSQL()

	if len(sql) == 0 {
		t.Error("GetInitSQL() returned empty slice")
	}

	requiredExtensions := []string{"pgcrypto", "uuid-ossp", "pg_stat_statements"}
	for _, ext := range requiredExtensions {
		found := false
		for _, s := range sql {
			if strings.Contains(s, ext) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Extension %s not found in init SQL", ext)
		}
	}

	requiredSchemas := []string{"auth", "storage", "realtime"}
	for _, schema := range requiredSchemas {
		found := false
		for _, s := range sql {
			if strings.Contains(s, "CREATE SCHEMA") && strings.Contains(s, schema) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Schema %s not found in init SQL", schema)
		}
	}

	requiredRoles := []string{"authenticator", "anon", "service_role"}
	for _, role := range requiredRoles {
		found := false
		for _, s := range sql {
			if strings.Contains(s, "CREATE ROLE") && strings.Contains(s, role) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Role %s not found in init SQL", role)
		}
	}

	grantFound := false
	for _, s := range sql {
		if strings.Contains(s, "GRANT") {
			grantFound = true
			break
		}
	}
	if !grantFound {
		t.Error("No GRANT statements found in init SQL")
	}
}

func TestInitConfig(t *testing.T) {
	config := InitConfig{
		Host:     "localhost",
		Port:     "5432",
		Database: "postgres",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
	}

	if config.Host == "" {
		t.Error("Host should not be empty")
	}
	if config.Port == "" {
		t.Error("Port should not be empty")
	}
	if config.Database == "" {
		t.Error("Database should not be empty")
	}
	if config.Username == "" {
		t.Error("Username should not be empty")
	}
	if config.Password == "" {
		t.Error("Password should not be empty")
	}
	if config.SSLMode == "" {
		t.Error("SSLMode should not be empty")
	}
}
