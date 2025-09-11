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
	var parts []string
	if len(o.Partitions) > 0 {
		parts = append(parts, "partitions:"+strings.Join(o.Partitions, ","))
	}
	if len(o.Indexes) > 0 {
		parts = append(parts, "indexes:"+strings.Join(o.Indexes, ","))
	}
	if o.From != nil {
		parts = append(parts, "from:"+o.From.String())
	}
	if o.To != nil {
		parts = append(parts, "to:"+o.To.String())
	}
	return strings.Join(parts, "|")
}
