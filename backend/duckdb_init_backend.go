package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

const (
	DuckDBInitConnectionStringPrefix = "duckdbinit:"
)

type DuckDBInitBackend struct {
	initScript string
	rowReader  RowReader
}

func NewDuckDBInitBackend(connString string) (*DuckDBInitBackend, error) {
	// remove any leading or trailing whitespace
	connString = strings.TrimSpace(connString)
	// remove the prefix
	connString = strings.TrimPrefix(connString, DuckDBInitConnectionStringPrefix)
	return &DuckDBInitBackend{
		initScript: connString,
		rowReader:  newDuckDBRowReader(),
	}, nil
}

// Close attempts to remove the init script file if it exists
func (b *DuckDBInitBackend) Close()error {
	if b.initScript != "" {
		if _, err := os.Stat(b.initScript); err == nil {
			// file exists - try to remove it
			if err := os.Remove(b.initScript); err != nil {
				// just log error
				slog.Warn("Failed to remove duckdb init script file", "file", b.initScript, "error", err)
			}
		}
	}
	return nil
}

// Connect implements Backend.
func (b *DuckDBInitBackend) Connect(ctx context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	db, err := sql.Open("duckdb", "")
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

	// Execute the init script if provided
	if b.initScript != "" {
		err = b.executeInitScript(ctx, db)
		if err != nil {
			db.Close()
			return nil, sperr.WrapWithMessage(err, "failed to execute init script")
		}
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

// executeInitScript reads and executes the SQL init script file
func (b *DuckDBInitBackend) executeInitScript(ctx context.Context, db *sql.DB) error {
	content, err := os.ReadFile(b.initScript)
	if err != nil {
		return fmt.Errorf("failed to read init script %q: %w", b.initScript, err)
	}

	if _, err := db.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute init script %q: %w", b.initScript, err)
	}
	return nil
}
