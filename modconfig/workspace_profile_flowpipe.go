package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type FlowpipeWorkspaceProfile struct {
	ProfileName string                    `hcl:"name,label" cty:"name"`
	Base        *FlowpipeWorkspaceProfile `hcl:"base"`

	Host        *string `hcl:"host" cty:"host"`
	Input       *bool   `hcl:"input" cty:"input"`
	Insecure    *bool   `hcl:"insecure" cty:"insecure"`
	Listen      *string `hcl:"listen" cty:"port"`
	LogLevel    *string `hcl:"log_level" cty:"log_level"`
	MemoryMaxMb *int    `hcl:"memory_max_mb" cty:"memory_max_mb"`
	Output      *string `hcl:"output" cty:"output"`
	Port        *int    `hcl:"port" cty:"port"`
	Progress    *bool   `hcl:"progress" cty:"progress"`
	Telemetry   *string `hcl:"telemetry" cty:"telemetry"`
	UpdateCheck *string `hcl:"update_check" cty:"update_check"`
	Watch       *bool   `hcl:"watch" cty:"watch"`

	DeclRange hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *FlowpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Flowpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
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

	if p.Host == nil {
		p.Host = p.Base.Host
	}
	if p.Input == nil {
		p.Input = p.Base.Input
	}
	if p.Insecure == nil {
		p.Insecure = p.Base.Insecure
	}
	if p.Listen == nil && p.Base.Listen != nil {
		p.Listen = p.Base.Listen
	}
	if p.LogLevel == nil && p.Base.LogLevel != nil {
		p.LogLevel = p.Base.LogLevel
	}
	if p.MemoryMaxMb == nil && p.Base.MemoryMaxMb != nil {
		p.MemoryMaxMb = p.Base.MemoryMaxMb
	}
	if p.Output == nil {
		p.Output = p.Base.Output
	}
	if p.Port == nil && p.Base.Port != nil {
		p.Port = p.Base.Port
	}
	if p.Progress == nil {
		p.Progress = p.Base.Progress
	}
	if p.UpdateCheck == nil && p.Base.UpdateCheck != nil {
		p.UpdateCheck = p.Base.UpdateCheck
	}
	if p.Telemetry == nil && p.Base.Telemetry != nil {
		p.Telemetry = p.Base.Telemetry
	}
	if p.Watch == nil {
		p.Watch = p.Base.Watch
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *FlowpipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map
	res.SetStringItem(p.Host, constants.ArgHost)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Insecure, constants.ArgInsecure)
	res.SetStringItem(p.Listen, constants.ArgListen)
	res.SetStringItem(p.LogLevel, constants.ArgLogLevel)
	res.SetIntItem(p.MemoryMaxMb, constants.ArgMemoryMaxMb)
	res.SetStringItem(p.Output, constants.ArgOutput)
	res.SetIntItem(p.Port, constants.ArgPort)
	res.SetBoolItem(p.Progress, constants.ArgProgress)
	res.SetStringItem(p.Telemetry, constants.ArgTelemetry)
	res.SetStringItem(p.UpdateCheck, constants.ArgUpdateCheck)
	res.SetBoolItem(p.Watch, constants.ArgWatch)

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
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Flowpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}
