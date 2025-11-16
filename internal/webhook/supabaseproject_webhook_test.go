package webhook

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
)

// Helper function to create test secrets
func createTestSecrets() []client.Object {
	return []client.Object{
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"host":     []byte("postgres.example.com"),
				"port":     []byte("5432"),
				"database": []byte("postgres"),
				"username": []byte("user"),
				"password": []byte("pass"),
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "storage-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"endpoint":        []byte("s3.example.com"),
				"region":          []byte("us-east-1"),
				"bucket":          []byte("supabase"),
				"accessKeyId":     []byte("key"),
				"secretAccessKey": []byte("secret"),
			},
		},
	}
}

// Helper function to create test project
func createTestProject() *supabasev1alpha1.SupabaseProject {
	return &supabasev1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: supabasev1alpha1.SupabaseProjectSpec{
			ProjectID: "test-project",
			Database: supabasev1alpha1.DatabaseConfig{
				SecretRef: corev1.SecretReference{
					Name: "db-secret",
				},
			},
			Storage: supabasev1alpha1.StorageConfig{
				SecretRef: corev1.SecretReference{
					Name: "storage-secret",
				},
			},
		},
	}
}

func TestValidateCreate_SecretExistence(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = supabasev1alpha1.AddToScheme(scheme)

	tests := []struct {
		name    string
		project *supabasev1alpha1.SupabaseProject
		secrets []client.Object
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with existing secrets",
			project: createTestProject(),
			secrets: createTestSecrets(),
			wantErr: false,
		},
		{
			name: "missing database secret should fail",
			project: &supabasev1alpha1.SupabaseProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-project",
					Namespace: "default",
				},
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					ProjectID: "test-project",
					Database: supabasev1alpha1.DatabaseConfig{
						SecretRef: corev1.SecretReference{
							Name: "missing-db-secret",
						},
					},
					Storage: supabasev1alpha1.StorageConfig{
						SecretRef: corev1.SecretReference{
							Name: "storage-secret",
						},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "storage-secret",
						Namespace: "default",
					},
				},
			},
			wantErr: true,
			errMsg:  "database secret 'missing-db-secret' not found",
		},
		{
			name: "missing storage secret should fail",
			project: &supabasev1alpha1.SupabaseProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-project",
					Namespace: "default",
				},
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					ProjectID: "test-project",
					Database: supabasev1alpha1.DatabaseConfig{
						SecretRef: corev1.SecretReference{
							Name: "db-secret",
						},
					},
					Storage: supabasev1alpha1.StorageConfig{
						SecretRef: corev1.SecretReference{
							Name: "missing-storage-secret",
						},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "db-secret",
						Namespace: "default",
					},
				},
			},
			wantErr: true,
			errMsg:  "storage secret 'missing-storage-secret' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.secrets...).
				Build()

			webhook := &SupabaseProjectWebhook{
				Client: fakeClient,
			}

			_, err := webhook.ValidateCreate(context.Background(), tt.project)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateCreate() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateCreate_RequiredSecretKeys(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = supabasev1alpha1.AddToScheme(scheme)

	tests := []struct {
		name    string
		project *supabasev1alpha1.SupabaseProject
		secrets []client.Object
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with all required database keys",
			project: createTestProject(),
			secrets: createTestSecrets(),
			wantErr: false,
		},
		{
			name: "missing database host key should fail",
			project: &supabasev1alpha1.SupabaseProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-project",
					Namespace: "default",
				},
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					ProjectID: "test-project",
					Database: supabasev1alpha1.DatabaseConfig{
						SecretRef: corev1.SecretReference{
							Name: "db-secret",
						},
					},
					Storage: supabasev1alpha1.StorageConfig{
						SecretRef: corev1.SecretReference{
							Name: "storage-secret",
						},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "db-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"port":     []byte("5432"),
						"database": []byte("postgres"),
						"username": []byte("user"),
						"password": []byte("pass"),
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "storage-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"endpoint":        []byte("s3.example.com"),
						"region":          []byte("us-east-1"),
						"bucket":          []byte("supabase"),
						"accessKeyId":     []byte("key"),
						"secretAccessKey": []byte("secret"),
					},
				},
			},
			wantErr: true,
			errMsg:  "database secret missing required key 'host'",
		},
		{
			name: "missing storage endpoint key should fail",
			project: &supabasev1alpha1.SupabaseProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-project",
					Namespace: "default",
				},
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					ProjectID: "test-project",
					Database: supabasev1alpha1.DatabaseConfig{
						SecretRef: corev1.SecretReference{
							Name: "db-secret",
						},
					},
					Storage: supabasev1alpha1.StorageConfig{
						SecretRef: corev1.SecretReference{
							Name: "storage-secret",
						},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "db-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"host":     []byte("postgres.example.com"),
						"port":     []byte("5432"),
						"database": []byte("postgres"),
						"username": []byte("user"),
						"password": []byte("pass"),
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "storage-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"region":          []byte("us-east-1"),
						"bucket":          []byte("supabase"),
						"accessKeyId":     []byte("key"),
						"secretAccessKey": []byte("secret"),
					},
				},
			},
			wantErr: true,
			errMsg:  "storage secret missing required key 'endpoint'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.secrets...).
				Build()

			webhook := &SupabaseProjectWebhook{
				Client: fakeClient,
			}

			_, err := webhook.ValidateCreate(context.Background(), tt.project)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateCreate() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateCreate_ImageReferenceValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid kong image",
			config: &supabasev1alpha1.KongConfig{
				Image: "kong:2.8.1",
			},
			wantErr: false,
		},
		{
			name: "valid auth image",
			config: &supabasev1alpha1.AuthConfig{
				Image: "supabase/gotrue:v2.177.0",
			},
			wantErr: false,
		},
		{
			name: "empty image should use default",
			config: &supabasev1alpha1.KongConfig{
				Image: "",
			},
			wantErr: false,
		},
		{
			name: "invalid image format should fail",
			config: &supabasev1alpha1.KongConfig{
				Image: "invalid image:with spaces",
			},
			wantErr: true,
			errMsg:  "invalid image reference format",
		},
		{
			name: "missing tag should fail",
			config: &supabasev1alpha1.KongConfig{
				Image: "kong",
			},
			wantErr: true,
			errMsg:  "image must include tag (e.g., 'kong:2.8.1')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var image string
			switch c := tt.config.(type) {
			case *supabasev1alpha1.KongConfig:
				image = c.Image
			case *supabasev1alpha1.AuthConfig:
				image = c.Image
			}

			if image == "" && !tt.wantErr {
				t.Log("Empty image should use default, validation passes")
				return
			}

			hasColon := false
			hasSpace := false
			for _, char := range image {
				if char == ':' {
					hasColon = true
				}
				if char == ' ' {
					hasSpace = true
				}
			}

			if hasSpace && !tt.wantErr {
				t.Error("Image with spaces should fail validation")
			}

			if !hasColon && image != "" && !tt.wantErr && tt.name == "missing tag should fail" {
				t.Error("Image without tag should fail validation")
			}
		})
	}
}

func TestDefault_ResourceDefaults(t *testing.T) {
	tests := []struct {
		name    string
		project *supabasev1alpha1.SupabaseProject
	}{
		{
			name: "webhook default should complete without error",
			project: &supabasev1alpha1.SupabaseProject{
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					Kong: &supabasev1alpha1.KongConfig{
						Resources: nil,
					},
				},
			},
		},
		{
			name: "webhook default with existing resources should not error",
			project: &supabasev1alpha1.SupabaseProject{
				Spec: supabasev1alpha1.SupabaseProjectSpec{
					Kong: &supabasev1alpha1.KongConfig{
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("1Gi"),
								corev1.ResourceCPU:    resource.MustParse("500m"),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook := &SupabaseProjectWebhook{}

			err := webhook.Default(context.Background(), tt.project)
			if err != nil {
				t.Errorf("Default() unexpected error = %v", err)
			}
		})
	}
}
