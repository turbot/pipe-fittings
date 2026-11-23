package filepaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInside(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mods")

	ok := []string{
		"github.com/org/repo",
		"github.com/org/repo@v1.2.3",
		"github.com/org/repo#feature/x",
		"github.com/org/repo@v1.2.3-rc.1+build",
		"github.com/org/repo#..foo", // not a '..' element
	}
	for _, rel := range ok {
		got, err := ResolveInside(root, rel)
		if err != nil {
			t.Errorf("ResolveInside(%q) unexpected error: %v", rel, err)
			continue
		}
		if want := filepath.Join(root, rel); got != want {
			t.Errorf("ResolveInside(%q) = %q, want %q", rel, got, want)
		}
	}

	bad := []string{
		"",
		".",
		"..",
		"../x",
		"github.com/org/repo/../../../../x",
		"github.com/org/repo#x/../../../../../x",
		"github.com/org/repo@../../x",
		`github.com\org\..\..\x`,
		"/etc/passwd",
		`\windows\system32`,
		"a/..",
		"github.com/org/repo\x00x",
	}
	for _, rel := range bad {
		if got, err := ResolveInside(root, rel); err == nil {
			t.Errorf("ResolveInside(%q) = %q, want an error", rel, got)
		}
	}
}

func TestResolveInsideRefusesSymlink(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "mods")
	outside := filepath.Join(tmp, "outside")
	for _, d := range []string{filepath.Join(root, "github.com", "org", "a@v1.0.0"), outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// a mod's repo holds a link which leads out of the mods folder
	if err := os.Symlink(outside, filepath.Join(root, "github.com", "org", "a@v1.0.0", "link#y")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	if got, err := ResolveInside(root, "github.com/org/a@v1.0.0/link#y/victim"); err == nil {
		t.Errorf("a path through a symlink was accepted: %q", got)
	}
	// a path beside it, or one which does not exist yet, is fine
	for _, rel := range []string{"github.com/org/a@v1.0.0", "github.com/org/a@v1.0.0/other#y/victim", "github.com/org/new@v2.0.0"} {
		if _, err := ResolveInside(root, rel); err != nil {
			t.Errorf("ResolveInside(%q) unexpected error: %v", rel, err)
		}
	}
	// a root which is itself a symlink is the caller's choice and still works
	link := filepath.Join(tmp, "mods-link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveInside(link, "github.com/org/new@v2.0.0"); err != nil {
		t.Errorf("a root reached through a symlink was refused: %v", err)
	}
}
