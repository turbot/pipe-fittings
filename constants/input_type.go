package constants

const (
	InputTypeButton      = "button"
	InputTypeMultiSelect = "multiselect"
	InputTypeSelect      = "select"
	InputTypeText        = "text"
)

func IsValidInputType(s string) bool {
	switch s {
	case InputTypeButton, InputTypeMultiSelect, InputTypeSelect, InputTypeText:
		return true
	default:
		return false
	}
}

const (
	InputStyleInfo    = "info"
	InputStyleOk      = "ok"
	InputStyleAlert   = "alert"
	InputStyleDefault = "default"
)

func IsValidInputStyleType(s string) bool {
	switch s {
	case InputStyleInfo, InputStyleOk, InputStyleAlert, InputStyleDefault:
		return true
	default:
		return false
	}
}
