package querydisplay

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/queryresult"

	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/constants"
)

// columnNames builds a list of name from a slice of column defs - respecting the original name if present
func columnNames(columns []*queryresult.ColumnDef) []string {
	var colNames = make([]string, len(columns))
	for i, c := range columns {
		// respect original name
		if c.OriginalName != "" {
			colNames[i] = c.OriginalName
		} else {
			colNames[i] = c.Name
		}
	}

	return colNames
}

type columnValueSettings struct{ nullString string }

type ColumnValueOption func(opt *columnValueSettings)

func WithNullString(nullString string) ColumnValueOption {
	return func(opt *columnValueSettings) {
		opt.nullString = nullString
	}
}

// ColumnValuesAsString converts a slice of columns into strings
func ColumnValuesAsString(values []interface{}, columns []*queryresult.ColumnDef, opts ...ColumnValueOption) ([]string, error) {
	rowAsString := make([]string, len(columns))
	for idx, val := range values {
		val, err := ColumnValueAsString(val, columns[idx], opts...)
		if err != nil {
			return nil, err
		}
		rowAsString[idx] = val
	}
	return rowAsString, nil
}

// ColumnValueAsString converts column value to string
func ColumnValueAsString(val interface{}, col *queryresult.ColumnDef, opts ...ColumnValueOption) (result string, err error) {
	opt := &columnValueSettings{nullString: constants.NullString}
	for _, o := range opts {
		o(opt)
	}

	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("%v", val)
		}
	}()

	if val == nil {
		return opt.nullString, nil
	}

	//log.Printf("[TRACE] ColumnValueAsString type %s", colType.DatabaseTypeName())
	// possible types for colType are defined in pq/oid/types.go
	switch col.DataType {
	case "JSON", "JSONB":
		bytes, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	case "TIMESTAMP", "DATE", "TIME", "INTERVAL":
		t, ok := val.(time.Time)
		if ok {
			return t.Format("2006-01-02 15:04:05"), nil
		}
		fallthrough
	case "NAME":
		result := string(val.([]uint8))
		return result, nil
	case "UUID":
		// duckdb returns UUID as []uint8 which if parsed as string will return illegible data, need to convert to uuid
		// postgres returns UUID correctly and doesn't need conversion
		b, ok := val.([]uint8)
		if ok {
			v, err := uuid.FromBytes(b)
			if err != nil {
				return "", err
			}
			return v.String(), nil
		}
		fallthrough
	default:
		if strings.HasPrefix(col.DataType, "DECIMAL") {
			// attempt to convert decimal to string, if this fails will fall through to generic formatting code
			if str, ok := columnValueForDuckDBDecimal(val); ok {
				return str, nil
			}
		}

		return typeHelpers.ToString(val), nil
	}
}

// ParseJSONOutputColumnValue segregate data types, ignore string conversion for certain data types :
// JSON, JSONB, BOOL and so on..
func ParseJSONOutputColumnValue(val interface{}, col *queryresult.ColumnDef) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch col.DataType {
	// we can revise/increment the list of DT's in future
	case "JSON", "JSONB", "BOOL", "INT2", "INT4", "INT8", "FLOAT8", "FLOAT4":
		return val, nil
	default:
		return ColumnValueAsString(val, col)
	}
}

// columnValueForDuckDBDecimal converts duckdb decimal to string
// we want to use the String() method on the duckDb Decimal object,
// but we do not want to reference go-duckdb as that requires Cgo bindings
// so instead invoke the function via reflection
func columnValueForDuckDBDecimal(val interface{}) (string, bool) {
	if val == nil {
		return "", false
	}

	s, err := helpers.ExecuteMethod(val, "String")
	if err == nil && len(s) == 1 {
		if str, ok := s[0].(string); ok {
			return str, true
		}
	}

	return "", false
}
