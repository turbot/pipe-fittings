//go:build cgo

package backend

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/turbot/pipe-fittings/v2/filepaths"
)

// duckDBTestFolders is
//
//	<base>/in'side/a.csv     the permitted folder (its name carries a quote, so the SQL escaping is tested)
//	<base>/out/b.csv         a folder outside it
//	<base>/in'sidex/c.csv    a sibling whose name has the permitted folder's as a prefix
type duckDBTestFolders struct {
	in, out, sibling string
}

// newDuckDBTest makes the folders, points the pipes install folder at a temp folder and empties the extension
// setup, so no test needs the network
func newDuckDBTest(t *testing.T) duckDBTestFolders {
	t.Helper()
	prevInstall := filepaths.PipesInstallDir
	filepaths.PipesInstallDir = t.TempDir()
	prevSetup := duckDbExtensionSetup
	duckDbExtensionSetup = func(exec func(string) error) error { return nil }
	t.Cleanup(func() {
		filepaths.PipesInstallDir = prevInstall
		duckDbExtensionSetup = prevSetup
	})

	base := t.TempDir()
	f := duckDBTestFolders{
		in:      filepath.Join(base, "in'side"),
		out:     filepath.Join(base, "out"),
		sibling: filepath.Join(base, "in'sidex"),
	}
	for dir, file := range map[string]string{f.in: "a.csv", f.out: "b.csv", f.sibling: "c.csv"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, file), []byte("x\n1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

// q quotes a path as a DuckDB string literal
func q(path string) string { return duckDBQuote(path) }

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func mustRead(t *testing.T, db sqlRunner, path string) {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), "SELECT count(*) FROM read_csv("+q(path)+")").Scan(&n); err != nil {
		t.Errorf("reading %s: %v", path, err)
	} else if n != 1 {
		t.Errorf("reading %s: %d rows, want 1", path, n)
	}
}

// mustFail checks that query fails because of the lock: a permission error, or a locked setting
func mustFail(t *testing.T, db sqlRunner, what, query string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), query)
	switch {
	case err == nil:
		t.Errorf("%s: want an error, got none (%s)", what, query)
	case !strings.Contains(err.Error(), "Permission Error") && !strings.Contains(err.Error(), "configuration has been locked"):
		t.Errorf("%s: want the lock's error, got: %v", what, err)
	}
}

// assertLocked checks that SQL on db reads inside the folder and nothing outside it, and cannot lift the lock
func assertLocked(t *testing.T, db sqlRunner, f duckDBTestFolders) {
	t.Helper()
	mustRead(t, db, filepath.Join(f.in, "a.csv"))

	outCSV := filepath.Join(f.out, "b.csv")
	mustFail(t, db, "read_csv outside", "SELECT * FROM read_csv("+q(outCSV)+")")
	mustFail(t, db, "read_text outside", "SELECT * FROM read_text("+q(outCSV)+")")
	mustFail(t, db, "read_csv in the sibling", "SELECT * FROM read_csv("+q(filepath.Join(f.sibling, "c.csv"))+")")
	mustFail(t, db, "read_csv through ..", "SELECT * FROM read_csv("+q(f.in+"/../out/b.csv")+")")
	// DuckDB allows its own temp folder; it must not lead out either
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(wd, outCSV)
	if err != nil {
		t.Fatal(err)
	}
	mustFail(t, db, "read_csv through the temp folder", "SELECT * FROM read_csv("+q(".tmp/../"+rel)+")")

	copied := filepath.Join(f.out, "copied.csv")
	mustFail(t, db, "COPY TO outside", "COPY (SELECT 1 AS x) TO "+q(copied))
	if _, err := os.Stat(copied); err == nil {
		t.Errorf("COPY TO wrote %s", copied)
	}
	attached := filepath.Join(f.out, "attached.duckdb")
	mustFail(t, db, "ATTACH outside", "ATTACH "+q(attached)+" AS outside_db")
	if _, err := os.Stat(attached); err == nil {
		t.Errorf("ATTACH created %s", attached)
	}
	ext := filepath.Join(f.out, "evil.duckdb_extension")
	if err := os.WriteFile(ext, []byte("not an extension"), 0o600); err != nil {
		t.Fatal(err)
	}
	mustFail(t, db, "INSTALL from outside", "INSTALL "+q(ext))
	mustFail(t, db, "LOAD from outside", "LOAD "+q(ext))

	mustFail(t, db, "SET enable_external_access", "SET enable_external_access = true")
	mustFail(t, db, "SET allowed_directories", "SET allowed_directories = ['/']")
	mustFail(t, db, "SET allowed_paths", "SET allowed_paths = ["+q(outCSV)+"]")
	mustFail(t, db, "SET lock_configuration", "SET lock_configuration = false")
	mustFail(t, db, "RESET enable_external_access", "RESET enable_external_access")
}

