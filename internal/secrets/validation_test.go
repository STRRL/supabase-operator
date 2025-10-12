package secrets

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateDatabaseSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  *corev1.Secret
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid database secret with all required keys",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"host":     []byte("postgres.example.com"),
					"port":     []byte("5432"),
					"database": []byte("postgres"),
					"username": []byte("user"),
					"password": []byte("pass"),
				},
			},
			wantErr: false,
		},
		{
			name: "missing host key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"port":     []byte("5432"),
					"database": []byte("postgres"),
					"username": []byte("user"),
					"password": []byte("pass"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'host'",
		},
		{
			name: "missing port key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"host":     []byte("postgres.example.com"),
					"database": []byte("postgres"),
					"username": []byte("user"),
					"password": []byte("pass"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'port'",
		},
		{
			name: "missing database key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"host":     []byte("postgres.example.com"),
					"port":     []byte("5432"),
					"username": []byte("user"),
					"password": []byte("pass"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'database'",
		},
		{
			name: "missing username key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"host":     []byte("postgres.example.com"),
					"port":     []byte("5432"),
					"database": []byte("postgres"),
					"password": []byte("pass"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'username'",
		},
		{
			name: "missing password key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{
					"host":     []byte("postgres.example.com"),
					"port":     []byte("5432"),
					"database": []byte("postgres"),
					"username": []byte("user"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'password'",
		},
		{
			name: "empty secret data should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: map[string][]byte{},
			},
			wantErr: true,
			errMsg:  "missing required key 'host'",
		},
		{
			name: "nil secret data should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "db-secret",
				},
				Data: nil,
			},
			wantErr: true,
			errMsg:  "missing required key 'host'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseSecret(tt.secret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDatabaseSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateDatabaseSecret() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateStorageSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  *corev1.Secret
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid storage secret with all required keys",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"endpoint":        []byte("s3.example.com"),
					"region":          []byte("us-east-1"),
					"bucket":          []byte("supabase"),
					"accessKeyId":     []byte("key"),
					"secretAccessKey": []byte("secret"),
				},
			},
			wantErr: false,
		},
		{
			name: "missing endpoint key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"region":          []byte("us-east-1"),
					"bucket":          []byte("supabase"),
					"accessKeyId":     []byte("key"),
					"secretAccessKey": []byte("secret"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'endpoint'",
		},
		{
			name: "missing region key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"endpoint":        []byte("s3.example.com"),
					"bucket":          []byte("supabase"),
					"accessKeyId":     []byte("key"),
					"secretAccessKey": []byte("secret"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'region'",
		},
		{
			name: "missing bucket key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"endpoint":        []byte("s3.example.com"),
					"region":          []byte("us-east-1"),
					"accessKeyId":     []byte("key"),
					"secretAccessKey": []byte("secret"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'bucket'",
		},
		{
			name: "missing accessKeyId key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"endpoint":        []byte("s3.example.com"),
					"region":          []byte("us-east-1"),
					"bucket":          []byte("supabase"),
					"secretAccessKey": []byte("secret"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'accessKeyId'",
		},
		{
			name: "missing secretAccessKey key should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{
					"endpoint":    []byte("s3.example.com"),
					"region":      []byte("us-east-1"),
					"bucket":      []byte("supabase"),
					"accessKeyId": []byte("key"),
				},
			},
			wantErr: true,
			errMsg:  "missing required key 'secretAccessKey'",
		},
		{
			name: "empty secret data should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: map[string][]byte{},
			},
			wantErr: true,
			errMsg:  "missing required key 'endpoint'",
		},
		{
			name: "nil secret data should fail",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "storage-secret",
				},
				Data: nil,
			},
			wantErr: true,
			errMsg:  "missing required key 'endpoint'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStorageSecret(tt.secret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStorageSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateStorageSecret() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
