package printers

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/sanitize"
	"reflect"
)

type TableRow struct {
	Columns []string
	Cells   []any
	// map of field name to field render options
	Opts map[string]FieldRenderOptions
}

func NewTableRow(fields ...FieldValue) *TableRow {
	row := &TableRow{
		Columns: make([]string, 0, len(fields)),
		Opts:    make(map[string]FieldRenderOptions),
	}
	for _, f := range fields {
		row.Columns = append(row.Columns, f.Name)

		value := f.Value
		if !helpers.IsNil(value) {
			value = dereferencePointer(value)
		}

		row.Cells = append(row.Cells, value)

		// create field render opts
		row.Opts[f.Name] = newFieldRenderOptions(f)

	}
	return row
}

func (r *TableRow) Merge(other *TableRow) {
	if other == nil {
		return
	}
	r.Columns = append(r.Columns, other.Columns...)
	r.Cells = append(r.Cells, other.Cells...)
	for k, v := range other.Opts {
		r.Opts[k] = v
	}
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
	Columns   []string
	FieldOpts map[string]FieldRenderOptions
}

func NewTable() *Table {
	return &Table{
		FieldOpts: make(map[string]FieldRenderOptions),
	}
}

func (t *Table) WithData(tableRows []TableRow, columns []string) *Table {
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

			t.Columns = append(t.Columns, f.Name)
			row.Cells = append(row.Cells, value)
		}

		// create field render opts
		t.FieldOpts[f.Name] = newFieldRenderOptions(f)
	}

	t.Rows = append(t.Rows, row)
	return t
}

func (t *Table) AddRow(row *TableRow) {
	t.Rows = append(t.Rows, *row)
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
