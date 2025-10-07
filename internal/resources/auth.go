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

	image := "supabase/gotrue:v2.177.0"
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
