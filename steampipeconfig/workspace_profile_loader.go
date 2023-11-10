package steampipeconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"os"
	"path/filepath"
)

var defaultWorkspaceSampleFileName = "workspaces.spc.sample"

type WorkspaceProfileLoader[T modconfig.WorkspaceProfile] struct {
	workspaceProfiles          map[string]T
	globalWorkspaceProfilePath string
	localWorkspaceProfilePath  string
	DefaultProfile             T
	ConfiguredProfile          T
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

func NewWorkspaceProfileLoader[T modconfig.WorkspaceProfile](globalWorkspaceProfilePath, localWorkspaceProfilePath string) (*WorkspaceProfileLoader[T], error) {
	// write the workspaces.spc.sample file
	if err := ensureDefaultWorkspaceFile(globalWorkspaceProfilePath); err != nil {
		return nil,
			sperr.WrapWithMessage(
				err,
				"could not create sample workspace",
			)
	}
	loader := &WorkspaceProfileLoader[T]{
		globalWorkspaceProfilePath: globalWorkspaceProfilePath,
		localWorkspaceProfilePath:  localWorkspaceProfilePath,
	}

	// do the load
	err := loader.load()
	if err != nil {
		return nil, err
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
	// load all workspaces in the global config location
	globalWorkspaces, err := parse.LoadWorkspaceProfiles[T](l.globalWorkspaceProfilePath)
	if err != nil {
		return err
	}
	// load all workspaces in the mod location
	localWorkspaces, err := parse.LoadWorkspaceProfiles[T](l.localWorkspaceProfilePath)
	if err != nil {
		return err
	}

	// determine the default workspace
	if err := l.setDefault(localWorkspaces, globalWorkspaces); err != nil {
		return err
	}

	l.setWorkspaces(localWorkspaces, globalWorkspaces)

	// try to set the configured workspace
	if viper.IsSet(constants.ArgWorkspaceProfile) {
		configuredProfile, ok := l.get(viper.GetString(constants.ArgWorkspaceProfile))
		if !ok {
			// could not find configured profile
			return err
		}
		l.ConfiguredProfile = configuredProfile
	}

	return nil
}

func (l *WorkspaceProfileLoader[T]) setDefault(localWorkspaces map[string]T, globalWorkspaces map[string]T) error {
	// if there is a 'local' workspace defined in localWorkspaces, use it as the default
	defaultProfile, ok := localWorkspaces["local"]
	if !ok {
		// no local profile - look for a global default
		defaultProfile, ok = globalWorkspaces["default"]
		if !ok {
			var diags hcl.Diagnostics
			defaultProfile, diags = modconfig.NewDefaultWorkspaceProfile[T]()
			if diags.HasErrors() {
				return plugin.DiagsToError("failed to create default workspace", diags)
			}
		}
	}
	l.DefaultProfile = defaultProfile
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
	//	log.Printf("[TRACE] getImplicitWorkspace - %s is implicit workspace: SnapshotLocation=%s, WorkspaceDatabase=%s", name, name, name)
	//	return &modconfig.SteampipeWorkspaceProfile{
	//		SnapshotLocation:  utils.ToStringPointer(name),
	//		WorkspaceDatabase: utils.ToStringPointer(name),
	//	}
	//}
	var w T
	return w
}

func (l *WorkspaceProfileLoader[T]) setWorkspaces(globalWorkspaces, localWorkspaces map[string]T) {
	l.workspaceProfiles = make(map[string]T)
	// assign global workspaces
	for k, v := range globalWorkspaces {
		l.workspaceProfiles[k] = v
	}
	// assign local workspaces with higher precedence (i.e. these overwrite global workspaces with same name)
	for k, v := range localWorkspaces {
		l.workspaceProfiles[k] = v
	}
}
