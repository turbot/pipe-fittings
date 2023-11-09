package cmdconfig

type bootstrapConfig struct {
	configDefaults       map[string]any
	directoryEnvMappings map[string]EnvMapping
}

func newBootstrapConfig() *bootstrapConfig {
	return &bootstrapConfig{
		configDefaults:       make(map[string]any),
		directoryEnvMappings: make(map[string]EnvMapping),
	}
}

type bootstrapOption func(*bootstrapConfig)

func WithConfigDefaults(configDefaults map[string]any) bootstrapOption {
	return func(c *bootstrapConfig) {
		c.configDefaults = configDefaults
	}
}
func WithDirectoryEnvMappings(directoryEnvMappings map[string]EnvMapping) bootstrapOption {
	return func(c *bootstrapConfig) {
		c.directoryEnvMappings = directoryEnvMappings
	}
}
