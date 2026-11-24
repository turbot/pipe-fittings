package permitted_paths

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
)

const (
	testPrefix = "TESTAPP_"
	testEnv    = testPrefix + "PERMITTED_ROOT_PATHS"
)

// set points the app's env var keys at a test prefix, sets the setting and reads it, as an app does at start-up
func set(t *testing.T, value string) error {
	t.Helper()
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys(testPrefix)
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prev)
		Reset()
	})
	t.Setenv(testEnv, value)
	return Load()
}

// tree makes <root>/in/a.txt, <root>/in/link -> ../out/b.txt, <root>/in/dir_link -> ../out, <root>/out/b.txt
func tree(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	for _, d := range []string{"in", "out"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"in/a.txt", "out/b.txt"} {
		if err := os.WriteFile(filepath.Join(root, f), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for link, target := range map[string]string{"in/link": "../out/b.txt", "in/dir_link": "../out", "in/loop": "loop", "in/inner": "a.txt"} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestLoad_UnsetOrEmptyIsNoLimit(t *testing.T) {
	for _, value := range []string{"", "  "} {
		if err := set(t, value); err != nil {
			t.Fatalf("value %q: %v", value, err)
		}
		if Enabled() {
			t.Errorf("value %q: Enabled = true, want false", value)
		}
		got, err := Check("/etc/hosts")
		if err != nil || got != "/etc/hosts" {
			t.Errorf("value %q: Check = %q, %v; want the path back", value, got, err)
		}
		if in := []string{"/a", "/b"}; len(Filter(in)) != 2 {
			t.Errorf("value %q: Filter dropped entries", value)
		}
		if !Within("/etc/hosts") {
			t.Errorf("value %q: Within = false, want true", value)
		}
	}
}

func TestLoad_NoAppKeysIsNoLimit(t *testing.T) {
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys("")
	t.Cleanup(func() { app_specific.SetAppSpecificEnvVarKeys(prev); Reset() })
	if err := Load(); err != nil || Enabled() {
		t.Errorf("Load = %v, Enabled = %v; want no limit", err, Enabled())
	}
}

func TestLoad_ParsesTheList(t *testing.T) {
	if err := set(t, " /srv/a , ,/srv/b/../c/ ,"); err != nil {
		t.Fatal(err)
	}
	want := []string{"/srv/a", "/srv/c"}
	if got := Roots(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Roots = %v, want %v", got, want)
	}
	if Setting() != testEnv {
		t.Errorf("Setting = %q, want %q", Setting(), testEnv)
	}
}

func TestLoad_RelativeEntryNamesItself(t *testing.T) {
	err := set(t, "/srv/a,docs/x")
	if err == nil || !strings.Contains(err.Error(), `"docs/x"`) || !strings.Contains(err.Error(), testEnv) {
		t.Fatalf("want an error naming the entry and the setting, got: %v", err)
	}
	if !Enabled() {
		t.Error("a setting that could not be read must not read as no limit")
	}
	if _, err := Check("/srv/a/x"); err == nil {
		t.Error("Check must fail closed while the setting is unreadable")
	}
	if Within("/srv/a/x") {
		t.Error("Within must be false while the setting is unreadable")
	}
	if got := Filter([]string{"/srv/a/x"}); len(got) != 0 {
		t.Errorf("Filter = %v, want none", got)
	}
}

func TestCheck_InsideAndOutside(t *testing.T) {
	root := tree(t)
	if err := set(t, filepath.Join(root, "in")); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{"in/a.txt", "in/inner", "in/missing.txt", "in/../in/a.txt", "in"} {
		if _, err := Check(filepath.Join(root, p)); err != nil {
			t.Errorf("Check(%s) = %v, want it permitted", p, err)
		}
	}
	got, err := Check(filepath.Join(root, "in", "inner"))
	if err != nil || got != filepath.Join(root, "in", "a.txt") {
		t.Errorf("Check(inner) = %q, %v; want the resolved path", got, err)
	}

	for _, p := range []string{"out/b.txt", "in/link", "in/dir_link/b.txt", "in/../out/b.txt", "in/dir_link/missing", ".."} {
		_, err := Check(filepath.Join(root, p))
		var outside *OutsideError
		if !errors.As(err, &outside) {
			t.Errorf("Check(%s) = %v, want an OutsideError", p, err)
			continue
		}
		if !strings.Contains(err.Error(), filepath.Join(root, p)) || !strings.Contains(err.Error(), testEnv) {
			t.Errorf("Check(%s) error %q must name the path and the setting", p, err)
		}
	}

	// a loop is an error, and not read as outside
	_, err = Check(filepath.Join(root, "in", "loop"))
	var outside *OutsideError
	if err == nil || errors.As(err, &outside) {
		t.Errorf("Check(loop) = %v, want a loop error", err)
	}
}

func TestCheck_ManyFoldersAndFolderThroughASymlink(t *testing.T) {
	root := tree(t)
	linkToIn := filepath.Join(root, "link_to_in")
	if err := os.Symlink("in", linkToIn); err != nil {
		t.Fatal(err)
	}
	if err := set(t, linkToIn+","+filepath.Join(root, "out")); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"in/a.txt", "link_to_in/a.txt", "out/b.txt", "in/link", "in/dir_link/b.txt"} {
		if _, err := Check(filepath.Join(root, p)); err != nil {
			t.Errorf("Check(%s) = %v, want it permitted", p, err)
		}
	}
	if _, err := Check(filepath.Join(filepath.Dir(root), "x")); err == nil {
		t.Error("a path above both folders must be refused")
	}
}

func TestFilter_LeavesOutWhatIsOutside(t *testing.T) {
	root := tree(t)
	if err := set(t, filepath.Join(root, "in")); err != nil {
		t.Fatal(err)
	}
	in := []string{
		filepath.Join(root, "in", "a.txt"),
		filepath.Join(root, "in", "link"),
		filepath.Join(root, "out", "b.txt"),
		filepath.Join(root, "in", "dir_link", "b.txt"),
		filepath.Join(root, "in", "inner"),
		filepath.Join(root, "in", "loop"),
	}
	want := []string{in[0], in[4]}
	if got := Filter(in); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Filter = %v, want %v", got, want)
	}
}

func TestWithin_ReadsThePathAsWritten(t *testing.T) {
	root := tree(t)
	if err := set(t, filepath.Join(root, "in")); err != nil {
		t.Fatal(err)
	}
	if !Within(filepath.Join(root, "in", "*", "x")) || Within(filepath.Join(root, "in", "..", "out", "*")) {
		t.Error("Within must judge the pattern by where it points")
	}
}

func TestOutsideError_ReasonHasNoPath(t *testing.T) {
	e := &OutsideError{Path: "/etc/hosts", Setting: testEnv}
	if strings.Contains(e.Reason(), "/etc/hosts") || !strings.Contains(e.Reason(), testEnv) {
		t.Errorf("Reason = %q", e.Reason())
	}
	if !strings.Contains(e.Error(), `"/etc/hosts"`) || !strings.Contains(e.Error(), testEnv) {
		t.Errorf("Error = %q", e.Error())
	}
}

// modTree makes a mod folder <root>/in/mod with a file, a subfolder and a link to the file, plus <root>/out/b.txt
func modTree(t *testing.T) (root, mod string) {
	t.Helper()
	root = tree(t)
	mod = filepath.Join(root, "in", "mod")
	if err := os.MkdirAll(filepath.Join(mod, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mod, "mod.pp"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("mod.pp", filepath.Join(mod, "inner")); err != nil {
		t.Fatal(err)
	}
	return root, mod
}

func TestCheckSymlinks_LinksInsidePass(t *testing.T) {
	root, mod := modTree(t)
	if err := set(t, filepath.Join(root, "in")); err != nil {
		t.Fatal(err)
	}
	// to a file in the mod, to a missing file in the mod, to the folders' own files outside the mod, to a folder
	for link, target := range map[string]string{"sub/up": "../mod.pp", "missing": "nothing.pp", "sibling": "../a.txt", "dir": "sub", "abs": filepath.Join(root, "in", "a.txt")} {
		if err := os.Symlink(target, filepath.Join(mod, link)); err != nil {
			t.Fatal(err)
		}
	}
	if err := CheckSymlinks(mod); err != nil {
		t.Errorf("CheckSymlinks = %v, want nil", err)
	}
	// a mod folder given through a link inside the folders is walked where it leads
	viaLink := filepath.Join(root, "in", "mod_link")
	if err := os.Symlink("mod", viaLink); err != nil {
		t.Fatal(err)
	}
	if err := CheckSymlinks(viaLink); err != nil {
		t.Errorf("CheckSymlinks(link to the mod) = %v, want nil", err)
	}
	// a missing folder has nothing to check
	if err := CheckSymlinks(filepath.Join(root, "in", "no_such_mod")); err != nil {
		t.Errorf("CheckSymlinks(missing) = %v, want nil", err)
	}
}

func TestCheckSymlinks_LinkOutsideIsRefusedByName(t *testing.T) {
	cases := map[string]string{
		"relative file":    "../../out/b.txt",
		"absolute file":    "@ROOT@/out/b.txt",
		"folder":           "../../out",
		"missing outside":  "../../out/missing.txt",
		"through a link":   "../link",
		"nested/deep/link": "../../../../out/b.txt",
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			root, mod := modTree(t)
			if err := set(t, filepath.Join(root, "in")); err != nil {
				t.Fatal(err)
			}
			link := "evil"
			if strings.Contains(name, "/") {
				link = name
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(mod, link)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(strings.ReplaceAll(target, "@ROOT@", root), filepath.Join(mod, link)); err != nil {
				t.Fatal(err)
			}
			err := CheckSymlinks(mod)
			var symlinkErr *SymlinkError
			var outside *OutsideError
			if !errors.As(err, &symlinkErr) || !errors.As(err, &outside) {
				t.Fatalf("CheckSymlinks = %v, want a SymlinkError for a link outside", err)
			}
			if symlinkErr.Link != filepath.Join(mod, link) {
				t.Errorf("Link = %q, want %q", symlinkErr.Link, filepath.Join(mod, link))
			}
			if !strings.Contains(err.Error(), filepath.Join(mod, link)) || !strings.Contains(err.Error(), testEnv) {
				t.Errorf("error %q must name the link and the setting", err)
			}
		})
	}
}

func TestCheckSymlinks_LoopIsRefused(t *testing.T) {
	root, mod := modTree(t)
	if err := set(t, filepath.Join(root, "in")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("loop", filepath.Join(mod, "loop")); err != nil {
		t.Fatal(err)
	}
	err := CheckSymlinks(mod)
	var symlinkErr *SymlinkError
	if !errors.As(err, &symlinkErr) || !errors.Is(err, ErrSymlinkLoop) || !strings.Contains(err.Error(), filepath.Join(mod, "loop")) {
		t.Errorf("CheckSymlinks = %v, want a SymlinkError naming the loop", err)
	}
}

func TestCheckSymlinks_UnsetReadsNothing(t *testing.T) {
	root, mod := modTree(t)
	if err := os.Symlink("../../out/b.txt", filepath.Join(mod, "evil")); err != nil {
		t.Fatal(err)
	}
	if err := set(t, ""); err != nil {
		t.Fatal(err)
	}
	if err := CheckSymlinks(mod); err != nil {
		t.Errorf("CheckSymlinks with no limit = %v, want nil", err)
	}
	// an unreadable setting fails closed
	if err := set(t, "relative/x"); err == nil {
		t.Fatal("want the relative entry refused")
	}
	if err := CheckSymlinks(filepath.Join(root, "in")); err == nil {
		t.Error("CheckSymlinks must fail closed while the setting is unreadable")
	}
}
