package modinstaller

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

// some predefined update checkers
type mockUpdateChecker struct {
	updateStrategy  string
	updateAvailable bool
	commitAvailable bool ``
}

func (m mockUpdateChecker) newerVersionAvailable(_ *modconfig.ModVersionConstraint, _ *semver.Version) (bool, error) {
	return m.updateAvailable, nil
}

func (m mockUpdateChecker) newCommitAvailable(_ *versionmap.InstalledModVersion) (bool, error) {
	return m.commitAvailable, nil
}

func (m mockUpdateChecker) getUpdateStrategy() string {
	return m.updateStrategy
}

// ctor
func newMockUpdateChecker(updateStrategy string, updateAvailable bool, commitAvailable bool) *mockUpdateChecker {
	return &mockUpdateChecker{
		updateStrategy:  updateStrategy,
		updateAvailable: updateAvailable,
		commitAvailable: commitAvailable,
	}
}

func Test_shouldUpdateMod(t *testing.T) {
	// version constraints
	installed_tag := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: &versionmap.ResolvedVersionConstraint{
			DependencyVersion: modconfig.DependencyVersion{
				Tag: "mytag",
			},
		},
	}
	installed_branch := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: &versionmap.ResolvedVersionConstraint{
			DependencyVersion: modconfig.DependencyVersion{
				Branch: "mybranch",
			},
		},
	}
	installed_file := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: &versionmap.ResolvedVersionConstraint{
			DependencyVersion: modconfig.DependencyVersion{
				FilePath: "/path/to/mod",
			},
		},
	}
	installed_v1_0 := &versionmap.InstalledModVersion{
		ResolvedVersionConstraint: &versionmap.ResolvedVersionConstraint{
			DependencyVersion: modconfig.DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
		},
	}
	required_v1_1, _ := modconfig.NewModVersionConstraint("foo@1.1")
	required_any, _ := modconfig.NewModVersionConstraint("foo")
	required_file, _ := modconfig.NewModVersionConstraint("/path/to/mod")
	required_tag, _ := modconfig.NewModVersionConstraint("foo@mytag")
	required_branch, _ := modconfig.NewModVersionConstraint("foo#mybranch")

	type args struct {
		installedVersion        *versionmap.InstalledModVersion
		requiredModVersion      *modconfig.ModVersionConstraint
		commandTargettingParent bool
		updateChecker           updateChecker
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// not targetting mod
		{
			name: "not targetting mod (full)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: false,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not targetting mod (latest)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: false,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not targetting mod (development)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: false,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not targetting mod (minimal)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: false,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    false,
			wantErr: false,
		},
		// version constraint met
		{
			name: "version constraint met (full)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "version constraint met (latest)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "version constraint met (development)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "version constraint met (minimal)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    false,
			wantErr: false,
		},
		// version constraint not met
		{
			name: "version constraint not met (full)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version constraint not met (latest)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version constraint not met (development)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version constraint not met (minimal)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_v1_1,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    true,
			wantErr: false,
		},
		// new version available
		{
			name: "new version available (full)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", true, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "new version available (latest)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", true, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "new version available (development)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", true, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "new version available (minimal)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", true, false),
			},
			want:    false,
			wantErr: false,
		},
		// new commit available (version constraint)
		{
			name: "new commit available (version constraint)  (full)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, true),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "new commit available (version constraint)  (latest)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, true),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "new commit available (version constraint)  (development)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, true),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "new commit available (version constraint)  (minimal)",
			args: args{
				installedVersion:        installed_v1_0,
				requiredModVersion:      required_any,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, true),
			},
			want:    false,
			wantErr: false,
		},

		// installed (file) (always update)
		{
			name: "installed (file) (full)",
			args: args{
				installedVersion:        installed_file,
				requiredModVersion:      required_file,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (file) (latest)",
			args: args{
				installedVersion:        installed_file,
				requiredModVersion:      required_file,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (file) (development)",
			args: args{
				installedVersion:        installed_file,
				requiredModVersion:      required_file,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (file) (minimal)",
			args: args{
				installedVersion:        installed_file,
				requiredModVersion:      required_file,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    true,
			wantErr: false,
		},
		// installed (tag)
		{
			name: "installed (tag)  (full)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (tag)  (latest)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (tag)  (development)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (tag)  (minimal)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    false,
			wantErr: false,
		},
		// new commit available (tag)
		{
			name: "installed (tag)  (full)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, true),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (tag)  (latest)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, true),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (tag)  (development)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, true),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (tag)  (minimal)",
			args: args{
				installedVersion:        installed_tag,
				requiredModVersion:      required_tag,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, true),
			},
			want:    false,
			wantErr: false,
		},
		// installed (branch)
		{
			name: "installed (branch)  (full)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (branch)  (latest)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (branch)  (development)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, false),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "installed (branch)  (minimal)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, false),
			},
			want:    false,
			wantErr: false,
		},
		// new commit available (branch)
		{
			name: "installed (branch)  (full)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("full", false, true),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (branch)  (latest)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("latest", false, true),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (branch)  (development)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("development", false, true),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "installed (branch)  (minimal)",
			args: args{
				installedVersion:        installed_branch,
				requiredModVersion:      required_branch,
				commandTargettingParent: true,
				updateChecker:           newMockUpdateChecker("minimal", false, true),
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shouldUpdateMod(tt.args.installedVersion, tt.args.requiredModVersion, tt.args.commandTargettingParent, tt.args.updateChecker)
			if (err != nil) != tt.wantErr {
				t.Errorf("shouldUpdateMod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("shouldUpdateMod() got = %v, want %v", got, tt.want)
			}
		})
	}
}
