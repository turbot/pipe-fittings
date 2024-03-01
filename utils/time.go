package utils

import "time"

const RFC3339WithMS = "2006-01-02T15:04:05.000Z07:00"

// TimePtr returns a pointer to the time.Time value passed in.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// TimeNowPtr returns a pointer to the current time in UTC.
func TimeNow() *time.Time {
	return TimePtr(time.Now().UTC())
}
