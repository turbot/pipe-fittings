package hclhelpers

import (
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

func IsValueCompatibleWithType(ctyType cty.Type, value cty.Value) bool {
	if ctyType == cty.DynamicPseudoType {
		return true
	}

	valueType := value.Type()

	/**
	  Is this correct? This is to allow declaration such as

	  variable "accessanalyzer_tag_rules" {
		type = object({
			add           = optional(map(string))
			remove        = optional(list(string))
			remove_except = optional(list(string))
			update_keys   = optional(map(list(string)))
			update_values = optional(map(map(list(string))))
		})
		description = "Access Analyzers specific tag rules"
		default     = null
	}


	*/
	if value.IsNull() {
		return true
	}

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
	return typ.IsCollectionType() || typ.IsTupleType() || typ.IsListType() || typ.IsSetType()
}

func IsListLike(typ cty.Type) bool {
	return typ.IsListType() || typ.IsTupleType() || typ.IsSetType()
}

func IsComplexType(typ cty.Type) bool {
	return typ.IsMapType() || typ.IsObjectType() || IsCollectionOrTuple(typ) || typ.IsCapsuleType()
}

func IsMapLike(typ cty.Type) bool {
	return typ.IsMapType() || typ.IsObjectType()
}

func IsNestedCapsuleType(t cty.Type) (reflect.Type, bool) {
	if t.IsCollectionType() {
		// Recursively check the element type if it's a list
		elementType := t.ElementType()
		encapsulatedGoType, ok := IsNestedCapsuleType(elementType)
		return encapsulatedGoType, ok
	} else if t.IsObjectType() {
		// If there's at least one encapsulated type, we say it is a nested capsule type
		attributeTypes := t.AttributeTypes()
		for _, attributeType := range attributeTypes {
			encapsulatedGoType, ok := IsNestedCapsuleType(attributeType)
			if ok {
				return encapsulatedGoType, ok
			}
		}
		return nil, false

	} else if t.IsCapsuleType() {
		// If it's a capsule type, return the encapsulated Go type
		return t.EncapsulatedType(), true
	}
	// Return false if no capsule type is found
	return nil, false
}
