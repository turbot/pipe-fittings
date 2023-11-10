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

type PowerpipeWorkspaceProfile struct {
	ProfileName       string                     `hcl:"name,label" cty:"name"`
	CloudHost         *string                    `hcl:"cloud_host,optional" cty:"cloud_host"`
	CloudToken        *string                    `hcl:"cloud_token,optional" cty:"cloud_token"`
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
	Base              *PowerpipeWorkspaceProfile `hcl:"base"`

	// options
	QueryOptions     *options.Query                     `cty:"query-options"`
	CheckOptions     *options.Check                     `cty:"check-options"`
	DashboardOptions *options.WorkspaceProfileDashboard `cty:"dashboard-options"`
	DeclRange        hcl.Range
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (p *PowerpipeWorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Query:
		if p.QueryOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.QueryOptions = o
	case *options.Check:
		if p.CheckOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.CheckOptions = o
	case *options.WorkspaceProfileDashboard:
		if p.DashboardOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.DashboardOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Workspace", reflect.TypeOf(o).Name()),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
	}
	return diags
}

func (p *PowerpipeWorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}
func (p *PowerpipeWorkspaceProfile) ShortName() string {
	return p.ProfileName
}

func (p *PowerpipeWorkspaceProfile) CtyValue() (cty.Value, error) {
	return GetCtyValue(p)
}

func (p *PowerpipeWorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *PowerpipeWorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}

	if p.CloudHost == nil {
		p.CloudHost = p.Base.CloudHost
	}
	if p.CloudToken == nil {
		p.CloudToken = p.Base.CloudToken
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

	// nested inheritance strategy:
	//
	// if my nested struct is a nil
	//		-> use the base struct
	//
	// if I am not nil (and base is not nil)
	//		-> only inherit the properties which are nil in me and not in base
	//
	if p.QueryOptions == nil {
		p.QueryOptions = p.Base.QueryOptions
	} else {
		p.QueryOptions.SetBaseProperties(p.Base.QueryOptions)
	}
	if p.CheckOptions == nil {
		p.CheckOptions = p.Base.CheckOptions
	} else {
		p.CheckOptions.SetBaseProperties(p.Base.CheckOptions)
	}
	if p.DashboardOptions == nil {
		p.DashboardOptions = p.Base.DashboardOptions
	} else {
		p.DashboardOptions.SetBaseProperties(p.Base.DashboardOptions)
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *PowerpipeWorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map

	res.SetStringItem(p.CloudHost, constants.ArgCloudHost)
	res.SetStringItem(p.CloudToken, constants.ArgCloudToken)
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
	if cmd.Name() == constants.CmdNameCheck && p.CheckOptions != nil {
		res.PopulateConfigMapForOptions(p.CheckOptions)
	}
	if cmd.Name() == constants.CmdNameDashboard && p.DashboardOptions != nil {
		res.PopulateConfigMapForOptions(p.DashboardOptions)
	}

	return res
}

func (p *PowerpipeWorkspaceProfile) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

func (p *PowerpipeWorkspaceProfile) GetInstallDir() *string {
	return p.InstallDir
}

func (p *PowerpipeWorkspaceProfile) IsNil() bool {
	return p == nil
}
