package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFileData_EmptyPaths(t *testing.T) {
	fileData, diags := LoadFileData()

	assert.Empty(t, diags)
	assert.Empty(t, fileData)
}

func TestLoadFileData_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.pp")
	content := `query "test" { sql = "SELECT 1" }`

	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	fileData, diags := LoadFileData(path)

	assert.Empty(t, diags)
	assert.Len(t, fileData, 1)
	assert.Equal(t, content, string(fileData[path]))
}

func TestLoadFileData_SequentialForSmallSets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 3 files (should use sequential path)
	paths := make([]string, 3)
	expectedData := make(map[string]string)

	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		err := os.WriteFile(path, []byte(content), 0600)
		require.NoError(t, err)
		paths[i] = path
		expectedData[path] = content
	}

	fileData, diags := LoadFileData(paths...)

	assert.Empty(t, diags)
	assert.Len(t, fileData, 3)

	for path, expected := range expectedData {
		assert.Equal(t, expected, string(fileData[path]))
	}
}

func TestLoadFileData_ParallelForLargeSets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 files (should use parallel path)
	numFiles := 50
	paths := make([]string, numFiles)
	expectedData := make(map[string]string)

	for i := 0; i < numFiles; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		err := os.WriteFile(path, []byte(content), 0600)
		require.NoError(t, err)
		paths[i] = path
		expectedData[path] = content
	}

	fileData, diags := LoadFileData(paths...)

	assert.Empty(t, diags)
	assert.Len(t, fileData, numFiles)

	for path, expected := range expectedData {
		assert.Equal(t, expected, string(fileData[path]))
	}
}

func TestLoadFileData_HandlesMissingFiles(t *testing.T) {
	paths := []string{"/nonexistent/file.pp"}

	fileData, diags := LoadFileData(paths...)

	assert.Len(t, diags, 1)
	assert.Equal(t, "failed to read config file /nonexistent/file.pp", diags[0].Summary)
	assert.Len(t, fileData, 0)
}

func TestLoadFileData_HandlesMixedValidAndInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create one valid file
	validPath := filepath.Join(tmpDir, "valid.pp")
	validContent := `query "valid" { sql = "SELECT 1" }`
	err := os.WriteFile(validPath, []byte(validContent), 0600)
	require.NoError(t, err)

	// Include one invalid path
	invalidPath := "/nonexistent/invalid.pp"

	fileData, diags := LoadFileData(validPath, invalidPath)

	// Should have one diagnostic for the missing file
	assert.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary, "invalid.pp")

	// Should have loaded the valid file
	assert.Len(t, fileData, 1)
	assert.Equal(t, validContent, string(fileData[validPath]))
}

func TestLoadFileData_ParallelHandlesMixedValidAndInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 5 valid files (triggers parallel path)
	numValid := 5
	paths := make([]string, numValid+1)
	expectedData := make(map[string]string)

	for i := 0; i < numValid; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		err := os.WriteFile(path, []byte(content), 0600)
		require.NoError(t, err)
		paths[i] = path
		expectedData[path] = content
	}

	// Add one invalid path
	paths[numValid] = "/nonexistent/invalid.pp"

	fileData, diags := LoadFileData(paths...)

	// Should have one diagnostic for the missing file
	assert.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary, "invalid.pp")

	// Should have loaded all valid files
	assert.Len(t, fileData, numValid)

	for path, expected := range expectedData {
		assert.Equal(t, expected, string(fileData[path]))
	}
}

func TestLoadFileData_ExactlyFourFiles(t *testing.T) {
	// Test the boundary condition: exactly 4 files should trigger parallel
	tmpDir := t.TempDir()

	paths := make([]string, 4)
	expectedData := make(map[string]string)

	for i := 0; i < 4; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		err := os.WriteFile(path, []byte(content), 0600)
		require.NoError(t, err)
		paths[i] = path
		expectedData[path] = content
	}

	fileData, diags := LoadFileData(paths...)

	assert.Empty(t, diags)
	assert.Len(t, fileData, 4)

	for path, expected := range expectedData {
		assert.Equal(t, expected, string(fileData[path]))
	}
}

