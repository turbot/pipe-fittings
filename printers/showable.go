package printers

import (
	"github.com/turbot/go-kit/helpers"
	"golang.org/x/exp/maps"
	"strings"
)

// Showable is an interface implemented by objects which support the `show` command for pretty/plain output format
type Showable interface {
	GetShowData() *RowData
}

func IsShowable(value any) bool {
	return AsShowable(value) != nil
}

func AsShowable(value any) Showable {
	if s, ok := value.(Showable); ok {
		return s
	}
	// check the pointer
	pVal := &value
	if s, ok := any(pVal).(Showable); ok {
		return s
	}
	return nil
}

// Listable is an interface implemented by objects which support the `list` command for pretty/plain output format
type Listable interface {
	GetListData() *RowData
}

// RowData is a struct that holds the data for a single row of a table
type RowData struct {
	// slice of column names (converted to lower case)
	Columns []string
	// map of column name to field display name (as provided in NewRowData)
	displayNameMap map[string]string
	// map of fields keyed by lower case column name
	Fields map[string]FieldValue
}

func NewRowData(fields ...FieldValue) *RowData {
	data := &RowData{
		Fields:         make(map[string]FieldValue),
		displayNameMap: make(map[string]string),
		Columns:        make([]string, 0, len(fields)),
	}
	for _, f := range fields {
		data.AddField(f)
	}
	return data
}

// Merge merges the other RowData
// columns from other are placed first, however our fields take precedence
// this is to ensure when merging base showdata the base columns come first, but derived typeds can override values
func (d *RowData) Merge(other *RowData) {
	if other == nil {
		return
	}

	// combine columns, putting other first, omitting dupes
	d.Columns = helpers.AppendSliceUnique(other.Columns, d.Columns)

	// merge Fields from other, retaining our value in the case of conflict
	otherClone := maps.Clone(other.Fields)
	maps.Copy(otherClone, d.Fields)
	d.Fields = otherClone

	// merge display name map - we do not expect conflict here
	maps.Copy(d.displayNameMap, other.displayNameMap)
	d.Fields = otherClone
}

func (d *RowData) AddField(f FieldValue) {
	// convert name into lower case for column name - we store the original name in the displayNameMap

	columnName := strings.ToLower(f.Name)
	d.Columns = append(d.Columns, columnName)
	d.displayNameMap[columnName] = f.Name
	d.Fields[columnName] = f
}

func (d *RowData) GetRow() *TableRow {
	row := NewTableRow()
	row.Columns = d.Columns
	for _, c := range d.Columns {
		row.Cells = append(row.Cells, d.Fields[c].Value)
	}
	return row
}
