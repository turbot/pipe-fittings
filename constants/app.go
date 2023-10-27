package constants

import "github.com/Masterminds/semver/v3"

// Application constants which MUST be set by the application

var AppName string
var ClientConnectionAppNamePrefix string
var ServiceConnectionAppNamePrefix string
var ClientSystemConnectionAppNamePrefix string
var DefaultInstallDir string
var AppVersion *semver.Version
