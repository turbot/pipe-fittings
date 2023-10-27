package cmdconfig

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/steampipeconfig"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// BootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func BootstrapViper(loader *steampipeconfig.WorkspaceProfileLoader, cmd *cobra.Command, configDefaults map[string]any, directoryEnvMappings map[string]EnvMapping) error {
	if loader == nil {
		return perr.BadRequestWithMessage("workspace profile loader cannot be nil")
	}

	// set defaults  for keys which do not have a corresponding command flag
	setBaseDefaults(configDefaults)

	// set defaults from defaultWorkspaceProfile
	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap(cmd))

	// set defaults for install dir and mod location from env vars
	// this needs to be done since the workspace profile definitions exist in the
	// default install dir
	SetDefaultsFromEnv(directoryEnvMappings)

	// NOTE: if an explicit workspace profile was set, default the mod location and install dir _now_
	// All other workspace profile values are defaults _after defaulting to the connection config options
	// to give them higher precedence, but these must be done now as subsequent operations depend on them
	// (and they cannot be set from hcl options)
	if loader.ConfiguredProfile != nil {
		if loader.ConfiguredProfile.ModLocation != nil {
			log.Printf("[TRACE] setting mod location from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.ModLocation)
			viper.SetDefault(constants.ArgModLocation, *loader.ConfiguredProfile.ModLocation)
		}
		if loader.ConfiguredProfile.InstallDir != nil {
			log.Printf("[TRACE] setting install dir from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.InstallDir)
			viper.SetDefault(constants.ArgInstallDir, *loader.ConfiguredProfile.InstallDir)
		}
	}

	// tildefy all paths in viper
	return TildefyPaths()
}

// TildefyPaths cleans all path config values and replaces '~' with the home directory
func TildefyPaths() error {
	pathArgs := []string{
		constants.ArgSnapshotLocation,
		constants.ArgModLocation,
		constants.ArgInstallDir,
		constants.ArgOutputDir,
		constants.ArgLogDir,
	}
	var err error
	for _, argName := range pathArgs {
		if argVal := viper.GetString(argName); argVal != "" {
			if argVal, err = filehelpers.Tildefy(argVal); err != nil {
				return err
			}
			if viper.IsSet(argName) {
				// if the value was already set re-set
				viper.Set(argName, argVal)
			} else {
				// otherwise just update the default
				viper.SetDefault(argName, argVal)
			}
		}
	}
	return nil
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
// any option setting, workspace profile property or env var which does not have a command line
// MUST have a default (unless we want the zero value to take effect)
func setBaseDefaults(configDefaults map[string]any) {
	for k, v := range configDefaults {
		viper.SetDefault(k, v)
	}
}

func SetDefaultsFromEnv(envMappings map[string]EnvMapping) {
	for envVar, mapping := range envMappings {
		SetConfigFromEnv(envVar, mapping.ConfigVar, mapping.VarType)
	}
}

// SetDefaultsFromConfig overrides viper default values from hcl config values
func SetDefaultsFromConfig(configMap map[string]any) {
	for k, v := range configMap {
		viper.SetDefault(k, v)
	}
}

type EnvMapping struct {
	ConfigVar []string
	VarType   EnvVarType
}

func SetConfigFromEnv(envVar string, configs []string, varType EnvVarType) {
	for _, configVar := range configs {
		SetDefaultFromEnv(envVar, configVar, varType)
	}
}

func SetDefaultFromEnv(k string, configVar string, varType EnvVarType) {
	if val, ok := os.LookupEnv(k); ok {
		switch varType {
		case EnvVarTypeString:
			viper.SetDefault(configVar, val)
		case EnvVarTypeBool:
			if boolVal, err := types.ToBool(val); err == nil {
				viper.SetDefault(configVar, boolVal)
			}
		case EnvVarTypeInt:
			if intVal, err := types.ToInt64(val); err == nil {
				viper.SetDefault(configVar, intVal)
			}
		default:
			// must be an invalid value in the map above
			panic(fmt.Sprintf("invalid env var mapping type: %v", varType))
		}
	}
}
