package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
)

// FindModFilePath search up the directory tree to find the modfile
func FindModFilePath(folder string) (string, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return "", err
	}
	for _, modFilePath := range app_specific.ModFilePaths(folder) {
		_, err = os.Stat(modFilePath)
		if err == nil {
			// found the modfile
			return modFilePath, nil
		}
	}

	// if the file wasn't found, search in the parent directory
	parent := filepath.Dir(folder)
	if folder == parent {
		// this typically means that we are already in the root directory
		return "", ErrorNoModDefinition
	}
	return FindModFilePath(filepath.Dir(folder))
}

func HomeDirectoryModfileCheck(ctx context.Context, workspacePath string) error {
	// bypass all the checks if ConfigKeyBypassHomeDirModfileWarning is set - it means home dir modfile check
	// has already happened before
	if viper.GetBool(constants.ConfigKeyBypassHomeDirModfileWarning) {
		return nil
	}
	// get the cmd and home dir
	home, _ := os.UserHomeDir()
	var modFileExists bool
	for _, modFilePath := range app_specific.ModFilePaths(workspacePath) {
		if _, err := os.Stat(modFilePath); err != nil {
			modFileExists = true
		}
	}
	// check if your workspace path is home dir and if modfile exists
	if workspacePath == home && modFileExists {
		// for other cmds - if home dir has modfile, just warn
		defaultModFileName := app_specific.ModFileNames()[0]
		warningText := fmt.Sprintf("You have a %s file in your home directory. This is not recommended.\nAs a result, %s will try to load all the files in home and its sub-directories, which can cause performance issues.\nBest practice is to put %s files in their own directories.\nHit Ctrl+C to stop.\n",
			defaultModFileName,
			app_specific.AppName,
			defaultModFileName)
		error_helpers.ShowWarning(warningText)
	}

	return nil
}
