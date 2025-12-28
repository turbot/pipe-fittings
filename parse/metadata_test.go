package parse

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// generateTestFile creates a test file with the specified number of lines
func generateTestFile(numLines int) []byte {
	var sb strings.Builder
	for i := 1; i <= numLines; i++ {
		sb.WriteString("line ")
		sb.WriteString(string(rune('0' + i%10)))
		sb.WriteString(" content here with some padding to make it realistic\n")
	}
	return []byte(sb.String())
}

func TestGetSourceDefinition(t *testing.T) {
	fileContent := []byte("line 1\nline 2\nline 3\nline 4\nline 5\n")
	fileData := map[string][]byte{
		"test.pp": fileContent,
	}

	tests := []struct {
		name      string
		srcRange  hcl.Range
		expected  string
	}{
		{
			name: "single line",
			srcRange: hcl.Range{
				Filename: "test.pp",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 7},
			},
			expected: "line 2",
		},
		{
			name: "multiple lines",
			srcRange: hcl.Range{
				Filename: "test.pp",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 4, Column: 7},
			},
			expected: "line 2\nline 3\nline 4",
		},
		{
			name: "first line",
			srcRange: hcl.Range{
				Filename: "test.pp",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 7},
			},
			expected: "line 1",
		},
		{
			name: "last line",
			srcRange: hcl.Range{
				Filename: "test.pp",
				Start:    hcl.Pos{Line: 5, Column: 1},
				End:      hcl.Pos{Line: 5, Column: 7},
			},
			expected: "line 5",
		},
		{
			name: "file not found",
			srcRange: hcl.Range{
				Filename: "nonexistent.pp",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 7},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSourceDefinition(tt.srcRange, fileData)
			if result != tt.expected {
				t.Errorf("getSourceDefinition() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// BenchmarkGetSourceDefinition_Small benchmarks with a small file (100 lines)
func BenchmarkGetSourceDefinition_Small(b *testing.B) {
	benchmarkGetSourceDefinition(b, 100, 10)
}

// BenchmarkGetSourceDefinition_Medium benchmarks with a medium file (1000 lines)
func BenchmarkGetSourceDefinition_Medium(b *testing.B) {
	benchmarkGetSourceDefinition(b, 1000, 50)
}

// BenchmarkGetSourceDefinition_Large benchmarks with a large file (10000 lines)
func BenchmarkGetSourceDefinition_Large(b *testing.B) {
	benchmarkGetSourceDefinition(b, 10000, 100)
}

func benchmarkGetSourceDefinition(b *testing.B, numLines int, numResources int) {
	fileContent := generateTestFile(numLines)
	fileData := map[string][]byte{
		"test.pp": fileContent,
	}

	// Create source ranges spread throughout the file
	ranges := make([]hcl.Range, numResources)
	linesPerResource := numLines / numResources
	for i := 0; i < numResources; i++ {
		startLine := i*linesPerResource + 1
		endLine := startLine + 5 // Each resource spans 5 lines
		if endLine > numLines {
			endLine = numLines
		}
		ranges[i] = hcl.Range{
			Filename: "test.pp",
			Start:    hcl.Pos{Line: startLine, Column: 1},
			End:      hcl.Pos{Line: endLine, Column: 1},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, r := range ranges {
			_ = getSourceDefinition(r, fileData)
		}
	}
}

// getSourceDefinitionOld is the original implementation for comparison
func getSourceDefinitionOld(sourceRange hcl.Range, fileData map[string][]byte) string {
	filename := sourceRange.Filename
	fileBytes, ok := fileData[filename]
	if !ok {
		return ""
	}

	source := strings.Join(
		strings.Split(string(fileBytes), "\n")[sourceRange.Start.Line-1:sourceRange.End.Line], "\n")
	return source
}

// BenchmarkGetSourceDefinition_ManyResourcesSameFile simulates many resources in one file
// This is the worst case for the old implementation
func BenchmarkGetSourceDefinition_ManyResourcesSameFile(b *testing.B) {
	numLines := 5000
	numResources := 200

	fileContent := generateTestFile(numLines)
	fileData := map[string][]byte{
		"test.pp": fileContent,
	}

	ranges := make([]hcl.Range, numResources)
	linesPerResource := numLines / numResources
	for i := 0; i < numResources; i++ {
		startLine := i*linesPerResource + 1
		endLine := startLine + 10
		if endLine > numLines {
			endLine = numLines
		}
		ranges[i] = hcl.Range{
			Filename: "test.pp",
			Start:    hcl.Pos{Line: startLine, Column: 1},
			End:      hcl.Pos{Line: endLine, Column: 1},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, r := range ranges {
			_ = getSourceDefinition(r, fileData)
		}
	}
}

// BenchmarkGetSourceDefinitionOld_ManyResourcesSameFile benchmarks the OLD implementation
// for comparison - this shows the allocation explosion we fixed
func BenchmarkGetSourceDefinitionOld_ManyResourcesSameFile(b *testing.B) {
	numLines := 5000
	numResources := 200

	fileContent := generateTestFile(numLines)
	fileData := map[string][]byte{
		"test.pp": fileContent,
	}

	ranges := make([]hcl.Range, numResources)
	linesPerResource := numLines / numResources
	for i := 0; i < numResources; i++ {
		startLine := i*linesPerResource + 1
		endLine := startLine + 10
		if endLine > numLines {
			endLine = numLines
		}
		ranges[i] = hcl.Range{
			Filename: "test.pp",
			Start:    hcl.Pos{Line: startLine, Column: 1},
			End:      hcl.Pos{Line: endLine, Column: 1},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, r := range ranges {
			_ = getSourceDefinitionOld(r, fileData)
		}
	}
}
