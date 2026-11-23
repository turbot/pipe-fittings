//go:build cgo

package backend

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
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

	// DuckDB keeps persistent secrets and extensions under HOME: never the real one
	t.Setenv("HOME", t.TempDir())

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
		// DuckDB adds its temp folder to allowed_directories, so one outside would open that folder: the lock
		// puts its own in its place
		"duckdb://?temp_directory=" + f.out,
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

// duckDBDefaultTempDir is the temp_directory DuckDB uses for an in-memory database: .tmp in the working folder
const duckDBDefaultTempDir = ".tmp"

// TestDuckDBPermittedTempDir shows that an in-memory DuckDB, whose temp folder is by default .tmp in the
// working folder, keeps its temp folder inside the permitted folders when the working folder lies outside
// them, so neither .tmp nor a symlink there reaches outside
func TestDuckDBPermittedTempDir(t *testing.T) {
	for _, linked := range []bool{false, true} {
		for _, backend := range []string{"duckdb", "duckdbinit"} {
			name := backend + " with .tmp a folder"
			if linked {
				name = backend + " with .tmp a symlink outside"
			}
			t.Run(name, func(t *testing.T) {
				f := newDuckDBTest(t)
				work := filepath.Join(filepath.Dir(f.in), "work")
				if err := os.MkdirAll(work, 0o755); err != nil {
					t.Fatal(err)
				}
				// where .tmp leads: f.out itself through the symlink, else a real folder outside
				tmpTarget := filepath.Join(work, ".tmp")
				if linked {
					if err := os.Symlink(f.out, tmpTarget); err != nil {
						t.Fatal(err)
					}
					tmpTarget = f.out
				} else {
					if err := os.MkdirAll(tmpTarget, 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(tmpTarget, "b.csv"), []byte("x\n1\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				t.Chdir(work)
				permitFolders(t, f.in)

				var db *sql.DB
				var err error
				if backend == "duckdb" {
					db, err = connectDuckDB(t, "duckdb://", 2)
				} else {
					db, err = connectDuckDBInit(t, "")
				}
				if err != nil {
					t.Fatalf("Connect: %v", err)
				}
				mustFail(t, db, "read_csv in .tmp", "SELECT * FROM read_csv('.tmp/b.csv')")
				mustFail(t, db, "COPY TO .tmp", "COPY (SELECT 1 AS x) TO '.tmp/plain.csv'")
				if _, err := os.Stat(filepath.Join(tmpTarget, "plain.csv")); err == nil {
					t.Errorf("COPY TO wrote %s", filepath.Join(tmpTarget, "plain.csv"))
				}
				var tempDir string
				if err := db.QueryRow("SELECT current_setting('temp_directory')").Scan(&tempDir); err != nil {
					t.Fatal(err)
				}
				if !isPermitted(tempDir) {
					t.Errorf("temp_directory = %q, want a folder inside %s", tempDir, f.in)
				}
				assertLocked(t, db, f)
			})
		}
	}

	t.Run("setting unset", func(t *testing.T) {
		f := newDuckDBTest(t)
		work := filepath.Join(filepath.Dir(f.in), "work")
		if err := os.MkdirAll(work, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(f.out, filepath.Join(work, ".tmp")); err != nil {
			t.Fatal(err)
		}
		t.Chdir(work)
		if err := setPermittedEnv(t, ""); err != nil {
			t.Fatal(err)
		}
		db, err := connectDuckDB(t, "duckdb://", 2)
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		var tempDir string
		if err := db.QueryRow("SELECT current_setting('temp_directory')").Scan(&tempDir); err != nil || tempDir != duckDBDefaultTempDir {
			t.Errorf("temp_directory = %q, %v; want DuckDB's default %q", tempDir, err, duckDBDefaultTempDir)
		}
		mustRead(t, db, filepath.Join(".tmp", "b.csv"))
	})
}

func noLinks(duckDBTestFolders) map[string]string { return nil }

// TestDuckDBPermittedSymlinks shows that DuckDB, which matches allowed_directories as text and follows
// symlinks, opens only when no symlink in the folders leads outside them
func TestDuckDBPermittedSymlinks(t *testing.T) {
	type row struct {
		name string
		// links maps a symlink, relative to the permitted folder, to its target (made from the folders)
		links func(f duckDBTestFolders) map[string]string
		// modes sets, after the links are made, the mode of each folder (relative to the permitted folder, made
		// first if missing); each is given back 0755 when the test ends so its temp folder can be removed
		modes map[string]os.FileMode
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
		// DuckDB cannot read through a folder this process cannot enter, or through a symlink loop, so neither
		// stops the open
		{name: "unreadable subfolder", links: noLinks, modes: map[string]os.FileMode{"sub/locked": 0}},
		{name: "symlink loop", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"loopa": "loopb", "loopb": "loopa"}
		}},
		// a folder that can be entered but not listed may hold a symlink DuckDB can reach by name
		{name: "searchable but unlistable folder", links: noLinks, modes: map[string]os.FileMode{"sub/hidden": 0o111}, wantErr: "hidden"},
		// the walk goes on past a skipped folder: sub/locked comes before sub/zlink
		{name: "outward symlink beside an unreadable subfolder", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"sub/zlink": f.out}
		}, modes: map[string]os.FileMode{"sub/locked": 0}, escape: "sub/zlink/b.csv", wantErr: "zlink"},
		{name: "loop through a folder outside", links: func(f duckDBTestFolders) map[string]string {
			return map[string]string{"loopa": filepath.Join(f.out, "loopc"), "../out/loopc": filepath.Join(f.in, "loopa")}
		}, wantErr: "loopa"},
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
			for dir, mode := range r.modes {
				path := filepath.Join(f.in, dir)
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(path, mode); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
			}
			if len(r.modes) > 0 && os.Geteuid() == 0 {
				t.Skip("root enters and lists every folder")
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

// TestDuckDBPermittedPersistentSecrets shows that mod SQL cannot write, replace or drop DuckDB's persistent
// secrets, which DuckDB keeps outside the folders in <HOME>/.duckdb/stored_secrets. HOME is a temp folder.
func TestDuckDBPermittedPersistentSecrets(t *testing.T) {
	f := newDuckDBTest(t)
	secrets := filepath.Join(os.Getenv("HOME"), ".duckdb", "stored_secrets")
	if err := os.MkdirAll(secrets, 0o700); err != nil {
		t.Fatal(err)
	}
	// a secret the user stored before
	const stored = "users_secret"
	if err := setPermittedEnv(t, ""); err != nil {
		t.Fatal(err)
	}
	before, err := connectDuckDB(t, "duckdb://", 1)
	if err != nil {
		t.Fatalf("Connect without the setting: %v", err)
	}
	if _, err := before.Exec("CREATE PERSISTENT SECRET " + stored + " (TYPE http, BEARER_TOKEN 'keep')"); err != nil {
		t.Fatalf("storing the user's secret: %v", err)
	}
	_ = before.Close()
	storedFile := filepath.Join(secrets, stored+".duckdb_secret")
	want, err := os.ReadFile(storedFile)
	if err != nil {
		t.Fatalf("the user's secret was not stored: %v", err)
	}

	permitFolders(t, f.in)
	db, err := connectDuckDB(t, "duckdb://", 2)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	for _, stmt := range []string{
		"CREATE PERSISTENT SECRET mod_secret (TYPE http, BEARER_TOKEN 'x')",
		"CREATE OR REPLACE PERSISTENT SECRET " + stored + " (TYPE http, BEARER_TOKEN 'replaced')",
		"DROP PERSISTENT SECRET " + stored,
		"SET allow_persistent_secrets = true",
		"SET secret_directory = " + q(f.in),
	} {
		if _, err := db.Exec(stmt); err == nil {
			t.Errorf("%s: want an error, got none", stmt)
		}
	}
	if _, err := os.Stat(filepath.Join(secrets, "mod_secret.duckdb_secret")); err == nil {
		t.Error("mod SQL wrote a persistent secret")
	}
	if got, err := os.ReadFile(storedFile); err != nil || string(got) != string(want) {
		t.Errorf("the user's stored secret was changed or removed: %v", err)
	}
	assertLocked(t, db, f)
}

// TestDuckDBPermittedSymlinkWalkVanish shows that the symlink search goes on past a folder that vanishes while
// it runs, as DuckDB's own temp folder does when another database closes: the folder is removed after the walk
// has listed it and before it reads it, so the read fails with "not exist". An outward symlink after it is
// still found.
func TestDuckDBPermittedSymlinkWalkVanish(t *testing.T) {
	f := newDuckDBTest(t)
	permitFolders(t, f.in)
	resolved, err := filepath.EvalSymlinks(f.in)
	if err != nil {
		t.Fatal(err)
	}
	vanishing := filepath.Join(resolved, duckDBTempDirName, "123-spill")
	if err := os.MkdirAll(filepath.Join(vanishing, "deeper"), 0o700); err != nil {
		t.Fatal(err)
	}
	walk := func(t *testing.T) error {
		removed := false
		search := duckDBSymlinkWalk(resolved)
		err := filepath.WalkDir(resolved, func(path string, d fs.DirEntry, err error) error {
			if path == vanishing && err == nil && !removed {
				// WalkDir reads a folder's entries after this call returns
				if rmErr := os.RemoveAll(vanishing); rmErr != nil {
					t.Fatal(rmErr)
				}
				removed = true
			}
			return search(path, d, err)
		})
		if !removed {
			t.Fatal("the walk never reached the vanishing folder")
		}
		return err
	}
	if err := walk(t); err != nil {
		t.Errorf("walk with a folder vanishing = %v; want none", err)
	}

	if err := os.MkdirAll(filepath.Join(vanishing, "deeper"), 0o700); err != nil {
		t.Fatal(err)
	}
	// "zlink" sorts after ".duckdb_tmp", so the walk reaches it after the folder vanished
	if err := os.Symlink(f.out, filepath.Join(resolved, "zlink")); err != nil {
		t.Fatal(err)
	}
	if err := walk(t); err == nil || !strings.Contains(err.Error(), "zlink") {
		t.Errorf("walk with a folder vanishing before an outward symlink = %v; want an error naming zlink", err)
	}
}

// TestDuckDBPermittedTempDirReadOnlyRoot shows that a first permitted folder this process cannot write to does
// not stop DuckDB opening when a later one is writable: the temp folder goes in the writable one, and the open
// stays locked. Only when no folder is writable does the open fail.
func TestDuckDBPermittedTempDirReadOnlyRoot(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes to every folder")
	}
	f := newDuckDBTest(t)
	readOnly := filepath.Join(filepath.Dir(f.in), "readonly")
	if err := os.MkdirAll(readOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(readOnly, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0o755) })

	permitFolders(t, readOnly, f.in)
	for _, backend := range []string{"duckdb", "duckdbinit"} {
		t.Run(backend, func(t *testing.T) {
			var db *sql.DB
			var err error
			if backend == "duckdb" {
				db, err = connectDuckDB(t, "duckdb://", 2)
			} else {
				db, err = connectDuckDBInit(t, "")
			}
			if err != nil {
				t.Fatalf("Connect: %v", err)
			}
			var tempDir string
			if err := db.QueryRow("SELECT current_setting('temp_directory')").Scan(&tempDir); err != nil {
				t.Fatal(err)
			}
			in, err := filepath.EvalSymlinks(f.in)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(tempDir, filepath.Join(f.in, duckDBTempDirName)) && !strings.HasPrefix(tempDir, filepath.Join(in, duckDBTempDirName)) {
				t.Errorf("temp_directory = %q, want a folder under %s", tempDir, filepath.Join(f.in, duckDBTempDirName))
			}
			assertLocked(t, db, f)
		})
	}

	t.Run("no folder writable", func(t *testing.T) {
		permitFolders(t, readOnly)
		db, err := connectDuckDB(t, "duckdb://", 2)
		if err == nil || !strings.Contains(err.Error(), testPermittedEnv) {
			t.Errorf("Connect = %v; want an error naming the setting", err)
		}
		if db != nil {
			t.Error("Connect returned a pool with the error")
		}
	})
}
