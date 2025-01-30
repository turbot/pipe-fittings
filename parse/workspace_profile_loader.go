package parse

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/sperr"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
)

func DefaultWorkspaceSampleFileName() string {
	return fmt.Sprintf("workspaces%s.sample", app_specific.ConfigExtension)
}

type WorkspaceProfileLoader[T workspace_profile.WorkspaceProfile] struct {
	workspaceProfiles map[string]T
	// list of config locations, in order or DECREASING precedence
	workspaceProfilePaths []string
	DefaultProfile        T
	ConfiguredProfile     T
	ParseOpts             []ParseHclOpt
}

func NewWorkspaceProfileLoader[T workspace_profile.WorkspaceProfile](workspaceProfilePaths ...string) (*WorkspaceProfileLoader[T], error) {
	loader := &WorkspaceProfileLoader[T]{
		workspaceProfilePaths: workspaceProfilePaths,
	}

	// must be at config one location
	if len(workspaceProfilePaths) == 0 {
		return nil, fmt.Errorf("no workspace profile locations specified")
	}

	// if a config paths location was NOT passed, write the workspaces.spc.sample file to the lowest precedence location
	// (assumed to be the global config folder)
	if !viper.IsSet(constants.ArgConfigPath) {
		if err := loader.ensureDefaultWorkspaceFile(workspaceProfilePaths); err != nil {
			return nil,
				sperr.WrapWithMessage(
					err,
					"could not create sample workspace",
				)
		}
	}

	return loader, nil
}

func (l *WorkspaceProfileLoader[T]) ensureDefaultWorkspaceFile(workspaceProfilePaths []string) error {
	globalConfigPath := workspaceProfilePaths[len(workspaceProfilePaths)-1]
	var empty T

	var sampleContent string
	switch any(empty).(type) {
	case *workspace_profile.SteampipeWorkspaceProfile:
		sampleContent = constants.DefaultSteampipeWorkspaceContent
	case *workspace_profile.FlowpipeWorkspaceProfile:
		sampleContent = constants.DefaultFlowpipeWorkspaceContent
	case *workspace_profile.PowerpipeWorkspaceProfile:
		sampleContent = constants.DefaultPowerpipeWorkspaceContent
	case *workspace_profile.TailpipeWorkspaceProfile:
		sampleContent = constants.DefaultTailpipeWorkspaceContent
	}
	// always write the workspaces sample file; i.e. workspaces.spc.sample
	err := os.MkdirAll(globalConfigPath, 0755)
	if err != nil {
		return err
	}

	defaultWorkspaceSampleFile := filepath.Join(globalConfigPath, DefaultWorkspaceSampleFileName())
	//nolint: gosec // this file is safe to be read by all users
	err = os.WriteFile(defaultWorkspaceSampleFile, []byte(sampleContent), 0755)
	if err != nil {
		return err
	}
	return nil
}

func (l *WorkspaceProfileLoader[T]) GetActiveWorkspaceProfile() T {
	if !l.ConfiguredProfile.IsNil() {
		return l.ConfiguredProfile
	}
	return l.DefaultProfile
}

func (l *WorkspaceProfileLoader[T]) get(name string) (T, bool) {
	var emptyProfile T

	if workspaceProfile, ok := l.workspaceProfiles[name]; ok {
		return workspaceProfile, true
	}

	if implicitWorkspace := l.getImplicitWorkspace(name); !implicitWorkspace.IsNil() {
		return implicitWorkspace, true
	}

	return emptyProfile, false // fmt.Errorf("workspace profile %s does not exist", name)
}

