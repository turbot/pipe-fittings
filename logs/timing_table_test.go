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

// the Time column is formatted with time.StampMilli, which pads days 1-9 with a space
func TestDisplayTable_stampMilliSingleDigitDay(t *testing.T) {
	var buf bytes.Buffer
	displayTable(&buf, [][]string{
		{"Load config", "Jan  2 15:04:05.000", "< 1ms", "< 1ms"},
		{"Next", "Jan 12 15:04:05.123", "120.25ms", "121.75ms"},
	})

	// tablewriter v0.0.5 output for the same data and settings
	expected := `+-------------+---------------------+----------+------------+
|  OPERATION  |        TIME         | ELAPSED  | CUMULATIVE |
+-------------+---------------------+----------+------------+
| Load config | Jan  2 15:04:05.000 | < 1ms    | < 1ms      |
| Next        | Jan 12 15:04:05.123 | 120.25ms | 121.75ms   |
+-------------+---------------------+----------+------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}

// an Operation name that needs no wrapping keeps its runs of spaces, as with tablewriter v0
func TestDisplayTable_shortOperationKeepsSpaces(t *testing.T) {
	var buf bytes.Buffer
	displayTable(&buf, [][]string{
		{"LoadMod start:  mod", "Jan 12 15:04:05.123", "1ms", "1ms"},
	})

	expected := `+---------------------+---------------------+---------+------------+
|      OPERATION      |        TIME         | ELAPSED | CUMULATIVE |
+---------------------+---------------------+---------+------------+
| LoadMod start:  mod | Jan 12 15:04:05.123 | 1ms     | 1ms        |
+---------------------+---------------------+---------+------------+
`
	if buf.String() != expected {
		t.Errorf("unexpected table:\n%s\nexpected:\n%s", buf.String(), expected)
	}
}
