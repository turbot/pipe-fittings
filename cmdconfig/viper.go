package cmdconfig

import (
	"fmt"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/error_helpers"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/steampipeconfig"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// BootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func BootstrapViper[T modconfig.WorkspaceProfile](loader *steampipeconfig.WorkspaceProfileLoader[T], cmd *cobra.Command, opts ...bootstrapOption) {
	config := newBootstrapConfig()
	for _, opt := range opts {
		opt(config)
	}

	// set defaults  for keys which do not have a corresponding command flag
	setBaseDefaults(config.configDefaults)

	// set defaults from defaultWorkspaceProfile
	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap(cmd))

	// set defaults for install dir and mod location from env vars
	// this needs to be done since the workspace profile definitions exist in the
	// default install dir
	SetDefaultsFromEnv(config.directoryEnvMappings)

	// NOTE: if an explicit workspace profile was set, default the install dir _now_
	// All other workspace profile values are defaults _after defaulting to the connection config options
	// to give them higher precedence, but these must be done now as subsequent operations depend on them
	// (and they cannot be set from hcl options)
	if !loader.ConfiguredProfile.IsNil() {
		if installDir := loader.ConfiguredProfile.GetInstallDir(); installDir != nil {
			slog.Debug("setting install dir", "configured profile", loader.ConfiguredProfile.Name(), "install dir", *installDir)
			setDefault(constants.ArgInstallDir, *installDir)
		}
	}
	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Debug("mod directory", "modLocation", modLocation)

}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
// any option setting, workspace profile property or env var which does not have a command line
// MUST have a default (unless we want the zero value to take effect)
func setBaseDefaults(configDefaults map[string]any) {
	for k, v := range configDefaults {
		setDefault(k, v)
	}
}

// set the viper default for the given key
// NOTE: if the key is a filepath, tildefy it
func setDefault(k string, v any) {
	if _, isFilePath := filePathViperKeys[k]; isFilePath {
		var err error
		v, err = filehelpers.Tildefy(v.(string))
		error_helpers.FailOnError(err)
	}
	viper.SetDefault(k, v)
}

func SetDefaultsFromEnv(envMappings map[string]EnvMapping) {
	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Debug("mod directory", "modLocation", modLocation)

	for envVar, mapping := range envMappings {
		SetConfigFromEnv(envVar, mapping.ConfigVar, mapping.VarType)
	}
	modLocation = viper.GetString(constants.ArgModLocation)
	slog.Debug("mod directory", "modLocation", modLocation)

}

// SetDefaultsFromConfig overrides viper default values from hcl config values
func SetDefaultsFromConfig(configMap map[string]any) {
	for k, v := range configMap {
		setDefault(k, v)
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
			setDefault(configVar, val)
		case EnvVarTypeBool:
			if boolVal, err := types.ToBool(val); err == nil {
				setDefault(configVar, boolVal)
			}
		case EnvVarTypeInt:
			if intVal, err := types.ToInt64(val); err == nil {
				setDefault(configVar, intVal)
			}
		default:
			// must be an invalid value in the map above
			panic(fmt.Sprintf("invalid env var mapping type: %v", varType))
		}
	}
}
