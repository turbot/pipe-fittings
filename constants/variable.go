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
)

var ValidVariableFormats = []string{
	VariableFormatText,
	VariableFormatMultiline,
}

func IsValidVariableFormat(format string) bool {
	return slices.Contains(ValidVariableFormats, format)
}
