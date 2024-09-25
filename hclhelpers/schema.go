package hclhelpers

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"reflect"
)

func HclSchemaForStruct(target any) (*hcl.BodySchema, error) {
	// Initialize the schema directly
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     []hcl.BlockHeaderSchema{},
	}

	// Reflect on connImpl to find HCL tags
	connImplType := reflect.TypeOf(target)
	// dereference the pointer type
	if connImplType.Kind() == reflect.Ptr {
		connImplType = connImplType.Elem()
	}

	// Iterate over the fields of ConnectionImpl to identify attributes and blocks
	for i := 0; i < connImplType.NumField(); i++ {
		field := connImplType.Field(i)

		// Get the HCL tag from the field and parse it
		tagStr, ok := field.Tag.Lookup("hcl")
		if !ok {
			continue
		}
		hclTag, err := NewHclTag(tagStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hcl tag for field %s: %v", field.Name, err)
		}

		// Add the tag to the schema based on whether it's a block or an attribute
		if hclTag.Block {
			schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{Type: hclTag.Tag})
		} else {
			schema.Attributes = append(schema.Attributes, hcl.AttributeSchema{Name: hclTag.Tag})
		}
	}
	return schema, nil
}
