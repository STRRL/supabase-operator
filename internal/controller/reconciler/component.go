package reconciler

import (
	"context"
	"fmt"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/component"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ComponentReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ComponentReconciler) ReconcileComponent(
	ctx context.Context,
	project *supabasev1alpha1.SupabaseProject,
	builder component.ComponentBuilder,
) error {
	deployment, err := builder.BuildDeployment(project)
	if err != nil {
		return fmt.Errorf("failed to build %s deployment: %w", builder.Name(), err)
	}
	if err := controllerutil.SetControllerReference(project, deployment, r.Scheme); err != nil {
		return err
	}

	existingDeployment := &appsv1.Deployment{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: deployment.Namespace, Name: deployment.Name}, existingDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Client.Create(ctx, deployment); err != nil {
				return fmt.Errorf("failed to create %s deployment: %w", builder.Name(), err)
			}
		} else {
			return err
		}
	} else {
		existingDeployment.Spec = deployment.Spec
		if err := r.Client.Update(ctx, existingDeployment); err != nil {
			return fmt.Errorf("failed to update %s deployment: %w", builder.Name(), err)
		}
	}

	service, err := builder.BuildService(project)
	if err != nil {
		return fmt.Errorf("failed to build %s service: %w", builder.Name(), err)
	}
	if err := controllerutil.SetControllerReference(project, service, r.Scheme); err != nil {
		return err
	}

	existingService := &corev1.Service{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: service.Namespace, Name: service.Name}, existingService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Client.Create(ctx, service); err != nil {
				return fmt.Errorf("failed to create %s service: %w", builder.Name(), err)
			}
		} else {
			return err
		}
	}

	return nil
}
