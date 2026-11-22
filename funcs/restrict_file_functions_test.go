package funcs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/terraform-components/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

const testRestrictEnv = "TEST_APP_RESTRICT_FILE_FUNCTIONS"

// fileFuncFixture is a mod folder beside a folder outside it:
//
//	<root>/outside/secret.txt         the file a mod must not read
//	<root>/outside/locked/key         a file in a folder that cannot be searched
//	<root>/mod/docs/x.md              a file inside the mod
//	<root>/mod/link_in -> docs/x.md   a symlink that stays inside
//	<root>/mod/link_out -> ../outside/secret.txt
//	<root>/mod/linkdir_out -> ../outside
//	<root>/modlink -> mod             a mod folder reached through a symlink
type fileFuncFixture struct {
	root, mod, modLink, outside, secret, locked string
}

func newFileFuncFixture(t *testing.T) fileFuncFixture {
	t.Helper()
	root := t.TempDir()
	f := fileFuncFixture{
		root:    root,
		mod:     filepath.Join(root, "mod"),
		modLink: filepath.Join(root, "modlink"),
		outside: filepath.Join(root, "outside"),
		secret:  filepath.Join(root, "outside", "secret.txt"),
		locked:  filepath.Join(root, "outside", "locked"),
	}
	mustMkdir(t, filepath.Join(f.mod, "docs"))
	mustMkdir(t, f.outside)
	mustWrite(t, filepath.Join(f.mod, "docs", "x.md"), "hello")
	mustWrite(t, f.secret, "tls-key")
	mustMkdir(t, f.locked)
	mustWrite(t, filepath.Join(f.locked, "key"), "locked-key")
	if err := os.Chmod(f.locked, 0); err != nil {
		t.Fatal(err)
	}
	// let t.TempDir remove it
	t.Cleanup(func() { _ = os.Chmod(f.locked, 0o755) })
	mustSymlink(t, filepath.Join("docs", "x.md"), filepath.Join(f.mod, "link_in"))
	mustSymlink(t, filepath.Join("..", "outside", "secret.txt"), filepath.Join(f.mod, "link_out"))
	mustSymlink(t, filepath.Join("..", "outside"), filepath.Join(f.mod, "linkdir_out"))
	mustSymlink(t, "mod", f.modLink)
	// a HOME outside the mod, so ~ paths are tested without reading the real home folder
	t.Setenv("HOME", f.outside)
	return f
}

// setRestrict points the app's env var key at a test-only name and sets its value; value "" unsets it
func setRestrict(t *testing.T, value string) {
	t.Helper()
	prev := app_specific.EnvRestrictFileFunctions
	app_specific.EnvRestrictFileFunctions = testRestrictEnv
	t.Cleanup(func() { app_specific.EnvRestrictFileFunctions = prev })
	t.Setenv(testRestrictEnv, value)
	if value == "" {
		if err := os.Unsetenv(testRestrictEnv); err != nil {
			t.Fatal(err)
		}
	}
}

// eval evaluates an HCL expression with the context functions for baseDir, as the mod parser does
func eval(t *testing.T, baseDir, expr string) (cty.Value, hcl.Diagnostics) {
	t.Helper()
	parsed, diags := hclsyntax.ParseExpression([]byte(expr), "test.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse %s: %s", expr, diags.Error())
	}
	return parsed.Value(&hcl.EvalContext{Functions: ContextFunctions(baseDir)})
}

func mustEval(t *testing.T, baseDir, expr string) cty.Value {
	t.Helper()
	v, diags := eval(t, baseDir, expr)
	if diags.HasErrors() {
		t.Fatalf("%s: unexpected error: %s", expr, diags.Error())
	}
	return v
}

func mustRefuse(t *testing.T, baseDir, expr string) {
	t.Helper()
	v, diags := eval(t, baseDir, expr)
	if !diags.HasErrors() {
		t.Fatalf("%s: expected it to be refused, got %#v", expr, v)
	}
	if !strings.Contains(diags.Error(), errOutsideFolder) {
		t.Fatalf("%s: expected %q in the error, got: %s", expr, errOutsideFolder, diags.Error())
	}
}

