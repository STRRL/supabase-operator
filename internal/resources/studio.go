package resources

import (
	"fmt"

	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildStudioDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.Studio != nil && project.Spec.Studio.Replicas > 0 {
		replicas = project.Spec.Studio.Replicas
	}

	image := "supabase/studio:2025.10.01-sha-8460121"
	if project.Spec.Studio != nil && project.Spec.Studio.Image != "" {
		image = project.Spec.Studio.Image
	}

	resources := getStudioDefaultResources()
	if project.Spec.Studio != nil && project.Spec.Studio.Resources != nil {
		resources = *project.Spec.Studio.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "studio",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "studio",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	apiURL := fmt.Sprintf("http://%s-kong:8000", project.Name)
	publicURL := apiURL
	if project.Spec.Studio != nil && project.Spec.Studio.PublicURL != "" {
		publicURL = project.Spec.Studio.PublicURL
	}

	metaURL := fmt.Sprintf("http://%s-meta:8080", project.Name)

	env := []corev1.EnvVar{
		{Name: "PORT", Value: "3000"},
		{Name: "HOSTNAME", Value: "0.0.0.0"},
		{Name: "SUPABASE_URL", Value: apiURL},
		{Name: "SUPABASE_PUBLIC_URL", Value: publicURL},
		{Name: "NEXT_PUBLIC_SUPABASE_URL", Value: publicURL},
		{Name: "NEXT_PUBLIC_GOTRUE_URL", Value: fmt.Sprintf("%s/auth/v1", publicURL)},
		{Name: "NEXT_PUBLIC_SITE_URL", Value: publicURL},
		{Name: "STUDIO_PG_META_URL", Value: metaURL},
		{Name: "NEXT_PUBLIC_ENABLE_LOGS", Value: "false"},
		{Name: "NEXT_ANALYTICS_BACKEND_PROVIDER", Value: "postgres"},
		{
			Name: "POSTGRES_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Spec.Database.SecretRef.Name},
					Key:                  "password",
				},
			},
		},
		{
			Name: "SUPABASE_ANON_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "anon-key",
				},
			},
		},
		{
			Name: "NEXT_PUBLIC_SUPABASE_ANON_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "anon-key",
				},
			},
		},
		{
			Name: "SUPABASE_SERVICE_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "service-role-key",
				},
			},
		},
		{
			Name: "AUTH_JWT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "jwt-secret",
				},
			},
		},
		{
			Name: "PG_META_CRYPTO_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "pg-meta-crypto-key",
				},
			},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-studio",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "studio",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 3000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.Studio != nil && len(project.Spec.Studio.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Studio.ExtraEnv...,
		)
	}

	return deployment
}

func BuildStudioService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "studio",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "studio",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-studio",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getStudioDefaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("256Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("512Mi"),
			corev1.ResourceCPU:    resource.MustParse("500m"),
		},
	}
}
