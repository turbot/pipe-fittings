package hclhelpers

import "github.com/zclconf/go-cty/cty"

func IsValueCompatibleWithType(ctyType cty.Type, value cty.Value) bool {

	if ctyType == cty.DynamicPseudoType {
		return true
	}

	valType := value.Type()
	if ctyType.IsCollectionType() {
		if ctyType.ElementType() == cty.DynamicPseudoType {
			return true
		}

		if valType.IsCollectionType() {
			return ctyType.ElementType() == valType.ElementType()
		}

		if valType.IsTupleType() {
			tupleElementTypes := valType.TupleElementTypes()
			for _, tupleElementType := range tupleElementTypes {
				if !ctyType.ElementType().Equals(tupleElementType) {
					return false
				}
			}
			return true
		}
	}

	if ctyType.IsMapType() {
		if valType.IsMapType() {
			return true
		}

		if valType.IsObjectType() {
			return true
		}
	}

	return ctyType == valType
}
