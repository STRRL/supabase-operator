package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/resources"
	"github.com/strrl/supabase-operator/internal/secrets"
	"github.com/strrl/supabase-operator/internal/status"
)

const (
	finalizerName  = "supabase.strrl.dev/finalizer"
	jwtSecretKey   = "jwt-secret"
	anonKeyKey     = "anon-key"
	serviceRoleKey = "service-role-key"
)

type SupabaseProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=supabase.strrl.dev,resources=supabaseprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.strrl.dev,resources=supabaseprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.strrl.dev,resources=supabaseprojects/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *SupabaseProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	project := &supabasev1alpha1.SupabaseProject{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if project.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(project, finalizerName) {
			if err := r.handleDeletion(ctx, project); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(project, finalizerName)
			if err := r.Update(ctx, project); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(project, finalizerName) {
		controllerutil.AddFinalizer(project, finalizerName)
		if err := r.Update(ctx, project); err != nil {
			return ctrl.Result{}, err
		}
	}

	originalProject := project.DeepCopy()

	if project.Status.Phase == "" {
		project.Status.Phase = status.PhasePending
		project.Status.Message = status.GetPhaseMessage(status.PhasePending)
	}

	project.Status.Conditions = status.SetCondition(
		project.Status.Conditions,
		status.NewProgressingCondition(metav1.ConditionTrue, "Reconciling", "Reconciliation in progress"),
	)

	if err := r.validateDependencies(ctx, project); err != nil {
		logger.Error(err, "Failed to validate dependencies")
		project.Status.Phase = status.PhaseFailed
		project.Status.Message = fmt.Sprintf("Dependency validation failed: %v", err)
		project.Status.Conditions = status.SetCondition(
			project.Status.Conditions,
			status.NewReadyCondition(metav1.ConditionFalse, "DependencyValidationFailed", err.Error()),
		)
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Generate JWT secrets first (needed by database init job)
	project.Status.Phase = status.PhaseDeployingSecrets
	project.Status.Message = status.GetPhaseMessage(status.PhaseDeployingSecrets)

	if err := r.ensureJWTSecrets(ctx, project); err != nil {
		logger.Error(err, "Failed to ensure JWT secrets")
		project.Status.Phase = status.PhaseFailed
		project.Status.Message = fmt.Sprintf("Secret generation failed: %v", err)
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Initialize database with required extensions and roles via Kubernetes Job
	project.Status.Phase = status.PhaseInitializingDatabase
	project.Status.Message = status.GetPhaseMessage(status.PhaseInitializingDatabase)

	jobResult, err := r.ensureDatabaseInitJob(ctx, project)
	if err != nil {
		logger.Error(err, "Failed to ensure database init job")
		project.Status.Phase = status.PhaseFailed
		project.Status.Message = fmt.Sprintf("Database initialization job failed: %v", err)
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// If job is still running, requeue and check later
	if jobResult.RequeueAfter > 0 {
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return jobResult, nil
	}

	project.Status.Phase = status.PhaseDeployingComponents

	componentsStatus, err := r.reconcileAllComponents(ctx, project)
	if err != nil {
		return ctrl.Result{}, err
	}

	project.Status.Components = componentsStatus
	project.Status.Phase = status.PhaseRunning
	project.Status.Message = status.GetPhaseMessage(status.PhaseRunning)
	project.Status.Conditions = status.SetCondition(
		project.Status.Conditions,
		status.NewReadyCondition(metav1.ConditionTrue, "AllComponentsReady", "All components are running"),
	)
	project.Status.Conditions = status.SetCondition(
		project.Status.Conditions,
		status.NewProgressingCondition(metav1.ConditionFalse, "ReconciliationComplete", "Reconciliation complete"),
	)
	project.Status.ObservedGeneration = project.Generation
	now := metav1.Now()
	project.Status.LastReconcileTime = &now

	if err := r.Status().Update(ctx, project); err != nil {
		return ctrl.Result{}, err
	}

	if !apiequality(originalProject.Status, project.Status) {
		logger.Info("Successfully reconciled SupabaseProject")
	}

	return ctrl.Result{}, nil
}

func (r *SupabaseProjectReconciler) reconcileAllComponents(ctx context.Context, project *supabasev1alpha1.SupabaseProject) (supabasev1alpha1.ComponentsStatus, error) {
	logger := log.FromContext(ctx)
	componentsStatus := supabasev1alpha1.ComponentsStatus{}

	// Create Kong ConfigMap
	kongConfigMap := resources.BuildKongConfigMap(project)
	if err := controllerutil.SetControllerReference(project, kongConfigMap, r.Scheme); err != nil {
		return componentsStatus, err
	}
	existingKongConfigMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: kongConfigMap.Namespace, Name: kongConfigMap.Name}, existingKongConfigMap); err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, kongConfigMap); err != nil {
			logger.Error(err, "Failed to create Kong ConfigMap")
			return componentsStatus, err
		}
	}

	if err := r.reconcileComponent(ctx, project, "kong", resources.BuildKongDeployment, resources.BuildKongService); err != nil {
		logger.Error(err, "Failed to reconcile Kong")
		return componentsStatus, err
	}
	kongImage := "kong:2.8.1"
	if project.Spec.Kong != nil && project.Spec.Kong.Image != "" {
		kongImage = project.Spec.Kong.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "Kong",
		status.NewComponentStatus(status.PhaseRunning, kongImage, 1, 1))

	if err := r.reconcileComponent(ctx, project, "auth", resources.BuildAuthDeployment, resources.BuildAuthService); err != nil {
		logger.Error(err, "Failed to reconcile Auth")
		return componentsStatus, err
	}
	authImage := "supabase/gotrue:v2.180.0"
	if project.Spec.Auth != nil && project.Spec.Auth.Image != "" {
		authImage = project.Spec.Auth.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "Auth",
		status.NewComponentStatus(status.PhaseRunning, authImage, 1, 1))

	if err := r.reconcileComponent(ctx, project, "postgrest", resources.BuildPostgRESTDeployment, resources.BuildPostgRESTService); err != nil {
		logger.Error(err, "Failed to reconcile PostgREST")
		return componentsStatus, err
	}
	postgrestImage := "postgrest/postgrest:v13.0.7"
	if project.Spec.PostgREST != nil && project.Spec.PostgREST.Image != "" {
		postgrestImage = project.Spec.PostgREST.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "PostgREST",
		status.NewComponentStatus(status.PhaseRunning, postgrestImage, 1, 1))

	if err := r.reconcileComponent(ctx, project, "realtime", resources.BuildRealtimeDeployment, resources.BuildRealtimeService); err != nil {
		logger.Error(err, "Failed to reconcile Realtime")
		return componentsStatus, err
	}
	realtimeImage := "supabase/realtime:v2.51.11"
	if project.Spec.Realtime != nil && project.Spec.Realtime.Image != "" {
		realtimeImage = project.Spec.Realtime.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "Realtime",
		status.NewComponentStatus(status.PhaseRunning, realtimeImage, 1, 1))

	if err := r.reconcileComponent(ctx, project, "storage", resources.BuildStorageDeployment, resources.BuildStorageService); err != nil {
		logger.Error(err, "Failed to reconcile Storage")
		return componentsStatus, err
	}
	storageImage := "supabase/storage-api:v1.28.0"
	if project.Spec.StorageAPI != nil && project.Spec.StorageAPI.Image != "" {
		storageImage = project.Spec.StorageAPI.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "StorageAPI",
		status.NewComponentStatus(status.PhaseRunning, storageImage, 1, 1))

	if err := r.reconcileComponent(ctx, project, "meta", resources.BuildMetaDeployment, resources.BuildMetaService); err != nil {
		logger.Error(err, "Failed to reconcile Meta")
		return componentsStatus, err
	}
	metaImage := "supabase/postgres-meta:v0.102.0"
	if project.Spec.Meta != nil && project.Spec.Meta.Image != "" {
		metaImage = project.Spec.Meta.Image
	}
	componentsStatus = status.SetComponentStatus(componentsStatus, "Meta",
		status.NewComponentStatus(status.PhaseRunning, metaImage, 1, 1))

	return componentsStatus, nil
}

