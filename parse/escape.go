package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"regexp"
	"strings"
)

// EscapeTemplateTokens escapes template expressions in any properties specified by disableTemplateForProperties
func EscapeTemplateTokens(fileData []byte, filePath string, disableTemplateForProperties []string) ([]byte, hcl.Diagnostics) {
	// so we need to escape template expressions in the file for the specified properties
	// the initial parse may have failed, but we can still get the ranges of the properties

	// the file will have been at least partially parsed - get the ranges of all attributes
	// for which we are disabling templates

	// do an initial parse of the file
	// NOTE: we do not want to use the parse as we do not want to cache the result
	file, diags := hclsyntax.ParseConfig(fileData, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})

	var attrRanges []*hcl.Range
	for _, attrName := range disableTemplateForProperties {
		moreRanges, moreDiags := getAttributeRanges(file, attrName)
		if moreDiags.HasErrors() {
			return nil, diags
		}
		attrRanges = append(attrRanges, moreRanges...)
	}

	// now for each attribute we found, escape template expression opening '%{' to '%%{'
	// and reparse the file

	// if we did not find any attributes, we are done
	if len(attrRanges) == 0 {
		return fileData, diags
	}

	end := 0
	var sections []string
	for _, attrRange := range attrRanges {
		prev := fileData[end:attrRange.Start.Byte]
		attr := fileData[attrRange.Start.Byte:attrRange.End.Byte]
		// Regex pattern to match unescaped "%{" (not preceded by "%")
		// Regex to match "%{" only if NOT preceded by "%"
		re := regexp.MustCompile(`([^%]|^)%{`)

		// Replace "%{" with "%%{" (but only if not already escaped) while preserving the preceding character
		escapedAttr := re.ReplaceAllString(string(attr), "${1}%%{")

		sections = append(sections, string(prev), escapedAttr)
		end = attrRange.End.Byte
	}
	sections = append(sections, string(fileData[end:]))

	fileData = []byte(strings.Join(sections, ""))

	// reparse the file
	file, diags = hclsyntax.ParseConfig(fileData, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if file == nil {
		return nil, diags
	}
	if diags.HasErrors() {
		return file.Bytes, diags
	}

	// and we are done
	return file.Bytes, diags
}

func getAttributeRanges(file *hcl.File, name string) ([]*hcl.Range, hcl.Diagnostics) {
	return getAttributeRangesFromBody(file.Body, name)
}

func getAttributeRangesFromBody(body hcl.Body, name string) ([]*hcl.Range, hcl.Diagnostics) {
	var ranges []*hcl.Range
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to parse file",
				Detail:   "Failed to parse file body",
			},
		}
	}
	for _, attribute := range syntaxBody.Attributes {
		if attribute.Name == name {
			ranges = append(ranges, &attribute.SrcRange)
		}
	}
	for _, block := range syntaxBody.Blocks {
		moreRanges, moreDiags := getAttributeRangesFromBody(block.Body, name)
		if moreDiags.HasErrors() {
			return nil, moreDiags
		}
		ranges = append(ranges, moreRanges...)
	}
	return ranges, nil
}
