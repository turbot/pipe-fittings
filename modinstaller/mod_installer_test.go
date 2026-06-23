package modinstaller

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestCommitShadowOverwritesReadOnlyFiles reproduces a mod re-install/upgrade where the
// destination already holds read-only git pack files (as written by go-git >= v5.17).
// (pipe-fittings' own go.mod pins go-git below this; consumers such as flowpipe float it higher.)
// commitShadow must overwrite them rather than fail with "permission denied".
func TestCommitShadowOverwritesReadOnlyFiles(t *testing.T) {
	tmp := t.TempDir()
	modsPath := filepath.Join(tmp, "mods")
	shadowPath := filepath.Join(tmp, "shadow")

	// a previously installed mod, with a read-only pack file (mimics go-git output)
	packRel := filepath.Join("github.com", "turbot", "mod", ".git", "objects", "pack", "pack-abc.idx")
	destPack := filepath.Join(modsPath, packRel)
	if err := os.MkdirAll(filepath.Dir(destPack), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destPack, []byte("old"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(destPack, 0o444); err != nil {
		t.Fatal(err)
	}

	// the shadow directory holds the new version of the same file, also read-only
	shadowPack := filepath.Join(shadowPath, packRel)
	if err := os.MkdirAll(filepath.Dir(shadowPack), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shadowPack, []byte("new"), 0o444); err != nil {
		t.Fatal(err)
	}

	i := &ModInstaller{modsPath: modsPath, shadowDirPath: shadowPath}
	if err := i.commitShadow(context.Background()); err != nil {
		t.Fatalf("commitShadow failed overwriting read-only destination: %v", err)
	}

	got, err := os.ReadFile(destPack)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("destination not updated: got %q, want %q", string(got), "new")
	}
}
