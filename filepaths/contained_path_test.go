package filepaths

import (
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
