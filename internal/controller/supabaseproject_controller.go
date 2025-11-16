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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
	"github.com/strrl/supabase-operator/internal/component"
	"github.com/strrl/supabase-operator/internal/controller/reconciler"
	"github.com/strrl/supabase-operator/internal/secrets"
	"github.com/strrl/supabase-operator/internal/status"
)

const (
	finalizerName   = "supabase.strrl.dev/finalizer"
	jwtSecretKey    = "jwt-secret"
	anonKeyKey      = "anon-key"
	serviceRoleKey  = "service-role-key"
	pgMetaCryptoKey = "pg-meta-crypto-key"
)

type SupabaseProjectReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
		r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonPhaseChanged, EventMessagePhasePending)
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
		r.Recorder.Eventf(project, corev1.EventTypeWarning, EventReasonValidationFailed, EventMessageDependencyValidationFailedFmt, err)
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonDependenciesValidated, EventMessageDependenciesValidated)

	// Generate JWT secrets first (needed by database init job)
	project.Status.Phase = status.PhaseDeployingSecrets
	project.Status.Message = status.GetPhaseMessage(status.PhaseDeployingSecrets)
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonPhaseChanged, EventMessageDeployingSecrets)

	if err := r.ensureJWTSecrets(ctx, project); err != nil {
		logger.Error(err, "Failed to ensure JWT secrets")
		project.Status.Phase = status.PhaseFailed
		project.Status.Message = fmt.Sprintf("Secret generation failed: %v", err)
		r.Recorder.Eventf(project, corev1.EventTypeWarning, EventReasonSecretsFailed, EventMessageSecretsFailedFmt, err)
		if updateErr := r.Status().Update(ctx, project); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonSecretsCreated, EventMessageSecretsCreated)

	// Initialize database with required extensions and roles via Kubernetes Job
	project.Status.Phase = status.PhaseInitializingDatabase
	project.Status.Message = status.GetPhaseMessage(status.PhaseInitializingDatabase)
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonPhaseChanged, EventMessageInitializingDatabase)

	jobResult, err := r.ensureDatabaseInitJob(ctx, project)
	if err != nil {
		logger.Error(err, "Failed to ensure database init job")
		r.Recorder.Eventf(project, corev1.EventTypeWarning, EventReasonDatabaseInitFailed, EventMessageDatabaseInitFailedFmt, err)
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
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonDatabaseInitialized, EventMessageDatabaseInitialized)

	project.Status.Phase = status.PhaseDeployingComponents
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonPhaseChanged, EventMessageDeployingComponents)

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
	r.Recorder.Event(project, corev1.EventTypeNormal, EventReasonPhaseChanged, EventMessageSupabaseProjectRunning)
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

	componentReconciler := &reconciler.ComponentReconciler{
		Client: r.Client,
		Scheme: r.Scheme,
	}

	kongConfigMap := component.BuildKongConfigMap(project)
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

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.KongBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Kong")
		return componentsStatus, err
	}
	kongDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-kong"}, kongDeploy); err != nil {
		logger.Error(err, "Failed to get Kong deployment status")
	} else {
		replicas := int32(0)
		if kongDeploy.Spec.Replicas != nil {
			replicas = *kongDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "Kong",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.Kong.Image, replicas, kongDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.AuthBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Auth")
		return componentsStatus, err
	}
	authDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-auth"}, authDeploy); err != nil {
		logger.Error(err, "Failed to get Auth deployment status")
	} else {
		replicas := int32(0)
		if authDeploy.Spec.Replicas != nil {
			replicas = *authDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "Auth",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.Auth.Image, replicas, authDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.PostgRESTBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile PostgREST")
		return componentsStatus, err
	}
	postgrestDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-postgrest"}, postgrestDeploy); err != nil {
		logger.Error(err, "Failed to get PostgREST deployment status")
	} else {
		replicas := int32(0)
		if postgrestDeploy.Spec.Replicas != nil {
			replicas = *postgrestDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "PostgREST",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.PostgREST.Image, replicas, postgrestDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.RealtimeBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Realtime")
		return componentsStatus, err
	}
	realtimeDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-realtime"}, realtimeDeploy); err != nil {
		logger.Error(err, "Failed to get Realtime deployment status")
	} else {
		replicas := int32(0)
		if realtimeDeploy.Spec.Replicas != nil {
			replicas = *realtimeDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "Realtime",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.Realtime.Image, replicas, realtimeDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.StorageBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Storage")
		return componentsStatus, err
	}
	storageDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-storage"}, storageDeploy); err != nil {
		logger.Error(err, "Failed to get Storage deployment status")
	} else {
		replicas := int32(0)
		if storageDeploy.Spec.Replicas != nil {
			replicas = *storageDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "StorageAPI",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.StorageAPI.Image, replicas, storageDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.MetaBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Meta")
		return componentsStatus, err
	}
	metaDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-meta"}, metaDeploy); err != nil {
		logger.Error(err, "Failed to get Meta deployment status")
	} else {
		replicas := int32(0)
		if metaDeploy.Spec.Replicas != nil {
			replicas = *metaDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "Meta",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.Meta.Image, replicas, metaDeploy.Status.ReadyReplicas))
	}

	if err := componentReconciler.ReconcileComponent(ctx, project, &component.StudioBuilder{}); err != nil {
		logger.Error(err, "Failed to reconcile Studio")
		return componentsStatus, err
	}
	studioDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: project.Namespace, Name: project.Name + "-studio"}, studioDeploy); err != nil {
		logger.Error(err, "Failed to get Studio deployment status")
	} else {
		replicas := int32(0)
		if studioDeploy.Spec.Replicas != nil {
			replicas = *studioDeploy.Spec.Replicas
		}
		componentsStatus = status.SetComponentStatus(componentsStatus, "Studio",
			status.NewComponentStatus(status.PhaseRunning, project.Spec.Studio.Image, replicas, studioDeploy.Status.ReadyReplicas))
	}

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

	if project.Spec.Studio != nil && project.Spec.Studio.DashboardBasicAuthSecretRef != nil {
		secretRef := project.Spec.Studio.DashboardBasicAuthSecretRef
		namespace := secretRef.Namespace
		if namespace == "" {
			namespace = project.Namespace
		}

		dashboardSecret := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretRef.Name}, dashboardSecret); err != nil {
			return fmt.Errorf("failed to get dashboard basic auth secret: %w", err)
		}

		if err := secrets.ValidateBasicAuthSecret(dashboardSecret); err != nil {
			return fmt.Errorf("dashboard basic auth secret validation failed: %w", err)
		}
	}

	return nil
}

