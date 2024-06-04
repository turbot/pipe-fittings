package utils

import (
	"crypto/rand"
	"encoding/base64"
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

// RandomString generates a random lowercase string of length n.
func RandomString(n int) string {
	if n <= 0 {
		return ""
	}

	// Generate n random bytes
	randomBytes := make([]byte, n)
	if _, err := rand.Read(randomBytes); err != nil {
		return ""
	}

	// Encode the random bytes to a base64 string and return the first n characters
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return strings.ToLower(randomString[:n])
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

// GetUniqueName returns a unique name based on the input name
// If the input name is not unique, a random lowercase string is appended to the name
func (g *UniqueNameGenerator) GetUniqueName(name string) string {
	// ensure a unique column name
	for {
		// check the lookup to see if this name exists
		if _, exists := g.lookup[name]; !exists {
			// name is unique - we are done
			break
		}
		// name is not unique - generate a new name
		// store the original name
		originalName := name

		// generate a new name
		name = fmt.Sprintf("%s_%s", originalName, RandomString(4))
	}
	// add the unique name into the lookup
	g.lookup[name] = struct{}{}
	return name
}
