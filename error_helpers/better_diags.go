package error_helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/perr"
)

var credentialTypeRegistry = make(map[string]struct{}, 0)

func RegisterCredentialType(credentialType string) {
	credentialTypeRegistry[credentialType] = struct{}{}
}

// This is not the best guess, but it's sufficient for now. We can improve this later to add more context or
// making sure that the error is relevant. The plumbing is at least here.
func BetterHclDiagsToError(prefix string, diags hcl.Diagnostics) error {
	if !diags.HasErrors() {
		return nil
	}
	errStrings := betterDiagsToString(diags, hcl.DiagError)

	var res string
	if len(errStrings) > 0 {
		res = strings.Join(errStrings, "\n")
		if len(errStrings) > 1 {
			res += "\n"
		}
		return perr.InternalWithMessage(fmt.Sprintf("%s: %s", prefix, res))
	}

	return diags.Errs()[0]
}

func betterDiagsToString(diags hcl.Diagnostics, severity hcl.DiagnosticSeverity) (strs []string) { // convert the first diag into an error

	// store list of messages (without the range) and use for de-duping (we may get the same message for multiple ranges)
	var msgMap = make(map[string]struct{})
	for _, diag := range diags {

		// check if diag has "connection or credential" expression on it
		if diag.Expression != nil {
			if scopeTraverals, ok := diag.Expression.(*hclsyntax.ScopeTraversalExpr); ok {
				if len(scopeTraverals.Traversal) > 0 {
					if traverserRoot, ok := scopeTraverals.Traversal[0].(hcl.TraverseRoot); ok {
						if traverserRoot.Name == "connection" || traverserRoot.Name == "credential" {
							str := "Missing credential: " + diag.Detail + " Ensures that credentials are correctly defined in the configuration file."
							if _, ok := msgMap[str]; !ok {
								msgMap[str] = struct{}{}
								// now add in the subject and add to the output array
								if diag.Subject != nil && len(diag.Subject.Filename) > 0 {
									str += fmt.Sprintf("\n(%s)", diag.Subject.String())
								}

								strs = append(strs, str)
							}
							continue
						}
					}
				}
			}
		}

		if diag.Severity == severity {
			str := matchSpecificIssue(diag.Summary, diag.Detail)

			if _, ok := msgMap[str]; !ok {
				msgMap[str] = struct{}{}
				// now add in the subject and add to the output array
				if diag.Subject != nil && len(diag.Subject.Filename) > 0 {
					str += fmt.Sprintf("\n(%s)", diag.Subject.String())
				}

				strs = append(strs, str)
			}
		}
	}

	return strs
}

func matchSpecificIssue(summary, detail string) string {

	if detail != "" {
		credTypeAll := ""
		for k := range credentialTypeRegistry {
			credTypeAll += k + "|"
		}
		if len(credTypeAll) > 0 {
			credTypeAll = credTypeAll[:len(credTypeAll)-1]
		}

		pattern := `This object does not have an attribute named "(` + credTypeAll + `)"\.`
		matched, err := regexp.MatchString(pattern, detail)
		if err != nil {
			return detail
		}
		if matched {
			newErrorMessage := "Missing credential: " + detail + " Ensures that credentials are correctly defined in the configuration file."
			return newErrorMessage
		}
		return summary + ": " + detail
	}

	return summary
}
