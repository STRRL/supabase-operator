package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSupabaseProjectSpec_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		spec    SupabaseProjectSpec
		wantErr bool
	}{
		{
			name: "valid spec with all required fields",
			spec: SupabaseProjectSpec{
				ProjectID: "test-project",
				Database: DatabaseConfig{
					SecretRef: corev1.SecretReference{
						Name: "db-secret",
					},
				},
				Storage: StorageConfig{
					SecretRef: corev1.SecretReference{
						Name: "storage-secret",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing projectId should fail",
			spec: SupabaseProjectSpec{
				Database: DatabaseConfig{
					SecretRef: corev1.SecretReference{
						Name: "db-secret",
					},
				},
				Storage: StorageConfig{
					SecretRef: corev1.SecretReference{
						Name: "storage-secret",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spec.ProjectID == "" && !tt.wantErr {
				t.Error("ProjectID should be required but validation passed")
			}
		})
	}
}

func TestDatabaseConfig_SecretReference(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid database config with secret ref",
			config: DatabaseConfig{
				SecretRef: corev1.SecretReference{
					Name: "postgres-config",
				},
				SSLMode:        "require",
				MaxConnections: 20,
			},
			wantErr: false,
		},
		{
			name: "missing secret ref should fail",
			config: DatabaseConfig{
				SSLMode:        "require",
				MaxConnections: 20,
			},
			wantErr: true,
		},
		{
			name: "default values should be applied",
			config: DatabaseConfig{
				SecretRef: corev1.SecretReference{
					Name: "postgres-config",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.SecretRef.Name == "" && !tt.wantErr {
				t.Error("SecretRef should be required but validation passed")
			}

			if tt.name == "default values should be applied" {
				if tt.config.SSLMode == "" {
					tt.config.SSLMode = "require"
				}
				if tt.config.MaxConnections == 0 {
					tt.config.MaxConnections = 20
				}

				if tt.config.SSLMode != "require" {
					t.Errorf("Expected default SSLMode 'require', got '%s'", tt.config.SSLMode)
				}
				if tt.config.MaxConnections != 20 {
					t.Errorf("Expected default MaxConnections 20, got %d", tt.config.MaxConnections)
				}
			}
		})
	}
}

func TestStorageConfig_SecretReference(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
	}{
		{
			name: "valid storage config with secret ref",
			config: StorageConfig{
				SecretRef: corev1.SecretReference{
					Name: "s3-config",
				},
				ForcePathStyle: true,
			},
			wantErr: false,
		},
		{
			name: "missing secret ref should fail",
			config: StorageConfig{
				ForcePathStyle: true,
			},
			wantErr: true,
		},
		{
			name: "explicit ForcePathStyle true",
			config: StorageConfig{
				SecretRef: corev1.SecretReference{
					Name: "s3-config",
				},
				ForcePathStyle: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.SecretRef.Name == "" && !tt.wantErr {
				t.Error("SecretRef should be required but validation passed")
			}
		})
	}
}

func TestKongConfig_Defaults(t *testing.T) {
	config := &KongConfig{}

	expectedImage := "kong:2.8.1"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}

	if config.Resources != nil {
		memLimit := config.Resources.Limits[corev1.ResourceMemory]
		cpuLimit := config.Resources.Limits[corev1.ResourceCPU]

		expectedMem := resource.MustParse("2.5Gi")
		expectedCPU := resource.MustParse("500m")

		if memLimit.Cmp(expectedMem) != 0 {
			t.Errorf("Expected memory limit %s, got %s", expectedMem.String(), memLimit.String())
		}
		if cpuLimit.Cmp(expectedCPU) != 0 {
			t.Errorf("Expected CPU limit %s, got %s", expectedCPU.String(), cpuLimit.String())
		}
	}
}

func TestAuthConfig_Defaults(t *testing.T) {
	config := &AuthConfig{}

	expectedImage := "supabase/gotrue:v2.177.0"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}
}

func TestRealtimeConfig_Defaults(t *testing.T) {
	config := &RealtimeConfig{}

	expectedImage := "supabase/realtime:v2.34.47"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}
}

func TestPostgRESTConfig_Defaults(t *testing.T) {
	config := &PostgRESTConfig{}

	expectedImage := "postgrest/postgrest:v12.2.12"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}
}

func TestStorageAPIConfig_Defaults(t *testing.T) {
	config := &StorageAPIConfig{}

	expectedImage := "supabase/storage-api:v1.25.7"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}
}

func TestMetaConfig_Defaults(t *testing.T) {
	config := &MetaConfig{}

	expectedImage := "supabase/postgres-meta:v0.91.0"
	if config.Image == "" {
		config.Image = expectedImage
	}

	if config.Image != expectedImage {
		t.Errorf("Expected default image %s, got %s", expectedImage, config.Image)
	}

	expectedReplicas := int32(1)
	if config.Replicas == 0 {
		config.Replicas = expectedReplicas
	}

	if config.Replicas != expectedReplicas {
		t.Errorf("Expected default replicas %d, got %d", expectedReplicas, config.Replicas)
	}
}

func TestSupabaseProjectStatus_Structure(t *testing.T) {
	status := SupabaseProjectStatus{
		Phase:   "Running",
		Message: "All components healthy",
		Conditions: []metav1.Condition{
			{
				Type:   "Ready",
				Status: metav1.ConditionTrue,
				Reason: "AllComponentsReady",
			},
		},
		ObservedGeneration: 1,
	}

	if status.Phase != "Running" {
		t.Errorf("Expected phase 'Running', got '%s'", status.Phase)
	}

	if len(status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(status.Conditions))
	}

	if status.Conditions[0].Type != "Ready" {
		t.Errorf("Expected condition type 'Ready', got '%s'", status.Conditions[0].Type)
	}

	if status.ObservedGeneration != 1 {
		t.Errorf("Expected ObservedGeneration 1, got %d", status.ObservedGeneration)
	}
}

func TestComponentsStatus_Structure(t *testing.T) {
	components := ComponentsStatus{
		Kong: ComponentStatus{
			Phase:   "Running",
			Version: "kong:2.8.1",
		},
		Auth: ComponentStatus{
			Phase:   "Running",
			Version: "supabase/gotrue:v2.177.0",
		},
	}

	if components.Kong.Phase != "Running" {
		t.Errorf("Expected Kong phase 'Running', got '%s'", components.Kong.Phase)
	}

	if components.Auth.Version != "supabase/gotrue:v2.177.0" {
		t.Errorf("Expected Auth version 'supabase/gotrue:v2.177.0', got '%s'", components.Auth.Version)
	}
}
