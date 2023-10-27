package hclhelpers

import (
	"testing"
)

type jsonToHclTestCase struct {
	input    string
	expected string
}

var jsonToHclTests = map[string]jsonToHclTestCase{
	"simple": {
		input:    `{ "profile": "foo"}`,
		expected: "profile = \"foo\"\n",
	},
}

func TestResolveAsString(t *testing.T) {
	for name, test := range jsonToHclTests {

		res, err := JSONToHcl(test.input)
		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED : \nunexpected error %v", name, err)
			}
			continue
		}
		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}
		if test.expected != res {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v, \ngot:\n %v\n", name, test.expected, res)
		}
	}
}
