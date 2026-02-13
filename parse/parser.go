package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"sigs.k8s.io/yaml"

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
// For 4 or more files, parsing is parallelized for better CPU utilization
func ParseHclFiles(fileDataMap map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	if len(fileDataMap) == 0 {
		return hcl.EmptyBody(), nil
	}

	// For small number of files, sequential is fine (avoids goroutine overhead)
	if len(fileDataMap) < 4 {
		return parseHclFilesSequential(fileDataMap)
	}

	return parseHclFilesParallel(fileDataMap)
}

// parseHclFilesSequential parses files one at a time
func parseHclFilesSequential(fileDataMap map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	filePaths := buildOrderedFileNameList(fileDataMap)
	parsedConfigFiles := make([]*hcl.File, 0, len(filePaths))

	for _, filePath := range filePaths {
		file, moreDiags := parseHclFile(filePath, fileDataMap[filePath])
		diags = append(diags, moreDiags...)
		if file != nil {
			parsedConfigFiles = append(parsedConfigFiles, file)
		}
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

// parseResult holds the result of parsing a single file
type parseResult struct {
	path  string
	file  *hcl.File
	diags hcl.Diagnostics
}

// parseHclFilesParallel parses files concurrently using a worker pool
func parseHclFilesParallel(fileDataMap map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	filePaths := buildOrderedFileNameList(fileDataMap)
	numFiles := len(filePaths)

	// Use worker pool pattern
	numWorkers := runtime.NumCPU()
	if numWorkers > numFiles {
		numWorkers = numFiles
	}

	workChan := make(chan string, numFiles)
	resultsChan := make(chan parseResult, numFiles)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range workChan {
				file, diags := parseHclFile(path, fileDataMap[path])
				resultsChan <- parseResult{path: path, file: file, diags: diags}
			}
		}()
	}

	// Send work
	for _, path := range filePaths {
		workChan <- path
	}
	close(workChan)

	// Wait for completion in separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results - need to maintain order for deterministic output
	resultMap := make(map[string]parseResult, numFiles)
	for result := range resultsChan {
		resultMap[result.path] = result
	}

	// Build ordered output
	var diags hcl.Diagnostics
	parsedConfigFiles := make([]*hcl.File, 0, numFiles)

	for _, path := range filePaths {
		result := resultMap[path]
		diags = append(diags, result.diags...)
		if result.file != nil {
			parsedConfigFiles = append(parsedConfigFiles, result.file)
		}
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

// parseHclFile parses a single file based on its extension using already-loaded data
func parseHclFile(filePath string, data []byte) (*hcl.File, hcl.Diagnostics) {
	ext := filepath.Ext(filePath)

	switch {
	case ext == constants.JsonExtension:
		return json.Parse(data, filePath)
	case constants.IsYamlExtension(ext):
		return parseYamlData(data, filePath)
	default:
		parser := hclparse.NewParser()
		return parser.ParseHCL(data, filePath)
	}
}

// parseYamlData parses YAML from already-loaded data
func parseYamlData(data []byte, filename string) (*hcl.File, hcl.Diagnostics) {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to convert YAML to JSON",
				Detail:   fmt.Sprintf("Error converting %s: %v", filename, err),
			},
		}
	}
	return json.Parse(jsonData, filename)
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

