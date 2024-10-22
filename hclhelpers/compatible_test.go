package hclhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

type compatibleTest struct {
	ctyType  cty.Type
	value    cty.Value
	expected bool
}

var compatibleTests = map[string]compatibleTest{
	"string": {
		ctyType:  cty.String,
		value:    cty.StringVal("foo"),
		expected: true,
	},
	"string vs number": {
		ctyType:  cty.String,
		value:    cty.NumberIntVal(42),
		expected: false,
	},
	"string vs bool": {
		ctyType:  cty.String,
		value:    cty.True,
		expected: false,
	},
	"number": {
		ctyType:  cty.Number,
		value:    cty.NumberIntVal(42),
		expected: true,
	},
	"number vs string": {
		ctyType:  cty.Number,
		value:    cty.StringVal("foo"),
		expected: false,
	},
	"number vs bool": {
		ctyType:  cty.Number,
		value:    cty.True,
		expected: false,
	},
	"bool": {
		ctyType:  cty.Bool,
		value:    cty.True,
		expected: true,
	},
	"bool vs string": {
		ctyType:  cty.Bool,
		value:    cty.StringVal("foo"),
		expected: false,
	},
	"bool vs number": {
		ctyType:  cty.Bool,
		value:    cty.NumberIntVal(42),
		expected: false,
	},
	"list of string": {
		ctyType:  cty.List(cty.String),
		value:    cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		expected: true,
	},
	"list of string vs list of number": {
		ctyType:  cty.List(cty.String),
		value:    cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43)}),
		expected: false,
	},
	"list of string vs list of bool": {
		ctyType:  cty.List(cty.String),
		value:    cty.ListVal([]cty.Value{cty.True, cty.False}),
		expected: false,
	},
	"list of number": {
		ctyType:  cty.List(cty.Number),
		value:    cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43), cty.NumberIntVal(44)}),
		expected: true,
	},
	"list of number vs list of string": {
		ctyType:  cty.List(cty.Number),
		value:    cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		expected: false,
	},
	"list of number vs list of bool": {
		ctyType:  cty.List(cty.Number),
		value:    cty.ListVal([]cty.Value{cty.True, cty.False}),
		expected: false,
	},
	"list of bool": {
		ctyType:  cty.List(cty.Bool),
		value:    cty.ListVal([]cty.Value{cty.True, cty.False}),
		expected: true,
	},
	"list of bool vs list of string": {
		ctyType:  cty.List(cty.Bool),
		value:    cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		expected: false,
	},
	"list of any": {
		ctyType:  cty.List(cty.DynamicPseudoType),
		value:    cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		expected: true,
	},
	"tuple of string - matched": {
		ctyType:  cty.Tuple([]cty.Type{cty.String, cty.String}),
		value:    cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
		expected: true,
	},
	"tuple of string and bool and number": {
		ctyType:  cty.Tuple([]cty.Type{cty.String, cty.Bool, cty.Bool}),
		value:    cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.True, cty.False}),
		expected: true,
	},
	"tuple of string - not matched": {
		ctyType:  cty.Tuple([]cty.Type{cty.String}),
		value:    cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
		expected: true,
	},
	"list of any vs tuple of any mixed": {
		ctyType:  cty.List(cty.DynamicPseudoType),
		value:    cty.TupleVal([]cty.Value{cty.StringVal("foo"), cty.NumberIntVal(42)}),
		expected: true,
	},
	"list of list of string": {
		ctyType:  cty.List(cty.List(cty.String)),
		value:    cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")})}),
		expected: true,
	},
	"list of list of string vs tuple of list of string": {
		ctyType:  cty.List(cty.List(cty.String)),
		value:    cty.TupleVal([]cty.Value{cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}), cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")})}),
		expected: true,
	},
	"list of list of string vs list of list of number": {
		ctyType:  cty.List(cty.List(cty.String)),
		value:    cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43)}), cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43)})}),
		expected: false,
	},
	"list of list of string vs tuple of list of number": {
		ctyType:  cty.List(cty.List(cty.String)),
		value:    cty.TupleVal([]cty.Value{cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43)}), cty.ListVal([]cty.Value{cty.NumberIntVal(42), cty.NumberIntVal(43)})}),
		expected: false,
	},
	"map of string": {
		ctyType:  cty.Map(cty.String),
		value:    cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
		expected: true,
	},
	"map of string vs map of number": {
		ctyType:  cty.Map(cty.String),
		value:    cty.MapVal(map[string]cty.Value{"foo": cty.NumberIntVal(42), "bar": cty.NumberIntVal(43)}),
		expected: false,
	},
	"map of number": {
		ctyType:  cty.Map(cty.Number),
		value:    cty.MapVal(map[string]cty.Value{"foo": cty.NumberIntVal(42), "bar": cty.NumberIntVal(43)}),
		expected: true,
	},
	"map of number vs map of string": {
		ctyType:  cty.Map(cty.Number),
		value:    cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
		expected: false,
	},
	// this is how HCL will parse
	// default = map(
	//	"foo": "foo"
	//	"bar": "bar"
	// )
	"map of string vs object of map of string": {
		ctyType:  cty.Map(cty.String),
		value:    cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("foo"), "bar": cty.StringVal("bar")}),
		expected: true,
	},
	"map of string vs object of map of number": {
		ctyType:  cty.Map(cty.String),
		value:    cty.ObjectVal(map[string]cty.Value{"foo": cty.NumberIntVal(42), "bar": cty.NumberIntVal(43)}),
		expected: false,
	},
	"list of map of bool": {
		ctyType:  cty.List(cty.Map(cty.Bool)),
		value:    cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"foo": cty.True, "bar": cty.False})}),
		expected: true,
	},
	/*
			this is how HCL parse
			    default = [
		      {
		        "foo": true
		        "bar": false
		      },
		      {
		        "baz": true
		        "qux": false
		      }
		    ]

			as tuple of map of bool
		**/
	"list of map of bool vs list of object of map of bool": {
		ctyType:  cty.List(cty.Map(cty.Bool)),
		value:    cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"foo": cty.True, "bar": cty.False}), cty.ObjectVal(map[string]cty.Value{"bar": cty.True, "qux": cty.False})}),
		expected: true,
	},
	"list of map of bool vs tuple of map of bool": {
		ctyType:  cty.List(cty.Map(cty.Bool)),
		value:    cty.TupleVal([]cty.Value{cty.MapVal(map[string]cty.Value{"foo": cty.True, "bar": cty.False}), cty.MapVal(map[string]cty.Value{"bar": cty.True, "qux": cty.False})}),
		expected: true,
	},
	"list of map of bool vs tuple of map of number": {
		ctyType:  cty.List(cty.Map(cty.Bool)),
		value:    cty.TupleVal([]cty.Value{cty.MapVal(map[string]cty.Value{"foo": cty.NumberIntVal(42), "bar": cty.NumberIntVal(43)}), cty.MapVal(map[string]cty.Value{"bar": cty.NumberIntVal(42), "qux": cty.NumberIntVal(43)})}),
		expected: false,
	},
	"list of list of list of list of string": {
		ctyType:  cty.List(cty.List(cty.List(cty.List(cty.String)))),
		value:    cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")})})})}),
		expected: true,
	},
	"list of list of list of list of string - invalid": {
		ctyType:  cty.List(cty.List(cty.List(cty.List(cty.String)))),
		value:    cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{cty.NumberIntVal(23), cty.NumberIntVal(42)})})})}),
		expected: false,
	},
	"map of an object": {
		ctyType:  cty.Map(cty.Object(map[string]cty.Type{"foo": cty.String, "bar": cty.String})),
		value:    cty.MapVal(map[string]cty.Value{"foo": cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("foo"), "bar": cty.StringVal("bar")})}),
		expected: true,
	},
	"map of an object 2": {
		ctyType:  cty.Map(cty.Object(map[string]cty.Type{"foo": cty.String, "bar": cty.Number})),
		value:    cty.ObjectVal(map[string]cty.Value{"foo": cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("foo"), "bar": cty.NumberIntVal(42)})}),
		expected: true,
	},
}

func TestCompatible(t *testing.T) {

	for name, test := range compatibleTests {

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(test.expected, IsValueCompatibleWithType(test.ctyType, test.value))
		})
	}
}
