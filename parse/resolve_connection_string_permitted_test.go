package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/schema"
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

// databaseContent parses a body holding one database attribute, takes its content as decodeMod does
func databaseContent(t *testing.T, database string) *hcl.BodyContent {
	t.Helper()
	body, diags := hclsyntax.ParseConfig([]byte("database = \""+database+"\"\n"), "mod.sp", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}
	content, _, diags := body.Body.PartialContent(&hcl.BodySchema{Attributes: []hcl.AttributeSchema{
		{Name: schema.AttributeTypeDatabase},
	}})
	if diags.HasErrors() {
		t.Fatal(diags.Error())
	}
	return content
}

func TestResolveConnectionStringPermitted(t *testing.T) {
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

	t.Run("outside is an error on the attribute", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		content := databaseContent(t, "duckdb://"+filepath.Join(outside, "a.db"))
		csp, _, _, diags := ResolveConnectionString(content, &hcl.EvalContext{})
		if !diags.HasErrors() {
			t.Fatal("want an error diagnostic")
		}
		if csp != nil {
			t.Fatalf("want no connection string provider, got %v", csp)
		}
		want := content.Attributes[schema.AttributeTypeDatabase].Range
		if diags[0].Subject == nil || *diags[0].Subject != want {
			t.Fatalf("want the diagnostic on the database attribute %v, got %v", want, diags[0].Subject)
		}
		if !strings.Contains(diags[0].Summary, "outside the folders "+testPermittedEnv) {
			t.Fatalf("unexpected summary %q", diags[0].Summary)
		}
	})

	t.Run("inside resolves to the resolved path", func(t *testing.T) {
		setPermittedForTest(t, permitted)
		inside := filepath.Join(permitted, "a.db")
		csp, _, _, diags := ResolveConnectionString(databaseContent(t, "duckdb://"+inside), &hcl.EvalContext{})
		if diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		got, err := csp.GetConnectionString()
		if err != nil {
			t.Fatal(err)
		}
		if want := "duckdb://" + inside; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("unset leaves the value unchanged", func(t *testing.T) {
		setPermittedForTest(t, "")
		value := "duckdb://" + filepath.Join(outside, "a.db") + "?access_mode=read_only"
		csp, _, _, diags := ResolveConnectionString(databaseContent(t, value), &hcl.EvalContext{})
		if diags.HasErrors() {
			t.Fatal(diags.Error())
		}
		got, err := csp.GetConnectionString()
		if err != nil {
			t.Fatal(err)
		}
		if got != value {
			t.Fatalf("got %q, want %q", got, value)
		}
	})
}
