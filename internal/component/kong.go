package component

import (
	"strings"

	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// kongDeclarativeConfigPath is where the entrypoint script writes the rendered
// declarative config. It must be writable by the kong user (uid 1001) in the
// official kong/kong image.
const kongDeclarativeConfigPath = "/usr/local/kong/kong.yml"

// kongEntrypointScript mirrors upstream supabase docker/volumes/api/kong-entrypoint.sh.
// The operator does not support opaque API keys yet, so only the legacy
// passthrough branch of the Lua expressions is kept. Environment variable
// substitution uses awk instead of eval/echo to preserve YAML quoting.
const kongEntrypointScript = `#!/bin/sh
export LUA_AUTH_EXPR="\$((headers.authorization ~= nil and headers.authorization:sub(1, 10) ~= 'Bearer sb_' and headers.authorization) or headers.apikey)"
export LUA_RT_WS_EXPR="\$(query_params.apikey)"

awk '{
  result = ""
  rest = $0
  while (match(rest, /\$[A-Za-z_][A-Za-z_0-9]*/)) {
    varname = substr(rest, RSTART + 1, RLENGTH - 1)
    if (varname in ENVIRON) {
      result = result substr(rest, 1, RSTART - 1) ENVIRON[varname]
    } else {
      result = result substr(rest, 1, RSTART + RLENGTH - 1)
    }
    rest = substr(rest, RSTART + RLENGTH)
  }
  print result rest
}' /etc/kong/kong.yml > "$KONG_DECLARATIVE_CONFIG"

exec /entrypoint.sh kong docker-start
`

// kongDeclarativeConfigTemplate mirrors upstream supabase docker/volumes/api/kong.yml.
// {{PROJECT}} is replaced with the SupabaseProject name. Routes for components
// the operator does not deploy (edge functions, analytics) are omitted.
const kongDeclarativeConfigTemplate = `_format_version: '2.1'
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
    username: '$DASHBOARD_USERNAME'
    password: '$DASHBOARD_PASSWORD'

services:
  - name: auth-v1-open
    url: http://{{PROJECT}}-auth:9999/verify
    routes:
      - name: auth-v1-open
        strip_path: true
        paths:
          - /auth/v1/verify
    plugins:
      - name: cors

  - name: auth-v1-open-callback
    url: http://{{PROJECT}}-auth:9999/callback
    routes:
      - name: auth-v1-open-callback
        strip_path: true
        paths:
          - /auth/v1/callback
    plugins:
      - name: cors

  - name: auth-v1-open-authorize
    url: http://{{PROJECT}}-auth:9999/authorize
    routes:
      - name: auth-v1-open-authorize
        strip_path: true
        paths:
          - /auth/v1/authorize
    plugins:
      - name: cors

  - name: auth-v1-open-jwks
    url: http://{{PROJECT}}-auth:9999/.well-known/jwks.json
    routes:
      - name: auth-v1-open-jwks
        strip_path: true
        paths:
          - /auth/v1/.well-known/jwks.json
    plugins:
      - name: cors

  - name: auth-v1-open-sso-acs
    url: http://{{PROJECT}}-auth:9999/sso/saml/acs
    routes:
      - name: auth-v1-open-sso-acs
        strip_path: true
        paths:
          - /auth/v1/sso/saml/acs
    plugins:
      - name: cors

  - name: auth-v1-open-sso-metadata
    url: http://{{PROJECT}}-auth:9999/sso/saml/metadata
    routes:
      - name: auth-v1-open-sso-metadata
        strip_path: true
        paths:
          - /auth/v1/sso/saml/metadata
    plugins:
      - name: cors

  - name: auth-v1
    url: http://{{PROJECT}}-auth:9999/
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
      - name: request-transformer
        config:
          add:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

  - name: rest-v1-openapi
    url: http://{{PROJECT}}-postgrest:3000/
    routes:
      - name: rest-v1-openapi-root
        strip_path: true
        expression: 'http.path == "/rest/v1/"'
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: request-transformer
        config:
          add:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin

  - name: rest-v1
    url: http://{{PROJECT}}-postgrest:3000/
    routes:
      - name: rest-v1-all
        strip_path: true
        paths:
          - /rest/v1/
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: request-transformer
        config:
          add:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

  - name: graphql-v1
    url: http://{{PROJECT}}-postgrest:3000/rpc/graphql
    routes:
      - name: graphql-v1-all
        strip_path: true
        paths:
          - /graphql/v1
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: request-transformer
        config:
          add:
            headers:
              - "Content-Profile: graphql_public"
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

  - name: realtime-v1-ws
    url: http://{{PROJECT}}-realtime:4000/socket
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
      - name: request-transformer
        config:
          add:
            headers:
              - "x-api-key:$LUA_RT_WS_EXPR"
          replace:
            querystring:
              - "apikey:$LUA_RT_WS_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

  - name: realtime-v1-rest-openapi
    url: http://{{PROJECT}}-realtime:4000/api/openapi
    protocol: http
    routes:
      - name: realtime-v1-rest-openapi
        strip_path: true
        paths:
          - /realtime/v1/api/openapi
    plugins:
      - name: request-termination
        config:
          status_code: 403
          message: "Access is forbidden."

  - name: realtime-v1-rest-tenants
    url: http://{{PROJECT}}-realtime:4000/api/tenants
    protocol: http
    routes:
      - name: realtime-v1-rest-tenants
        strip_path: true
        paths:
          - /realtime/v1/api/tenants
    plugins:
      - name: request-termination
        config:
          status_code: 403
          message: "Access is forbidden."

  - name: realtime-v1-rest
    url: http://{{PROJECT}}-realtime:4000/api
    protocol: http
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
      - name: request-transformer
        config:
          add:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon

  - name: storage-v1
    url: http://{{PROJECT}}-storage:5000/
    routes:
      - name: storage-v1-all
        strip_path: true
        paths:
          - /storage/v1/
    plugins:
      - name: cors
      - name: request-transformer
        config:
          add:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
          replace:
            headers:
              - "Authorization: $LUA_AUTH_EXPR"
      - name: post-function
        config:
          access:
            - |
              local auth = kong.request.get_header("authorization")
              if auth == nil or auth == "" or auth:find("^%s*$") then
                kong.service.request.clear_header("authorization")
              end

  - name: well-known-oauth
    url: http://{{PROJECT}}-auth:9999/.well-known/oauth-authorization-server
    routes:
      - name: well-known-oauth
        strip_path: true
        paths:
          - /.well-known/oauth-authorization-server
    plugins:
      - name: cors

  - name: meta
    url: http://{{PROJECT}}-meta:8080/
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

  - name: mcp-blocker
    url: http://{{PROJECT}}-studio:3000/api/mcp
    routes:
      - name: mcp-blocker-route
        strip_path: true
        paths:
          - /api/mcp
    plugins:
      - name: request-termination
        config:
          status_code: 403
          message: "Access is forbidden."

  - name: dashboard
    url: http://{{PROJECT}}-studio:3000/
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
`

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

	image := v1alpha1.DefaultKongImage
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

	plugins := "request-transformer,cors,key-auth,acl,basic-auth,request-termination,ip-restriction,post-function"

	env := []corev1.EnvVar{
		{
			Name:  "KONG_DATABASE",
			Value: "off",
		},
		{
			Name:  "KONG_DECLARATIVE_CONFIG",
			Value: kongDeclarativeConfigPath,
		},
		{
			Name:  "KONG_ROUTER_FLAVOR",
			Value: "expressions",
		},
		{
			Name:  "KONG_PROXY_ACCESS_LOG",
			Value: "/dev/stdout combined",
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
			Name:  "KONG_DNS_NOT_FOUND_TTL",
			Value: "1",
		},
		{
			Name:  "KONG_PLUGINS",
			Value: plugins,
		},
		{
			Name:  "KONG_NGINX_PROXY_PROXY_BUFFER_SIZE",
			Value: "160k",
		},
		{
			Name:  "KONG_NGINX_PROXY_PROXY_BUFFERS",
			Value: "64 160k",
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

	healthProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"kong", "health"},
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		FailureThreshold:    5,
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
							Command: []string{
								"/bin/sh",
								"/etc/kong/kong-entrypoint.sh",
							},
							ReadinessProbe: healthProbe,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"kong", "health"},
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
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

	kongConfig := strings.ReplaceAll(kongDeclarativeConfigTemplate, "{{PROJECT}}", project.Name)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      project.Name + "-kong-config",
			Namespace: project.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"kong.yml":           kongConfig,
			"kong-entrypoint.sh": kongEntrypointScript,
		},
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