func setStrings(t *testing.T, v cty.Value) []string {
	t.Helper()
	var res []string
	for it := v.ElementIterator(); it.Next(); {
		_, e := it.Element()
		res = append(res, e.AsString())
	}
	sort.Strings(res)
	return res
}

func TestRestrictFileFunctions_On_ReadsInsideTheModFolder(t *testing.T) {
	f := newFileFuncFixture(t)
	setRestrict(t, "true")

	for _, expr := range []string{
		`file("./docs/x.md")`,
		`file("docs/x.md")`,
		`file("docs/../docs/x.md")`,
		`file("link_in")`,
		`file("` + filepath.Join(f.mod, "docs", "x.md") + `")`,
	} {
		if got := mustEval(t, f.mod, expr); got.AsString() != "hello" {
			t.Errorf("%s = %q, want %q", expr, got.AsString(), "hello")
		}
	}

	// the other single-path functions read inside the folder too
	for _, name := range singlePathFileFunctions {
		mustEval(t, f.mod, name+`("docs/x.md")`)
	}

	// a missing file inside the folder is reported missing, not refused
	if got := mustEval(t, f.mod, `fileexists("docs/missing.md")`); got.True() {
		t.Errorf(`fileexists("docs/missing.md") = true, want false`)
	}
	if got := mustEval(t, f.mod, `fileexists("docs/x.md")`); !got.True() {
		t.Errorf(`fileexists("docs/x.md") = false, want true`)
	}
	_, diags := eval(t, f.mod, `file("docs/missing.md")`)
	if !diags.HasErrors() || strings.Contains(diags.Error(), errOutsideFolder) {
		t.Errorf(`file("docs/missing.md"): want a missing-file error, got: %v`, diags)
	}
}

func TestRestrictFileFunctions_On_ModFolderReachedThroughASymlink(t *testing.T) {
	f := newFileFuncFixture(t)
	setRestrict(t, "true")

	if got := mustEval(t, f.modLink, `file("docs/x.md")`); got.AsString() != "hello" {
		t.Errorf(`file("docs/x.md") = %q, want "hello"`, got.AsString())
	}
	mustRefuse(t, f.modLink, `file("../outside/secret.txt")`)
	mustRefuse(t, f.modLink, `file("link_out")`)
}

func TestRestrictFileFunctions_On_RefusesPathsOutsideTheModFolder(t *testing.T) {
	f := newFileFuncFixture(t)
	setRestrict(t, "true")

	outsidePaths := []string{
		f.secret,                // absolute
		"/etc/passwd",           // absolute, outside every test folder
		"../outside/secret.txt", // ..
		"docs/../../outside/secret.txt",
		"link_out",                // symlink to a file outside
		"linkdir_out/secret.txt",  // through a symlinked folder outside
		"linkdir_out/missing.txt", // missing, through a symlinked folder outside
		"~/secret.txt",            // home folder
		"..",                      // the parent folder itself
		// outside paths whose lookup fails other than "not found" get the same refusal, so the error
		// never tells a mod what exists outside
		"/etc/passwd/x",                // a file used as a folder
		"linkdir_out/secret.txt/x",     // the same, through a symlinked folder
		filepath.Join(f.locked, "key"), // inside a folder that cannot be searched
		"linkdir_out/locked/key",       // the same, through a symlinked folder
	}
	for _, name := range singlePathFileFunctions {
		for _, p := range outsidePaths {
			mustRefuse(t, f.mod, name+`("`+p+`")`)
		}
	}
}

func TestRestrictFileFunctions_On_ErrorNamesTheCall(t *testing.T) {
	f := newFileFuncFixture(t)
	setRestrict(t, "true")

	_, diags := eval(t, f.mod, `file("/etc/passwd")`)
	want := `file("/etc/passwd"): path is outside the mod folder`
	if !strings.Contains(diags.Error(), want) {
		t.Fatalf("want %q in the error, got: %s", want, diags.Error())
	}
}

func TestRestrictFileFunctions_On_SensitivePathIsRedacted(t *testing.T) {
	err := confineError("file", "/secret/path", cty.NewValueMarks(marks.Sensitive), os.ErrPermission)
	if strings.Contains(err.Error(), "/secret/path") || !strings.Contains(err.Error(), "(sensitive value)") {
		t.Fatalf("want the path redacted, got: %s", err)
	}
}

