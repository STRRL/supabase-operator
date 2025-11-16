package component

import (
	"github.com/strrl/supabase-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ComponentBuilder interface {
	BuildDeployment(project *v1alpha1.SupabaseProject) (*appsv1.Deployment, error)
	BuildService(project *v1alpha1.SupabaseProject) (*corev1.Service, error)
	Name() string
}