func connectDuckDB(t *testing.T, dsn string, maxOpen int) (*sql.DB, error) {
	t.Helper()
	b, err := NewDuckDBBackend(dsn)
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

func TestDuckDBPermittedLockdown(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)
	ctx := context.Background()

	// the extension setup runs once per database, before the lock: external access is still on
	setups := 0
	duckDbExtensionSetup = func(exec func(string) error) error {
		setups++
		return exec("SELECT * FROM read_csv(" + q(filepath.Join(f.out, "b.csv")) + ")")
	}

	db, err := connectDuckDB(t, "duckdb://", 4)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// four connections held at once, so each is its own driver connection
	var conns []*sql.Conn
	for i := 0; i < 4; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}
	if open := db.Stats().OpenConnections; open != 4 {
		t.Fatalf("open connections = %d, want 4", open)
	}
	for i, conn := range conns {
		t.Run("connection "+string(rune('1'+i)), func(t *testing.T) { assertLocked(t, conn, f) })
	}
	for _, conn := range conns {
		_ = conn.Close()
	}
	if setups != 1 {
		t.Errorf("extension setup ran %d times, want once", setups)
	}

	// let every connection expire, so the next one is opened afresh
	db.SetConnMaxLifetime(time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("fresh connection: %v", err)
	}
	defer conn.Close()
	if closed := db.Stats().MaxLifetimeClosed; closed == 0 {
		t.Fatal("no connection expired; the fresh connection is not new")
	}
	assertLocked(t, conn, f)
}

func TestDuckDBPermittedSetupFails(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)
	duckDbExtensionSetup = func(exec func(string) error) error { return errors.New("no network") }

	db, err := connectDuckDB(t, "duckdb://", 2)
	if err == nil || !strings.Contains(err.Error(), "no network") {
		t.Errorf("Connect = %v; want the setup's error", err)
	}
	if db != nil {
		t.Error("Connect returned a pool with the error")
	}
}

func TestDuckDBPermittedDatabaseFile(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)

	dbFile := filepath.Join(f.in, "test.duckdb")
	db, err := connectDuckDB(t, DuckDBConnectionStringPrefix+dbFile, 2)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE t AS SELECT 42 AS v"); err != nil {
		t.Fatalf("create: %v", err)
	}
	var v int
	if err := db.QueryRow("SELECT v FROM t").Scan(&v); err != nil || v != 42 {
		t.Fatalf("query = %d, %v; want 42", v, err)
	}
	// a file in the folder can be attached too, but not one in the sibling folder
	if _, err := db.Exec("ATTACH " + q(filepath.Join(f.in, "other.duckdb")) + " AS other"); err != nil {
		t.Errorf("ATTACH inside: %v", err)
	}
	mustFail(t, db, "ATTACH in the sibling", "ATTACH "+q(filepath.Join(f.sibling, "other.duckdb"))+" AS sib")
	assertLocked(t, db, f)
}

