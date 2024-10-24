package flowpipe

import (
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeWorkspace struct {
	workspace.WorkspaceBase[*flowpipe.Mod]
	// Credentials are something different, it's not part of the mod, it's not part of the workspace, it is at the same level
	// with mod and workspace. However, it can be referenced by the mod, so it needs to be in the parse context
	Credentials         map[string]credential.Credential
	PipelingConnections map[string]connection.PipelingConnection
	Integrations        map[string]flowpipe.Integration
	Notifiers           map[string]flowpipe.Notifier
}
