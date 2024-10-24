package workspace_profile

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type FlowpipeWorkspaceProfile struct {
	ProfileName string                    `hcl:"name,label" cty:"name"`
	Base        *FlowpipeWorkspaceProfile `hcl:"base"`

	Host                    *string `hcl:"host" cty:"host"`
	Input                   *bool   `hcl:"input" cty:"input"`
	Insecure                *bool   `hcl:"insecure" cty:"insecure"`
	Listen                  *string `hcl:"listen" cty:"port"`
	LogLevel                *string `hcl:"log_level" cty:"log_level"`
	MemoryMaxMb             *int    `hcl:"memory_max_mb" cty:"memory_max_mb"`
	Output                  *string `hcl:"output" cty:"output"`
	Port                    *int    `hcl:"port" cty:"port"`
	Progress                *bool   `hcl:"progress" cty:"progress"`
	Telemetry               *string `hcl:"telemetry" cty:"telemetry"`
	UpdateCheck             *string `hcl:"update_check" cty:"update_check"`
	Watch                   *bool   `hcl:"watch" cty:"watch"`
	MaxConcurrencyHttp      *int    `hcl:"max_concurrency_http" cty:"max_concurrency_http"`
	MaxConcurrencyContainer *int    `hcl:"max_concurrency_container" cty:"max_concurrency_container"`
	MaxConcurrencyFunction  *int    `hcl:"max_concurrency_function" cty:"max_concurrency_function"`
	MaxConcurrencyQuery     *int    `hcl:"max_concurrency_query" cty:"max_concurrency_query"`
	ProcessRetention        *int    `hcl:"process_retention" cty:"process_retention"`
	BaseUrl                 *string `hcl:"base_url" cty:"base_url"`

	DeclRange hcl.Range
}

func (p *FlowpipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *FlowpipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *FlowpipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(p)
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
	if p.MaxConcurrencyContainer == nil {
		p.MaxConcurrencyContainer = p.Base.MaxConcurrencyContainer
	}
	if p.MaxConcurrencyFunction == nil {
		p.MaxConcurrencyFunction = p.Base.MaxConcurrencyFunction
	}
	if p.MaxConcurrencyHttp == nil {
		p.MaxConcurrencyHttp = p.Base.MaxConcurrencyHttp
	}
	if p.MaxConcurrencyQuery == nil {
		p.MaxConcurrencyQuery = p.Base.MaxConcurrencyQuery
	}
	if p.ProcessRetention == nil {
		p.ProcessRetention = p.Base.ProcessRetention
	}
	if p.BaseUrl == nil {
		p.BaseUrl = p.Base.BaseUrl
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
	res.SetIntItem(p.MaxConcurrencyContainer, constants.ArgMaxConcurrencyContainer)
	res.SetIntItem(p.MaxConcurrencyFunction, constants.ArgMaxConcurrencyFunction)
	res.SetIntItem(p.MaxConcurrencyHttp, constants.ArgMaxConcurrencyHttp)
	res.SetIntItem(p.MaxConcurrencyQuery, constants.ArgMaxConcurrencyQuery)
	res.SetIntItem(p.ProcessRetention, constants.ArgProcessRetention)
	res.SetStringItem(p.BaseUrl, constants.ArgBaseUrl)

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

// SetOptions sets the options on the Workspace
// FlowpipeWorkspaceProfile does not support options
func (p *FlowpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Flowpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

func (p *FlowpipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Flowpipe workspaces do not support options",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}
