package hclhelpers

import (
	"github.com/zclconf/go-cty/cty"
)

func IsValueCompatibleWithType(ctyType cty.Type, value cty.Value) bool {
	if ctyType == cty.DynamicPseudoType {
		return true
	}

	valueType := value.Type()

	if ctyType.IsMapType() || ctyType.IsObjectType() {
		if valueType.IsMapType() || valueType.IsObjectType() {
			if ctyType.IsCollectionType() {
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
			} else if ctyType.IsObjectType() {
				typeMapTypes := ctyType.AttributeTypes()
				for name, typeValue := range typeMapTypes {
					if valueType.HasAttribute(name) {
						innerValue := value.GetAttr(name)
						if !IsValueCompatibleWithType(typeValue, innerValue) {
							return false
						}
					} else {
						return false
					}
				}
				return true
			}
		}
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
					// must be primitive type
				} else if !IsValueCompatibleWithType(ctyType.ElementType(), cty.UnknownVal(tupleElementType)) {
					return false
				}
			}
			return true
		}
	}

	return ctyType.Equals(valueType)
}

func IsEnumValueCompatibleWithType(ctyType cty.Type, enumValues cty.Value) bool {
	if ctyType == cty.DynamicPseudoType {
		return true
	}

	innerCtyType := ctyType
	// if ctyType is not a scalar type, then pull the element type
	if ctyType.IsCollectionType() {
		innerCtyType = ctyType.ElementType()
	} else if ctyType.IsTupleType() {
		innerCtyType = ctyType.TupleElementTypes()[0]
	}

	if innerCtyType == cty.DynamicPseudoType {
		return false
	}

	valueType := enumValues.Type()
	if !valueType.IsCollectionType() && !valueType.IsTupleType() {
		return false
	}

	var mapElementType cty.Type
	if valueType.IsCollectionType() {
		mapElementType = valueType.ElementType()
	} else {
		mapElementType = valueType.TupleElementTypes()[0]
	}

	if mapElementType != innerCtyType {
		return false
	}

	return true
}

// Checks if the given type is a collection or a tuple
func IsCollectionOrTuple(typ cty.Type) bool {
	return typ.IsCollectionType() || typ.IsTupleType() || typ.IsListType()
}
