package load_mod

import (
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/steampipeconfig"
)

func LoadWorkspaceProfile() (*steampipeconfig.WorkspaceProfileLoader, error) {
	// set viper default for workspace profile, using STEAMPIPE_WORKSPACE env var
	cmdconfig.SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, cmdconfig.EnvVarTypeString)
	// set viper default for install dir, using STEAMPIPE_INSTALL_DIR env var
	cmdconfig.SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, cmdconfig.EnvVarTypeString)

	// resolve the workspace profile dir
	installDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	if err != nil {
		return nil, err
	}

	workspaceProfileDir, err := filepaths.WorkspaceProfileDir(installDir)
	if err != nil {
		return nil, err
	}

	// create loader
	loader, err := steampipeconfig.NewWorkspaceProfileLoader(workspaceProfileDir)
	if err != nil {
		return nil, err
	}

	return loader, nil
}
