package utils

import (
	"math/rand"
	"time"
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

// RandomString generates a random string of length n
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