func (r *SupabaseProjectReconciler) validateDependencies(ctx context.Context, project *supabasev1alpha1.SupabaseProject) error {
	dbSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      project.Spec.Database.SecretRef.Name,
	}, dbSecret); err != nil {
		return fmt.Errorf("failed to get database secret: %w", err)
	}

	if err := secrets.ValidateDatabaseSecret(dbSecret); err != nil {
		return fmt.Errorf("database secret validation failed: %w", err)
	}

	storageSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      project.Spec.Storage.SecretRef.Name,
	}, storageSecret); err != nil {
		return fmt.Errorf("failed to get storage secret: %w", err)
	}

	if err := secrets.ValidateStorageSecret(storageSecret); err != nil {
		return fmt.Errorf("storage secret validation failed: %w", err)
	}

	return nil
}

func (r *SupabaseProjectReconciler) ensureDatabaseInitJob(ctx context.Context, project *supabasev1alpha1.SupabaseProject) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Create ConfigMap with SQL scripts
	configMap := resources.BuildDatabaseInitConfigMap(project)
	if err := controllerutil.SetControllerReference(project, configMap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Namespace: configMap.Namespace, Name: configMap.Name}, existingConfigMap)
	if err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, configMap); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create database init configmap: %w", err)
		}
		logger.Info("Created database init ConfigMap")
	}

	// Create Job
	job := resources.BuildDatabaseInitJob(project)
	if err := controllerutil.SetControllerReference(project, job, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	existingJob := &batchv1.Job{}
	err = r.Get(ctx, client.ObjectKey{Namespace: job.Namespace, Name: job.Name}, existingJob)
	if err != nil && apierrors.IsNotFound(err) {
		if err := r.Create(ctx, job); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create database init job: %w", err)
		}
		logger.Info("Created database init Job")
		// Job just created, requeue to check status
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Check job status
	if existingJob.Status.Succeeded > 0 {
		logger.Info("Database initialization job completed successfully")
		return ctrl.Result{}, nil
	}

	if existingJob.Status.Failed > 0 {
		// Check if we've exhausted retries
		if existingJob.Spec.BackoffLimit != nil && existingJob.Status.Failed > *existingJob.Spec.BackoffLimit {
			return ctrl.Result{}, fmt.Errorf("database initialization job failed after %d attempts", existingJob.Status.Failed)
		}
		// Still retrying
		logger.Info("Database initialization job failed, will retry", "attempts", existingJob.Status.Failed)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Job is still running
	logger.Info("Database initialization job is running", "active", existingJob.Status.Active)
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *SupabaseProjectReconciler) ensureJWTSecrets(ctx context.Context, project *supabasev1alpha1.SupabaseProject) error {
	secretName := project.Name + "-jwt"

	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: secretName}, existingSecret)

	if err == nil {
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return err
	}

	jwtSecret, err := secrets.GenerateJWTSecret()
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	anonKey, err := secrets.GenerateAnonKey(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to generate anon key: %w", err)
	}

	serviceRole, err := secrets.GenerateServiceRoleKey(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to generate service role key: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Namespace,
		},
		StringData: map[string]string{
			jwtSecretKey:   jwtSecret,
			anonKeyKey:     anonKey,
			serviceRoleKey: serviceRole,
		},
	}

	if err := controllerutil.SetControllerReference(project, secret, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, secret)
}

