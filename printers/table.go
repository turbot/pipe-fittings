package printers

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/sanitize"
	"reflect"
)

type TableRow struct {
	Cells []any
	// map of field name to field render options
	Opts map[string]FieldRenderOptions
}

type FieldValue struct {
	Name  string
	Value any
	// a function implementing custom rendering logic to display the value
	RenderValueFunc func(opts sanitize.RenderOptions) string
	// a function implementing custom rendering logic to display the key AND value
	RenderKeyValueFunc func(opts sanitize.RenderOptions) string
	// the number of spaces to indent the value
	Indent int
}
type RenderFunc func(opts sanitize.RenderOptions) string
type Table struct {
	Rows      []TableRow
	Columns   []TableColumnDefinition
	FieldOpts map[string]FieldRenderOptions
}

func NewTable() *Table {
	return &Table{
		FieldOpts: make(map[string]FieldRenderOptions),
	}
}

func (t *Table) WithData(tableRows []TableRow, columns []TableColumnDefinition) *Table {
	t.Rows = tableRows
	t.Columns = columns
	return t
}
func (t *Table) WithRow(fields ...FieldValue) *Table {
	row := TableRow{}
	for _, f := range fields {

		value := f.Value
		if !helpers.IsNil(value) {
			value = dereferencePointer(value)

			t.Columns = append(t.Columns, TableColumnDefinition{
				Name: f.Name,
			})
			row.Cells = append(row.Cells, value)
		}

		// create field render opts
		t.FieldOpts[f.Name] = newFieldRenderOptions(f)
	}

	t.Rows = append(t.Rows, row)
	return t
}

func dereferencePointer(value any) any {
	val := reflect.ValueOf(value)

	// Check if the value is a pointer
	if val.Kind() == reflect.Ptr {
		// Dereference the pointer and update the value
		value = val.Elem().Interface()

	}
	return value
}
