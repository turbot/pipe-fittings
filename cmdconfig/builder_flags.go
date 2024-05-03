package cmdconfig

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
)

// basic flags

// AddStringFlag is a helper function to add a string flag to a command
func (c *CmdBuilder) AddStringFlag(name string, defaultValue string, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().String(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}

	return c
}

// AddIntFlag is a helper function to add an integer flag to a command
func (c *CmdBuilder) AddIntFlag(name string, defaultValue int, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().Int(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddBoolFlag ia s helper function to add a boolean flag to a command
func (c *CmdBuilder) AddBoolFlag(name string, defaultValue bool, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().Bool(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringSliceFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringSliceFlag(name string, defaultValue []string, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().StringSlice(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringArrayFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringArrayFlag(name string, defaultValue []string, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().StringArray(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringMapStringFlag is a helper function to add a flag that accepts a map of strings
func (c *CmdBuilder) AddStringMapStringFlag(name string, defaultValue map[string]string, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().StringToString(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// specific flags

// AddCloudFlags is helper function to add the cloud flags to a command
func (c *CmdBuilder) AddCloudFlags() *CmdBuilder {
	return c.
		AddStringFlag(constants.ArgPipesHost, constants.DefaultPipesHost, "Turbot Pipes host").
		AddStringFlag(constants.ArgPipesToken, "", "Turbot Pipes authentication token")
}

// AddModLocationFlag is helper function to add the mod-location flag to a command
func (c *CmdBuilder) AddModLocationFlag() *CmdBuilder {
	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)
	return c.AddStringFlag(constants.ArgModLocation, cwd, "Sets the workspace working directory. If not specified, the workspace directory will be set to the current working directory.")
}

func (c *CmdBuilder) AddVarFlag(value pflag.Value, name string, usage string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().Var(value, name, usage)

	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}
