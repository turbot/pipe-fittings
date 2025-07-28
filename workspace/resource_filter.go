package workspace

import (
	"fmt"
	"log"
	"net/url"
	"slices"
	"strings"

	"github.com/danwakefield/fnmatch"
	"github.com/turbot/pipe-fittings/v2/filter"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/printers"
	"github.com/turbot/pipe-fittings/v2/sperr"
	"golang.org/x/exp/maps"
)

type ResourceFilter struct {
	Where          string
	Tags           map[string][]string
	WherePredicate func(item modconfig.HclResource) bool
}

// ResourceFilterFromTags creates a ResourceFilter from a list of tag values of the form 'key=value'
func ResourceFilterFromTags(tags []string) ResourceFilter {
	var res = ResourceFilter{
		Tags: make(map[string][]string),
	}

	// 'tags' should be KV Pairs of the form: 'benchmark=pic' or 'cis_level=1'
	for _, tag := range tags {
		value, _ := url.ParseQuery(tag)
		for k, v := range value {
			if _, ok := res.Tags[k]; !ok {
				res.Tags[k] = []string{}
			}
			res.Tags[k] = append(res.Tags[k], v...)
		}
	}
	return res
}

func (f *ResourceFilter) Empty() bool {
	return f.Where == "" && len(f.Tags) == 0
}

func (f *ResourceFilter) getPredicate() (func(resource modconfig.HclResource) bool, error) {
	// if a where predicate has been provided just use that
	if f.WherePredicate != nil {
		if f.Tags != nil || f.Where != "" {
			return nil, sperr.New("cannot specify 'where' or 'tags' when 'wherePredicate' is provided")
		}
		return f.WherePredicate, nil
	}
	// If there is a 'where' clause, parse it
	wherePredicate, err := f.parseFilter()
	if err != nil {
		return nil, err
	}
	tagPredicate := f.getTagPredicate()

	// combine these
	res := func(resource modconfig.HclResource) bool {
		return wherePredicate(resource) && tagPredicate(resource)
	}

	return res, nil
}

func (f *ResourceFilter) getTagPredicate() func(resource modconfig.HclResource) bool {
	if f.Tags == nil {
		return func(resource modconfig.HclResource) bool {
			return true
		}
	}
	tagPredicate := func(resource modconfig.HclResource) bool {
		tags := resource.GetTags()
		for k, v := range f.Tags {
			if !slices.Contains(v, tags[k]) {
				return false
			}
		}
		return true

	}
	return tagPredicate
}

func (f *ResourceFilter) parseFilter() (func(resource modconfig.HclResource) bool, error) {
	if f.Where == "" {
		return func(resource modconfig.HclResource) bool {
			return true
		}, nil
	}

	// Use the existing filter parser for all expressions, including JSON path expressions
	parsed, err := filter.Parse("", []byte(f.Where))
	if err != nil {
		log.Printf("err %v", err)
		return nil, sperr.New("failed to parse 'where' property: %s", err.Error())
	}

	// convert table schema into a column map
	columnFilter, err := newColumnFilter(parsed.(filter.ComparisonNode))
	if err != nil {
		return nil, err
	}

	// now build the predicate
	p := func(resource modconfig.HclResource) bool {
		data := resource.GetShowData()

		// Check if the column exists (for non-JSON path expressions)
		if !columnFilter.isJSONPath {
			if _, containsColumn := data.Fields[columnFilter.column]; !containsColumn {
				return false
			}
		}

		return columnFilter.evaluate(data)
	}
	return p, nil
}

type columnFilter struct {
	column     string
	operator   string
	values     []string
	isJSONPath bool
	jsonPath   string // Base field name for JSON path expressions
	jsonKey    string // Key to extract from JSON field
}

