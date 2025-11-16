package controller

// Event constants follow Kubernetes API conventions for Events.
//
// Event Reason Format:
//   - Must use CamelCase format (e.g., "PhaseChanged", "DatabaseInitialized")
//   - Should be short, unique, and suitable for automation/switch statements
//   - Maximum 128 characters
//   - Represents a programmatic category identifier for kubectl get output
//
// Event Message Format:
//   - Human-readable phrase or sentence with first letter capitalized
//   - May contain specific details and multiple words
//   - Used in kubectl describe output for detailed status explanations
//   - Product names and acronyms should maintain their original casing
//
// References:
//   - Kubernetes API Conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#events
//   - Kubebuilder Event Guide: https://book.kubebuilder.io/reference/raising-events
//   - Kubernetes Event API: https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/event-v1/
//   - Core Event Examples: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/events/event.go

const (
	EventReasonPhaseChanged             = "PhaseChanged"
	EventReasonDependenciesValidated    = "DependenciesValidated"
	EventReasonValidationFailed         = "ValidationFailed"
	EventReasonSecretsCreated           = "SecretsCreated"
	EventReasonSecretsFailed            = "SecretsFailed"
	EventReasonDatabaseInitialized      = "DatabaseInitialized"
	EventReasonDatabaseInitFailed       = "DatabaseInitFailed"
	EventReasonComponentDeploymentReady = "ComponentDeploymentReady"
	EventReasonReconciliationComplete   = "ReconciliationComplete"
)

const (
	EventMessagePhasePending                  = "Entered Pending phase"
	EventMessageDependenciesValidated         = "Successfully validated external dependencies"
	EventMessageDeployingSecrets              = "Deploying JWT secrets"
	EventMessageSecretsCreated                = "JWT secrets generated successfully"
	EventMessageInitializingDatabase          = "Initializing PostgreSQL database"
	EventMessageDatabaseInitialized           = "PostgreSQL database initialized successfully"
	EventMessageDeployingComponents           = "Deploying Supabase components"
	EventMessageSupabaseProjectRunning        = "SupabaseProject is now Running"
	EventMessageDependencyValidationFailedFmt = "Dependency validation failed: %v"
	EventMessageSecretsFailedFmt              = "Failed to generate JWT secrets: %v"
	EventMessageDatabaseInitFailedFmt         = "Failed to initialize database: %v"
)
