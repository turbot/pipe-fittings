package backend

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/sperr"
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
func (b *DuckDBBackend) Connect(ctx context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	db, err := sql.Open("duckdb", b.connectionString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)

	// Install and load the JSON extension
	err = installAndLoadExtensions(db)
	if err != nil {
		return nil, err
	}

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

func installAndLoadExtensions(db *sql.DB) error {
	// set the extension directory
	if _, err := db.Exec(fmt.Sprintf("SET extension_directory = '%s';", filepaths.EnsurePipesDuckDbExtensionsDir())); err != nil {
		return fmt.Errorf("failed to set extension_directory: %w", err)
	}

	// install and load the extensions
	for _, extension := range constants.DuckDbExtensions {
		if _, err := db.Exec(fmt.Sprintf("INSTALL '%s'; LOAD '%s';", extension, extension)); err != nil {
			return fmt.Errorf("failed to install and load extension %s: %s", extension, err.Error())
		}
	}

	return nil
}
