package cmdconfig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
)

// a relative entry in <APP>_PERMITTED_ROOT_PATHS stops the app at bootstrap, before it reads anything else, and names the entry
func TestBootstrapViper_RelativePermittedRootPathStopsTheApp(t *testing.T) {
	prev := app_specific.EnvAppPrefix
	app_specific.SetAppSpecificEnvVarKeys("TESTAPP_")
	t.Cleanup(func() {
		app_specific.SetAppSpecificEnvVarKeys(prev)
		permitted_paths.Reset()
	})
	t.Setenv("TESTAPP_PERMITTED_ROOT_PATHS", "/srv/ok,relative/dir")

	defer func() {
		got := fmt.Sprint(recover())
		for _, want := range []string{"TESTAPP_PERMITTED_ROOT_PATHS", `"relative/dir"`} {
			if !strings.Contains(got, want) {
				t.Errorf("want %q in the message, got: %s", want, got)
			}
		}
	}()
	// the loader is never reached
	BootstrapViper[workspace_profile.WorkspaceProfile](nil, nil)
	t.Fatal("want the app to stop")
}
