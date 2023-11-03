package app_specific

import "github.com/Masterminds/semver/v3"

// Application specific constants which MUST be set by the application

var AppName string
var AppVersion *semver.Version

var DefaultVarsFileName string
var ModFileName string
var WorkspaceIgnoreFile string
var WorkspaceDataDir string

var ClientConnectionAppNamePrefix string
var ServiceConnectionAppNamePrefix string
var ClientSystemConnectionAppNamePrefix string

var DefaultInstallDir string
var DefaultWorkspaceDatabase string

var ModDataExtension string
var VariablesExtension string
var AutoVariablesExtension string

// EnvAppPrefix is the prefix for all app specific environment variables (e.g. ("STEAMPIPE_")
var EnvAppPrefix string

// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
var EnvInputVarPrefix string
