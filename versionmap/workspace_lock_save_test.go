package versionmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

func newTestLock(t *testing.T, tag string) (*WorkspaceLock, string) {
	t.Helper()
	workspace := t.TempDir()
	lock := &WorkspaceLock{
		WorkspacePath: workspace,
		InstallCache: InstalledDependencyVersionsMap{
			"root": {
				"github.com/org/a": &InstalledModVersion{
					ResolvedVersionConstraint: &ResolvedVersionConstraint{
						DependencyVersion: modconfig.DependencyVersion{Tag: tag},
						Name:              "github.com/org/a",
					},
					Alias: "a",
				},
			},
		},
	}
	lockPath := filepaths.WorkspaceLockPath(workspace)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	return lock, lockPath
}

func setPast(t *testing.T, path string) time.Time {
	t.Helper()
	past := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	return past
}

func TestSaveLeavesIdenticalLockFileUntouched(t *testing.T) {
	lock, lockPath := newTestLock(t, "v1")
	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}
	past := setPast(t, lockPath)

	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(past) {
		t.Errorf("modified time moved from %v to %v for an unchanged lock", past, info.ModTime())
	}
}

func TestSaveWritesChangedLock(t *testing.T) {
	lock, lockPath := newTestLock(t, "v1")
	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}
	past := setPast(t, lockPath)

	changed, _ := newTestLock(t, "v2")
	lock.InstallCache = changed.InstallCache
	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), `"v2"`) || strings.Contains(string(after), `"v1"`) {
		t.Errorf("lock file does not hold the new content: %s", after)
	}
	info, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().After(past) {
		t.Error("modified time did not move for a changed lock")
	}
}

func TestSaveWritesWhenLockFileMissing(t *testing.T) {
	lock, lockPath := newTestLock(t, "v1")
	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("lock file not written: %v", err)
	}
	if !strings.Contains(string(after), `"v1"`) {
		t.Errorf("unexpected content: %s", after)
	}
}

func TestSaveWritesWhenLockFileCannotBeRead(t *testing.T) {
	lock, lockPath := newTestLock(t, "v1")
	// a directory sits where the file should be: it cannot be read as a file, and the write fails loudly
	if err := os.Mkdir(lockPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := lock.Save(); err == nil {
		t.Fatal("expected the failed write to be reported")
	}
}
