package flowpipe

import (
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type LoadFlowpipeWorkspaceOption func(*LoadFlowpipeWorkspaceConfig)

type LoadFlowpipeWorkspaceConfig struct {
	credentials                 map[string]credential.Credential
	integrations                map[string]flowpipe.Integration
	notifiers                   map[string]flowpipe.Notifier
	skipResourceLoadIfNoModfile bool
	pipelingConnections         map[string]connection.PipelingConnection
	blockTypeInclusions         []string
	validateVariables           bool
	supportLateBinding          bool
}

func newLoadFlowpipeWorkspaceConfig() *LoadFlowpipeWorkspaceConfig {
	return &LoadFlowpipeWorkspaceConfig{
		credentials:         make(map[string]credential.Credential),
		integrations:        make(map[string]flowpipe.Integration),
		notifiers:           make(map[string]flowpipe.Notifier),
		pipelingConnections: make(map[string]connection.PipelingConnection),
		validateVariables:   true,
		supportLateBinding:  true,
	}
}

func WithPipelingConnections(pipelingConnections map[string]connection.PipelingConnection) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.pipelingConnections = pipelingConnections
	}
}

func WithLateBinding(enabled bool) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.supportLateBinding = enabled
	}
}

func WithCredentials(credentials map[string]credential.Credential) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.credentials = credentials
	}
}

func WithIntegrations(integrations map[string]flowpipe.Integration) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.integrations = integrations
	}
}

func WithNotifiers(notifiers map[string]flowpipe.Notifier) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.notifiers = notifiers
	}
}

func WithBlockType(blockTypeInclusions []string) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.blockTypeInclusions = blockTypeInclusions
	}
}

func WithVariableValidation(enabled bool) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.validateVariables = enabled
	}
}

// TODO this is only needed as Pipe fittings tests rely on loading workspaces without modfiles
func WithSkipResourceLoadIfNoModfile(enabled bool) LoadFlowpipeWorkspaceOption {
	return func(m *LoadFlowpipeWorkspaceConfig) {
		m.skipResourceLoadIfNoModfile = enabled
	}
}
