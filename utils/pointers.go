package utils

// ToStringPointer converts a string into its pointer
func ToStringPointer(s string) *string {
	return &s
}

// ToIntegerPointer converts an integer into its pointer
func ToIntegerPointer(i int) *int {
	return &i
}

func StringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
