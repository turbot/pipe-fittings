package workspace

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/inputvars"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/statushooks"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/terraform-components/terraform"
	"log/slog"
)

func HandleWorkspaceLoadError(ctx context.Context, err error, workspacePath string) error {
	var missingVariablesError steampipeconfig.MissingVariableError
	ok := errors.As(err, &missingVariablesError)
	// if there was an error which is NOT a MissingVariableError, return it
	if !ok {
		return err
	}
	// if there are missing transitive dependency variables, fail as we do not prompt for these
	if len(missingVariablesError.MissingTransitiveVariables) > 0 {
		return err
	}
	// if interactive input is disabled, return the missing variables error
	if !viper.GetBool(constants.ArgInput) {
		return missingVariablesError
	}
	// so we have missing variables - prompt for them
	// first hide spinner if it is there
	statushooks.Done(ctx)
	if err := promptForMissingVariables(ctx, missingVariablesError.MissingVariables, workspacePath); err != nil {
		slog.Debug("Interactive variables prompting returned error", "error", err)
		return err
	}
	return nil
}

func promptForMissingVariables(ctx context.Context, missingVariables []*modconfig.Variable, workspacePath string) error {
	fmt.Println()                                       //nolint:forbidigo // UI formatting
	fmt.Println("Variables defined with no value set.") //nolint:forbidigo // UI formatting
	for _, v := range missingVariables {
		variableName := v.ShortName
		variableDisplayName := fmt.Sprintf("var.%s", v.ShortName)
		// if this variable is NOT part of the workspace mod, add the mod name to the variable name
		if v.Mod.GetModPath() != workspacePath {
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
