package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

// customType is an interface that custom cty types must implement
type customType interface {
	CustomType()
}
type lateBindingType interface {
	LateBinding()
}

// IsLateBinding returns true if the  type is late binding, i.e. the value is resolved at run time rather than at parse time.
func IsLateBindingType(ty cty.Type) bool {
	encapsulatedGoType, nestedCapsule := hclhelpers.IsNestedCapsuleType(ty)
	if !nestedCapsule {
		return false
	}

	// dereference the pointer
	if encapsulatedGoType.Kind() == reflect.Ptr {
		encapsulatedGoType = encapsulatedGoType.Elem()
	}
	encapulatedInstanceNew := reflect.New(encapsulatedGoType)

	_, isLateBindingType := encapulatedInstanceNew.Interface().(lateBindingType)

	return isLateBindingType
}

// IsCustomType returns true if the given cty.Type is a custom type, as determined by the customType interface
func IsCustomType(ty cty.Type) bool {
	encapsulatedGoType, nestedCapsule := hclhelpers.IsNestedCapsuleType(ty)
	if !nestedCapsule {
		return false
	}

	// dereference the pointer
	if encapsulatedGoType.Kind() == reflect.Ptr {
		encapsulatedGoType = encapsulatedGoType.Elem()
	}
	encapulatedInstanceNew := reflect.New(encapsulatedGoType)

	_, isCustomType := encapulatedInstanceNew.Interface().(customType)

	return isCustomType
}

func ValidateValueMatchesType(ctyVal cty.Value, ty cty.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	if ty != cty.DynamicPseudoType && ctyVal.Type() != ty {
		if IsCustomType(ty) {
			return CustomTypeValidation(ctyVal, ty, sourceRange)
		}
		if !hclhelpers.IsValueCompatibleWithType(ty, ctyVal) {
			return hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("default value type mismatched - expected %s, got %s", ty.FriendlyName(), ctyVal.Type().FriendlyName()),
					Subject:  sourceRange},
			}
		}
	}

	return nil
}

func CustomTypeValidation(ctyVal cty.Value, ctyType cty.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// It must be a capsule type OR a list where the element type is a capsule
	encapsulatedGoType, ok := hclhelpers.IsNestedCapsuleType(ctyType)
	if !ok {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Type must be a capsule",
			Subject:  sourceRange,
		}

		diags = append(diags, diag)
		return diags
	}

	if ctyType.IsCapsuleType() {
		return customTypeValidationSingle(ctyVal, encapsulatedGoType, sourceRange)
	}

	return customTypeValidation(ctyVal, ctyType, encapsulatedGoType, sourceRange)
}

func validateMapAttribute(valueMap map[string]cty.Value, key string, errMsg string, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if valueMap[key] == cty.NilVal {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  errMsg,
			Subject:  sourceRange,
		}

		return append(diags, diag)
	}
	return diags
}

func customTypeValidationSingle(ctyVal cty.Value, encapsulatedGoType reflect.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var valueMap map[string]cty.Value
	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		valueMap = ctyVal.AsValueMap()
	}

	if valueMap == nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "value must be a map if the type is a capsule",
			Subject:  sourceRange,
		}

		return append(diags, diag)
	}

	encapulatedInstanceNew := reflect.New(encapsulatedGoType)
	if connInterface, ok := encapulatedInstanceNew.Interface().(connection.PipelingConnection); ok {
		diags := validateMapAttribute(valueMap, "type", "missing type in value", sourceRange)
		if len(diags) > 0 {
			return diags
		}

		if connInterface.GetConnectionType() == valueMap["type"].AsString() {
			return diags
		}
	} else if encapsulatedGoType.String() == "*connection.ConnectionImpl" {
		diags := validateMapAttribute(valueMap, "resource_type", "missing resource_type in value", sourceRange)
		if len(diags) > 0 {
			return diags
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeConnection {
			return diags
		}
	} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
		diags := validateMapAttribute(valueMap, "resource_type", "missing resource_type in value", sourceRange)
		if len(diags) > 0 {
			return diags
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeNotifier {
			return diags
		}
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "value type mismatched with the capsule type",
		Subject:  sourceRange,
	}

	return append(diags, diag)
}

func customTypeCheckResourceTypeCorrect(val cty.Value, encapsulatedGoType reflect.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if val.Type().IsMapType() || val.Type().IsObjectType() {
		valueMap := val.AsValueMap()

		diags := validateMapAttribute(valueMap, "resource_type", "missing resource_type in value", sourceRange)
		if len(diags) > 0 {
			return diags
		}

		encapsulatedInstanceNew := reflect.New(encapsulatedGoType)
		valid := false
		if pc, ok := encapsulatedInstanceNew.Interface().(connection.PipelingConnection); ok {
			// Validate list of capsule type
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeConnection && valueMap["type"].AsString() == pc.GetConnectionType()
		} else if encapsulatedGoType.String() == "*connection.ConnectionImpl" {
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeConnection
		} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
			// Validate internal notifier resource
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeNotifier
		}

		if !valid {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "value type mismatched with the capsule type",
				Subject:  sourceRange,
			}

			return append(diags, diag)
		}

		return diags
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "value must be a map if the type is a list of capsules",
		Subject:  sourceRange,
	}

	return append(diags, diag)

}

func customTypeValidation(ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// short circuit .. if it's object type we can't validate .. it's too complicated right now
	//
	// i.e. object(string, connection.aws, bool)
	if settingType.IsObjectType() {
		return diags
	}

	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		// Validate map or object type
		diags = customTypeValidateMapOrObject(ctyVal, settingType, encapsulatedGoType, sourceRange)
	} else if hclhelpers.IsListLike(ctyVal.Type()) {
		// Validate list type, including nested lists or maps/objects
		diags = customTypeValidateList(ctyVal, settingType, encapsulatedGoType, sourceRange)
	}

	return diags
}

func customTypeValidateMapOrObject(ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	valueMap := ctyVal.AsValueMap()

	for _, val := range valueMap {
		if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object types
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(val, encapsulatedGoType, sourceRange)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(val, settingType, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		} else if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list types
			nestedDiags := customTypeValidateList(val, settingType, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(val, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}

func customTypeValidateList(ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type, sourceRange *hcl.Range) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for _, val := range ctyVal.AsValueSlice() {
		if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list
			nestedDiags := customTypeValidateList(val, settingType, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		} else if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object inside list
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(val, encapsulatedGoType, sourceRange)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(val, settingType, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(val, encapsulatedGoType, sourceRange)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}