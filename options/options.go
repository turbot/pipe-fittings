package options

// hcl options block types
const (
	ServerBlock     = "server"
	ConnectionBlock = "connection"
	QueryBlock      = "query"
	CheckBlock      = "check"
	DashboardBlock  = "dashboard"
	DatabaseBlock   = "database"
	GeneralBlock    = "general"
	PluginBlock     = "plugin"
)

type Options interface {
	// map of config keys to values - used to populate viper
	ConfigMap() map[string]interface{}
	// merge with another options of same type
	Merge(otherOptions Options)
}
