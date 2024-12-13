package filepaths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/pipe-fittings/error_helpers"
)

// PipesInstallDir is the location of config files common between pipelings
// this must be set by the application at startup
var DefaultPipesInstallDir = ""
var PipesInstallDir = ""

func ensurePipesInstallSubDir(dirName string) string {
	subDir := pipesInstallSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

func pipesInstallSubDir(dirName string) string {
	if PipesInstallDir == "" {
		panic(fmt.Errorf("cannot call any pipes directory functions before PipesInstallDir is set"))
	}
	return filepath.Join(PipesInstallDir, dirName)
}

// PipesInternalDir returns the path to the pipes internal directory (creates if missing)
func PipesInternalDir() string {
	return pipesInstallSubDir("internal")
}

// EnsurePipesInternalDir returns the path to the pipes internal directory (creates if missing)
func EnsurePipesInternalDir() string {
	return ensurePipesInstallSubDir("internal")
}

// EnsurePipesDuckDbExtensionsDir returns the path to the pipes duckdb extensions directory (creates if missing)
func EnsurePipesDuckDbExtensionsDir() string {
	return ensurePipesInstallSubDir(filepath.Join("extensions", "duckdb"))
}
