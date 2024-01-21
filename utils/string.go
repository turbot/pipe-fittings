package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"unicode"
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
