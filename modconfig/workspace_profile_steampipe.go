package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

type SteampipeWorkspaceProfile struct {
	ProfileName string `hcl:"name,label" cty:"name"`

	//  dashboard / api server options
	Host   *string `hcl:"string" cty:"string"`
	Port   *int    `hcl:"port" cty:"port"`
	Listen *string `hcl:"listen" cty:"port"`

	// general options
	UpdateCheck *string `hcl:"update_check" cty:"update_check"`
	Telemetry   *string `hcl:"telemetry" cty:"telemetry"`
	LogLevel    *string `hcl:"log_level" cty:"log_level"`
	MemoryMaxMb *int    `hcl:"memory_max_mb" cty:"memory_max_mb"`

	// pipes integration options
	CloudHost        *string `hcl:"cloud_host,optional" cty:"cloud_host"`
	CloudToken       *string `hcl:"cloud_token,optional" cty:"cloud_token"`
	SnapshotLocation *string `hcl:"snapshot_location,optional" cty:"snapshot_location"`

	ModLocation *string `hcl:"mod_location,optional" cty:"mod_location"`

	Watch    *bool `hcl:"watch" cty:"watch"`
	Input    *bool `hcl:"input" cty:"input"`
	Progress *bool `hcl:"progress" cty:"progress"`

	// "default" db settings
	WorkspaceDatabase *string `hcl:"workspace_database" cty:"workspace_database"`
	QueryTimeout      *int    `hcl:"query_timeout,optional" cty:"query_timeout"`
	MaxParallel       *int    `hcl:"max_parallel" cty:"max-parallel"`

	// (postgres-specific) search path settings
	SearchPath       *string `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix *string `hcl:"search_path_prefix" cty:"search_path_prefix"`

	// terminal options
	Header    *bool   `hcl:"header" cty:"header"`
	Output    *string `hcl:"output"`
	Separator *string `hcl:"separator"`
	Timing    *bool   `hcl:"timing"`

	Base      *SteampipeWorkspaceProfile `hcl:"base"`
	DeclRange hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *SteampipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "options blocks are supported",
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

	if p.Host == nil {
		p.Host = p.Base.Host
	}
	if p.Port == nil {
		p.Port = p.Base.Port
	}
	if p.Listen == nil {
		p.Listen = p.Base.Listen
	}

	if p.UpdateCheck == nil {
		p.UpdateCheck = p.Base.UpdateCheck
	}
	if p.Telemetry == nil {
		p.Telemetry = p.Base.Telemetry
	}
	if p.LogLevel == nil {
		p.LogLevel = p.Base.LogLevel
	}
	if p.MemoryMaxMb == nil {
		p.MemoryMaxMb = p.Base.MemoryMaxMb
	}
	if p.CloudHost == nil {
		p.CloudHost = p.Base.CloudHost
	}
	if p.CloudToken == nil {
		p.CloudToken = p.Base.CloudToken
	}
	if p.SnapshotLocation == nil {
		p.SnapshotLocation = p.Base.SnapshotLocation
	}
	if p.ModLocation == nil {
		p.ModLocation = p.Base.ModLocation
	}

	if p.Watch == nil {
		p.Watch = p.Base.Watch
	}
	if p.Input == nil {
		p.Input = p.Base.Input
	}
	if p.Progress == nil {
		p.Progress = p.Base.Progress
	}

	if p.WorkspaceDatabase == nil {
		p.WorkspaceDatabase = p.Base.WorkspaceDatabase
	}
	if p.QueryTimeout == nil {
		p.QueryTimeout = p.Base.QueryTimeout
	}
	if p.MaxParallel == nil {
		p.MaxParallel = p.Base.MaxParallel
	}

	if p.SearchPath == nil {
		p.SearchPath = p.Base.SearchPath
	}
	if p.SearchPathPrefix == nil {
		p.SearchPathPrefix = p.Base.SearchPathPrefix
	}

	if p.Header == nil {
		p.Header = p.Base.Header
	}
	if p.Output == nil {
		p.Output = p.Base.Output
	}
	if p.Separator == nil {
		p.Separator = p.Base.Separator
	}

	if p.Timing == nil {
		p.Timing = p.Base.Timing
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *SteampipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map
	res.SetStringItem(p.Host, constants.ArgHost)
	res.SetIntItem(p.Port, constants.ArgPort)
	res.SetStringItem(p.Listen, constants.ArgListen)

	res.SetStringItem(p.UpdateCheck, constants.ArgUpdateCheck)
	res.SetStringItem(p.Telemetry, constants.ArgTelemetry)
	res.SetStringItem(p.LogLevel, constants.ArgLogLevel)
	res.SetIntItem(p.MemoryMaxMb, constants.ArgMemoryMaxMb)

	res.SetStringItem(p.CloudHost, constants.ArgCloudHost)
	res.SetStringItem(p.CloudToken, constants.ArgCloudToken)
	res.SetStringItem(p.SnapshotLocation, constants.ArgSnapshotLocation)

	res.SetStringItem(p.ModLocation, constants.ArgModLocation)

	res.SetBoolItem(p.Watch, constants.ArgWatch)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Progress, constants.ArgProgress)

	res.SetStringItem(p.WorkspaceDatabase, constants.ArgWorkspaceDatabase)
	res.SetIntItem(p.QueryTimeout, constants.ArgDatabaseQueryTimeout)
	res.SetIntItem(p.MaxParallel, constants.ArgMaxParallel)

	res.SetStringSliceItem(searchPathFromString(p.SearchPath, ","), constants.ArgSearchPath)
	res.SetStringSliceItem(searchPathFromString(p.SearchPathPrefix, ","), constants.ArgSearchPathPrefix)

	res.SetBoolItem(p.Header, constants.ArgHeader)
	res.SetStringItem(p.Output, constants.ArgOutput)
	res.SetStringItem(p.Separator, constants.ArgSeparator)
	res.SetBoolItem(p.Timing, constants.ArgTiming)

	return res
}

func (p *SteampipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

func (p *SteampipeWorkspaceProfile) GetInstallDir() *string {
	return nil
}

func (p *SteampipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

func (p *SteampipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Type),
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

// searchPathFromString checks that `str` is `nil` and returns a string slice with `str`
// separated with `separator`
// If `str` is `nil`, this returns a `nil`
func searchPathFromString(str *string, separator string) []string {
	if str == nil {
		return nil
	}
	// convert comma separated list to array
	searchPath := strings.Split(*str, separator)
	// strip whitespace
	for i, s := range searchPath {
		searchPath[i] = strings.TrimSpace(s)
	}
	return searchPath
}
