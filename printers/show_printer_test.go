package printers

import (
	"github.com/turbot/pipe-fittings/v2/sanitize"
	"github.com/turbot/pipe-fittings/v2/utils"
	"testing"
)

type showable2 struct {
	Name             string
	StringField      string
	NumberPtrField   *int
	StructSliceField []nonShowable
}

func (s showable2) GetShowData() *RowData {
	return NewRowData(
		NewFieldValue("Name", s.Name, WithListKey()),
		NewFieldValue("StringField", s.StringField),
		NewFieldValue("NumberPtrField", s.NumberPtrField),
		NewFieldValue("StructSliceField", s.StructSliceField),
	)
}

type nonShowable struct {
	Name       string
	Value      string
	Other      *int
	AndAnother string
	MapField   map[string]string
}

// some test data types
type showable1 struct {
	Name                string
	ShowableField       showable2
	ShowablePtrField    *showable2
	StructField         nonShowable
	StructPtrField      *nonShowable
	StringField         string
	StringPtrField      *string
	NumberField         int
	NumberPtrField      *int
	MapField            map[string]string
	MapOfMapsField      map[string]map[string]string
	PrimitiveSliceField []int
	StructSliceField    []nonShowable
	SliceSliceField     [][]string
	ShowableSliceField  []showable2
	MapSliceField       []map[string]string
}

func (s showable1) GetShowData() *RowData {
	return NewRowData(
		NewFieldValue("Name", s.Name, WithListKey()),
		NewFieldValue("ShowableField", s.ShowableField),
		NewFieldValue("ShowablePtrField", s.ShowablePtrField),
		NewFieldValue("StructField", s.StructField),
		NewFieldValue("StructPtrField", s.StructPtrField),
		NewFieldValue("StringField", s.StringField),
		NewFieldValue("StringPtrField", s.StringPtrField),
		NewFieldValue("NumberField", s.NumberField),
		NewFieldValue("NumberPtrField", s.NumberPtrField),
		NewFieldValue("MapField", s.MapField),
		NewFieldValue("MapOfMapsField", s.MapOfMapsField),
		NewFieldValue("PrimitiveSliceField", s.PrimitiveSliceField),
		NewFieldValue("StructSliceField", s.StructSliceField),
		NewFieldValue("SliceSliceField", s.SliceSliceField),
		NewFieldValue("ShowableSliceField", s.ShowableSliceField),
		NewFieldValue("MapSliceField", s.MapSliceField),
	)
}

func TestShowPrinter(t *testing.T) {
	type testCase[T showable1] struct {
		name string
		data Showable
		want string
	}
	tests := []testCase[showable1]{
		{
			name: "simple",
			data: showable1{
				Name: "simple",
			},
			want: "Name:                simple\nNumberField:         0\n",
		},
		{
			name: "number slice",
			data: showable1{
				Name:                "number slice",
				PrimitiveSliceField: []int{1, 2, 3},
			},
			want: "Name:                number slice\nNumberField:         0\nPrimitiveSliceField: \n  1\n  2\n  3\n",
		},
		{
			name: "struct slice",
			data: showable1{
				Name:             "struct slice",
				StructSliceField: []nonShowable{{Name: "one", Value: "value 1", Other: utils.ToIntegerPointer(1), AndAnother: "and another 1"}, {Name: "two", Value: "value 2", Other: utils.ToIntegerPointer(2), AndAnother: "and another 2"}},
			},
			want: "Name:                struct slice\nNumberField:         0\nStructSliceField:    \n\n    AndAnother: and another 1\n    Name:       one\n    Other:      1\n    Value:      value 1\n\n    AndAnother: and another 2\n    Name:       two\n    Other:      2\n    Value:      value 2\n",
		},
		{
			name: "showable slice",
			data: showable1{
				Name:               "showable slice",
				ShowableSliceField: []showable2{{Name: "one", StringField: "value 1", NumberPtrField: utils.ToIntegerPointer(1)}, {Name: "two", StringField: "value 2", NumberPtrField: utils.ToIntegerPointer(2)}},
			},
			want: "Name:                showable slice\nNumberField:         0\nShowableSliceField:  \n  Name:               one\n    StringField:        value 1\n    NumberPtrField:     1\n  Name:               two\n    StringField:        value 2\n    NumberPtrField:     2\n",
		},
		{
			name: "slice of slices",
			data: showable1{
				Name:            "slice of slices",
				SliceSliceField: [][]string{{"one", "two"}, {"three", "four"}},
			},
			want: "Name:                slice of slices\nNumberField:         0\nSliceSliceField:     \n    one\n    two\n    three\n    four\n",
		},
		{
			name: "showable field",
			data: showable1{
				Name:          "showable field",
				ShowableField: showable2{Name: "one", StringField: "value 1", NumberPtrField: utils.ToIntegerPointer(1)},
			},
			want: "Name:                showable field\nShowableField:       \n  Name:             one\n  StringField:      value 1\n  NumberPtrField:   1\nNumberField:         0\n",
		},
		{
			name: "struct field",
			data: showable1{
				Name:        "struct field",
				StructField: nonShowable{Name: "one", Value: "value 1", Other: utils.ToIntegerPointer(1), AndAnother: "and another 1"},
			},
			want: "Name:                struct field\nStructField:         \n  AndAnother: and another 1\n  Name:       one\n  Other:      1\n  Value:      value 1\nNumberField:         0\n",
		},
	}

	testFilter := ""
	for _, tt := range tests {
		if len(testFilter) > 0 && tt.name != testFilter {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewShowPrinter[showable1]()
			if err != nil {
				t.Fatalf("error creating printer: %v", err)
			}
			res, err := p.render(tt.data, sanitize.RenderOptions{ColorEnabled: false})
			if err != nil {
				t.Fatalf("error rendering: %v", err)
			}
			//fmt.Println(res)
			if res != tt.want {
				t.Errorf("got: %v, want: %v", res, tt.want)
			}
		})
	}
}
