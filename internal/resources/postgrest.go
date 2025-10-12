package resources

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildPostgRESTDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.PostgREST != nil && project.Spec.PostgREST.Replicas > 0 {
		replicas = project.Spec.PostgREST.Replicas
	}

	image := "postgrest/postgrest:v13.0.7"
	if project.Spec.PostgREST != nil && project.Spec.PostgREST.Image != "" {
		image = project.Spec.PostgREST.Image
	}

	resources := getPostgRESTDefaultResources()
	if project.Spec.PostgREST != nil && project.Spec.PostgREST.Resources != nil {
		resources = *project.Spec.PostgREST.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "postgrest",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "rest-api",
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
			Name:  "PGRST_DB_URI",
			Value: "postgres://authenticator:$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=" + sslMode,
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
			Name:  "PGRST_DB_ANON_ROLE",
			Value: "anon",
		},
		{
			Name:  "PGRST_DB_SCHEMA",
			Value: "public",
		},
		{
			Name:  "PGRST_DB_EXTRA_SEARCH_PATH",
			Value: "public",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-postgrest",
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
							Name:      "postgrest",
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

	if project.Spec.PostgREST != nil && len(project.Spec.PostgREST.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.PostgREST.ExtraEnv...,
		)
	}

	return deployment
}

func BuildPostgRESTService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "postgrest",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "rest-api",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-postgrest",
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

func getPostgRESTDefaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("256Mi"),
			corev1.ResourceCPU:    resource.MustParse("200m"),
		},
	}
}
