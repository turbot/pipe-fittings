package workspace

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"log/slog"
	"time"

	"github.com/turbot/pipe-fittings/error_helpers"
)

func LoadWorkspacePromptingForVariables(ctx context.Context, workspacePath string, opts ...LoadWorkspaceOption) (*Workspace, error_helpers.ErrorAndWarnings) {
	// do not load resources if there is no modfile
	opts = append(opts, WithSkipResourceLoadIfNoModfile(true))

	t := time.Now()
	defer func() {
		slog.Debug("Workspace load complete", "duration (ms)", time.Since(t).Milliseconds())
	}()
	w, errAndWarnings := Load(ctx, workspacePath, opts...)
	if errAndWarnings.GetError() == nil {
		return w, errAndWarnings
	}

	// kif there wqs an error check if it was a missing variable error and if so prompt for variables
	if err := HandleWorkspaceLoadError(ctx, errAndWarnings.GetError(), workspacePath); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// ok we should have all variables now - reload workspace
	return Load(ctx, workspacePath, opts...)
}

// Load_ creates a Workspace and loads the workspace mod

func Load(ctx context.Context, workspacePath string, opts ...LoadWorkspaceOption) (w *Workspace, ew error_helpers.ErrorAndWarnings) {
	cfg := newLoadFlowpipeWorkspaceConfig()
	for _, o := range opts {
		o(cfg)
	}

	utils.LogTime("w.Load start")
	defer utils.LogTime("w.Load end")

	w = &Workspace{
		Path:              workspacePath,
		VariableValues:    make(map[string]string),
		ValidateVariables: true,
		Mod:               modconfig.NewMod("local", workspacePath, hcl.Range{}),
	}

	// check whether the workspace contains a modfile
	// this will determine whether we load files recursively, and create pseudo resources for sql files
	w.SetModfileExists()

	// load the .steampipe ignore file
	if err := w.LoadExclusions(); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	w.PipelingConnections = cfg.pipelingConnections
	w.SupportLateBinding = cfg.supportLateBinding
	w.BlockTypeInclusions = cfg.blockTypeInclusions
	w.ValidateVariables = cfg.validateVariables

	w.configValueMaps = cfg.configValueMaps
	w.decoderOptions = cfg.decoderOptions

	// if there is a mod file (or if we are loading resources even with no modfile), load them
	if w.ModfileExists() || !cfg.skipResourceLoadIfNoModfile {
		ew = w.LoadWorkspaceMod(ctx)
	}
	return
}
