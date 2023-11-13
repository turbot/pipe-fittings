package cmdconfig

import (
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"os"
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

// AddFilepathFlag is a helper function to add a string flag containing a file path to a command
// Note:this also stores the config key in filePathViperKeys,
// ensuring the value of the key has `~` converted to the home dir
func (c *CmdBuilder) AddFilepathFlag(name string, defaultValue string, desc string, opts ...FlagOption) *CmdBuilder {
	c.cmd.Flags().String(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}

	filePathViperKeys = append(filePathViperKeys, name)

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
		AddStringFlag(constants.ArgCloudHost, constants.DefaultCloudHost, "Turbot Pipes host").
		AddStringFlag(constants.ArgCloudToken, "", "Turbot Pipes authentication token")
}

// AddWorkspaceDatabaseFlag is helper function to add the workspace-databse flag to a command
func (c *CmdBuilder) AddWorkspaceDatabaseFlag() *CmdBuilder {
	return c.
		AddStringFlag(constants.ArgWorkspaceDatabase, app_specific.DefaultWorkspaceDatabase, "Turbot Pipes workspace database")
}

// AddModLocationFlag is helper function to add the mod-location flag to a command
func (c *CmdBuilder) AddModLocationFlag() *CmdBuilder {
	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)
	return c.AddFilepathFlag(constants.ArgModLocation, cwd, "Path to the workspace working directory")
}
