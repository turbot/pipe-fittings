package filepaths

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/turbot/pipe-fittings/v2/utils"
)

func EnsurePidTempDir(parent string) string {
	// add a PID directory to the collection directory
	pidTempDir := filepath.Join(parent, fmt.Sprintf("%d", os.Getpid()))

	// create the directory if it doesn't exist
	if _, err := os.Stat(pidTempDir); os.IsNotExist(err) {
		err := os.MkdirAll(pidTempDir, 0755)
		if err != nil {
			slog.Error("failed to create collection temp dir", "error", err)
		}
	}
	return pidTempDir
}


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
			DeleteTempDir(filepath.Join(dir, file.Name()))
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

func DeleteTempDir(tempDir string) {
	// Remove the specific temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		log.Printf("[TRACE] Failed to delete temp dir '%s' after installing plugin: %s", tempDir, err)
		return
	}

	// Check if the parent temp directory is empty and clean it up if so
	parentTempDir := filepath.Dir(tempDir)
	fmt.Println("Parent temp dir", parentTempDir)
	if filepath.Base(parentTempDir) == "temp" {
		isEmpty, err := IsDirEmpty(parentTempDir)
		if err != nil {
			log.Printf("[TRACE] Failed to check if temp parent dir '%s' is empty: %s", parentTempDir, err)
			return
		}

		if isEmpty {
			if err := os.Remove(parentTempDir); err != nil {
				log.Printf("[TRACE] Failed to remove empty temp parent dir '%s': %s", parentTempDir, err)
			} else {
				log.Printf("[TRACE] Cleaned up empty temp parent dir '%s'", parentTempDir)
			}
		}
	}
}
