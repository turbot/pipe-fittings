package backend

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/sperr"
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

func FromConnectionString(ctx context.Context, str string) (Backend, error) {
	switch {
	case IsPostgresConnectionString(str):
		pgBackend, err := NewPostgresBackend(ctx, str)
		if err != nil {
			return nil, err
		}
		// check if this is in fact a steampipe backend
		if isSteampipeBackend(ctx, pgBackend) {
			return NewSteampipeBackend(ctx, *pgBackend)
		}
		return pgBackend, nil

	case IsMySqlConnectionString(str):
		return NewMySQLBackend(str), nil
	case IsDuckDBConnectionString(str):
		return NewDuckDBBackend(str), nil
	case IsSqliteConnectionString(str):
		return NewSqliteBackend(str), nil
	default:
		return nil, sperr.WrapWithMessage(ErrUnknownBackend, "could not evaluate backend: '%s'", str)
	}
}

func HasBackend(str string) bool {
	switch {
	case
		IsPostgresConnectionString(str),
		IsMySqlConnectionString(str),
		IsDuckDBConnectionString(str),
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

// IsMySqlConnectionString returns true if the connection string is for mysql
// looks for the mysql:// prefix
func IsMySqlConnectionString(connString string) bool {
	return strings.HasPrefix(connString, mysqlConnectionStringPrefix)
}
