package backend

import (
	"slices"
	"strings"
	"time"
)

type DatabaseFilters struct {
	// partition wildcards
	Partitions []string
	// the indexes to include
	Indexes []string
	// the data range
	From *time.Time
	To   *time.Time
}

func (o *DatabaseFilters) Equals(other *DatabaseFilters) bool {
	if (o == nil) != (other == nil) ||
		!slices.Equal(o.Partitions, other.Partitions) ||
		!slices.Equal(o.Indexes, other.Indexes) ||
		(o.From == nil) != (other.From == nil) ||
		o.From != nil && !o.From.Equal(*other.From) ||
		(o.To == nil) != (other.To == nil) ||
		o.To != nil && !o.To.Equal(*other.To) {
		return false
	}

	return true
}

func (o *DatabaseFilters) String() string {
	var str strings.Builder
	if len(o.Partitions) > 0 {
		str.WriteString("partitions: ")
		str.WriteString(strings.Join(o.Partitions, ","))
	}
	if len(o.Indexes) > 0 {
		str.WriteString("indexes: ")
		str.WriteString(strings.Join(o.Indexes, ","))
	}
	if o.From != nil {
		str.WriteString("from: ")
		str.WriteString(o.From.String())
	}
	if o.To != nil {
		str.WriteString("to: ")
		str.WriteString(o.To.String())
	}
	return str.String()
}
