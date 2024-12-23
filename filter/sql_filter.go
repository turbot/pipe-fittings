package filter

import (
	"fmt"
	"github.com/danwakefield/fnmatch"
	"github.com/turbot/steampipe-plugin-sdk/v5/filter"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"golang.org/x/exp/maps"
	"log"
	"strings"
)

type SqlFilter struct {
	Filter filter.ComparisonNode
	raw    string
}

func NewSqlFilter(raw string) (*SqlFilter, error) {
	parsed, err := filter.Parse("", []byte(raw))
	if err != nil {
		log.Printf("err %v", err)
		return nil, sperr.New("failed to parse 'where' property: %s", err.Error())
	}

	res := &SqlFilter{
		Filter: parsed.(filter.ComparisonNode),
		raw:    raw,
	}

	// do a test run of the filter to ensure all operators are supported
	if _, _, err := sqlFilterSatisfied(res.Filter, map[string]string{}); err != nil {
		return nil, err
	}

	return res, nil

}

func (f *SqlFilter) Satisfied(values map[string]string) bool {
	res, _, _ := sqlFilterSatisfied(f.Filter, values)
	return res
}

func (f *SqlFilter) GetFieldNames() ([]string, error) {
	_, fieldNames, err := sqlFilterSatisfied(f.Filter, map[string]string{})
	return fieldNames, err
}

// sqlFilterSatisfied evaluates a filter against a set of values.
// It returns whether the filter is satisfied as well as the field names which were checked
func sqlFilterSatisfied(c filter.ComparisonNode, values map[string]string) (bool, []string, error) {
	var fieldNames = map[string]struct{}{}
	switch c.Type {
	case "identifier":
		// not sure when this would be used
		return false, nil, invalidScopeOperatorError(c.Operator.Value)
	case "is":
		// 'is' is not (currently) supported
		return false, nil, invalidScopeOperatorError(c.Operator.Value)
	case "like": // (also ilike?)
		codeNodes, ok := c.Values.([]filter.CodeNode)
		if !ok {
			return false, nil, fmt.Errorf("failed to parse filter")
		}
		if len(codeNodes) != 2 {
			return false, nil, fmt.Errorf("failed to parse filter")
		}

		fieldName := codeNodes[0].Value
		fieldNames[fieldName] = struct{}{}
		lval := values[fieldName]
		pattern := codeNodes[1].Value

		// Evaluate the condition
		var res bool
		switch c.Operator.Value {
		case "like":
			res = evaluateLike(lval, pattern, 0)
		case "not like":
			res = !evaluateLike(lval, pattern, 0)
		case "ilike":
			res = evaluateLike(lval, pattern, fnmatch.FNM_IGNORECASE)
		case "not ilike":
			res = !evaluateLike(lval, pattern, fnmatch.FNM_IGNORECASE)
		default:
			return false, nil, invalidScopeOperatorError(c.Operator.Value)
		}
		return res, maps.Keys(fieldNames), nil

	case "compare":
		codeNodes, ok := c.Values.([]filter.CodeNode)
		if !ok {
			return false, nil, fmt.Errorf("failed to parse filter")
		}
		if len(codeNodes) != 2 {
			return false, nil, fmt.Errorf("failed to parse filter")
		}

		fieldName := codeNodes[0].Value
		fieldNames[fieldName] = struct{}{}
		lval := values[fieldName]
		rval := codeNodes[1].Value

		var res bool
		switch c.Operator.Value {
		case "=":
			res = lval == rval
		case "!=", "<>":
			res = lval != rval
		// as we (currently) only support string scopes, < and > are not supported
		case "<=", ">=", "<", ">":
			return false, nil, invalidScopeOperatorError(c.Operator.Value)
		}
		return res, maps.Keys(fieldNames), nil

	case "in":
		codeNodes, ok := c.Values.([]filter.CodeNode)
		if !ok {
			return false, nil, fmt.Errorf("failed to parse filter")
		}
		if len(codeNodes) < 2 {
			return false, nil, fmt.Errorf("failed to parse filter")
		}

		fieldName := codeNodes[0].Value
		fieldNames[fieldName] = struct{}{}

		// Build a lookup of possible values
		rvals := make(map[string]struct{}, len(codeNodes)-1)
		for _, c := range codeNodes[1:] {
			rvals[c.Value] = struct{}{}
		}

		lval := values[fieldName]

		// Check if the value exists in rvals
		_, rvalsContainValue := rvals[lval]

		// Operator determines the result
		var res bool
		switch c.Operator.Value {
		case "in":
			res = rvalsContainValue
		case "not in":
			res = !rvalsContainValue
		}
		return res, maps.Keys(fieldNames), nil

	case "not":
		// TODO have not identified queries which give a top-level 'not'
		return false, nil, fmt.Errorf("unsupported location for 'not' operator")

	case "or":
		nodes, ok := c.Values.([]any)
		if !ok {
			return false, nil, fmt.Errorf("failed to parse filter")
		}
		satisfied := false
		for _, n := range nodes {
			c, ok := n.(filter.ComparisonNode)
			if !ok {
				return false, nil, fmt.Errorf("failed to parse filter")
			}
			// If any child nodes are satisfied, return true
			childSatisfied, childFieldNames, err := sqlFilterSatisfied(c, values)
			if err != nil {
				return false, nil, err
			}
			for _, c := range childFieldNames {
				fieldNames[c] = struct{}{}
			}

			if childSatisfied {
				satisfied = true
			}
		}
		// No child nodes satisfied - return false
		return satisfied, maps.Keys(fieldNames), nil

	case "and":
		nodes, ok := c.Values.([]any)
		if !ok {
			return false, nil, fmt.Errorf("failed to parse filter")
		}
		satisfied := true
		for _, n := range nodes {
			c, ok := n.(filter.ComparisonNode)
			if !ok {
				return false, nil, fmt.Errorf("failed to parse filter")
			}
			// If any child nodes are unsatisfied, return false
			childSatisfied, childFieldNames, err := sqlFilterSatisfied(c, values)
			if err != nil {
				return false, nil, err
			}
			for _, c := range childFieldNames {
				fieldNames[c] = struct{}{}
			}
			if !childSatisfied {
				satisfied = false
			}
		}
		// All child nodes satisfied - return true
		return satisfied, maps.Keys(fieldNames), nil
	}

	return false, nil, fmt.Errorf("failed to parse filter")
}

func evaluateLike(val, pattern string, flag int) bool {
	pattern = strings.ReplaceAll(pattern, "_", "?")
	pattern = strings.ReplaceAll(pattern, "%", "*")
	return fnmatch.Match(pattern, val, flag)

}

func invalidScopeOperatorError(operator string) error {
	return fmt.Errorf("invalid scope filter operator '%s'", operator)
}
