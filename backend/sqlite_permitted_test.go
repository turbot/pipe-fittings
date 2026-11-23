//go:build cgo

package backend

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func connectSqlite(t *testing.T, dbPath string, maxOpen int) (*sql.DB, error) {
	t.Helper()
	b, err := NewSqliteBackend(SqliteConnectionStringPrefix + dbPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg := NewBackendConfig(nil)
	cfg.MaxOpenConns = maxOpen
	db, err := b.Connect(context.Background(), WithConfig(cfg))
	if db != nil {
		t.Cleanup(func() { _ = db.Close() })
	}
	return db, err
}

// TestSqlitePermittedUnset shows that, with no setting, SQL on a SQLite database can attach, and so create, a
// file outside the folder the database lives in; the setting exists to stop that.
func TestSqlitePermittedUnset(t *testing.T) {
	f := newPermittedFixture(t)
	if err := setPermittedEnv(t, ""); err != nil {
		t.Fatal(err)
	}
	db, err := connectSqlite(t, filepath.Join(f.permitted, "main.db"), 2)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	outFile := filepath.Join(f.outside, "x.db")
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "ATTACH '"+outFile+"' AS o"); err != nil {
		t.Fatalf("ATTACH outside: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "CREATE TABLE o.t(a)"); err != nil {
		t.Fatalf("CREATE TABLE in the outside file: %v", err)
	}
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("the outside file was not created: %v", err)
	}
}

// sqliteAttachFails checks that attaching file fails on conn and leaves no file at created
func sqliteAttachFails(t *testing.T, conn *sql.Conn, what, file, created string) {
	t.Helper()
	sqliteAttachExprFails(t, conn, what, "'"+file+"'", created)
}

// sqliteAttachExprFails checks that ATTACH of the expression expr (with args bound to it) fails on conn and
// leaves no file at created
func sqliteAttachExprFails(t *testing.T, conn *sql.Conn, what, expr, created string, args ...any) {
	t.Helper()
	_, err := conn.ExecContext(context.Background(), "ATTACH "+expr+" AS x_"+strings.NewReplacer(" ", "_", "-", "_").Replace(what), args...)
	if err == nil {
		t.Errorf("%s: ATTACH %s: want an error, got none", what, expr)
	} else if !strings.Contains(err.Error(), "not authorized") {
		t.Errorf("%s: ATTACH %s: want the authorizer's refusal, got: %v", what, expr, err)
	}
	if created != "" {
		if _, err := os.Stat(created); err == nil {
			t.Errorf("%s: ATTACH created %s", what, created)
		}
	}
}

func TestSqlitePermittedAttach(t *testing.T) {
	f := newPermittedFixture(t)
	permitFolders(t, f.permitted)
	ctx := context.Background()

	db, err := connectSqlite(t, filepath.Join(f.permitted, "main.db"), 3)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	var conns []*sql.Conn
	for i := 0; i < 3; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}
	defer func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}()
	if open := db.Stats().OpenConnections; open != 3 {
		t.Fatalf("open connections = %d, want 3", open)
	}

	outFile := filepath.Join(f.outside, "x.db")
	for i, conn := range conns {
		t.Run("connection "+string(rune('1'+i)), func(t *testing.T) {
			sqliteAttachFails(t, conn, "outside", outFile, outFile)
			sqliteAttachFails(t, conn, "outside uri", "file:"+outFile+"?mode=rwc", outFile)
			sqliteAttachFails(t, conn, "uri inside", "file:"+filepath.Join(f.permitted, "u.db"), filepath.Join(f.permitted, "u.db"))
			link := filepath.Join(f.permitted, "linkdir", "y.db")
			sqliteAttachFails(t, conn, "symlink", link, filepath.Join(f.outside, "y.db"))
			sqliteAttachFails(t, conn, "dotdot", filepath.Join(f.permitted, "..", "outside", "z.db"), filepath.Join(f.outside, "z.db"))
			// SQLite tells the authorizer no file name for an expression or a bound parameter
			exprFile := filepath.Join(f.outside, "expr.db")
			dir, base := filepath.Split(exprFile)
			sqliteAttachExprFails(t, conn, "concatenation", "'"+dir+"' || '"+base+"'", exprFile)
			paramFile := filepath.Join(f.outside, "param.db")
			sqliteAttachExprFails(t, conn, "bound parameter", "?", paramFile, paramFile)
			sqliteAttachExprFails(t, conn, "bound parameter inside", "?", "", filepath.Join(f.permitted, "param.db"))

			inFile := filepath.Join(f.permitted, "in"+string(rune('1'+i))+".db")
			name := "i" + string(rune('1'+i))
			if _, err := conn.ExecContext(ctx, "ATTACH '"+inFile+"' AS "+name); err != nil {
				t.Errorf("ATTACH inside: %v", err)
			}
			if _, err := conn.ExecContext(ctx, "CREATE TABLE "+name+".t(a)"); err != nil {
				t.Errorf("CREATE TABLE in the attached file: %v", err)
			}
			if _, err := conn.ExecContext(ctx, "ATTACH ':memory:' AS m"); err != nil {
				t.Errorf("ATTACH :memory:: %v", err)
			}
			if _, err := conn.ExecContext(ctx, "CREATE TABLE m.t(a)"); err != nil {
				t.Errorf("CREATE TABLE in memory: %v", err)
			}
		})
	}
}

func TestSqlitePermittedBadSetting(t *testing.T) {
	f := newPermittedFixture(t)
	if err := setPermittedEnv(t, "relative/folder"); err == nil {
		t.Fatal("want the relative entry refused")
	}
	db, err := connectSqlite(t, filepath.Join(f.permitted, "main.db"), 2)
	if err == nil || !strings.Contains(err.Error(), testPermittedEnv) {
		t.Errorf("Connect = %v; want an error naming the setting", err)
	}
	if db != nil {
		t.Error("Connect returned a pool with the error")
	}
}

// fakeConn is a driver.Conn with no RegisterAuthorizer
type fakeConn struct{}

func (fakeConn) Prepare(string) (driver.Stmt, error) { return nil, nil }
func (fakeConn) Close() error                        { return nil }
func (fakeConn) Begin() (driver.Tx, error)           { return nil, nil }

func TestSqlitePermittedNoAuthorizer(t *testing.T) {
	f := newPermittedFixture(t)
	permitFolders(t, f.permitted)
	init, err := sqlitePermittedInit()
	if err != nil {
		t.Fatal(err)
	}
	if err := init(context.Background(), fakeConn{}); err == nil || !strings.Contains(err.Error(), "authorizer") {
		t.Errorf("init on a connection without RegisterAuthorizer = %v; want an error naming the authorizer", err)
	}
}
