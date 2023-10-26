package cmdconfig

import (
	"fmt"
	"os"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/constants"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

//// BootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
//func BootstrapViper(loader *steampipeconfig.WorkspaceProfileLoader, cmd *cobra.Command) error {
//	if loader == nil {
//		return perr.BadRequestWithMessage("workspace profile loader cannot be nil")
//	}
//
//	// set defaults  for keys which do not have a corresponding command flag
//	setBaseDefaults()
//
//	// set defaults from defaultWorkspaceProfile
//	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap(cmd))
//
//	// set defaults for install dir and mod location from env vars
//	// this needs to be done since the workspace profile definitions exist in the
//	// default install dir
//	setDirectoryDefaultsFromEnv()
//
//	// NOTE: if an explicit workspace profile was set, default the mod location and install dir _now_
//	// All other workspace profile values are defaults _after defaulting to the connection config options
//	// to give them higher precedence, but these must be done now as subsequent operations depend on them
//	// (and they cannot be set from hcl options)
//	if loader.ConfiguredProfile != nil {
//		if loader.ConfiguredProfile.ModLocation != nil {
//			log.Printf("[TRACE] setting mod location from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.ModLocation)
//			viper.SetDefault(constants.ArgModLocation, *loader.ConfiguredProfile.ModLocation)
//		}
//		if loader.ConfiguredProfile.InstallDir != nil {
//			log.Printf("[TRACE] setting install dir from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.InstallDir)
//			viper.SetDefault(constants.ArgInstallDir, *loader.ConfiguredProfile.InstallDir)
//		}
//	}
//
//	// tildefy all paths in viper
//	return TildefyPaths()
//}

// TildefyPaths cleans all path config values and replaces '~' with the home directory
func TildefyPaths() error {
	pathArgs := []string{
		constants.ArgModLocation,
		constants.ArgInstallDir,
		constants.ArgModLocation,
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

// SetDefaultsFromConfig overrides viper default values from hcl config values
func SetDefaultsFromConfig(configMap map[string]interface{}) {
	for k, v := range configMap {
		viper.SetDefault(k, v)
	}
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
// any option setting, workspace profile property or env var which does not have a command line
// MUST have a default (unless we want the zero value to take effect)
func setBaseDefaults() {
	defaults := map[string]interface{}{
		// global general options
		constants.ArgTelemetry:   constants.TelemetryInfo,
		constants.ArgUpdateCheck: true,

		// workspace profile
		constants.ArgAutoComplete:  true,
		constants.ArgIntrospection: constants.IntrospectionNone,

		// from global database options
		constants.ArgDatabasePort:         constants.DatabaseDefaultPort,
		constants.ArgDatabaseStartTimeout: constants.DBStartTimeout.Seconds(),
		constants.ArgServiceCacheEnabled:  true,
		constants.ArgCacheMaxTtl:          300,
		constants.ArgMaxCacheSizeMb:       constants.DefaultMaxCacheSizeMb,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar []string
	varType   EnvVarType
}

// set default values of INSTALL_DIR and ModLocation from env vars
func setDirectoryDefaultsFromEnv() {
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, EnvVarTypeString},
		constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, EnvVarTypeString},
		constants.EnvModLocation:    {[]string{constants.ArgModLocation}, EnvVarTypeString},
	}

	for envVar, mapping := range envMappings {
		SetConfigFromEnv(envVar, mapping.configVar, mapping.varType)
	}
}

// TODO KAI look at this
//// set default values from env vars
//func SetDefaultsFromEnv() {
//	// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
//	// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)
//
//	// a map of known environment variables to map to viper keys
//	envMappings := map[string]envMapping{
//		constants.EnvInstallDir:           {[]string{constants.ArgInstallDir}, cmdconfig.EnvVarTypeString},
//		constants.EnvWorkspaceChDir:       {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
//		constants.EnvModLocation:          {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
//		constants.EnvIntrospection:        {[]string{constants.ArgIntrospection}, cmdconfig.EnvVarTypeString},
//		constants.EnvTelemetry:            {[]string{constants.ArgTelemetry}, cmdconfig.EnvVarTypeString},
//		constants.EnvUpdateCheck:          {[]string{constants.ArgUpdateCheck}, cmdconfig.EnvVarTypeBool},
//		constants.EnvCloudHost:            {[]string{constants.ArgCloudHost}, cmdconfig.EnvVarTypeString},
//		constants.EnvCloudToken:           {[]string{constants.ArgCloudToken}, cmdconfig.EnvVarTypeString},
//		constants.EnvSnapshotLocation:     {[]string{constants.ArgSnapshotLocation}, cmdconfig.EnvVarTypeString},
//		constants.EnvWorkspaceDatabase:    {[]string{constants.ArgWorkspaceDatabase}, cmdconfig.EnvVarTypeString},
//		constants.EnvServicePassword:      {[]string{constants.ArgServicePassword}, cmdconfig.EnvVarTypeString},
//		constants.EnvCheckDisplayWidth:    {[]string{constants.ArgCheckDisplayWidth}, cmdconfig.EnvVarTypeInt},
//		constants.EnvMaxParallel:          {[]string{constants.ArgMaxParallel}, cmdconfig.EnvVarTypeInt},
//		constants.EnvQueryTimeout:         {[]string{constants.ArgDatabaseQueryTimeout}, cmdconfig.EnvVarTypeInt},
//		constants.EnvDatabaseStartTimeout: {[]string{constants.ArgDatabaseStartTimeout}, cmdconfig.EnvVarTypeInt},
//		constants.EnvCacheTTL:             {[]string{constants.ArgCacheTtl}, cmdconfig.EnvVarTypeInt},
//		constants.EnvCacheMaxTTL:          {[]string{constants.ArgCacheMaxTtl}, cmdconfig.EnvVarTypeInt},
//
//		// we need this value to go into different locations
//		constants.EnvCacheEnabled: {[]string{
//			constants.ArgClientCacheEnabled,
//			constants.ArgServiceCacheEnabled,
//		}, cmdconfig.EnvVarTypeBool},
//	}
//
//	for envVar, v := range envMappings {
//		SetConfigFromEnv(envVar, v.configVar, v.varType)
//	}
//}

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
