package workspace

import (
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/zclconf/go-cty/cty"
)

type LoadWorkspaceOption func(*LoadWorkspaceConfig)

type LoadWorkspaceConfig struct {
	skipResourceLoadIfNoModfile bool
	pipelingConnections         map[string]connection.PipelingConnection
	blockTypeInclusions         []string
	validateVariables           bool
	supportLateBinding          bool
	configValueMaps             map[string]map[string]cty.Value
	decoderOptions              []parse.DecoderOption
}

func newLoadFlowpipeWorkspaceConfig() *LoadWorkspaceConfig {
	return &LoadWorkspaceConfig{
		pipelingConnections: make(map[string]connection.PipelingConnection),
		validateVariables:   true,
		supportLateBinding:  true,
		configValueMaps:     make(map[string]map[string]cty.Value),
	}
}

func WithPipelingConnections(pipelingConnections map[string]connection.PipelingConnection) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.pipelingConnections = pipelingConnections
	}
}

func WithLateBinding(enabled bool) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.supportLateBinding = enabled
	}
}

func WithConfigValueMap(name string, valueMap map[string]cty.Value) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.configValueMaps[name] = valueMap
	}
}

func WithBlockType(blockTypeInclusions []string) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.blockTypeInclusions = blockTypeInclusions
	}
}

func WithVariableValidation(enabled bool) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.validateVariables = enabled
	}
}
func WithDecoderOptions(opts ...parse.DecoderOption) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.decoderOptions = opts
	}
}

// TODO this is only needed as Pipe fittings tests rely on loading workspaces without modfiles
func WithSkipResourceLoadIfNoModfile(enabled bool) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.skipResourceLoadIfNoModfile = enabled
	}
}
