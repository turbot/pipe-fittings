package modinstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
)

type InstallOpts struct {
	WorkspaceMod   *modconfig.Mod
	Command        string
	ModArgs        []string
	DryRun         bool
	Force          bool
	PluginVersions *modconfig.PluginVersionMap
	UpdateStrategy string
}

func NewInstallOpts(workspaceMod *modconfig.Mod, updateStrategy string, modsToInstall ...string) *InstallOpts {
	cmdName := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	opts := &InstallOpts{
		WorkspaceMod:   workspaceMod,
		DryRun:         viper.GetBool(constants.ArgDryRun),
		Force:          viper.GetBool(constants.ArgForce),
		ModArgs:        utils.TrimGitUrls(modsToInstall),
		Command:        cmdName,
		UpdateStrategy: updateStrategy,
	}
	opts.ModArgs = utils.TrimGitUrls(opts.ModArgs)
	return opts
}
