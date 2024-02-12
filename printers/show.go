package printers

import "strings"

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
	// slice of column names (converted to lower case)
	Columns []string
	// map of column name to field display name (as provided in NewShowData)
	displayNameMap map[string]string
	Fields         map[string]FieldValue
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

func (d *ShowData) Merge(other *ShowData) {
	if other == nil {
		return
	}
	d.Columns = append(d.Columns, other.Columns...)
	for k, v := range other.Fields {
		d.Fields[k] = v
	}
}

func (d *ShowData) GetRow() *TableRow {
	row := NewTableRow()
	for _, c := range d.Columns {
		row.Cells = append(row.Cells, d.Fields[c].Value)
	}
	return row
}
