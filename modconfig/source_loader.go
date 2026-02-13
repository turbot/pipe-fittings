package modconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// SourceLoader provides lazy loading of HCL source definitions from files.
// Rather than storing the full HCL source in memory for each resource,
// this allows loading source on-demand using file location metadata.
type SourceLoader struct {
	mu sync.RWMutex

	// Cache of loaded sources (optional)
	cache map[string]string

	// Whether to cache loaded sources
	enableCache bool

	// Maximum cache size
	maxCacheSize int
}

// DefaultSourceLoader is the default source loader instance.
// By default, caching is disabled since source loading is typically
// only needed for the occasional "View Source" UI request.
var DefaultSourceLoader = NewSourceLoader(false, 0)

// NewSourceLoader creates a source loader with optional caching.
// If enableCache is true and maxCacheSize > 0, loaded sources will be cached
// up to the specified maximum number of entries.
func NewSourceLoader(enableCache bool, maxCacheSize int) *SourceLoader {
	return &SourceLoader{
		cache:        make(map[string]string),
		enableCache:  enableCache,
		maxCacheSize: maxCacheSize,
	}
}

// LoadSource loads source definition from file using metadata.
// Returns empty string if metadata is nil or has no file name.
func (sl *SourceLoader) LoadSource(meta *ResourceMetadata) (string, error) {
	if meta == nil || meta.FileName == "" {
		return "", nil
	}

	cacheKey := sl.cacheKey(meta)

	// Check cache
	if sl.enableCache {
		sl.mu.RLock()
		if cached, ok := sl.cache[cacheKey]; ok {
			sl.mu.RUnlock()
			return cached, nil
		}
		sl.mu.RUnlock()
	}

	// Load from file
	source, err := sl.loadFromFile(meta)
	if err != nil {
		return "", err
	}

	// Cache if enabled
	if sl.enableCache && sl.maxCacheSize > 0 {
		sl.mu.Lock()
		if len(sl.cache) < sl.maxCacheSize {
			sl.cache[cacheKey] = source
		}
		sl.mu.Unlock()
	}

	return source, nil
}

func (sl *SourceLoader) loadFromFile(meta *ResourceMetadata) (string, error) {
	file, err := os.Open(meta.FileName)
	if err != nil {
		return "", fmt.Errorf("opening source file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var content strings.Builder
	lineNum := 0

	for {
		line, err := reader.ReadString('\n')
		lineNum++

		if lineNum >= meta.StartLineNumber && lineNum <= meta.EndLineNumber {
			content.WriteString(line)
		}

		if lineNum >= meta.EndLineNumber || err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf("reading source file: %w", err)
		}
	}

	// Trim trailing newline to match original HCL parsing behavior
	return strings.TrimSuffix(content.String(), "\n"), nil
}

func (sl *SourceLoader) cacheKey(meta *ResourceMetadata) string {
	return fmt.Sprintf("%s:%d:%d", meta.FileName, meta.StartLineNumber, meta.EndLineNumber)
}

// ClearCache clears the source cache
func (sl *SourceLoader) ClearCache() {
	sl.mu.Lock()
	sl.cache = make(map[string]string)
	sl.mu.Unlock()
}

// LoadSourceDefinition loads a source definition using the default loader.
// This is a convenience function for use when lazy loading is needed.
// Returns empty string on error for graceful degradation.
func LoadSourceDefinition(meta *ResourceMetadata) string {
	source, err := DefaultSourceLoader.LoadSource(meta)
	if err != nil {
		return "" // Return empty on error for graceful degradation
	}
	return source
}
