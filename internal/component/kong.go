package component

import (
	"fmt"
	"strings"

	"github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/webhook"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type KongBuilder struct{}

var _ ComponentBuilder = (*KongBuilder)(nil)

func (b *KongBuilder) Name() string {
	return "kong"
}

func (b *KongBuilder) BuildDeployment(project *v1alpha1.SupabaseProject) (*appsv1.Deployment, error) {
	replicas := int32(1)
	if project.Spec.Kong != nil && project.Spec.Kong.Replicas > 0 {
		replicas = project.Spec.Kong.Replicas
	}

	image := webhook.DefaultKongImage
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

	plugins := "request-transformer,cors,key-auth,acl,basic-auth"
	declConfigPath := "/tmp/kong.yml"

	env := []corev1.EnvVar{
		{
			Name:  "KONG_DATABASE",
			Value: "off",
		},
		{
			Name:  "KONG_DECLARATIVE_CONFIG",
			Value: declConfigPath,
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
			Value: plugins,
		},
	}

	env = append(env,
		corev1.EnvVar{
			Name: "SUPABASE_ANON_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "anon-key",
				},
			},
		},
		corev1.EnvVar{
			Name: "SUPABASE_SERVICE_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: project.Name + "-jwt"},
					Key:                  "service-role-key",
				},
			},
		},
	)

	usernameEnv := corev1.EnvVar{Name: "DASHBOARD_USERNAME", Value: "supabase"}
	passwordEnv := corev1.EnvVar{Name: "DASHBOARD_PASSWORD", Value: "this_password_is_insecure_and_should_be_updated"}
	if project.Spec.Studio != nil && project.Spec.Studio.DashboardBasicAuthSecretRef != nil {
		secretRef := project.Spec.Studio.DashboardBasicAuthSecretRef
		usernameEnv = corev1.EnvVar{
			Name: "DASHBOARD_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name},
					Key:                  "username",
				},
			},
		}
		passwordEnv = corev1.EnvVar{
			Name: "DASHBOARD_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name},
					Key:                  "password",
				},
			},
		}
	}
	env = append(env, usernameEnv, passwordEnv)

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

	deployment.Spec.Template.Spec.Containers[0].Command = []string{
		"bash",
		"-lc",
		"eval \"echo \\\"$$(cat /etc/kong/kong.yml)\\\"\" > /tmp/kong.yml && export KONG_DECLARATIVE_CONFIG=/tmp/kong.yml && /docker-entrypoint.sh kong docker-start",
	}

	if project.Spec.Kong != nil && len(project.Spec.Kong.ExtraEnv) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			project.Spec.Kong.ExtraEnv...,
		)
	}

	return deployment, nil
}

func BuildKongConfigMap(project *v1alpha1.SupabaseProject) *corev1.ConfigMap {
	labels := map[string]string{
		"app.kubernetes.io/name":       "kong",
		"app.kubernetes.io/instance":   project.Name,
		"app.kubernetes.io/component":  "api-gateway",
		"app.kubernetes.io/part-of":    "supabase",
		"app.kubernetes.io/managed-by": "supabase-operator",
	}

	var builder strings.Builder
	builder.WriteString(`_format_version: '2.1'
_transform: true

consumers:
  - username: DASHBOARD
  - username: anon
    keyauth_credentials:
      - key: $SUPABASE_ANON_KEY
  - username: service_role
    keyauth_credentials:
      - key: $SUPABASE_SERVICE_KEY

acls:
  - consumer: anon
    group: anon
  - consumer: service_role
    group: admin

basicauth_credentials:
  - consumer: DASHBOARD
    username: $DASHBOARD_USERNAME
    password: $DASHBOARD_PASSWORD

services:
`)

	builder.WriteString(fmt.Sprintf(`  - name: auth-v1-open
    url: http://%s-auth:9999/verify
    routes:
      - name: auth-v1-open
        strip_path: true
        paths:
          - /auth/v1/verify
    plugins:
      - name: cors

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: auth-v1-open-callback
    url: http://%s-auth:9999/callback
    routes:
      - name: auth-v1-open-callback
        strip_path: true
        paths:
          - /auth/v1/callback
    plugins:
      - name: cors

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: auth-v1-open-authorize
    url: http://%s-auth:9999/authorize
    routes:
      - name: auth-v1-open-authorize
        strip_path: true
        paths:
          - /auth/v1/authorize
    plugins:
      - name: cors

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: auth-v1
    url: http://%s-auth:9999/
    routes:
      - name: auth-v1-all
        strip_path: true
        paths:
          - /auth/v1/
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: rest-v1
    url: http://%s-postgrest:3000/
    routes:
      - name: rest-v1-all
        strip_path: true
        paths:
          - /rest/v1/
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: true
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: graphql-v1
    url: http://%s-postgrest:3000/rpc/graphql
    routes:
      - name: graphql-v1-all
        strip_path: true
        paths:
          - /graphql/v1
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: true
      - name: request-transformer
        config:
          add:
            headers:
              - Content-Profile:graphql_public
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: realtime-v1-ws
    url: http://%s-realtime:4000/socket
    protocol: ws
    routes:
      - name: realtime-v1-ws
        strip_path: true
        paths:
          - /realtime/v1/
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: realtime-v1-rest
    url: http://%s-realtime:4000/api
    routes:
      - name: realtime-v1-rest
        strip_path: true
        paths:
          - /realtime/v1/api
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: storage-v1
    url: http://%s-storage:5000/
    routes:
      - name: storage-v1-all
        strip_path: true
        paths:
          - /storage/v1/
    plugins:
      - name: cors

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: meta
    url: http://%s-meta:8080/
    routes:
      - name: meta-all
        strip_path: true
        paths:
          - /pg/
    plugins:
      - name: key-auth
        config:
          hide_credentials: false
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin

`, project.Name))
	builder.WriteString(fmt.Sprintf(`  - name: dashboard
    url: http://%s-studio:3000/
    routes:
      - name: dashboard-all
        strip_path: true
        paths:
          - /
    plugins:
      - name: cors
      - name: basic-auth
        config:
          hide_credentials: true
`, project.Name))

	kongConfig := builder.String()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-kong-config",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{"kong.yml": kongConfig},
	}
}

func (b *KongBuilder) BuildService(project *v1alpha1.SupabaseProject) (*corev1.Service, error) {
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
	}, nil
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
