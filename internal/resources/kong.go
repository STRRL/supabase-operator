package resources

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildKongDeployment(project *v1alpha1.SupabaseProject) *appsv1.Deployment {
	replicas := int32(1)
	if project.Spec.Kong != nil && project.Spec.Kong.Replicas > 0 {
		replicas = project.Spec.Kong.Replicas
	}

	image := "kong:2.8.1"
	if project.Spec.Kong != nil && project.Spec.Kong.Image != "" {
		image = project.Spec.Kong.Image
	}

	resources := getKongDefaultResources()
	if project.Spec.Kong != nil && project.Spec.Kong.Resources != nil {
		resources = *project.Spec.Kong.Resources
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "kong",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "api-gateway",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	env := []corev1.EnvVar{
		{
			Name:  "KONG_DATABASE",
			Value: "off",
		},
		{
			Name:  "KONG_DECLARATIVE_CONFIG",
			Value: "/etc/kong/kong.yml",
		},
		{
			Name:  "KONG_PROXY_ACCESS_LOG",
			Value: "/dev/stdout",
		},
		{
			Name:  "KONG_ADMIN_ACCESS_LOG",
			Value: "/dev/stdout",
		},
		{
			Name:  "KONG_PROXY_ERROR_LOG",
			Value: "/dev/stderr",
		},
		{
			Name:  "KONG_ADMIN_ERROR_LOG",
			Value: "/dev/stderr",
		},
		{
			Name:  "KONG_ADMIN_LISTEN",
			Value: "0.0.0.0:8001",
		},
		{
			Name:  "KONG_DNS_ORDER",
			Value: "LAST,A,CNAME",
		},
		{
			Name:  "KONG_PLUGINS",
			Value: "request-transformer,cors,key-auth,acl",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-kong",
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
							Name:      "kong",
							Image:     image,
							Resources: resources,
							Env:       env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "proxy",
									ContainerPort: 8000,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "proxy-ssl",
									ContainerPort: 8443,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "admin",
									ContainerPort: 8001,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kong-config",
									MountPath: "/etc/kong",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kong-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: project.Name + "-kong-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if project.Spec.Kong != nil && len(project.Spec.Kong.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Kong.ExtraEnv...,
		)
	}

	return deployment
}

func BuildKongConfigMap(project *v1alpha1.SupabaseProject) *corev1.ConfigMap {
	labels := map[string]string{
		"app.kubernetes.io/name":       "kong",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "api-gateway",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	kongConfig := `_format_version: "1.1"

services:
  - name: auth-v1-open
    url: http://` + project.Name + `-auth:9999/verify
    routes:
      - name: auth-v1-open
        strip_path: true
        paths:
          - /auth/v1/verify
    plugins:
      - name: cors

  - name: auth-v1-open-callback
    url: http://` + project.Name + `-auth:9999/callback
    routes:
      - name: auth-v1-open-callback
        strip_path: true
        paths:
          - /auth/v1/callback
    plugins:
      - name: cors

  - name: auth-v1-open-authorize
    url: http://` + project.Name + `-auth:9999/authorize
    routes:
      - name: auth-v1-open-authorize
        strip_path: true
        paths:
          - /auth/v1/authorize
    plugins:
      - name: cors

  - name: auth-v1
    url: http://` + project.Name + `-auth:9999/
    routes:
      - name: auth-v1-all
        strip_path: true
        paths:
          - /auth/v1/
    plugins:
      - name: cors

  - name: rest-v1
    url: http://` + project.Name + `-postgrest:3000/
    routes:
      - name: rest-v1-all
        strip_path: true
        paths:
          - /rest/v1/
    plugins:
      - name: cors

  - name: realtime-v1
    url: http://` + project.Name + `-realtime:4000/socket/
    routes:
      - name: realtime-v1-all
        strip_path: true
        paths:
          - /realtime/v1/
    plugins:
      - name: cors

  - name: storage-v1
    url: http://` + project.Name + `-storage:5000/
    routes:
      - name: storage-v1-all
        strip_path: true
        paths:
          - /storage/v1/
    plugins:
      - name: cors

  - name: meta-v1
    url: http://` + project.Name + `-meta:8080/
    routes:
      - name: meta-v1-all
        strip_path: true
        paths:
          - /pg/
    plugins:
      - name: cors
`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-kong-config",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"kong.yml": kongConfig,
		},
	}
}

func BuildKongService(project *v1alpha1.SupabaseProject) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name":       "kong",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "api-gateway",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-kong",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "proxy-ssl",
					Port:       8443,
					TargetPort: intstr.FromInt(8443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getKongDefaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1Gi"),
			corev1.ResourceCPU:    resource.MustParse("250m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("2.5Gi"),
			corev1.ResourceCPU:    resource.MustParse("500m"),
		},
	}
}
