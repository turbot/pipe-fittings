package workspace

import (
	"fmt"

	"github.com/turbot/pipe-fittings/constants"
)

var (
	// TODO: this only works with Steampipe given the content. Flowpipe doesn't reach the point where this error is thrown.
	ErrorNoModDefinition = fmt.Errorf("This command requires a mod definition file (mod.sp) - could not find in the current directory tree.\n\nYou can either clone a mod repository or install a mod using %s and run this command from the cloned/installed mod directory.\nPlease refer to: https://steampipe.io/docs/mods/overview", constants.Bold("steampipe mod install"))
)
