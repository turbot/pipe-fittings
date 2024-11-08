package backend

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/constants"

	"github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/sperr"
)

const (
	mysqlConnectionStringPrefix = "mysql://"
)

type MySQLBackend struct {
	connectionString string
	rowreader        RowReader
}

func NewMySQLBackend(connString string) *MySQLBackend {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	connString = strings.TrimPrefix(connString, mysqlConnectionStringPrefix)

	return &MySQLBackend{
		connectionString: connString,
		rowreader:        newMySqlRowReader(),
	}
}

// Connect implements Backend.
func (b *MySQLBackend) Connect(_ context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	db, err := sql.Open("mysql", b.connectionString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to mysql backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)
	return db, nil
}

func (b *MySQLBackend) ConnectionString() string {
	return b.connectionString
}

func (b *MySQLBackend) Name() string {
	return constants.MySQLBackendName
}

// RowReader implements Backend.
func (b *MySQLBackend) RowReader() RowReader {
	return b.rowreader
}

type mysqlRowReader struct {
	BasicRowReader
}

func newMySqlRowReader() RowReader {
	return &mysqlRowReader{
		BasicRowReader: BasicRowReader{
			CellReader: mysqlReadCell,
		},
	}
}

func mysqlReadCell(columnValue any, col *queryresult.ColumnDef) (result any, err error) {
	if columnValue != nil {
		switch t := columnValue.(type) {
		case []byte:
			asStr := string(t)
			switch col.DataType {
			case "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "YEAR":
				result, err = strconv.ParseInt(asStr, 10, 64)
			case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
				result, err = strconv.ParseFloat(asStr, 64)
			case "DATE":
				result, err = time.Parse(time.DateOnly, asStr)
			case "TIME":
				result, err = time.Parse(time.TimeOnly, asStr)
			case "DATETIME", "TIMESTAMP":
				result, err = time.Parse(time.DateTime, asStr)
			case "BIT":
				result = columnValue.([]byte)
			// case "CHAR", "VARCHAR", "TEXT", "BINARY", "VARBINARY", "ENUM", "SET":
			default:
				result = asStr
			}
		default:
			return columnValue, nil
		}
	}
	return result, err
}
