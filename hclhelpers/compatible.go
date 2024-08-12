package hclhelpers

import "github.com/zclconf/go-cty/cty"

func IsValueCompatibleWithType(ctyType cty.Type, value cty.Value) bool {
	if ctyType == cty.DynamicPseudoType {
		return true
	}

	valueType := value.Type()

	if ctyType.IsMapType() {
		if valueType.IsMapType() || valueType.IsObjectType() {
			mapElementType := ctyType.ElementType()

			// Ensure the value is known before iterating over it to avoid panic
			if value.IsKnown() {
				for it := value.ElementIterator(); it.Next(); {
					_, mapValue := it.Element()
					if !IsValueCompatibleWithType(mapElementType, mapValue) {
						return false
					}
				}
				return true
			} else {
				return false
			}
		}

		// if valType.IsObjectType() {
		// 	return true
		// }
	}

	if ctyType.IsCollectionType() {
		if ctyType.ElementType() == cty.DynamicPseudoType {
			return true
		}

		if valueType.IsCollectionType() {
			elementType := ctyType.ElementType()
			for it := value.ElementIterator(); it.Next(); {
				_, elementValue := it.Element()
				// Recursive check for nested collection types
				if !IsValueCompatibleWithType(elementType, elementValue) {
					return false
				}
			}
			return true
		}

		if valueType.IsTupleType() {
			tupleElementTypes := valueType.TupleElementTypes()
			for i, tupleElementType := range tupleElementTypes {
				if tupleElementType.IsObjectType() || tupleElementType.IsMapType() {
					nestedValue := value.Index(cty.NumberIntVal(int64(i)))
					if !IsValueCompatibleWithType(ctyType.ElementType(), nestedValue) {
						return false
					}
				} else if tupleElementType.IsCollectionType() {
					nestedValue := value.Index(cty.NumberIntVal(int64(i)))
					if !IsValueCompatibleWithType(ctyType.ElementType(), nestedValue) {
						return false
					}
				} else {
					// must be primitive type
					if !IsValueCompatibleWithType(ctyType.ElementType(), cty.UnknownVal(tupleElementType)) {
						return false
					}
				}
			}
			return true
		}
	}

	return ctyType.Equals(valueType)
}
