package parse

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"sigs.k8s.io/yaml"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
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

	// NOTE: Handle grok properties
	// if necessary, escape template tokens in any requested properties
	//for filename, fileData := range fileDataMap {
	// first check for grok function calls - execute these
	//fileData = escapeGrokArgs(fileData, filename)

	// TODO do we need to support the old mechanism and if so for how long?
	//// handle deprecated disableTemplateForProperties
	//if len(config.disableTemplateForProperties) > 0 {
	//	updatedFileData, moreDiags := EscapeTemplateTokens(fileData, filename, config.disableTemplateForProperties)
	//	if moreDiags.HasErrors() {
	//		diags = append(diags, moreDiags...)
	//		continue
	//	}
	//
	//	// if this modified the file data, it means the grok function is not being used - raise a warning
	//	if string(updatedFileData) != string(fileData) {
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagWarning,
	//			Summary:  "Grok expressions should be wrapped in a 'grok' function call",
	//			Detail:   fmt.Sprintf("The file %q contains a Grok expression that is not wrapped in a 'grok' function call. This has been escaped, but this funcitonalityis deprecated and will be removed in a future version.", filename),
	//		})
	//	}
	//	fileData = updatedFileData
	//}

	//	fileDataMap[filename] = fileData
	//}

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
			// TODO K use option to determine whethe rwe need to escape grok expressions
			// first check for grok function calls - execute these to escape grok expressions
			fileData, moreDiags := escapeGrokArgs(fileDataMap[filePath], filePath)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
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
