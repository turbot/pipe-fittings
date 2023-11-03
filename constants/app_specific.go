package constants

import "github.com/Masterminds/semver/v3"

// Application specific constants which MUST be set by the application

var AppName string
var ClientConnectionAppNamePrefix string
var ServiceConnectionAppNamePrefix string
var ClientSystemConnectionAppNamePrefix string
var DefaultInstallDir string
var AppVersion *semver.Version
var DefaultWorkspaceDatabase string
var ModDataExtension string
var VariablesExtension string
var AutoVariablesExtension string

// Pipes Components overrides
var PipesComponentModDataExtension = ModDataExtension
