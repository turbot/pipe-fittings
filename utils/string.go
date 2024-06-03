package utils

import (
	"crypto/rand"
	"encoding/base64"
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

// RandomString generates a random string of length n.
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
	return randomString[:n]
}
