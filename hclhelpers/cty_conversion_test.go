package hclhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestConvertInterfaceToCtyValue(t *testing.T) {
	assert := assert.New(t)

	cty, err := ConvertInterfaceToCtyValue("foo")
	assert.Nil(err)

	assert.Equal("foo", cty.AsString())

	cty, err = ConvertInterfaceToCtyValue(map[string]interface{}{
		"foo":  "bar",
		"baz":  "qux",
		"quux": "baz",
	})

	assert.Nil(err)
	assert.Equal("bar", cty.GetAttr("foo").AsString())
	assert.Equal("qux", cty.GetAttr("baz").AsString())
	assert.Equal("baz", cty.GetAttr("quux").AsString())

	cty, err = ConvertInterfaceToCtyValue([]interface{}{"foo", "bar", "baz", 3})
	assert.Nil(err)

	ctySlice := cty.AsValueSlice()
	assert.Equal(4, len(ctySlice))
	assert.Equal("foo", ctySlice[0].AsString())
	assert.Equal("bar", ctySlice[1].AsString())
	assert.Equal("baz", ctySlice[2].AsString())
	val, _ := ctySlice[3].AsBigFloat().Float64()
	assert.Equal(float64(3), val)

	cty, err = ConvertInterfaceToCtyValue([]string{"foo", "bar", "baz"})
	assert.Nil(err)

	ctySlice = cty.AsValueSlice()
	assert.Equal(3, len(ctySlice))
	assert.Equal("foo", ctySlice[0].AsString())
	assert.Equal("bar", ctySlice[1].AsString())
	assert.Equal("baz", ctySlice[2].AsString())

	complexMap := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
			"man": "chu",
			"baz": map[string]interface{}{
				"qux": "quux",
				"quux": []interface{}{
					"foo",
					40,
					"baz",
				},
			},
			"quux": []interface{}{
				"foo",
				27,
				"baz",
			},
			"qux": []string{
				"foo",
				"bar",
				"baz",
			},
			"baz_baz": "qux",
			"bar_bar": 3,
			"foo_bar": []int{
				1,
				2,
				3,
			},
		},
		"bar": []interface{}{
			"foo",
			"bar",
			"baz",
			5,
		},
	}

	cty, err = ConvertInterfaceToCtyValue(complexMap)
	assert.Nil(err)

	ctyMap := cty.AsValueMap()
	assert.Equal("baz", ctyMap["foo"].GetAttr("bar").AsString())

	ctySlice = ctyMap["foo"].GetAttr("quux").AsValueSlice()
	assert.Equal(3, len(ctySlice))
	assert.Equal("foo", ctySlice[0].AsString())

	val, _ = ctySlice[1].AsBigFloat().Float64()
	assert.Equal(float64(27), val)

	assert.Equal("baz", ctySlice[2].AsString())

	ctyMapNested := ctyMap["foo"].GetAttr("baz").AsValueMap()
	assert.Equal("quux", ctyMapNested["qux"].AsString())

	ctySlice = ctyMapNested["quux"].AsValueSlice()
	assert.Equal(3, len(ctySlice))
	assert.Equal("foo", ctySlice[0].AsString())
	val, _ = ctySlice[1].AsBigFloat().Float64()
	assert.Equal(float64(40), val)
	assert.Equal("baz", ctySlice[2].AsString())
}

func TestConvertInterfaceToCtyValue2(t *testing.T) {
	assert := assert.New(t)

	stringMap := map[string]string{
		"foo":  "bar",
		"baz":  "qux",
		"quux": "baz",
	}

	cty, err := ConvertInterfaceToCtyValue(stringMap)
	assert.Nil(err)

	assert.Equal("bar", cty.GetAttr("foo").AsString())
	assert.Equal("qux", cty.GetAttr("baz").AsString())
	assert.Equal("baz", cty.GetAttr("quux").AsString())

	intMap := map[string]int{
		"foo":  1,
		"baz":  2,
		"quux": 4,
	}

	cty, err = ConvertInterfaceToCtyValue(intMap)
	assert.Nil(err)

	val, _ := cty.GetAttr("foo").AsBigFloat().Float64()
	assert.Equal(float64(1), val)
	val, _ = cty.GetAttr("baz").AsBigFloat().Float64()
	assert.Equal(float64(2), val)
	val, _ = cty.GetAttr("quux").AsBigFloat().Float64()
	assert.Equal(float64(4), val)

	boolMap := map[string]bool{
		"foo":  true,
		"baz":  false,
		"quux": true,
	}

	cty, err = ConvertInterfaceToCtyValue(boolMap)
	assert.Nil(err)

	assert.Equal(true, cty.GetAttr("foo").True())
	assert.Equal(false, cty.GetAttr("baz").True())
	assert.Equal(true, cty.GetAttr("quux").True())
}

func TestConvertInterfaceToCtyValueWithStruct(t *testing.T) {
	assert := assert.New(t)

	type Foo struct {
		Bar string
		Baz string
	}

	foo := Foo{
		Bar: "bar",
		Baz: "baz",
	}

	cty, err := ConvertInterfaceToCtyValue(foo)
	assert.Nil(err)

	assert.Equal("bar", cty.GetAttr("Bar").AsString())
	assert.Equal("baz", cty.GetAttr("Baz").AsString())
}

