package cmdconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	parse2 "github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
	app_specific2 "github.com/turbot/pipe-helpers/app_specific"
	"github.com/turbot/pipe-helpers/constants"
	"github.com/turbot/pipe-helpers/files"
)

// GetWorkspaceProfileLoader creates a WorkspaceProfileLoader which loads the configured workspace
func GetWorkspaceProfileLoader[T workspace_profile.WorkspaceProfile](parseOpts ...parse2.ParseHclOpt) (*parse2.WorkspaceProfileLoader[T], error) {
	// NOTE: we need to setup some viper defaults to enable workspace profile loading
	// the rest are set up in BootstrapViper

	// set viper default for workspace profile, using ArgWorkspaceProfile env var
	SetDefaultFromEnv(app_specific2.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, EnvVarTypeString)
	// set viper default for install dir, using ArgInstallDir env var
	SetDefaultFromEnv(app_specific2.EnvInstallDir, constants.ArgInstallDir, EnvVarTypeString)

	// create loader and load the workspace
	configPaths, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	loader, err := parse2.NewWorkspaceProfileLoader[T](configPaths...)
	if err != nil {
		return nil, err
	}

	// set the parse options (if any)
	loader.ParseOpts = parseOpts

	// do the load
	if err = loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}

// GetConfigPath builds a list of possible config file locations, starting with the HIGHEST priority
func GetConfigPath() ([]string, error) {
	configPathArg := app_specific2.DefaultConfigPath

	// config-path is a colon separated path of decreasing precedence that config (fpc) files are loaded from
	// default to the cmod location and the global config dir
	configPathEnv := app_specific2.EnvConfigPath
	if envVal, ok := os.LookupEnv(configPathEnv); ok {
		configPathArg = envVal
	}
	if viper.IsSet(constants.ArgConfigPath) {
		configPathArg = viper.GetString(constants.ArgConfigPath)
	}
	if len(configPathArg) == 0 {
		return nil, fmt.Errorf("no config path specified")
	}
	configPaths := strings.Split(configPathArg, ":")

	for i, p := range configPaths {
		// special case for "." - use the mod location
		if p == "." {
			p = viper.GetString(constants.ArgModLocation)
		}
		absPath, err := files.Tildefy(p)
		if err != nil {
			return nil, err
		}
		configPaths[i] = absPath

	}
	return configPaths, nil
}
