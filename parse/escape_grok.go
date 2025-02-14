package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	for i, diag := range diags {
		if isBacktickError(diag) {
			// only replace if the backtick is at the start of an attribute value
			a := getAttributeForRange(file.Body.(*hclsyntax.Body), diag.Subject)
			if a == nil {
				continue
			}

			replaceStartByte := diag.Subject.Start.Byte
			replaceEndByte := replaceStartByte

			for i := replaceStartByte + 1; i < len(fileData); i++ {
				if fileData[i] == '`' {
					replaceEndByte = i
					break
				}
			}
			if replaceEndByte == replaceStartByte {
				// replace the diagnostic
				diags[i].Summary = "Could not find closing backtick"
				diags[i].Detail = "A backtick escape was found but the closing backtick could not be found"
				return nil, diags
			}

			// Extract the value to be escaped
			hclVal := fileData[replaceStartByte : replaceEndByte+1]
			// trim whitespace
			escapedAttr := strings.TrimSpace(string(hclVal))
			// remove the opening and closing `
			escapedAttr = strings.TrimPrefix(escapedAttr, "`")
			escapedAttr = strings.TrimSuffix(escapedAttr, "`")
			// Efficiently escape "%{" while keeping existing "%%{" unchanged
			escapedAttr = escapeGrokCapturePattern(escapedAttr)
			// Add quotes and escape the string
			escapedAttr = strconv.Quote(escapedAttr)

			// put in quotes
			//escapedAttr = fmt.Sprintf(`"%s"`, escapedAttr)

			// add to list of replacements
			replacements = append(replacements, replacement{start: replaceStartByte, end: replaceEndByte, value: escapedAttr})
		}
	}

	for i := len(replacements) - 1; i >= 0; i-- {
		// Perform replacements in reverse order to avoid changing the indices
		replacement := replacements[i]
		fileData = append(fileData[:replacement.start], append([]byte(replacement.value), fileData[replacement.end+1:]...)...)
	}

	return fileData, nil
}

func isBacktickError(diag *hcl.Diagnostic) bool {
	return strings.HasPrefix(diag.Detail, "The \"`\" character is not valid.")
}

// escapeGrokCapturePattern ensures "%{" is escaped as "%%{" but does NOT double-escape existing "%%{"
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
