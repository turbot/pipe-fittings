package filepaths

import (
	"fmt"
	"github.com/turbot/pipe-fittings/error_helpers"
	"os"
	"path/filepath"
)

var InstallDir string

func EnsureInstallSubDir(dirName string) string {
	subDir := getInstallSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

func getInstallSubDir(dirName string) string {
	if InstallDir == "" {
		panic(fmt.Errorf("cannot call any Steampipe directory functions before InstallDir is set"))
	}
	return filepath.Join(InstallDir, dirName)
}
