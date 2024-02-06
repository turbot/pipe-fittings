package flowpipeconfig

import (
	"log/slog"
	"maps"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/zclconf/go-cty/cty"

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

	lastErrorLength := 0

	for {

		var diags hcl.Diagnostics
		for i := len(configPaths) - 1; i >= 0; i-- {
			configPath := configPaths[i]
			moreDiags := res.loadFlowpipeConfigBlocks(configPath, loadOptions)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
			}
		}

		if len(diags) == 0 {
			break
		}

		if len(diags) > 0 && lastErrorLength == len(diags) {
			return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load Flowpipe config", diags)
		}

		lastErrorLength = len(diags)
	}

	return res, errorsAndWarnings
}

func (f *FlowpipeConfig) loadFlowpipeConfigBlocks(configPath string, opts *loadConfigOptions) hcl.Diagnostics {

	configPaths, err := filehelpers.ListFiles(configPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
		Exclude: []string{filepaths.WorkspaceLockFileName},
	})

	if err != nil {
		slog.Warn("failed to get config file paths", "error", err)
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to get config file paths",
				Detail:   err.Error(),
			},
		}
	}

	if len(configPaths) == 0 {
		return hcl.Diagnostics{}
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		slog.Warn("failed to load all config files", "error", err)
		return diags
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return diags
	}

	// do a partial decode
	content, diags := body.Content(parse.ConfigBlockSchema)
	if diags.HasErrors() {
		return diags
	}

	var credentials = map[string]credential.Credential{}
	var integrations = map[string]modconfig.Integration{}
	var notifiers = map[string]modconfig.Notifier{}

	// Parse credentials and integration first
	for _, block := range content.Blocks {
		switch block.Type {
		case schema.BlockTypeCredential:
			credential, moreDiags := parse.DecodeCredential(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode credential block")
				continue
			}

			credentials[credential.GetUnqualifiedName()] = credential
		case schema.BlockTypeIntegration:
			integration, moreDiags := parse.DecodeIntegration(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode integration block")
				continue
			}

			integrations[integration.GetUnqualifiedName()] = integration

		case schema.BlockTypeNotifier:
			evalContext, moreDiags := buildEvalContextWithIntegrationsOnly(configPath, f.Integrations)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			notifier, moreDiags := parse.DecodeNotifier(configPath, block, evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode notifier block")
				continue
			}

			notifiers[notifier.GetUnqualifiedName()] = *notifier
		}
	}

	// Copy over what's avaiable

	if len(credentials) > 0 {
		maps.Copy(f.Credentials, credentials)
	}

	if len(integrations) > 0 {
		maps.Copy(f.Integrations, integrations)
	}

	if len(notifiers) > 0 {
		maps.Copy(f.Notifiers, notifiers)
	}

	if len(diags) > 0 {
		return diags
	}

	return diags
}

func buildEvalContextWithIntegrationsOnly(configPath string, integrations map[string]modconfig.Integration) (*hcl.EvalContext, hcl.Diagnostics) {

	diags := hcl.Diagnostics{}
	variables := make(map[string]cty.Value)

	slack := make(map[string]cty.Value)
	email := make(map[string]cty.Value)
	webform := make(map[string]cty.Value)

	for k, v := range integrations {
		parts := strings.Split(k, ".")
		if len(parts) != 2 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "invalid integration name",
				Detail:   "integration name must be in the format <type>.<name>",
				Subject:  v.GetDeclRange(),
			})
			continue
		}

		var vars map[string]cty.Value

		switch parts[0] {
		case "slack":
			vars = slack
		case "email":
			vars = email
		case "webform":
			vars = webform
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "invalid integration type",
				Detail:   "integration type must be one of slack, email or webform",
				Subject:  v.GetDeclRange(),
			})
			continue
		}

		ctyVal, err := v.CtyValue()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to convert integration to its cty value",
				Detail:   err.Error(),
				Subject:  v.GetDeclRange(),
			})
		}
		vars[parts[1]] = ctyVal
	}
	if len(diags) > 0 {
		return nil, diags
	}

	integrationVariables := make(map[string]cty.Value)
	if len(slack) > 0 {
		integrationVariables["slack"] = cty.ObjectVal(slack)
	}
	if len(email) > 0 {
		integrationVariables["email"] = cty.ObjectVal(email)
	}
	if len(webform) > 0 {
		integrationVariables["webform"] = cty.ObjectVal(webform)
	}

	variables["integration"] = cty.ObjectVal(integrationVariables)

	return &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: variables,
	}, diags
}