func TestLoadFileData_LargeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with larger content
	numFiles := 10
	paths := make([]string, numFiles)
	expectedData := make(map[string]string)

	for i := 0; i < numFiles; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		// Create larger content (~10KB per file)
		content := strings.Repeat(fmt.Sprintf(`query "q%d_%d" { sql = "SELECT %d" }`+"\n", i, i, i), 200)
		err := os.WriteFile(path, []byte(content), 0600)
		require.NoError(t, err)
		paths[i] = path
		expectedData[path] = content
	}

	fileData, diags := LoadFileData(paths...)

	assert.Empty(t, diags)
	assert.Len(t, fileData, numFiles)

	for path, expected := range expectedData {
		assert.Equal(t, expected, string(fileData[path]))
	}
}

// Benchmarks

func BenchmarkLoadFileData_Sequential(b *testing.B) {
	paths := setupBenchmarkFiles(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = loadFileDataSequential(paths)
	}
}

func BenchmarkLoadFileData_Parallel(b *testing.B) {
	paths := setupBenchmarkFiles(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = loadFileDataParallel(paths)
	}
}

func BenchmarkLoadFileData_SmallSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = LoadFileData(paths...)
	}
}

func BenchmarkLoadFileData_MediumSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = LoadFileData(paths...)
	}
}

func BenchmarkLoadFileData_LargeSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = LoadFileData(paths...)
	}
}

