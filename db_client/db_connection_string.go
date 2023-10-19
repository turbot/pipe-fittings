package db_client

import "strings"

// getUseableConnectionString returns a connection string that can be used by the database driver
// this is derived from the connection string passed in by the user and the driver name
func getUseableConnectionString(driver string, connString string) string {
	// using this to remove the sqlite3?:// prefix from the connection string
	// this is necessary because the sqlite3 driver doesn't recognize the sqlite3?:// prefix
	connString = strings.TrimPrefix(connString, "sqlite3://")
	connString = strings.TrimPrefix(connString, "sqlite://")
	return connString
}

// isPostgresConnectionString returns true if the connection string is for postgres
// looks for the postgresql:// or postgres:// prefix
func isPostgresConnectionString(connString string) bool {
	return strings.HasPrefix(connString, "postgresql://") || strings.HasPrefix(connString, "postgres://")
}

// isSqliteConnectionString returns true if the connection string is for sqlite
// looks for the sqlite:// prefix
func isSqliteConnectionString(connString string) bool {
	return strings.HasPrefix(connString, "sqlite://")
}

// isMySqlConnectionString returns true if the connection string is for mysql
// looks for the mysql:// prefix
func isMySqlConnectionString(connString string) bool {
	return strings.HasPrefix(connString, "mysql://")
}

func IsConnectionString(connString string) bool {
	isPostgres := isPostgresConnectionString(connString)
	isSqlite := isSqliteConnectionString(connString)
	return isPostgres || isSqlite
}