type coerceValueTest struct {
	title    string
	input    string
	expected interface{}
	ctyType  cty.Type
}

var coerceValueTests = []coerceValueTest{
	{
		title:    "string",
		input:    "foo",
		expected: "foo",
		ctyType:  cty.String,
	},
	{
		// This is a bit of a weird test, but it's to ensure that we can handle
		// this use case: --arg 'region="us-east-2"'
		//
		// intuitively we'd expect the value to be us-east-2, not literal "us-east-2"
		//
		// this is why we need to strip the quotes if they are present in the beginning AND and the
		// end of the string
		title:    "string with quotes",
		input:    "\"foo\"",
		expected: "foo",
		ctyType:  cty.String,
	},
	{
		title:    "string with quotes 2",
		input:    "\"foo bar\"",
		expected: "foo bar",
		ctyType:  cty.String,
	},
	{
		title:    "string with quotes 3",
		input:    "\"foo bar baz\"\"",
		expected: "foo bar baz\"",
		ctyType:  cty.String,
	},
	{
		title:    "string with quotes - unbalanced",
		input:    "\"foo",
		expected: "\"foo",
		ctyType:  cty.String,
	},
	{
		title:    "string with quotes - unbalanced 2",
		input:    "foo\"",
		expected: "foo\"",
		ctyType:  cty.String,
	},
	{
		title:    "string with quotes - unbalanced 3",
		input:    "\"\"foo",
		expected: "\"\"foo",
		ctyType:  cty.String,
	},
	{
		title:    "bool",
		input:    "true",
		expected: true,
		ctyType:  cty.Bool,
	},
	{
		title:    "int",
		input:    "3",
		expected: 3,
		ctyType:  cty.Number,
	},
	{
		title:    "float",
		input:    "3.14",
		expected: 3.14,
		ctyType:  cty.Number,
	},
}

func TestCoerceValue(tm *testing.T) {
	for _, tc := range coerceValueTests {
		tm.Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)

			ctyValue, err := CoerceStringToGoBasedOnCtyType(tc.input, tc.ctyType)
			if err != nil {
				assert.Fail(err.Error())
				return
			}

			assert.Equal(tc.expected, ctyValue)
		})
	}
}

type ctyTypeToHclTypeTest struct {
	input    cty.Type
	expected string
}

var ctyTypeToHclTypeTests = map[string]ctyTypeToHclTypeTest{
	"dynamic pseudo type": {
		input:    cty.DynamicPseudoType, // this comes as cty.NilType so the underlying type is unknown
		expected: "",
	},
	"empty object": {
		input: cty.EmptyObject,
		// make sure there are 2 spaces after the open bracket
		expected: `{
  
}`,
	},
	"empty tuple": {
		input:    cty.EmptyTuple,
		expected: "tuple([])",
	},
	"simple string": {
		input:    cty.String,
		expected: "string",
	},
	"simple bool": {
		input:    cty.Bool,
		expected: "bool",
	},
	"simple number": {
		input:    cty.Number,
		expected: "number",
	},
	"list of string": {
		input:    cty.List(cty.String),
		expected: "list(string)",
	},
	"list of number": {
		input:    cty.List(cty.Number),
		expected: "list(number)",
	},
	"list of bool": {
		input:    cty.List(cty.Bool),
		expected: "list(bool)",
	},
	"list of list of string": {
		input:    cty.List(cty.List(cty.String)),
		expected: "list(list(string))",
	},
	"map of string": {
		input:    cty.Map(cty.String),
		expected: "map(string)",
	},
	"map of number": {
		input:    cty.Map(cty.Number),
		expected: "map(number)",
	},
	"map of bool": {
		input:    cty.Map(cty.Bool),
		expected: "map(bool)",
	},
	"map of map of string": {
		input:    cty.Map(cty.Map(cty.String)),
		expected: "map(map(string))",
	},
	"map of a list of string": {
		input:    cty.Map(cty.List(cty.String)),
		expected: "map(list(string))",
	},
	"map of a list of number": {
		input:    cty.Map(cty.List(cty.Number)),
		expected: "map(list(number))",
	},
	"map of a list of bool": {
		input:    cty.Map(cty.List(cty.Bool)),
		expected: "map(list(bool))",
	},
	"map of a list of a map of a list of a bool": {
		input:    cty.Map(cty.List(cty.Map(cty.List(cty.Bool)))),
		expected: "map(list(map(list(bool))))",
	},
	"list of a list of a list of a map of a list of a number": {
		input:    cty.List(cty.List(cty.List(cty.Map(cty.List(cty.Number))))),
		expected: "list(list(list(map(list(number)))))",
	},
	"object": {
		input: cty.Object(map[string]cty.Type{"foo": cty.String, "bar": cty.Number}),
		expected: `{
  bar = number
  foo = string
}`,
	},
	"object with list": {
		input: cty.Object(map[string]cty.Type{"foo": cty.String, "bar": cty.List(cty.Number)}),
		expected: `{
  bar = list(number)
  foo = string
}`,
	},
}

func TestCtyTypeToHclType(t *testing.T) {

	for name, test := range ctyTypeToHclTypeTests {

		t.Run(name, func(t *testing.T) {
			res := CtyTypeToHclType(test.input)
			if test.expected != res {
				t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, res)
			}
		})
	}
}
