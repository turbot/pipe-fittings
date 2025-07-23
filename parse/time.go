package parse

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/constants"
)

// ParseTime parses a time string into a time.Time object.
func ParseTime(input string, now time.Time) (time.Time, error) {
	var t time.Time
	var err error
	// check if time is relative
	if strings.HasPrefix(input, "T-") {
		t, err = parseRelativeTime(input, now)
	} else {
		// Handle absolute time formats using go-kit helpers.ParseTime
		t, err = helpers.ParseTime(input)
	}
	if err != nil {
		return time.Time{}, err
	}

	// normalize to UTC
	return t.UTC(), nil
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
