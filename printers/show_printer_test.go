package printers

import (
	"fmt"
	"github.com/turbot/pipe-fittings/sanitize"
	"testing"
)

type showable2 struct {
	Name        string
	StringField string
	NumberField int
}

func (s showable2) GetShowData() *ShowData {
	return NewShowData(
		NewFieldValue("Name", s.Name, WithListKey()),
		NewFieldValue("StringField", s.StringField),
		NewFieldValue("NumberField", s.NumberField),
	)
}

type nonShowable struct {
	Name       string
	Value      string
	Other      int
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

func (s showable1) GetShowData() *ShowData {
	return NewShowData(
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
			want: "Name: test\n",
		},
		{
			name: "number slice",
			data: showable1{
				Name:                "number slice",
				PrimitiveSliceField: []int{1, 2, 3},
			},
			want: "Name: number slice\nPrimitiveSliceField:\n  1\n  2\n  3\n",
		},
		{
			name: "struct slice",
			data: showable1{
				Name:             "struct slice",
				StructSliceField: []nonShowable{{Name: "one", Value: "value 1", Other: 1, AndAnother: "and another 1"}, {Name: "two", Value: "value 2", Other: 2, AndAnother: "and another 2"}},
			},
			want: "Name: number slice\nPrimitiveSliceField:\n  1\n  2\n  3\n",
		},
		{
			name: "showable slice",
			data: showable1{
				Name:               "showable slice",
				ShowableSliceField: []showable2{{Name: "one", StringField: "value 1", NumberField: 1}, {Name: "two", StringField: "value 2", NumberField: 2}},
			},
			want: "Name: number slice\nPrimitiveSliceField:\n  1\n  2\n  3\n",
		},
		{
			name: "slice of slices",
			data: showable1{
				Name:            "slice of slices",
				SliceSliceField: [][]string{{"one", "two"}, {"three", "four"}},
			},
			want: "Name: number slice\nPrimitiveSliceField:\n  1\n  2\n  3\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewShowPrinter[showable1]()
			if err != nil {
				t.Fatalf("error creating printer: %v", err)
			}
			res, err := p.render(tt.data, sanitize.RenderOptions{ColorEnabled: false})
			if err != nil {
				t.Fatalf("error rendering: %v", err)
			}
			fmt.Println(res)
			//if res != tt.want {
			//	t.Errorf("got: %v, want: %v", res, tt.want)
			//}
		})
	}
}
