package resources

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildAuthDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.Auth != nil && project.Spec.Auth.Replicas > 0 {
		replicas = project.Spec.Auth.Replicas
	}

	image := "supabase/gotrue:v2.180.0"
	if project.Spec.Auth != nil && project.Spec.Auth.Image != "" {
		image = project.Spec.Auth.Image
	}

	resources := getAuthDefaultResources()
	if project.Spec.Auth != nil && project.Spec.Auth.Resources != nil {
		resources = *project.Spec.Auth.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "auth",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "authentication",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	// Build SSL mode
	sslMode := project.Spec.Database.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}

	env := []corev1.EnvVar{
		{
			Name:  "API_EXTERNAL_URL",
			Value: "http://localhost:8000",
		},
		{
			Name:  "GOTRUE_SITE_URL",
			Value: "http://localhost:8000",
		},
		{
			Name:  "GOTRUE_API_PORT",
			Value: "9999",
		},
		{
			Name:  "GOTRUE_DB_DRIVER",
			Value: "postgres",
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
			Name:  "GOTRUE_DB_DATABASE_URL",
			Value: "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)",
		},
		{
			Name: "GOTRUE_JWT_SECRET",
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
			Name:  "GOTRUE_JWT_EXP",
			Value: "3600",
		},
		{
			Name:  "GOTRUE_JWT_DEFAULT_GROUP_NAME",
			Value: "authenticated",
		},
		{
			Name:  "GOTRUE_DISABLE_SIGNUP",
			Value: "false",
		},
		{
			Name:  "GOTRUE_EXTERNAL_EMAIL_ENABLED",
			Value: "true",
		},
		{
			Name:  "GOTRUE_MAILER_AUTOCONFIRM",
			Value: "true",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-auth",
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
							Name:      "auth",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9999,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.Auth != nil && len(project.Spec.Auth.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Auth.ExtraEnv...,
		)
	}

	return deployment
}

func BuildAuthService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "auth",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "authentication",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-auth",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       9999,
					TargetPort: intstr.FromInt(9999),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getAuthDefaultResources() corev1.ResourceRequirements {
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
