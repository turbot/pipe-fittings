package utils

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

// ToStringPointer converts a string into its pointer
func ToStringPointer(s string) *string {
	return &s
}

// ToIntegerPointer converts an integer into its pointer
func ToIntegerPointer(i int) *int {
	return &i
}

func ToPointer[T any](value T) *T {
	return &value
}

func IsPointer[T any](v T) bool {
	// Reflect on the type of the value to determine if it's a pointer
	return reflect.TypeOf(v).Kind() == reflect.Ptr
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
