package component

import (
	"strings"
	"testing"

	"github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/webhook"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildKongDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &KongBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-kong" {
		t.Errorf("Expected name 'test-project-kong', got '%s'", deployment.Name)
	}

	if deployment.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", deployment.Namespace)
	}

	if *deployment.Spec.Replicas != 1 {
		t.Errorf("Expected 1 replica, got %d", *deployment.Spec.Replicas)
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != webhook.DefaultKongImage {
		t.Errorf("Expected image '%s', got '%s'", webhook.DefaultKongImage, deployment.Spec.Template.Spec.Containers[0].Image)
	}

	resources := deployment.Spec.Template.Spec.Containers[0].Resources
	if resources.Limits.Memory().Cmp(resource.MustParse("2.5Gi")) != 0 {
		t.Errorf("Expected memory limit 2.5Gi, got %v", resources.Limits.Memory())
	}
}

func TestBuildKongDeployment_CustomConfig(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
			Kong: &v1alpha1.KongConfig{
				Image:    "kong:3.0.0",
				Replicas: 3,
				Resources: &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				},
			},
		},
	}

	builder := &KongBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if *deployment.Spec.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", *deployment.Spec.Replicas)
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != "kong:3.0.0" {
		t.Errorf("Expected image 'kong:3.0.0', got '%s'", deployment.Spec.Template.Spec.Containers[0].Image)
	}

	resources := deployment.Spec.Template.Spec.Containers[0].Resources
	if resources.Limits.Memory().Cmp(resource.MustParse("4Gi")) != 0 {
		t.Errorf("Expected memory limit 4Gi, got %v", resources.Limits.Memory())
	}
}

func TestBuildKongService(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
	}

	builder := &KongBuilder{}
	service, err := builder.BuildService(project)
	if err != nil {
		t.Fatalf("Failed to build service: %v", err)
	}

	if service.Name != "test-project-kong" {
		t.Errorf("Expected name 'test-project-kong', got '%s'", service.Name)
	}

	if service.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Errorf("Expected type ClusterIP, got %v", service.Spec.Type)
	}

	if len(service.Spec.Ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(service.Spec.Ports))
	}
}

func TestBuildKongWithDashboardBasicAuth(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
			Studio: &v1alpha1.StudioConfig{
				DashboardBasicAuthSecretRef: &corev1.SecretReference{Name: "dashboard-creds"},
			},
		},
	}

	builder := &KongBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}
	container := deployment.Spec.Template.Spec.Containers[0]

	var pluginValue string
	var hasUsernameEnv, hasPasswordEnv bool
	var hasAnonKeyEnv, hasServiceKeyEnv bool
	for _, env := range container.Env {
		switch env.Name {
		case "KONG_PLUGINS":
			pluginValue = env.Value
		case "DASHBOARD_USERNAME":
			hasUsernameEnv = env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key == "username"
		case "DASHBOARD_PASSWORD":
			hasPasswordEnv = env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key == "password"
		case "SUPABASE_ANON_KEY":
			hasAnonKeyEnv = env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key == "anon-key"
		case "SUPABASE_SERVICE_KEY":
			hasServiceKeyEnv = env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key == "service-role-key"
		}
	}

	if !strings.Contains(pluginValue, "basic-auth") {
		t.Fatalf("expected KONG_PLUGINS to include basic-auth, got %s", pluginValue)
	}

	if !hasUsernameEnv || !hasPasswordEnv || !hasAnonKeyEnv || !hasServiceKeyEnv {
		t.Fatalf("expected env vars to source dashboard and key credentials from secrets")
	}

	if len(container.Command) == 0 {
		t.Fatalf("expected custom command to render kong config when dashboard auth enabled")
	}

	configMap := BuildKongConfigMap(project)
	config := configMap.Data["kong.yml"]

	checks := []string{
		"basicauth_credentials",
		"keyauth_credentials",
		"acls:",
		"protocol: ws",
		"- name: graphql-v1",
	}
	for _, token := range checks {
		if !strings.Contains(config, token) {
			t.Fatalf("expected kong config to include %s, got: %s", token, config)
		}
	}

	if !strings.Contains(config, "- name: dashboard") {
		t.Fatalf("expected kong config to include dashboard service, got: %s", config)
	}
}

func TestBuildAuthDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &AuthBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-auth" {
		t.Errorf("Expected name 'test-project-auth', got '%s'", deployment.Name)
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != webhook.DefaultAuthImage {
		t.Errorf("Expected default image '%s', got '%s'", webhook.DefaultAuthImage, deployment.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestBuildPostgRESTDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &PostgRESTBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-postgrest" {
		t.Errorf("Expected name 'test-project-postgrest', got '%s'", deployment.Name)
	}
}

func TestBuildRealtimeDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &RealtimeBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-realtime" {
		t.Errorf("Expected name 'test-project-realtime', got '%s'", deployment.Name)
	}
}

func TestBuildStorageDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &StorageBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-storage" {
		t.Errorf("Expected name 'test-project-storage', got '%s'", deployment.Name)
	}
}

func TestBuildMetaDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
		},
	}

	builder := &MetaBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-meta" {
		t.Errorf("Expected name 'test-project-meta', got '%s'", deployment.Name)
	}
}

func TestBuildStudioDeployment(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
		Spec: v1alpha1.SupabaseProjectSpec{
			ProjectID: "test",
			Database: v1alpha1.DatabaseConfig{
				SecretRef: corev1.SecretReference{Name: "postgres-config"},
			},
		},
	}

	builder := &StudioBuilder{}
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		t.Fatalf("Failed to build deployment: %v", err)
	}

	if deployment.Name != "test-project-studio" {
		t.Errorf("Expected name 'test-project-studio', got '%s'", deployment.Name)
	}

	container := deployment.Spec.Template.Spec.Containers[0]
	if container.Image != webhook.DefaultStudioImage {
		t.Errorf("Expected default image '%s', got '%s'", webhook.DefaultStudioImage, container.Image)
	}
}

func TestBuildStudioService(t *testing.T) {
	project := &v1alpha1.SupabaseProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-project",
			Namespace: "default",
		},
	}

	builder := &StudioBuilder{}
	service, err := builder.BuildService(project)
	if err != nil {
		t.Fatalf("Failed to build service: %v", err)
	}

	if service.Name != "test-project-studio" {
		t.Errorf("Expected name 'test-project-studio', got '%s'", service.Name)
	}

	if len(service.Spec.Ports) != 1 {
		t.Errorf("Expected 1 port, got %d", len(service.Spec.Ports))
	}
}