func TestRestrictFileFunctions_On_FilesetNeverListsOutside(t *testing.T) {
	f := newFileFuncFixture(t)
	setRestrict(t, "true")

	got := setStrings(t, mustEval(t, f.mod, `fileset(".", "**")`))
	if !contains(got, "docs/x.md") {
		t.Errorf(`fileset(".", "**") = %v, want it to list docs/x.md`, got)
	}
	for _, m := range got {
		if strings.HasPrefix(m, "link_out") || strings.HasPrefix(m, "linkdir_out") || strings.Contains(m, "..") {
			t.Errorf(`fileset(".", "**") lists %q, which is outside the mod folder`, m)
		}
	}
	got = setStrings(t, mustEval(t, f.mod, `fileset(".", "*")`))
	if contains(got, "link_out") {
		t.Errorf(`fileset(".", "*") = %v, lists link_out`, got)
	}

	// a brace alternative hides ".." from the pattern check; nothing outside is listed
	got = setStrings(t, mustEval(t, f.mod, `fileset(".", "{docs,../outside}/*")`))
	if len(got) != 1 || got[0] != "docs/x.md" {
		t.Errorf(`fileset(".", "{docs,../outside}/*") = %v, want [docs/x.md]`, got)
	}

	mustRefuse(t, f.mod, `fileset("..", "**")`)
	mustRefuse(t, f.mod, `fileset("`+f.outside+`", "*")`)
	mustRefuse(t, f.mod, `fileset("linkdir_out", "*")`)
	mustRefuse(t, f.mod, `fileset(".", "../outside/*")`)
	mustRefuse(t, f.mod, `fileset("docs", "../../**")`)
}

func TestRestrictFileFunctions_Off_BehaviourUnchanged(t *testing.T) {
	for _, value := range []string{"", "false", "0", "yes"} {
		t.Run("value="+value, func(t *testing.T) {
			f := newFileFuncFixture(t)
			setRestrict(t, value)

			for _, expr := range []string{
				`file("` + f.secret + `")`,
				`file("../outside/secret.txt")`,
				`file("link_out")`,
				`file("linkdir_out/secret.txt")`,
			} {
				if got := mustEval(t, f.mod, expr); got.AsString() != "tls-key" {
					t.Errorf("%s = %q, want %q", expr, got.AsString(), "tls-key")
				}
			}
			if got := mustEval(t, f.mod, `file("docs/x.md")`); got.AsString() != "hello" {
				t.Errorf(`file("docs/x.md") = %q, want "hello"`, got.AsString())
			}
			if got := mustEval(t, f.mod, `fileexists("../outside/secret.txt")`); !got.True() {
				t.Errorf(`fileexists("../outside/secret.txt") = false, want true`)
			}
			got := setStrings(t, mustEval(t, f.mod, `fileset("..", "outside/*")`))
			if !contains(got, "outside/secret.txt") {
				t.Errorf(`fileset("..", "outside/*") = %v, want outside/secret.txt`, got)
			}
			// the symlink to a file outside is listed, which the setting stops
			got = setStrings(t, mustEval(t, f.mod, `fileset(".", "*")`))
			if !contains(got, "link_out") {
				t.Errorf(`fileset(".", "*") = %v, want link_out`, got)
			}
		})
	}
}

func TestRestrictFileFunctionsEnabled(t *testing.T) {
	for value, want := range map[string]bool{
		"true": true, "TRUE": true, "1": true,
		"": false, "false": false, "0": false, "yes": false,
	} {
		setRestrict(t, value)
		if got := RestrictFileFunctionsEnabled(); got != want {
			t.Errorf("value %q: RestrictFileFunctionsEnabled() = %v, want %v", value, got, want)
		}
	}

	// an app that has not set its env var keys has no setting to read
	prev := app_specific.EnvRestrictFileFunctions
	app_specific.EnvRestrictFileFunctions = ""
	t.Cleanup(func() { app_specific.EnvRestrictFileFunctions = prev })
	if RestrictFileFunctionsEnabled() {
		t.Error("RestrictFileFunctionsEnabled() = true with no env var key set")
	}
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}
