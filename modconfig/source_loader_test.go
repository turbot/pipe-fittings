package modconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceLoader_LoadSource(t *testing.T) {
	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")

	content := `dashboard "test" {
    title = "Test Dashboard"
    description = "A test"
}

query "test_query" {
    sql = "SELECT 1"
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	loader := NewSourceLoader(false, 0)

	// Load dashboard source (lines 1-4)
	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 1,
		EndLineNumber:   4,
	}

	source, err := loader.LoadSource(meta)
	require.NoError(t, err)

	assert.Contains(t, source, "dashboard \"test\"")
	assert.Contains(t, source, "title = \"Test Dashboard\"")
	assert.NotContains(t, source, "query \"test_query\"")
}

func TestSourceLoader_LoadSource_PartialContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")

	content := `line 1
line 2
line 3
line 4
line 5
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	loader := NewSourceLoader(false, 0)

	// Load lines 2-4
	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 2,
		EndLineNumber:   4,
	}

	source, err := loader.LoadSource(meta)
	require.NoError(t, err)

	assert.Contains(t, source, "line 2")
	assert.Contains(t, source, "line 3")
	assert.Contains(t, source, "line 4")
	assert.NotContains(t, source, "line 1")
	assert.NotContains(t, source, "line 5")
}

func TestSourceLoader_WithCache(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")
	require.NoError(t, os.WriteFile(testFile, []byte("query \"test\" { sql = \"SELECT 1\" }"), 0644))

	loader := NewSourceLoader(true, 100)

	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	// First load
	source1, err := loader.LoadSource(meta)
	require.NoError(t, err)

	// Second load (from cache)
	source2, err := loader.LoadSource(meta)
	require.NoError(t, err)

	assert.Equal(t, source1, source2)
}

func TestSourceLoader_CacheClearing(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")
	require.NoError(t, os.WriteFile(testFile, []byte("query \"test\" { sql = \"SELECT 1\" }"), 0644))

	loader := NewSourceLoader(true, 100)

	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	// Load to populate cache
	_, err := loader.LoadSource(meta)
	require.NoError(t, err)

	// Verify cache is populated
	loader.mu.RLock()
	cacheLen := len(loader.cache)
	loader.mu.RUnlock()
	assert.Equal(t, 1, cacheLen)

	// Clear cache
	loader.ClearCache()

	// Verify cache is empty
	loader.mu.RLock()
	cacheLen = len(loader.cache)
	loader.mu.RUnlock()
	assert.Equal(t, 0, cacheLen)
}

func TestSourceLoader_NilMetadata(t *testing.T) {
	loader := NewSourceLoader(false, 0)

	source, err := loader.LoadSource(nil)
	require.NoError(t, err)
	assert.Equal(t, "", source)
}

func TestSourceLoader_EmptyFileName(t *testing.T) {
	loader := NewSourceLoader(false, 0)

	meta := &ResourceMetadata{
		FileName:        "",
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	source, err := loader.LoadSource(meta)
	require.NoError(t, err)
	assert.Equal(t, "", source)
}

func TestSourceLoader_FileNotFound(t *testing.T) {
	loader := NewSourceLoader(false, 0)

	meta := &ResourceMetadata{
		FileName:        "/nonexistent/file.pp",
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	source, err := loader.LoadSource(meta)
	assert.Error(t, err)
	assert.Equal(t, "", source)
}

func TestResourceMetadata_GetSourceDefinition_Cached(t *testing.T) {
	meta := &ResourceMetadata{
		SourceDefinition: "cached source content",
	}

	source := meta.GetSourceDefinition()
	assert.Equal(t, "cached source content", source)
}

func TestResourceMetadata_GetSourceDefinition_Lazy(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")
	require.NoError(t, os.WriteFile(testFile, []byte("query \"test\" { sql = \"SELECT 1\" }"), 0644))

	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	source := meta.GetSourceDefinition()
	assert.Contains(t, source, "query \"test\"")
}

func TestResourceMetadata_ClearSourceDefinition(t *testing.T) {
	meta := &ResourceMetadata{
		SourceDefinition: "cached source",
	}

	assert.Equal(t, "cached source", meta.GetSourceDefinition())

	meta.ClearSourceDefinition()

	// After clear, GetSourceDefinition returns empty (no file to load from)
	assert.Equal(t, "", meta.GetSourceDefinition())
}

func TestResourceMetadata_SetSourceDefinition(t *testing.T) {
	meta := &ResourceMetadata{}

	meta.SetSourceDefinition("new source content")
	assert.Equal(t, "new source content", meta.SourceDefinition)
	assert.Equal(t, "new source content", meta.GetSourceDefinition())
}

func TestResourceMetadata_LazyAfterClear(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")
	testContent := "control \"test\" { title = \"Test Control\" }"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	meta := &ResourceMetadata{
		FileName:         testFile,
		StartLineNumber:  1,
		EndLineNumber:    1,
		SourceDefinition: "original cached source",
	}

	// Initially returns cached value
	assert.Equal(t, "original cached source", meta.GetSourceDefinition())

	// Clear the cached source
	meta.ClearSourceDefinition()

	// Now should load from file
	source := meta.GetSourceDefinition()
	assert.Contains(t, source, "control \"test\"")
	assert.Contains(t, source, "Test Control")
}

func TestLoadSourceDefinition_ConvenienceFunction(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pp")
	require.NoError(t, os.WriteFile(testFile, []byte("benchmark \"test\" { }"), 0644))

	meta := &ResourceMetadata{
		FileName:        testFile,
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	source := LoadSourceDefinition(meta)
	assert.Contains(t, source, "benchmark \"test\"")
}

func TestLoadSourceDefinition_GracefulError(t *testing.T) {
	meta := &ResourceMetadata{
		FileName:        "/nonexistent/file.pp",
		StartLineNumber: 1,
		EndLineNumber:   1,
	}

	// Should return empty string on error (graceful degradation)
	source := LoadSourceDefinition(meta)
	assert.Equal(t, "", source)
}
