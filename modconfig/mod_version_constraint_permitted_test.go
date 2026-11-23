package modconfig

import (
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

func TestModVersionConstraintDatabasePermitted(t *testing.T) {
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
	// a require block pointing at a local mod, so Initialise has nothing to do but the database check
	newConstraint := func(database string) *ModVersionConstraint {
		return &ModVersionConstraint{
			Name:     "github.com/example/mod",
			FilePath: "../mod",
			Database: &database,
			DefRange: hcl.Range{Filename: "mod.sp", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 3, Column: 8}},
		}
	}

	t.Run("outside is refused on the block", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		m := newConstraint("duckdb://" + filepath.Join(outside, "a.db"))
		diags := m.Initialise(nil)
		if !diags.HasErrors() {
			t.Fatal("want an error diagnostic")
		}
		if diags[0].Subject == nil || *diags[0].Subject != m.DefRange {
			t.Fatalf("want the diagnostic on the block %v, got %v", m.DefRange, diags[0].Subject)
		}
		if !strings.Contains(diags[0].Summary, "outside the folders "+testPermittedEnv) {
			t.Fatalf("unexpected summary %q", diags[0].Summary)
		}
	})

	t.Run("an option is refused", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		m := newConstraint("duckdb://" + filepath.Join(permitted, "a.db") + "?access_mode=read_only")
		if diags := m.Initialise(nil); !diags.HasErrors() {
			t.Fatal("want an error diagnostic")
		}
	})

	t.Run("inside is rewritten to the resolved path", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		inside := filepath.Join(permitted, "sub", "..", "a.db")
		m := newConstraint("duckdb://" + inside)
		if diags := m.Initialise(nil); diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		if want := "duckdb://" + filepath.Join(permitted, "a.db"); *m.Database != want {
			t.Fatalf("got %q, want %q", *m.Database, want)
		}
	})

	t.Run("unset leaves the value unchanged", func(t *testing.T) {
		setPermittedForTest(t, "")
		value := "duckdb://" + filepath.Join(outside, "a.db") + "?access_mode=read_only"
		m := newConstraint(value)
		if diags := m.Initialise(nil); diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		if *m.Database != value {
			t.Fatalf("got %q, want %q", *m.Database, value)
		}
	})

	t.Run("no database override is untouched", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		m := newConstraint("")
		m.Database = nil
		if diags := m.Initialise(nil); diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		if m.Database != nil {
			t.Fatalf("want no database, got %q", *m.Database)
		}
	})
}
