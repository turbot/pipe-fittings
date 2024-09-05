package hclhelpers

import (
	"github.com/sagikazarmark/slog-shim"
	"github.com/zclconf/go-cty/cty"
)

func ValidateSettingWithEnum(setting cty.Value, enum cty.Value) (bool, error) {
	if setting.IsNull() {
		return true, nil
	}

	if enum.IsNull() {
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

	if setting.Type().IsCollectionType() || setting.Type().IsTupleType() {
		settingsGo, ok := settingGo.([]interface{})
		if !ok {
			slog.Debug("setting is not a slice", "setting", settingGo, "enum", enumGo)
			return false, nil
		}

		res, err := compareValuesSlice(settingsGo, enumGo)

		return res, err
	}

	res, err := compareValues(settingGo, enumGo)

	return res, err
}

func compareValuesSlice(settingGo []interface{}, enumGo interface{}) (bool, error) {

	for _, settingValue := range settingGo {
		res, err := compareValues(settingValue, enumGo)
		if err != nil {
			return false, err
		}

		if !res {
			return false, nil
		}
	}

	return true, nil
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
