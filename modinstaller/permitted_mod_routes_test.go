package modinstaller

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
)

const (
	modRoutesPrefix    = "TESTAPP_"
	modRoutesEnv       = modRoutesPrefix + "PERMITTED_ROOT_PATHS"
	modRoutesFixture   = "../workspace/test_data/permitted_db_routes"
	outsidePlaceholder = "@OUTSIDE@"
	upPlaceholder      = "@UP@"
	// the mod a git route clones, and the branch it clones
	clonedModName   = "github.com/org/cloned"
	clonedModBranch = "main"
)

// modRoute is one line of the test mod's mod_routes.txt
type modRoute struct {
	kind, name, value string
	refused           bool
}

// routesOfKind reads the routes of one kind from the test mod's mod_routes.txt, and fails if there are none
func routesOfKind(t *testing.T, kind string) []modRoute {
	t.Helper()
	f, err := os.Open(filepath.Join(modRoutesFixture, "mod_routes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var routes []modRoute
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 || (fields[1] != "refused" && fields[1] != "allowed") {
			t.Fatalf("mod_routes.txt: bad line %q", line)
		}
		if fields[0] == kind {
			routes = append(routes, modRoute{kind: fields[0], refused: fields[1] == "refused", name: fields[2], value: fields[3]})
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(routes) == 0 {
		t.Fatalf("mod_routes.txt has no %s lines", kind)
	}
	return routes
}

// modRoutesDirs is
//
//	<base>/permitted/          a copy of the test mod, the one permitted folder when the setting is set
//	<base>/permitted/child/    a mod
//	<base>/permitted/root/     a workspace mod, for the routes that need one besides the test mod
//	<base>/outside/            a folder that is not permitted, holding secret.sql and a mod in child/
type modRoutesDirs struct {
	permitted, outside string
}

func (d modRoutesDirs) root() string { return filepath.Join(d.permitted, "root") }

// value fills the placeholders in a route's value
func (d modRoutesDirs) value(r modRoute) string {
	return strings.NewReplacer(upPlaceholder, strings.Repeat("../", 64), outsidePlaceholder, d.outside).Replace(r.value)
}

// setModRoutesPermitted sets the app's env var keys to a test-only prefix, sets the setting to value ("" unsets
// it) and reads it, as an app does once at start-up
func setModRoutesPermitted(t *testing.T, value string) {
	t.Helper()
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys(modRoutesPrefix)
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prev)
		permitted_paths.Reset()
	})
	t.Setenv(modRoutesEnv, value)
	if value == "" {
		if err := os.Unsetenv(modRoutesEnv); err != nil {
			t.Fatal(err)
		}
	}
	if err := permitted_paths.Load(); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// copyTestMod copies the test mod's files into folder, filling the outside placeholder
func (d modRoutesDirs) copyTestMod(t *testing.T, folder string) {
	t.Helper()
	entries, err := os.ReadDir(modRoutesFixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		content, err := os.ReadFile(filepath.Join(modRoutesFixture, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(folder, e.Name()), strings.ReplaceAll(string(content), outsidePlaceholder, d.outside))
	}
}

// makeLink makes a symlink route's link in folder
func (d modRoutesDirs) makeLink(t *testing.T, folder string, r modRoute) {
	t.Helper()
	link := filepath.Join(folder, r.name)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(d.value(r), link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
}

// modRoutesResources is the smallest ModResources a mod needs to be loaded: the test mods hold no resources but
// their mod blocks (the app registers its own in a real install)
type modRoutesResources struct{}

func (modRoutesResources) WalkResources(func(modconfig.HclResource) (bool, error)) error { return nil }
func (modRoutesResources) AddResource(modconfig.HclResource) hcl.Diagnostics             { return nil }
func (modRoutesResources) GetResource(*modconfig.ParsedResourceName) (modconfig.HclResource, bool) {
	return nil, false
}
func (modRoutesResources) Equals(modconfig.ModResources) bool                     { return true }
func (modRoutesResources) AddReference(*modconfig.ResourceReference)              {}
func (modRoutesResources) GetReferences() map[string]*modconfig.ResourceReference { return nil }
func (modRoutesResources) GetVariables() map[string]*modconfig.Variable           { return nil }
func (modRoutesResources) GetMods() map[string]*modconfig.Mod                     { return nil }
func (r modRoutesResources) TopLevelResources() modconfig.ModResources            { return r }
func (modRoutesResources) AddMaps(...modconfig.ModResources)                      {}

// newModRoutesDirs lays out the folders, sets the app constants an install reads, and makes <base>/permitted the
// permitted folder when permit is set
func newModRoutesDirs(t *testing.T, permit bool) modRoutesDirs {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := modRoutesDirs{permitted: filepath.Join(base, "permitted"), outside: filepath.Join(base, "outside")}
	d.copyTestMod(t, d.permitted)
	writeFile(t, filepath.Join(d.permitted, "child", "mod.pp"), "mod \"child\" {}\n")
	writeFile(t, filepath.Join(d.outside, "child", "mod.pp"), "mod \"child\" {}\n")
	writeFile(t, filepath.Join(d.outside, "secret.sql"), "select 'secret'\n")
	writeFile(t, filepath.Join(d.root(), "mod.pp"), "mod \"root\" {}\n")

	prevExt, prevDataDir := app_specific.ModDataExtensions, app_specific.WorkspaceDataDir
	app_specific.ModDataExtensions = []string{".pp"}
	app_specific.WorkspaceDataDir = ".testapp"
	prevResources := modconfig.AppSpecificNewModResourcesFunc
	modconfig.AppSpecificNewModResourcesFunc = func(*modconfig.Mod, ...modconfig.ModResources) modconfig.ModResources {
		return modRoutesResources{}
	}
	t.Cleanup(func() {
		app_specific.ModDataExtensions, app_specific.WorkspaceDataDir = prevExt, prevDataDir
		modconfig.AppSpecificNewModResourcesFunc = prevResources
	})

	if permit {
		setModRoutesPermitted(t, d.permitted)
	} else {
		setModRoutesPermitted(t, "")
	}
	return d
}

// newRoutesInstaller makes an installer for the workspace mod in folder, as `mod install` does
func newRoutesInstaller(t *testing.T, folder string, force bool) (*ModInstaller, error) {
	t.Helper()
	workspaceMod := newTestMod(filepath.Base(folder), folder)
	workspaceMod.Require = modconfig.NewRequire()
	return NewModInstaller(&InstallOpts{WorkspaceMod: workspaceMod, Command: "install", Force: force})
}

// makeSourceRepo makes a git repo holding a copy of the test mod with the route's link, on the branch the git
// routes clone, and points the installer's clone at it
func (d modRoutesDirs) makeSourceRepo(t *testing.T, r modRoute) {
	t.Helper()
	source := filepath.Join(filepath.Dir(d.permitted), "source")
	d.copyTestMod(t, source)
	d.makeLink(t, source, r)
	repo, err := git.PlainInitWithOptions(source, &git.PlainInitOptions{
		InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName(clonedModBranch)},
	})
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if err := wt.AddGlob("."); err != nil {
		t.Fatal(err)
	}
	sig := &object.Signature{Name: "test", Email: "test@example.com", When: time.Unix(0, 0)}
	if _, err := wt.Commit("test mod", &git.CommitOptions{Author: sig, Committer: sig}); err != nil {
		t.Fatal(err)
	}
	prev := modGitUrl
	modGitUrl = func(string, GitUrlMode) string { return "file://" + filepath.ToSlash(source) }
	t.Cleanup(func() { modGitUrl = prev })
}

func assertRoute(t *testing.T, r modRoute, permit bool, err error, names string) {
	t.Helper()
	if permit && r.refused {
		if err == nil {
			t.Fatalf("expected the route to be refused")
		}
		if !strings.Contains(err.Error(), names) || !strings.Contains(err.Error(), modRoutesEnv) {
			t.Errorf("the error should name %q and the setting, got: %v", names, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("expected the route to work, got: %v", err)
	}
}

func settings() map[string]bool { return map[string]bool{"setting set": true, "setting unset": false} }

// TestPermittedModInstallRoutes installs with each route of the test mod's mod_routes.txt, with the setting set
// (the refused routes fail, naming the link or the path, and nothing is cloned or copied) and unset (every route
// installs, as it did before the setting covered mods)
func TestPermittedModInstallRoutes(t *testing.T) {
	for setting, permit := range settings() {
		for _, r := range routesOfKind(t, "symlink") {
			// the workspace mod holds the link
			t.Run(setting+"/root/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				d.makeLink(t, d.permitted, r)
				i, err := newRoutesInstaller(t, d.permitted, true)
				if err != nil {
					t.Fatal(err)
				}
				// --force does not let it through
				assertRoute(t, r, permit, i.InstallWorkspaceDependencies(context.Background()), r.name)
			})
			// a mod cloned from git holds the link; with --force the install goes on past the refusal, and the
			// refused clone must still not reach the mods folder
			for _, force := range []bool{false, true} {
				name := setting + "/git/" + r.name
				if force {
					name += "/force"
				}
				t.Run(name, func(t *testing.T) {
					d := newModRoutesDirs(t, permit)
					d.makeSourceRepo(t, r)
					i, err := newRoutesInstaller(t, d.root(), force)
					if err != nil {
						t.Fatal(err)
					}
					i.workspaceMod.Require.Mods = []*modconfig.ModVersionConstraint{{Name: clonedModName, BranchName: clonedModBranch}}
					err = i.installMods(context.Background(), i.workspaceMod)
					installed := filepath.Join(i.modsPath, modconfig.BuildModDependencyPath(clonedModName, &modconfig.DependencyVersion{Branch: clonedModBranch}))
					_, statErr := os.Stat(filepath.Join(installed, "mod.pp"))
					if permit && r.refused {
						if !force {
							assertRoute(t, r, permit, err, r.name)
						}
						if statErr == nil {
							t.Errorf("the refused mod was copied to the mods folder")
						}
						return
					}
					assertRoute(t, r, permit, err, r.name)
					if statErr != nil {
						t.Errorf("the mod was not installed: %v", statErr)
					}
					if _, err := os.Lstat(filepath.Join(installed, r.name)); err != nil {
						t.Errorf("the installed mod lost its link: %v", err)
					}
				})
			}
			// a mod required by its local path holds the link
			t.Run(setting+"/local/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				d.makeLink(t, d.permitted, r)
				i, err := newRoutesInstaller(t, d.root(), false)
				if err != nil {
					t.Fatal(err)
				}
				i.workspaceMod.Require.Mods = []*modconfig.ModVersionConstraint{{Name: d.permitted, FilePath: d.permitted}}
				assertRoute(t, r, permit, i.installMods(context.Background(), i.workspaceMod), r.name)
			})
		}

		// the test mod requires a mod by its local path
		for _, r := range routesOfKind(t, "require") {
			t.Run(setting+"/require/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				i, err := newRoutesInstaller(t, d.permitted, false)
				if err != nil {
					t.Fatal(err)
				}
				value := d.value(r)
				i.workspaceMod.Require.Mods = []*modconfig.ModVersionConstraint{{Name: value, FilePath: value}}
				err = i.installMods(context.Background(), i.workspaceMod)
				names := value
				if !filepath.IsAbs(names) {
					names = filepath.Join(d.permitted, value)
				}
				assertRoute(t, r, permit, err, names)
				if _, err := os.Stat(i.modsPath); err == nil {
					t.Errorf("nothing should be copied to the mods folder for a local mod")
				}
			})
		}

		// the test mod's lock file names a local path
		for _, r := range routesOfKind(t, "lock") {
			t.Run(setting+"/lock/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				value := d.value(r)
				if !filepath.IsAbs(value) {
					value = filepath.Join(d.permitted, value)
				}
				lock := map[string]map[string]map[string]string{
					"permitted": {value: {"name": value, "path": value, "alias": "child"}},
				}
				content, err := json.Marshal(lock)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepaths.WorkspaceLockPath(d.permitted), string(content))
				// the lock is read before anything is installed
				_, err = newRoutesInstaller(t, d.permitted, false)
				assertRoute(t, r, permit, err, value)
			})
		}
	}
}

