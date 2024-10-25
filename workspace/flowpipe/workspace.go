package flowpipe

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeWorkspace struct {
	workspace.WorkspaceBase[*flowpipe.ModResources]
}
