package utils

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const RFC3339WithMS = "2006-01-02T15:04:05.000Z07:00"

// TimePtr returns a pointer to the time.Time value passed in.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// TimeNowPtr returns a pointer to the current time in UTC.
func TimeNow() *time.Time {
	return TimePtr(time.Now().UTC())
}

// FormatTime formats time as RFC3339 in UTC
func FormatTime(localTime time.Time) string {
	loc, _ := time.LoadLocation("UTC")
	utcTime := localTime.In(loc)
	return (utcTime.Format(time.RFC3339))
}

// JSONTime is a timestamp that can be serialized into JSON format
type JSONTime struct {
	time.Time
}

// MarshalJSON converts a JSONTime struct into a JSON byte buf
func (j *JSONTime) MarshalJSON() ([]byte, error) {
	if j.IsZero() {
		return []byte("null"), nil
	}
	ts := fmt.Sprintf("\"%s\"", j.Format(RFC3339WithMS))
	return []byte(ts), nil
}

// UnmarshalJSON converts buf into a JSONTime struct
func (j *JSONTime) UnmarshalJSON(buf []byte) error {
	s := string(buf)
	// If it's set to null, then return nil for empty / null
	if s == "null" {
		return nil
	}
	// Otherwise, assume it's likely a datetime string (quoted in JSON)
	rawStr := strings.Trim(string(buf), `"`)
	bt, err := time.Parse(RFC3339WithMS, rawStr)
	if err != nil {
		// oops ... couldn't parse the time
		return err
	}
	j.Time = bt
	return nil
}

// https://stackoverflow.com/questions/61630216/converting-time-from-db-to-custom-time-fails
func (j *JSONTime) Scan(src interface{}) error {
	if t, ok := src.(time.Time); ok {
		j.Time = t.UTC()
	}
	return nil
}

func (j *JSONTime) Value() (driver.Value, error) {
	return j.Time, nil
}
