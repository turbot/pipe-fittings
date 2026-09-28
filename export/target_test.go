package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type noopExporter struct{}

func (e *noopExporter) Export(_ context.Context, _ ExportSourceData, _ string) error { return nil }
func (e *noopExporter) FileExtension() string                                        { return ".json" }
func (e *noopExporter) Name() string                                                 { return "noop" }
func (e *noopExporter) Alias() string                                                { return "" }

func TestTarget_Export_Message(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	testCases := []struct {
		name     string
		filePath string
		expect   string
	}{
		{
			name:     "relative path",
			filePath: "output.json",
			expect:   fmt.Sprintf("File exported to %s", filepath.Join(pwd, "output.json")),
		},
		{
			name:     "absolute path",
			filePath: filepath.Join(pwd, "abs", "output.json"),
			expect:   fmt.Sprintf("File exported to %s", filepath.Join(pwd, "abs", "output.json")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := &Target{
				exporter: &noopExporter{},
				filePath: tc.filePath,
			}
			msg, err := target.Export(context.Background(), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg != tc.expect {
				t.Errorf("expected message %q, got %q", tc.expect, msg)
			}
		})
	}
}
