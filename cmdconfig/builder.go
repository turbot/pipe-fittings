package cmdconfig

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/utils"
)

var CustomPreRunHook func(cmd *cobra.Command, args []string) error
var CustomPostRunHook func(cmd *cobra.Command, args []string) error

type CmdBuilder struct {
	cmd      *cobra.Command
	bindings map[string]*pflag.Flag
}

// OnCmd starts a config builder wrapping over the provided *cobra.Command
func OnCmd(cmd *cobra.Command) *CmdBuilder {
	cfg := new(CmdBuilder)
	cfg.cmd = cmd
	cfg.bindings = map[string]*pflag.Flag{}

	setPreRunHook(cfg)

	setPostRunHook(cfg)

	// wrap over the original Run function
	originalRun := cfg.cmd.Run
	cfg.cmd.Run = func(cmd *cobra.Command, args []string) {
		utils.LogTime(fmt.Sprintf("cmd.%s.Run start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.Run end", cmd.CommandPath()))

		// run the original Run
		if originalRun != nil {
			originalRun(cmd, args)
		}
	}

	return cfg
}

func setPreRunHook(cfg *CmdBuilder) {
	/* update the command pre run hook to:
	 	- bind command flags
		- run any custom pre run hook
		- run the existing pre run hook
	*/

	// we will wrap over these two function - need references to call them

	// override PreRunE no PreRun as this has precedence
	originalPreRun := cfg.cmd.PreRunE
	if originalPreRun == nil && cfg.cmd.PreRun != nil {
		originalPreRun = func(cmd *cobra.Command, args []string) error {
			cfg.cmd.PreRun(cmd, args)
			return nil
		}
	}
	cfg.cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		utils.LogTime(fmt.Sprintf("cmd.%s.PreRun start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.PreRun end", cmd.CommandPath()))
		// bind flags
		for flagName, flag := range cfg.bindings {
			if err := viper.GetViper().BindPFlag(flagName, flag); err != nil {
				// we can panic here since this is bootstrap code and not execution path specific
				panic(fmt.Sprintf("flag for %s cannot be nil", flagName))
			}
		}

		// now that we have done all the flag bindings, run the custom pre run hook (if set)
		if CustomPreRunHook != nil {
			if err := CustomPreRunHook(cmd, args); err != nil {
				return err
			}
		}

		// run the original PreRun
		if originalPreRun != nil {
			return originalPreRun(cmd, args)
		}

		return nil
	}
}

func setPostRunHook(cfg *CmdBuilder) {
	// override PostRunE not PostRun as this has precedence
	originalPostRun := cfg.cmd.PostRunE

	if originalPostRun == nil && cfg.cmd.PostRun != nil {
		originalPostRun = func(cmd *cobra.Command, args []string) error {
			cfg.cmd.PostRun(cmd, args)
			return nil
		}
	}

	cfg.cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		utils.LogTime(fmt.Sprintf("cmd.%s.PostRun start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.PostRun end", cmd.CommandPath()))
		// run the original PostRun
		if originalPostRun != nil {
			if err := originalPostRun(cmd, args); err != nil {
				return err
			}
		}

		// run the custom post run hook (if there is one)
		if CustomPostRunHook != nil {
			return CustomPostRunHook(cmd, args)
		}
		return nil
	}
}
