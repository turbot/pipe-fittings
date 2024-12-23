package workspace_profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/turbot/pipe-fittings/pipes"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/zclconf/go-cty/cty"
)

type PowerpipeWorkspaceProfile struct {
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

	// execution timeouts
	BenchmarkTimeout *int `hcl:"benchmark_timeout" cty:"benchmark_timeout"`
	DashboardTimeout *int `hcl:"dashboard_timeout" cty:"dashboard_timeout"`

	// pipes integration options
	PipesHost        *string `hcl:"pipes_host,optional" cty:"pipes_host"`
	PipesToken       *string `hcl:"pipes_token,optional" cty:"pipes_token"`
	SnapshotLocation *string `hcl:"snapshot_location,optional" cty:"snapshot_location"`

	ModLocation *string `hcl:"mod_location,optional" cty:"mod_location"`

	Watch    *bool `hcl:"watch" cty:"watch"`
	Input    *bool `hcl:"input" cty:"input"`
	Progress *bool `hcl:"progress" cty:"progress"`

	// "default" db settings
	Database     *string `hcl:"database" cty:"database"` // deprecated
	QueryTimeout *int    `hcl:"query_timeout,optional" cty:"query_timeout"`
	MaxParallel  *int    `hcl:"max_parallel" cty:"max-parallel"`

	// (postgres-specific) search path settings
	SearchPath       *string `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix *string `hcl:"search_path_prefix" cty:"search_path_prefix"`

	// terminal options
	Header    *bool   `hcl:"header" cty:"header"`
	Output    *string `hcl:"output"`
	Separator *string `hcl:"separator"`
	Timing    *bool   `hcl:"timing"`

	// set if this is an implicit profile for a cloud workspace
	CloudWorkspace *string

	Base      *PowerpipeWorkspaceProfile `hcl:"base"`
	DeclRange hcl.Range
}

func (p *PowerpipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *PowerpipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *PowerpipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(p)
}

func (p *PowerpipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *PowerpipeWorkspaceProfile) setBaseProperties() {
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
	if p.BenchmarkTimeout == nil {
		p.BenchmarkTimeout = p.Base.BenchmarkTimeout
	}
	if p.DashboardTimeout == nil {
		p.DashboardTimeout = p.Base.DashboardTimeout
	}
	if p.PipesHost == nil {
		p.PipesHost = p.Base.PipesHost
	}
	if p.PipesToken == nil {
		p.PipesToken = p.Base.PipesToken
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

	if p.Database == nil {
		p.Database = p.Base.Database
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
func (p *PowerpipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map
	res.SetStringItem(p.Host, constants.ArgHost)
	res.SetIntItem(p.Port, constants.ArgPort)
	res.SetStringItem(p.Listen, constants.ArgListen)

	res.SetStringItem(p.UpdateCheck, constants.ArgUpdateCheck)
	res.SetStringItem(p.Telemetry, constants.ArgTelemetry)
	res.SetStringItem(p.LogLevel, constants.ArgLogLevel)
	res.SetIntItem(p.MemoryMaxMb, constants.ArgMemoryMaxMb)

	res.SetIntItem(p.BenchmarkTimeout, constants.ArgBenchmarkTimeout)
	res.SetIntItem(p.DashboardTimeout, constants.ArgDashboardTimeout)

	res.SetStringItem(p.PipesHost, constants.ArgPipesHost)
	res.SetStringItem(p.PipesToken, constants.ArgPipesToken)
	res.SetStringItem(p.SnapshotLocation, constants.ArgSnapshotLocation)

	res.SetStringItem(p.ModLocation, constants.ArgModLocation)

	res.SetBoolItem(p.Watch, constants.ArgWatch)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Progress, constants.ArgProgress)

	res.SetStringItem(p.Database, constants.ArgDatabase)
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

func (p *PowerpipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

func (p *PowerpipeWorkspaceProfile) GetInstallDir() *string {
	return nil
}

func (p *PowerpipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

func (p *PowerpipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Type),
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

// SetOptions sets the options on the connection
// PowerpipeWorkspaceProfile does not support options
func (p *PowerpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "options blocks are supported",
		Subject:  hclhelpers.BlockRangePointer(block),
	}}
}

func (p *PowerpipeWorkspaceProfile) IsCloudWorkspace() bool {
	return p.CloudWorkspace != nil
}

// GetPipesMetadata returns the cloud metadata for the cloud workspace
// note: call IsCloudWorkspace before calling this to ensure it is a cloud workspace
func (p *PowerpipeWorkspaceProfile) GetPipesMetadata() (*steampipeconfig.PipesMetadata, error_helpers.ErrorAndWarnings) {
	if !p.IsCloudWorkspace() {
		return nil, error_helpers.NewErrorsAndWarning(fmt.Errorf("workspace profile is not a cloud workspace"))
	}

	// verify the cloud token was provided
	cloudToken := viper.GetString(constants.ArgPipesToken)
	if cloudToken == "" {
		return nil, error_helpers.NewErrorsAndWarning(error_helpers.MissingCloudTokenError())
	}

	// so we have a database and a token - build the connection string and set it in viper
	pipesMetadata, err := pipes.GetPipesMetadata(context.Background(), *p.CloudWorkspace, cloudToken)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// set the default conneciton to the cloud metadata
	return pipesMetadata, error_helpers.ErrorAndWarnings{}
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
