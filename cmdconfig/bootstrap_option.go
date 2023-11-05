package cmdconfig

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/steampipeconfig"
)

type bootstrapConfig struct {
	configDefaults         map[string]any
	directoryEnvMappings   map[string]EnvMapping
	workspaceProfileLoader *steampipeconfig.WorkspaceProfileLoader
}

func newBootstrapConfig() *bootstrapConfig {
	return &bootstrapConfig{
		configDefaults:       make(map[string]any),
		directoryEnvMappings: make(map[string]EnvMapping),
		/// create empty loader
		workspaceProfileLoader: &steampipeconfig.WorkspaceProfileLoader{DefaultProfile: &modconfig.WorkspaceProfile{}},
	}
}

type bootstrapOption func(*bootstrapConfig)
