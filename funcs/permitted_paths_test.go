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
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/zclconf/go-cty/cty"
)

const (
	testAppPrefix    = "TESTAPP_"
	testPermittedEnv = testAppPrefix + "PERMITTED_ROOT_PATHS"
	outsideErr       = "path is outside the folders " + testPermittedEnv + " permits"
)

// fileFunctionNames are the functions whose first argument is the path of one file they read or stat
var fileFunctionNames = []string{
	"file", "fileexists", "filebase64", "filebase64sha256", "filebase64sha512",
	"filemd5", "filesha1", "filesha256", "filesha512",
}

// fileFuncFixture is a mod folder beside a folder outside it:
//
//	<root>/outside/secret.txt         the file a mod must not read
//	<root>/outside/locked/key         a file in a folder that cannot be searched
//	<root>/mod/docs/x.md              a file inside the mod
//	<root>/mod/link_in -> docs/x.md   a symlink that stays inside
//	<root>/mod/link_out -> ../outside/secret.txt
//	<root>/mod/linkdir_out -> ../outside
//	<root>/mod/dang_out -> ../outside/missing.txt
//	<root>/mod/locked_out -> ../outside/locked/key
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
	mustSymlink(t, filepath.Join("..", "outside", "missing.txt"), filepath.Join(f.mod, "dang_out"))
	mustSymlink(t, filepath.Join("..", "outside", "locked", "key"), filepath.Join(f.mod, "locked_out"))
	mustSymlink(t, "mod", f.modLink)
	// a HOME outside the mod, so ~ paths are tested without reading the real home folder
	t.Setenv("HOME", f.outside)
	return f
}

