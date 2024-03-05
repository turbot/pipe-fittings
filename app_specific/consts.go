package app_specific

import (
	"github.com/Masterminds/semver/v3"
	"path/filepath"
)

// Application specific constants which MUST be set by the application

// app name and version

var AppName string
var AppVersion *semver.Version

// filepaths

var DefaultVarsFileName string
var LegacyDefaultVarsFileName string

// TODO KAI  we need to provide a default (for now) as the flowpipe test code does not (always)
// call SetAppSpecificConstants so this may be empty we need a proper solution to this
var ModFileNameDeprecated string = "mod.sp"

func ModFileNames() []string {
	var res []string
	for _, ext := range ModDataExtensions {
		res = append(res, "mod"+ext)

	}
	return res
}

func ModFilePaths(modFolder string) []string {
	var res []string
	for _, filename := range ModFileNames() {
		res = append(res, filepath.Join(modFolder, filename))
	}
	return res
}

func DefaultModFileName() string {
	return ModFileNames()[0]
}
func DefaultModFilePath(modFolder string) string {
	return filepath.Join(modFolder, DefaultModFileName())
}

var WorkspaceIgnoreFile string
var WorkspaceDataDir string
var InstallDir string
var DefaultInstallDir string
var DefaultConfigPath string

// extensions

var ConfigExtension string
var ModDataExtensions []string
var VariablesExtensions []string
var AutoVariablesExtensions []string

// args

var DefaultDatabase string

// env vars

// EnvAppPrefix is the prefix for all app specific environment variables (e.g. ("STEAMPIPE_")
var EnvAppPrefix string

// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
var EnvInputVarPrefix string

// Update check
var VersionCheckHost string
var VersionCheckPath string