func TestDuckDBPermittedValueCannotLift(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)

	for _, dsn := range []string{
		"duckdb://?enable_external_access=true",
		DuckDBConnectionStringPrefix + filepath.Join(f.in, "lift.duckdb") + "?enable_external_access=true",
	} {
		db, err := connectDuckDB(t, dsn, 2)
		if err != nil {
			t.Fatalf("Connect(%s): %v", dsn, err)
		}
		assertLocked(t, db, f)
	}

	for _, dsn := range []string{
		"duckdb://?lock_configuration=true",
		DuckDBConnectionStringPrefix + filepath.Join(f.in, "prelocked.duckdb") + "?lock_configuration=true",
		// DuckDB adds its temp folder to allowed_directories, so one outside would open that folder
		"duckdb://?temp_directory=" + f.out,
	} {
		db, err := connectDuckDB(t, dsn, 2)
		if err == nil || !strings.Contains(err.Error(), testPermittedEnv) {
			t.Errorf("Connect(%s) = %v; want an error naming the setting", dsn, err)
		}
		if db != nil {
			t.Errorf("Connect(%s): returned a pool with the error", dsn)
		}
	}
}

func connectDuckDBInit(t *testing.T, script string) (*sql.DB, error) {
	t.Helper()
	b, err := NewDuckDBInitBackend(DuckDBInitConnectionStringPrefix + script)
	if err != nil {
		t.Fatal(err)
	}
	db, err := b.Connect(context.Background())
	if db != nil {
		t.Cleanup(func() { _ = db.Close() })
	}
	return db, err
}

