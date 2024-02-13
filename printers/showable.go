package printers

import (
	"github.com/turbot/go-kit/helpers"
	"golang.org/x/exp/maps"
	"strings"
)

// Showable is an interface implemented by objects which support the `show` command for pretty/plain output format
type Showable interface {
	GetShowData() *ShowData
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
	GetListData() *ShowData
}

func AsListable(value any) Listable {
	if l, ok := value.(Listable); ok {
		return l
	}
	// check the pointer
	pVal := &value
	if l, ok := any(pVal).(Listable); ok {
		return l
	}
	return nil
}

type ShowData struct {
	ListKeyField *ListKeyField
	// slice of column names (converted to lower case)
	Columns []string
	// map of column name to field display name (as provided in NewShowData)
	displayNameMap map[string]string
	Fields         map[string]FieldValue
}

// ListKeyField is a struct that holds the name and value of a field to be used as the key (grouping) field when
// displaying a slice of resource
type ListKeyField struct {
	Name            string
	RenderValueFunc RenderFunc
}

func NewShowData(fields ...FieldValue) *ShowData {
	data := &ShowData{
		Fields:         make(map[string]FieldValue),
		displayNameMap: make(map[string]string),
		Columns:        make([]string, 0, len(fields)),
	}
	for _, f := range fields {
		// convert name into lower case for column name - we store the original name in the displayNameMap
		columnName := strings.ToLower(f.Name)
		data.Columns = append(data.Columns, columnName)
		data.displayNameMap[columnName] = f.Name
		data.Fields[columnName] = f
	}
	return data
}

// Merge merges the other ShowData
// columns from other are placed first, however our fields take precedence
// this is to ensure when merging base showdata the base columns come first, but derived typeds can override values
func (d *ShowData) Merge(other *ShowData) {
	if other == nil {
		return
	}

	// combine columns, putting other first, omitting dupes
	d.Columns = helpers.AppendUnique(other.Columns, d.Columns)

	// merge Fields from other, retaining our value in the case of conflict
	otherClone := maps.Clone(other.Fields)
	maps.Copy(otherClone, d.Fields)
	d.Fields = otherClone

	// merge display name map - we do not expect conflict here
	maps.Copy(d.displayNameMap, other.displayNameMap)
	d.Fields = otherClone
}

func (d *ShowData) GetRow() *TableRow {
	row := NewTableRow()
	for _, c := range d.Columns {
		row.Cells = append(row.Cells, d.Fields[c].Value)
	}
	return row
}
