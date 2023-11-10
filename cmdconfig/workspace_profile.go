package cmdconfig

import (
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
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
	SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, EnvVarTypeString)
	// set viper default for install dir, using ArgInstallDir env var
	SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, EnvVarTypeString)

	globalWorkspaceProfileDir, err := getGlobalWorkspaceDir()
	if err != nil {
		return nil, err
	}

	localWorkspaceProfileDir, err := getLocalWorkspaceDir()
	if err != nil {
		return nil, err
	}

	// create loader and load the workspace
	loader, err := steampipeconfig.NewWorkspaceProfileLoader[T](globalWorkspaceProfileDir, localWorkspaceProfileDir)
	if err != nil {
		return nil, err
	}

	return loader, nil
}

func getGlobalWorkspaceDir() (string, error) {
	installDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	if err != nil {
		return "", err
	}
	return filepaths.GlobalWorkspaceProfileDir(installDir)
}

func getLocalWorkspaceDir() (string, error) {
	modDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgModLocation))
	if err != nil {
		return "", err
	}
	return filepaths.LocalWorkspaceProfileDir(modDir)
}
