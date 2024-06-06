package queryresult

import "reflect"

// ColumnDef is a struct used to store column information from query results
type ColumnDef struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	UniqueName string `json:"unique_name,omitempty"`
	isScalar   *bool
}

// IsScalar checks if the given value is a scalar value
// it also mutates the containing ColumnDef so that it doesn't have to reflect
// for all values in a column
func (c *ColumnDef) IsScalar(v any) bool {
	if c.isScalar == nil {
		var scalar bool
		switch reflect.ValueOf(v).Kind() {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			scalar = false
		default:
			scalar = true
		}
		c.isScalar = &scalar
	}
	return *c.isScalar
}

// GetUniqueName returns the unique name of the column
func (c *ColumnDef) GetUniqueName() string {
	// if UniqueName is set, that indicates the original column name exists more than once in a row
	// so we have generated a unique name for it. Return this unique name
	if c.UniqueName != "" {
		return c.UniqueName
	}
	// otherwise return the original column name
	return c.Name
}
