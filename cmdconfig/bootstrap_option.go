package cmdconfig

type BootstrapConfig struct {
	ConfigDefaults       map[string]any
	DirectoryEnvMappings map[string]EnvMapping
}

func NewBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		ConfigDefaults:       make(map[string]any),
		DirectoryEnvMappings: make(map[string]EnvMapping),
	}
}

type BootstrapOption func(*BootstrapConfig)

func WithConfigDefaults(configDefaults map[string]any) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.ConfigDefaults = configDefaults
	}
}
func WithDirectoryEnvMappings(directoryEnvMappings map[string]EnvMapping) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.DirectoryEnvMappings = directoryEnvMappings
	}
}
