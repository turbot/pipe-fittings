package utils

import "time"

type TimeRange struct {
	From *time.Time `json:"from"`
	To   *time.Time `json:"to"`
}

func (t TimeRange) Equals(other TimeRange) bool {
	if (t.From == nil) != (other.From == nil) ||
		(t.To == nil) != (other.To == nil) ||
		t.From != nil && !t.From.Equal(*other.From) ||
		t.To != nil && !t.To.Equal(*other.To) {
		return false
	}
	return true
}

func (t TimeRange) Empty() bool {
	return t.From == nil && t.To == nil
}
