package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}
}

func TestWriteFileAtomicCreatesFileWithDefaultPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")
	if err := WriteFileAtomic(path, []byte("hello"), 0o640); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "hello" {
		t.Errorf("content = %q", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o640 {
		t.Errorf("permissions = %v, want 0640", info.Mode().Perm())
	}
	assertNoTempFiles(t, dir)
}

func TestWriteFileAtomicKeepsPermissionsOfExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mod.pp")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permissions = %v, want the old 0600", info.Mode().Perm())
	}
	if got := readFile(t, path); got != "new" {
		t.Errorf("content = %q", got)
	}
	assertNoTempFiles(t, dir)
}

func TestWriteFileAtomicReplacesTheFileByRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mod.pp")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// a rename puts a new file at the path; an in-place write would keep the old one
	if os.SameFile(before, after) {
		t.Error("the file was written in place, not replaced")
	}
}

func TestWriteFileAtomicKeepsASymlink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real", "mod.pp")
	if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "mod.pp")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	if err := WriteFileAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("the symlink was replaced by a regular file")
	}
	if got := readFile(t, real); got != "new" {
		t.Errorf("the file the link points to holds %q, want the new content", got)
	}
	assertNoTempFiles(t, dir)
	assertNoTempFiles(t, filepath.Dir(real))
}

func TestWriteFileAtomicFailureLeavesTheFileAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-folder", "mod.pp")
	if err := WriteFileAtomic(path, []byte("new"), 0o644); err == nil {
		t.Fatal("expected an error writing into a folder that does not exist")
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("a file was written")
	}

	// a rename onto a folder fails: the old folder must stay and no temporary file may be left
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(target, []byte("new"), 0o644); err == nil {
		t.Fatal("expected an error replacing a folder")
	}
	assertNoTempFiles(t, dir)
}
