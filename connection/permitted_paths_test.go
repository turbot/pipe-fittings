package connection

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
)

const (
	testAppPrefix    = "TESTAPP_"
	testPermittedEnv = testAppPrefix + "PERMITTED_ROOT_PATHS"
)

// setPermittedForTest sets the app's env var keys to a test-only prefix, as an app does at start-up, and
// sets the setting to value (value "" unsets it)
func setPermittedForTest(t *testing.T, value string) {
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
	if err := permitted_paths.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionPermitted(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	permitted := filepath.Join(base, "permitted")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{permitted, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	in := func(name string) string { return filepath.Join(permitted, name) }
	out := func(name string) string { return filepath.Join(outside, name) }
	ptr := func(s string) *string { return &s }
	const secret = "s3cretpw"
	pg := func(query string) string {
		return "postgres://alice:" + secret + "@db.example.com:5432/app?" + query
	}
	declRange := hcl.Range{Filename: "conn.ppc"}

	type row struct {
		name string
		conn func() ConnectionStringProvider
		// want is the string returned; wantErr is a fragment of the error, "" for none
		want    string
		wantErr string
		// unset is the string returned with the setting unset
		unset string
	}
	duck := func(f func(c *DuckDbConnection)) func() ConnectionStringProvider {
		return func() ConnectionStringProvider {
			c := NewDuckDbConnection("d", declRange).(*DuckDbConnection)
			f(c)
			return c
		}
	}
	sqlite := func(f func(c *SqliteConnection)) func() ConnectionStringProvider {
		return func() ConnectionStringProvider {
			c := NewSqliteConnection("s", declRange).(*SqliteConnection)
			f(c)
			return c
		}
	}
	postgres := func(f func(c *PostgresConnection)) func() ConnectionStringProvider {
		return func() ConnectionStringProvider {
			c := NewPostgresConnection("p", declRange).(*PostgresConnection)
			f(c)
			return c
		}
	}
	mysql := func(f func(c *MysqlConnection)) func() ConnectionStringProvider {
		return func() ConnectionStringProvider {
			c := NewMysqlConnection("m", declRange).(*MysqlConnection)
			f(c)
			return c
		}
	}

	rows := []row{
		// DuckDB, by its fields and by its connection_string
		{name: "duckdb file_name outside", conn: duck(func(c *DuckDbConnection) { c.FileName = ptr(out("a.db")) }),
			wantErr: "outside the folders " + testPermittedEnv, unset: "duckdb://" + out("a.db")},
		{name: "duckdb file_name inside", conn: duck(func(c *DuckDbConnection) { c.FileName = ptr(in("sub/../a.db")) }),
			want: "duckdb://" + in("a.db"), unset: "duckdb://" + in("sub/../a.db")},
		{name: "duckdb connection_string outside", conn: duck(func(c *DuckDbConnection) { c.ConnectionString = ptr("duckdb://" + out("a.db")) }),
			wantErr: "outside the folders", unset: "duckdb://" + out("a.db")},
		{name: "duckdb connection_string inside", conn: duck(func(c *DuckDbConnection) { c.ConnectionString = ptr("duckdb://" + in("sub/../a.db")) }),
			want: "duckdb://" + in("a.db"), unset: "duckdb://" + in("sub/../a.db")},
		{name: "duckdb connection_string options", conn: duck(func(c *DuckDbConnection) { c.ConnectionString = ptr("duckdb://" + in("a.db") + "?threads=1") }),
			wantErr: `option "?…"`, unset: "duckdb://" + in("a.db") + "?threads=1"},
		{name: "duckdb connection_string duckdbinit", conn: duck(func(c *DuckDbConnection) { c.ConnectionString = ptr("duckdbinit://" + in("init.sql")) }),
			wantErr: "duckdbinit://", unset: "duckdbinit://" + in("init.sql")},
		// SQLite
		{name: "sqlite file_name outside", conn: sqlite(func(c *SqliteConnection) { c.FileName = ptr(out("a.sqlite")) }),
			wantErr: "outside the folders " + testPermittedEnv, unset: "sqlite://" + out("a.sqlite")},
		{name: "sqlite file_name inside", conn: sqlite(func(c *SqliteConnection) { c.FileName = ptr(in("sub/../a.sqlite")) }),
			want: "sqlite://" + in("a.sqlite"), unset: "sqlite://" + in("sub/../a.sqlite")},
		{name: "sqlite connection_string outside", conn: sqlite(func(c *SqliteConnection) { c.ConnectionString = ptr("sqlite://" + out("a.sqlite")) }),
			wantErr: "outside the folders", unset: "sqlite://" + out("a.sqlite")},
		{name: "sqlite connection_string inside", conn: sqlite(func(c *SqliteConnection) { c.ConnectionString = ptr("sqlite://" + in("sub/../a.sqlite")) }),
			want: "sqlite://" + in("a.sqlite"), unset: "sqlite://" + in("sub/../a.sqlite")},
		// Postgres
		{name: "postgres fields", conn: postgres(func(c *PostgresConnection) {
			c.Host, c.DbName, c.UserName = ptr("db.example.com"), ptr("app"), ptr("alice")
		}), want: "postgresql://alice@db.example.com:5432/app", unset: "postgresql://alice@db.example.com:5432/app"},
		{name: "postgres sslkey outside", conn: postgres(func(c *PostgresConnection) { c.ConnectionString = ptr(pg("sslkey=" + out("k.pem"))) }),
			wantErr: `option "sslkey"`, unset: pg("sslkey=" + out("k.pem"))},
		{name: "postgres sslkey inside", conn: postgres(func(c *PostgresConnection) { c.ConnectionString = ptr(pg("sslkey=" + in("sub/../k.pem"))) }),
			want: pg("sslkey=" + url.QueryEscape(in("k.pem"))), unset: pg("sslkey=" + in("sub/../k.pem"))},
		// MySQL
		{name: "mysql fields", conn: mysql(func(c *MysqlConnection) {
			c.Host, c.DbName, c.UserName, c.Password = ptr("db.example.com"), ptr("app"), ptr("alice"), ptr(secret)
		}), want: "mysql://alice:" + secret + "@tcp(db.example.com:3306)/app", unset: "mysql://alice:" + secret + "@tcp(db.example.com:3306)/app"},
		{name: "mysql connection_string allowAllFiles", conn: mysql(func(c *MysqlConnection) {
			c.ConnectionString = ptr("mysql://alice:" + secret + "@tcp(db.example.com:3306)/app?allowAllFiles=true")
		}), wantErr: `"allowAllFiles"`, unset: "mysql://alice:" + secret + "@tcp(db.example.com:3306)/app?allowAllFiles=true"},
		{name: "mysql connection_string without it", conn: mysql(func(c *MysqlConnection) {
			c.ConnectionString = ptr("mysql://alice@tcp(db.example.com:3306)/app?parseTime=true")
		}), want: "mysql://alice@tcp(db.example.com:3306)/app?parseTime=true", unset: "mysql://alice@tcp(db.example.com:3306)/app?parseTime=true"},
	}

	t.Run("setting set", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		for _, r := range rows {
			t.Run(r.name, func(t *testing.T) {
				got, err := r.conn().GetConnectionString()
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
		setPermittedForTest(t, "")
		for _, r := range rows {
			t.Run(r.name, func(t *testing.T) {
				got, err := r.conn().GetConnectionString()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != r.unset {
					t.Fatalf("got %q, want %q unchanged", got, r.unset)
				}
			})
		}
	})
}
