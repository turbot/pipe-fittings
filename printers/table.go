package printers

import (
	"github.com/turbot/go-kit/helpers"
	"reflect"
)

type TableRow struct {
	Cells []any
}

type FieldValue struct {
	Name  string
	Value any
	// TODO render opts
}
type Table struct {
	Rows    []TableRow
	Columns []TableColumnDefinition
}

func NewTable() *Table {
	return &Table{}
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
			// add an empty row if there are no rows
			if len(t.Rows) == 0 {
				t.Rows = append(t.Rows, TableRow{})
			}
			t.Rows[0].Cells = append(t.Rows[0].Cells, value)
		}
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
