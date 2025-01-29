package workspace_profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type TailpipeWorkspaceProfile struct {
	ProfileName string `hcl:"name,label" cty:"name"`

	Local            *string `hcl:"local" cty:"local"`
	Remote           *string `hcl:"remote" cty:"remote"`
	RemoteConnection *string `hcl:"remote_connection" cty:"remote_connection"`

	// general options
	UpdateCheck *string `hcl:"update_check" cty:"update_check"`
	LogLevel    *string `hcl:"log_level" cty:"log_level"`
	MemoryMaxMb *int    `hcl:"memory_max_mb" cty:"memory_max_mb"`

	// terminal options
	Timing *bool `hcl:"timing" cty:"timing"`

	Base *TailpipeWorkspaceProfile `hcl:"base"`

	DeclRange hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *TailpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Powerpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

func (p *TailpipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *TailpipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *TailpipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(p)
}

func (p *TailpipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *TailpipeWorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *TailpipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map
	res.SetStringItem(p.Local, constants.ArgLocal)
	res.SetStringItem(p.Remote, constants.ArgRemote)
	res.SetStringItem(p.RemoteConnection, constants.ArgRemoteConnection)
	res.SetStringItem(p.UpdateCheck, constants.ArgUpdateCheck)
	res.SetStringItem(p.LogLevel, constants.ArgLogLevel)
	res.SetIntItem(p.MemoryMaxMb, constants.ArgMemoryMaxMb)
	res.SetBoolItem(p.Timing, constants.ArgTiming)

	return res
}

func (p *TailpipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// TODO this is (currently) required by interface
func (p *TailpipeWorkspaceProfile) GetInstallDir() *string {
	return nil
}

func (p *TailpipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

func (p *TailpipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Powerpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

// EnsureWorkspaceDirs creates all necessary workspace directories
func (p *TailpipeWorkspaceProfile) EnsureWorkspaceDirs() error {
	workspaceDirs := []string{p.GetDataDir(), p.GetCollectionDir()}

	// create if necessary
	for _, dir := range workspaceDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *TailpipeWorkspaceProfile) GetDataDir() string {
	var dataDir string
	if p.Local != nil {
		dataDir = *p.Local
	} else {
		dataDir = filepath.Join(filepaths.GetDataDir(), p.ProfileName)
	}
	return dataDir
}

// GetCollectionDir returns the path to the collection data directory
// - this is located  in ~/.turbot/internal/collection/<profile_name>
// this will contain the collection temp dir (which should only exist during collection) and the collection state
func (p *TailpipeWorkspaceProfile) GetCollectionDir() string {
	return filepath.Join(filepaths.GetInternalDir(), "collection", p.ProfileName)
}
