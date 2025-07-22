package backend

import (
	"context"
	"database/sql"
	"fmt"
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

func ConnectDucklake(ctx context.Context, db *sql.DB, dbPath, dataPath string) error {
	// 1. Install sqlite extension
	_, err := db.ExecContext(ctx, "install sqlite")
	if err != nil {
		return fmt.Errorf("failed to install sqlite extension: %v", err)
	}

	// 2. Install ducklake extension
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
	attachQuery := fmt.Sprintf("attach 'ducklake:sqlite:%s' AS %s (data_path '%s/')", dbPath, constants.DuckLakeCatalog, dataPath)
	_, err = db.ExecContext(ctx, attachQuery)
	if err != nil {
		return fmt.Errorf("failed to attach sqlite database: %v", err)
	}

	// set default catalog to ducklake
	_, err = db.ExecContext(ctx, fmt.Sprintf("use %s", constants.DuckLakeCatalog))
	if err != nil {
		return fmt.Errorf("failed to set catalog: %w", err)
	}

	return nil
}

// OnConnection implements ConnectionInitializer.
// This function is called by the dbClient after obtaining a new connection
// We use it to set the default catalog to tailpipe_ducklake
func (b *DucklakeBackend) OnConnection(ctx context.Context, conn *sql.Conn) error {
	// set default catalog to ducklake
	_, err := conn.ExecContext(ctx, fmt.Sprintf("use %s", constants.DuckLakeCatalog))
	return err
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
