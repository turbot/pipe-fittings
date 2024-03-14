package flowpipeconfig

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/inputvars"
	"github.com/turbot/terraform-components/tfdiags"
)

func sanitiseTriggerNames(src []byte) ([]byte, map[string]string) {
	// replace syntax `<modname>.<varname>=<var_value>` with `____flowpipe_mod_<modname>_<varname>____=<var_value>

	lines := strings.Split(string(src), "\n")
	// make map of varname aliases
	var depVarAliases = make(map[string]string)

	pattern := `^\s*([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)\s*=\s*(true|false|"[^"]*")\s*$`
	re := regexp.MustCompile(pattern)

	for _, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) == 3 {
				fullKey := match[1]
				value := match[2]

				fmt.Println(value)

				parts := strings.Split(fullKey, ".")
				if len(parts) == 3 {
					// root mod for example:
					// schedule.my_scheduled_trigger.enabled = true
				} else if len(parts) == 5 {
					// dependent mod for example:
					// my_mod.trigger.http.http_trigger.enabled = true
				}
			}
		}
	}

	// now try again
	src = []byte(strings.Join(lines, "\n"))
	return src, depVarAliases
}

func LoadFlowpipeTriggerFile(filename string) (map[string]inputvars.UnparsedVariableValueExpression, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	src, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read Flowpipe trigger configuration file",
				fmt.Sprintf("Given Flowpipe trigger configuration file %s does not exist.", filename),
			))
		} else {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read Flowpipe trigger configuration file",
				fmt.Sprintf("Error while reading %s: %s.", filename, err),
			))
		}
		return nil, diags
	}
	sanitisedSrc, depVarAliases := sanitiseTriggerNames(src)

	fmt.Printf("sanitisedSrc: %s\n", sanitisedSrc)
	fmt.Printf("depVarAliases: %s\n", depVarAliases)

	var f *hcl.File
	var hclDiags hcl.Diagnostics

	// attempt to parse the config
	f, hclDiags = hclsyntax.ParseConfig(sanitisedSrc, filename, hcl.Pos{Line: 1, Column: 1})
	diags = diags.Append(hclDiags)
	if f == nil || f.Body == nil {
		return nil, diags
	}

	if len(diags) > 0 {
		return nil, diags
	}

	attrs, hclDiags := f.Body.JustAttributes()
	diags = diags.Append(hclDiags)

	results := make(map[string]inputvars.UnparsedVariableValueExpression, 0)
	for name, attr := range attrs {
		// check for aliases
		if alias, ok := depVarAliases[name]; ok {
			name = alias
		}
		results[name] = inputvars.UnparsedVariableValueExpression{
			Expr: attr.Expr,
		}
	}

	return results, diags
}
