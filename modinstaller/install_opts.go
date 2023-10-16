package modinstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
)

type InstallOpts struct {
	WorkspaceMod *modconfig.Mod
	Command      string
	ModArgs      []string
	DryRun       bool
	Force        bool
	GitUrlMode   GitUrlMode
}

func NewInstallOpts(workspaceMod *modconfig.Mod, modsToInstall ...string) *InstallOpts {
	cmdName := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	opts := &InstallOpts{
		WorkspaceMod: workspaceMod,
		DryRun:       viper.GetBool(constants.ArgDryRun),
		Force:        viper.GetBool(constants.ArgForce),
		ModArgs:      modsToInstall,
		Command:      cmdName,
	}
	return opts
}
