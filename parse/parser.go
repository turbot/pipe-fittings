package parse

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"sigs.k8s.io/yaml"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
)

// LoadFileData builds a map of filepath to file data
// For 4 or more files, reads are parallelized for better I/O performance
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	if len(paths) == 0 {
		return map[string][]byte{}, nil
	}

	// For small number of files, sequential is fine (avoids goroutine overhead)
	if len(paths) < 4 {
		return loadFileDataSequential(paths)
	}

	return loadFileDataParallel(paths)
}

// loadFileDataSequential reads files one at a time
func loadFileDataSequential(paths []string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	fileData := make(map[string][]byte, len(paths))

	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error(),
			})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}

// fileReadResult holds the result of reading a single file
type fileReadResult struct {
	path string
	data []byte
	err  error
}

// loadFileDataParallel reads files concurrently using a worker pool
func loadFileDataParallel(paths []string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	fileData := make(map[string][]byte, len(paths))

	// Use worker pool pattern - limit workers to avoid too many open files
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8 // Cap at 8 to avoid file descriptor limits
	}
	if numWorkers > len(paths) {
		numWorkers = len(paths)
	}

	pathsChan := make(chan string, len(paths))
	resultsChan := make(chan fileReadResult, len(paths))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathsChan {
				data, err := os.ReadFile(path)
				resultsChan <- fileReadResult{path: path, data: data, err: err}
			}
		}()
	}

	// Send work
	for _, path := range paths {
		pathsChan <- path
	}
	close(pathsChan)

	// Wait for completion in separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		if result.err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("failed to read config file %s", result.path),
				Detail:   result.err.Error(),
			})
			continue
		}
		fileData[result.path] = result.data
	}

	return fileData, diags
}

// ParseHclFiles parses hcl, json or yaml file data and returns the hcl body object
func ParseHclFiles(fileDataMap map[string][]byte) (hcl.Body, hcl.Diagnostics) {
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
