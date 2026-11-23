package backend

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
)

const (
	testAppPrefix    = "TESTAPP_"
	testPermittedEnv = testAppPrefix + "PERMITTED_ROOT_PATHS"
)

// setPermittedEnv sets the app's env var keys to a test-only prefix, as an app does at start-up, sets the
// setting to value (value "" unsets it) and reads it, as the app does once at start-up
func setPermittedEnv(t *testing.T, value string) error {
	t.Helper()
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys(testAppPrefix)
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prev)
		permitted_paths.Reset()
	})
	t.Setenv(testPermittedEnv, value)
	if value == "" {
		if err := os.Unsetenv(testPermittedEnv); err != nil {
			t.Fatal(err)
		}
	}
	return permitted_paths.Load()
}

// permitFolders limits the readers to the folders
func permitFolders(t *testing.T, folders ...string) {
	t.Helper()
	if err := setPermittedEnv(t, strings.Join(folders, ",")); err != nil {
		t.Fatal(err)
	}
}

// permittedFixture is
//
//	<base>/permitted/                the one permitted folder
//	<base>/permitted/linkdir -> ../outside   a symlink inside it that points outside
//	<base>/outside/                  a folder that is not permitted
type permittedFixture struct {
	// base holds both folders, with its symlinks resolved (a temp dir can sit behind one)
	base      string
	permitted string
	outside   string
}

func newPermittedFixture(t *testing.T) permittedFixture {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := permittedFixture{base: base, permitted: filepath.Join(base, "permitted"), outside: filepath.Join(base, "outside")}
	for _, dir := range []string{f.permitted, f.outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join("..", "outside"), filepath.Join(f.permitted, "linkdir")); err != nil {
		t.Fatal(err)
	}
	return f
}

