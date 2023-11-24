package steampipeconfig

import (
	"log"
	"path/filepath"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"

	filehelpers "github.com/turbot/go-kit/files"
)

type loadConfigOptions struct {
	include []string
}

func LoadFlowpipeConfig(modLocation string) (*modconfig.FlowpipeConfig, *error_helpers.ErrorAndWarnings) {

	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	//nolint:gocritic // appendAssign: append result not assigned to the same slice - this is fine could be temporary as well
	connectionConfigExtensions := append(constants.YamlExtensions, app_specific.ConfigExtension, constants.JsonExtension)

	include := filehelpers.InclusionsFromExtensions(connectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}

	flowpipeConfig := modconfig.NewFlowpipeConfig()

	ew := loadConfig(filepaths.EnsureConfigDir(), flowpipeConfig, loadOptions)

	if ew != nil {
		if ew.GetError() != nil {
			return nil, ew
		}
		// merge the warning from this call
		errorsAndWarnings.AddWarning(ew.Warnings...)
	}

	if modLocation != "" {
		ew := loadConfig(filepath.Join(modLocation, ".flowpipe/config"), flowpipeConfig, loadOptions)
		if ew != nil {
			if ew.GetError() != nil {
				return nil, ew
			}
			// merge the warning from this call
			errorsAndWarnings.AddWarning(ew.Warnings...)
		}
	}

	return flowpipeConfig, errorsAndWarnings
}

func loadConfig(configFolder string, flowpipeConfig *modconfig.FlowpipeConfig, opts *loadConfigOptions) *error_helpers.ErrorAndWarnings {

	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return error_helpers.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return error_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return error_helpers.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	for _, block := range content.Blocks {
		switch block.Type {

		case schema.BlockTypeCredential:
			credential, moreDiags := parse.DecodeCredential(block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				log.Printf("[WARN] loadConfig: failed to decode credential block: %v\n", err)
			}

			flowpipeConfig.Credentials[credential.GetUnqualifiedName()] = credential
		}
	}

	if len(diags) > 0 {
		return error_helpers.DiagsToErrorsAndWarnings("Failed to load Flowpipe config", diags)
	}
	return nil
}
