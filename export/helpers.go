package export

import (
	"fmt"
	"io"
	"os"
	"time"
)

func GenerateDefaultExportFileName(executionName, fileExtension string) string {
	now := time.Now()
	timeFormatted := fmt.Sprintf("%d%02d%02dT%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return fmt.Sprintf("%s.%s%s", executionName, timeFormatted, fileExtension)
}

// Write creates the file and copies exportData into it.
// If the copy fails, the incomplete file is removed.
func Write(filePath string, exportData io.Reader) error {
	// create the output file
	destination, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(destination, exportData)
	if closeErr := destination.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(filePath)
	}
	return err
}
