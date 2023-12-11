package cmdconfig

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/error_helpers"
	"log/slog"
)

// persistent flags

// AddPersistentStringFlag is a helper function to add a string flag to a command
func (c *CmdBuilder) AddPersistentStringFlag(name string, defaultValue string, desc string) *CmdBuilder {
	c.cmd.PersistentFlags().String(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c
}

// AddPersistentFilepathFlag is a helper function to add a persistent string flag containing a file path to a command
// Note:this also stores the config key in filePathViperKeys,
// ensuring the value of the key has `~` converted to the home dir
func (c *CmdBuilder) AddPersistentFilepathFlag(name string, defaultValue string, desc string) *CmdBuilder {
	defaultValue, err := filehelpers.Tildefy(defaultValue)
	error_helpers.FailOnError(err)

	c.cmd.PersistentFlags().String(name, defaultValue, desc)
	s := viper.IsSet(name)
	slog.Debug("is set", "name", name, "isSet", s)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	// add the key to the map of keys to tildefy
	filePathViperKeys[name] = struct{}{}

	return c
}

// AddPersistentIntFlag is a helper function to add an integer flag to a command
func (c *CmdBuilder) AddPersistentIntFlag(name string, defaultValue int, desc string) *CmdBuilder {
	c.cmd.PersistentFlags().Int(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c

}

// AddPersistentBoolFlag ia s helper function to add a boolean flag to a command
func (c *CmdBuilder) AddPersistentBoolFlag(name string, defaultValue bool, desc string) *CmdBuilder {
	c.cmd.PersistentFlags().Bool(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c
}

// AddPersistentStringSliceFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddPersistentStringSliceFlag(name string, defaultValue []string, desc string) *CmdBuilder {
	c.cmd.PersistentFlags().StringSlice(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c
}

// AddPersistentStringArrayFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddPersistentStringArrayFlag(name string, defaultValue []string, desc string) *CmdBuilder {
	c.cmd.PersistentFlags().StringArray(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c
}

func (c *CmdBuilder) AddPersistentVarFlag(value pflag.Value, name string, usage string) *CmdBuilder {
	c.cmd.PersistentFlags().Var(value, name, usage)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	return c
}