func (r *SupabaseProjectReconciler) reconcileComponent(
	ctx context.Context,
	project *supabasev1alpha1.SupabaseProject,
	name string,
	deploymentBuilder func(*supabasev1alpha1.SupabaseProject) *appsv1.Deployment,
	serviceBuilder func(*supabasev1alpha1.SupabaseProject) *corev1.Service,
) error {
	deployment := deploymentBuilder(project)
	if err := controllerutil.SetControllerReference(project, deployment, r.Scheme); err != nil {
		return err
	}

	existingDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Namespace: deployment.Namespace, Name: deployment.Name}, existingDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				return fmt.Errorf("failed to create %s deployment: %w", name, err)
			}
		} else {
			return err
		}
	} else {
		// Update existing deployment
		existingDeployment.Spec = deployment.Spec
		if err := r.Update(ctx, existingDeployment); err != nil {
			return fmt.Errorf("failed to update %s deployment: %w", name, err)
		}
	}

	service := serviceBuilder(project)
	if err := controllerutil.SetControllerReference(project, service, r.Scheme); err != nil {
		return err
	}

	existingService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Namespace: service.Namespace, Name: service.Name}, existingService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, service); err != nil {
				return fmt.Errorf("failed to create %s service: %w", name, err)
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *SupabaseProjectReconciler) handleDeletion(ctx context.Context, project *supabasev1alpha1.SupabaseProject) error {
	return nil
}

func (r *SupabaseProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasev1alpha1.SupabaseProject{}).
		Owns(&appsv1.Deployment{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Named("supabaseproject").
		Complete(r)
}

func apiequality(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
