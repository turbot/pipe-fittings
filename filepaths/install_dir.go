package filepaths

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/error_helpers"
)

// Constants for Config

const (
	connectionsStateFileName     = "connection.json"
	versionFileName              = "versions.json"
	databaseRunningInfoFileName  = "steampipe.json"
	pluginManagerStateFileName   = "plugin_manager.json"
	dashboardServerStateFileName = "dashboard_service.json"
	stateFileName                = "update_check.json"
	legacyStateFileName          = "update-check.json"
	availableVersionsFileName    = "available_versions.json"
	legacyNotificationsFileName  = "notifications.json"

	PipesComponentInternal = "internal"
)

func ensureInstallSubDir(dirName string) string {
	subDir := installSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

func installSubDir(dirName string) string {
	if app_specific.InstallDir == "" {
		panic(fmt.Errorf("cannot call any %s directory functions before InstallDir is set", app_specific.AppName))
	}
	return filepath.Join(app_specific.InstallDir, dirName)
}

// EnsureTemplateDir returns the path to the templates directory (creates if missing)
func EnsureTemplateDir() string {
	return ensureInstallSubDir(filepath.Join("check", "templates"))
}

// EnsurePluginDir returns the path to the plugins directory (creates if missing)
func EnsurePluginDir() string {
	return ensureInstallSubDir("plugins")
}

// EnsureConfigDir returns the path to the config directory (creates if missing)
func EnsureConfigDir() string {
	return ensureInstallSubDir("config")
}

// EnsureInternalDir returns the path to the internal directory (creates if missing)
func EnsureInternalDir() string {
	return ensureInstallSubDir("internal")
}

// EnsureBackupsDir returns the path to the backups directory (creates if missing)
func EnsureBackupsDir() string {
	return ensureInstallSubDir("backups")
}

// BackupsDir returns the path to the backups directory
func BackupsDir() string {
	return installSubDir("backups")
}

// GlobalWorkspaceProfileDir returns the path to the workspace profiles directory
// if  STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set use that
// otherwise look in the config folder
// NOTE: unlike other path functions this accepts the install-dir as arg
// this is because of the slightly complex bootstrapping process required because the
// install-dir may be set in the workspace profile
func GlobalWorkspaceProfileDir(installDir string) (string, error) {
	if workspaceProfileLocation, ok := os.LookupEnv(app_specific.EnvWorkspaceProfileLocation); ok {
		return filehelpers.Tildefy(workspaceProfileLocation)
	}
	return filepath.Join(installDir, "config"), nil

}

// LocalWorkspaceProfileDir returns the path to the local workspace profiles directory.
// i.e. the workspace profiles which may be specified in the mod-location
func LocalWorkspaceProfileDir(modLocation string) (string, error) {
	return filepath.Join(modLocation, app_specific.WorkspaceDataDir, "config"), nil
}

// EnsureDatabaseDir returns the path to the db directory (creates if missing)
func EnsureDatabaseDir() string {
	return ensureInstallSubDir("db")
}

// EnsureLogDir returns the path to the db log directory (creates if missing)
func EnsureLogDir() string {
	return ensureInstallSubDir("logs")
}

func EnsureDashboardAssetsDir() string {
	return ensureInstallSubDir(filepath.Join("dashboard", "assets"))
}

// LegacyDashboardAssetsDir returns the path to the legacy report assets folder
func LegacyDashboardAssetsDir() string {
	return installSubDir("report")
}

// LegacyStateFilePath returns the path of the legacy update-check.json state file
func LegacyStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), legacyStateFileName)
}

// StateFilePath returns the path of the update_check.json state file
func StateFilePath() string {
	return filepath.Join(EnsureInternalDir(), stateFileName)
}

// AvailableVersionsFilePath returns the path of the json file used to store cache available versions of installed plugins and the CLI
func AvailableVersionsFilePath() string {
	return filepath.Join(EnsureInternalDir(), availableVersionsFileName)
}

// LegacyNotificationsFilePath returns the path of the (legacy) notifications.json file used to store update notifications
func LegacyNotificationsFilePath() string {
	return filepath.Join(EnsureInternalDir(), legacyNotificationsFileName)
}

// ConnectionStatePath returns the path of the connections state file
func ConnectionStatePath() string {
	return filepath.Join(EnsureInternalDir(), connectionsStateFileName)
}

// LegacyVersionFilePath returns the legacy version file path
func LegacyVersionFilePath() string {
	return filepath.Join(EnsureInternalDir(), versionFileName)
}

// PluginVersionFilePath returns the plugin version file path
func PluginVersionFilePath() string {
	return filepath.Join(EnsurePluginDir(), versionFileName)
}

// DatabaseVersionFilePath returns the plugin version file path
func DatabaseVersionFilePath() string {
	return filepath.Join(EnsureDatabaseDir(), versionFileName)
}

// ReportAssetsVersionFilePath returns the report assets version file path
func ReportAssetsVersionFilePath() string {
	return filepath.Join(EnsureDashboardAssetsDir(), versionFileName)
}

func RunningInfoFilePath() string {
	return filepath.Join(EnsureInternalDir(), databaseRunningInfoFileName)
}

func PluginManagerStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), pluginManagerStateFileName)
}

func DashboardServiceStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), dashboardServerStateFileName)
}

func StateFileName() string {
	return stateFileName
}
