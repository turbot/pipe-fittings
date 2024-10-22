package utils

import (
	"reflect"
	"slices"

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

func IsPointer[T any](val T) bool {
	// Check if the type is nil to avoid panics
	if reflect.TypeOf(val) == nil {
		return false
	}
	return reflect.TypeOf(val).Kind() == reflect.Ptr
}

func IsInterface[T any](val T) bool {
	// Check if the type is nil to avoid panics
	if reflect.TypeOf(val) == nil {
		return false
	}
	return reflect.TypeOf(val).Kind() == reflect.Interface
}

// Deref safely dereferences a pointer, returning a default value if the pointer is nil
func Deref[T any](p *T, defaultVal T) T {
	if p != nil {
		return *p
	}
	return defaultVal
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

func SlicePtrEqual[T constraints.Ordered](a, b *[]T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return slices.Equal(*a, *b)
}
