package flowpipeconfig

import (
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

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

func LoadFlowpipeConfig(configPaths []string) (config *FlowpipeConfig, errorsAndWarnings error_helpers.ErrorAndWarnings) {
	errorsAndWarnings = error_helpers.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	connectionConfigExtensions := []string{app_specific.ConfigExtension}

	include := filehelpers.InclusionsFromExtensions(connectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}

	config = NewFlowpipeConfig(configPaths)

	lastErrorLength := 0

	for {

		var diags hcl.Diagnostics
		for i := len(configPaths) - 1; i >= 0; i-- {
			configPath := configPaths[i]
			moreDiags := config.loadFlowpipeConfigBlocks(configPath, loadOptions)
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

	if errorsAndWarnings.Error != nil {
		return config, errorsAndWarnings
	}

	err := config.importCredentials()
	if err != nil {
		slog.Error("failed to import credentials", "error", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	err := config.credentialsToConnection()
	if err != nil {
		slog.Error("failed to convert credentials to connections", "error", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	//err := config.importConnections()
	//if err != nil {
	//	slog.Error("failed to import connections", "error", err)
	//	return nil, error_helpers.NewErrorsAndWarning(err)
	//}

	return config, errorsAndWarnings
}

func (f *FlowpipeConfig) importCredentials() error {
	if len(f.CredentialImports) == 0 {
		return nil
	}

	credentials := map[string]credential.Credential{}
	for _, credentialImport := range f.CredentialImports {
		if credentialImport.Source == nil {
			continue
		}

		// This can't be encapsulated in CredentialImports due to crucial function the `parse` package
		// it will result in circular dependency
		filePaths, err := parse.ResolveCredentialImportSource(credentialImport.Source)
		if err != nil {
			return err
		}

		fileData, diags := parse.LoadFileData(filePaths...)
		if diags.HasErrors() {
			slog.Error("loadConfig: failed to load all config files", "error", err)
			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
		}

		body, diags := parse.ParseHclFiles(fileData)
		if diags.HasErrors() {
			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
		}

		// do a partial decode
		content, moreDiags := body.Content(parse.ConfigBlockSchema)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
		}

		for _, block := range content.Blocks {
			if block.Type == schema.BlockTypeConnection {
				connection, moreDiags := parse.DecodeConnection(block)
				diags = append(diags, moreDiags...)
				if moreDiags.HasErrors() {
					continue
				}

				// If the plugin name contains slash('/'), takes the last part of the name
				connectionType := connection.PluginAlias
				if strings.Contains(connectionType, "/") {
					strParts := strings.Split(connectionType, "/")
					connectionType = strParts[len(strParts)-1]
				}
				connectionName := block.Labels[0]

				if len(credentialImport.Connections) > 0 {
					if !isRequiredConnection(connectionName, credentialImport.Connections) {
						continue
					}
				}

				if credentialImport.Prefix != nil && *credentialImport.Prefix != "" {
					connectionName = fmt.Sprintf("%s%s", *credentialImport.Prefix, connectionName)
				}
				credentialShortName := connectionName
				credentialFullName := fmt.Sprintf("%s.%s", connectionType, connectionName)

				// Return error if the flowpipe already has a creds with same type and name
				if f.Credentials[credentialFullName] != nil {
					return perr.BadRequestWithMessage(fmt.Sprintf("Credential with name '%s' already exists", credentialFullName))
				}

				if credentials[credentialFullName] != nil {
					return perr.BadRequestWithMessage(fmt.Sprintf("Credential with name '%s' already exists", credentialFullName))
				}

				// Parse the config string
				configString := []byte(connection.Config)

				// filename and range may not have been passed (for older versions of CLI)
				filename := ""
				startPos := hcl.Pos{}

				body, diags := parse.ParseConfig(configString, filename, startPos)
				if diags.HasErrors() {
					return error_helpers.HclDiagsToError("Flowpipe Config", diags)
				}
				evalCtx := &hcl.EvalContext{
					Variables: make(map[string]cty.Value),
					Functions: make(map[string]function.Function),
				}

				configStruct, err := credential.InstantiateCredentialConfig(connectionType)
				if err != nil {
					return err
				}

				// configStruct will be nil if the credential type is not supported by the Flowpipe.
				// In that case, skip the connection
				if configStruct == nil {
					continue
				}

				moreDiags = gohcl.DecodeBody(body, evalCtx, configStruct)
				diags = append(diags, moreDiags...)
				if diags.HasErrors() {
					return error_helpers.HclDiagsToError("Flowpipe Config", diags)
				}

				cred := configStruct.GetCredential(credentialFullName, credentialShortName)
				if cred == nil {
					return perr.InternalWithMessage("Failed to get credential")
				}

				credentials[credentialFullName] = cred
			}
		}
	}

	maps.Copy(f.Credentials, credentials)
	return nil
}

//
//type ImportTarget interface {
//	ImportTarget()
//}

//func importSteampipeConnections[I ImportSteampipeConnection, T modconfig.ConnectionImportTarget](imports []I) error {
//	if len(imports) == 0 {
//		return nil
//	}
//
//	credentials := map[string]T{}
//	for _, connectionImport := range imports {
//		if connectionImport.GetSource() == nil {
//			continue
//		}
//
//		// This can't be encapsulated in CredentialImports due to crucial function the `parse` package
//		// it will result in circular dependency
//		filePaths, err := parse.ResolveCredentialImportSource(connectionImport.GetSource())
//		if err != nil {
//			return err
//		}
//
//		fileData, diags := parse.LoadFileData(filePaths...)
//		if diags.HasErrors() {
//			slog.Error("loadConfig: failed to load all config files", "error", err)
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		body, diags := parse.ParseHclFiles(fileData)
//		if diags.HasErrors() {
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		// do a partial decode
//		content, moreDiags := body.Content(parse.ConfigBlockSchema)
//		if moreDiags.HasErrors() {
//			diags = append(diags, moreDiags...)
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		connections := connectionImport.GetConnections()
//		prefix := connectionImport.GetPrefix()
//
//		for _, block := range content.Blocks {
//			if block.Type == schema.BlockTypeConnection {
//				connection, moreDiags := parse.DecodeConnection(block)
//				diags = append(diags, moreDiags...)
//				if moreDiags.HasErrors() {
//					continue
//				}
//
//				// If the plugin name contains slash('/'), takes the last part of the name
//				connectionType := connection.PluginAlias
//				if strings.Contains(connectionType, "/") {
//					strParts := strings.Split(connectionType, "/")
//					connectionType = strParts[len(strParts)-1]
//				}
//				connectionName := block.Labels[0]
//
//				if len(connections) > 0 {
//					if !isRequiredConnection(connectionName, connections) {
//						continue
//					}
//				}
//
//				if prefix != nil && *prefix != "" {
//					connectionName = fmt.Sprintf("%s%s", *prefix, connectionName)
//				}
//				credentialShortName := connectionName
//				credentialFullName := fmt.Sprintf("%s.%s", connectionType, connectionName)
//
//				// Return error if the flowpipe already has a creds with same type and name
//				if f.Credentials[credentialFullName] != nil {
//					return perr.BadRequestWithMessage(fmt.Sprintf("Credential with name '%s' already exists", credentialFullName))
//				}
//
//				if credentials[credentialFullName] != nil {
//					return perr.BadRequestWithMessage(fmt.Sprintf("Credential with name '%s' already exists", credentialFullName))
//				}
//
//				// Parse the config string
//				configString := []byte(connection.Config)
//
//				// filename and range may not have been passed (for older versions of CLI)
//				filename := ""
//				startPos := hcl.Pos{}
//
//				body, diags := parse.ParseConfig(configString, filename, startPos)
//				if diags.HasErrors() {
//					return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//				}
//				evalCtx := &hcl.EvalContext{
//					Variables: make(map[string]cty.Value),
//					Functions: make(map[string]function.Function),
//				}
//
//				cred, err := steampipeConnectionToCredential(connectionType, body, evalCtx, credentialFullName, credentialShortName)
//				if err != nil {
//					return err
//				}
//
//				credentials[credentialFullName] = cred
//			}
//		}
//	}
//
//	maps.Copy(f.Credentials, credentials)
//	return nil
//}

//func steampipeConnectionToCredential(connectionType string, body hcl.Body, evalCtx *hcl.EvalContext, credentialFullName string, credentialShortName string) (credential.Credential, error) {
//	configStruct, err := credential.InstantiateCredentialConfig(connectionType)
//	if err != nil {
//		return nil, err
//	}
//
//	// configStruct will be nil if the credential type is not supported by the Flowpipe.
//	// In that case, skip the connection
//	if configStruct == nil {
//		//continue
//	}
//
//	diags := gohcl.DecodeBody(body, evalCtx, configStruct)
//	if diags.HasErrors() {
//		return nil, error_helpers.HclDiagsToError("Flowpipe Config", diags)
//	}
//
//	cred := configStruct.GetCredential(credentialFullName, credentialShortName)
//	if cred == nil {
//		return nil, perr.InternalWithMessage("Failed to get credential")
//	}
//	return cred, nil
//}

//func (f *FlowpipeConfig) importConnections() error {
//	if len(f.ConnectionImports) == 0 {
//		return nil
//	}
//
//	connections := map[string]connection.PipelingConnection{}
//	for _, connectionImport := range f.ConnectionImports {
//		if connectionImport.Source == nil {
//			continue
//		}
//
//		// This can't be encapsulated in ConnectionImports due to crucial function the `parse` package
//		// it will result in circular dependency
//		filePaths, err := parse.ResolveConnectionImportSource(connectionImport.Source)
//		if err != nil {
//			return err
//		}
//
//		fileData, diags := parse.LoadFileData(filePaths...)
//		if diags.HasErrors() {
//			slog.Error("loadConfig: failed to load all config files", "error", err)
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		body, diags := parse.ParseHclFiles(fileData)
//		if diags.HasErrors() {
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		// do a partial decode
//		content, moreDiags := body.Content(parse.ConfigBlockSchema)
//		if moreDiags.HasErrors() {
//			diags = append(diags, moreDiags...)
//			return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//		}
//
//		for _, block := range content.Blocks {
//			if block.Type == schema.BlockTypeConnection {
//				connection, moreDiags := parse.DecodeConnection(block)
//				diags = append(diags, moreDiags...)
//				if moreDiags.HasErrors() {
//					continue
//				}
//
//				// If the plugin name contains slash('/'), takes the last part of the name
//				connectionType := connection.PluginAlias
//				if strings.Contains(connectionType, "/") {
//					strParts := strings.Split(connectionType, "/")
//					connectionType = strParts[len(strParts)-1]
//				}
//				connectionName := block.Labels[0]
//
//				if len(connectionImport.Connections) > 0 {
//					if !isRequiredConnection(connectionName, connectionImport.Connections) {
//						continue
//					}
//				}
//
//				if connectionImport.Prefix != nil && *connectionImport.Prefix != "" {
//					connectionName = fmt.Sprintf("%s%s", *connectionImport.Prefix, connectionName)
//				}
//				connectionShortName := connectionName
//				connectionFullName := fmt.Sprintf("%s.%s", connectionType, connectionName)
//
//				// if the flowpipe already has a connection with same type and name, do not import
//				if f.PipelingConnections[connectionFullName] != nil {
//					continue
//				}
//				// ensure this block name has not already been imported
//				if connections[connectionFullName] != nil {
//					return perr.BadRequestWithMessage(fmt.Sprintf("Connection with name '%s' already exists", connectionFullName))
//				}
//
//				// Parse the config string
//				configString := []byte(connection.Config)
//
//				// filename and range may not have been passed (for older versions of CLI)
//				filename := ""
//				startPos := hcl.Pos{}
//
//				body, diags := connection.ParseConfig(configString, filename, startPos)
//				if diags.HasErrors() {
//					return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//				}
//				evalCtx := &hcl.EvalContext{
//					Variables: make(map[string]cty.Value),
//					Functions: make(map[string]function.Function),
//				}
//
//				configStruct, err := connection.InstantiateConnectionConfig(connectionType)
//				if err != nil {
//					return err
//				}
//
//				// configStruct will be nil if the connection type is not supported by the Flowpipe.
//				// In that case, skip the connection
//				if configStruct == nil {
//					continue
//				}
//
//				moreDiags = gohcl.DecodeBody(body, evalCtx, configStruct)
//				diags = append(diags, moreDiags...)
//				if diags.HasErrors() {
//					return error_helpers.HclDiagsToError("Flowpipe Config", diags)
//				}
//
//				cred := configStruct.GetConnection(connectionFullName, connectionShortName)
//				if cred == nil {
//					return perr.InternalWithMessage("Failed to get connection")
//				}
//
//				connections[connectionFullName] = cred
//			}
//		}
//	}
//
//	maps.Copy(f.Connections, connections)
//	return nil
//}

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
	content, diags := body.Content(parse.FlowpipeConfigBlockSchema)
	if diags.HasErrors() {
		return diags
	}

	// Parse credentials and integration first
	for _, block := range content.Blocks {
		switch block.Type {
		case schema.BlockTypeCredentialImport:
			credentialImport, moreDiags := parse.DecodeCredentialImport(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode credential import block")
				continue
			}

			f.CredentialImports[credentialImport.GetUnqualifiedName()] = *credentialImport

		case schema.BlockTypeCredential:
			credential, moreDiags := parse.DecodeCredential(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode credential block")
				continue
			}

			f.Credentials[credential.GetUnqualifiedName()] = credential

		case schema.BlockTypeIntegration:
			integration, moreDiags := parse.DecodeIntegration(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode integration block")
				continue
			}

			f.Integrations[integration.GetUnqualifiedName()] = integration

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

			f.Notifiers[notifier.GetUnqualifiedName()] = notifier
		case schema.BlockTypeConnection:

			conn, moreDiags := parse.DecodePipelingConnection(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode connection block")
				continue
			}

			f.PipelingConnections[conn.Name()] = conn
		case schema.BlockTypeConnectionImport:
			connectionImport, moreDiags := parse.DecodeConnectionImport(configPath, block)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				slog.Debug("failed to decode connection import block")
				continue
			}

			f.ConnectionImports[connectionImport.GetUnqualifiedName()] = *connectionImport
		}
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
	http := make(map[string]cty.Value)
	teams := make(map[string]cty.Value)

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
		case schema.IntegrationTypeSlack:
			vars = slack
		case schema.IntegrationTypeEmail:
			vars = email
		case schema.IntegrationTypeHttp:
			vars = http
		case schema.IntegrationTypeMsTeams:
			vars = teams
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "invalid integration type",
				Detail:   "integration type must be one of slack, email, msteams or http",
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
		integrationVariables[schema.IntegrationTypeSlack] = cty.ObjectVal(slack)
	}
	if len(email) > 0 {
		integrationVariables[schema.IntegrationTypeEmail] = cty.ObjectVal(email)
	}
	if len(http) > 0 {
		integrationVariables[schema.IntegrationTypeHttp] = cty.ObjectVal(http)
	}
	if len(teams) > 0 {
		integrationVariables[schema.IntegrationTypeMsTeams] = cty.ObjectVal(teams)
	}

	variables["integration"] = cty.ObjectVal(integrationVariables)

	return &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: variables,
	}, diags
}
