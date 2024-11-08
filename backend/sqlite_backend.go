package backend

import (
	"context"
	"database/sql"
	"strings"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/sperr"
)

const (
	sqliteConnectionStringPrefix = "sqlite:"
)

type SqliteBackend struct {
	connectionString string
	rowReader        RowReader
}

func NewSqliteBackend(connString string) *SqliteBackend {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	connString = strings.TrimPrefix(connString, sqliteConnectionStringPrefix)
	return &SqliteBackend{
		connectionString: connString,
		rowReader:        newSqliteRowReader(),
	}
}

// Connect implements Backend.
func (b *SqliteBackend) Connect(_ context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)
	db, err := sql.Open("sqlite3", b.connectionString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to sqlite backend")
	}
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)
	return db, nil
}

func (b *SqliteBackend) ConnectionString() string {
	return b.connectionString
}

func (b *SqliteBackend) Name() string {
	return constants.SQLiteBackendName
}

// RowReader implements Backend.
func (b *SqliteBackend) RowReader() RowReader {
	return b.rowReader
}

type sqliteRowReader struct {
	BasicRowReader
}

func newSqliteRowReader() *sqliteRowReader {
	return &sqliteRowReader{
		// use the generic row reader - there's no real difference between sqlite and generic
		BasicRowReader: *NewBasicRowReader(),
	}
}
