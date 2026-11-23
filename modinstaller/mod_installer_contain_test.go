package modinstaller

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/versionmap"
)

const outsideMsg = "outside the mods folder"

// build a mod directly: modconfig.NewMod needs the app to register its resource constructor
func newTestMod(shortName, modPath string) *modconfig.Mod {
	m := &modconfig.Mod{ModPath: modPath}
	m.ShortName = shortName
	return m
}

func newTestInstaller(t *testing.T) (*ModInstaller, string) {
	t.Helper()
	tmp := t.TempDir()
	workspace := filepath.Join(tmp, "workspace")
	modsPath := filepath.Join(workspace, ".mods", "mods")
	if err := os.MkdirAll(modsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	workspaceMod := newTestMod("root", workspace)
	workspaceMod.Require = modconfig.NewRequire()
	lock := &versionmap.WorkspaceLock{
		WorkspacePath:       workspace,
		ModInstallationPath: modsPath,
		InstallCache:        make(versionmap.InstalledDependencyVersionsMap),
		MissingVersions:     make(versionmap.InstalledDependencyVersionsMap),
	}
	i := &ModInstaller{
		workspacePath: workspace,
		workspaceMod:  workspaceMod,
		modsPath:      modsPath,
		shadowDirPath: filepath.Join(workspace, ".mods", "mods.shadow"),
		installData:   NewInstallData(lock, workspaceMod),
		command:       "install",
	}
	return i, tmp
}

// a require which resolves to a folder outside the mods folder
func evilConstraints() map[string]*modconfig.ModVersionConstraint {
	return map[string]*modconfig.ModVersionConstraint{
		"name":        {Name: "github.com/org/../../../../outside", Tag: "v1"},
		"name no ver": {Name: "../../../../outside"},
		"absolute":    {Name: "/tmp/outside"},
		"branch":      {Name: "github.com/org/repo", BranchName: "x/../../../../../../outside"},
		"tag":         {Name: "github.com/org/repo", Tag: "x/../../../../../outside"},
		"version":     {Name: "github.com/org/repo", VersionString: "../../../../outside"},
	}
}

func TestValidateRequirement(t *testing.T) {
	i, _ := newTestInstaller(t)

	for name, c := range evilConstraints() {
		t.Run("refuses "+name, func(t *testing.T) {
			err := i.validateRequirement(c)
			if err == nil {
				t.Fatalf("expected %+v to be refused", c)
			}
			if !strings.Contains(err.Error(), outsideMsg) {
				t.Errorf("error should say the mod would be outside the mods folder, got: %v", err)
			}
		})
	}

	ok := []string{
		"github.com/org/repo",
		"github.com/org/repo@v1.2.3",
		"github.com/org/repo@^1.2",
		"github.com/org/repo@latest",
	}
	for _, arg := range ok {
		c, err := modconfig.NewModVersionConstraint(arg)
		if err != nil {
			t.Fatal(err)
		}
		if err := i.validateRequirement(c); err != nil {
			t.Errorf("%s should be accepted: %v", arg, err)
		}
	}
	accepted := []*modconfig.ModVersionConstraint{
		{Name: "github.com/org/repo", BranchName: "feature/x"},
		{Name: "github.com/org/repo", Tag: "v1.2.3"},
		{Name: "github.com/org/repo", VersionString: "v1.2.3"},
		// a local path dependency is not installed under the mods folder
		{Name: "local", FilePath: "../elsewhere"},
	}
	for _, c := range accepted {
		if err := i.validateRequirement(c); err != nil {
			t.Errorf("%+v should be accepted: %v", c, err)
		}
	}
}

// an install whose require would escape the mods folder fails, and touches nothing outside it
func TestInstallRefusesRequireOutsideModsFolder(t *testing.T) {
	for name, c := range evilConstraints() {
		t.Run(name, func(t *testing.T) {
			i, tmp := newTestInstaller(t)

			// a folder an escaping path would land on (it would be deleted before a reinstall)
			outside := filepath.Join(tmp, "outside")
			sentinel := filepath.Join(outside, "keep.txt")
			if err := os.MkdirAll(outside, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}

			i.workspaceMod.Require.Mods = []*modconfig.ModVersionConstraint{c}
			err := i.installMods(context.Background(), i.workspaceMod)
			if err == nil {
				t.Fatal("expected the install to be refused")
			}
			if !strings.Contains(err.Error(), outsideMsg) {
				t.Errorf("error should say the mod would be outside the mods folder, got: %v", err)
			}
			if b, err := os.ReadFile(sentinel); err != nil || string(b) != "keep" {
				t.Errorf("file outside the mods folder was changed: %q, %v", b, err)
			}
			if _, err := os.Stat(i.shadowDirPath); err == nil {
				t.Errorf("nothing should have been written to the shadow folder")
			}
		})
	}
}

// the same holds for a dependency of a dependency
func TestInstallRefusesChildRequireOutsideModsFolder(t *testing.T) {
	for name, c := range evilConstraints() {
		t.Run(name, func(t *testing.T) {
			i, _ := newTestInstaller(t)

			parentMod := newTestMod("parent", filepath.Join(i.modsPath, "github.com", "org", "parent@v1.0.0"))
			parentMod.Require = modconfig.NewRequire()
			parentMod.Require.Mods = []*modconfig.ModVersionConstraint{c}

			parentConstraint := &modconfig.ModVersionConstraint{Name: "github.com/org/parent", VersionString: "v1.0.0"}
			parentDep := &DependencyMod{
				Mod: parentMod,
				InstalledVersion: &versionmap.InstalledModVersion{
					ResolvedVersionConstraint: versionmap.NewResolvedVersionConstraint(&modconfig.DependencyVersion{Tag: "v1.0.0"}, "github.com/org/parent", nil),
				},
			}

			err := i.installModDependenciesRecursively(context.Background(), parentConstraint, parentDep, i.workspaceMod, true)
			if err == nil {
				t.Fatal("expected the child dependency to be refused")
			}
			if !strings.Contains(err.Error(), outsideMsg) {
				t.Errorf("error should say the mod would be outside the mods folder, got: %v", err)
			}
		})
	}
}

// the paths the installer builds from a dependency name are checked wherever they are used
func TestDependencyPathsStayInside(t *testing.T) {
	i, _ := newTestInstaller(t)
	evil := "github.com/org/repo#x/../../../../../../outside"

	if _, err := i.getDependencyDestPath(evil); err == nil {
		t.Error("getDependencyDestPath accepted an escaping path")
	}
	if _, err := i.getDependencyShadowPath(evil); err == nil {
		t.Error("getDependencyShadowPath accepted an escaping path")
	}
	if _, err := i.loadDependencyModFromRoot(context.Background(), i.modsPath, evil); err == nil {
		t.Error("loadDependencyModFromRoot accepted an escaping path")
	}
	// the commit checks open the repo in the mods folder
	v := &versionmap.InstalledModVersion{ResolvedVersionConstraint: versionmap.NewResolvedVersionConstraint(&modconfig.DependencyVersion{Branch: "x/../../../../../../outside"}, "github.com/org/repo", nil)}
	if _, err := i.getLatestCommitForBranch(v); err == nil || !strings.Contains(err.Error(), "invalid dependency") {
		t.Errorf("getLatestCommitForBranch should refuse an escaping branch, got %v", err)
	}

	for _, ok := range []string{"github.com/org/repo@v1.2.3", "github.com/org/repo#feature/x"} {
		if _, err := i.getDependencyDestPath(ok); err != nil {
			t.Errorf("%s refused: %v", ok, err)
		}
	}
}

// a symlink inside an installed mod's repo must not lead a child dependency out of the mods folder
func TestInstallRefusesChildThroughSymlink(t *testing.T) {
	i, tmp := newTestInstaller(t)
	victim := filepath.Join(tmp, "home", "victim")
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(victim, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	parentDir := filepath.Join(i.shadowDirPath, "github.com", "org", "a@v1.0.0")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(tmp, "home"), filepath.Join(parentDir, "link#y")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	c := &modconfig.ModVersionConstraint{Name: "github.com/org/a@v1.0.0/link", BranchName: "y/victim"}
	if err := i.validateRequirement(c); err == nil {
		t.Fatal("expected a require reaching a symlink to be refused")
	}
	if _, err := i.getDependencyShadowPath("github.com/org/a@v1.0.0/link#y/victim"); err == nil {
		t.Fatal("expected the shadow path through a symlink to be refused")
	}
	if b, err := os.ReadFile(sentinel); err != nil || string(b) != "keep" {
		t.Errorf("file outside the mods folder was changed: %q, %v", b, err)
	}
}

// once a local 'path' dependency is no longer required, prune has nothing to delete for it and does not fail
func TestPruneIgnoresRemovedLocalPathDependency(t *testing.T) {
	i, tmp := newTestInstaller(t)
	local := filepath.Join(tmp, "localmod")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	lockJSON := `{"root":{"local":{"name":"` + filepath.ToSlash(local) + `","path":"` + filepath.ToSlash(local) + `","alias":"local"}}}`
	lockPath := filepaths.WorkspaceLockPath(i.workspacePath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, []byte(lockJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	lock, err := versionmap.LoadWorkspaceLock(i.workspacePath)
	if err != nil {
		t.Fatal(err)
	}
	// the install completes with nothing required: the new lock is empty but shares the installed mods
	i.installData = NewInstallData(lock, i.workspaceMod)
	i.installData.Lock = i.installData.NewLock

	unused, err := i.Prune()
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}
	if len(unused) != 1 {
		t.Errorf("expected the local mod to be listed as unreferenced, got %v", unused)
	}
	if _, err := os.Stat(local); err != nil {
		t.Errorf("local mod folder was touched: %v", err)
	}
}

// a mods folder with a folder a person has linked in is still readable (a symlink is only refused where cloned)
func TestModsFolderSymlinkStillResolves(t *testing.T) {
	i, tmp := newTestInstaller(t)
	linked := filepath.Join(tmp, "clone")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(linked, filepath.Join(i.modsPath, "github.com")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	if _, err := i.getDependencyDestPath("github.com/org/a@v1.0.0"); err != nil {
		t.Errorf("a mods folder holding a symlink was refused: %v", err)
	}
}
