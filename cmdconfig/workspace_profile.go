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

	// set viper default for workspace profile, using STEAMPIPE_WORKSPACE env var
	SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, EnvVarTypeString)
	// set viper default for install dir, using STEAMPIPE_INSTALL_DIR env var
	SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, EnvVarTypeString)

	// resolve the workspace profile dir
	installDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	if err != nil {
		return nil, err
	}

	workspaceProfileDir, err := filepaths.WorkspaceProfileDir(installDir)
	if err != nil {
		return nil, err
	}

	// create loader and load ther workspace
	loader, err := steampipeconfig.NewWorkspaceProfileLoader[T](workspaceProfileDir)
	if err != nil {
		return nil, err
	}

	return loader, nil
}
