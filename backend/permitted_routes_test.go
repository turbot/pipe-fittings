//go:build cgo

package backend_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	_ "github.com/marcboeker/go-duckdb/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/app_specific_connection"
	"github.com/turbot/pipe-fittings/v2/backend"
	"github.com/turbot/pipe-fittings/v2/connection"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/zclconf/go-cty/cty"
)

const (
	routesAppPrefix    = "TESTAPP_"
	routesPermittedEnv = routesAppPrefix + "PERMITTED_ROOT_PATHS"
	routesFixture      = "../workspace/test_data/permitted_db_routes"
	outsidePlaceholder = "@OUTSIDE@"
)

// routesStubResources is the smallest ModResources a mod needs to be parsed: the mod block holds no resources
type routesStubResources struct{ modconfig.ModResources }

// routesDirs is
//
//	<base>/permitted/   a copy of the test mod: the one permitted folder, and the mod's folder
//	<base>/outside/     a folder that is not permitted
type routesDirs struct {
	permitted, outside string
}

func (d routesDirs) in(name string) string  { return filepath.Join(d.permitted, name) }
func (d routesDirs) out(name string) string { return filepath.Join(d.outside, name) }

// setPermittedForRoutes sets the app's env var keys to a test-only prefix, as an app does at start-up, sets the
// setting to value (value "" unsets it) and reads it, as the app does once at start-up
func setPermittedForRoutes(t *testing.T, value string) {
	t.Helper()
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys(routesAppPrefix)
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prev)
		permitted_paths.Reset()
	})
	t.Setenv(routesPermittedEnv, value)
	if value == "" {
		if err := os.Unsetenv(routesPermittedEnv); err != nil {
			t.Fatal(err)
		}
	}
	if err := permitted_paths.Load(); err != nil {
		t.Fatal(err)
	}
}

