package utils

import (
	"bytes"
	"testing"
)

// the expected strings are the output of tablewriter v0.0.5 for the same data,
// with the callers' v0 settings (auto-wrap off, left alignment)
func TestNewASCIITable_withHeader(t *testing.T) {
	var buf bytes.Buffer
	table := NewASCIITable(&buf)
	table.Header([]string{"Operation", "Elapsed", "Cumulative"})
	if err := table.Bulk([][]string{
		{"Load config", "1.5ms", "1.5ms"},
		{"A much longer operation name here", "120.25ms", "121.75ms"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := table.Render(); err != nil {
		t.Fatal(err)
	}

	expected := `+-----------------------------------+----------+------------+
|             OPERATION             | ELAPSED  | CUMULATIVE |
+-----------------------------------+----------+------------+
| Load config                       | 1.5ms    | 1.5ms      |
| A much longer operation name here | 120.25ms | 121.75ms   |
+-----------------------------------+----------+------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}

func TestNewASCIITable_noHeader(t *testing.T) {
	var buf bytes.Buffer
	table := NewASCIITable(&buf)
	if err := table.Bulk([][]string{
		{""},
		{"A new version of Steampipe is available! 2.4.4 → 2.4.5"},
		{"You can update by downloading from https://steampipe.io/downloads"},
		{""},
	}); err != nil {
		t.Fatal(err)
	}
	if err := table.Render(); err != nil {
		t.Fatal(err)
	}

	expected := `+-------------------------------------------------------------------+
|                                                                   |
| A new version of Steampipe is available! 2.4.4 → 2.4.5            |
| You can update by downloading from https://steampipe.io/downloads |
|                                                                   |
+-------------------------------------------------------------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}
