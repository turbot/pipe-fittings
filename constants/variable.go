package constants

import (
	"slices"
)

const (
	LateBindingVarsKey = "late_binding_vars"
)

const (
	VariableFormatText      = "text"
	VariableFormatMultiline = "multiline"
	VariableFormatJson      = "json"
)

var ValidVariableFormats = []string{
	VariableFormatText,
	VariableFormatMultiline,
	VariableFormatJson,
}

func IsValidVariableFormat(format string) bool {
	return slices.Contains(ValidVariableFormats, format)
}
