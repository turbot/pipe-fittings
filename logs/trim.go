package logs

import (
	"context"
	"github.com/turbot/pipe-fittings/constants"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/pipe-fittings/filepaths"
)

const logRetentionDays = 7

func TrimLogs() {
	fileLocation := filepaths.EnsureLogDir()
	files, err := os.ReadDir(fileLocation)
	if err != nil {
		slog.Log(context.Background(), constants.LevelTrace, "error listing db log directory", err)
	}
	for _, file := range files {
		fi, err := file.Info()
		if err != nil {
			slog.Log(context.Background(), constants.LevelTrace, "error reading file info. continuing", "file", file.Name())
			continue
		}

		fileName := fi.Name()
		if filepath.Ext(fileName) != ".log" {
			continue
		}

		age := time.Since(fi.ModTime()).Hours()
		if age > logRetentionDays*24 {
			logPath := filepath.Join(fileLocation, fileName)
			err := os.Remove(logPath)
			if err != nil {
				slog.Log(context.Background(), constants.LevelTrace, "failed to delete log file %s\n", logPath)
			}
		}
	}
}
