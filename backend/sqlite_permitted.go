package backend

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/sperr"
)

// SQLite's authorizer action and result codes (sqlite3.h). Declared here because pipe-fittings imports no
// SQLite driver; they are fixed by SQLite's file format and C interface.
const (
	sqliteOK     = 0  // SQLITE_OK: allow the action
	sqliteDeny   = 1  // SQLITE_DENY: fail the statement with "not authorized"
	sqliteAttach = 24 // SQLITE_ATTACH: ATTACH DATABASE; the first argument is the file name when written as a literal, else empty
)

// sqliteAuthorizerConn is what a SQLite driver connection must have for the lock: mattn/go-sqlite3's
// *SQLiteConn has it
type sqliteAuthorizerConn interface {
	RegisterAuthorizer(func(op int, arg1, arg2, arg3 string) int)
}

// connectSqlitePermitted opens the SQLite dsn so that no connection of the pool can ATTACH a file outside the
// permitted folders: every connection the pool opens runs sqlitePermittedInit before it is handed out.
// configure sets the pool's limits. One connection is opened here, so a driver that cannot be locked fails
// the open, not the first query; on any failure the pool is closed and the error returned.
func connectSqlitePermitted(ctx context.Context, dsn string, configure func(*sql.DB)) (*sql.DB, error) {
	init, err := sqlitePermittedInit()
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to sqlite backend")
	}
	connector, err := NewConnInitConnector("sqlite3", dsn, init)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not connect to sqlite backend")
	}
	db := sql.OpenDB(connector)
	configure(db)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, sperr.WrapWithMessage(err, "could not connect to sqlite backend")
	}
	return db, nil
}

// sqlitePermittedInit returns the init that registers, on each new SQLite connection, an authorizer that
// refuses an ATTACH of a file outside the permitted folders. The file name is judged as the permitted_paths
// check judges every path (made absolute, symlinks resolved). A name that starts with file: is a URI, whose
// real path and options this cannot judge, so it is refused. SQLite passes the name only when the statement
// writes it as a literal: for an expression or a bound parameter (ATTACH 'a' || 'b', ATTACH ?) the authorizer
// gets an empty name, which cannot be judged, so an empty name is refused, and with it an ATTACH of the empty
// string, which opens a temporary database. A plain VACUUM is refused too, as SQLite runs it by attaching the
// empty string (VACUUM INTO a file is judged like an ATTACH of that file). :memory: opens no file and stays
// allowed, as does every other action.
//
// A bad setting fails here, so the open fails closed. A connection without RegisterAuthorizer cannot be
// locked, and fails its init.
func sqlitePermittedInit() (ConnInitFunc, error) {
	if _, err := permitted_paths.ResolvedRoots(); err != nil {
		return nil, err
	}
	if !permitted_paths.Enabled() {
		return nil, fmt.Errorf("%s names no folder", permitted_paths.Setting())
	}
	return func(_ context.Context, conn driver.Conn) error {
		c, ok := conn.(sqliteAuthorizerConn)
		if !ok {
			return fmt.Errorf("the sqlite driver has no authorizer (RegisterAuthorizer), so %s cannot be applied", permitted_paths.Setting())
		}
		c.RegisterAuthorizer(sqliteAuthorize)
		return nil
	}, nil
}

// sqliteAuthorize is the authorizer callback: it denies an ATTACH outside the permitted folders
func sqliteAuthorize(op int, file, _, _ string) int {
	if op != sqliteAttach {
		return sqliteOK
	}
	if file == ":memory:" {
		return sqliteOK
	}
	// an empty name is an expression or a bound parameter, whose file SQLite has not worked out yet
	if file == "" || strings.HasPrefix(file, "file:") {
		return sqliteDeny
	}
	if _, err := permitted_paths.Check(file); err != nil {
		return sqliteDeny
	}
	return sqliteOK
}
