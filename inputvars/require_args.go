package inputvars

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
)

const ValueFromModFile terraform.ValueSourceType = 'M'

func CollectVariableValuesFromModRequire(m *modconfig.Mod, lock *versionmap.WorkspaceLock) (terraform.InputValues, error) {
	res := make(terraform.InputValues)
	if require := m.Require; require != nil {
		for _, depModConstraint := range require.Mods {
			if args := depModConstraint.Args; args != nil {
				// find the loaded dep mod which satisfies this constraint
				resolvedConstraint := lock.GetMod(depModConstraint.Name, m)
				if resolvedConstraint == nil {
					return nil, fmt.Errorf("dependency mod %s is not loaded", depModConstraint.Name)
				}
				for varName, varVal := range args {
					varFullName := fmt.Sprintf("%s.var.%s", resolvedConstraint.Alias, varName)

					sourceRange := tfdiags.SourceRange{
						Filename: require.DeclRange.Filename,
						Start: tfdiags.SourcePos{
							Line:   require.DeclRange.Start.Line,
							Column: require.DeclRange.Start.Column,
							Byte:   require.DeclRange.Start.Byte,
						},
						End: tfdiags.SourcePos{
							Line:   require.DeclRange.End.Line,
							Column: require.DeclRange.End.Column,
							Byte:   require.DeclRange.End.Byte,
						},
					}

					res[varFullName] = &terraform.InputValue{
						Value:       varVal,
						SourceType:  ValueFromModFile,
						SourceRange: sourceRange,
					}
				}
			}
		}
	}
	return res, nil
}
