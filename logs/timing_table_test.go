package logs

import (
	"bytes"
	"testing"
)

// the expected strings are the output of tablewriter v0.0.5 for the same data,
// except that its padding of an over-long column to 50 characters is not reproduced
func TestDisplayTable(t *testing.T) {
	var buf bytes.Buffer
	displayTable(&buf, [][]string{
		{"Load config", "2026-01-01T10:00:00Z", "1.5ms", "1.5ms"},
		{"A much longer operation name here", "2026-01-01T10:00:01Z", "120.25ms", "121.75ms"},
	})

	expected := `+-----------------------------------+----------------------+----------+------------+
|             OPERATION             |         TIME         | ELAPSED  | CUMULATIVE |
+-----------------------------------+----------------------+----------+------------+
| Load config                       | 2026-01-01T10:00:00Z | 1.5ms    | 1.5ms      |
| A much longer operation name here | 2026-01-01T10:00:01Z | 120.25ms | 121.75ms   |
+-----------------------------------+----------------------+----------+------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}

func TestDisplayTable_wrapsLongOperation(t *testing.T) {
	var buf bytes.Buffer
	displayTable(&buf, [][]string{
		{"A much longer operation name here that goes well beyond the fifty column width setting", "2026-01-01T10:00:01Z", "120.25ms", "121.75ms"},
	})

	expected := `+--------------------------------------------------+----------------------+----------+------------+
|                    OPERATION                     |         TIME         | ELAPSED  | CUMULATIVE |
+--------------------------------------------------+----------------------+----------+------------+
| A much longer operation name here that goes well | 2026-01-01T10:00:01Z | 120.25ms | 121.75ms   |
| beyond the fifty column width setting            |                      |          |            |
+--------------------------------------------------+----------------------+----------+------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}
