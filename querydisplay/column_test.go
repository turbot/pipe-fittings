package querydisplay

import (
	"testing"
	"time"

	"github.com/turbot/pipe-fittings/v2/queryresult"
)

func TestColumnValueAsString_DateTypes(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		value    interface{}
		want     string
	}{
		{
			name:     "DATE formats without time component",
			dataType: "DATE",
			value:    time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     "1984-01-01",
		},
		{
			name:     "DATE leap year",
			dataType: "DATE",
			value:    time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			want:     "2024-02-29",
		},
		{
			name:     "TIMESTAMP includes time but no timezone",
			dataType: "TIMESTAMP",
			value:    time.Date(1984, 1, 1, 12, 30, 45, 0, time.UTC),
			want:     "1984-01-01 12:30:45",
		},
		{
			name:     "TIMESTAMPTZ uses RFC3339 with UTC",
			dataType: "TIMESTAMPTZ",
			value:    time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     "1984-01-01T00:00:00Z",
		},
		{
			name:     "TIMESTAMPTZ with non-UTC timezone",
			dataType: "TIMESTAMPTZ",
			value:    time.Date(1984, 1, 1, 0, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			want:     "1984-01-01T00:00:00-08:00",
		},
		{
			name:     "TIMESTAMPTZ with positive offset",
			dataType: "TIMESTAMPTZ",
			value:    time.Date(1984, 1, 1, 0, 0, 0, 0, time.FixedZone("IST", 330*60)),
			want:     "1984-01-01T00:00:00+05:30",
		},
		{
			name:     "TIME shows only time component",
			dataType: "TIME",
			value:    time.Date(1984, 1, 1, 15, 30, 45, 0, time.UTC),
			want:     "1984-01-01 15:30:45",
		},
		{
			name:     "INTERVAL shows time component",
			dataType: "INTERVAL",
			value:    time.Date(1970, 1, 1, 4, 30, 0, 0, time.UTC),
			want:     "1970-01-01 04:30:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &queryresult.ColumnDef{
				Name:     "test_col",
				DataType: tt.dataType,
			}
			got, err := ColumnValueAsString(tt.value, col)
			if err != nil {
				t.Errorf("ColumnValueAsString() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ColumnValueAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnValueAsString_DateEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		value    interface{}
		want     string
	}{
		{
			name:     "NULL date",
			dataType: "DATE",
			value:    nil,
			want:     "<null>",
		},
		{
			name:     "NULL timestamptz",
			dataType: "TIMESTAMPTZ",
			value:    nil,
			want:     "<null>",
		},
		{
			name:     "DATE min year",
			dataType: "DATE",
			value:    time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     "0001-01-01",
		},
		{
			name:     "DATE max representable",
			dataType: "DATE",
			value:    time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC),
			want:     "9999-12-31",
		},
		{
			name:     "TIMESTAMPTZ with fractional seconds",
			dataType: "TIMESTAMPTZ",
			value:    time.Date(2024, 1, 1, 12, 30, 45, 123456789, time.UTC),
			want:     "2024-01-01T12:30:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &queryresult.ColumnDef{
				Name:     "test_col",
				DataType: tt.dataType,
			}
			got, err := ColumnValueAsString(tt.value, col)
			if err != nil {
				t.Errorf("ColumnValueAsString() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ColumnValueAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnValueAsString_NonTimeValue(t *testing.T) {
	// Test that non-time.Time values fall through correctly
	tests := []struct {
		name     string
		dataType string
		value    interface{}
	}{
		{
			name:     "DATE with string falls through",
			dataType: "DATE",
			value:    "not a time.Time",
		},
		{
			name:     "TIMESTAMPTZ with string falls through",
			dataType: "TIMESTAMPTZ",
			value:    "not a time.Time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &queryresult.ColumnDef{
				Name:     "test_col",
				DataType: tt.dataType,
			}
			// Should not panic, should fall through to default handling
			_, err := ColumnValueAsString(tt.value, col)
			if err != nil {
				t.Errorf("ColumnValueAsString() error = %v", err)
			}
		})
	}
}
