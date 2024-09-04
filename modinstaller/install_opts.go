package modinstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/plugin"
	"github.com/turbot/pipe-fittings/utils"
)

// TODO KAI why does powerpipe care about plugins???

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
	cmdName := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()

	// for install command, if there is a target mod, and if the pull strategy has not been explicitly set, set it to latest
	if cmdName == "install" && len(modsToInstall) > 0 && !viper.IsSet(constants.ArgPull) {
		viper.Set(constants.ArgPull, constants.ModUpdateLatest)
	}
	// for uninstall default to minimal
	if cmdName == "uninstall" {
		viper.Set(constants.ArgPull, constants.ModUpdateIdMinimal)
	}

	opts := &InstallOpts{
		WorkspaceMod:   workspaceMod,
		DryRun:         viper.GetBool(constants.ArgDryRun),
		Force:          viper.GetBool(constants.ArgForce),
		ModArgs:        utils.TrimGitUrls(modsToInstall),
		Command:        cmdName,
		UpdateStrategy: viper.GetString(constants.ArgPull),
	}

	opts.ModArgs = utils.TrimGitUrls(opts.ModArgs)
	return opts
}
