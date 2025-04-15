package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"strconv"
	"strings"
)

// ApplyPropertyEscaping escapes properties within backticks, and optionally escaped properties specified by disableTemplateForProperties
func ApplyPropertyEscaping(fileDataMap map[string][]byte, opts ...ParseHclOpt) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var res = make(map[string][]byte, len(fileDataMap))
	config := &ParseHclConfig{}
	for _, opt := range opts {
		opt(config)
	}

	for filePath := range fileDataMap {

		fileData := fileDataMap[filePath]
		var moreDiags hcl.Diagnostics
		if config.escapeBackticks {
			// check backtick surrounded property values - escape the contents
			fileData, moreDiags = escapeBackticks(fileDataMap[filePath], filePath)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
		}
		// handle deprecated disableTemplateForProperties
		fileData, moreDiags = applyDisableTemplateForProperties(fileData, filePath, config.disableTemplateForProperties)
		diags = append(diags, moreDiags...)
		if diags.HasErrors() {
			continue
		}
		res[filePath] = fileData
	}
	return res, diags
}

// escapeBackticks implements hcl backtick escaping
// - any data between backticks will be escaped, including hcl tempate expressions %{ (which are used for grok)
func escapeBackticks(fileData []byte, filePath string) ([]byte, hcl.Diagnostics) {
	for {
		updatedFileData, diags := doEscapeBackticks(fileData, filePath)
		if diags.HasErrors() {
			return fileData, diags
		}

		if string(updatedFileData) == string(fileData) {
			return updatedFileData, nil
		}
		fileData = updatedFileData
	}
}

func doEscapeBackticks(f []byte, filePath string) ([]byte, hcl.Diagnostics) {
	// clone fileData so we can mutate it
	fileData := make([]byte, len(f))
	copy(fileData, f)

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
			escapedAttr = escapeTemplateAndInterpolationPatterns(escapedAttr)
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

// escapeTemplateAndInterpolationPatterns ensures "%{" is escaped as "%%{" and "${" as "$${" in the input string.
// NOTE: we DO NOT escape "%%{" or "$${" as they are already escaped.
func escapeTemplateAndInterpolationPatterns(input string) string {
	var sb strings.Builder
	n := len(input)

	for i := 0; i < n; i++ {
		if input[i] == '%' && i+1 < n && input[i+1] == '{' {
			// escape "%{" to "%%{" but not if it's already "%%{"

			// If it's already "%%{", keep it as is
			if i > 0 && input[i-1] == '%' {
				sb.WriteString("%{") // Keep it unchanged
			} else {
				sb.WriteString("%%{") // Escape "%{" to "%%{"
			}
			i++ // Skip '{' since we already processed it
		} else if input[i] == '$' && i+1 < n && input[i+1] == '{' {
			// escape "${" to "$${" but not if it's already "$${"

			// If it's already "$${", keep it as is
			if i > 0 && input[i-1] == '$' {
				sb.WriteString("${") // Keep it unchanged
			} else {
				sb.WriteString("$${") // Escape "${" to "$${"
			}
			i++ // Skip '{' since we already processed it
		} else {
			// Copy the character as is
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

// applyDisableTemplateForProperties is the deprecated mechanism for escaping template tokens
func applyDisableTemplateForProperties(fileData []byte, filePath string, disableTemplateForProperties []string) ([]byte, hcl.Diagnostics) {
	updatedFileData, diags := EscapeTemplateTokens(fileData, filePath, disableTemplateForProperties)

	// if this modified the file data, it means the grok function is not being used - raise a warning
	if string(updatedFileData) != string(fileData) {
		msg := getEscapeTemplateWarningMessage(filePath, fileData, updatedFileData)

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  msg,
		})
		fileData = updatedFileData
	}
	return fileData, diags
}

// getEscapeTemplateWarningMessage generates a warning message	 for the deprecated disableTemplateForProperties
func getEscapeTemplateWarningMessage(filePath string, fileData, updatedFileData []byte) string {
	differentLines := findDifferentLines(fileData, updatedFileData)
	lineStr := make([]string, 0, len(differentLines))
	for _, line := range differentLines {
		lineStr = append(lineStr, fmt.Sprintf("%d", line))
	}

	return fmt.Sprintf("Config contains reserved reserved characters. These have been auto-escaped for you, but future versions will not do this. \nPlease use backticks to escape the property: file_layout = `${val}`. (%s:%s)\n",
		filePath,
		strings.Join(lineStr, ", "))

}

// findDifferentLines compares two byte slices line by line and returns the line numbers where they differ
func findDifferentLines(original, updated []byte) []int {
	originalLines := strings.Split(string(original), "\n")
	updatedLines := strings.Split(string(updated), "\n")

	var differentLines []int

	// Compare each line
	maxLines := len(originalLines)
	if len(updatedLines) > maxLines {
		maxLines = len(updatedLines)
	}

	for i := 0; i < maxLines; i++ {
		if i >= len(originalLines) || i >= len(updatedLines) {
			differentLines = append(differentLines, i+1)
			continue
		}
		if originalLines[i] != updatedLines[i] {
			differentLines = append(differentLines, i+1)
		}
	}

	return differentLines
}
