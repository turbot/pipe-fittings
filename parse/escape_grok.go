package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"slices"
	"strconv"
	"strings"
)

// used to give warning that grok expressions should be wrapped in a 'grok' function call
var grokConfigProperties = []string{"log_format", "file_layout", "layout"}

// GrokEscape is the implementation of the hcl 'grok' function
// which is used to wrap grok expressions and escape the Grok pattern so the hcl parse does not fail
// NOTE: we must implement this explicitly as part of the parse rather than a standard context function as
// a grok pattern containing the string "%{" will cause the initial hcl parse to fail
func GrokEscape(f []byte, filePath string) ([]byte, hcl.Diagnostics) {
	// clone fileData
	fileData := make([]byte, len(f))
	copy(fileData, f)

	for {
		// because the parse will return errors for a single attribute at a time, we may need to call doEscapeGrokArgs
		// multiple times

		updatedFileData, diags := doEscapeGrokArgs(fileData, filePath)
		if diags.HasErrors() {
			return fileData, diags
		}

		if string(updatedFileData) == string(fileData) {
			return updatedFileData, nil
		}
		fileData = updatedFileData
	}
}

func doEscapeGrokArgs(fileData []byte, filePath string) ([]byte, hcl.Diagnostics) {
	// Parse HCL file without caching
	file, diags := hclsyntax.ParseConfig(fileData, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})

	// Return original data if no errors or failed parsing
	if !diags.HasErrors() || file == nil {
		return fileData, nil
	}

	type replacement struct {
		start int
		end   int
		value string
	}

	var replacements []replacement

	// Iterate over diagnostics to find Grok pattern errors
	for _, diag := range diags {
		if isGrokPatternError(diag) {

			a := getAttributeForRange(file.Body.(*hclsyntax.Body), diag.Subject)
			if a == nil {
				continue
			}

			// we only do this escaping if the attribute is a 'grok' function call
			f, ok := a.Expr.(*hclsyntax.FunctionCallExpr)
			if !ok || f.Name != "grok" {
				// so there is a potential grok pattern error, but it is not in a grok function call
				// is this a known grok property?
				if slices.Contains(grokConfigProperties, a.Name) {
					return fileData, hcl.Diagnostics{
						&hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unescaped Grok expression",
							Detail:   fmt.Sprintf("The attribute '%s' in file %q looks like a Grok pattern. This should be wrapped in a 'grok()' function call (with no surrounding quotes).", a.Name, filePath),
							Subject:  diag.Subject,
						}}
				}
			}
			if len(f.Args) == 0 {
				continue
			}

			startByte := a.EqualsRange.End.Byte
			replaceStartByte := f.Args[0].StartRange().Start.Byte
			// The end byte will not be set as the parse of the arg failed, so just take the while line
			// Find the end of the current line
			// Default to end of file
			endByte := len(fileData) - 1
			for i := replaceStartByte + 1; i < len(fileData); i++ {
				if fileData[i] == '\n' {
					endByte = i - 1
					break
				}
			}
			// now search back to the last close bracket
			replaceEndByte := endByte
			for i := endByte; i > replaceStartByte; i-- {
				if fileData[i] == ')' {
					replaceEndByte = i
					break
				}
			}
			// Extract the value to be escaped
			hclVal := fileData[replaceStartByte:replaceEndByte]

			// Efficiently escape "%{" while keeping existing "%%{" unchanged
			escapedAttr := escapeGrokCapturePattern(string(hclVal))

			// Escape the string
			escapedAttr = strconv.Quote(escapedAttr)

			// put in quotes
			//escapedAttr = fmt.Sprintf(`"%s"`, escapedAttr)

			// add to list of replacements
			replacements = append(replacements, replacement{start: startByte, end: endByte, value: escapedAttr})
		}
	}

	for i := len(replacements) - 1; i >= 0; i-- {
		// Perform replacements in reverse order to avoid changing the indices
		replacement := replacements[i]
		fileData = append(fileData[:replacement.start], append([]byte(replacement.value), fileData[replacement.end+1:]...)...)
	}

	return fileData, nil
}

func isGrokPatternError(diag *hcl.Diagnostic) bool {
	return diag.Summary == "Invalid template control keyword" || diag.Detail == "Expected the start of an expression, but found an invalid expression token."
}

// escapeGrokCapturePattern ensures "%{" is escaped as "%%{" but does NOT double-escape existing "%%{"
// escapeGrokCapturePattern escapes "%{" as "%%{" but does NOT double-escape existing "%%{"
func escapeGrokCapturePattern(input string) string {
	var sb strings.Builder
	n := len(input)

	for i := 0; i < n; i++ {
		if input[i] == '%' && i+1 < n && input[i+1] == '{' {
			// If it's already "%%{", keep it as is
			if i > 0 && input[i-1] == '%' {
				sb.WriteString("%{") // Keep it unchanged
			} else {
				sb.WriteString("%%{") // Escape "%{" to "%%{"
			}
			i++ // Skip '{' since we already processed it
		} else {
			sb.WriteByte(input[i])
		}
	}

	return sb.String()
}

func getAttributeForRange(syntaxBody *hclsyntax.Body, subject *hcl.Range) *hclsyntax.Attribute {
	for _, attribute := range syntaxBody.Attributes {
		if attribute.Expr.Range().Start.Byte <= subject.Start.Byte && attribute.Expr.Range().End.Byte >= subject.End.Byte {
			return attribute
		}
	}
	for _, block := range syntaxBody.Blocks {
		attr := getAttributeForRange(block.Body, subject)
		if attr != nil {
			return attr
		}
	}

	return nil

}
