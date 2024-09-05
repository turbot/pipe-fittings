package hclhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

type enumTest struct {
	setting  cty.Value
	enum     cty.Value
	expected bool
}

var enumTests = map[string]enumTest{
	"string": {
		setting:  cty.StringVal("foo"),
		enum:     cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
		expected: true,
	},
	"string in list": {
		setting:  cty.StringVal("foo"),
		enum:     cty.TupleVal([]cty.Value{cty.StringVal("bar"), cty.StringVal("foo")}),
		expected: true,
	},
	"string not in list": {
		setting:  cty.StringVal("foo"),
		enum:     cty.TupleVal([]cty.Value{cty.StringVal("bar"), cty.StringVal("baz")}),
		expected: false,
	},
	"number - int": {
		setting:  cty.NumberIntVal(1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberIntVal(1)}),
		expected: true,
	},
	"number - int in list": {
		setting:  cty.NumberIntVal(1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberIntVal(2), cty.NumberIntVal(1)}),
		expected: true,
	},
	"number - int not in list": {
		setting:  cty.NumberIntVal(1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberIntVal(2), cty.NumberIntVal(3)}),
		expected: false,
	},
	"number - float": {
		setting:  cty.NumberFloatVal(1.1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberFloatVal(1.1)}),
		expected: true,
	},
	"number - float in list": {
		setting:  cty.NumberFloatVal(1.1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberFloatVal(2.2), cty.NumberFloatVal(1.1)}),
		expected: true,
	},
	"number - float not in list": {
		setting:  cty.NumberFloatVal(1.1),
		enum:     cty.TupleVal([]cty.Value{cty.NumberFloatVal(2.2), cty.NumberFloatVal(3.3)}),
		expected: false,
	},
	"bool": {
		setting:  cty.BoolVal(true),
		enum:     cty.TupleVal([]cty.Value{cty.BoolVal(true)}),
		expected: true,
	},
	"bool in list": {
		setting:  cty.BoolVal(true),
		enum:     cty.TupleVal([]cty.Value{cty.BoolVal(false), cty.BoolVal(true)}),
		expected: true,
	},
	"bool not in list": {
		setting:  cty.BoolVal(true),
		enum:     cty.TupleVal([]cty.Value{cty.BoolVal(false), cty.BoolVal(false)}),
		expected: false,
	},
	"list of string in list": {
		setting:  cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		enum:     cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar"), cty.StringVal("baz")}),
		expected: true,
	},
	"list of string not in list": {
		setting:  cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		enum:     cty.TupleVal([]cty.Value{cty.StringVal("baz"), cty.StringVal("qux")}),
		expected: false,
	},
	"list of number in list": {
		setting:  cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)}),
		enum:     cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)}),
		expected: true,
	},
	"list of number not in list": {
		setting:  cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)}),
		enum:     cty.TupleVal([]cty.Value{cty.NumberIntVal(3), cty.NumberIntVal(4)}),
		expected: false,
	},
	"list of float in list": {
		setting:  cty.TupleVal([]cty.Value{cty.NumberFloatVal(1.1), cty.NumberFloatVal(2.2)}),
		enum:     cty.TupleVal([]cty.Value{cty.NumberFloatVal(1.1), cty.NumberFloatVal(2.2), cty.NumberFloatVal(3.3)}),
		expected: true,
	},
	"list of float not in list": {
		setting:  cty.TupleVal([]cty.Value{cty.NumberFloatVal(1.1), cty.NumberFloatVal(2.2)}),
		enum:     cty.TupleVal([]cty.Value{cty.NumberFloatVal(3.3), cty.NumberFloatVal(4.4)}),
		expected: false,
	},
}

func TestEnum(t *testing.T) {

	for name, test := range enumTests {

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			res, err := ValidateSettingWithEnum(test.setting, test.enum)
			if err != nil {
				assert.Fail(err.Error())
				return
			}

			assert.Equal(test.expected, res)
		})
	}
}
