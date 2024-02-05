package flowpipeconfig

import (
	"log/slog"
	"maps"

	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/filepaths"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
)

type loadConfigOptions struct {
	include []string
}

func LoadFlowpipeConfig(configPaths []string) (*FlowpipeConfig, error_helpers.ErrorAndWarnings) {
	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	connectionConfigExtensions := []string{app_specific.ConfigExtension}

	include := filehelpers.InclusionsFromExtensions(connectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}

	var res = NewFlowpipeConfig()
	var credentialMap = res.Credentials
	var integrationMap = res.Integrations

	// load from the config paths in reverse order (i.e. lowest precedence first)
	for i := len(configPaths) - 1; i >= 0; i-- {
		configPath := configPaths[i]

		c, i, ew := loadFlowpipeConfigBlocks(configPath, loadOptions)
		if ew.GetError() != nil {
			return nil, ew
		}
		// merge the warning from this call
		errorsAndWarnings.AddWarning(ew.Warnings...)

		// copy creds over the top of credentialMap (i.e. with greater precedence)
		maps.Copy(credentialMap, c)
		maps.Copy(integrationMap, i)
	}
	if len(credentialMap) > 0 {
		res.Credentials = credentialMap
	}

	if len(integrationMap) > 0 {
		res.Integrations = integrationMap
	}

	return res, errorsAndWarnings
}

func loadFlowpipeConfigBlocks(configPath string, opts *loadConfigOptions) (map[string]credential.Credential, map[string]modconfig.Integration, error_helpers.ErrorAndWarnings) {
	var credentials = map[string]credential.Credential{}
	var integrations = map[string]modconfig.Integration{}

	configPaths, err := filehelpers.ListFiles(configPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
		Exclude: []string{filepaths.WorkspaceLockFileName},
	})

	if err != nil {
		slog.Warn("failed to get config file paths", "error", err)
		return nil, nil, error_helpers.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return nil, nil, error_helpers.ErrorAndWarnings{}
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		slog.Warn("failed to load all config files", "error", err)
		return nil, nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case schema.BlockTypeCredential:
			credential, moreDiags := parse.DecodeCredential(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Warn("failed to decode credential block", "error", err)
				continue
			}

			credentials[credential.GetUnqualifiedName()] = credential
		case schema.BlockTypeIntegration:
			integration, moreDiags := parse.DecodeIntegration(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Warn("failed to decode integration block", "error", err)
				continue
			}

			integrations[integration.GetUnqualifiedName()] = integration
		}
	}

	if len(diags) > 0 {
		return nil, nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load Flowpipe config", diags)
	}
	return credentials, integrations, error_helpers.ErrorAndWarnings{}
}
