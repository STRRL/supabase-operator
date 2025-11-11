package controller

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
