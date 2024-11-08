package backend

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/sperr"
	"github.com/turbot/pipe-fittings/utils"
)

var postgresConnectionStringPrefixes = []string{"postgresql://", "postgres://"}

type PostgresBackend struct {
	originalConnectionString string
	originalSearchPath       []string
	schemaNames              []string
	rowReader                RowReader
	// if a custom search path or a prefix is used, store the resolved search path
	// NOTE: only applies to postgres backend
	requiredSearchPath []string
}

func NewPostgresBackend(ctx context.Context, connString string) (*PostgresBackend, error) {
	b := &PostgresBackend{
		originalConnectionString: connString,
		rowReader:                newPgxRowReader(),
	}

	if err := b.init(ctx); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *PostgresBackend) init(ctx context.Context) error {
	db, err := b.Connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := b.loadSearchPath(ctx, db); err != nil {
		return err
	}

	return b.loadSchemaNames(db)
}

// Connect implements Backend.
func (b *PostgresBackend) Connect(ctx context.Context, opts ...BackendOption) (*sql.DB, error) {
	connString := b.originalConnectionString
	connector, err := NewPgxConnector(connString, b.afterConnectFunc)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "Unable to parse connection string")
	}

	config := NewBackendConfig(opts)

	db := sql.OpenDB(connector)
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)
	db.SetConnMaxLifetime(config.MaxConnLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)

	// resolve the required search path
	if err := b.resolveDesiredSearchPath(ctx, db, config.SearchPathConfig); err != nil {
		return nil, err
	}
	return db, nil
}

func (b *PostgresBackend) ConnectionString() string {
	return b.originalConnectionString
}

func (b *PostgresBackend) Name() string {
	return constants.PostgresBackendName
}

// RowReader implements Backend.
func (b *PostgresBackend) RowReader() RowReader {
	return b.rowReader
}

// OriginalSearchPath implements SearchPathProvider.
func (b *PostgresBackend) OriginalSearchPath() []string {
	return b.originalSearchPath
}

// RequiredSearchPath implements SearchPathProvider
func (b *PostgresBackend) RequiredSearchPath() []string {
	return b.requiredSearchPath
}

// ResolvedSearchPath implements SearchPathProvider
func (b *PostgresBackend) ResolvedSearchPath() []string {
	if len(b.requiredSearchPath) != 0 {
		return b.requiredSearchPath
	}
	return b.originalSearchPath
}

// afterConnectFunc is called after the connection is established
func (b *PostgresBackend) afterConnectFunc(ctx context.Context, conn driver.Conn) error {
	if len(b.requiredSearchPath) == 0 {
		return nil
	}
	connPc, ok := conn.(driver.ConnPrepareContext)
	if !ok {
		return fmt.Errorf("stdlib driver does not implement ConnPrepareContext")
	}
	ps, err := connPc.PrepareContext(ctx, "SET search_path TO "+strings.Join(b.requiredSearchPath, ","))
	if err != nil {
		return err
	}
	ec, ok := ps.(driver.StmtExecContext)
	if !ok {
		return fmt.Errorf("prepared statement does not implement StmtExecContext")
	}
	defer ps.Close()

	_, err = ec.ExecContext(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

// loadSearchPath gets the current search path from the database
func (b *PostgresBackend) loadSearchPath(ctx context.Context, db *sql.DB) error {
	// Get a connection from the database
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Execute the query
	row := conn.QueryRowContext(ctx, "SHOW search_path")
	if row.Err() != nil {
		return row.Err()
	}

	var searchPathStr string
	// Scan the result into the searchPath variable
	err = row.Scan(&searchPathStr)
	if err != nil {
		return err
	}

	// Split the search path into individual schemas
	searchPath := strings.Split(searchPathStr, ",")
	// Trim spaces from each path
	for i, path := range searchPath {
		searchPath[i] = strings.TrimSpace(path)
	}

	b.originalSearchPath = searchPath
	return nil
}

func (b *PostgresBackend) loadSchemaNames(db *sql.DB) error {
	// SQL query to select all schema names
	query := `SELECT schema_name FROM information_schema.schemata;`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to read schema names from the database")
	}
	defer rows.Close()

	var schemaNames []string

	// Iterate over the results
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return sperr.WrapWithMessage(err, "failed to read a schema name from the database")
		}
		schemaNames = append(schemaNames, schemaName)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return sperr.WrapWithMessage(err, "error encountered while reading schema names from the database")
	}

	// Here, you can use the schemaNames slice which contains all the schema names
	// For example, storing the schema names in the struct
	b.schemaNames = schemaNames

	return nil
}

