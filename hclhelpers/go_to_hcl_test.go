package hclhelpers

import (
	"testing"
)

type goToHclTestCase struct {
	input    any
	expected string
}

var goToHclTestCases = map[string]goToHclTestCase{
	"simple map": {
		input:    map[string]any{"profile": "foo"},
		expected: "{profile = \"foo\"}",
	},
	"just string": {
		input:    "foo",
		expected: `"foo"`,
	},
	"list of string": {
		input:    []string{"foo", "bar"},
		expected: `["foo", "bar"]`,
	},
	"list of numbers": {
		input:    []any{1, 2, 2.3, 3.2, -1, 0},
		expected: `[1, 2, 2.3, 3.2, -1, 0]`,
	},
	"list of any": {
		input:    []any{"foo", 1, 2.3},
		expected: `["foo", 1, 2.3]`,
	},
	"map of string": {
		input:    map[string]string{"profile": "foo", "region": "us-west-1"},
		expected: "{profile = \"foo\", region = \"us-west-1\"}",
	},
	"map of numbers": {
		input:    map[string]any{"profile": 1, "region": 2.3},
		expected: "{profile = 1, region = 2.3}",
	},
	"list of map": {
		input:    []any{map[string]any{"profile": "foo"}, map[string]any{"region": "us-west-1"}},
		expected: "[{profile = \"foo\"}, {region = \"us-west-1\"}]",
	},
	"list of list": {
		input:    []any{[]any{"foo", "bar"}, []any{"us-west-1", "us-west-2"}},
		expected: "[[\"foo\", \"bar\"], [\"us-west-1\", \"us-west-2\"]]",
	},
	"map of any": {
		input:    map[string]any{"profile": "foo", "region": 2.3},
		expected: "{profile = \"foo\", region = 2.3}",
	},
	"map of any complex": {
		input:    map[string]any{"profile": "foo", "region": []any{"us-west-1", "us-west-2"}},
		expected: "{profile = \"foo\", region = [\"us-west-1\", \"us-west-2\"]}",
	},
	"list of complex": {
		input:    []any{map[string]any{"title": nil, "profile": "foo", "region": "us-west-1", "notifiers": []any{map[string]any{"name": "foo", "age": 23}}}, map[string]any{"profile": "bar", "region": "us-west-2"}},
		expected: `[{notifiers = [{age = 23, name = "foo"}], profile = "foo", region = "us-west-1", title = null}, {profile = "bar", region = "us-west-2"}]`,
	},
	"with null": {
		input:    map[string]any{"profile": "foo", "region": nil},
		expected: "{profile = \"foo\", region = null}",
	},
}

func TestGoToHcl(t *testing.T) {
	for name, test := range goToHclTestCases {

		t.Run(name, func(t *testing.T) {
			res, err := GoToHCLString(test.input)
			if err != nil {
				if test.expected != "ERROR" {
					t.Errorf("Test: '%s'' FAILED : \nunexpected error %v", name, err)
				}
				return
			}
			if test.expected == "ERROR" {
				t.Errorf("Test: '%s'' FAILED - expected error", name)
				return
			}
			if test.expected != res {
				t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, res)
			}
		})
	}
}