// newRoutesDirs copies the test mod into a temp folder (filling the outside placeholder in every file), makes
// that copy the permitted folder when permit is set, and puts a CSV and a file that is not an extension outside it
func newRoutesDirs(t *testing.T, permit bool) routesDirs {
	t.Helper()
	// DuckDB keeps persistent secrets and extensions under HOME: never the real one
	t.Setenv("HOME", t.TempDir())
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := routesDirs{permitted: filepath.Join(base, "permitted"), outside: filepath.Join(base, "outside")}
	for _, dir := range []string{d.permitted, d.outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(routesFixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		content, err := os.ReadFile(filepath.Join(routesFixture, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.ReplaceAll(string(content), outsidePlaceholder, d.outside))
		if err := os.WriteFile(d.in(e.Name()), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for file, content := range map[string]string{"data.csv": "x\n1\n", "evil.duckdb_extension": "not an extension"} {
		if err := os.WriteFile(d.out(file), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// no test needs the network: DuckDB's extension install is replaced, and its extension folder is a temp one
	prevInstall := filepaths.PipesInstallDir
	filepaths.PipesInstallDir = t.TempDir()
	extDir := t.TempDir()
	restore := backend.SetDuckDbExtensionSetupForTest(func(exec func(string) error) error {
		return exec("SET extension_directory = '" + extDir + "'")
	})
	t.Cleanup(func() {
		filepaths.PipesInstallDir = prevInstall
		restore()
	})
	// a mod's resources are not needed to read its database
	prevResources := modconfig.AppSpecificNewModResourcesFunc
	modconfig.AppSpecificNewModResourcesFunc = func(*modconfig.Mod, ...modconfig.ModResources) modconfig.ModResources {
		return routesStubResources{}
	}
	t.Cleanup(func() { modconfig.AppSpecificNewModResourcesFunc = prevResources })

	if permit {
		setPermittedForRoutes(t, d.permitted)
	} else {
		setPermittedForRoutes(t, "")
	}
	return d
}

// loadPpcConnections decodes the .ppc file in the mod's folder with the decoder the hosts use, and returns the
// `connection` variable of the eval context, as the hosts build it: connection.<type>.<name>
func loadPpcConnections(t *testing.T, d routesDirs) cty.Value {
	t.Helper()
	prevRegistry := app_specific_connection.ConnectionTypeRegistry
	app_specific_connection.ConnectionTypeRegistry = nil
	app_specific_connection.RegisterConnections(connection.NewDuckDbConnection, connection.NewSqliteConnection)
	t.Cleanup(func() { app_specific_connection.ConnectionTypeRegistry = prevRegistry })

	path := d.in("connections.ppc")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body, diags := parse.ParseConfig(content, path, hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}
	hclContent, diags := body.Content(&hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: "connection", LabelNames: []string{"type", "name"}}}})
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}
	byType := map[string]map[string]cty.Value{}
	for _, block := range hclContent.Blocks {
		conn, diags := parse.DecodePipelingConnection(path, block)
		if diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		value, err := conn.CtyValue()
		if err != nil {
			t.Fatal(err)
		}
		if byType[block.Labels[0]] == nil {
			byType[block.Labels[0]] = map[string]cty.Value{}
		}
		byType[block.Labels[0]][block.Labels[1]] = value
	}
	types := map[string]cty.Value{}
	for ty, conns := range byType {
		types[ty] = cty.ObjectVal(conns)
	}
	return cty.ObjectVal(types)
}

// routeRow is one way a mod can reach a database or a file through one
type routeRow struct {
	name string
	// database is the value the mod's `database` takes, from the eval context
	database func(d routesDirs) string
	// databaseExpr replaces `var.database` in the copy of mod.pp, to name a .ppc connection
	databaseExpr string
	// ppc loads connections.ppc into the eval context
	ppc bool
	// open opens the value through the backend, which is DuckDB or SQLite, and runs sql on it
	open bool
	// sqlFile is a file of the test mod that open runs on the database
	sqlFile string
	// refused: with the setting set, the value is refused at the mod parser (or at GetConnectionString); else
	// it works, as it does with the setting unset
	refused bool
	// sqlRefused: with the setting set, the SQL fails with one of these errors
	sqlRefused []string
	// unsetErr: with the setting unset, the SQL runs and fails with this error that is not the lock's (a
	// non-extension file cannot be installed)
	unsetErr string
	// outsideFile is a file the SQL creates outside the permitted folder when it runs unlocked
	outsideFile string
}

var lockErrors = []string{"Permission Error", "configuration has been locked"}

func postgresOption(option string) func(d routesDirs) string {
	return func(d routesDirs) string {
		return "postgres://alice:s3cretpw@db.example.com:5432/app?sslmode=require&" + option + "=" + url.QueryEscape(d.out(option+".pem"))
	}
}

// TestPermittedDatabaseRoutes drives every route a mod has to a database or to a file through a database, with
// <APP>_PERMITTED_ROOT_PATHS set (each is refused) and unset (each works, as it did before the setting).
//
// Every row goes through the mod parser (parse.ParseModDefinition) on a copy of the test mod. A DuckDB or SQLite
// value is then opened through the real backend and the row's SQL runs on it. There is no Postgres or MySQL server
// in the unit gate, so for those "works" means the parsed value reaches GetConnectionString unchanged.
func TestPermittedDatabaseRoutes(t *testing.T) {
	// A new database route is one row here.
	rows := []routeRow{
		// 1. a DuckDB file as the mod's database
		{name: "duckdb file outside", database: func(d routesDirs) string { return "duckdb://" + d.out("a.duckdb") }, open: true, refused: true},
		{name: "duckdb file inside", database: func(d routesDirs) string { return "duckdb://" + d.in("a.duckdb") }, open: true},
		// 2. a SQLite file as the mod's database
		{name: "sqlite file outside", database: func(d routesDirs) string { return "sqlite://" + d.out("a.sqlite") }, open: true, refused: true},
		{name: "sqlite file inside", database: func(d routesDirs) string { return "sqlite://" + d.in("a.sqlite") }, open: true},
		// 3. options on a DuckDB value
		{name: "duckdb options", database: func(d routesDirs) string { return "duckdb://" + d.in("a.duckdb") + "?access_mode=read_write" }, open: true, refused: true},
		// 4. duckdbinit:// from mod content
		{name: "duckdbinit", database: func(d routesDirs) string { return "duckdbinit://" + d.in("init.sql") }, open: true, refused: true},
		// 5. each Postgres option that makes the driver read a file
		{name: "postgres sslcert", database: postgresOption("sslcert"), refused: true},
		{name: "postgres sslkey", database: postgresOption("sslkey"), refused: true},
		{name: "postgres sslrootcert", database: postgresOption("sslrootcert"), refused: true},
		{name: "postgres passfile", database: postgresOption("passfile"), refused: true},
		{name: "postgres servicefile", database: postgresOption("servicefile"), refused: true},
		// 6. MySQL local files
		{name: "mysql allowAllFiles", database: func(routesDirs) string { return "mysql://root@tcp(localhost:3306)/app?allowAllFiles=true" }, refused: true},
		// 7. SQL a mod runs on DuckDB (the database itself is in memory, which is allowed)
		{name: "duckdb sql reads outside", database: func(routesDirs) string { return "duckdb://" }, open: true, sqlFile: "duckdb_read_outside.sql", sqlRefused: lockErrors},
		{name: "duckdb sql writes outside", database: func(routesDirs) string { return "duckdb://" }, open: true, sqlFile: "duckdb_write_outside.sql", sqlRefused: lockErrors, outsideFile: "written.csv"},
		{name: "duckdb sql attaches outside", database: func(routesDirs) string { return "duckdb://" }, open: true, sqlFile: "duckdb_attach_outside.sql", sqlRefused: lockErrors, outsideFile: "attached.duckdb"},
		{name: "duckdb sql installs from outside", database: func(routesDirs) string { return "duckdb://" }, open: true, sqlFile: "duckdb_install_outside.sql", sqlRefused: lockErrors, unsetErr: "extension"},
		// 8. SQL a mod runs on SQLite
		{name: "sqlite sql attaches outside", database: func(d routesDirs) string { return "sqlite://" + d.in("a.sqlite") }, open: true, sqlFile: "sqlite_attach_outside.sql", sqlRefused: []string{"not authorized"}},
		// 9. a .ppc file in the mod's folder, named by the mod's database as connection.<type>.<name>
		{name: "ppc duckdb connection outside", ppc: true, databaseExpr: "connection.duckdb.outside_duck", database: func(routesDirs) string { return "" }, refused: true},
		{name: "ppc sqlite connection outside", ppc: true, databaseExpr: "connection.sqlite.outside_lite", database: func(routesDirs) string { return "" }, refused: true},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Run("set", func(t *testing.T) { runRoute(t, row, true) })
			t.Run("unset", func(t *testing.T) { runRoute(t, row, false) })
		})
	}
}

