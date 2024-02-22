package constants

// Argument name constants
const (
	ArgArg                     = "arg"
	ArgAutoComplete            = "auto-complete"
	ArgCacheMaxTtl             = "cache-max-ttl"
	ArgCacheTtl                = "cache-ttl"
	ArgClientCacheEnabled      = "client-cache-enabled"
	ArgDashboardStartTimeout   = "dashboard-start-timeout"
	ArgDatabaseListenAddresses = "database-listen"
	ArgDatabasePort            = "database-port"
	ArgDatabaseQueryTimeout    = "query-timeout"
	ArgDatabaseStartTimeout    = "database-start-timeout"
	ArgDetach                  = "detach"
	ArgDisplayWidth            = "display-width"
	ArgDryRun                  = "dry-run"
	ArgExport                  = "export"
	ArgForce                   = "force"
	ArgHeader                  = "header"
	ArgHelp                    = "help"
	ArgHost                    = "host"
	ArgInput                   = "input"
	ArgInsecure                = "insecure"
	ArgInstallDir              = "install-dir"
	ArgIntrospection           = "introspection"
	ArgListen                  = "listen"
	ArgLogLevel                = "log-level"
	ArgMaxCacheSizeMb          = "max-cache-size-mb"
	ArgMaxParallel             = "max-parallel"
	ArgMemoryMaxMb             = "memory-max-mb"
	ArgMemoryMaxMbPlugin       = "memory-max-mb-plugin"
	ArgModInstall              = "mod-install"
	ArgModLocation             = "mod-location"
	ArgMultiLine               = "multi-line"
	ArgOff                     = "off"
	ArgOn                      = "on"
	ArgOutput                  = "output"
	ArgPipesInstallDir         = "pipes-install-dir"
	ArgPipesHost               = "pipes-host"
	ArgPipesToken              = "pipes-token"
	ArgPort                    = "port"
	ArgProgress                = "progress"
	ArgPrune                   = "prune"
	ArgSearchPath              = "search-path"
	ArgSearchPathPrefix        = "search-path-prefix"
	ArgSeparator               = "separator"
	ArgServiceCacheEnabled     = "service-cache-enabled"
	ArgShare                   = "share"
	ArgSnapshot                = "snapshot"
	ArgSnapshotLocation        = "snapshot-location"
	ArgSnapshotTag             = "snapshot-tag"
	ArgSnapshotTitle           = "snapshot-title"
	ArgTag                     = "tag"
	ArgTelemetry               = "telemetry"
	ArgTiming                  = "timing"
	ArgUpdateCheck             = "update-check"
	ArgVarFile                 = "var-file"
	ArgVariable                = "var"
	ArgVerbose                 = "verbose"
	ArgWatch                   = "watch"
	ArgWhere                   = "where"
	ArgDatabase                = "database"
	ArgWorkspaceProfile        = "workspace"
	ArgConfigPath              = "config-path"
	ArgBaseUrl                 = "base-url"

	// Flowpipe concurrency
	ArgMaxConcurrencyHttp      = "max-concurrency-http"
	ArgMaxConcurrencyQuery     = "max-concurrency-query"
	ArgMaxConcurrencyContainer = "max-concurrency-container"
	ArgMaxConcurrencyFunction  = "max-concurrency-function"
	ArgProcessRetention        = "process-retention"
)

// BoolToOnOff converts a boolean value onto the string "on" or "off"
func BoolToOnOff(val bool) string {
	if val {
		return ArgOn
	}
	return ArgOff
}

// BoolToEnableDisable converts a boolean value onto the string "enable" or "disable"
func BoolToEnableDisable(val bool) string {
	if val {
		return "enable"
	}
	return "disable"

}
