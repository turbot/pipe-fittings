package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type SteampipeWorkspaceProfile struct {
	ProfileName string                     `hcl:"name,label" cty:"name"`
	Base        *SteampipeWorkspaceProfile `hcl:"base"`

	DeclRange hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *SteampipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Steampipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

func (p *SteampipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *SteampipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *SteampipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return GetCtyValue(p)
}

func (p *SteampipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *SteampipeWorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}

}

// ConfigMap creates a config map containing all options to pass to viper
func (p *SteampipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}

	return res
}

func (p *SteampipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// TODO this is (currently) required by interface
func (p *SteampipeWorkspaceProfile) GetInstallDir() *string {
	return nil
}

func (p *SteampipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

func (p *SteampipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Steampipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}
