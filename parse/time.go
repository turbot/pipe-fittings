package parse

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/constants"
)

// ParseTime parses a time string into a time.Time object.
func ParseTime(input string, now time.Time) (time.Time, error) {
	// Handle absolute time formats
	absoluteLayouts := []string{
		"2006-01-02",              // ISO 8601 date
		"2006-01-02T15:04:05",     // ISO 8601 datetime
		"2006-01-02T15:04:05.000", // ISO 8601 datetime with milliseconds
		time.RFC3339,              // RFC 3339 datetime with timezone
	}

	for _, layout := range absoluteLayouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t.UTC(), nil // Normalize to UTC
		}
	}

	// Handle relative formats
	if strings.HasPrefix(input, "T-") {
		return parseRelativeTime(input, now)
	}

	return time.Time{}, errors.New(constants.InvalidTimeFormat)
}

// parseRelativeTime parses relative time strings.
func parseRelativeTime(input string, now time.Time) (time.Time, error) {
	if len(input) < 3 || !strings.HasPrefix(input, "T-") {
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}

	// Extract the value and unit
	relative := input[2:]
	unit := relative[len(relative)-1]
	value, err := strconv.Atoi(relative[:len(relative)-1])
	if err != nil {
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}

	// Calculate the resulting time
	switch unit {
	case 'Y': // Years
		return now.AddDate(-value, 0, 0), nil
	case 'm': // Months
		return now.AddDate(0, -value, 0), nil
	case 'W': // Weeks
		return now.AddDate(0, 0, -value*7), nil
	case 'd': // Days
		return now.AddDate(0, 0, -value), nil
	case 'H': // Hours
		return now.Add(time.Duration(-value) * time.Hour), nil
	case 'M': // Minutes
		return now.Add(time.Duration(-value) * time.Minute), nil
	default:
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}
}
