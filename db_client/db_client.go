package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_client/backend"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"golang.org/x/sync/semaphore"
	"log"
	"strings"
)

// define func type for StartQuery
type startQueryFunc func(ctx context.Context, dbConn *sql.Conn, query string, args ...any) (*sql.Rows, error)

// DbClient wraps over `sql.DB` and gives an interface to the database
type DbClient struct {
	connectionString string

	// connection UserPool for user initiated queries
	UserPool *sql.DB

	// connection used to run system/plumbing queries (connection state, server settings)
	ManagementPool *sql.DB

	// function to start the query - defaults to startquery
	// steampipe overrides this with startQueryWithRetries
	startQueryFunc startQueryFunc

	// concurrency management for db session access
	parallelSessionInitLock *semaphore.Weighted

	// the backend type of the dbclient backend
	backend backend.DBClientBackendType

	// a reader which can be used to read rows from a pgx.Rows object
	rowReader backend.RowReader

	// TODO KAI new hook
	BeforeExecuteHook func(context.Context, *sql.Conn) error

	// if a custom search path or a prefix is used, store it here
	CustomSearchPath []string
	SearchPathPrefix []string
	// the default user search path
	UserSearchPath []string

	// TODO KAI NEEDED?
	//onConnectionCallback DbConnectionCallback

}

