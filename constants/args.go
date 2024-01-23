package constants

// Argument name constants
const (
	//ArgAll          = "all"
	ArgArg          = "arg"
	ArgAutoComplete = "auto-complete"
	//ArgBrowser                 = "browser"
	ArgCacheMaxTtl = "cache-max-ttl"
	ArgCacheTtl    = "cache-ttl"
	//ArgClear                   = "clear"
	ArgClientCacheEnabled = "client-cache-enabled"
	ArgCloudHost          = "cloud-host"
	ArgCloudToken         = "cloud-token"
	ArgConnectionString   = "connection-string"
	//ArgDashboard               = "dashboard"
	ArgDashboardListen         = "dashboard-listen"
	ArgDashboardPort           = "dashboard-port"
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
	//ArgForeground              = "foreground"
	//ArgFunctions               = "functions"
	ArgHeader        = "header"
	ArgHelp          = "help"
	ArgHost          = "host"
	ArgInput         = "input"
	ArgInsecure      = "insecure"
	ArgInstallDir    = "install-dir"
	ArgIntrospection = "introspection"
	//ArgInvoker                 = "invoker"
	ArgListen = "listen"
	//ArgLogDir                  = "log-dir"
	ArgLogLevel          = "log-level"
	ArgMaxCacheSizeMb    = "max-cache-size-mb"
	ArgMaxParallel       = "max-parallel"
	ArgMemoryMaxMb       = "memory-max-mb"
	ArgMemoryMaxMbPlugin = "memory-max-mb-plugin"
	ArgModInstall        = "mod-install"
	ArgModLocation       = "mod-location"
	ArgMultiLine         = "multi-line"
	ArgOff               = "off"
	ArgOn                = "on"
	ArgOutput            = "output"
	//ArgOutputDir               = "output-dir"
	//ArgOutputOnly              = "output-only"
	ArgPort = "port"
	//ArgPortHttps               = "port-https"
	ArgProgress = "progress"
	ArgPrune    = "prune"
	//ArgSchemaComments          = "schema-comments"
	ArgSearchPath          = "search-path"
	ArgSearchPathPrefix    = "search-path-prefix"
	ArgSeparator           = "separator"
	ArgServiceCacheEnabled = "service-cache-enabled"
	//ArgServiceMode         = "service-mode"
	//ArgServicePassword         = "database-password"
	//ArgServiceShowPassword     = "show-password"
	ArgShare = "share"
	//ArgSkipConfig              = "skip-config"
	ArgSnapshot         = "snapshot"
	ArgSnapshotLocation = "snapshot-location"
	ArgSnapshotTag      = "snapshot-tag"
	ArgSnapshotTitle    = "snapshot-title"
	ArgTag              = "tag"
	ArgTelemetry        = "telemetry"
	//ArgTheme            = "theme"
	ArgTiming      = "timing"
	ArgUpdateCheck = "update-check"
	ArgVarFile     = "var-file"
	ArgVariable    = "var"
	ArgVerbose     = "verbose"
	//ArgVersion           = "version"
	ArgWatch             = "watch"
	ArgWhere             = "where"
	ArgWorkspaceDatabase = "workspace-database"
	ArgWorkspaceProfile  = "workspace"
	ArgConfigPath        = "config-path"

	// Flowpipe concurrency
	ArgMaxConcurrencyHttp      = "max-concurrency-http"
	ArgMaxConcurrencyQuery     = "max-concurrency-query"
	ArgMaxConcurrencyContainer = "max-concurrency-container"
	ArgMaxConcurrencyFunction  = "max-concurrency-function"
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
