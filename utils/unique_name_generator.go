package utils

import (
	"fmt"
)

type UniqueNameGenerator struct {
	// lookup storing how many times each name has been used
	lookup map[string]int
}

// NewUniqueNameGenerator creates a new UniqueNameGenerator
func NewUniqueNameGenerator() *UniqueNameGenerator {
	return &UniqueNameGenerator{
		lookup: make(map[string]int),
	}
}

// GetUniqueName returns a unique name based on the input name
// If the input name is not unique, a random lowercase string is appended to the name
// This is used in steampipe and powerpipe to ensure unique column names in JSON output
// when same columns are requested.
func (g *UniqueNameGenerator) GetUniqueName(name string) (string, error) {
	uniqueName := name

	// check the lookup to see if this name exists
	nameCount := g.lookup[name]
	if nameCount != 0 {
		// name is not unique - generate a new name using the hash of the name and the name count
		hash, err := Base36Hash(name, 4)
		if err != nil {
			return "", err
		}

		uniqueName = fmt.Sprintf("%s_%s%d", name, hash, nameCount)
	}

	// increment the name count
	nameCount++
	// store the count for the original name in the lookup
	g.lookup[name] = nameCount

	return uniqueName, nil
}
