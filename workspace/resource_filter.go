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

	// Check if this is a JSON path expression (contains ->)
	if strings.Contains(f.Where, "->") {
		return f.parseJSONPathFilter()
	}

	// Try to parse as a simple property filter first
	if simpleFilter, err := f.parseSimplePropertyFilter(); err == nil {
		return simpleFilter, nil
	}

	// Fall back to the original filter parser
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

		if _, containsColumn := data.Fields[columnFilter.column]; !containsColumn {
			return false
		}

		return columnFilter.evaluate(data)
	}
	return p, nil
}

// parseSimplePropertyFilter handles simple property filters like cis_type='automated'
func (f *ResourceFilter) parseSimplePropertyFilter() (func(resource modconfig.HclResource) bool, error) {
	// Parse expressions like: cis_type='automated' or severity='high'
	// or: cis_type in ('automated', 'manual')
	parts := strings.Fields(f.Where)
	if len(parts) < 3 {
		return nil, sperr.New("invalid simple property filter: %s", f.Where)
	}

	propertyName := parts[0] // e.g., "tag_property"
	operator := parts[1]     // e.g., "=" or "in"

	// Handle "not in" operator
	if operator == "not" && len(parts) >= 4 && parts[2] == "in" {
		operator = "not in"
		parts = append(parts[:2], parts[3:]...)
	}

	// Extract values
	var values []string
	if operator == "in" || operator == "not in" {
		// Handle list like ('automated', 'manual')
		valuePart := strings.Join(parts[2:], " ")
		if strings.HasPrefix(valuePart, "(") && strings.HasSuffix(valuePart, ")") {
			valuePart = strings.Trim(valuePart, "()")
			values = parseQuotedList(valuePart)
		} else {
			return nil, sperr.New("invalid list format in filter: %s", f.Where)
		}
	} else {
		// Handle single value
		value := strings.Trim(parts[2], "'")
		values = []string{value}
	}

	// Build the predicate
	p := func(resource modconfig.HclResource) bool {
		data := resource.GetShowData()

		// Get the field value
		fieldValue, exists := data.Fields[propertyName]
		if !exists {
			return false
		}

		// Compare the values
		fieldStr := fieldValue.ValueString()
		switch operator {
		case "=":
			if len(values) == 1 {
				return fieldStr == values[0]
			}
			return false
		case "!=":
			if len(values) == 1 {
				return fieldStr != values[0]
			}
			return false
		case "in":
			for _, v := range values {
				if fieldStr == v {
					return true
				}
			}
			return false
		case "not in":
			for _, v := range values {
				if fieldStr == v {
					return false
				}
			}
			return true
		default:
			return false
		}
	}

	return p, nil
}

// parseJSONPathFilter handles PostgreSQL JSON path expressions like tags->>'service'
func (f *ResourceFilter) parseJSONPathFilter() (func(resource modconfig.HclResource) bool, error) {
	// Parse expressions like: tags->>'service' != 'Azure/ActiveDirectory'
	// or: tags->>'service' not in ('Azure/EntraID','Azure/ActiveDirectory')
	// or: tags->>'service' = 'Azure/ActiveDirectory'

	// Extract the JSON path and operator
	parts := strings.Fields(f.Where)
	if len(parts) < 3 {
		return nil, sperr.New("invalid JSON path expression: %s", f.Where)
	}

	jsonPath := parts[0] // e.g., "tags->>'tag_property'"
	operator := parts[1] // e.g., "!=", "=", "in", "not"

	// Handle "not in" operator
	if operator == "not" && len(parts) >= 4 && parts[2] == "in" {
		operator = "not in"
		parts = append(parts[:2], parts[3:]...)
	}

	// Extract values (handle both single value and list)
	var values []string
	if operator == "in" || operator == "not in" {
		valuePart := strings.Join(parts[2:], " ")
		if strings.HasPrefix(valuePart, "(") && strings.HasSuffix(valuePart, ")") {
			valuePart = strings.Trim(valuePart, "()")
			values = parseQuotedList(valuePart)
		} else {
			return nil, sperr.New("invalid list format in filter: %s", f.Where)
		}
	} else {
		valuePart := strings.Join(parts[2:], " ")
		if strings.HasPrefix(valuePart, "'") && strings.HasSuffix(valuePart, "'") {
			values = []string{strings.Trim(valuePart, "'")}
		} else {
			values = []string{valuePart}
		}
	}

	// Parse the JSON path
	pathParts := strings.Split(jsonPath, "->")
	if len(pathParts) != 2 {
		return nil, sperr.New("invalid JSON path: %s", jsonPath)
	}

	fieldName := strings.TrimSpace(pathParts[0])                   // e.g., "tags"
	keyName := strings.Trim(strings.TrimSpace(pathParts[1]), ">'") // e.g., "tag_property"

	// Build the predicate
	p := func(resource modconfig.HclResource) bool {
		data := resource.GetShowData()

		// Get the field value
		fieldValue, exists := data.Fields[fieldName]
		if !exists {
			return false
		}

		// If it's a map (like tags), extract the key value
		if tags, ok := fieldValue.Value.(map[string]string); ok {
			tagValue, tagExists := tags[keyName]
			if !tagExists {
				return false
			}

			// Apply the operator
			switch operator {
			case "=":
				if len(values) == 1 {
					return tagValue == values[0]
				}
				return false
			case "!=":
				if len(values) == 1 {
					return tagValue != values[0]
				}
				return false
			case "in":
				for _, v := range values {
					if tagValue == v {
						return true
					}
				}
				return false
			case "not in":
				for _, v := range values {
					if tagValue == v {
						return false
					}
				}
				return true
			default:
				return false
			}
		}

		return false
	}

	return p, nil
}

// parseQuotedList parses a comma-separated list of quoted strings
func parseQuotedList(listStr string) []string {
	var values []string
	parts := strings.Split(listStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			values = append(values, strings.Trim(part, "'"))
		}
	}
	return values
}

type columnFilter struct {
	column   string
	operator string
	values   []string
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

	case "in":
		res.operator = cn.Operator.Value

		codeNodes, ok := cn.Values.([]filter.CodeNode)
		if !ok || len(codeNodes) < 2 {
			return res, fmt.Errorf("failed to parse cn")
		}
		res.column = codeNodes[0].Value

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
	switch f.operator {
	case "=":
		return data.Fields[f.column].ValueString() == f.values[0]
	case "!=":
		return data.Fields[f.column].ValueString() != f.values[0]
	// TODO cast as number??
	//case "<":
	//case "<=":
	//case ">":
	//case ">=":
	case "~~", "like":
		return SqlLike(data.Fields[f.column].ValueString(), f.values[0], true)
	case "!~~", "not like":
		return !SqlLike(data.Fields[f.column].ValueString(), f.values[0], true)
	case "~~*", "ilike":
		return SqlLike(data.Fields[f.column].ValueString(), f.values[0], false)
	case "!~~*", "not ilike":
		return !SqlLike(data.Fields[f.column].ValueString(), f.values[0], false)
	case "in":
		for _, v := range f.values {
			if data.Fields[f.column].ValueString() == v {
				return true
			}
		}
		return false
	case "not in":
		for _, v := range f.values {
			if data.Fields[f.column].ValueString() == v {
				return false
			}
		}
		return true
	default:
		return false
	}
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
