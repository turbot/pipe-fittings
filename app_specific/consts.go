package app_specific

import (
	"github.com/Masterminds/semver/v3"
)

// Application specific constants which MUST be set by the application

// app name and version

var AppName string
var AppVersion *semver.Version

// filepaths

var DefaultVarsFileName string

// TODO KAI  we need to provide a default (for now) as the flowpipe test code does not (always)
// call SetAppSpecificConstants so this may be empty we need a proper solution to this
var ModFileName string = "mod.sp"

var WorkspaceIgnoreFile string
var WorkspaceDataDir string
var InstallDir string
var DefaultInstallDir string
var DefaultConfigPath string

// db client app names

var ClientConnectionAppNamePrefix string
var ServiceConnectionAppNamePrefix string
var ClientSystemConnectionAppNamePrefix string

// extensions

var ConfigExtension string
var ModDataExtension string
var VariablesExtension string
var AutoVariablesExtension string

// args

var DefaultWorkspaceDatabase string

// env vars

// EnvAppPrefix is the prefix for all app specific environment variables (e.g. ("STEAMPIPE_")
var EnvAppPrefix string

// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
var EnvInputVarPrefix string
