package status

const (
	PhasePending                = "Pending"
	PhaseValidatingDependencies = "ValidatingDependencies"
	PhaseDeployingSecrets       = "DeployingSecrets"
	PhaseDeployingNetwork       = "DeployingNetwork"
	PhaseDeployingComponents    = "DeployingComponents"
	PhaseConfiguring            = "Configuring"
	PhaseRunning                = "Running"
	PhaseUpdating               = "Updating"
	PhaseFailed                 = "Failed"
	PhaseTerminating            = "Terminating"
)

func GetPhaseMessage(phase string) string {
	messages := map[string]string{
		PhasePending:                "SupabaseProject is pending",
		PhaseValidatingDependencies: "Validating external dependencies",
		PhaseDeployingSecrets:       "Deploying secrets and credentials",
		PhaseDeployingNetwork:       "Deploying network resources",
		PhaseDeployingComponents:    "Deploying Supabase components",
		PhaseConfiguring:            "Configuring components",
		PhaseRunning:                "All components running",
		PhaseUpdating:               "Updating components",
		PhaseFailed:                 "Reconciliation failed",
		PhaseTerminating:            "Terminating resources",
	}

	if msg, ok := messages[phase]; ok {
		return msg
	}
	return "Unknown phase"
}

func IsPhaseTerminal(phase string) bool {
	return phase == PhaseRunning || phase == PhaseFailed || phase == PhaseTerminating
}

func IsPhaseHealthy(phase string) bool {
	return phase == PhaseRunning
}

func CanTransitionTo(currentPhase, targetPhase string) bool {
	if targetPhase == PhaseFailed {
		return true
	}

	if currentPhase == PhaseFailed && targetPhase == PhaseValidatingDependencies {
		return true
	}

	transitions := map[string][]string{
		PhasePending:                {PhaseValidatingDependencies},
		PhaseValidatingDependencies: {PhaseDeployingSecrets},
		PhaseDeployingSecrets:       {PhaseDeployingNetwork},
		PhaseDeployingNetwork:       {PhaseDeployingComponents},
		PhaseDeployingComponents:    {PhaseConfiguring},
		PhaseConfiguring:            {PhaseRunning},
		PhaseRunning:                {PhaseUpdating, PhaseTerminating},
		PhaseUpdating:               {PhaseRunning, PhaseConfiguring},
	}

	allowed, ok := transitions[currentPhase]
	if !ok {
		return false
	}

	for _, allowed := range allowed {
		if allowed == targetPhase {
			return true
		}
	}

	return false
}
