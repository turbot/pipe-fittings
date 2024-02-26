package utils

import "golang.org/x/exp/constraints"

// ToStringPointer converts a string into its pointer
func ToStringPointer(s string) *string {
	return &s
}

// ToIntegerPointer converts an integer into its pointer
func ToIntegerPointer(i int) *int {
	return &i
}

func PtrEqual[T constraints.Ordered](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func BoolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
