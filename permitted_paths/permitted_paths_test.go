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
