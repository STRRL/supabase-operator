package component

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/webhook"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type StorageBuilder struct{}

var _ ComponentBuilder = (*StorageBuilder)(nil)

func (b *StorageBuilder) Name() string {
	return "storage"
}

func (b *StorageBuilder) BuildDeployment(project *v1alpha1.SupabaseProject) (*appsv1.Deployment, error) {
	replicas := int32(1)
	if project.Spec.StorageAPI != nil && project.Spec.StorageAPI.Replicas > 0 {
		replicas = project.Spec.StorageAPI.Replicas
	}

	image := webhook.DefaultStorageAPIImage
	if project.Spec.StorageAPI != nil && project.Spec.StorageAPI.Image != "" {
		image = project.Spec.StorageAPI.Image
	}

	resources := getStorageDefaultResources()
	if project.Spec.StorageAPI != nil && project.Spec.StorageAPI.Resources != nil {
		resources = *project.Spec.StorageAPI.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "storage",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "storage-api",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	// Build SSL mode
	sslMode := project.Spec.Database.SSLMode
	if sslMode == "" {
		sslMode = defaultSSLMode
	}

	env := []corev1.EnvVar{
		{
			Name: "ANON_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Name + "-jwt",
					},
					Key: "anon-key",
				},
			},
		},
		{
			Name: "SERVICE_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Name + "-jwt",
					},
					Key: "service-role-key",
				},
			},
		},
		{
			Name: "PGRST_JWT_SECRET",
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
			Value: "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
		},
		{
			Name:  "FILE_SIZE_LIMIT",
			Value: "52428800",
		},
		{
			Name:  "STORAGE_BACKEND",
			Value: "s3",
		},
		{
			Name: "GLOBAL_S3_BUCKET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Spec.Storage.SecretRef.Name,
					},
					Key: "bucket",
				},
			},
		},
		{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Spec.Storage.SecretRef.Name,
					},
					Key: "accessKeyId",
				},
			},
		},
		{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Spec.Storage.SecretRef.Name,
					},
					Key: "secretAccessKey",
				},
			},
		},
		{
			Name: "AWS_DEFAULT_REGION",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Spec.Storage.SecretRef.Name,
					},
					Key: "region",
				},
			},
		},
		{
			Name: "S3_ENDPOINT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: project.Spec.Storage.SecretRef.Name,
					},
					Key: "endpoint",
				},
			},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-storage",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "storage",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 5000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.StorageAPI != nil && len(project.Spec.StorageAPI.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.StorageAPI.ExtraEnv...,
		)
	}

	return deployment, nil
}

func (b *StorageBuilder) BuildService(project *v1alpha1.SupabaseProject) (*corev1.Service, error) {
	labels := map[string]string{
		"app.kubernetes.io/name":       "storage",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "storage-api",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-storage",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}, nil
}

func getStorageDefaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("64Mi"),
			corev1.ResourceCPU:    resource.MustParse("50m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
	}
}
