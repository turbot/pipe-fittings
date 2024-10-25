package flowpipe

import (
	"context"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeWorkspace struct {
	workspace.WorkspaceBase[*flowpipe.FlowpipeResourceMaps]
	// Credentials are something different, it's not part of the mod, it's not part of the workspace, it is at the same level
	// with mod and workspace. However, it can be referenced by the mod, so it needs to be in the parse context
	Credentials  map[string]credential.Credential
	Integrations map[string]flowpipe.Integration
	Notifiers    map[string]flowpipe.Notifier
}

// build options used to load workspace
func (w *FlowpipeWorkspace) GetParseContext(ctx context.Context) (*parse.ModParseContext, error) {
	parseCtx, err := w.WorkspaceBase.GetParseContext(ctx)
	if err != nil {
		return nil, err
	}
	// TODO K make an interface for parse context and have a flowpipe implementation
	parseCtx.Credentials = w.Credentials
	parseCtx.Integrations = w.Integrations
	parseCtx.Notifiers = w.Notifiers

	// I don't think we need CredentialImports here .. it's fully resolved to credentials at startup

	return parseCtx, nil
}
