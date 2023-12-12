package steampipeconfig

import (
	"log/slog"

	"github.com/turbot/pipe-fittings/filepaths"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"golang.org/x/exp/maps"
)

type loadConfigOptions struct {
	include []string
}

func LoadFlowpipeConfig(configPaths []string) (*modconfig.FlowpipeConfig, *error_helpers.ErrorAndWarnings) {
	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	connectionConfigExtensions := []string{app_specific.ConfigExtension}

	include := filehelpers.InclusionsFromExtensions(connectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}

	var res = modconfig.NewFlowpipeConfig()
	var credentialMap = res.Credentials
	// load from the config paths in reverse order (i.e. lowest precedence first)
	for i := len(configPaths) - 1; i >= 0; i-- {
		configPath := configPaths[i]

		c, ew := loadCredentials(configPath, loadOptions)

		if ew != nil {
			if ew.GetError() != nil {
				return nil, ew
			}
			// merge the warning from this call
			errorsAndWarnings.AddWarning(ew.Warnings...)
		}
		// copy creds over the top of credentialMap (i.e. with greater precedence)
		maps.Copy(credentialMap, c)
	}
	if len(credentialMap) > 0 {
		res.Credentials = credentialMap
	}
	return res, errorsAndWarnings
}

func loadCredentials(configFolder string, opts *loadConfigOptions) (map[string]modconfig.Credential, *error_helpers.ErrorAndWarnings) {
	var res = map[string]modconfig.Credential{}
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
		Exclude: []string{filepaths.WorkspaceLockFileName},
	})

	if err != nil {
		slog.Warn("loadCredentials: failed to get config file paths", "error", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return nil, nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		slog.Warn("loadCredentials: failed to load all config files", "error", err)
		return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	for _, block := range content.Blocks {
		switch block.Type {

		case schema.BlockTypeCredential:
			credential, moreDiags := parse.DecodeCredential(block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Warn("loadCredentials: failed to decode credential block", "error", err)
				continue
			}

			res[credential.GetUnqualifiedName()] = credential
		}
	}

	if len(diags) > 0 {
		return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load Flowpipe config", diags)
	}
	return res, nil
}
