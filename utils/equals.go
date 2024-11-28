package utils

import typehelpers "github.com/turbot/go-kit/types"

func SafeStringsEqual(s1, s2 interface{}) bool {
	return typehelpers.SafeString(s1) == typehelpers.SafeString(s2)
}
func SafeIntEqual(i1, i2 *int) bool {
	if i1 != nil {
		if i2 == nil {
			return false
		}
		return *i1 == *i2
	}
	return i2 == nil
}

func SafeBoolEqual(b1, b2 *bool) bool {
	if b1 != nil {
		if b2 == nil {
			return false
		}
		return *b1 == *b2
	}
	return b2 == nil
}
