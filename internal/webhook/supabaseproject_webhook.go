package webhook

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
)

// +kubebuilder:object:generate=false
type SupabaseProjectWebhook struct {
	Client client.Client
}

func (r *SupabaseProjectWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&supabasev1alpha1.SupabaseProject{}).
		WithValidator(r).
		WithDefaulter(r).
		Complete()
}

var _ webhook.CustomValidator = &SupabaseProjectWebhook{}
var _ webhook.CustomDefaulter = &SupabaseProjectWebhook{}

var (
	requiredDatabaseSecretKeys = []string{"host", "port", "database", "username", "password"}
	requiredStorageSecretKeys  = []string{"endpoint", "region", "bucket", "accessKeyId", "secretAccessKey"}
)

func (r *SupabaseProjectWebhook) Default(ctx context.Context, obj runtime.Object) error {
	project, ok := obj.(*supabasev1alpha1.SupabaseProject)
	if !ok {
		return fmt.Errorf("expected SupabaseProject, got %T", obj)
	}

	if project.Spec.Kong == nil {
		project.Spec.Kong = &supabasev1alpha1.KongConfig{}
	}
	if project.Spec.Kong.Image == "" {
		project.Spec.Kong.Image = DefaultKongImage
	}

	if project.Spec.Auth == nil {
		project.Spec.Auth = &supabasev1alpha1.AuthConfig{}
	}
	if project.Spec.Auth.Image == "" {
		project.Spec.Auth.Image = DefaultAuthImage
	}

	if project.Spec.PostgREST == nil {
		project.Spec.PostgREST = &supabasev1alpha1.PostgRESTConfig{}
	}
	if project.Spec.PostgREST.Image == "" {
		project.Spec.PostgREST.Image = DefaultPostgRESTImage
	}

	if project.Spec.Realtime == nil {
		project.Spec.Realtime = &supabasev1alpha1.RealtimeConfig{}
	}
	if project.Spec.Realtime.Image == "" {
		project.Spec.Realtime.Image = DefaultRealtimeImage
	}

	if project.Spec.StorageAPI == nil {
		project.Spec.StorageAPI = &supabasev1alpha1.StorageAPIConfig{}
	}
	if project.Spec.StorageAPI.Image == "" {
		project.Spec.StorageAPI.Image = DefaultStorageAPIImage
	}

	if project.Spec.Meta == nil {
		project.Spec.Meta = &supabasev1alpha1.MetaConfig{}
	}
	if project.Spec.Meta.Image == "" {
		project.Spec.Meta.Image = DefaultMetaImage
	}

	if project.Spec.Studio == nil {
		project.Spec.Studio = &supabasev1alpha1.StudioConfig{}
	}
	if project.Spec.Studio.Image == "" {
		project.Spec.Studio.Image = DefaultStudioImage
	}

	return nil
}

func (r *SupabaseProjectWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	project, ok := obj.(*supabasev1alpha1.SupabaseProject)
	if !ok {
		return nil, fmt.Errorf("expected SupabaseProject, got %T", obj)
	}

	// Validate secret references (format only)
	if err := r.validateSecretReferences(project); err != nil {
		return nil, err
	}

	// Validate required secrets and keys
	if err := r.validateSecrets(ctx, project); err != nil {
		return nil, err
	}

	// Validate image references
	if err := r.validateImages(project); err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *SupabaseProjectWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return r.ValidateCreate(ctx, newObj)
}

func (r *SupabaseProjectWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// validateSecretReferences validates the format of secret references (not their existence or content)
func (r *SupabaseProjectWebhook) validateSecretReferences(project *supabasev1alpha1.SupabaseProject) error {
	// Validate database secret reference
	if project.Spec.Database.SecretRef.Name == "" {
		return fmt.Errorf("database.secretRef.name cannot be empty")
	}

	// Validate storage secret reference
	if project.Spec.Storage.SecretRef.Name == "" {
		return fmt.Errorf("storage.secretRef.name cannot be empty")
	}

	return nil
}

func (r *SupabaseProjectWebhook) validateSecrets(ctx context.Context, project *supabasev1alpha1.SupabaseProject) error {
	if r.Client == nil {
		return fmt.Errorf("webhook client is not initialized")
	}

	dbSecret, err := r.getSecret(ctx, project, project.Spec.Database.SecretRef, "database")
	if err != nil {
		return err
	}

	storageSecret, err := r.getSecret(ctx, project, project.Spec.Storage.SecretRef, "storage")
	if err != nil {
		return err
	}

	if err := ensureSecretKeys(dbSecret, requiredDatabaseSecretKeys, "database"); err != nil {
		return err
	}

	if err := ensureSecretKeys(storageSecret, requiredStorageSecretKeys, "storage"); err != nil {
		return err
	}

	return nil
}

func (r *SupabaseProjectWebhook) getSecret(ctx context.Context, project *supabasev1alpha1.SupabaseProject, ref corev1.SecretReference, secretType string) (*corev1.Secret, error) {
	namespace := ref.Namespace
	if namespace == "" {
		namespace = project.Namespace
	}

	var secret corev1.Secret
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: ref.Name}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s secret '%s' not found", secretType, ref.Name)
		}
		return nil, fmt.Errorf("failed to fetch %s secret '%s': %w", secretType, ref.Name, err)
	}

	return &secret, nil
}

func ensureSecretKeys(secret *corev1.Secret, keys []string, secretType string) error {
	if secret.Data == nil {
		return fmt.Errorf("%s secret missing required key '%s'", secretType, keys[0])
	}

	for _, key := range keys {
		value, ok := secret.Data[key]
		if !ok || len(value) == 0 {
			return fmt.Errorf("%s secret missing required key '%s'", secretType, key)
		}
	}

	return nil
}

func (r *SupabaseProjectWebhook) validateImages(project *supabasev1alpha1.SupabaseProject) error {
	imagesToValidate := make(map[string]string)

	if project.Spec.Kong != nil && project.Spec.Kong.Image != "" {
		imagesToValidate["kong"] = project.Spec.Kong.Image
	}
	if project.Spec.Auth != nil && project.Spec.Auth.Image != "" {
		imagesToValidate["auth"] = project.Spec.Auth.Image
	}
	if project.Spec.Realtime != nil && project.Spec.Realtime.Image != "" {
		imagesToValidate["realtime"] = project.Spec.Realtime.Image
	}
	if project.Spec.PostgREST != nil && project.Spec.PostgREST.Image != "" {
		imagesToValidate["postgrest"] = project.Spec.PostgREST.Image
	}
	if project.Spec.StorageAPI != nil && project.Spec.StorageAPI.Image != "" {
		imagesToValidate["storage"] = project.Spec.StorageAPI.Image
	}
	if project.Spec.Meta != nil && project.Spec.Meta.Image != "" {
		imagesToValidate["meta"] = project.Spec.Meta.Image
	}

	for component, image := range imagesToValidate {
		if strings.Contains(image, " ") {
			return fmt.Errorf("invalid image reference format")
		}

		if !strings.Contains(image, ":") {
			return fmt.Errorf("image must include tag (e.g., 'kong:2.8.1')")
		}

		_ = component
	}

	return nil
}
