package backend

import (
	"context"
	"database/sql"
	"github.com/turbot/pipe-fittings/v2/constants"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

const (
	duckDBConnectionStringPrefix = "duckdb:"
)

type DuckDBBackend struct {
	connectionString string
	rowreader        RowReader
}

func NewDuckDBBackend(connString string) *DuckDBBackend {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	connString = strings.TrimPrefix(connString, duckDBConnectionStringPrefix)
	return &DuckDBBackend{
		connectionString: connString,
		rowreader:        newDuckDBRowReader(),
	}
}

// Connect implements Backend.
func (b *DuckDBBackend) Connect(_ context.Context, options ...ConnectOption) (*sql.DB, error) {
	config := NewConnectConfig(options)
	db, err := sql.Open("duckdb", b.connectionString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)
	return db, nil
}

func (b *DuckDBBackend) ConnectionString() string {
	return b.connectionString
}

func (b *DuckDBBackend) Name() string {
	return constants.DuckDBBackendName
}

// RowReader implements Backend.
func (b *DuckDBBackend) RowReader() RowReader {
	return b.rowreader
}

type duckdbRowReader struct {
	BasicRowReader
}

func newDuckDBRowReader() *duckdbRowReader {
	return &duckdbRowReader{
		// use the generic row reader - there's no real difference between sqlite and duckdb
		BasicRowReader: *NewBasicRowReader(),
	}
}
