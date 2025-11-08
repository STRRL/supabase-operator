package webhook

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
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

	if err := r.validateSecretExists(ctx, project.Namespace, project.Spec.Database.SecretRef.Name); err != nil {
		return nil, fmt.Errorf("database secret '%s' not found", project.Spec.Database.SecretRef.Name)
	}

	if err := r.validateSecretExists(ctx, project.Namespace, project.Spec.Storage.SecretRef.Name); err != nil {
		return nil, fmt.Errorf("storage secret '%s' not found", project.Spec.Storage.SecretRef.Name)
	}

	dbSecret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      project.Spec.Database.SecretRef.Name,
	}, dbSecret); err == nil {
		if err := r.validateDatabaseSecretKeys(dbSecret); err != nil {
			return nil, err
		}
	}

	storageSecret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      project.Spec.Storage.SecretRef.Name,
	}, storageSecret); err == nil {
		if err := r.validateStorageSecretKeys(storageSecret); err != nil {
			return nil, err
		}
	}

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

func (r *SupabaseProjectWebhook) validateSecretExists(ctx context.Context, namespace, name string) error {
	secret := &corev1.Secret{}
	return r.Client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
}

func (r *SupabaseProjectWebhook) validateDatabaseSecretKeys(secret *corev1.Secret) error {
	requiredKeys := []string{"host", "port", "database", "username", "password"}
	for _, key := range requiredKeys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("database secret missing required key '%s'", key)
		}
	}
	return nil
}

func (r *SupabaseProjectWebhook) validateStorageSecretKeys(secret *corev1.Secret) error {
	requiredKeys := []string{"endpoint", "region", "bucket", "accessKeyId", "secretAccessKey"}
	for _, key := range requiredKeys {
		if _, ok := secret.Data[key]; !ok {
			return fmt.Errorf("storage secret missing required key '%s'", key)
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
