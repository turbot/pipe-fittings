package filepaths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/pipe-fittings/app_specific"
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

// EnsureControlTemplateDir returns the path to the templates directory (creates if missing)
func EnsureControlTemplateDir() string {
	return ensureInstallSubDir(filepath.Join("check", "templates"))
}

// EnsureDetectionTemplateDir returns the path to the detection templates directory (creates if missing)
func EnsureDetectionTemplateDir() string {
	return ensureInstallSubDir(filepath.Join("check", "detection_templates"))
}

// EnsureConfigDir returns the path to the config directory (creates if missing)
func EnsureConfigDir() string {
	return ensureInstallSubDir("config")
}

// GetInternalDir returns the path to the internal directory
func GetInternalDir() string {
	return installSubDir("internal")
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

// GetDataDir returns the path to the data directory
func GetDataDir() string {
	return installSubDir("data")
}

// EnsureDataDir returns the path to the data directory (creates if missing)
func EnsureDataDir() string {
	return ensureInstallSubDir("data")
}
