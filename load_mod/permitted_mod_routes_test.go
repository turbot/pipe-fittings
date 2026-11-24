package load_mod_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/load_mod"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/versionmap"
	"github.com/zclconf/go-cty/cty"
)

const (
	modRoutesPrefix    = "TESTAPP_"
	modRoutesEnv       = modRoutesPrefix + "PERMITTED_ROOT_PATHS"
	modRoutesFixture   = "../workspace/test_data/permitted_db_routes"
	outsidePlaceholder = "@OUTSIDE@"
	upPlaceholder      = "@UP@"
)

// modRoute is one line of the test mod's mod_routes.txt
type modRoute struct {
	kind, name, value string
	refused           bool
}

// readModRoutes reads the test mod's mod_routes.txt
func readModRoutes(t *testing.T) []modRoute {
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
		routes = append(routes, modRoute{kind: fields[0], refused: fields[1] == "refused", name: fields[2], value: fields[3]})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return routes
}

// routesOfKind returns the routes of one kind, and fails if there are none
func routesOfKind(t *testing.T, kind string) []modRoute {
	t.Helper()
	var res []modRoute
	for _, r := range readModRoutes(t) {
		if r.kind == kind {
			res = append(res, r)
		}
	}
	if len(res) == 0 {
		t.Fatalf("mod_routes.txt has no %s lines", kind)
	}
	return res
}

// modRoutesDirs is
//
//	<base>/permitted/          a copy of the test mod, the one permitted folder when the setting is set
//	<base>/permitted/child/    a mod
//	<base>/outside/            a folder that is not permitted, holding secret.sql and a mod in child/
type modRoutesDirs struct {
	permitted, outside string
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

// modRoutesResources is the smallest ModResources a mod needs to be loaded: the test mods hold no resources but
// their mod blocks (the app registers its own in a real load)
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

// newModRoutesDirs copies the test mod into a temp folder (filling the outside placeholder in every file), adds
// the two child mods, sets the app constants a mod load reads, and makes the copy the permitted folder when
// permit is set
func newModRoutesDirs(t *testing.T, permit bool) modRoutesDirs {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d := modRoutesDirs{permitted: filepath.Join(base, "permitted"), outside: filepath.Join(base, "outside")}
	for _, dir := range []string{filepath.Join(d.permitted, "child"), filepath.Join(d.outside, "child")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "mod.pp"), []byte(`mod "child" {}`+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(d.outside, "secret.sql"), []byte("select 'secret'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(modRoutesFixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		content, err := os.ReadFile(filepath.Join(modRoutesFixture, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.ReplaceAll(string(content), outsidePlaceholder, d.outside))
		if err := os.WriteFile(filepath.Join(d.permitted, e.Name()), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	prevExt, prevDataDir := app_specific.ModDataExtensions, app_specific.WorkspaceDataDir
	app_specific.ModDataExtensions = []string{".pp"}
	app_specific.WorkspaceDataDir = ".testapp"
	prevResources := modconfig.AppSpecificNewModResourcesFunc
	modconfig.AppSpecificNewModResourcesFunc = func(*modconfig.Mod, ...modconfig.ModResources) modconfig.ModResources {
		return modRoutesResources{}
	}
	prevDecoder := parse.ModDecoderFunc
	parse.ModDecoderFunc = func(...parse.DecoderOption) parse.Decoder {
		d := parse.NewDecoderImpl()
		return &d
	}
	t.Cleanup(func() {
		app_specific.ModDataExtensions, app_specific.WorkspaceDataDir = prevExt, prevDataDir
		modconfig.AppSpecificNewModResourcesFunc = prevResources
		parse.ModDecoderFunc = prevDecoder
	})

	if permit {
		setModRoutesPermitted(t, d.permitted)
	} else {
		setModRoutesPermitted(t, "")
	}
	return d
}

// value fills the placeholders in a route's value
func (d modRoutesDirs) value(r modRoute) string {
	return strings.NewReplacer(upPlaceholder, strings.Repeat("../", 64), outsidePlaceholder, d.outside).Replace(r.value)
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

// loadRoot loads the mod in folder as the root mod, as a workspace does: its lock file, then the mod
func loadRoot(t *testing.T, folder string) error {
	t.Helper()
	lock, err := versionmap.LoadWorkspaceLock(folder)
	if err != nil {
		return err
	}
	parseCtx, err := parse.NewModParseContext(lock, folder)
	if err != nil {
		t.Fatal(err)
	}
	parseCtx.EvalCtx.Variables["var"] = cty.ObjectVal(map[string]cty.Value{"database": cty.StringVal("")})
	_, ew := load_mod.LoadMod(context.Background(), folder, parseCtx)
	return ew.GetError()
}

// writeRoot writes a root mod in <permitted>/root that requires the mod at the local path depPath, and its lock
// file, as an install leaves them
func writeRoot(t *testing.T, d modRoutesDirs, depPath string) string {
	t.Helper()
	root := filepath.Join(d.permitted, "root")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	modFile := fmt.Sprintf("mod \"root\" {\n  require {\n    mod %q {\n      path = %q\n    }\n  }\n}\n", depPath, depPath)
	if err := os.WriteFile(filepath.Join(root, "mod.pp"), []byte(modFile), 0o600); err != nil {
		t.Fatal(err)
	}
	lock := map[string]map[string]map[string]string{
		"root": {depPath: {"name": depPath, "path": depPath, "alias": filepath.Base(depPath)}},
	}
	content, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepaths.WorkspaceLockPath(root), content, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
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

// TestPermittedModLoadRoutes loads the test mod with each symlink route of mod_routes.txt, as the root mod and
// as a dependency, and with each lock route, with the setting set (the refused routes fail, naming the link or
// the path) and unset (every route loads, as it did before the setting covered mods)
func TestPermittedModLoadRoutes(t *testing.T) {
	for _, permit := range []bool{true, false} {
		setting := "set"
		if !permit {
			setting = "unset"
		}
		for _, r := range routesOfKind(t, "symlink") {
			t.Run("setting "+setting+"/root/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				d.makeLink(t, d.permitted, r)
				assertRoute(t, r, permit, loadRoot(t, d.permitted), r.name)
			})
			t.Run("setting "+setting+"/dependency/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				d.makeLink(t, d.permitted, r)
				root := writeRoot(t, d, d.permitted)
				assertRoute(t, r, permit, loadRoot(t, root), r.name)
			})
		}
		for _, r := range routesOfKind(t, "lock") {
			t.Run("setting "+setting+"/lock/"+r.name, func(t *testing.T) {
				d := newModRoutesDirs(t, permit)
				depPath := d.value(r)
				if !filepath.IsAbs(depPath) {
					depPath = filepath.Join(d.permitted, depPath)
				}
				root := writeRoot(t, d, depPath)
				assertRoute(t, r, permit, loadRoot(t, root), depPath)
			})
		}
	}
}
