package steampipeconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/terraform-components/tfdiags"
)

type MissingVariableError struct {
	MissingVariables           []*modconfig.Variable
	MissingTransitiveVariables map[DependencyPathKey][]*modconfig.Variable
	workspaceMod               *modconfig.Mod
}

func NewMissingVarsError(workspaceMod *modconfig.Mod) MissingVariableError {
	return MissingVariableError{
		MissingTransitiveVariables: make(map[DependencyPathKey][]*modconfig.Variable),
		workspaceMod:               workspaceMod,
	}
}

func (m MissingVariableError) Error() string {
	//allMissing := append(m.MissingVariables, m.MissingTransitiveVariables...)
	missingCount := len(m.MissingVariables)
	for _, missing := range m.MissingTransitiveVariables {
		missingCount += len(missing)
	}

	return fmt.Sprintf("missing %d variable %s:\n%s%s",
		missingCount,
		utils.Pluralize("value", missingCount),
		m.getVariableMissingString(),
		m.getTransitiveVariableMissingString(),
	)
}

func (m MissingVariableError) getVariableMissingString() string {
	var sb strings.Builder

	varNames := make([]string, len(m.MissingVariables))
	for i, v := range m.MissingVariables {
		varNames[i] = m.getVariableName(v)
	}

	// sort names for top level first
	sort.Slice(varNames, func(i, j int) bool {
		if len(strings.Split(varNames[i], ".")) < len(strings.Split(varNames[j], ".")) {
			return true
		} else {
			return false
		}
	})

	for _, v := range varNames {
		sb.WriteString(fmt.Sprintf("\t%s not set\n", v))
	}
	return sb.String()
}

func (m MissingVariableError) getTransitiveVariableMissingString() string {
	var sb strings.Builder
	for modPath, missingVars := range m.MissingTransitiveVariables {
		parentPath := modPath.GetParent()
		varCount := len(missingVars)

		varNames := make([]string, len(missingVars))
		for i, v := range missingVars {
			varNames[i] = m.getVariableName(v)
		}

		pluralizer := pluralize.NewClient()
		pluralizer.AddIrregularRule("has", "have")
		pluralizer.AddIrregularRule("an arg", "args")
		varsString := strings.Join(varNames, ",")

		sb.WriteString(
			fmt.Sprintf("\tdependency mod %s cannot be loaded because %s %s %s no value.  Mod %s must pass %s for %s in the `require` block of its %s\n",
				modPath,
				pluralizer.Pluralize("variable", varCount, false),
				varsString,
				pluralizer.Pluralize("has", varCount, false),
				parentPath,
				pluralizer.Pluralize("a value", varCount, false),
				varsString,
				app_specific.DefaultModFileName(),
			))

	}
	return sb.String()
}

func (m MissingVariableError) getVariableName(v *modconfig.Variable) string {
	if v.Mod.Name() == m.workspaceMod.Name() {
		return v.ShortName
	}
	return fmt.Sprintf("%s.%s", v.Mod.ShortName, v.ShortName)
}

type VariableValidationFailedError struct {
	diags tfdiags.Diagnostics
}

func NewVariableValidationFailedError(diags tfdiags.Diagnostics) VariableValidationFailedError {
	return VariableValidationFailedError{diags: diags}
}
func (m VariableValidationFailedError) Error() string {
	var sb strings.Builder

	for i, diag := range m.diags {

		if diag.Severity() == tfdiags.Error {
			detailedErrorString := fmt.Sprintf("%s: %s",
				diag.Description().Summary,
				diag.Description().Detail)

			if diag.Source().Subject != nil {
				detailedErrorString = fmt.Sprintf("%s\n(%s)", detailedErrorString, diag.Source().Subject.StartString())
			}

			sb.WriteString(detailedErrorString)
			if i < len(m.diags)-1 {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}
