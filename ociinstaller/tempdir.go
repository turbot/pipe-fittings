package ociinstaller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/pipe-fittings/v2/error_helpers"
	pfilepaths"github.com/turbot/pipe-fittings/v2/filepaths"
)

type tempDir struct {
	Path string
}

// NewTempDir creates a directory under the given parent directory.
func NewTempDir(parent string) *tempDir {
	return &tempDir{
		Path: getOrCreateTempDir(parent),
	}
}

func getOrCreateTempDir(parent string) string {
	cacheDir := filepath.Join(parent, pfilepaths.SafeDirName(fmt.Sprintf("tmp-%s", pfilepaths.GenerateTempDirName())))

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create cache directory")
	}
	return cacheDir
}

func (d *tempDir) Delete() error {
	return os.RemoveAll(d.Path)
}
