package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeReady       = "Ready"
	ConditionTypeProgressing = "Progressing"
	ConditionTypeAvailable   = "Available"
	ConditionTypeDegraded    = "Degraded"

	ConditionTypeKongReady       = "KongReady"
	ConditionTypeAuthReady       = "AuthReady"
	ConditionTypeRealtimeReady   = "RealtimeReady"
	ConditionTypePostgRESTReady  = "PostgRESTReady"
	ConditionTypeStorageAPIReady = "StorageAPIReady"
	ConditionTypeMetaReady       = "MetaReady"

	ConditionTypePostgreSQLConnected = "PostgreSQLConnected"
	ConditionTypeS3Connected         = "S3Connected"
	ConditionTypeSecretsReady        = "SecretsReady"
	ConditionTypeNetworkReady        = "NetworkReady"
)

func NewReadyCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return newCondition(ConditionTypeReady, status, reason, message)
}

func NewProgressingCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return newCondition(ConditionTypeProgressing, status, reason, message)
}

func NewAvailableCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return newCondition(ConditionTypeAvailable, status, reason, message)
}

func NewDegradedCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return newCondition(ConditionTypeDegraded, status, reason, message)
}

func NewComponentCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return newCondition(conditionType, status, reason, message)
}

func newCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: 0,
	}
}

func SetCondition(conditions []metav1.Condition, newCondition metav1.Condition) []metav1.Condition {
	for i, cond := range conditions {
		if cond.Type == newCondition.Type {
			if cond.Status != newCondition.Status || cond.Reason != newCondition.Reason || cond.Message != newCondition.Message {
				newCondition.LastTransitionTime = metav1.Now()
			} else {
				newCondition.LastTransitionTime = cond.LastTransitionTime
			}
			conditions[i] = newCondition
			return conditions
		}
	}
	conditions = append(conditions, newCondition)
	return conditions
}

func IsConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	for _, cond := range conditions {
		if cond.Type == conditionType {
			return cond.Status == metav1.ConditionTrue
		}
	}
	return false
}

func GetCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i, cond := range conditions {
		if cond.Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