// setPermitted sets the app's env var keys to a test-only prefix, as an app does at start-up, sets the
// setting to value (value "" unsets it) and reads it, as the app does once at start-up
func setPermitted(t *testing.T, value string) error {
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

// permit limits the file functions to the folders
func permit(t *testing.T, folders ...string) {
	t.Helper()
	if err := setPermitted(t, strings.Join(folders, ",")); err != nil {
		t.Fatal(err)
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
	if !strings.Contains(diags.Error(), outsideErr) {
		t.Fatalf("%s: expected %q in the error, got: %s", expr, outsideErr, diags.Error())
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

func TestPermittedPaths_On_ReadsInsideTheModFolder(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

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
	for _, name := range fileFunctionNames {
		mustEval(t, f.mod, name+`("docs/x.md")`)
	}

	// a missing file inside the folder is reported missing, not refused
	if got := mustEval(t, f.mod, `fileexists("docs/missing.md")`); got.True() {
		t.Errorf(`fileexists("docs/missing.md") = true, want false`)
	}
	if got := mustEval(t, f.mod, `fileexists("docs/x.md")`); !got.True() {
		t.Errorf(`fileexists("docs/x.md") = false, want true`)
	}
	// so is one reached through a symlink that stays inside
	mustSymlink(t, filepath.Join("docs", "missing.md"), filepath.Join(f.mod, "dang_in"))
	mustSymlink(t, "loop", filepath.Join(f.mod, "loop"))
	if got := mustEval(t, f.mod, `fileexists("dang_in")`); got.True() {
		t.Errorf(`fileexists("dang_in") = true, want false`)
	}
	// a symlink loop fails, and is not read as outside
	_, loopDiags := eval(t, f.mod, `file("loop")`)
	if !loopDiags.HasErrors() || strings.Contains(loopDiags.Error(), outsideErr) {
		t.Errorf(`file("loop"): want a symlink loop error, got: %v`, loopDiags)
	}
	_, diags := eval(t, f.mod, `file("docs/missing.md")`)
	if !diags.HasErrors() || strings.Contains(diags.Error(), outsideErr) {
		t.Errorf(`file("docs/missing.md"): want a missing-file error, got: %v`, diags)
	}
}

func TestPermittedPaths_On_ModFolderReachedThroughASymlink(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

	if got := mustEval(t, f.modLink, `file("docs/x.md")`); got.AsString() != "hello" {
		t.Errorf(`file("docs/x.md") = %q, want "hello"`, got.AsString())
	}
	mustRefuse(t, f.modLink, `file("../outside/secret.txt")`)
	mustRefuse(t, f.modLink, `file("link_out")`)
}

func TestPermittedPaths_On_RefusesPathsOutsideTheModFolder(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

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
		"dang_out",                     // a symlink to a missing file outside
		"locked_out",                   // a symlink into a folder outside that cannot be searched
	}
	for _, name := range fileFunctionNames {
		for _, p := range outsidePaths {
			mustRefuse(t, f.mod, name+`("`+p+`")`)
		}
	}
}

// a symlink whose target climbs back with ".." past a part that cannot be looked up must not be cleaned into
// a path that opens another symlink, unchecked
func TestPermittedPaths_On_SymlinkClimbingPastAMissingPart(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)
	lockedIn := filepath.Join(f.mod, "lockedin")
	mustMkdir(t, filepath.Join(lockedIn, "x"))
	if err := os.Chmod(lockedIn, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(lockedIn, 0o755) })
	mustSymlink(t, "missing/../link_out", filepath.Join(f.mod, "via_missing"))
	mustSymlink(t, "lockedin/x/../../link_out", filepath.Join(f.mod, "via_locked"))
	mustSymlink(t, "../outside/missing/../secret.txt", filepath.Join(f.mod, "via_outside_missing"))

	for _, name := range fileFunctionNames {
		for _, link := range []string{"via_missing", "via_locked", "via_outside_missing"} {
			v, diags := eval(t, f.mod, name+`("`+link+`")`)
			if !diags.HasErrors() {
				t.Errorf("%s(%q) = %#v, want an error", name, link, v)
			}
		}
		// a failed lookup outside the folder is reported as outside
		mustRefuse(t, f.mod, name+`("via_outside_missing")`)
	}
	got := setStrings(t, mustEval(t, f.mod, `fileset(".", "via_*")`))
	if len(got) != 0 {
		t.Errorf(`fileset(".", "via_*") = %v, want none`, got)
	}
}

func TestPermittedPaths_On_ErrorNamesTheCall(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

	_, diags := eval(t, f.mod, `file("/etc/passwd")`)
	want := `file("/etc/passwd"): ` + outsideErr
	if !strings.Contains(diags.Error(), want) {
		t.Fatalf("want %q in the error, got: %s", want, diags.Error())
	}
}

func TestPermittedPaths_On_SensitivePathIsRedacted(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

	_, diags := eval(t, f.mod, `file(sensitive("/etc/passwd"))`)
	if !diags.HasErrors() || strings.Contains(diags.Error(), "/etc/passwd") || !strings.Contains(diags.Error(), "(sensitive value)") {
		t.Fatalf("want the refusal with the path redacted, got: %v", diags)
	}
}

func TestPermittedPaths_On_FilesetNeverListsOutside(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)

	got := setStrings(t, mustEval(t, f.mod, `fileset(".", "**")`))
	if !contains(got, "docs/x.md") {
		t.Errorf(`fileset(".", "**") = %v, want it to list docs/x.md`, got)
	}
	for _, m := range got {
		if strings.HasPrefix(m, "link_out") || strings.HasPrefix(m, "linkdir_out") || strings.Contains(m, "..") {
			t.Errorf(`fileset(".", "**") lists %q, which is outside the mod folder`, m)
		}
	}
	// symlinks outside, to a file that exists, is missing or cannot be reached, are all left out alike
	got = setStrings(t, mustEval(t, f.mod, `fileset(".", "*")`))
	if want := []string{"link_in"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf(`fileset(".", "*") = %v, want %v`, got, want)
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

func TestPermittedPaths_Off_BehaviourUnchanged(t *testing.T) {
	for _, value := range []string{"", "  "} {
		t.Run("value="+value, func(t *testing.T) {
			f := newFileFuncFixture(t)
			if err := setPermitted(t, value); err != nil {
				t.Fatal(err)
			}
			// the setting #815 shipped on a branch has no effect
			t.Setenv(testAppPrefix+"RESTRICT_FILE_FUNCTIONS", "true")

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
			got = setStrings(t, mustEval(t, f.mod, `fileset(".", "link_*")`))
			if !contains(got, "link_out") {
				t.Errorf(`fileset(".", "link_*") = %v, want link_out`, got)
			}
			// and a symlink to a missing file outside makes fileset fail, which tells a mod it is missing
			if _, diags := eval(t, f.mod, `fileset(".", "dang_*")`); !diags.HasErrors() {
				t.Errorf(`fileset(".", "dang_*"): want a stat error with the setting off`)
			}
		})
	}
}

func TestPermittedPaths_RelativeEntryIsRefusedAtStartUp(t *testing.T) {
	f := newFileFuncFixture(t)
	for _, value := range []string{"mod", f.mod + ",docs", "./mod", "~/mod", f.mod + " , ../outside"} {
		err := setPermitted(t, value)
		if err == nil {
			t.Fatalf("value %q: want an error for the relative entry", value)
		}
		for _, want := range []string{testPermittedEnv, "not an absolute path"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("value %q: want %q in the error, got: %v", value, want, err)
			}
		}
	}
	err := setPermitted(t, "docs")
	if err == nil || !strings.Contains(err.Error(), `"docs"`) {
		t.Errorf(`want the error to name the entry "docs", got: %v`, err)
	}
	if err := setPermitted(t, " , "); err == nil || !strings.Contains(err.Error(), "names no folder") {
		t.Errorf("a list of nothing: want a names-no-folder error, got: %v", err)
	}
}

// a setting that could not be read is never read as no limit
func TestPermittedPaths_UnreadableSettingFailsClosed(t *testing.T) {
	f := newFileFuncFixture(t)
	if err := setPermitted(t, "docs"); err == nil {
		t.Fatal("want an error")
	}
	_, diags := eval(t, f.mod, `file("docs/x.md")`)
	if !diags.HasErrors() || !strings.Contains(diags.Error(), testPermittedEnv) {
		t.Fatalf("want every read refused and the setting named, got: %v", diags)
	}
	// and the same when the app never called Load
	permitted_paths.Reset()
	_, diags = eval(t, f.mod, `file("docs/x.md")`)
	if !diags.HasErrors() {
		t.Fatalf("want every read refused, got: %v", diags)
	}
}

func TestPermittedPaths_ListOfFolders(t *testing.T) {
	f := newFileFuncFixture(t)
	// spaces round the entries and empty entries are fine
	if err := setPermitted(t, " "+f.mod+" , ,"+f.outside+","); err != nil {
		t.Fatal(err)
	}
	if got := mustEval(t, f.mod, `file("docs/x.md")`); got.AsString() != "hello" {
		t.Errorf("got %q", got.AsString())
	}
	// the second folder is read too, from the mod and by its absolute path
	if got := mustEval(t, f.mod, `file("../outside/secret.txt")`); got.AsString() != "tls-key" {
		t.Errorf("got %q", got.AsString())
	}
	if got := mustEval(t, f.mod, `file("`+f.secret+`")`); got.AsString() != "tls-key" {
		t.Errorf("got %q", got.AsString())
	}
	// a folder in neither is refused
	mustRefuse(t, f.mod, `file("/etc/passwd")`)
	mustRefuse(t, f.mod, `file("../")`)
}

// the folder is the permitted one, not the mod: a parent folder lets the mod read its neighbours, and nothing above it
func TestPermittedPaths_ParentFolderPermitsNeighbours(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.root)
	for _, expr := range []string{`file("../outside/secret.txt")`, `file("link_out")`} {
		if got := mustEval(t, f.mod, expr); got.AsString() != "tls-key" {
			t.Errorf("%s = %q, want tls-key", expr, got.AsString())
		}
	}
	mustEval(t, f.mod, `file("docs/x.md")`)
	mustRefuse(t, f.mod, `file("/etc/passwd")`)
	mustRefuse(t, f.mod, `file("../../x")`)
	mustRefuse(t, f.mod, `file("`+filepath.Dir(f.root)+`/x")`)
	// a listing across the neighbours
	got := setStrings(t, mustEval(t, f.mod, `fileset("..", "outside/*")`))
	if !contains(got, "outside/secret.txt") {
		t.Errorf(`fileset("..", "outside/*") = %v, want outside/secret.txt`, got)
	}
	mustRefuse(t, f.mod, `fileset("../..", "*")`)
}

// a permitted folder that is itself reached through a symlink, whichever way it is written
func TestPermittedPaths_FolderThroughASymlink(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.modLink)
	for _, base := range []string{f.mod, f.modLink} {
		if got := mustEval(t, base, `file("docs/x.md")`); got.AsString() != "hello" {
			t.Errorf("base %s: got %q", base, got.AsString())
		}
		mustRefuse(t, base, `file("link_out")`)
		mustRefuse(t, base, `file("../outside/secret.txt")`)
	}
}

// a permitted folder holding a symlink out of it still refuses what the link leads to
func TestPermittedPaths_SymlinkOutOfTheFolderIsRefusedWhateverTheBase(t *testing.T) {
	f := newFileFuncFixture(t)
	permit(t, f.mod)
	// the function's base folder is outside the permitted folder
	mustRefuse(t, f.outside, `file("secret.txt")`)
	if got := mustEval(t, f.outside, `file("`+filepath.Join(f.mod, "docs", "x.md")+`")`); got.AsString() != "hello" {
		t.Errorf("got %q", got.AsString())
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
