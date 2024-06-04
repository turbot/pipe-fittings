package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ContainsUpper returns true if the string contains any uppercase characters
func ContainsUpper(s string) bool {
	hasUpper := false
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	return hasUpper
}

// ToTitleCase correctly returns a Title cased string
func ToTitleCase(s string) string {
	return cases.Title(language.English).String(s)
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

type UniqueNameGenerator struct {
	lookup map[string]struct{}
}

// ctor
func NewUniqueNameGenerator() *UniqueNameGenerator {
	return &UniqueNameGenerator{
		lookup: make(map[string]struct{}),
	}
}

// randomString generates a random lowercase string of length n.
func (g *UniqueNameGenerator) randomString(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	if n <= 0 {
		return ""
	}

	// Generate n random bytes
	randomBytes := make([]byte, n)
	if _, err := rand.Read(randomBytes); err != nil {
		return ""
	}

	// Map each random byte to a character in the alphabet
	var sb strings.Builder
	for _, b := range randomBytes {
		sb.WriteByte(alphabet[int(b)%len(alphabet)])
	}

	return sb.String()
}

// GetUniqueName returns a unique name based on the input name
// If the input name is not unique, a random lowercase string is appended to the name
// This is used in steampipe and powerpipe to ensure unique column names in JSON output
// when same columns are requested.
func (g *UniqueNameGenerator) GetUniqueName(name string) string {
	// store the original name
	originalName := name

	// ensure a unique column name
	for {
		// check the lookup to see if this name exists
		if _, exists := g.lookup[name]; !exists {
			// name is unique - we are done
			break
		}
		// name is not unique - generate a new name
		name = fmt.Sprintf("%s_%s", originalName, g.randomString(4))
	}
	// add the unique name into the lookup
	g.lookup[name] = struct{}{}
	return name
}