func NewDbClient(ctx context.Context, connectionString string, opts ...ClientOption) (_ *DbClient, err error) {
	utils.LogTime("db_client.NewDbClient start")
	defer utils.LogTime("db_client.NewDbClient end")

	backendType, err := backend.GetBackendFromConnectionString(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	client := &DbClient{
		// a weighted semaphore to control the maximum number parallel
		// initializations under way
		parallelSessionInitLock: semaphore.NewWeighted(constants.MaxParallelClientInits),
		connectionString:        connectionString,
		backend:                 backendType,
		rowReader:               backend.RowReaderFactory(backendType),
	}

	// set the start query func
	client.startQueryFunc = client.StartQuery

	defer func() {
		if err != nil {
			// try closing the client
			client.Close(ctx)
		}
	}()

	config := clientConfig{}
	for _, o := range opts {
		o(&config)
	}

	if err := client.establishConnectionPool(ctx, config); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *DbClient) closePools() {
	if c.UserPool != nil {
		c.UserPool.Close()
	}
	if c.ManagementPool != nil {
		c.ManagementPool.Close()
	}
}

func (c *DbClient) GetConnectionString() string {
	return c.connectionString
}

// RegisterNotificationListener has an empty implementation
// NOTE: we do not (currently) support notifications from remote connections
func (c *DbClient) RegisterNotificationListener(func(notification *pgconn.Notification)) {}

// closes the connection to the database and shuts down the backend
func (c *DbClient) Close(context.Context) error {
	log.Printf("[TRACE] DbClient.Close %v", c.UserPool)
	c.closePools()

	return nil
}

// TODO KAI STEAMPIPE ONLY
// GetSchemaFromDB  retrieves schemas for all steampipe connections (EXCEPT DISABLED CONNECTIONS)
// NOTE: it optimises the schema extraction by extracting schema information for
// connections backed by distinct plugins and then fanning back out.
func (c *DbClient) GetSchemaFromDB(ctx context.Context) (*db_common.SchemaMetadata, error) {
	log.Printf("[INFO] DbClient GetSchemaFromDB")
	// mgmtConn, err := c.ManagementPool.Acquire(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// defer mgmtConn.Release()

	// TODO KAI not needed for powerpipe
	return nil, sperr.New("not supported in Powerpipe")
	//
	//// for optimisation purposes, try to load connection state and build a map of schemas to load
	//// (if we are connected to a remote server running an older CLI,
	//// this load may fail, in which case bypass the optimisation)
	//connectionStateMap, err := modconfig.LoadConnectionState(ctx, mgmtConn.Conn(), modconfig.WithWaitUntilLoading())
	//// NOTE: if we failed to load connection state, this may be because we are connected to an older version of the CLI
	//// use legacy (v0.19.x) schema loading code
	//if err != nil {
	//	return c.GetSchemaFromDBLegacy(ctx, mgmtConn)
	//}
	//
	//// build a ConnectionSchemaMap object to identify the schemas to load
	//connectionSchemaMap := modconfig.NewConnectionSchemaMap(ctx, connectionStateMap, c.GetRequiredSessionSearchPath())
	//if err != nil {
	//	return nil, err
	//}
	//
	//// get the unique schema - we use this to limit the schemas we load from the database
	//schemas := maps.Keys(connectionSchemaMap)
	//
	//// build a query to retrieve these schemas
	//query := c.buildSchemasQuery(schemas...)
	//
	//// build schema metadata from query result
	//metadata, err := db_common.LoadSchemaMetadata(ctx, mgmtConn.Conn(), query)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// we now need to add in all other schemas which have the same schemas as those we have loaded
	//for loadedSchema, otherSchemas := range connectionSchemaMap {
	//	// all 'otherSchema's have the same schema as loadedSchema
	//	exemplarSchema, ok := metadata.Schemas[loadedSchema]
	//	if !ok {
	//		// should can happen in the case of a dynamic plugin with no tables - use empty schema
	//		exemplarSchema = make(map[string]db_common.TableSchema)
	//	}
	//
	//	for _, s := range otherSchemas {
	//		metadata.Schemas[s] = exemplarSchema
	//	}
	//}
	//
	//return metadata, nil
}

func (c *DbClient) GetSchemaFromDBLegacy(ctx context.Context, conn *sql.Conn) (*db_common.SchemaMetadata, error) {
	// build a query to retrieve these schemas
	query := c.buildSchemasQueryLegacy()

	// build schema metadata from query result
	return db_common.LoadSchemaMetadata(ctx, conn, query)
}

// TODO KAI STEAMPIPE ONLY
// Unimplemented (sql.DB does not have a mechanism to reset pools) - refreshDbClient terminates the current connection and opens up a new connection to the service.
func (c *DbClient) ResetPools(ctx context.Context) {
	log.Println("[TRACE] db_client.ResetPools start")
	defer log.Println("[TRACE] db_client.ResetPools end")

	// c.UserPool.Reset()
	// c.ManagementPool.Reset()
}

func (c *DbClient) buildSchemasQuery(schemas ...string) string {
	for idx, s := range schemas {
		schemas[idx] = fmt.Sprintf("'%s'", s)
	}

	// build the schemas filter clause
	schemaClause := ""
	if len(schemas) > 0 {
		schemaClause = fmt.Sprintf(`
    cols.table_schema in (%s)
	OR`, strings.Join(schemas, ","))
	}

	query := fmt.Sprintf(`
SELECT
		table_name,
		column_name,
		column_default,
		is_nullable,
		data_type,
		udt_name,
		table_schema,
		(COALESCE(pg_catalog.col_description(c.oid, cols.ordinal_position :: int),'')) as column_comment,
		(COALESCE(pg_catalog.obj_description(c.oid),'')) as table_comment
FROM
    information_schema.columns cols
LEFT JOIN
    pg_catalog.pg_namespace nsp ON nsp.nspname = cols.table_schema
LEFT JOIN
    pg_catalog.pg_class c ON c.relname = cols.table_name AND c.relnamespace = nsp.oid
WHERE %s
	LEFT(cols.table_schema,8) = 'pg_temp_'
`, schemaClause)
	return query
}
func (c *DbClient) buildSchemasQueryLegacy() string {

	query := `
WITH distinct_schema AS (
	SELECT DISTINCT(foreign_table_schema) 
	FROM 
		information_schema.foreign_tables 
	WHERE 
		foreign_table_schema <> 'steampipe_command'
)
SELECT
    table_name,
    column_name,
    column_default,
    is_nullable,
    data_type,
    udt_name,
    table_schema,
    (COALESCE(pg_catalog.col_description(c.oid, cols.ordinal_position :: int),'')) as column_comment,
    (COALESCE(pg_catalog.obj_description(c.oid),'')) as table_comment
FROM
    information_schema.columns cols
LEFT JOIN
    pg_catalog.pg_namespace nsp ON nsp.nspname = cols.table_schema
LEFT JOIN
    pg_catalog.pg_class c ON c.relname = cols.table_name AND c.relnamespace = nsp.oid
WHERE
	cols.table_schema in (select * from distinct_schema)
	OR
    LEFT(cols.table_schema,8) = 'pg_temp_'

`
	return query
}
