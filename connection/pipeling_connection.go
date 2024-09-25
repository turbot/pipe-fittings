package connection

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

type PipelingConnection interface {
	GetConnectionType() string
	GetShortName() string
	Name() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (PipelingConnection, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	GetEnv() map[string]cty.Value

	Equals(PipelingConnection) bool
	GetConnectionImpl() *ConnectionImpl

	SetTtl(int)
}

func validateMapAttribute(attr *hcl.Attribute, valueMap map[string]cty.Value, key, errMsg string) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if valueMap[key] == cty.NilVal {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  errMsg,
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}
		return append(diags, diag)
	}
	return diags
}

func customTypeValidationSingle(attr *hcl.Attribute, ctyVal cty.Value, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var valueMap map[string]cty.Value
	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		valueMap = ctyVal.AsValueMap()
	}

	if valueMap == nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "value must be a map if the type is a capsule",
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}

		return append(diags, diag)
	}

	encapulatedInstanceNew := reflect.New(encapsulatedGoType)
	if connInterface, ok := encapulatedInstanceNew.Interface().(PipelingConnection); ok {
		diags := validateMapAttribute(attr, valueMap, "type", "missing type in value")
		if len(diags) > 0 {
			return diags
		}

		if connInterface.GetConnectionType() == valueMap["type"].AsString() {
			return diags
		}
	} else if encapsulatedGoType.String() == "*connection.ConnectionImpl" {
		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
		if len(diags) > 0 {
			return diags
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeConnection {
			return diags
		}
	} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
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
	}
	if attr != nil {
		diag.Subject = &attr.Range
	}

	return append(diags, diag)
}

func customTypeCheckResourceTypeCorrect(attr *hcl.Attribute, val cty.Value, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if val.Type().IsMapType() || val.Type().IsObjectType() {
		valueMap := val.AsValueMap()

		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
		if len(diags) > 0 {
			return diags
		}

		encapsulatedInstanceNew := reflect.New(encapsulatedGoType)
		valid := false
		if pc, ok := encapsulatedInstanceNew.Interface().(PipelingConnection); ok {
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
			}
			if attr != nil {
				diag.Subject = &attr.Range
			}
			return append(diags, diag)
		}

		return diags
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "value must be a map if the type is a list of capsules",
	}
	if attr != nil {
		diag.Subject = &attr.Range
	}
	return append(diags, diag)

}

func customTypeValidation(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// short circuit .. if it's object type we can't validate .. it's too complicated right now
	//
	// i.e. object(string, connection.aws, bool)
	if settingType.IsObjectType() {
		return diags
	}

	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		// Validate map or object type
		diags = customTypeValidateMapOrObject(attr, ctyVal, settingType, encapsulatedGoType)
	} else if hclhelpers.IsListLike(ctyVal.Type()) {
		// Validate list type, including nested lists or maps/objects
		diags = customTypeValidateList(attr, ctyVal, settingType, encapsulatedGoType)
	}

	return diags
}

func customTypeValidateMapOrObject(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	valueMap := ctyVal.AsValueMap()

	for _, val := range valueMap {
		if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object types
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list types
			nestedDiags := customTypeValidateList(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}

func customTypeValidateList(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for _, val := range ctyVal.AsValueSlice() {
		if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list
			nestedDiags := customTypeValidateList(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object inside list
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}

func CustomTypeValidation(attr *hcl.Attribute, ctyVal cty.Value, ctyType cty.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// It must be a capsule type OR a list where the element type is a capsule
	encapsulatedGoType, ok := hclhelpers.IsNestedCapsuleType(ctyType)
	if !ok {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Type must be a capsule",
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}
		diags = append(diags, diag)
		return diags
	}

	if ctyType.IsCapsuleType() {
		return customTypeValidationSingle(attr, ctyVal, encapsulatedGoType)
	}

	return customTypeValidation(attr, ctyVal, ctyType, encapsulatedGoType)
}
