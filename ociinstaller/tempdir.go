package ociinstaller

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/pipe-fittings/v2/error_helpers"

	"github.com/google/uuid"
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
	// Create temp subdirectory under parent if it doesn't exist
	tempParentDir := filepath.Join(parent, "temp")
	if _, err := os.Stat(tempParentDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempParentDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create temp parent directory")
	}

	// Create the actual temp directory
	cacheDir := filepath.Join(tempParentDir, safeDirName(generateTempDirName()))

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, "could not create cache directory")
	}
	return cacheDir
}

func (d *tempDir) Delete() error {
	return os.RemoveAll(d.Path)
}

func safeDirName(dirName string) string {
	newName := strings.ReplaceAll(dirName, "/", "_")
	newName = strings.ReplaceAll(newName, ":", "@")

	return newName
}

func generateTempDirName() string {
	u, err := uuid.NewRandom()
	if err != nil {
		// Should never happen?
		panic(err)
	}
	s := u.String()
	return s[9:23]
}