func (l *WorkspaceProfileLoader[T]) Load() error {
	// load workspaces from all locations
	var workspacesPrecedenceList = make([]map[string]T, 0, len(l.workspaceProfilePaths))

	// load from the config paths in reverse order (i.e. lowest precedence first)
	for _, configPath := range l.workspaceProfilePaths {
		// load all workspaces in the global config location
		// NOTE: pass parse opts if we have any
		workspaces, err := LoadWorkspaceProfiles[T](configPath, l.ParseOpts...)
		if err != nil {
			return err
		}

		// add to workspacesPrecedenceList
		if len(workspaces) > 0 {
			workspacesPrecedenceList = append(workspacesPrecedenceList, workspaces)
		}
	}

	// determine the default workspace
	if err := l.setDefault(workspacesPrecedenceList); err != nil {
		return err
	}

	l.setWorkspaces(workspacesPrecedenceList)

	// try to set the configured workspace
	if viper.IsSet(constants.ArgWorkspaceProfile) {
		name := viper.GetString(constants.ArgWorkspaceProfile)
		configuredProfile, ok := l.get(name)
		if !ok {
			// could not find configured profile
			return fmt.Errorf("workspace '%s' not found in config path %s", name, strings.Join(l.workspaceProfilePaths, ", "))
		}
		l.ConfiguredProfile = configuredProfile
	}

	return nil
}

func (l *WorkspaceProfileLoader[T]) setDefault(workspacesPrecedenceList []map[string]T) error {
	// workspacesPrecedenceList is in order of decreasing precedence

	// create an empty default
	var diags hcl.Diagnostics
	defaultWorkspace, diags := workspace_profile.NewDefaultWorkspaceProfile[T]()
	if diags.HasErrors() {
		return error_helpers.HclDiagsToError("failed to create default workspace", diags)
	}

	// now travers the list of paths in reverse order (i.e. order if INCREASING precedence)
	for i := len(workspacesPrecedenceList) - 1; i >= 0; i-- {
		workspaces := workspacesPrecedenceList[i]
		// if there is a 'default' workspace defined in localWorkspaces, use it as the default
		if localDefault, ok := workspaces["default"]; ok {
			defaultWorkspace = localDefault
			break
		}
	}

	l.DefaultProfile = defaultWorkspace
	return nil
}

/*
Named workspaces follow normal standards for hcl identities, thus they cannot contain the slash (/) character.

If you pass a value to --workspace or STEAMPIPE_WORKSPACE in the form of {identity_handle}/{workspace_handle},
it will be interpreted as an implicit workspace.

Implicit workspaces, as the name suggests, do not need to be specified in the workspaces.spc file.

Instead they will be assumed to refer to a Turbot Pipes workspace,
which will be used as both the database and snapshot location.

Essentially, --workspace acme/dev is equivalent to:

	workspace "acme/dev" {
	  database = "acme/dev"
	  snapshot_location  = "acme/dev"
	}
*/
func (l *WorkspaceProfileLoader[T]) getImplicitWorkspace(name string) T {
	var empty T
	if steampipeconfig.IsPipesWorkspaceIdentifier(name) {
		switch any(empty).(type) {
		case *workspace_profile.PowerpipeWorkspaceProfile:
			slog.Debug("getImplicitWorkspace - creating implicit workspace", "name", name)
			var res workspace_profile.WorkspaceProfile = &workspace_profile.PowerpipeWorkspaceProfile{
				SnapshotLocation: &name,
				// NOTE: we no longer set the deprecated Database property - instead set the CloudWorkspace property
				CloudWorkspace: &name,
			}
			return res.(T)
		case *workspace_profile.SteampipeWorkspaceProfile:
			slog.Debug("getImplicitWorkspace - creating implicit workspace", "name", name)
			var res workspace_profile.WorkspaceProfile = &workspace_profile.SteampipeWorkspaceProfile{
				SnapshotLocation:  &name,
				WorkspaceDatabase: &name,
			}
			return res.(T)
		}
	}
	return empty
}

func (l *WorkspaceProfileLoader[T]) setWorkspaces(workspacesPrecedenceList []map[string]T) {
	l.workspaceProfiles = make(map[string]T)

	// workspacesPrecedenceList is in order of decreasing precedence
	// iterate _back_ through the list
	for i := len(workspacesPrecedenceList) - 1; i >= 0; i-- {
		workspaces := workspacesPrecedenceList[i]
		for k, v := range workspaces {
			l.workspaceProfiles[k] = v
		}
	}
}
