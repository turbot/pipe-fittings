package steampipeconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

func defaultWorkspaceSampleFileName() string {
	return fmt.Sprintf("workspaces%s.sample", app_specific.ConfigExtension)
}

type WorkspaceProfileLoader[T modconfig.WorkspaceProfile] struct {
	workspaceProfiles map[string]T
	// list of config locations, in order or DECREASING precedence
	workspaceProfilePaths []string
	DefaultProfile        T
	ConfiguredProfile     T
}

func NewWorkspaceProfileLoader[T modconfig.WorkspaceProfile](workspaceProfilePaths ...string) (*WorkspaceProfileLoader[T], error) {
	loader := &WorkspaceProfileLoader[T]{
		workspaceProfilePaths: workspaceProfilePaths,
	}

	// must be at config one location
	if len(workspaceProfilePaths) == 0 {
		return nil, fmt.Errorf("no workspace profile locations specified")
	}

	// write the workspaces.spc.sample file to the lowest precedence location (assumed to be the gloabl config folder)
	if err := loader.ensureDefaultWorkspaceFile(workspaceProfilePaths[0]); err != nil {
		return nil,
			sperr.WrapWithMessage(
				err,
				"could not create sample workspace",
			)
	}

	// do the load
	err := loader.load()
	if err != nil {
		return nil, err
	}

	return loader, nil
}

func (l *WorkspaceProfileLoader[T]) ensureDefaultWorkspaceFile(configFolder string) error {
	var empty T

	var sampleContent string
	switch any(empty).(type) {
	case *modconfig.FlowpipeWorkspaceProfile:
		sampleContent = constants.DefaultFlowpipeWorkspaceContent
	case *modconfig.SteampipeWorkspaceProfile:
		sampleContent = constants.DefaultSteampipeWorkspaceContent
	case *modconfig.PowerpipeWorkspaceProfile:
		sampleContent = constants.DefaultPowerpipeWorkspaceContent
	}
	// always write the workspaces sample file; i.e. workspaces.spc.sample
	err := os.MkdirAll(configFolder, 0755)
	if err != nil {
		return err
	}

	defaultWorkspaceSampleFile := filepath.Join(configFolder, defaultWorkspaceSampleFileName())
	//nolint: gosec // this file is safe to be read by all users
	err = os.WriteFile(defaultWorkspaceSampleFile, []byte(sampleContent), 0755)
	if err != nil {
		return err
	}
	return nil
}

func (l *WorkspaceProfileLoader[T]) GetActiveWorkspaceProfile() T {
	// TODO KAI nicer way to nil check
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

func (l *WorkspaceProfileLoader[T]) load() error {
	// load workspaces from all locations
	var workspacesPrecedenceList = make([]map[string]T, len(l.
		workspaceProfilePaths))

	// load from the config paths in reverse order (i.e. lowest precedence first)
	for i := len(l.workspaceProfilePaths) - 1; i >= 0; i-- {
		configPath := l.workspaceProfilePaths[i]
		// load all workspaces in the global config location
		workspaces, err := parse.LoadWorkspaceProfiles[T](configPath)
		if err != nil {
			return err
		}

		workspacesPrecedenceList[i] = workspaces
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
			return fmt.Errorf("the configured profile '%s' does not exist", name)
		}
		l.ConfiguredProfile = configuredProfile
	}

	return nil
}

func (l *WorkspaceProfileLoader[T]) setDefault(workspacesPrecedenceList []map[string]T) error {
	// the first element in the workspacesPrecedenceList is the global config location
	globalWorkspaces := workspacesPrecedenceList[0]

	// get the global default workspace
	// no local profile - look for a global default
	defaultWorkspace, ok := globalWorkspaces["default"]
	if !ok {
		var diags hcl.Diagnostics
		defaultWorkspace, diags = modconfig.NewDefaultWorkspaceProfile[T]()
		if diags.HasErrors() {
			return plugin.DiagsToError("failed to create default workspace", diags)
		}
	}

	if len(workspacesPrecedenceList) > 1 {
		for _, workspaces := range workspacesPrecedenceList[1:] {
			// if there is a 'local' workspace defined in localWorkspaces, use it as the default
			if localDefault, ok := workspaces["local"]; ok {
				defaultWorkspace = localDefault
				break
			}
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
	  workspace_database = "acme/dev"
	  snapshot_location  = "acme/dev"
	}
*/
func (l *WorkspaceProfileLoader[T]) getImplicitWorkspace(name string) T {
	// TODO KAI FIX ME <WORKSPACE>
	//if IsCloudWorkspaceIdentifier(name) {
	//	slog.Debug("getImplicitWorkspace - %s is implicit workspace: SnapshotLocation=%s, WorkspaceDatabase=%s", name, name, name)
	//	return &modconfig.SteampipeWorkspaceProfile{
	//		SnapshotLocation:  utils.ToStringPointer(name),
	//		WorkspaceDatabase: utils.ToStringPointer(name),
	//	}
	//}
	var w T
	return w
}

func (l *WorkspaceProfileLoader[T]) setWorkspaces(workspacesPrecedenceList []map[string]T) {
	l.workspaceProfiles = make(map[string]T)

	for _, workspaces := range workspacesPrecedenceList {
		for k, v := range workspaces {
			l.workspaceProfiles[k] = v
		}
	}
}
