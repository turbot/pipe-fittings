package cty_helpers

import (
	"log/slog"
	"reflect"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/configs/configschema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// GetCtyTypes builds a map of cty types for all tagged properties.
// It is used to convert the struct to a cty value
func GetCtyTypes(item interface{}) (_ map[string]cty.Type, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("GetCtyTypes failed with panic", "panic", r)
			err = helpers.ToError(r)
		}
	}()
	var res = make(map[string]cty.Type)

	t := reflect.TypeOf(helpers.DereferencePointer(item))
	val := reflect.ValueOf(item)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		structField := t.Field(i)
		attribute, ok := structField.Tag.Lookup("cty")
		if ok && attribute != "-" {
			valField := val.Field(i)
			// get cty type
			ctyType, err := gocty.ImpliedType(valField.Interface())
			if err != nil {
				panic(err)
			}

			res[attribute] = ctyType
		}
	}
	return res, nil
}

// GetCtyValue converts the item into a cty value
func GetCtyValue(item interface{}) (cty.Value, error) {
	// reflect on the item to get the cty types as specified in cty tags
	types, err := GetCtyTypes(item)
	if err != nil {
		return cty.NilVal, err
	}

	// build the block schema
	var block = configschema.Block{Attributes: make(map[string]*configschema.Attribute)}
	for attribute, ctyType := range types {
		block.Attributes[attribute] = &configschema.Attribute{Optional: true, Type: ctyType}
	}

	// get cty spec
	spec := block.DecoderSpec()
	// convert the spec to cty type
	ty := hcldec.ImpliedType(spec)
	// now convert the item to cty value
	return gocty.ToCtyValue(item, ty)
}
