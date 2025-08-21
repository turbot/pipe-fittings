package backend

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

var ErrUnknownBackend = errors.New("unknown backend")

type RowReader interface {
	Read(columnValues []any, cols []*queryresult.ColumnDef) ([]any, error)
}

type Backend interface {
	Connect(context.Context, ...BackendOption) (*sql.DB, error)
	RowReader() RowReader
	ConnectionString() string
	Name() string
}

type SearchPathProvider interface {
	OriginalSearchPath() []string
	RequiredSearchPath() []string
	ResolvedSearchPath() []string
}

// NameFromConnectionString returns the name of the backend from the connection string
// NOTE: this function is does not create the backend so cannot check if a postgres the backend is a steampipe backend
func NameFromConnectionString(ctx context.Context, cs string) (string, error) {
	switch {
	case IsPostgresConnectionString(cs):
		// NOTE: this _may be_ a steampipe backend but this function does not check as we want it to be inexpensive
		return constants.PostgresBackendName, nil
	case IsMySqlConnectionString(cs):
		return constants.MySQLBackendName, nil
	case IsDuckDBConnectionString(cs):
		return constants.DuckDBBackendName, nil
	case IsSqliteConnectionString(cs):
		return constants.SQLiteBackendName, nil
	default:
		return "", sperr.WrapWithMessage(ErrUnknownBackend, "could not evaluate backend: '%s'", cs)
	}
}

// FromConnectionString creates a backend from a connection string
func FromConnectionString(ctx context.Context, cs string) (Backend, error) {
	switch {
	case IsPostgresConnectionString(cs):
		pgBackend, err := NewPostgresBackend(ctx, cs)
		if err != nil {
			return nil, err
		}
		// check if this is in fact a steampipe backend
		if isSteampipeBackend(ctx, pgBackend) {
			return NewSteampipeBackend(ctx, *pgBackend)
		}
		return pgBackend, nil

	case IsMySqlConnectionString(cs):
		return NewMySQLBackend(cs)
	case IsDuckDBConnectionString(cs):
		return NewDuckDBBackend(cs)
	case IsDucklakeConnectionString(cs):
		return NewDucklakeBackend(cs)
	case IsSqliteConnectionString(cs):
		return NewSqliteBackend(cs)
	default:
		return nil, sperr.WrapWithMessage(ErrUnknownBackend, "could not evaluate backend: '%s'", cs)
	}
}

func HasBackend(str string) bool {
	switch {
	case
		IsPostgresConnectionString(str),
		IsMySqlConnectionString(str),
		IsDuckDBConnectionString(str),
		IsDucklakeConnectionString(str),
		IsSqliteConnectionString(str):
		return true
	default:

		return false
	}
}
func isSteampipeBackend(ctx context.Context, s *PostgresBackend) bool {
	db, err := s.Connect(ctx)
	if err != nil {
		return false
	}
	defer db.Close()

	// Query to check if tables exist
	query := `SELECT EXISTS (
                  SELECT FROM 
                      pg_tables
                  WHERE 
                      schemaname = 'steampipe_internal' AND 
                      tablename  IN ('steampipe_plugin', 'steampipe_connection')
              );`

	// Execute the query
	var exists bool
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		return false
	}

	// Check if tables exist
	return exists
}

// IsPostgresConnectionString returns true if the connection string is for postgres
// looks for the postgresql:// or postgres:// prefix
func IsPostgresConnectionString(connString string) bool {
	for _, v := range postgresConnectionStringPrefixes {
		if strings.HasPrefix(connString, v) {
			return true
		}
	}
	return false
}

// IsSqliteConnectionString returns true if the connection string is for sqlite
// looks for the sqlite:// prefix
func IsSqliteConnectionString(connString string) bool {
	return strings.HasPrefix(connString, sqliteConnectionStringPrefix)
}

// IsDuckDBConnectionString returns true if the connection string is for duckdb
// looks for the duckdb:// prefix
func IsDuckDBConnectionString(connString string) bool {
	return strings.HasPrefix(connString, duckDBConnectionStringPrefix)
}

// IsDucklakeConnectionString returns true if the connection string is for ducklake
// looks for the ducklake:// prefix
func IsDucklakeConnectionString(connString string) bool {
	return strings.HasPrefix(connString, ducklakeConnectionStringPrefix)
}

// IsMySqlConnectionString returns true if the connection string is for mysql
// looks for the mysql:// prefix
func IsMySqlConnectionString(connString string) bool {
	return strings.HasPrefix(connString, mysqlConnectionStringPrefix)
}
