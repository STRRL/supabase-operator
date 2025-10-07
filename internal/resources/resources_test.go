package resources

import (
	"testing"

	"github.com/strrl/supabase-operator/api/v1alpha1"
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

	deployment := BuildKongDeployment(project)

	if deployment.Name != "test-project-kong" {
		t.Errorf("Expected name 'test-project-kong', got '%s'", deployment.Name)
	}

	if deployment.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", deployment.Namespace)
	}

	if *deployment.Spec.Replicas != 1 {
		t.Errorf("Expected 1 replica, got %d", *deployment.Spec.Replicas)
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != "kong:2.8.1" {
		t.Errorf("Expected image 'kong:2.8.1', got '%s'", deployment.Spec.Template.Spec.Containers[0].Image)
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

	deployment := BuildKongDeployment(project)

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

	service := BuildKongService(project)

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

	deployment := BuildAuthDeployment(project)

	if deployment.Name != "test-project-auth" {
		t.Errorf("Expected name 'test-project-auth', got '%s'", deployment.Name)
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != "supabase/gotrue:v2.177.0" {
		t.Errorf("Expected default image, got '%s'", deployment.Spec.Template.Spec.Containers[0].Image)
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

	deployment := BuildPostgRESTDeployment(project)

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

	deployment := BuildRealtimeDeployment(project)

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

	deployment := BuildStorageDeployment(project)

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

	deployment := BuildMetaDeployment(project)

	if deployment.Name != "test-project-meta" {
		t.Errorf("Expected name 'test-project-meta', got '%s'", deployment.Name)
	}
}