func setupBenchmarkFiles(b *testing.B, count int) []string {
	b.Helper()
	tmpDir := b.TempDir()
	paths := make([]string, count)

	for i := 0; i < count; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file_%d.pp", i))
		// Create realistic-sized file content (~2KB per file)
		content := strings.Repeat(fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`+"\n", i, i), 50)
		err := os.WriteFile(path, []byte(content), 0600)
		if err != nil {
			b.Fatalf("failed to create benchmark file: %v", err)
		}
		paths[i] = path
	}

	return paths
}

// Tests for ParseHclFiles parallel parsing

func TestParseHclFiles_EmptyMap(t *testing.T) {
	body, diags := ParseHclFiles(map[string][]byte{})

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_SingleFile(t *testing.T) {
	fileData := map[string][]byte{
		"/test/file.pp": []byte(`query "test" { sql = "SELECT 1" }`),
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_SequentialForSmallSets(t *testing.T) {
	// 3 files should use sequential path
	fileData := make(map[string][]byte)
	for i := 0; i < 3; i++ {
		path := fmt.Sprintf("/test/file_%d.pp", i)
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		fileData[path] = []byte(content)
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_ParallelForLargeSets(t *testing.T) {
	// 20 files should use parallel path
	numFiles := 20
	fileData := make(map[string][]byte)
	for i := 0; i < numFiles; i++ {
		path := fmt.Sprintf("/test/file_%02d.pp", i)
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		fileData[path] = []byte(content)
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_DeterministicOrder(t *testing.T) {
	// Create 10 files
	fileData := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/test/file_%02d.pp", i)
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		fileData[path] = []byte(content)
	}

	// Parse multiple times and verify same result order
	var firstBlockLabels []string
	for run := 0; run < 5; run++ {
		body, diags := ParseHclFiles(fileData)
		assert.Empty(t, diags)

		content, _ := body.Content(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "query", LabelNames: []string{"name"}},
			},
		})

		var labels []string
		for _, block := range content.Blocks {
			labels = append(labels, block.Labels[0])
		}

		if run == 0 {
			firstBlockLabels = labels
		} else {
			assert.Equal(t, firstBlockLabels, labels, "block order should be deterministic (run %d)", run)
		}
	}
}

func TestParseHclFiles_WithParseErrors(t *testing.T) {
	fileData := map[string][]byte{
		"/test/valid.pp":   []byte(`query "valid" { sql = "SELECT 1" }`),
		"/test/invalid.pp": []byte(`query "invalid" { sql = `), // Syntax error
	}

	body, diags := ParseHclFiles(fileData)

	// Should still return body with valid file
	assert.NotNil(t, body)
	// Should have error from invalid file
	assert.True(t, diags.HasErrors())
}

func TestParseHclFiles_MixedFormats(t *testing.T) {
	fileData := map[string][]byte{
		"/test/a.pp":   []byte(`query "hcl" { sql = "SELECT 1" }`),
		"/test/b.json": []byte(`{"query": {"json": {"sql": "SELECT 2"}}}`),
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_ParallelWithParseErrors(t *testing.T) {
	// 5+ files to trigger parallel, with one invalid
	fileData := make(map[string][]byte)
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/test/file_%d.pp", i)
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		fileData[path] = []byte(content)
	}
	// Add invalid file
	fileData["/test/invalid.pp"] = []byte(`query "bad" { sql = `)

	body, diags := ParseHclFiles(fileData)

	// Should return body with valid files
	assert.NotNil(t, body)
	// Should have error from invalid file
	assert.True(t, diags.HasErrors())
}

func TestParseHclFiles_ExactlyFourFiles(t *testing.T) {
	// Test boundary: exactly 4 files should trigger parallel
	fileData := make(map[string][]byte)
	for i := 0; i < 4; i++ {
		path := fmt.Sprintf("/test/file_%d.pp", i)
		content := fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		fileData[path] = []byte(content)
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

func TestParseHclFiles_YamlFormat(t *testing.T) {
	// Test YAML parsing through the parallel path
	fileData := make(map[string][]byte)
	for i := 0; i < 4; i++ {
		var path string
		var content string
		if i == 0 {
			path = "/test/file.yaml"
			content = `query:
  yaml_query:
    sql: "SELECT 1"`
		} else {
			path = fmt.Sprintf("/test/file_%d.pp", i)
			content = fmt.Sprintf(`query "q%d" { sql = "SELECT %d" }`, i, i)
		}
		fileData[path] = []byte(content)
	}

	body, diags := ParseHclFiles(fileData)

	assert.Empty(t, diags)
	assert.NotNil(t, body)
}

// Benchmarks for ParseHclFiles

func BenchmarkParseHclFiles_Sequential(b *testing.B) {
	fileData := setupBenchmarkFileData(b, 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parseHclFilesSequential(fileData)
	}
}

func BenchmarkParseHclFiles_Parallel(b *testing.B) {
	fileData := setupBenchmarkFileData(b, 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parseHclFilesParallel(fileData)
	}
}

func BenchmarkParseHclFiles_SmallSet(b *testing.B) {
	fileData := setupBenchmarkFileData(b, 3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ParseHclFiles(fileData)
	}
}

func BenchmarkParseHclFiles_MediumSet(b *testing.B) {
	fileData := setupBenchmarkFileData(b, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ParseHclFiles(fileData)
	}
}

func BenchmarkParseHclFiles_LargeSet(b *testing.B) {
	fileData := setupBenchmarkFileData(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ParseHclFiles(fileData)
	}
}

func setupBenchmarkFileData(b *testing.B, count int) map[string][]byte {
	b.Helper()
	fileData := make(map[string][]byte, count)

	for i := 0; i < count; i++ {
		path := fmt.Sprintf("/test/file_%d.pp", i)
		// Create realistic HCL content
		content := fmt.Sprintf(`
query "query_%d" {
    title = "Query %d"
    description = "A test query for benchmarking parallel HCL parsing"
    sql = <<-EOQ
        SELECT
            id,
            name,
            created_at,
            updated_at
        FROM
            table_%d
        WHERE
            status = 'active'
        ORDER BY
            created_at DESC
        LIMIT 100
    EOQ

    param "filter" {
        description = "Filter parameter"
        default = "all"
    }
}
`, i, i, i)
		fileData[path] = []byte(content)
	}

	return fileData
}
