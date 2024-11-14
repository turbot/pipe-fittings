package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// GetGenericTypeName returns lower case form of type unqualified name
func GetGenericTypeName[T any]() string {
	longName := fmt.Sprintf("%T", *new(T))
	split := strings.Split(longName, ".")
	return split[len(split)-1]
}

func InstanceOf[T any]() T {
	var x T
	// is x a pointee
	if reflect.ValueOf(x).Kind() != reflect.Ptr {
		return x
	}
	var y interface{} = x
	rv := reflect.ValueOf(y)
	t := rv.Type().Elem()
	y = reflect.New(t).Interface()
	return y.(T)
}
