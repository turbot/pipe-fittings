package cmdconfig

import (
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/error_helpers"
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
	c.cmd.PersistentFlags().String(name, defaultValue, desc)
	error_helpers.FailOnError(viper.BindPFlag(name, c.cmd.PersistentFlags().Lookup(name)))
	// add the key to the list of keys to tildefy
	filePathViperKeys = append(filePathViperKeys, name)

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
