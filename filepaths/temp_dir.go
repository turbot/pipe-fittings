package filepaths

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/turbot/pipe-fittings/v2/utils"
)

// EnsurePidTempDir creates a temporary directory with the current process ID as the name, it returns the path to the temporary directory
func EnsurePidTempDir(parent string) string {
	// add a PID directory followed by a uuid to the collection directory
	pidTempDir := filepath.Join(parent, fmt.Sprintf("%d", os.Getpid()), SafeDirName(GenerateTempDirName()))

	// create the directory if it doesn't exist
	if _, err := os.Stat(pidTempDir); os.IsNotExist(err) {
		err := os.MkdirAll(pidTempDir, 0755)
		if err != nil {
			slog.Error("failed to create collection temp dir", "error", err)
		}
	}
	return pidTempDir
}

// CleanupPidTempDirs cleans up the temporary directories passed to it
func CleanupPidTempDirs(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		slog.Warn("failed to list files in dir", "error", err)
		return
	}
	for _, file := range files {
		// if the file is a directory and is not our temp dir, remove it
		if file.IsDir() {
			// the folder name is the PID - check whether that pid exists
			// if it doesn't, remove the folder
			// Attempt to find the process
			// try to parse the directory name as a pid
			pid, err := strconv.ParseInt(file.Name(), 10, 32)
			if err == nil {
				if utils.PidExists(int(pid)) {
					slog.Info(fmt.Sprintf("Cleaning existing temp dirs - skipping directory '%s' as process with PID %d exists", file.Name(), pid))
					continue
				}
			}
			slog.Debug("Removing directory", "dir", file.Name())
			RemoveDirAndEmptyParents(filepath.Join(dir, file.Name()))
		}
	}
}

// isDirEmpty checks if a directory is empty.
func IsDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// RemoveDirAndEmptyParents removes the given directory and cleans up empty parent
// directories up to (and including) the 'temp' directory.
func RemoveDirAndEmptyParents(tempDir string) {
	// Remove the specific temp directory (e.g., the uuid dir or a pid dir)
	if err := os.RemoveAll(tempDir); err != nil {
		slog.Warn("failed to delete temp dir after installing plugin", "dir", tempDir, "error", err)
		return
	}

	// Walk up and remove empty parent directories until we reach the 'temp' dir
	parentDir := filepath.Dir(tempDir)
	for {
		base := filepath.Base(parentDir)

		isEmpty, err := IsDirEmpty(parentDir)
		if err != nil {
			slog.Warn("failed to check if parent dir is empty", "dir", parentDir, "error", err)
			break
		}
		if !isEmpty {
			// parent has other contents (e.g., other UUID dirs) -> stop
			break
		}

		if err := os.Remove(parentDir); err != nil {
			slog.Warn("failed to remove empty parent dir", "dir", parentDir, "error", err)
			break
		} else {
			slog.Debug("cleaned up empty parent dir", "dir", parentDir)
		}

		// if we just removed the 'temp' directory, stop
		if base == "temp" {
			break
		}

		// continue walking up (uuid -> pid -> temp)
		parentDir = filepath.Dir(parentDir)
	}
}

func SafeDirName(dirName string) string {
	newName := strings.ReplaceAll(dirName, "/", "_")
	newName = strings.ReplaceAll(newName, ":", "@")

	return newName
}

func GenerateTempDirName() string {
	u, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	s := u.String()
	return s[9:23]
}
