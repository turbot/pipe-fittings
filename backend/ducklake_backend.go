package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

const (
	ducklakeConnectionStringPrefix = "ducklake:"
)

type DucklakeBackend struct {
	dbPath    string
	dataPath  string
	rowReader RowReader
	filters   *DatabaseFilters
}

func NewDucklakeBackend(connString string) (*DucklakeBackend, error) {
	connString = strings.TrimSpace(connString) // remove any leading or trailing whitespace
	dbPath, dataPath, err := ParseDucklakeConnectionString(connString)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not parse ducklake connection string: '%s'", connString)
	}

	return &DucklakeBackend{
		dbPath:    dbPath,
		dataPath:  dataPath,
		rowReader: newDuckDBRowReader(),
	}, nil
}

// Connect implements Backend.
func (b *DucklakeBackend) Connect(ctx context.Context, options ...BackendOption) (*sql.DB, error) {
	config := NewBackendConfig(options)

	// Store filters for later use in view creation
	b.filters = config.Filters

	db, err := sql.Open("duckdb", "")
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

	if err = ConnectDucklake(ctx, db, b.dbPath, b.dataPath); err != nil {
		return nil, err
	}

	if err := b.createViews(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to create views: %w", err)
	}

	return db, nil
}

func (b *DucklakeBackend) ConnectionString() string {
	return GetDucklakeConnectionString(b.dbPath, b.dataPath)
}

func (b *DucklakeBackend) Name() string {
	return constants.DucklakeBackendName
}

// RowReader implements Backend.
func (b *DucklakeBackend) RowReader() RowReader {
	return b.rowReader
}

func (b *DucklakeBackend) createViews(ctx context.Context, db *sql.DB) error {
	// get list of tables
	tableQuery := fmt.Sprintf("select table_name FROM %s.ducklake_table", constants.DuckLakeMetadataCatalog)

	// Execute the query
	rows, err := db.QueryContext(ctx, tableQuery)
	if err != nil {
		return fmt.Errorf("failed to query ducklake tables: %w", err)
	}
	defer rows.Close()
	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan ducklake table name: %w", err)
		}
		tableNames = append(tableNames, tableName)
	}
	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over ducklake tables: %w", err)
	}

	// Create views for each table
	for _, tableName := range tableNames {
		// build the (possibly empty) filter clause
		filterClause := b.buildFilterClause()
		// build the view creation query
		createViewQuery := fmt.Sprintf(`
				create view %s as 
				select * from %s.%s 
				%s`, tableName, constants.DuckLakeCatalog, tableName, filterClause)

		_, err = db.ExecContext(ctx, createViewQuery)
		if err != nil {
			return fmt.Errorf("failed to create view for table %s: %w", tableName, err)
		}
	}

	return nil
}

// buildFilterClause builds the WHERE clause for the filters
func (b *DucklakeBackend) buildFilterClause() string {
	if b.filters == nil {
		return ""
	}

	var conditions []string

	// Add partition filters
	if len(b.filters.Partitions) > 0 {
		partitionCondition := fmt.Sprintf("partition IN (%s)",
			strings.Join(b.filters.Partitions, ","))
		conditions = append(conditions, partitionCondition)
	}

	// Add index filters
	if len(b.filters.Indexes) > 0 {
		indexCondition := fmt.Sprintf("index IN (%s)",
			strings.Join(b.filters.Indexes, ","))
		conditions = append(conditions, indexCondition)
	}

	// Add time range filters
	if b.filters.From != nil {
		fromCondition := fmt.Sprintf("timestamp >= '%s'", b.filters.From.Format("2006-01-02 15:04:05"))
		conditions = append(conditions, fromCondition)
	}

	if b.filters.To != nil {
		toCondition := fmt.Sprintf("timestamp <= '%s'", b.filters.To.Format("2006-01-02 15:04:05"))
		conditions = append(conditions, toCondition)
	}

	if len(conditions) == 0 {
		return "" // No filters, return all rows
	}

	return "where " + strings.Join(conditions, " and ")
}

// TODO #DL: use default data location - remove DataPath everywhere

func ConnectDucklake(ctx context.Context, db *sql.DB, dbPath, dataPath string, creds ...string) error {
	// 1. Install sqlite extension
	_, err := db.ExecContext(ctx, "install sqlite")
	if err != nil {
		return fmt.Errorf("failed to install sqlite extension: %w", err)
	}

	// 2. Install extensions
	slog.Info("loading aws, parquet, httpfs extensions")
	// TODO #DL: enscapsulate extension loading and only load s3 related ones if needed
	// load aws, http and parquet for S3 support
	_, err = db.ExecContext(ctx, "install parquet")
	if err != nil {
		return fmt.Errorf("failed to install parquet extension: %w", err)
	}
	_, err = db.ExecContext(ctx, "load parquet")
	if err != nil {
		return fmt.Errorf("failed to load parquet extension: %v", err)
	}
	_, err = db.ExecContext(ctx, "install httpfs")
	if err != nil {
		return fmt.Errorf("failed to install httpfs extension: %w", err)
	}
	_, err = db.ExecContext(ctx, "load httpfs")
	if err != nil {
		return fmt.Errorf("failed to load httpfs extension: %w", err)
	}
	_, err = db.ExecContext(ctx, "install aws")
	if err != nil {
		return fmt.Errorf("failed to install aws extension: %w", err)
	}
	_, err = db.ExecContext(ctx, "load aws")
	if err != nil {
		return fmt.Errorf("failed to load aws extension: %w", err)
	}
	slog.Info("loading aws credentials")
	// load aws creds
	_, err = db.ExecContext(ctx, "call load_aws_credentials()")
	if err != nil {
		return fmt.Errorf("failed to load aws credentials: %w", err)
	}

	// TODO #DL change to using prod extension when stable
	//  https://github.com/turbot/tailpipe/issues/476
	//_, err = db.Exec("install ducklake;")
	_, err = db.ExecContext(ctx, "force install ducklake from core_nightly")
	if err != nil {
		return fmt.Errorf("failed to install ducklake nightly extension: %v", err)
	}
	_, err = db.ExecContext(ctx, "load ducklake")
	if err != nil {
		return fmt.Errorf("failed to load ducklake extension: %v", err)
	}

	// 3. Attach the sqlite database as my_ducklake
	slog.Info("attaching sqlite database", "dbPath", dbPath, "dataPath", dataPath)

	attachQuery := fmt.Sprintf("attach 'ducklake:sqlite:%s' AS %s (data_path '%s/')", dbPath, constants.DuckLakeCatalog, dataPath)
	_, err = db.ExecContext(ctx, attachQuery)
	if err != nil {
		return fmt.Errorf("failed to attach sqlite database: %v", err)
	}

	// TODO #DL figure out appropriate row group size
	// 4. Set the row group size for parquet files
	rowGroupQuery := fmt.Sprintf("call ducklake_set_option('%s', 'parquet_row_group_size', 10000);", constants.DuckLakeCatalog)
	_, err = db.ExecContext(ctx, rowGroupQuery)
	if err != nil {
		return fmt.Errorf("failed to attach sqlite database: %v", err)
	}

	return nil
}

func ParseDucklakeConnectionString(connectionString string) (string, string, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", "", err
	}

	// Db path comes from the Path component
	dbPath := u.Path

	// Data path comes from the query parameter
	dataDir := u.Query().Get("data_path")

	return dbPath, dataDir, err
}

func GetDucklakeConnectionString(dbPath, dataPath string) string {
	return fmt.Sprintf("ducklake://%s?data_path=%s", dbPath, dataPath)
}
