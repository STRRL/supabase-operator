package resources

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildMetaDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.Meta != nil && project.Spec.Meta.Replicas > 0 {
		replicas = project.Spec.Meta.Replicas
	}

	image := "supabase/postgres-meta:v0.91.6"
	if project.Spec.Meta != nil && project.Spec.Meta.Image != "" {
		image = project.Spec.Meta.Image
	}

	resources := getMetaDefaultResources()
	if project.Spec.Meta != nil && project.Spec.Meta.Resources != nil {
		resources = *project.Spec.Meta.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "meta",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "metadata",
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
			Name: "PG_META_DB_HOST",
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
			Name: "PG_META_DB_PORT",
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
			Name: "PG_META_DB_NAME",
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
			Name: "PG_META_DB_USER",
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
			Name: "PG_META_DB_PASSWORD",
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
			Name:  "PG_META_DB_SSL_MODE",
			Value: sslMode,
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-meta",
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
							Name:      "meta",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.Meta != nil && len(project.Spec.Meta.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Meta.ExtraEnv...,
		)
	}

	return deployment
}

func BuildMetaService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "meta",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "metadata",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-meta",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getMetaDefaultResources() corev1.ResourceRequirements {
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
