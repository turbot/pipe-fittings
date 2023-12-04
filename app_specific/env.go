package app_specific

// app specific env vars
var EnvUpdateCheck,
	EnvInstallDir,
	EnvInstallDatabase,
	EnvServicePassword,
	EnvMaxParallel,
	EnvDatabaseStartTimeout,
	EnvDashboardStartTimeout,
	EnvSnapshotLocation,
	EnvWorkspaceDatabase,
	EnvWorkspaceProfile,
	EnvCloudHost,
	EnvCloudToken,
	EnvDisplayWidth,
	EnvCacheEnabled,
	EnvCacheTTL,
	EnvCacheMaxTTL,
	EnvCacheMaxSize,
	EnvQueryTimeout,
	EnvConnectionWatcher,
	EnvWorkspaceChDir,
	EnvModLocation,
	EnvTelemetry,
	EnvIntrospection,
	EnvWorkspaceProfileLocation,
	// EnvConfigDump is an undocumented variable is subject to change in the future
	EnvConfigDump,
	EnvMemoryMaxMb,
	EnvMemoryMaxMbPlugin,
	EnvLogLevel string

func SetAppSpecificEnvVarKeys(envAppPrefix string) {
	// set prefix
	EnvAppPrefix = envAppPrefix

	EnvUpdateCheck = buildEnv("UPDATE_CHECK")
	EnvInstallDir = buildEnv("INSTALL_DIR")
	EnvInstallDatabase = buildEnv("INITDB_DATABASE_NAME")
	EnvServicePassword = buildEnv("DATABASE_PASSWORD")
	EnvMaxParallel = buildEnv("MAX_PARALLEL")
	EnvDatabaseStartTimeout = buildEnv("DATABASE_START_TIMEOUT")
	EnvDashboardStartTimeout = buildEnv("DASHBOARD_START_TIMEOUT")
	EnvSnapshotLocation = buildEnv("SNAPSHOT_LOCATION")
	EnvWorkspaceDatabase = buildEnv("WORKSPACE_DATABASE")
	EnvWorkspaceProfile = buildEnv("WORKSPACE")
	EnvCloudHost = buildEnv("CLOUD_HOST")
	EnvCloudToken = buildEnv("CLOUD_TOKEN")
	EnvDisplayWidth = buildEnv("DISPLAY_WIDTH")
	EnvCacheEnabled = buildEnv("CACHE")
	EnvCacheTTL = buildEnv("CACHE_TTL")
	EnvCacheMaxTTL = buildEnv("CACHE_MAX_TTL")
	EnvCacheMaxSize = buildEnv("CACHE_MAX_SIZE_MB")
	EnvQueryTimeout = buildEnv("QUERY_TIMEOUT")
	EnvConnectionWatcher = buildEnv("CONNECTION_WATCHER")
	EnvWorkspaceChDir = buildEnv("WORKSPACE_CHDIR")
	EnvModLocation = buildEnv("MOD_LOCATION")
	EnvTelemetry = buildEnv("TELEMETRY")
	EnvIntrospection = buildEnv("INTROSPECTION")
	EnvWorkspaceProfileLocation = buildEnv("WORKSPACE_PROFILES_LOCATION")
	EnvConfigDump = buildEnv("CONFIG_DUMP")
	EnvMemoryMaxMb = buildEnv("MEMORY_MAX_MB")
	EnvMemoryMaxMbPlugin = buildEnv("PLUGIN_MEMORY_MAX_MB")
	EnvLogLevel = buildEnv("LOG_LEVEL")
}

// buildEnv is a function to construct an application specific env var key
func buildEnv(suffix string) string {
	return EnvAppPrefix + suffix
}
