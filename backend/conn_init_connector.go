package backend

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
)

// ConnInitFunc runs on every new driver connection before database/sql hands it out
type ConnInitFunc func(ctx context.Context, conn driver.Conn) error

// ConnInitConnector is a driver.Connector that wraps the driver registered under a name and runs an init
// function on every new connection, as PgxConnector does for pgx. It imports no driver package: the host
// registers the driver, and a connection whose init fails is closed and never handed out.
type ConnInitConnector struct {
	inner driver.Connector
	drv   driver.Driver
	init  ConnInitFunc
}

// NewConnInitConnector returns a connector for dsn through the driver registered under driverName, running
// init on every new connection.
//
// database/sql has no lookup of a registered driver by name, so the driver is taken from a pool opened
// with an empty DSN and closed at once. sql.Open never connects, but for a driver.DriverContext it does call
// OpenConnector, which for DuckDB opens the database: opening the real DSN that way and again here would open
// the file twice (and DuckDB locks the file), so the throwaway pool uses the empty DSN, which DuckDB reads
// as a private in-memory database and a driver without DriverContext does not open at all.
func NewConnInitConnector(driverName, dsn string, init ConnInitFunc) (*ConnInitConnector, error) {
	probe, err := sql.Open(driverName, "")
	if err != nil {
		return nil, err
	}
	drv := probe.Driver()
	if err := probe.Close(); err != nil {
		return nil, err
	}

	c := &ConnInitConnector{drv: drv, init: init}
	if dc, ok := drv.(driver.DriverContext); ok {
		inner, err := dc.OpenConnector(dsn)
		if err != nil {
			return nil, err
		}
		c.inner = inner
	} else {
		c.inner = &dsnConnector{dsn: dsn, drv: drv}
	}
	return c, nil
}

// Connect opens a connection through the wrapped driver and runs the init on it
func (c *ConnInitConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.inner.Connect(ctx)
	if err != nil {
		return nil, err
	}
	if c.init != nil {
		if err := c.init(ctx, conn); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

// Driver returns the wrapped driver
func (c *ConnInitConnector) Driver() driver.Driver {
	return c.drv
}

// Close closes the wrapped connector when it can be closed; sql.DB.Close calls it
func (c *ConnInitConnector) Close() error {
	if closer, ok := c.inner.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// dsnConnector opens a connection per Connect through a driver without DriverContext, as database/sql does
type dsnConnector struct {
	dsn string
	drv driver.Driver
}

func (c *dsnConnector) Connect(context.Context) (driver.Conn, error) {
	return c.drv.Open(c.dsn)
}

func (c *dsnConnector) Driver() driver.Driver {
	return c.drv
}

// execConn runs one statement with no arguments on a raw driver connection
func execConn(ctx context.Context, conn driver.Conn, query string) error {
	if execer, ok := conn.(driver.ExecerContext); ok {
		_, err := execer.ExecContext(ctx, query, nil)
		if !errors.Is(err, driver.ErrSkip) {
			return err
		}
	}
	stmt, err := prepareConn(ctx, conn, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if se, ok := stmt.(driver.StmtExecContext); ok {
		_, err = se.ExecContext(ctx, nil)
		return err
	}
	_, err = stmt.Exec(nil) //nolint:staticcheck // the fallback for a driver without StmtExecContext
	return err
}

// queryConnValue runs one query with no arguments on a raw driver connection and returns the first column of
// its first row; a query that returns no row is an error
func queryConnValue(ctx context.Context, conn driver.Conn, query string) (driver.Value, error) {
	var rows driver.Rows
	var err error
	if queryer, ok := conn.(driver.QueryerContext); ok {
		rows, err = queryer.QueryContext(ctx, query, nil)
	} else {
		err = driver.ErrSkip
	}
	if errors.Is(err, driver.ErrSkip) {
		var stmt driver.Stmt
		stmt, err = prepareConn(ctx, conn, query)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()
		if sq, ok := stmt.(driver.StmtQueryContext); ok {
			rows, err = sq.QueryContext(ctx, nil)
		} else {
			rows, err = stmt.Query(nil) //nolint:staticcheck // the fallback for a driver without StmtQueryContext
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := rows.Columns()
	if len(cols) == 0 {
		return nil, fmt.Errorf("query %q returned no column", query)
	}
	dest := make([]driver.Value, len(cols))
	if err := rows.Next(dest); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("query %q returned no row", query)
		}
		return nil, err
	}
	return dest[0], nil
}

func prepareConn(ctx context.Context, conn driver.Conn, query string) (driver.Stmt, error) {
	if pc, ok := conn.(driver.ConnPrepareContext); ok {
		return pc.PrepareContext(ctx, query)
	}
	return conn.Prepare(query)
}
