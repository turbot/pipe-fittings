package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

type FlowpipeWorkspaceProfile struct {
	ProfileName string  `hcl:"name,label" cty:"name"`
	Output      *string `hcl:"output" cty:"max-output"`

	// TODO this is in general options
	MaxParallel *int `hcl:"max_parallel" cty:"max-parallel"`

	Watch    *bool   `hcl:"watch" cty:"watch"`
	Input    *bool   `hcl:"input" cty:"input"`
	Progress *bool   `hcl:"progress" cty:"progress"`
	Host     *string `hcl:"host" cty:"host"`

	Insecure *bool                     `hcl:"insecure" cty:"insecure"`
	Base     *FlowpipeWorkspaceProfile `hcl:"base"`

	// options
	ServerOptions  *options.Server  `cty:"server-options"`
	GeneralOptions *options.General `cty:"general-options"`

	DeclRange hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *FlowpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Server:
		if p.ServerOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.ServerOptions = o
	case *options.General:
		if p.GeneralOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.GeneralOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Workspace", reflect.TypeOf(o).Name()),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
	}
	return diags
}

func (p *FlowpipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *FlowpipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *FlowpipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return GetCtyValue(p)
}

func (p *FlowpipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *FlowpipeWorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}

	if p.Output == nil {
		p.Output = p.Base.Output
	}

	if p.Watch == nil {
		p.Watch = p.Base.Watch
	}

	if p.MaxParallel == nil {
		p.MaxParallel = p.Base.MaxParallel
	}

	if p.Input == nil {
		p.Input = p.Base.Input
	}

	if p.Progress == nil {
		p.Progress = p.Base.Progress
	}

	if p.Host == nil {
		p.Host = p.Base.Host
	}

	if p.Insecure == nil {
		p.Insecure = p.Base.Insecure
	}

	// nested inheritance strategy:
	//
	// if my nested struct is a nil
	//		-> use the base struct
	//
	// if I am not nil (and base is not nil)
	//		-> only inherit the properties which are nil in me and not in base
	//
	if p.ServerOptions == nil {
		p.ServerOptions = p.Base.ServerOptions
	} else {
		p.ServerOptions.SetBaseProperties(p.Base.ServerOptions)
	}
	if p.GeneralOptions == nil {
		p.GeneralOptions = p.Base.GeneralOptions
	} else {
		p.GeneralOptions.SetBaseProperties(p.Base.GeneralOptions)
	}

}

// ConfigMap creates a config map containing all options to pass to viper
func (p *FlowpipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map
	res.SetIntItem(p.MaxParallel, constants.ArgMaxParallel)
	res.SetStringItem(p.Output, constants.ArgOutput)
	res.SetBoolItem(p.Watch, constants.ArgWatch)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Progress, constants.ArgProgress)
	res.SetStringItem(p.Host, constants.ArgHost)
	res.SetBoolItem(p.Insecure, constants.ArgInsecure)

	if p.ServerOptions != nil {
		res.PopulateConfigMapForOptions(p.ServerOptions)
	}

	return res
}

func (p *FlowpipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// TODO this is (currently) required by interface
func (p *FlowpipeWorkspaceProfile) GetInstallDir() *string {
	return nil
}

func (p *FlowpipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

func (p *FlowpipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	switch block.Labels[0] {
	case options.ServerBlock:
		return new(options.Server), nil
	case options.GeneralBlock:
		return new(options.General), nil
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Type),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return nil, diags
	}
}
