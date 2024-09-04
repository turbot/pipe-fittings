package hclhelpers

import (
	"github.com/sagikazarmark/slog-shim"
	"github.com/zclconf/go-cty/cty"
)

func ValidateSettingWithEnum(setting cty.Value, enum cty.Value) (bool, error) {
	if setting.IsNull() {
		return true, nil
	}

	settingGo, err := CtyToGo(setting)
	if err != nil {
		return false, err
	}

	enumGo, err := CtyToGo(enum)
	if err != nil {
		return false, err
	}

	res, err := compareValues(settingGo, enumGo)

	return res, err
}

func compareValues[T comparable](settingGo T, enumGo interface{}) (bool, error) {
	enumGoSlice, ok := enumGo.([]interface{})
	if !ok {
		slog.Debug("enum is not a slice", "enum", enumGo, "setting", settingGo)
		return false, nil
	}

	for _, v := range enumGoSlice {
		if vTyped, ok := v.(T); ok && settingGo == vTyped {
			return true, nil
		}
		if !ok {
			slog.Debug("enum value type mismatch", "enum", enumGo, "setting", settingGo)
			return false, nil
		}
	}

	return false, nil
}
