package constants

import "github.com/turbot/pipe-fittings/app_specific"

// Environment Variables
const (
	EnvPipesHost  = "PIPES_HOST"
	EnvPipesToken = "PIPES_TOKEN"
)

// app specific env vars
var (
	EnvUpdateCheck           = app_specific.BuildEnv("UPDATE_CHECK")
	EnvInstallDir            = app_specific.BuildEnv("INSTALL_DIR")
	EnvInstallDatabase       = app_specific.BuildEnv("INITDB_DATABASE_NAME")
	EnvServicePassword       = app_specific.BuildEnv("DATABASE_PASSWORD")
	EnvMaxParallel           = app_specific.BuildEnv("MAX_PARALLEL")
	EnvDatabaseStartTimeout  = app_specific.BuildEnv("DATABASE_START_TIMEOUT")
	EnvDashboardStartTimeout = app_specific.BuildEnv("DASHBOARD_START_TIMEOUT")

	EnvSnapshotLocation  = app_specific.BuildEnv("SNAPSHOT_LOCATION")
	EnvWorkspaceDatabase = app_specific.BuildEnv("WORKSPACE_DATABASE")
	EnvWorkspaceProfile  = app_specific.BuildEnv("WORKSPACE")
	EnvCloudHost         = app_specific.BuildEnv("CLOUD_HOST")
	EnvCloudToken        = app_specific.BuildEnv("CLOUD_TOKEN")

	EnvDisplayWidth = app_specific.BuildEnv("DISPLAY_WIDTH")
	EnvCacheEnabled = app_specific.BuildEnv("CACHE")
	EnvCacheTTL     = app_specific.BuildEnv("CACHE_TTL")
	EnvCacheMaxTTL  = app_specific.BuildEnv("CACHE_MAX_TTL")
	EnvCacheMaxSize = app_specific.BuildEnv("CACHE_MAX_SIZE_MB")
	EnvQueryTimeout = app_specific.BuildEnv("QUERY_TIMEOUT")

	EnvConnectionWatcher        = app_specific.BuildEnv("CONNECTION_WATCHER")
	EnvWorkspaceChDir           = app_specific.BuildEnv("WORKSPACE_CHDIR")
	EnvModLocation              = app_specific.BuildEnv("MOD_LOCATION")
	EnvTelemetry                = app_specific.BuildEnv("TELEMETRY")
	EnvIntrospection            = app_specific.BuildEnv("INTROSPECTION")
	EnvWorkspaceProfileLocation = app_specific.BuildEnv("WORKSPACE_PROFILES_LOCATION")

	// EnvConfigDump is an undocumented variable is subject to change in the future
	EnvConfigDump = app_specific.BuildEnv("CONFIG_DUMP")

	EnvMemoryMaxMb       = app_specific.BuildEnv("MEMORY_MAX_MB")
	EnvMemoryMaxMbPlugin = app_specific.BuildEnv("PLUGIN_MEMORY_MAX_MB")
)
