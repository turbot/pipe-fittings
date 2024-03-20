package utils

import (
	"fmt"
	"strings"
)

// GetGenericTypeName returns lower case form of type unqualified name
func GetGenericTypeName[T any]() string {
	longName := fmt.Sprintf("%T", *new(T))
	split := strings.Split(longName, ".")
	return split[len(split)-1]
}
