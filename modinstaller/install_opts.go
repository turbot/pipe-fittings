package modinstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/plugin"
	constants2 "github.com/turbot/pipe-helpers/constants"
	"github.com/turbot/pipe-helpers/utils"
)

type InstallOpts struct {
	WorkspaceMod   *modconfig.Mod
	Command        string
	ModArgs        []string
	DryRun         bool
	Force          bool
	PluginVersions *plugin.PluginVersionMap
	UpdateStrategy string
}

func NewInstallOpts(workspaceMod *modconfig.Mod, modsToInstall ...string) *InstallOpts {
	cmdName := viper.Get(constants2.ConfigKeyActiveCommand).(*cobra.Command).Name()

	// for install command, if there is a target mod, and if the pull strategy has not been explicitly set, set it to latest
	if cmdName == "install" && len(modsToInstall) > 0 && !viper.IsSet(constants2.ArgPull) {
		viper.Set(constants2.ArgPull, constants2.ModUpdateLatest)
	}
	// for uninstall default to minimal
	if cmdName == "uninstall" {
		viper.Set(constants2.ArgPull, constants2.ModUpdateIdMinimal)
	}

	opts := &InstallOpts{
		WorkspaceMod:   workspaceMod,
		DryRun:         viper.GetBool(constants2.ArgDryRun),
		Force:          viper.GetBool(constants2.ArgForce),
		ModArgs:        utils.TrimGitUrls(modsToInstall),
		Command:        cmdName,
		UpdateStrategy: viper.GetString(constants2.ArgPull),
	}

	opts.ModArgs = utils.TrimGitUrls(opts.ModArgs)
	return opts
}
