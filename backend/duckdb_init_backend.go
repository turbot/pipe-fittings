package backend

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

const (
	duckDBInitConnectionStringPrefix = "duckdbinit:"
)

type DuckDBInitBackend struct {
	initScript string
	rowReader  RowReader
}

func NewDuckDBInitBackend(connString string) (*DuckDBInitBackend, error) {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	// connString is already the file path, no need to trim prefix
	return &DuckDBInitBackend{
		initScript: connString,
		rowReader:  newDuckDBRowReader(),
	}, nil
}

// Connect implements Backend.
func (b *DuckDBInitBackend) Connect(ctx context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	db, err := sql.Open("duckdb", fmt.Sprintf("init=%s", b.initScript))
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	// for duckdb, limit connections to 1 - DuckDB is designed for single-connection usage
	db.SetMaxOpenConns(1)

	// Install and load the standard extensions
	err = installAndLoadDuckDbExtensions(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (b *DuckDBInitBackend) ConnectionString() string {
	return b.initScript
}

func (b *DuckDBInitBackend) Name() string {
	return constants.DuckDBInitBackendName
}

// RowReader implements Backend.
func (b *DuckDBInitBackend) RowReader() RowReader {
	return b.rowReader
}
