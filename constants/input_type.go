package constants

const (
	InputTypeButton      = "button"
	InputTypeCombo       = "combo"
	InputTypeMultiCombo  = "multicombo"
	InputTypeMultiSelect = "multiselect"
	InputTypeSelect      = "select"
	InputTypeText        = "text"
)

func IsValidInputType(s string) bool {
	switch s {
	case InputTypeButton, InputTypeCombo, InputTypeMultiCombo, InputTypeMultiSelect, InputTypeSelect, InputTypeText:
		return true
	default:
		return false
	}
}
