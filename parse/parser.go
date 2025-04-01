package parse

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/utils"
	"sigs.k8s.io/yaml"
)

// LoadFileData builds a map of filepath to file data
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}

// ParseHclFiles parses hcl, json or yaml file data and returns the hcl body object
func ParseHclFiles(fileDataMap map[string][]byte, opts ...ParseHclOpt) (hcl.Body, hcl.Diagnostics) {
	config := &ParseHclConfig{}
	for _, opt := range opts {
		opt(config)
	}
	var diags hcl.Diagnostics

	if diags.HasErrors() {
		return nil, diags
	}

	// build ordered list of files so that we parse in a repeatable order
	filePaths := buildOrderedFileNameList(fileDataMap)
	var parsedConfigFiles []*hcl.File

	for _, filePath := range filePaths {
		var file *hcl.File
		var moreDiags hcl.Diagnostics
		ext := filepath.Ext(filePath)

		switch {
		case ext == constants.JsonExtension:
			file, moreDiags = json.ParseFile(filePath)
		case constants.IsYamlExtension(ext):
			file, moreDiags = parseYamlFile(filePath)
		default:
			fileData := fileDataMap[filePath]

			// check for grok function calls - execute these to escape grok expressions
			if config.escapeBackticks {
				fileData, moreDiags = EscapeBackticks(fileDataMap[filePath], filePath)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					continue
				}
			}

			// handle deprecated disableTemplateForProperties
			if len(config.disableTemplateForProperties) > 0 {
				fileData, moreDiags = applyDisableTemplateForProperties(fileData, filePath, config, diags)
				diags = append(diags, moreDiags...)
				if diags.HasErrors() {
					continue
				}
			}

			parser := hclparse.NewParser()
			file, moreDiags = parser.ParseHCL(fileData, filePath)
		}

		if moreDiags.HasErrors() {
			//  detect templata error for grok expressions and raise a warning to use grok function
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
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

func applyDisableTemplateForProperties(fileData []byte, filePath string, config *ParseHclConfig, diags hcl.Diagnostics) ([]byte, hcl.Diagnostics) {
	updatedFileData, moreDiags := EscapeTemplateTokens(fileData, filePath, config.disableTemplateForProperties)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	}

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

func getEscapeTemplateWarningMessage(filePath string, fileData, updatedFileData []byte) string {
	differentLines := findDifferentLines(fileData, updatedFileData)
	lineStr := make([]string, 0, len(differentLines))
	for _, line := range differentLines {
		lineStr = append(lineStr, fmt.Sprintf("%d", line))
	}
	if len(differentLines) == 1 {
		return fmt.Sprintf("The file %q contains a file_layout property containing hcl reserved characters (%s). This has been auto-escaped for you, but future versions will not do this. Please use backticks to escape the property: file_layout = `${val}`.",
			filePath,
			lineStr[0])
	}
	return fmt.Sprintf("The file %q contains file_layout properties containing hcl reserved characters (%s %s). These have been auto-escaped for you, but future versions will not do this. Please use backticks to escape the property: file_layout = `${val}`.",
		filePath,
		utils.Pluralize("line", len(differentLines)),
		strings.Join(lineStr, ", "))

}

func buildOrderedFileNameList(fileData map[string][]byte) []string {
	filePaths := make([]string, len(fileData))
	idx := 0
	for filePath := range fileData {
		filePaths[idx] = filePath
		idx++
	}
	sort.Strings(filePaths)
	return filePaths
}

// ModFileExists returns whether a mod file exists at the specified path and if so returns the filepath
func ModFileExists(modPath string) (string, bool) {
	for _, modFilePath := range app_specific.ModFilePaths(modPath) {
		if _, err := os.Stat(modFilePath); err == nil {
			return modFilePath, true
		}
	}
	return "", false
}

// parse a yaml file into a hcl.File object
func parseYamlFile(filename string) (*hcl.File, hcl.Diagnostics) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to open file",
				Detail:   fmt.Sprintf("The file %q could not be opened.", filename),
			},
		}
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while reading it.", filename),
			},
		}
	}
	jsonData, err := yaml.YAMLToJSON(src)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read convert YAML to JSON",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while converting it to JSON.", filename),
			},
		}
	}
	return json.Parse(jsonData, filename)
}
