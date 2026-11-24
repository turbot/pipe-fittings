package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
)

// noResources stands in for the app's resources: a load that stops before the mod is parsed calls none of them
type noResources struct{ modconfig.ModResources }

// TestLoadRefusesWorkspaceSymlinkOutside: with the setting set, a workspace holding a symlink that leads outside
// the permitted folders is refused before any of its files is read; with the setting unset the link is no reason
// to refuse it
func TestLoadRefusesWorkspaceSymlinkOutside(t *testing.T) {
	const prefix = "TESTAPP_"
	for _, permit := range []bool{true, false} {
		t.Run(map[bool]string{true: "setting set", false: "setting unset"}[permit], func(t *testing.T) {
			loadWithLink(t, prefix, permit)
		})
	}
}

// loadWithLink loads a workspace whose ignore file is a symlink to a file outside the permitted folder
func loadWithLink(t *testing.T, prefix string, permit bool) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace, outside := filepath.Join(base, "permitted"), filepath.Join(base, "outside")
	for _, dir := range []string{workspace, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// the ignore file is the first file a load reads
	if err := os.WriteFile(filepath.Join(outside, "ignore"), []byte("*.pp\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(workspace, ".testappignore")
	if err := os.Symlink("../outside/ignore", link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	prevPrefix, prevIgnore := app_specific.EnvAppPrefix, app_specific.WorkspaceIgnoreFile
	prevResources := modconfig.AppSpecificNewModResourcesFunc
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prevPrefix)
		app_specific.WorkspaceIgnoreFile = prevIgnore
		modconfig.AppSpecificNewModResourcesFunc = prevResources
		permitted_paths.Reset()
	})
	modconfig.AppSpecificNewModResourcesFunc = func(*modconfig.Mod, ...modconfig.ModResources) modconfig.ModResources {
		return noResources{}
	}
	app_specific.SetAppSpecificEnvVarKeys(prefix)
	app_specific.WorkspaceIgnoreFile = ".testappignore"
	value := ""
	if permit {
		value = workspace
	}
	t.Setenv(prefix+"PERMITTED_ROOT_PATHS", value)
	if err := permitted_paths.Load(); err != nil {
		t.Fatal(err)
	}

	// with no mod file and resources skipped, a load reads the ignore file and stops
	_, ew := Load(context.Background(), workspace, WithSkipResourceLoadIfNoModfile(true))
	err = ew.GetError()
	var symlinkErr *permitted_paths.SymlinkError
	if permit {
		if !errors.As(err, &symlinkErr) || !strings.Contains(err.Error(), link) {
			t.Errorf("setting set: want a refusal naming %s, got %v", link, err)
		}
	} else if err != nil {
		t.Errorf("setting unset: want the workspace to load, got %v", err)
	}
}
