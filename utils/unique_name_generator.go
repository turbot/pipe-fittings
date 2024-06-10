package utils

import (
	"fmt"
)

type UniqueNameGenerator struct {
	// lookup storing how many times each name has been used
	lookup map[string]struct{}
}

// NewUniqueNameGenerator creates a new UniqueNameGenerator
func NewUniqueNameGenerator() *UniqueNameGenerator {
	return &UniqueNameGenerator{
		lookup: make(map[string]struct{}),
	}
}

// GetUniqueName returns a unique name based on the input name
// If the input name is not unique, a random lowercase string is appended to the name
// This is used in steampipe and powerpipe to ensure unique column names in JSON output
// when same columns are requested.
func (g *UniqueNameGenerator) GetUniqueName(name string, index int) (string, error) {
	uniqueName := name

	// check the lookup to see if this column name already exists
	_, isDuplicate := g.lookup[name]
	if isDuplicate {
		// name is not unique - generate a new name using the hash of the name and the index
		hash, err := Base36Hash(fmt.Sprintf("%s%d", name, index), 4)
		if err != nil {
			return "", err
		}

		uniqueName = fmt.Sprintf("%s_%s", name, hash)
	}

	// store the count for the original name in the lookup
	g.lookup[name] = struct{}{}

	return uniqueName, nil
}