// TestValidateRequirementLocalPath: the require check itself refuses a local path outside the permitted folders,
// before the install looks at the folder
func TestValidateRequirementLocalPath(t *testing.T) {
	for setting, permit := range settings() {
		for _, r := range routesOfKind(t, "require") {
			t.Run(setting+"/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				i, err := newRoutesInstaller(t, d.permitted, false)
				if err != nil {
					t.Fatal(err)
				}
				value := d.value(r)
				names := value
				if !filepath.IsAbs(names) {
					names = filepath.Join(d.permitted, value)
				}
				assertRoute(t, r, permit, i.validateRequirement(&modconfig.ModVersionConstraint{Name: value, FilePath: value}, i.workspaceMod), names)
			})
		}
	}
}

// TestPermittedInstalledModRead: a mod installed before the setting was set, whose mod file now links outside the
// permitted folders, is refused by the reads an uninstall or update makes of installed mods (which do not walk
// the workspace first, as an install does), naming the link; with the setting unset it is read as before
func TestPermittedInstalledModRead(t *testing.T) {
	for setting, permit := range settings() {
		t.Run(setting, func(t *testing.T) {
			d := newModRoutesDirs(t, false)
			d.makeSourceRepo(t, modRoute{kind: "symlink", name: "inside_file.sql", value: "init.sql"})
			i, err := newRoutesInstaller(t, d.root(), false)
			if err != nil {
				t.Fatal(err)
			}
			cloned := &modconfig.ModVersionConstraint{Name: clonedModName, BranchName: clonedModBranch}
			i.workspaceMod.Require.Mods = []*modconfig.ModVersionConstraint{cloned}
			if err := i.installMods(context.Background(), i.workspaceMod); err != nil {
				t.Fatal(err)
			}
			if err := i.installData.Lock.Save(); err != nil {
				t.Fatal(err)
			}
			// the installed mod file becomes a link to a mod outside
			installed := filepath.Join(i.modsPath, modconfig.BuildModDependencyPath(clonedModName, &modconfig.DependencyVersion{Branch: clonedModBranch}))
			link := filepath.Join(installed, "mod.pp")
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(d.outside, "child", "mod.pp"), link); err != nil {
				t.Skipf("cannot create symlink: %v", err)
			}
			if permit {
				setModRoutesPermitted(t, d.permitted)
			}

			// uninstalling another mod reads the installed one, as the install step of an uninstall does
			un, err := newRoutesInstaller(t, d.root(), false)
			if err != nil {
				t.Fatal(err)
			}
			un.command = "uninstall"
			_, err = un.getModForRequirement(context.Background(), cloned, false)
			r := modRoute{refused: true}
			assertRoute(t, r, permit, err, link)

			// an update's check for a newer commit opens the installed repo
			locked, lockErr := un.installData.Lock.FindLockedModVersion(cloned)
			if lockErr != nil || locked == nil {
				t.Fatalf("the installed mod is not in the lock: %v", lockErr)
			}
			_, err = un.getLatestCommitForBranch(locked)
			if permit {
				assertRoute(t, r, permit, err, link)
			}
		})
	}
}
