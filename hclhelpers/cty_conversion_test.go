package hclhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
