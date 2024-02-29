package workspace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/inputvars"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/terraform-components/terraform"
)

type LoadWorkspaceOption func(*LoadWorkspaceConfig)

type LoadWorkspaceConfig struct {
	credentials         map[string]credential.Credential
	integrations        map[string]modconfig.Integration
	notifiers           map[string]modconfig.Notifier
	blockTypeInclusions []string
	validateVariables   bool
}

func newLoadWorkspaceConfig() *LoadWorkspaceConfig {
	return &LoadWorkspaceConfig{
		credentials:  make(map[string]credential.Credential),
		integrations: make(map[string]modconfig.Integration),
		notifiers:    make(map[string]modconfig.Notifier),
	}
}

func WithCredentials(credentials map[string]credential.Credential) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.credentials = credentials
	}
}

func WithIntegrations(integrations map[string]modconfig.Integration) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.integrations = integrations
	}
}

func WithNotifiers(notifiers map[string]modconfig.Notifier) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.notifiers = notifiers
	}
}

func WithBlockType(blockTypeInclusions []string) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.blockTypeInclusions = blockTypeInclusions
	}
}
func WithVariableValidation(enabled bool) LoadWorkspaceOption {
	return func(m *LoadWorkspaceConfig) {
		m.validateVariables = enabled
	}
}

func LoadWorkspacePromptingForVariables(ctx context.Context, workspacePath string, opts ...LoadWorkspaceOption) (*Workspace, error_helpers.ErrorAndWarnings) {
	t := time.Now()
	defer func() {
		slog.Debug("Workspace load complete", "duration (ms)", time.Since(t).Milliseconds())
	}()
	w, errAndWarnings := Load(ctx, workspacePath, opts...)
	if errAndWarnings.GetError() == nil {
		return w, errAndWarnings
	}
	var missingVariablesError steampipeconfig.MissingVariableError
	ok := errors.As(errAndWarnings.GetError(), &missingVariablesError)
	// if there was an error which is NOT a MissingVariableError, return it
	if !ok {
		return nil, errAndWarnings
	}
	// if there are missing transitive dependency variables, fail as we do not prompt for these
	if len(missingVariablesError.MissingTransitiveVariables) > 0 {
		return nil, errAndWarnings
	}
	// if interactive input is disabled, return the missing variables error
	if !viper.GetBool(constants.ArgInput) {
		return nil, error_helpers.NewErrorsAndWarning(missingVariablesError)
	}
	// so we have missing variables - prompt for them
	// first hide spinner if it is there
	statushooks.Done(ctx)
	if err := promptForMissingVariables(ctx, missingVariablesError.MissingVariables, workspacePath); err != nil {
		slog.Debug("Interactive variables prompting returned error %v", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}
	// ok we should have all variables now - reload workspace
	return Load(ctx, workspacePath, opts...)
}

func promptForMissingVariables(ctx context.Context, missingVariables []*modconfig.Variable, workspacePath string) error {
	fmt.Println()                                       //nolint:forbidigo // UI formatting
	fmt.Println("Variables defined with no value set.") //nolint:forbidigo // UI formatting
	for _, v := range missingVariables {
		variableName := v.ShortName
		variableDisplayName := fmt.Sprintf("var.%s", v.ShortName)
		// if this variable is NOT part of the workspace mod, add the mod name to the variable name
		if v.Mod.ModPath != workspacePath {
			variableDisplayName = fmt.Sprintf("%s.var.%s", v.ModName, v.ShortName)
			variableName = fmt.Sprintf("%s.%s", v.ModName, v.ShortName)
		}
		r, err := promptForVariable(ctx, variableDisplayName, v.GetDescription())
		if err != nil {
			return err
		}
		addInteractiveVariableToViper(variableName, r)
	}
	return nil
}

func promptForVariable(ctx context.Context, name, description string) (string, error) {
	uiInput := &inputvars.UIInput{}
	rawValue, err := uiInput.Input(ctx, &terraform.InputOpts{
		Id:          name,
		Query:       name,
		Description: description,
	})

	return rawValue, err
}

func addInteractiveVariableToViper(name string, rawValue string) {
	varMap := viper.GetStringMap(constants.ConfigInteractiveVariables)
	varMap[name] = rawValue
	viper.Set(constants.ConfigInteractiveVariables, varMap)
}
