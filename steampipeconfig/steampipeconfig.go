package steampipeconfig

import "github.com/turbot/pipe-fittings/modconfig"

// SteampipeConfig is a struct to hold Connection map and Steampipe options
type SteampipeConfig struct {
	// map of plugin configs, keyed by plugin image ref
	// (for each image ref we store an array of configs)
	Plugins map[string][]*modconfig.Plugin
	// map of plugin configs, keyed by plugin instance
	PluginsInstances map[string]*modconfig.Plugin
	// map of connection name to partially parsed connection config
	Connections map[string]*modconfig.Connection
}
