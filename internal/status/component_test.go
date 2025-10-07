package status

import (
	"testing"

	"github.com/strrl/supabase-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewComponentStatus(t *testing.T) {
	status := NewComponentStatus("Running", "kong:2.8.1", 1, 1)

	if status.Phase != "Running" {
		t.Errorf("Expected phase 'Running', got '%s'", status.Phase)
	}

	if status.Version != "kong:2.8.1" {
		t.Errorf("Expected version 'kong:2.8.1', got '%s'", status.Version)
	}

	if status.Replicas != 1 {
		t.Errorf("Expected replicas 1, got %d", status.Replicas)
	}

	if status.ReadyReplicas != 1 {
		t.Errorf("Expected ready replicas 1, got %d", status.ReadyReplicas)
	}

	if !status.Ready {
		t.Error("Expected Ready to be true when replicas match")
	}

	if status.LastUpdateTime == nil {
		t.Error("Expected LastUpdateTime to be set")
	}
}

func TestNewComponentStatus_NotReady(t *testing.T) {
	status := NewComponentStatus("Deploying", "kong:2.8.1", 3, 1)

	if status.Ready {
		t.Error("Expected Ready to be false when replicas don't match")
	}

	if status.ReadyReplicas != 1 {
		t.Errorf("Expected ready replicas 1, got %d", status.ReadyReplicas)
	}

	if status.Replicas != 3 {
		t.Errorf("Expected replicas 3, got %d", status.Replicas)
	}
}

func TestSetComponentStatus(t *testing.T) {
	componentsStatus := v1alpha1.ComponentsStatus{}

	kongStatus := NewComponentStatus("Running", "kong:2.8.1", 1, 1)
	componentsStatus = SetComponentStatus(componentsStatus, "Kong", kongStatus)

	if componentsStatus.Kong.Phase != "Running" {
		t.Errorf("Expected Kong phase 'Running', got '%s'", componentsStatus.Kong.Phase)
	}

	if componentsStatus.Kong.Version != "kong:2.8.1" {
		t.Errorf("Expected Kong version 'kong:2.8.1', got '%s'", componentsStatus.Kong.Version)
	}

	authStatus := NewComponentStatus("Running", "supabase/gotrue:v2.177.0", 1, 1)
	componentsStatus = SetComponentStatus(componentsStatus, "Auth", authStatus)

	if componentsStatus.Auth.Phase != "Running" {
		t.Errorf("Expected Auth phase 'Running', got '%s'", componentsStatus.Auth.Phase)
	}
}

func TestSetComponentStatus_AllComponents(t *testing.T) {
	componentsStatus := v1alpha1.ComponentsStatus{}

	components := map[string]v1alpha1.ComponentStatus{
		"Kong":       NewComponentStatus("Running", "kong:2.8.1", 1, 1),
		"Auth":       NewComponentStatus("Running", "supabase/gotrue:v2.177.0", 1, 1),
		"Realtime":   NewComponentStatus("Running", "supabase/realtime:v2.34.47", 1, 1),
		"PostgREST":  NewComponentStatus("Running", "postgrest/postgrest:v12.2.12", 1, 1),
		"StorageAPI": NewComponentStatus("Running", "supabase/storage-api:v1.25.7", 1, 1),
		"Meta":       NewComponentStatus("Running", "supabase/postgres-meta:v0.91.0", 1, 1),
	}

	for name, status := range components {
		componentsStatus = SetComponentStatus(componentsStatus, name, status)
	}

	if componentsStatus.Kong.Phase != "Running" {
		t.Error("Expected Kong to be Running")
	}
	if componentsStatus.Auth.Phase != "Running" {
		t.Error("Expected Auth to be Running")
	}
	if componentsStatus.Realtime.Phase != "Running" {
		t.Error("Expected Realtime to be Running")
	}
	if componentsStatus.PostgREST.Phase != "Running" {
		t.Error("Expected PostgREST to be Running")
	}
	if componentsStatus.StorageAPI.Phase != "Running" {
		t.Error("Expected StorageAPI to be Running")
	}
	if componentsStatus.Meta.Phase != "Running" {
		t.Error("Expected Meta to be Running")
	}
}

func TestIsComponentReady(t *testing.T) {
	tests := []struct {
		name     string
		status   v1alpha1.ComponentStatus
		expected bool
	}{
		{
			name:     "ready with matching replicas",
			status:   NewComponentStatus("Running", "kong:2.8.1", 1, 1),
			expected: true,
		},
		{
			name:     "not ready with mismatched replicas",
			status:   NewComponentStatus("Deploying", "kong:2.8.1", 3, 1),
			expected: false,
		},
		{
			name:     "ready with zero replicas",
			status:   NewComponentStatus("Running", "kong:2.8.1", 0, 0),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.Ready != tt.expected {
				t.Errorf("Expected Ready=%v, got %v", tt.expected, tt.status.Ready)
			}
		})
	}
}

func TestAreAllComponentsReady(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() v1alpha1.ComponentsStatus
		expected bool
	}{
		{
			name: "all components ready",
			setup: func() v1alpha1.ComponentsStatus {
				cs := v1alpha1.ComponentsStatus{}
				cs = SetComponentStatus(cs, "Kong", NewComponentStatus("Running", "kong:2.8.1", 1, 1))
				cs = SetComponentStatus(cs, "Auth", NewComponentStatus("Running", "supabase/gotrue:v2.177.0", 1, 1))
				cs = SetComponentStatus(cs, "Realtime", NewComponentStatus("Running", "supabase/realtime:v2.34.47", 1, 1))
				cs = SetComponentStatus(cs, "PostgREST", NewComponentStatus("Running", "postgrest/postgrest:v12.2.12", 1, 1))
				cs = SetComponentStatus(cs, "StorageAPI", NewComponentStatus("Running", "supabase/storage-api:v1.25.7", 1, 1))
				cs = SetComponentStatus(cs, "Meta", NewComponentStatus("Running", "supabase/postgres-meta:v0.91.0", 1, 1))
				return cs
			},
			expected: true,
		},
		{
			name: "one component not ready",
			setup: func() v1alpha1.ComponentsStatus {
				cs := v1alpha1.ComponentsStatus{}
				cs = SetComponentStatus(cs, "Kong", NewComponentStatus("Running", "kong:2.8.1", 1, 1))
				cs = SetComponentStatus(cs, "Auth", NewComponentStatus("Deploying", "supabase/gotrue:v2.177.0", 3, 1))
				return cs
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := tt.setup()
			result := AreAllComponentsReady(cs)
			if result != tt.expected {
				t.Errorf("AreAllComponentsReady() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSetComponentCondition(t *testing.T) {
	status := NewComponentStatus("Running", "kong:2.8.1", 1, 1)

	condition := metav1.Condition{
		Type:    "HealthCheck",
		Status:  metav1.ConditionTrue,
		Reason:  "Healthy",
		Message: "Component is healthy",
	}

	status = SetComponentCondition(status, condition)

	if len(status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(status.Conditions))
	}

	if status.Conditions[0].Type != "HealthCheck" {
		t.Errorf("Expected condition type 'HealthCheck', got '%s'", status.Conditions[0].Type)
	}
}
