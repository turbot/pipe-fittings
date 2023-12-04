package cmdconfig

import (
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/steampipeconfig"
)

// GetWorkspaceProfileLoader creates a WorkspaceProfileLoader which loads the configured workspace
func GetWorkspaceProfileLoader[T modconfig.WorkspaceProfile]() (*steampipeconfig.WorkspaceProfileLoader[T], error) {
	// NOTE: we need to setup some viper defaults to enable workspace profile loading
	// the rest are set up in BootstrapViper

	// set viper default for workspace profile, using ArgWorkspaceProfile env var
	SetDefaultFromEnv(app_specific.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, EnvVarTypeString)
	// set viper default for install dir, using ArgInstallDir env var
	SetDefaultFromEnv(app_specific.EnvInstallDir, constants.ArgInstallDir, EnvVarTypeString)

	// create loader and load the workspace
	configPaths, err := getWorkspaceLocations()
	if err != nil {
		return nil, err
	}
	loader, err := steampipeconfig.NewWorkspaceProfileLoader[T](configPaths...)
	if err != nil {
		return nil, err
	}

	return loader, nil
}

// build list of possible workspace locations
func getWorkspaceLocations() ([]string, error) {
	// if config-path was passed, use that
	configPaths := viper.GetStringSlice(constants.ArgConfigPath)
	if len(configPaths) > 0 {
		for i, p := range configPaths {
			absPath, err := files.Tildefy(p)
			if err != nil {
				return nil, err
			}
			configPaths[i] = absPath

		}
		return configPaths, nil
	}

	return []string{
		filepaths.GlobalWorkspaceProfileDir(viper.GetString(constants.ArgInstallDir)),
		viper.GetString(constants.ArgModLocation),
	}, nil
}