// resolveDesiredSearchPath resolves the desired search path from the prefix or the custom search path
func (b *PostgresBackend) resolveDesiredSearchPath(ctx context.Context, db *sql.DB, cfg SearchPathConfig) error {
	if len(cfg.SearchPath) > 0 && len(cfg.SearchPathPrefix) > 0 {
		return sperr.WrapWithMessage(ErrInvalidConfig, "cannot specify both search_path and search_path_prefix")
	}

	if len(cfg.SearchPath) == 0 && len(cfg.SearchPathPrefix) == 0 {
		return nil
	}

	if len(cfg.SearchPath) > 0 {
		b.requiredSearchPath = b.cleanSearchPath(cfg.SearchPath)
		return nil
	}

	// must be that the SearchPathPrefix is set
	requiredSearchPath, err := b.constructSearchPathFromPrefix(ctx, db, cfg)
	if err != nil {
		return err
	}
	b.requiredSearchPath = requiredSearchPath

	return nil
}

// constructSearchPathFromPrefix constructs the search path from the prefix and the original search path
func (b *PostgresBackend) constructSearchPathFromPrefix(ctx context.Context, db *sql.DB, cfg SearchPathConfig) ([]string, error) {
	searchPathPrefix := b.cleanSearchPath(cfg.SearchPathPrefix)
	return append(searchPathPrefix, b.originalSearchPath...), nil
}

// the prefix is prepended to the original search path
func (b *PostgresBackend) cleanSearchPath(searchPath []string) []string {
	return helpers.RemoveFromStringSlice(searchPath, "")
}

func newPgxRowReader() *pgxRowReader {
	return &pgxRowReader{
		BasicRowReader: BasicRowReader{
			CellReader: pgxReadCell,
		},
	}
}

// pgxRowReader is a RowReader implementation for the pgx database/sql driver
type pgxRowReader struct {
	BasicRowReader
}

func pgxReadCell(columnValue any, col *queryresult.ColumnDef) (any, error) {
	var result any
	if columnValue != nil {
		result = columnValue

		// add special handling for some types
		switch col.DataType {
		case "_TEXT":
			if arr, ok := columnValue.([]interface{}); ok {
				elements := utils.Map(arr, func(e interface{}) string { return e.(string) })
				result = strings.Join(elements, ",")
			}
		case "INET":
			if inet, ok := columnValue.(netip.Prefix); ok {
				result = strings.TrimSuffix(inet.String(), "/32")
			}
		case "UUID":
			if bytes, ok := columnValue.([16]uint8); ok {
				if u, err := uuid.FromBytes(bytes[:]); err == nil {
					result = u
				}
			}
		case "TIME":
			if t, ok := columnValue.(pgtype.Time); ok {
				result = time.UnixMicro(t.Microseconds).UTC().Format("15:04:05")
			}
		case "INTERVAL":
			if interval, ok := columnValue.(pgtype.Interval); ok {
				var sb strings.Builder
				years := interval.Months / 12
				months := interval.Months % 12
				if years > 0 {
					sb.WriteString(fmt.Sprintf("%d %s ", years, utils.Pluralize("year", int(years))))
				}
				if months > 0 {
					sb.WriteString(fmt.Sprintf("%d %s ", months, utils.Pluralize("mon", int(months))))
				}
				if interval.Days > 0 {
					sb.WriteString(fmt.Sprintf("%d %s ", interval.Days, utils.Pluralize("day", int(interval.Days))))
				}
				if interval.Microseconds > 0 {
					d := time.Duration(interval.Microseconds) * time.Microsecond
					formatStr := time.Unix(0, 0).UTC().Add(d).Format("15:04:05")
					sb.WriteString(formatStr)
				}
				result = sb.String()
			}

		case "NUMERIC":
			if numeric, ok := columnValue.(pgtype.Numeric); ok {
				if f, err := numeric.Float64Value(); err == nil {
					result = f.Float64
				}
			}
		case "JSON", "JSONB":
			var dst any
			err := json.Unmarshal(columnValue.([]byte), &dst)
			if err != nil {
				return nil, err
			}
			result = dst
		}
	}
	return result, nil
}
