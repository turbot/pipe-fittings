package filepaths

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/utils"
)

const (
	localPluginFolder = "local"
)

// EnsurePluginDir returns the path to the plugins directory (creates if missing)
func EnsurePluginDir() string {
	return ensureInstallSubDir("plugins")
}
func EnsurePluginInstallDir(pluginImageDisplayRef string) string {
	installDir := PluginInstallDir(pluginImageDisplayRef)

	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create plugin install directory")
	}

	return installDir
}

func PluginInstallDir(pluginImageDisplayRef string) string {
	osSafePath := filepath.FromSlash(pluginImageDisplayRef)

	fullPath := filepath.Join(EnsurePluginDir(), osSafePath)
	return fullPath
}

func PluginBinaryPath(pluginImageDisplayRef, pluginAlias string) string {
	return filepath.Join(PluginInstallDir(pluginImageDisplayRef), PluginAliasToLongName(pluginAlias)+".plugin")
}

func GetPluginPath(pluginImageRef, pluginAlias string) (string, error) {
	// the fully qualified name of the plugin is the relative path of the folder containing the plugin
	// calculate absolute folder path
	pluginFolder := filepath.Join(EnsurePluginDir(), pluginImageRef)

	// if the plugin folder is missing, it is possible the plugin path was truncated to create a schema name
	// - so search for a folder which when truncated would match the schema
	if _, err := os.Stat(pluginFolder); os.IsNotExist(err) {
		slog.Debug("plugin path not found - searching for folder using hashed name", "plugin path", pluginFolder)
		if pluginFolder, err = FindPluginFolder(pluginImageRef); err != nil {
			return "", err
		} else if pluginFolder == "" {
			return "", fmt.Errorf("no plugin installed matching '%s'", pluginAlias)
		}
	}

	// there should be just 1 file with extension pluginExtension (".plugin")
	entries, err := os.ReadDir(pluginFolder)
	if err != nil {
		return "", fmt.Errorf("failed to load plugin '%s': %v", pluginImageRef, err)
	}
	var matches []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == constants.PluginExtension {
			matches = append(matches, entry.Name())
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("plugin folder %s should contain a single plugin file. %d plugins were found ", pluginFolder, len(matches))
	}

	return filepath.Join(pluginFolder, matches[0]), nil
}

// FindPluginFolder searches for a folder which when hashed would match the schema
func FindPluginFolder(remoteSchema string) (string, error) {
	pluginDir := EnsurePluginDir()

	// first try searching by prefix - trim the schema name
	globPattern := filepath.Join(pluginDir, utils.TrimSchemaName(remoteSchema)) + "*"
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return "", err
	} else if len(matches) == 1 {
		return matches[0], nil
	}

	for _, match := range matches {
		// get the relative path to this match from the plugin folder
		folderRelativePath, err := filepath.Rel(pluginDir, match)
		if err != nil {
			// do not fail on error here
			continue
		}
		hashedName := utils.PluginFQNToSchemaName(folderRelativePath)
		if hashedName == remoteSchema {
			return filepath.Join(pluginDir, folderRelativePath), nil
		}
	}

	return "", nil
}

// LocalPluginPath returns the path to locally installed plugins
func LocalPluginPath() string {
	return filepath.Join(EnsurePluginDir(), localPluginFolder)
}
