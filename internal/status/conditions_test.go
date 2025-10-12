package status

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewReadyCondition(t *testing.T) {
	condition := NewReadyCondition(metav1.ConditionTrue, "AllReady", "All components ready")

	if condition.Type != ConditionTypeReady {
		t.Errorf("Expected type 'Ready', got '%s'", condition.Type)
	}

	if condition.Status != metav1.ConditionTrue {
		t.Errorf("Expected status True, got %v", condition.Status)
	}

	if condition.Reason != "AllReady" {
		t.Errorf("Expected reason 'AllReady', got '%s'", condition.Reason)
	}

	if condition.Message != "All components ready" {
		t.Errorf("Expected message 'All components ready', got '%s'", condition.Message)
	}

	if condition.LastTransitionTime.IsZero() {
		t.Error("Expected LastTransitionTime to be set")
	}
}

func TestNewProgressingCondition(t *testing.T) {
	condition := NewProgressingCondition(metav1.ConditionTrue, "Reconciling", "Reconciliation in progress")

	if condition.Type != ConditionTypeProgressing {
		t.Errorf("Expected type 'Progressing', got '%s'", condition.Type)
	}

	if condition.Status != metav1.ConditionTrue {
		t.Errorf("Expected status True, got %v", condition.Status)
	}
}

func TestNewAvailableCondition(t *testing.T) {
	condition := NewAvailableCondition(metav1.ConditionTrue, "ServicesReady", "All services available")

	if condition.Type != ConditionTypeAvailable {
		t.Errorf("Expected type 'Available', got '%s'", condition.Type)
	}
}

func TestNewDegradedCondition(t *testing.T) {
	condition := NewDegradedCondition(metav1.ConditionTrue, "ComponentFailed", "Kong deployment failed")

	if condition.Type != ConditionTypeDegraded {
		t.Errorf("Expected type 'Degraded', got '%s'", condition.Type)
	}
}

func TestNewComponentCondition(t *testing.T) {
	tests := []struct {
		name          string
		conditionType string
		expectedType  string
	}{
		{"KongReady", "KongReady", "KongReady"},
		{"AuthReady", "AuthReady", "AuthReady"},
		{"PostgreSQLConnected", "PostgreSQLConnected", "PostgreSQLConnected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := NewComponentCondition(tt.conditionType, metav1.ConditionTrue, "Ready", "Component ready")

			if condition.Type != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, condition.Type)
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	conditions := []metav1.Condition{}

	newCondition := NewReadyCondition(metav1.ConditionFalse, "NotReady", "Components not ready")
	conditions = SetCondition(conditions, newCondition)

	if len(conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(conditions))
	}

	if conditions[0].Type != ConditionTypeReady {
		t.Errorf("Expected condition type 'Ready', got '%s'", conditions[0].Type)
	}

	updatedCondition := NewReadyCondition(metav1.ConditionTrue, "AllReady", "All ready")
	conditions = SetCondition(conditions, updatedCondition)

	if len(conditions) != 1 {
		t.Errorf("Expected 1 condition after update, got %d", len(conditions))
	}

	if conditions[0].Status != metav1.ConditionTrue {
		t.Errorf("Expected status True after update, got %v", conditions[0].Status)
	}

	if conditions[0].Reason != "AllReady" {
		t.Errorf("Expected reason 'AllReady' after update, got '%s'", conditions[0].Reason)
	}
}

func TestSetCondition_MultipleConditions(t *testing.T) {
	conditions := []metav1.Condition{}

	conditions = SetCondition(conditions, NewReadyCondition(metav1.ConditionTrue, "Ready", "Ready"))
	conditions = SetCondition(conditions, NewProgressingCondition(metav1.ConditionFalse, "NotProgressing", "Not progressing"))
	conditions = SetCondition(conditions, NewAvailableCondition(metav1.ConditionTrue, "Available", "Available"))

	if len(conditions) != 3 {
		t.Errorf("Expected 3 conditions, got %d", len(conditions))
	}

	hasReady := false
	hasProgressing := false
	hasAvailable := false

	for _, cond := range conditions {
		switch cond.Type {
		case ConditionTypeReady:
			hasReady = true
		case ConditionTypeProgressing:
			hasProgressing = true
		case ConditionTypeAvailable:
			hasAvailable = true
		}
	}

	if !hasReady || !hasProgressing || !hasAvailable {
		t.Error("Expected all three condition types to be present")
	}
}

func TestIsConditionTrue(t *testing.T) {
	conditions := []metav1.Condition{
		NewReadyCondition(metav1.ConditionTrue, "Ready", "Ready"),
		NewProgressingCondition(metav1.ConditionFalse, "NotProgressing", "Not progressing"),
	}

	if !IsConditionTrue(conditions, ConditionTypeReady) {
		t.Error("Expected Ready condition to be true")
	}

	if IsConditionTrue(conditions, ConditionTypeProgressing) {
		t.Error("Expected Progressing condition to be false")
	}

	if IsConditionTrue(conditions, "NonExistent") {
		t.Error("Expected non-existent condition to be false")
	}
}