// rel is the path of target relative to the working directory, so a row can test a relative path
func (f permittedFixture) rel(t *testing.T, target string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(wd, target)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

func TestConfineModConnectionString(t *testing.T) {
	f := newPermittedFixture(t)
	in := func(name string) string { return filepath.Join(f.permitted, name) }
	out := func(name string) string { return filepath.Join(f.outside, name) }
	esc := url.QueryEscape
	const secret = "s3cretpw"
	pg := func(query string) string {
		return "postgres://alice:" + secret + "@db.example.com:5432/app?" + query
	}

	type row struct {
		name string
		in   string
		// want is the string returned; wantErr is a fragment of the error, "" for none
		want    string
		wantErr string
	}
	rows := []row{
		// DuckDB
		{name: "duckdb inside", in: "duckdb://" + in("a.db"), want: "duckdb://" + in("a.db")},
		{name: "duckdb inside through a dot segment", in: "duckdb://" + in("sub/../a.db"), want: "duckdb://" + in("a.db")},
		{name: "duckdb outside", in: "duckdb://" + out("a.db"), wantErr: "outside the folders " + testPermittedEnv},
		{name: "duckdb through an outward symlink", in: "duckdb://" + in("linkdir/a.db"), wantErr: "outside the folders"},
		{name: "duckdb relative inside", in: "duckdb://" + f.rel(t, in("a.db")), want: "duckdb://" + in("a.db")},
		{name: "duckdb relative outside", in: "duckdb://" + f.rel(t, out("a.db")), wantErr: "outside the folders"},
		{name: "duckdb options", in: "duckdb://" + in("a.db") + "?access_mode=read_only", wantErr: `option "?…"`},
		{name: "duckdb in-memory empty", in: "duckdb://", want: "duckdb://"},
		{name: "duckdb in-memory", in: "duckdb://:memory:", want: "duckdb://:memory:"},
		{name: "duckdb in-memory named", in: "duckdb://:memory:work", want: "duckdb://:memory:work"},
		{name: "duckdb in-memory with options", in: "duckdb://:memory:?threads=1", wantErr: `option "?…"`},
		{name: "duckdb motherduck", in: "duckdb://md:mydb", wantErr: "not a plain file path"},
		{name: "duckdb file uri", in: "duckdb://file:" + in("a.db"), wantErr: "not a plain file path"},
		{name: "duckdb s3", in: "duckdb://s3://bucket/a.db", wantErr: "not a plain file path"},
		{name: "duckdb leading space", in: " duckdb://" + out("a.db"), wantErr: "outside the folders"},
		{name: "duckdb leading space inside", in: " duckdb://" + in("a.db") + " ", want: "duckdb://" + in("a.db")},
		// SQLite
		{name: "sqlite inside", in: "sqlite://" + in("a.sqlite"), want: "sqlite://" + in("a.sqlite")},
		{name: "sqlite outside", in: "sqlite://" + out("a.sqlite"), wantErr: "outside the folders " + testPermittedEnv},
		{name: "sqlite through an outward symlink", in: "sqlite://" + in("linkdir/a.sqlite"), wantErr: "outside the folders"},
		{name: "sqlite relative inside", in: "sqlite://" + f.rel(t, in("a.sqlite")), want: "sqlite://" + in("a.sqlite")},
		{name: "sqlite relative outside", in: "sqlite://" + f.rel(t, out("a.sqlite")), wantErr: "outside the folders"},
		{name: "sqlite options", in: "sqlite://" + in("a.sqlite") + "?mode=ro", wantErr: `option "?…"`},
		{name: "sqlite in-memory", in: "sqlite://:memory:", want: "sqlite://:memory:"},
		{name: "sqlite in-memory named", in: "sqlite://:memory:work", want: "sqlite://:memory:work"},
		{name: "sqlite file uri", in: "sqlite://file:" + in("a.sqlite"), wantErr: "not a plain file path"},
		// duckdbinit
		{name: "duckdbinit", in: "duckdbinit://" + in("init.sql"), wantErr: "duckdbinit://"},
		{name: "duckdbinit empty", in: "duckdbinit://", wantErr: "duckdbinit://"},
		// Postgres: each file option inside (rewritten) and outside
		{name: "postgres no options", in: pg("sslmode=require"), want: pg("sslmode=require")},
		{name: "postgres no query", in: "postgresql://alice@localhost/app", want: "postgresql://alice@localhost/app"},
		{name: "postgres other options pass", in: pg("sslmode=require&application_name=a%20b"), want: pg("sslmode=require&application_name=a%20b")},
		{name: "postgres sslcert inside", in: pg("sslcert=" + in("c.pem")), want: pg("sslcert=" + esc(in("c.pem")))},
		{name: "postgres sslkey inside", in: pg("sslkey=" + in("k.pem")), want: pg("sslkey=" + esc(in("k.pem")))},
		{name: "postgres sslrootcert inside", in: pg("sslrootcert=" + in("r.pem")), want: pg("sslrootcert=" + esc(in("r.pem")))},
		{name: "postgres passfile inside", in: pg("passfile=" + in("pgpass")), want: pg("passfile=" + esc(in("pgpass")))},
		{name: "postgres servicefile inside", in: pg("servicefile=" + in("svc")), want: pg("servicefile=" + esc(in("svc")))},
		{name: "postgres sslcert outside", in: pg("sslcert=" + out("c.pem")), wantErr: `option "sslcert"`},
		{name: "postgres sslkey outside", in: pg("sslkey=" + out("k.pem")), wantErr: `option "sslkey"`},
		{name: "postgres sslrootcert outside", in: pg("sslrootcert=" + out("r.pem")), wantErr: `option "sslrootcert"`},
		{name: "postgres passfile outside", in: pg("passfile=" + out("pgpass")), wantErr: `option "passfile"`},
		{name: "postgres servicefile outside", in: pg("servicefile=" + out("svc")), wantErr: `option "servicefile"`},
		{name: "postgres escaped value outside", in: pg("sslkey=" + esc(out("k.pem"))), wantErr: `option "sslkey"`},
		{name: "postgres through an outward symlink", in: pg("sslkey=" + in("linkdir/k.pem")), wantErr: "outside the folders"},
		{name: "postgres upper-case key outside", in: pg("SSLKEY=" + out("k.pem")), wantErr: `option "SSLKEY"`},
		{name: "postgres upper-case key inside", in: pg("SslKey=" + in("k.pem")), want: pg("SslKey=" + esc(in("k.pem")))},
		{name: "postgres repeated key one outside", in: pg("sslkey=" + in("k.pem") + "&sslkey=" + out("k.pem")), wantErr: `option "sslkey"`},
		{name: "postgres repeated key all inside", in: pg("sslkey=" + in("k.pem") + "&sslkey=" + in("k2.pem")), want: pg("sslkey=" + esc(in("k.pem")) + "&sslkey=" + esc(in("k2.pem")))},
		{name: "postgres empty option", in: pg("sslkey=&sslmode=require"), want: pg("sslkey=&sslmode=require")},
		{name: "postgres unparseable", in: "postgres://alice:" + secret + "@db.example.com:bad/app", wantErr: "cannot be parsed"},
		// MySQL
		{name: "mysql plain", in: "mysql://root:" + secret + "@tcp(localhost:3306)/app", want: "mysql://root:" + secret + "@tcp(localhost:3306)/app"},
		{name: "mysql other options", in: "mysql://root@tcp(localhost:3306)/app?parseTime=true", want: "mysql://root@tcp(localhost:3306)/app?parseTime=true"},
		{name: "mysql allowAllFiles true", in: "mysql://root@tcp(localhost:3306)/app?allowAllFiles=true", wantErr: `"allowAllFiles"`},
		{name: "mysql allowAllFiles false", in: "mysql://root@tcp(localhost:3306)/app?allowAllFiles=false", wantErr: `"allowAllFiles"`},
		{name: "mysql ALLOWALLFILES 1", in: "mysql://root@tcp(localhost:3306)/app?ALLOWALLFILES=1", wantErr: `"allowAllFiles"`},
		{name: "mysql allowAllFiles among others", in: "mysql://root@tcp(localhost:3306)/app?parseTime=true&allowAllFiles=true", wantErr: `"allowAllFiles"`},
		// no known prefix
		{name: "no prefix", in: "somethingelse", want: "somethingelse"},
		{name: "empty", in: "", want: ""},
	}

	t.Run("setting set", func(t *testing.T) {
		permitFolders(t, f.permitted)
		for _, r := range rows {
			t.Run(r.name, func(t *testing.T) {
				got, err := ConfineModConnectionString(r.in)
				if r.wantErr != "" {
					if err == nil {
						t.Fatalf("want an error containing %q, got %q", r.wantErr, got)
					}
					if !strings.Contains(err.Error(), r.wantErr) {
						t.Fatalf("want an error containing %q, got %q", r.wantErr, err)
					}
					if strings.Contains(err.Error(), secret) {
						t.Fatalf("the error leaks the password: %q", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != r.want {
					t.Fatalf("got %q, want %q", got, r.want)
				}
			})
		}
	})

	t.Run("setting unset", func(t *testing.T) {
		if err := setPermittedEnv(t, ""); err != nil {
			t.Fatal(err)
		}
		for _, r := range rows {
			t.Run(r.name, func(t *testing.T) {
				got, err := ConfineModConnectionString(r.in)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != r.in {
					t.Fatalf("got %q, want the input %q unchanged", got, r.in)
				}
			})
		}
	})
}

func TestConfineModConnectionStringOutsideError(t *testing.T) {
	f := newPermittedFixture(t)
	permitFolders(t, f.permitted)
	_, err := ConfineModConnectionString("duckdb://" + filepath.Join(f.outside, "a.db"))
	var outside *permitted_paths.OutsideError
	if !errors.As(err, &outside) {
		t.Fatalf("want an *OutsideError, got %v", err)
	}
	_, err = ConfineModConnectionString("postgres://u@h/db?sslkey=" + url.QueryEscape(filepath.Join(f.outside, "k")))
	if !errors.As(err, &outside) {
		t.Fatalf("want an *OutsideError, got %v", err)
	}
}

func TestConfineModConnectionStringBadSetting(t *testing.T) {
	// a relative entry makes the setting unreadable: every check fails closed
	if err := setPermittedEnv(t, "relative/folder"); err == nil {
		t.Fatal("want an error for a relative entry")
	}
	if _, err := ConfineModConnectionString("duckdb:///some/where.db"); err == nil {
		t.Fatal("want a duckdb value refused when the setting is unreadable")
	}
	if _, err := ConfineModConnectionString("postgres://u@h/db?sslkey=/some/key"); err == nil {
		t.Fatal("want a postgres file option refused when the setting is unreadable")
	}
}