func newColumnFilter(cn filter.ComparisonNode) (columnFilter, error) {
	var res columnFilter

	switch cn.Type {

	case "compare", "like":
		codeNodes, ok := cn.Values.([]filter.CodeNode)
		if !ok {
			return res, fmt.Errorf("failed to parse cn")
		}
		if len(codeNodes) != 2 {
			return res, fmt.Errorf("failed to parse cn")
		}

		res.column = codeNodes[0].Value
		res.values = append(res.values, codeNodes[1].Value)
		res.operator = cn.Operator.Value

		// Check if this is a JSON path expression
		if len(codeNodes[0].JsonbSelector) > 0 {
			res.isJSONPath = true
			res.jsonPath = codeNodes[0].Value
			// Extract the key from the JSON selector
			if len(codeNodes[0].JsonbSelector) >= 2 {
				res.jsonKey = codeNodes[0].JsonbSelector[1].Value
			}
		}

	case "in":
		res.operator = cn.Operator.Value

		codeNodes, ok := cn.Values.([]filter.CodeNode)
		if !ok || len(codeNodes) < 2 {
			return res, fmt.Errorf("failed to parse cn")
		}
		res.column = codeNodes[0].Value

		// Check if this is a JSON path expression
		if len(codeNodes[0].JsonbSelector) > 0 {
			res.isJSONPath = true
			res.jsonPath = codeNodes[0].Value
			// Extract the key from the JSON selector
			if len(codeNodes[0].JsonbSelector) >= 2 {
				res.jsonKey = codeNodes[0].JsonbSelector[1].Value
			}
		}

		// Build look up of values to dedupe
		valuesMap := make(map[string]struct{}, len(codeNodes)-1)
		for _, c := range codeNodes[1:] {
			valuesMap[c.Value] = struct{}{}
		}
		res.values = maps.Keys(valuesMap)

	default:
		return res, fmt.Errorf("failed to convert 'where' arg to qual")
	}

	return res, nil
}

// evaluateFilter evaluates whether the f.column filter passes for the given resource
func (f columnFilter) evaluate(data *printers.RowData) bool {
	// Get the value to compare against
	var valueToCompare string
	if f.isJSONPath {
		valueToCompare = f.getJSONPathValue(data)
	} else {
		if field, exists := data.Fields[f.column]; exists {
			valueToCompare = field.ValueString()
		} else {
			return false
		}
	}

	// Apply the operator
	switch f.operator {
	case "=":
		return valueToCompare == f.values[0]
	case "!=":
		return valueToCompare != f.values[0]
	// TODO cast as number??
	//case "<":
	//case "<=":
	//case ">":
	//case ">=":
	case "~~", "like":
		return SqlLike(valueToCompare, f.values[0], true)
	case "!~~", "not like":
		return !SqlLike(valueToCompare, f.values[0], true)
	case "~~*", "ilike":
		return SqlLike(valueToCompare, f.values[0], false)
	case "!~~*", "not ilike":
		return !SqlLike(valueToCompare, f.values[0], false)
	case "in":
		for _, v := range f.values {
			if valueToCompare == v {
				return true
			}
		}
		return false
	case "not in":
		for _, v := range f.values {
			if valueToCompare == v {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// getJSONPathValue extracts the value from a JSON path expression
func (f columnFilter) getJSONPathValue(data *printers.RowData) string {
	fieldValue, exists := data.Fields[f.jsonPath]
	if !exists {
		return ""
	}

	// If it's a map (like tags), extract the key value
	if tags, ok := fieldValue.Value.(map[string]string); ok {
		if tagValue, tagExists := tags[f.jsonKey]; tagExists {
			return tagValue
		}
	}

	return ""
}

// SqlLike simulates SQL LIKE pattern matching using fnmatch, with an option for case sensitivity.
func SqlLike(input, pattern string, caseSensitive bool) bool {
	flag := 0
	if !caseSensitive {
		flag = fnmatch.FNM_CASEFOLD
	}
	// convert he sql pattern to fnmatch pattern
	fnmatchPattern := sqlLikeToFnmatch(pattern)
	return fnmatch.Match(fnmatchPattern, input, flag)

}

// sqlLikeToFnmatch converts a SQL LIKE pattern to an fnmatch pattern
func sqlLikeToFnmatch(pattern string) string {
	// Replace SQL '%' wildcard with fnmatch '*' wildcard
	pattern = strings.ReplaceAll(pattern, "%", "*")

	// Replace SQL '_' wildcard with fnmatch '?' wildcard
	pattern = strings.ReplaceAll(pattern, "_", "?")

	// Handle escaped '%' and '_' characters
	// This example assumes '\' is used as the escape character in the SQL pattern
	pattern = strings.ReplaceAll(pattern, "\\%", "%")
	pattern = strings.ReplaceAll(pattern, "\\_", "_")

	return pattern
}
