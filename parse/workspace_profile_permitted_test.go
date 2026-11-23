package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
)

// a workspace profile in a .ppc file in the mod folder is mod content: its database, mod_location and local
// snapshot_location must stay inside the permitted folders (when they are set)
func TestLoadPowerpipeWorkspaceProfilesPermitted(t *testing.T) {
	// nothing here reads a home folder; HOME still points at a scratch folder so nothing could
	t.Setenv("HOME", t.TempDir())
	prevExt := app_specific.ConfigExtension
	app_specific.ConfigExtension = ".ppc"
	t.Cleanup(func() { app_specific.ConfigExtension = prevExt })

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	permitted := filepath.Join(base, "permitted")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{permitted, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	inside := filepath.Join(permitted, "x.db")
	outsideDB := filepath.Join(outside, "x.db")

	rows := []struct {
		name    string
		setting string
		profile string
		// wantErr is a part of the error; empty means the load succeeds
		wantErr string
		// want is the config map the default profile gives, for the keys it names
		want map[string]string
	}{
		{
			name:    "outside database refused with the setting set",
			setting: permitted,
			profile: `database = "sqlite://` + outsideDB + `"`,
			wantErr: "outside the folders " + testPermittedEnv,
		},
		{
			name:    "outside duckdb database refused with the setting set",
			setting: permitted,
			profile: `database = "duckdb://` + outsideDB + `"`,
			wantErr: "outside the folders " + testPermittedEnv,
		},
		{
			name:    "database with options refused with the setting set",
			setting: permitted,
			profile: `database = "duckdb://` + inside + `?access_mode=read_only"`,
			wantErr: "is not allowed while " + testPermittedEnv,
		},
		{
			name:    "inside database works with the setting set",
			setting: permitted,
			profile: `database = "sqlite://` + inside + `"`,
			want:    map[string]string{constants.ArgDatabase: "sqlite://" + inside},
		},
		{
			name:    "outside database inherited from a base refused",
			setting: permitted,
			profile: `base = workspace.other` + "\n}\nworkspace \"other\" {\n" + `database = "sqlite://` + outsideDB + `"`,
			wantErr: "outside the folders " + testPermittedEnv,
		},
		{
			name:    "outside mod_location refused with the setting set",
			setting: permitted,
			profile: `mod_location = "` + outside + `"`,
			wantErr: "mod_location",
		},
		{
			name:    "outside snapshot_location refused with the setting set",
			setting: permitted,
			profile: `snapshot_location = "` + outside + `"`,
			wantErr: "snapshot_location",
		},
		{
			name:    "inside paths work with the setting set",
			setting: permitted,
			profile: `mod_location = "` + permitted + `"` + "\n" + `snapshot_location = "` + permitted + `"`,
			want:    map[string]string{constants.ArgModLocation: permitted, constants.ArgSnapshotLocation: permitted},
		},
		{
			name:    "pipes workspace snapshot_location is not a path",
			setting: permitted,
			profile: `snapshot_location = "acme/dev"`,
			want:    map[string]string{constants.ArgSnapshotLocation: "acme/dev"},
		},
		{
			name:    "unset leaves every value unchanged",
			setting: "",
			profile: `database = "duckdb://` + outsideDB + `?access_mode=read_only"` + "\n" + `mod_location = "` + outside + `"` + "\n" + `snapshot_location = "` + outside + `"`,
			want: map[string]string{
				constants.ArgDatabase:         "duckdb://" + outsideDB + "?access_mode=read_only",
				constants.ArgModLocation:      outside,
				constants.ArgSnapshotLocation: outside,
			},
		},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			setPermittedForTest(t, r.setting)
			modDir := t.TempDir()
			content := "workspace \"default\" {\n" + r.profile + "\n}\n"
			if err := os.WriteFile(filepath.Join(modDir, "workspaces.ppc"), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			profiles, err := LoadWorkspaceProfiles[*workspace_profile.PowerpipeWorkspaceProfile](modDir)
			if r.wantErr != "" {
				if err == nil {
					t.Fatalf("want an error containing %q, got none", r.wantErr)
				}
				if !strings.Contains(err.Error(), r.wantErr) {
					t.Fatalf("want an error containing %q, got %q", r.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			profile, ok := profiles["default"]
			if !ok {
				t.Fatal("no default profile loaded")
			}
			got := profile.ConfigMap(nil)
			for key, want := range r.want {
				if got[key] != want {
					t.Fatalf("%s: got %v, want %q", key, got[key], want)
				}
			}
		})
	}
}
