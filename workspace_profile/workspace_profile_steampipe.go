package workspace_profile

import (
	"fmt"

	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type SteampipeWorkspaceProfile struct {
	ProfileName       string                     `hcl:"name,label" cty:"name"`
	PipesHost         *string                    `hcl:"pipes_host,optional" cty:"pipes_host"`
	PipesToken        *string                    `hcl:"pipes_token,optional" cty:"pipes_token"`
	InstallDir        *string                    `hcl:"install_dir,optional" cty:"install_dir"`
	QueryTimeout      *int                       `hcl:"query_timeout,optional" cty:"query_timeout"`
	SnapshotLocation  *string                    `hcl:"snapshot_location,optional" cty:"snapshot_location"`
	WorkspaceDatabase *string                    `hcl:"workspace_database,optional" cty:"workspace_database"`
	SearchPath        *string                    `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix  *string                    `hcl:"search_path_prefix" cty:"search_path_prefix"`
	Watch             *bool                      `hcl:"watch" cty:"watch"`
	MaxParallel       *int                       `hcl:"max_parallel" cty:"max-parallel"`
	Introspection     *string                    `hcl:"introspection" cty:"introspection"`
	Input             *bool                      `hcl:"input" cty:"input"`
	Progress          *bool                      `hcl:"progress" cty:"progress"`
	Theme             *string                    `hcl:"theme" cty:"theme"`
	Cache             *bool                      `hcl:"cache" cty:"cache"`
	CacheTTL          *int                       `hcl:"cache_ttl" cty:"cache_ttl"`
	Base              *SteampipeWorkspaceProfile `hcl:"base"`

	// options
	QueryOptions *options.Query `cty:"query-options"`
	DeclRange    hcl.Range
	block        *hcl.Block
}

func NewSteampipeWorkspaceProfile(block *hcl.Block) *SteampipeWorkspaceProfile {
	return &SteampipeWorkspaceProfile{
		ProfileName: block.Labels[0],
		DeclRange:   hclhelpers.BlockRange(block),
		block:       block,
	}
}

func (p *SteampipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *SteampipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// GetOptionsForBlock returns the workspace profile options object for the given block
func (p *SteampipeWorkspaceProfile) GetOptionsForBlock(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	switch block.Labels[0] {

	case options.QueryBlock:
		return new(options.Query), nil
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Type),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return nil, diags
	}
}

func (p *SteampipeWorkspaceProfile) GetInstallDir() *string {
	return p.InstallDir
}

func (p *SteampipeWorkspaceProfile) IsNil() bool {
	return p == nil
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type
func (p *SteampipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Query:
		if p.QueryOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.QueryOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
	}
	return diags
}

func duplicateOptionsBlockDiag(block *hcl.Block) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("duplicate %s options block", block.Type),
		Subject:  hclhelpers.BlockRangePointer(block),
	}
}

func (p *SteampipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *SteampipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(p)
}

func (p *SteampipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *SteampipeWorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}
	if p.InstallDir == nil {
		p.InstallDir = p.Base.InstallDir
	}
	if p.SnapshotLocation == nil {
		p.SnapshotLocation = p.Base.SnapshotLocation
	}
	if p.WorkspaceDatabase == nil {
		p.WorkspaceDatabase = p.Base.WorkspaceDatabase
	}
	if p.QueryTimeout == nil {
		p.QueryTimeout = p.Base.QueryTimeout
	}
	if p.SearchPath == nil {
		p.SearchPath = p.Base.SearchPath
	}
	if p.SearchPathPrefix == nil {
		p.SearchPathPrefix = p.Base.SearchPathPrefix
	}
	if p.Watch == nil {
		p.Watch = p.Base.Watch
	}
	if p.MaxParallel == nil {
		p.MaxParallel = p.Base.MaxParallel
	}
	if p.Introspection == nil {
		p.Introspection = p.Base.Introspection
	}
	if p.Input == nil {
		p.Input = p.Base.Input
	}
	if p.Progress == nil {
		p.Progress = p.Base.Progress
	}
	if p.Theme == nil {
		p.Theme = p.Base.Theme
	}
	if p.Cache == nil {
		p.Cache = p.Base.Cache
	}
	if p.CacheTTL == nil {
		p.CacheTTL = p.Base.CacheTTL
	}
	if p.QueryOptions == nil {
		p.QueryOptions = p.Base.QueryOptions
	} else {
		p.QueryOptions.SetBaseProperties(p.Base.QueryOptions)
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *SteampipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map

	res.SetStringItem(p.PipesHost, constants.ArgPipesHost)
	res.SetStringItem(p.PipesToken, constants.ArgPipesToken)
	res.SetStringItem(p.InstallDir, constants.ArgInstallDir)
	res.SetStringItem(p.SnapshotLocation, constants.ArgSnapshotLocation)
	res.SetStringItem(p.WorkspaceDatabase, constants.ArgWorkspaceDatabase)
	res.SetIntItem(p.QueryTimeout, constants.ArgDatabaseQueryTimeout)
	res.SetBoolItem(p.Watch, constants.ArgWatch)
	res.SetIntItem(p.MaxParallel, constants.ArgMaxParallel)
	res.SetStringSliceItem(searchPathFromString(p.SearchPath, ","), constants.ArgSearchPath)
	res.SetStringSliceItem(searchPathFromString(p.SearchPathPrefix, ","), constants.ArgSearchPathPrefix)
	res.SetStringItem(p.Introspection, constants.ArgIntrospection)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Progress, constants.ArgProgress)
	res.SetStringItem(p.Theme, constants.ArgTheme)
	res.SetBoolItem(p.Cache, constants.ArgClientCacheEnabled)
	res.SetIntItem(p.CacheTTL, constants.ArgCacheTtl)

	if cmd.Name() == constants.CmdNameQuery && p.QueryOptions != nil {
		res.PopulateConfigMapForOptions(p.QueryOptions)
	}

	return res
}
