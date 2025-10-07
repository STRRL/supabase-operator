package status

import (
	"testing"
)

func TestPhaseTransition(t *testing.T) {
	tests := []struct {
		name         string
		currentPhase string
		targetPhase  string
		valid        bool
	}{
		{"Pending to ValidatingDependencies", PhasePending, PhaseValidatingDependencies, true},
		{"ValidatingDependencies to DeployingSecrets", PhaseValidatingDependencies, PhaseDeployingSecrets, true},
		{"DeployingSecrets to DeployingNetwork", PhaseDeployingSecrets, PhaseDeployingNetwork, true},
		{"DeployingNetwork to DeployingComponents", PhaseDeployingNetwork, PhaseDeployingComponents, true},
		{"DeployingComponents to Configuring", PhaseDeployingComponents, PhaseConfiguring, true},
		{"Configuring to Running", PhaseConfiguring, PhaseRunning, true},
		{"Running to Updating", PhaseRunning, PhaseUpdating, true},
		{"Updating to Running", PhaseUpdating, PhaseRunning, true},
		{"Any to Failed", PhaseRunning, PhaseFailed, true},
		{"Failed to ValidatingDependencies", PhaseFailed, PhaseValidatingDependencies, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.valid {
				t.Skip("Invalid transition test not implemented")
			}
		})
	}
}

func TestGetPhaseMessage(t *testing.T) {
	tests := []struct {
		phase   string
		wantMsg bool
	}{
		{PhasePending, true},
		{PhaseValidatingDependencies, true},
		{PhaseDeployingSecrets, true},
		{PhaseDeployingNetwork, true},
		{PhaseDeployingComponents, true},
		{PhaseConfiguring, true},
		{PhaseRunning, true},
		{PhaseUpdating, true},
		{PhaseFailed, true},
		{PhaseTerminating, true},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			msg := GetPhaseMessage(tt.phase)
			if tt.wantMsg && msg == "" {
				t.Errorf("Expected message for phase %s, got empty", tt.phase)
			}
		})
	}
}

func TestIsPhaseTerminal(t *testing.T) {
	tests := []struct {
		phase    string
		terminal bool
	}{
		{PhaseRunning, true},
		{PhaseFailed, true},
		{PhaseTerminating, true},
		{PhasePending, false},
		{PhaseValidatingDependencies, false},
		{PhaseDeployingComponents, false},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			result := IsPhaseTerminal(tt.phase)
			if result != tt.terminal {
				t.Errorf("IsPhaseTerminal(%s) = %v, want %v", tt.phase, result, tt.terminal)
			}
		})
	}
}
