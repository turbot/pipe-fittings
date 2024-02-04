package steampipeconfig

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

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
	credentialImportMap := map[string]modconfig.CredentialImport{}

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

		// Credential import
		ci, ew := loadCredentialImport(configPath, loadOptions)
		if ew != nil {
			if ew.GetError() != nil {
				return nil, ew
			}
			// merge the warning from this call
			errorsAndWarnings.AddWarning(ew.Warnings...)
		}
		maps.Copy(credentialImportMap, ci)

	}
	if len(credentialMap) > 0 {
		res.Credentials = credentialMap
	}

	credentials := make(map[string]modconfig.Credential)
	if len(credentialImportMap) > 0 {
		for _, credentialImport := range credentialImportMap {
			if credentialImport.Source != nil {
				filePaths, err := parse.ResolveCredentialImportSource(credentialImport.Source)
				if err != nil {
					// TODO: Build diags
					return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to resolve the source path", hcl.Diagnostics{})
				}

				fileData, diags := parse.LoadFileData(filePaths...)
				if diags.HasErrors() {
					slog.Error("loadConfig: failed to load all config files", "error", err)
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
					if block.Type == schema.BlockTypeConnection {
						connection, moreDiags := parse.DecodeConnection(block)
						diags = append(diags, moreDiags...)
						if moreDiags.HasErrors() {
							continue
						}

						connectionType := connection.PluginAlias
						connectionName := block.Labels[0]

						if credentialImport.Connections != nil && len(credentialImport.Connections) > 0 {
							if !isRequiredConnection(connectionName, credentialImport.Connections) {
								continue
							}
						}

						if credentialImport.Prefix != nil && *credentialImport.Prefix != "" {
							connectionName = fmt.Sprintf("%s%s", *credentialImport.Prefix, connectionName)
						}
						credentialName := fmt.Sprintf("%s.%s", connectionType, connectionName)

						// Return error if the flowpipe already has a creds with same type and name
						if res.Credentials[credentialName] != nil {
							return nil, error_helpers.DiagsToErrorsAndWarnings(fmt.Sprintf("credential already exists '%s'", connectionName), diags)
						}

						// Parse the config string
						configString := []byte(connection.Config)

						// filename and range may not have been passed (for older versions of CLI)
						filename := ""
						startPos := hcl.Pos{}

						body, diags := parseConfig(configString, filename, startPos)
						if diags.HasErrors() {
							return nil, error_helpers.DiagsToErrorsAndWarnings(fmt.Sprintf("failed to parse connection config for connection '%s'", connection.Name), diags)
						}
						evalCtx := &hcl.EvalContext{
							Variables: make(map[string]cty.Value),
							Functions: make(map[string]function.Function),
						}

						configStruct := modconfig.ResolveConfigStruct(connectionType)
						moreDiags = gohcl.DecodeBody(body, evalCtx, configStruct)
						diags = append(diags, moreDiags...)
						if diags.HasErrors() {
							return nil, error_helpers.DiagsToErrorsAndWarnings(fmt.Sprintf("failed to parse connection config for connection '%s'", connection.Name), diags)
						}

						switch v := configStruct.(type) {
						case *modconfig.AbuseIPDBCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.AwsCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.AzureCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.BitbucketCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.ClickUpCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.DatadogCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.DiscordCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.FreshdeskCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.GcpCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.GithubCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.GitLabCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.GuardrailsCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.IP2LocationIOCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.IPstackCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.JiraCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.JumpCloudCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.OktaCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.OpenAICredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.OpsgenieCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.PagerDutyCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.PipesCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.SendGridCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.ServiceNowCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.SlackCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.MicrosoftTeamsCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.TrelloCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.UptimeRobotCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.UrlscanCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.VaultCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.VirusTotalCredential:
							credentials[credentialName] = modconfig.Credential(v)
						case *modconfig.ZendeskCredential:
							credentials[credentialName] = modconfig.Credential(v)
						}
					}
				}
			}
		}
	}

	if len(credentials) > 0 {
		for k, v := range credentials {
			res.Credentials[k] = v
		}
	}

	return res, errorsAndWarnings
}

func loadCredentials(configPath string, opts *loadConfigOptions) (map[string]modconfig.Credential, *error_helpers.ErrorAndWarnings) {
	var res = map[string]modconfig.Credential{}
	configPaths, err := filehelpers.ListFiles(configPath, &filehelpers.ListOptions{
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
			credential, moreDiags := parse.DecodeCredential(configPath, block)
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

func loadCredentialImport(configPath string, opts *loadConfigOptions) (map[string]modconfig.CredentialImport, *error_helpers.ErrorAndWarnings) {
	var res = map[string]modconfig.CredentialImport{}
	configPaths, err := filehelpers.ListFiles(configPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
		Exclude: []string{filepaths.WorkspaceLockFileName},
	})

	if err != nil {
		slog.Warn("loadCredentialImport: failed to get config file paths", "error", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return nil, nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		slog.Warn("loadCredentialImport: failed to load all config files", "error", err)
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

		case schema.BlockTypeCredentialImport:
			credentialImport, moreDiags := parse.DecodeCredentialImport(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Warn("loadCredentialImport: failed to decode credential_import block", "error", err)
				continue
			}

			if credentialImport != nil {
				res[credentialImport.GetUnqualifiedName()] = *credentialImport
			}
		}
	}

	if len(diags) > 0 {
		return nil, error_helpers.DiagsToErrorsAndWarnings("Failed to load Flowpipe config", diags)
	}
	return res, nil
}

func isRequiredConnection(str string, patterns []string) bool {
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, str)
		if err != nil {
			slog.Warn("isRequiredConnection: error matching pattern", "pattern", pattern, "error", err)
			continue
		}

		if match {
			return true
		}
	}
	return false
}