// TestPermittedTailpipeRouteRefused: a tailpipe connection is refused before `tailpipe connect` runs, with the setting
// set, and the refusal names the setting. Unset, the connect is attempted; PATH holds no tailpipe, so it fails there,
// not on the setting. The tailpipe binary is never run.
func TestPermittedTailpipeRouteRefused(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		newRoutesDirs(t, true)
		conn := connection.NewTailpipeConnection("refused", hcl.Range{}).(*connection.TailpipeConnection)
		_, err := conn.GetConnectionString()
		if err == nil || !strings.Contains(err.Error(), routesPermittedEnv) {
			t.Fatalf("want a refusal naming %s, got: %v", routesPermittedEnv, err)
		}
	})
	t.Run("unset", func(t *testing.T) {
		newRoutesDirs(t, false)
		t.Setenv("PATH", t.TempDir())
		conn := connection.NewTailpipeConnection("unrefused", hcl.Range{}).(*connection.TailpipeConnection)
		_, err := conn.GetConnectionString()
		if err == nil || strings.Contains(err.Error(), routesPermittedEnv) {
			t.Fatalf("want the connect to fail for another reason, got: %v", err)
		}
	})
}

// parseRouteMod parses the copy of the test mod and returns the connection string its database gives
func parseRouteMod(t *testing.T, row routeRow, d routesDirs) (string, error) {
	t.Helper()
	modFile := d.in("mod.pp")
	if row.databaseExpr != "" {
		content, err := os.ReadFile(modFile)
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.ReplaceAll(string(content), "var.database", row.databaseExpr))
		if err := os.WriteFile(modFile, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	variables := map[string]cty.Value{"var": cty.ObjectVal(map[string]cty.Value{"database": cty.StringVal(row.database(d))})}
	if row.ppc {
		variables["connection"] = loadPpcConnections(t, d)
	}
	mod, res := parse.ParseModDefinition(modFile, &hcl.EvalContext{Variables: variables})
	if res.Diags.HasErrors() {
		return "", errors.New(res.Diags.Error())
	}
	provider := mod.GetDatabase()
	if provider == nil {
		t.Fatal("the mod has no database")
	}
	return provider.GetConnectionString()
}

func runRoute(t *testing.T, row routeRow, set bool) {
	d := newRoutesDirs(t, set)
	want := row.database(d)
	if row.ppc {
		want = ""
	}

	cs, err := parseRouteMod(t, row, d)
	if set && row.refused {
		if err == nil {
			t.Fatalf("want the value refused, got %q", cs)
		}
		if !strings.Contains(err.Error(), routesPermittedEnv) {
			t.Fatalf("want an error naming %s, got: %v", routesPermittedEnv, err)
		}
		if strings.Contains(err.Error(), "s3cretpw") {
			t.Fatalf("the error leaks the password: %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("want the value to work, got: %v", err)
	}
	switch {
	case row.ppc:
		// unset, the connection's file is returned as the .ppc names it
		if wantPrefix := map[string]string{"duckdb": "duckdb://" + d.out("ppc.duckdb"), "sqlite": "sqlite://" + d.out("ppc.sqlite")}[strings.Split(row.databaseExpr, ".")[1]]; cs != wantPrefix {
			t.Fatalf("got %q, want %q", cs, wantPrefix)
		}
	case !set:
		// unset, the value reaches the backend unchanged
		if cs != want {
			t.Fatalf("got %q, want the value unchanged, %q", cs, want)
		}
	}
	if !row.open {
		return
	}

	ctx := context.Background()
	b, err := backend.FromConnectionString(ctx, cs)
	if err != nil {
		t.Fatalf("FromConnectionString: %v", err)
	}
	db, err := b.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	var n int
	if err := db.QueryRowContext(ctx, "SELECT 42").Scan(&n); err != nil || n != 42 {
		t.Fatalf("SELECT 42 = %d, %v", n, err)
	}

	if row.sqlFile == "" {
		return
	}
	statement, err := os.ReadFile(d.in(row.sqlFile))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, string(statement))
	switch {
	case set:
		if err == nil {
			t.Fatalf("%s: want the lock to refuse it, got no error", row.sqlFile)
		}
		if !containsAny(err.Error(), row.sqlRefused) {
			t.Fatalf("%s: want an error containing one of %q, got: %v", row.sqlFile, row.sqlRefused, err)
		}
	case row.unsetErr != "":
		if err == nil || !strings.Contains(err.Error(), row.unsetErr) || containsAny(err.Error(), row.sqlRefused) {
			t.Fatalf("%s: want an error containing %q that is not the lock's, got: %v", row.sqlFile, row.unsetErr, err)
		}
	default:
		if err != nil {
			t.Fatalf("%s: want it to work, got: %v", row.sqlFile, err)
		}
	}
	if row.outsideFile != "" {
		_, statErr := os.Stat(d.out(row.outsideFile))
		if created := statErr == nil; created == set {
			t.Fatalf("%s: file created outside = %v, want %v", row.sqlFile, created, !set)
		}
	}
}

func containsAny(s string, fragments []string) bool {
	for _, f := range fragments {
		if strings.Contains(s, f) {
			return true
		}
	}
	return false
}
