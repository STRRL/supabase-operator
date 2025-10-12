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

package resources

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/database/migrations"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildDatabaseInitConfigMap creates a ConfigMap containing the database initialization SQL files
func BuildDatabaseInitConfigMap(project *v1alpha1.SupabaseProject) *corev1.ConfigMap {
	// Get all SQL files from embedded filesystem
	sqlFiles, err := migrations.GetAllSQLFiles()
	if err != nil {
		// This should never happen since files are embedded at compile time
		panic(err)
	}

	// Add execution order script
	sqlFiles["run-migrations.sh"] = generateMigrationScript()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-db-init",
			Namespace: project.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "db-init",
				"app.kubernetes.io/instance":   project.Name,
				"app.kubernetes.io/component":  "database",
				"app.kubernetes.io/part-of":    "supabase",
				"app.kubernetes.io/managed-by": "supabase-operator",
			},
		},
		Data: sqlFiles,
	}
}

// generateMigrationScript creates a shell script that executes SQL files in order
// This mimics the behavior of docker-compose init scripts
func generateMigrationScript() string {
	return `#!/bin/bash
set -e

echo "Starting Supabase database initialization..."
echo "Database: ${DB_NAME} at ${DB_HOST}:${DB_PORT}"
echo ""

# Set psql variables that upstream SQL files expect
export POSTGRES_USER="${DB_USER}"
export POSTGRES_PASSWORD="${DB_PASSWORD}"
export PGPASSWORD="${DB_PASSWORD}"

# Execute SQL files in order (matching docker-compose behavior)
SQL_FILES=(
	"00-initial-schema.sql"
	"01-roles.sql"
	"02-jwt.sql"
	"03-logs.sql"
	"04-webhooks.sql"
	"05-realtime.sql"
	"06-pooler.sql"
)

for sql_file in "${SQL_FILES[@]}"; do
	echo "Executing: ${sql_file}..."

	# Execute with error handling
	if psql "${DATABASE_URL}" -f "/scripts/${sql_file}" -v ON_ERROR_STOP=0; then
		echo "  ✓ ${sql_file} completed"
	else
		# Some scripts may fail (e.g., database already exists)
		# This is OK - services will handle their own migrations
		echo "  ⚠ ${sql_file} had errors (may be expected)"
	fi
	echo ""
done

echo "Database initialization complete!"
echo ""
echo "Note: Some errors above may be expected (e.g., resources already exist)."
echo "Supabase services will create their own schemas and run migrations on startup."
`
}

// BuildDatabaseInitJob creates a Kubernetes Job that runs the database initialization
func BuildDatabaseInitJob(project *v1alpha1.SupabaseProject) *batchv1.Job {
	// Build SSL mode
	sslMode := project.Spec.Database.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}

	// Job will retry up to 3 times on failure
	backoffLimit := int32(3)
	ttlSecondsAfterFinished := int32(600) // Keep job for 10 minutes after completion

	labels := map[string]string{
		"app.kubernetes.io/name":       "db-init",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "database",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-db-init",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "init",
							Image: "postgres:15-alpine",
							Command: []string{
								"bash",
								"/scripts/run-migrations.sh",
							},
							Env: []corev1.EnvVar{
								{
									Name: "DB_HOST",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Spec.Database.SecretRef.Name,
											},
											Key: "host",
										},
									},
								},
								{
									Name: "DB_PORT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Spec.Database.SecretRef.Name,
											},
											Key: "port",
										},
									},
								},
								{
									Name: "DB_NAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Spec.Database.SecretRef.Name,
											},
											Key: "database",
										},
									},
								},
								{
									Name: "DB_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Spec.Database.SecretRef.Name,
											},
											Key: "username",
										},
									},
								},
								{
									Name: "DB_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Spec.Database.SecretRef.Name,
											},
											Key: "password",
										},
									},
								},
								{
									Name:  "DB_SSL_MODE",
									Value: sslMode,
								},
								{
									Name:  "DATABASE_URL",
									Value: "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)",
								},
								{
									Name:  "PGPASSWORD",
									Value: "$(DB_PASSWORD)",
								},
								// JWT configuration for 02-jwt.sql
								{
									Name: "JWT_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: project.Name + "-jwt",
											},
											Key: "jwt-secret",
										},
									},
								},
								{
									Name:  "JWT_EXP",
									Value: "3600", // Default to 1 hour, matching docker-compose
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "init-scripts",
									MountPath: "/scripts",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "init-scripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: project.Name + "-db-init",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
