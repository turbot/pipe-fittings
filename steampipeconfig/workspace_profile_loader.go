package steampipeconfig

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"os"
	"path/filepath"
)

var defaultWorkspaceSampleFileName = "workspaces.spc.sample"

type WorkspaceProfileLoader[T modconfig.WorkspaceProfile] struct {
	workspaceProfiles    map[string]T
	workspaceProfilePath string
	DefaultProfile       T
	ConfiguredProfile    T
}

func ensureDefaultWorkspaceFile(configFolder string) error {
	// always write the workspaces.spc.sample file
	err := os.MkdirAll(configFolder, 0755)
	if err != nil {
		return err
	}
	defaultWorkspaceSampleFile := filepath.Join(configFolder, defaultWorkspaceSampleFileName)
	//nolint: gosec // this file is safe to be read by all users
	err = os.WriteFile(defaultWorkspaceSampleFile, []byte(constants.DefaultWorkspaceContent), 0755)
	if err != nil {
		return err
	}
	return nil
}

func NewWorkspaceProfileLoader[T modconfig.WorkspaceProfile](workspaceProfilePath string) (*WorkspaceProfileLoader[T], error) {
	// write the workspaces.spc.sample file
	if err := ensureDefaultWorkspaceFile(workspaceProfilePath); err != nil {
		return nil,
			sperr.WrapWithMessage(
				err,
				"could not create sample workspace",
			)
	}
	loader := &WorkspaceProfileLoader[T]{workspaceProfilePath: workspaceProfilePath}

	// do the load
	workspaceProfiles, err := loader.load()
	if err != nil {
		return nil, err
	}
	loader.workspaceProfiles = workspaceProfiles

	defaultProfile, err := loader.get("default")
	if err != nil {
		// there must always be a default - this should have been added by parse.LoadWorkspaceProfiles
		return nil, err
	}
	loader.DefaultProfile = defaultProfile

	if viper.IsSet(constants.ArgWorkspaceProfile) {
		configuredProfile, err := loader.get(viper.GetString(constants.ArgWorkspaceProfile))
		if err != nil {
			// could not find configured profile
			return nil, err
		}
		loader.ConfiguredProfile = configuredProfile
	}

	return loader, nil
}

func (l *WorkspaceProfileLoader[T]) GetActiveWorkspaceProfile() T {
	// TODO KAI nicer way to nil check
	if !l.ConfiguredProfile.IsNil() {
		return l.ConfiguredProfile
	}
	return l.DefaultProfile
}

func (l *WorkspaceProfileLoader[T]) get(name string) (T, error) {
	var emptyProfile T
	if workspaceProfile, ok := l.workspaceProfiles[name]; ok {
		return workspaceProfile, nil
	}

	if implicitWorkspace := l.getImplicitWorkspace(name); !implicitWorkspace.IsNil() {
		return implicitWorkspace, nil
	}

	return emptyProfile, fmt.Errorf("workspace profile %s does not exist", name)
}

func (l *WorkspaceProfileLoader[T]) load() (map[string]T, error) {
	// get all the config files in the directory
	return parse.LoadWorkspaceProfiles[T](l.workspaceProfilePath)
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
	//	log.Printf("[TRACE] getImplicitWorkspace - %s is implicit workspace: SnapshotLocation=%s, WorkspaceDatabase=%s", name, name, name)
	//	return &modconfig.SteampipeWorkspaceProfile{
	//		SnapshotLocation:  utils.ToStringPointer(name),
	//		WorkspaceDatabase: utils.ToStringPointer(name),
	//	}
	//}
	var w T
	return w
}
