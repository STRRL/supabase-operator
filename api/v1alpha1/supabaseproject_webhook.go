package v1alpha1

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
)

// +kubebuilder:object:generate=false
type SupabaseProjectWebhook struct {
	Client client.Client
}

func (r *SupabaseProjectWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&SupabaseProject{}).
		WithValidator(r).
		WithDefaulter(r).
		Complete()
}

var _ webhook.CustomValidator = &SupabaseProjectWebhook{}
var _ webhook.CustomDefaulter = &SupabaseProjectWebhook{}

func (r *SupabaseProjectWebhook) Default(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*SupabaseProject)
	if !ok {
		return fmt.Errorf("expected SupabaseProject, got %T", obj)
	}

	return nil
}

func (r *SupabaseProjectWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	project, ok := obj.(*SupabaseProject)
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

func (r *SupabaseProjectWebhook) validateImages(project *SupabaseProject) error {
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