func (r *SupabaseProjectReconciler) ensureDatabaseInitJob(ctx context.Context, project *supabasev1alpha1.SupabaseProject) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Create ConfigMap with SQL scripts
	configMap := component.BuildDatabaseInitConfigMap(project)
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
	job := component.BuildDatabaseInitJob(project)
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
		if _, ok := existingSecret.Data[pgMetaCryptoKey]; ok {
			return nil
		}

		metaCryptoKey, genErr := secrets.GeneratePGMetaCryptoKey()
		if genErr != nil {
			return fmt.Errorf("failed to generate PG meta crypto key: %w", genErr)
		}

		if existingSecret.Data == nil {
			existingSecret.Data = map[string][]byte{}
		}
		existingSecret.Data[pgMetaCryptoKey] = []byte(metaCryptoKey)

		return r.Update(ctx, existingSecret)
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

	metaCryptoKey, err := secrets.GeneratePGMetaCryptoKey()
	if err != nil {
		return fmt.Errorf("failed to generate PG meta crypto key: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: project.Namespace,
		},
		StringData: map[string]string{
			jwtSecretKey:    jwtSecret,
			anonKeyKey:      anonKey,
			serviceRoleKey:  serviceRole,
			pgMetaCryptoKey: metaCryptoKey,
		},
	}

	if err := controllerutil.SetControllerReference(project, secret, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, secret)
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
