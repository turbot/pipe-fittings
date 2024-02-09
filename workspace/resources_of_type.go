package workspace

import (
	"fmt"
	"github.com/danwakefield/fnmatch"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/printers"
	filter2 "github.com/turbot/steampipe-plugin-sdk/v5/filter"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"golang.org/x/exp/maps"
	"log"
)

// GetWorkspaceResourcesOfType returns all resources of type T from a workspace
func GetWorkspaceResourcesOfType[T modconfig.HclResource](w *Workspace) map[string]T {
	var res = map[string]T{}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		if item, ok := item.(T); ok {
			res[item.Name()] = item
		}
		return true, nil
	}

	// resource func does not return error
	_ = w.GetResourceMaps().WalkResources(resourceFunc)

	return res
}

// FilterWorkspaceResourcesOfType returns all resources of type T from a workspace which satisf          y the filter,
// which is specified as a SQL syntax where clause
func FilterWorkspaceResourcesOfType[T modconfig.HclResource](w *Workspace, where string) (map[string]T, error) {
	var res = map[string]T{}

	filterPredicate, err := parseFilter(where)
	if err != nil {
		return nil, err
	}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// if item is correct type and matches the predicate, add it to the result
		if item, ok := item.(T); ok && filterPredicate(item) {
			res[item.Name()] = item
		}
		return true, nil
	}

	// resource func does not return error
	_ = w.GetResourceMaps().WalkResources(resourceFunc)

	return res, nil
}

func parseFilter(raw string) (func(printers.Showable) bool, error) {
	parsed, err := filter2.Parse("", []byte(raw))
	if err != nil {
		log.Printf("err %v", err)
		return nil, sperr.New("failed to parse 'where' property: %s", err.Error())
	}

	// convert table schema into a column map

	filter := parsed.(filter2.ComparisonNode)

	var column string
	var values []string
	var operator string

	switch filter.Type {

	case "compare", "like":
		codeNodes, ok := filter.Values.([]filter2.CodeNode)
		if !ok {
			return nil, fmt.Errorf("failed to parse filter")
		}
		if len(codeNodes) != 2 {
			return nil, fmt.Errorf("failed to parse filter")
		}

		column = codeNodes[0].Value
		values = append(values, codeNodes[1].Value)
		operator = filter.Operator.Value

	case "in":
		operator = filter.Operator.Value

		codeNodes, ok := filter.Values.([]filter2.CodeNode)
		if !ok || len(codeNodes) < 2 {
			return nil, fmt.Errorf("failed to parse filter")
		}
		column = codeNodes[0].Value

		// Build look up of values to dedupe
		valuesMap := make(map[string]struct{}, len(codeNodes)-1)
		for _, c := range codeNodes[1:] {
			valuesMap[c.Value] = struct{}{}
		}
		values = maps.Keys(valuesMap)

	default:
		return nil, fmt.Errorf("failed to convert 'where' arg to qual")

	}

	// now build the predicate
	p := func(item printers.Showable) bool {
		data := item.GetShowData()

		if _, containsColumn := data.Fields[column]; !containsColumn {
			return false
		}

		return evaluateFilter(data, column, operator, values)
	}

	return p, nil
}

func evaluateFilter(data *printers.ShowData, column string, operator string, values []string) bool {
	switch operator {
	case "=":
		return data.Fields[column].ValueString() == values[0]
	case "!=":
		return data.Fields[column].ValueString() != values[0]
	// TODO cast as number??
	//case "<":
	//case "<=":
	//case ">":
	//case ">=":
	case "~~", "like":
		return sqlLike(data.Fields[column].ValueString(), values[0], true)
	case "!~~", "not like":
		return !sqlLike(data.Fields[column].ValueString(), values[0], true)
	case "~~*":
		return sqlLike(data.Fields[column].ValueString(), values[0], false)
	case "!~~*":
		return !sqlLike(data.Fields[column].ValueString(), values[0], false)
	case "in":
		for _, v := range values {
			if data.Fields[column].ValueString() == v {
				return true
			}
		}
		return false
	case "not in":
		for _, v := range values {
			if data.Fields[column].ValueString() == v {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Simulates SQL LIKE pattern matching in Go, with an option for case sensitivity.
func sqlLike(input, pattern string, caseSensitive bool) bool {
	flag := 0
	if !caseSensitive {
		flag = fnmatch.FNM_CASEFOLD
	}
	return fnmatch.Match(pattern, input, flag)

}

//
//func main() {
//	tests := []struct {
//		input          string
//		pattern        string
//		caseSensitive  bool
//		expectedMatch  bool
//	}{
//		{"Hello, world!", "hello, %", false, true}, // Case-insensitive match
//		{"Hello, world!", "hello, %", true, false},  // Case-sensitive match fails
//		{"Test_string_123", "Test\_string\__%", false, true},
//		{"abcdefg", "abc\_efg", true, false},
//		{"abc_efg", "ABC\_EFG", false, true},        // Case-insensitive match
//	}
//
//	for _, test := range tests {
//		match := sqlLike(test.input, test.pattern, test.caseSensitive)
//		fmt.Printf("Input: '%s', Pattern: '%s', Case Sensitive: %t, Expected: %t, Got: %t\n",
//			test.input, test.pattern, test.caseSensitive, test.expectedMatch, match)
//	}
//}
