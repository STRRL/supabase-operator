package status

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewComponentStatus(phase, version string, replicas, readyReplicas int32) v1alpha1.ComponentStatus {
	now := metav1.Now()
	return v1alpha1.ComponentStatus{
		Phase:          phase,
		Ready:          replicas == readyReplicas,
		Version:        version,
		ReadyReplicas:  readyReplicas,
		Replicas:       replicas,
		Conditions:     []metav1.Condition{},
		LastUpdateTime: &now,
	}
}

func SetComponentStatus(componentsStatus v1alpha1.ComponentsStatus, component string, status v1alpha1.ComponentStatus) v1alpha1.ComponentsStatus {
	switch component {
	case "Kong":
		componentsStatus.Kong = status
	case "Auth":
		componentsStatus.Auth = status
	case "Realtime":
		componentsStatus.Realtime = status
	case "PostgREST":
		componentsStatus.PostgREST = status
	case "StorageAPI":
		componentsStatus.StorageAPI = status
	case "Meta":
		componentsStatus.Meta = status
	}
	return componentsStatus
}

func AreAllComponentsReady(componentsStatus v1alpha1.ComponentsStatus) bool {
	components := []v1alpha1.ComponentStatus{
		componentsStatus.Kong,
		componentsStatus.Auth,
		componentsStatus.Realtime,
		componentsStatus.PostgREST,
		componentsStatus.StorageAPI,
		componentsStatus.Meta,
	}

	for _, comp := range components {
		if !comp.Ready {
			return false
		}
	}

	return true
}

func SetComponentCondition(status v1alpha1.ComponentStatus, condition metav1.Condition) v1alpha1.ComponentStatus {
	status.Conditions = SetCondition(status.Conditions, condition)
	return status
}

func GetComponentByName(componentsStatus v1alpha1.ComponentsStatus, name string) v1alpha1.ComponentStatus {
	switch name {
	case "Kong":
		return componentsStatus.Kong
	case "Auth":
		return componentsStatus.Auth
	case "Realtime":
		return componentsStatus.Realtime
	case "PostgREST":
		return componentsStatus.PostgREST
	case "StorageAPI":
		return componentsStatus.StorageAPI
	case "Meta":
		return componentsStatus.Meta
	default:
		return v1alpha1.ComponentStatus{}
	}
}
