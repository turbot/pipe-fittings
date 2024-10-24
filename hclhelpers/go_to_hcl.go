package hclhelpers

import (
	"fmt"
	"sort"

	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
)

// GoToHCLString converts a Go data structure to an HCL string.
func GoToHCLString(data interface{}) (string, error) {
	// Convert the data into a cty.Value, which is needed for HCL encoding
	ctyVal, err := ConvertInterfaceToCtyValue(data)
	if err != nil {
		return "", err
	}

	// Encode the cty.Value into HCL format directly without wrapping in a block
	val, err := ctyValueToHCLString(ctyVal)

	return val, err
}

func ctyValueToHCLString(val cty.Value) (string, error) {
	if val.IsNull() {
		return "null", nil
	}

	switch val.Type() {
	case cty.String:
		return fmt.Sprintf(`"%s"`, val.AsString()), nil
	case cty.Number:
		num, _ := val.AsBigFloat().Float64()
		return fmt.Sprintf("%v", num), nil
	case cty.Bool:
		return fmt.Sprintf("%v", val.True()), nil
	default:
		if val.Type().IsMapType() || val.Type().IsObjectType() {
			val, err := mapToHCLString(val)
			return val, err
		} else if val.Type().IsListType() || val.Type().IsTupleType() || val.Type().IsSetType() {
			elements := val.AsValueSlice()
			var hclElements []string
			for _, element := range elements {
				hclElement, err := ctyValueToHCLString(element)
				if err != nil {
					return "", err
				}
				hclElements = append(hclElements, hclElement)
			}
			return fmt.Sprintf("[%s]", join(hclElements, ", ")), nil
		}

		return "", perr.BadRequestWithMessage("Unsupported cty type " + val.Type().FriendlyName())
	}
}

// Convert a cty.Value of type map or object to its HCL string representation
func mapToHCLString(val cty.Value) (string, error) {
	if !val.CanIterateElements() {
		return "", perr.BadRequestWithMessage("Expected a map or object type")
	}

	valMap := val.AsValueMap()

	// need stable map otherwise testing is difficult
	keys := make([]string, 0, len(valMap))
	for key := range valMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var hclElements []string
	for _, key := range keys {
		element := valMap[key]
		hclElement, err := ctyValueToHCLString(element)
		if err != nil {
			return "", err
		}
		hclElements = append(hclElements, fmt.Sprintf(`%s = %s`, key, hclElement))
	}
	return fmt.Sprintf("{%s}", join(hclElements, ", ")), nil
}

// Join helper function to join elements with a separator
func join(elements []string, separator string) string {
	result := ""
	for i, element := range elements {
		if i > 0 {
			result += separator
		}
		result += element
	}
	return result
}

// // encodeAttributesToHCL encodes the attributes of a cty.Value to HCL
// func encodeAttributesToHCL(body *hclwrite.Body, val cty.Value) {
// 	if val.Type().IsObjectType() || val.Type().IsMapType() {
// 		for k, v := range val.AsValueMap() {
// 			encodeToHCL(body, k, v)
// 		}
// 	} else {

// 		fmt.Println("Expected an object or map at the top level.")
// 	}
// }

// // encodeToHCL recursively encodes cty.Value to HCL
// func encodeToHCL(body *hclwrite.Body, name string, val cty.Value) {
// 	switch val.Type() {
// 	case cty.String, cty.Number, cty.Bool:
// 		body.SetAttributeValue(name, val)
// 	default:
// 		if val.Type().IsMapType() || val.Type().IsObjectType() {
// 			block := body.AppendNewBlock(name, nil)
// 			content := block.Body()
// 			for k, v := range val.AsValueMap() {
// 				encodeToHCL(content, k, v)
// 			}
// 		} else if val.Type().IsListType() || val.Type().IsTupleType() || val.Type().IsSetType() {
// 			block := body.AppendNewBlock(name, nil)
// 			content := block.Body()
// 			for _, v := range val.AsValueSlice() {
// 				encodeToHCL(content, "item", v)
// 			}
// 		}
// 	}
// }
