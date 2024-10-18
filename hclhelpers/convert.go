package hclhelpers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/go-kit/helpers"
)

func JSONToHcl(jsonString string) (string, hcl.Diagnostics) {
	converted, diags := json.Parse([]byte(jsonString), "")
	if diags.HasErrors() {
		return "", diags
	}

	res, diags := HclBodyToHclString(converted.Body, nil)

	if diags.HasErrors() {
		return "", diags
	}
	return res, nil
}

// HclBodyToHclString builds a hcl string with all attributes in the connection config which are NOT specified in the connection block schema
// this is passed to the plugin who will validate and parse it
func HclBodyToHclString(body hcl.Body, excludeContent *hcl.BodyContent) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// this is a bit messy
	// we want to extract the attributes which are NOT in the connection block schema
	// the body passed in here is the 'rest' result returned from a partial decode, meaning all attributes and blocks
	// in the schema are marked as 'hidden'

	// body.JustAttributes() returns all attributes which are not hidden (i.e. all attributes NOT in the schema)
	//
	// however when calling JustAttributes for a hcl body, it will fail if there are any blocks
	// therefore this code will fail for hcl connection config which has any child blocks (e.g  connection options)
	//
	// it does work however for a json body as this implementation treats blocks as attributes,
	// so the options block is treated as a hidden attribute and excluded
	// we therefore need to treaty hcl and json body separately

	// store map of attribute expressions
	attrExpressionMap := make(map[string]hcl.Expression)

	if hclBody, ok := body.(*hclsyntax.Body); ok {
		// if we can cast to a hcl body, read all the attributes and manually exclude those which are in the schema
		for name, attr := range hclBody.Attributes {
			// if excludeContent was passed, exclude attributes we have already handled
			if excludeContent != nil {
				if _, ok := excludeContent.Attributes[name]; ok {
					continue
				}
			}

			attrExpressionMap[name] = attr.Expr

		}
	} else {
		// so the body was not hcl - we assume it is json
		// try to call JustAttributes
		attrs, diags := body.JustAttributes()
		if diags.HasErrors() {
			return "", diags
		}
		// the attributes returned will only be the ones not in the schema, i.e. we do not need to filter them ourselves
		for name, attr := range attrs {
			attrExpressionMap[name] = attr.Expr
		}
	}

	// build ordered list attributes
	var sortedKeys = helpers.SortedMapKeys(attrExpressionMap)
	for _, name := range sortedKeys {
		expr := attrExpressionMap[name]
		val, moreDiags := expr.Value(nil)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			rootBody.SetAttributeValue(name, val)
		}
	}

	return string(f.Bytes()), diags
}
