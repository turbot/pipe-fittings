package powerpipe

import (
	"github.com/turbot/pipe-fittings/modconfig/powerpipe"
	"github.com/turbot/pipe-fittings/workspace"
)

type PowerpipeWorkspace struct {
	workspace.WorkspaceBase[*powerpipe.Mod]
}
