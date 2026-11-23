package versionmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/filepaths"
)

func writeLock(t *testing.T, content string) string {
	t.Helper()
	workspace := t.TempDir()
	lockPath := filepaths.WorkspaceLockPath(workspace)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return workspace
}

func TestLoadWorkspaceLockRefusesEntriesOutsideModsFolder(t *testing.T) {
	cases := map[string]string{
		"branch": `{"root":{"github.com/org/repo":{"name":"github.com/org/repo","branch":"x/../../../../outside","alias":"r"}}}`,
		"tag":    `{"root":{"github.com/org/repo":{"name":"github.com/org/repo","tag":"../../../outside","alias":"r"}}}`,
		"name":   `{"root":{"github.com/org/repo":{"name":"../../../../outside","version":"1.0.0","alias":"r"}}}`,
		"child":  `{"root":{"github.com/org/ok":{"name":"github.com/org/ok","version":"1.0.0","alias":"ok"}},"github.com/org/ok@v1.0.0":{"github.com/org/repo":{"name":"/etc","tag":"t","alias":"r"}}}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			workspace := writeLock(t, content)
			_, err := LoadWorkspaceLock(workspace)
			if err == nil {
				t.Fatal("expected the lock file to be refused")
			}
			if !strings.Contains(err.Error(), "outside the mods folder") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadWorkspaceLockAcceptsOrdinaryEntries(t *testing.T) {
	content := `{"root":{
	  "github.com/org/a":{"name":"github.com/org/a","version":"1.2.3","alias":"a"},
	  "github.com/org/b":{"name":"github.com/org/b","branch":"feature/x","alias":"b"},
	  "github.com/org/c":{"name":"github.com/org/c","tag":"v1.2.3","alias":"c"}}}`
	workspace := writeLock(t, content)
	if _, err := LoadWorkspaceLock(workspace); err != nil {
		t.Fatalf("ordinary lock file refused: %v", err)
	}
}

func TestFindInstalledDependencyRefusesEscape(t *testing.T) {
	l := &WorkspaceLock{ModInstallationPath: t.TempDir()}
	dep := &ResolvedVersionConstraint{Name: "github.com/org/repo"}
	dep.Branch = "x/../../../../outside"
	if _, err := l.FindInstalledDependency(dep); err == nil || !strings.Contains(err.Error(), "outside the mods folder") {
		t.Errorf("expected refusal, got %v", err)
	}
}
