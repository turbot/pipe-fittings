package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	err := os.WriteFile(path, []byte(content), 0644)
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
		err := os.WriteFile(path, []byte(content), 0644)
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
		err := os.WriteFile(path, []byte(content), 0644)
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
	err := os.WriteFile(validPath, []byte(validContent), 0644)
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
		err := os.WriteFile(path, []byte(content), 0644)
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
		err := os.WriteFile(path, []byte(content), 0644)
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
		err := os.WriteFile(path, []byte(content), 0644)
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
		loadFileDataSequential(paths)
	}
}

func BenchmarkLoadFileData_Parallel(b *testing.B) {
	paths := setupBenchmarkFiles(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		loadFileDataParallel(paths)
	}
}

func BenchmarkLoadFileData_SmallSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		LoadFileData(paths...)
	}
}

func BenchmarkLoadFileData_MediumSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		LoadFileData(paths...)
	}
}

func BenchmarkLoadFileData_LargeSet(b *testing.B) {
	paths := setupBenchmarkFiles(b, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		LoadFileData(paths...)
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
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			b.Fatalf("failed to create benchmark file: %v", err)
		}
		paths[i] = path
	}

	return paths
}
