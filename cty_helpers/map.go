package cty_helpers

import "github.com/zclconf/go-cty/cty"

func StringValueFromCtyMap(valueMap map[string]cty.Value, key string) (string, bool) {
	if valueMap[key] == cty.NilVal ||
		valueMap[key].IsNull() ||
		valueMap[key].Type() != cty.String {
		return "", false
	}

	return valueMap[key].AsString(), true
}
