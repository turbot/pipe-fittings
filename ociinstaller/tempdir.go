package ociinstaller

import (
	"log"
	"os"
	"path/filepath"
	localfilepaths "github.com/turbot/pipe-fittings/v2/filepaths"
	"fmt"
)

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
		isEmpty, err := localfilepaths.IsDirEmpty(parentTempDir)
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
