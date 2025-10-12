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

func BuildRealtimeDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.Realtime != nil && project.Spec.Realtime.Replicas > 0 {
		replicas = project.Spec.Realtime.Replicas
	}

	image := "supabase/realtime:v2.51.11"
	if project.Spec.Realtime != nil && project.Spec.Realtime.Image != "" {
		image = project.Spec.Realtime.Image
	}

	resources := getRealtimeDefaultResources()
	if project.Spec.Realtime != nil && project.Spec.Realtime.Resources != nil {
		resources = *project.Spec.Realtime.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "realtime",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "realtime",
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
			Name:  "DATABASE_URL",
			Value: fmt.Sprintf("postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=%s", sslMode),
		},
		{
			Name: "JWT_SECRET",
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
			Name: "SECRET_KEY_BASE",
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
			Name:  "APP_NAME",
			Value: "realtime",
		},
		{
			Name:  "PORT",
			Value: "4000",
		},
		{
			Name:  "RLIMIT_NOFILE",
			Value: "10000",
		},
		{
			Name:  "SECURE_CHANNELS",
			Value: "true",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-realtime",
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
							Name:      "realtime",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 4000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.Realtime != nil && len(project.Spec.Realtime.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Realtime.ExtraEnv...,
		)
	}

	return deployment
}

func BuildRealtimeService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "realtime",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "realtime",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-realtime",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       4000,
					TargetPort: intstr.FromInt(4000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getRealtimeDefaultResources() corev1.ResourceRequirements {
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
