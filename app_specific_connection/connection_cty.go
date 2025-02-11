package app_specific_connection

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/turbot/pipe-fittings/v2/connection"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/perr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func CtyValueToConnection(value cty.Value) (_ connection.PipelingConnection, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = perr.BadRequestWithMessage("unable to decode connection: " + r.(string))
		}
	}()

	// get the name, block type and range and use to construct a connection
	shortName := value.GetAttr("short_name").AsString()
	connectionType := value.GetAttr("type").AsString()
	var declRange hclhelpers.Range

	err = gocty.FromCtyValue(value.GetAttr("decl_range"), &declRange)
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode decl_range: " + err.Error())
	}

	// now instantiate an empty connection of the correct type
	conn, err := NewPipelingConnection(connectionType, shortName, declRange.HclRange())
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}

	// split the cty value into fields for ConnectionImpl and the derived connection,
	// (NOTE: exclude the 'env', 'type', 'resource_type' fields, which are manually added)
	baseValue, derivedValue, err := getKnownCtyFields(value, conn.GetConnectionImpl(), "env", "type", "resource_type")
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}
	// decode the base fields into the ConnectionImpl
	err = gocty.FromCtyValue(baseValue, conn.GetConnectionImpl())
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode ConnectionImpl: " + err.Error())
	}
	// decode remaining fields into the derived connection
	err = gocty.FromCtyValue(derivedValue, &conn)
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}

	return conn, nil
}

// getKnownCtyFields splits the provided cty.Value object into known and unknown based on the struct's cty tags.
// It returns the two maps as cty.Value objects.
func getKnownCtyFields(ctyVal cty.Value, targetStruct interface{}, excludeFields ...string) (cty.Value, cty.Value, error) {
	if !ctyVal.Type().IsObjectType() {
		return cty.NilVal, cty.NilVal, fmt.Errorf("provided cty.Value is not an object")
	}

	// Extract the map from the cty.Value
	valueMap := ctyVal.AsValueMap()

	known := make(map[string]cty.Value)
	unknown := make(map[string]cty.Value)

	// Reflect on the struct's fields
	val := reflect.ValueOf(targetStruct)
	// dereference pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	// Create a set of known cty tags
	knownTags := make(map[string]bool)
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		ctyTag := field.Tag.Get("cty")
		if ctyTag != "" {
			knownTags[ctyTag] = true
		}
	}

	// Iterate over the provided value map
	for key, value := range valueMap {
		// If the key is a known cty tag, add it to the known map
		if knownTags[key] {
			known[key] = value
		} else if !slices.Contains(excludeFields, key) {
			// if we are not excluding this field, add to unknown map
			unknown[key] = value
		}
	}

	// Return the two maps as cty.Value objects
	return cty.ObjectVal(known), cty.ObjectVal(unknown), nil
}
