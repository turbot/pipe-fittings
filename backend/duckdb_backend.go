package backend

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

const (
	DuckDBConnectionStringPrefix = "duckdb://"
)

type DuckDBBackend struct {
	connectionString string
	rowReader        RowReader
}

func NewDuckDBBackend(connString string) (*DuckDBBackend, error) {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	connString = strings.TrimPrefix(connString, DuckDBConnectionStringPrefix)
	return &DuckDBBackend{
		connectionString: connString,
		rowReader:        newDuckDBRowReader(),
	}, nil
}

// Connect implements Backend.
func (b *DuckDBBackend) Connect(ctx context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	if permitted_paths.Enabled() {
		// open locked to the permitted folders; the extension setup runs inside the lock's init
		return connectDuckDBPermitted(ctx, b.connectionString, func(db *sql.DB) {
			db.SetConnMaxIdleTime(config.MaxConnIdleTime)
			db.SetConnMaxLifetime(config.MaxConnLifeTime)
			db.SetMaxOpenConns(config.MaxOpenConns)
		})
	}

	db, err := sql.Open("duckdb", b.connectionString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to duckdb backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)

	// Install and load the standard extensions
	err = installAndLoadDuckDbExtensions(db)
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
	return b.rowReader
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

// installAndLoadDuckDbExtensions installs and loads the default DuckDB extensions (json, inet)
func installAndLoadDuckDbExtensions(db *sql.DB) error {
	return duckDbExtensionSetup(func(query string) error {
		_, err := db.Exec(query)
		return err
	})
}

// duckDbExtensionSetup sets the extension directory and installs and loads the default DuckDB extensions,
// running each statement through exec. It is a variable so tests, which have no network, can replace it.
var duckDbExtensionSetup = func(exec func(query string) error) error {
	// set the extension directory
	if err := exec(fmt.Sprintf("set extension_directory = '%s';", filepaths.EnsurePipesDuckDbExtensionsDir())); err != nil {
		return fmt.Errorf("failed to set extension_directory: %w", err)
	}

	// install and load the extensions
	for _, extension := range constants.DuckDbExtensions {
		if err := exec(fmt.Sprintf("install '%s'; load '%s';", extension, extension)); err != nil {
			return fmt.Errorf("failed to install and load extension %s: %s", extension, err.Error())
		}
	}

	return nil
}
