package parse

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"2024-12-12", "2024-12-12T00:00:00Z", false},
		{"2024-12-12T15:00:00", "2024-12-12T15:00:00Z", false},
		{"2024-12-12T15:00:00.123", "2024-12-12T15:00:00.123Z", false},
		{"2024-12-12T15:00:00Z", "2024-12-12T15:00:00Z", false},
		{"2024-12-12T15:00:00+10:00", "2024-12-12T05:00:00Z", false},
		{"T-2Y", "2022-12-12T15:00:00Z", false},
		{"T-10m", "2024-02-12T15:00:00Z", false},
		{"T-10W", "2024-10-03T15:00:00Z", false},
		{"T-180d", "2024-06-15T15:00:00Z", false},
		{"T-9H", "2024-12-12T06:00:00Z", false},
		{"T-10M", "2024-12-12T14:50:00Z", false},
		{"invalid-date", "", true},
		{"T-invalid", "", true},
		{"T-10X", "", true},
	}

	// Mocked current time
	now := time.Date(2024, 12, 12, 15, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		parsedTime, err := ParseTime(tt.input, now)

		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for input '%s', got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("did not expect error for input '%s', got %v", tt.input, err)
			} else if parsedTime.Format(time.RFC3339Nano) != tt.expected {
				t.Errorf("expected '%s' for input '%s', got '%s'", tt.expected, tt.input, parsedTime.Format(time.RFC3339Nano))
			}
		}
	}
}