func TestDuckDBPermittedInit(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)

	script := filepath.Join(f.in, "init.sql")
	content := "CREATE TABLE t AS SELECT * FROM read_csv(" + q(filepath.Join(f.in, "a.csv")) + ");\n"
	if err := os.WriteFile(script, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	db, err := connectDuckDBInit(t, script)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	var n int
	if err := db.QueryRow("SELECT count(*) FROM t").Scan(&n); err != nil || n != 1 {
		t.Fatalf("table from the script = %d, %v; want 1 row", n, err)
	}
	if open := db.Stats().MaxOpenConnections; open != 1 {
		t.Errorf("MaxOpenConnections = %d, want 1", open)
	}
	assertLocked(t, db, f)

	bad := filepath.Join(f.in, "bad.sql")
	content = "CREATE TABLE t AS SELECT * FROM read_text(" + q(filepath.Join(f.out, "b.csv")) + ");\n"
	if err := os.WriteFile(bad, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	db, err = connectDuckDBInit(t, bad)
	if err == nil {
		t.Error("Connect with a script reading outside the folder: want an error")
	}
	if db != nil {
		t.Error("Connect returned a pool with the error")
	}
}

func TestDuckDBPermittedUnset(t *testing.T) {
	f := newDuckDBTest(t)
	if err := setPermittedEnv(t, ""); err != nil {
		t.Fatal(err)
	}

	db, err := connectDuckDB(t, "duckdb://", 2)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	mustRead(t, db, filepath.Join(f.out, "b.csv"))
	var external bool
	if err := db.QueryRow("SELECT current_setting('enable_external_access')").Scan(&external); err != nil || !external {
		t.Errorf("enable_external_access = %v, %v; want true", external, err)
	}

	script := filepath.Join(f.out, "init.sql")
	content := "CREATE TABLE t AS SELECT * FROM read_csv(" + q(filepath.Join(f.out, "b.csv")) + ");\n"
	if err := os.WriteFile(script, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	initDB, err := connectDuckDBInit(t, script)
	if err != nil {
		t.Fatalf("duckdbinit Connect: %v", err)
	}
	var n int
	if err := initDB.QueryRow("SELECT count(*) FROM t").Scan(&n); err != nil || n != 1 {
		t.Errorf("table from the script = %d, %v; want 1 row", n, err)
	}
}

func TestDuckDBPermittedBadSetting(t *testing.T) {
	newDuckDBTest(t)
	if err := setPermittedEnv(t, "relative/folder"); err == nil {
		t.Fatal("want the relative entry refused")
	}

	db, err := connectDuckDB(t, "duckdb://", 2)
	if err == nil || !strings.Contains(err.Error(), testPermittedEnv) {
		t.Errorf("Connect = %v; want an error naming the setting", err)
	}
	if db != nil {
		t.Error("Connect returned a pool with the error")
	}
	db, err = connectDuckDBInit(t, "")
	if err == nil {
		t.Error("duckdbinit Connect: want an error")
	}
	if db != nil {
		t.Error("duckdbinit Connect returned a pool with the error")
	}
}

// TestDuckDBPermittedSymlinks shows that DuckDB, which matches allowed_directories as text and follows
// symlinks, opens only when no symlink in the folders leads outside them
func TestDuckDBPermittedSymlinks(t *testing.T) {
	type row struct {
		name string
		// links maps a symlink, relative to the permitted folder, to its target (made from the folders)
		links func(f duckDBTestFolders) map[string]string
		// escape is a file, relative to the permitted folder, that would reach outside through a link
		escape string
		// wantErr names the symlink the open refuses; "" means the open works and stays locked
		wantErr string
	}
	rows := []row{
		{name: "folder outside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"linkdir": f.out}
		}, escape: "linkdir/b.csv", wantErr: "linkdir"},
		{name: "relative folder outside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"rel": filepath.Join("..", "out")}
		}, escape: "rel/b.csv", wantErr: "rel"},
		{name: "nested file outside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"sub/deeper/b.csv": filepath.Join(f.out, "b.csv")}
		}, escape: "sub/deeper/b.csv", wantErr: "b.csv"},
		{name: "missing file outside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"new.csv": filepath.Join(f.out, "new.csv")}
		}, wantErr: "new.csv"},
		{name: "the sibling folder", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"sib": f.sibling}
		}, escape: "sib/c.csv", wantErr: "sib"},
		{name: "folder inside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"inlink": filepath.Join(f.in, "sub")}
		}},
		{name: "file inside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"alias.csv": filepath.Join(f.in, "a.csv")}
		}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			f := newDuckDBTest(t)
			if err := os.MkdirAll(filepath.Join(f.in, "sub", "deeper"), 0o755); err != nil {
				t.Fatal(err)
			}
			for link, target := range r.links(f) {
				if err := os.Symlink(target, filepath.Join(f.in, link)); err != nil {
					t.Fatal(err)
				}
			}
			permitFolders(t, f.in)

			db, err := connectDuckDB(t, "duckdb://", 2)
			if r.wantErr == "" {
				if err != nil {
					t.Fatalf("Connect: %v", err)
				}
				assertLocked(t, db, f)
				return
			}
			if err == nil {
				if r.escape != "" {
					var n int
					if qErr := db.QueryRow("SELECT count(*) FROM read_csv(" + q(filepath.Join(f.in, r.escape)) + ")").Scan(&n); qErr == nil {
						t.Errorf("read %s outside the folder through a symlink", r.escape)
					}
				}
				t.Fatalf("Connect: want an error naming %q and %s, got none", r.wantErr, testPermittedEnv)
			}
			if !strings.Contains(err.Error(), r.wantErr) || !strings.Contains(err.Error(), testPermittedEnv) {
				t.Errorf("Connect = %v; want an error naming %q and %s", err, r.wantErr, testPermittedEnv)
			}
			if db != nil {
				t.Error("Connect returned a pool with the error")
			}
			// duckdbinit opens through the same lock
			initDB, err := connectDuckDBInit(t, "")
			if err == nil || !strings.Contains(err.Error(), r.wantErr) {
				t.Errorf("duckdbinit Connect = %v; want an error naming %q", err, r.wantErr)
			}
			if initDB != nil {
				t.Error("duckdbinit Connect returned a pool with the error")
			}
		})
	}

	t.Run("setting unset", func(t *testing.T) {
		f := newDuckDBTest(t)
		if err := os.Symlink(f.out, filepath.Join(f.in, "linkdir")); err != nil {
			t.Fatal(err)
		}
		if err := setPermittedEnv(t, ""); err != nil {
			t.Fatal(err)
		}
		db, err := connectDuckDB(t, "duckdb://", 2)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		mustRead(t, db, filepath.Join(f.in, "linkdir", "b.csv"))
	})
}
