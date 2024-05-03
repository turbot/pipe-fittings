package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/printers"
	"github.com/zclconf/go-cty/cty"
)

// HclResource must be implemented by resources defined in HCL
type HclResource interface {
	printers.Showable
	printers.Listable
	Name() string
	GetTitle() string
	GetUnqualifiedName() string
	GetShortName() string
	OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics
	GetDeclRange() *hcl.Range
	BlockType() string
	GetDescription() string
	GetDocumentation() string
	GetTags() map[string]string
	SetTopLevel(bool)
	IsTopLevel() bool
	GetBase() HclResource
	GetHclResourceImpl() *HclResourceImpl
}

// ModTreeItem must be implemented by elements of the mod resource hierarchy
// i.e. Control, Benchmark, Dashboard
type ModTreeItem interface {
	HclResource
	ModItem
	DatabaseItem

	AddParent(ModTreeItem) error
	GetParents() []ModTreeItem
	GetChildren() []ModTreeItem
	// GetPaths returns an array resource paths
	GetPaths() []NodePath
	SetPaths()
	GetModTreeItemImpl() *ModTreeItemImpl
	IsDependencyResource() bool
}

type DatabaseItem interface {
	GetDatabase() *string
	GetSearchPath() []string
	GetSearchPathPrefix() []string
}

type ModItem interface {
	GetMod() *Mod
}

// RuntimeDependencyProvider is implemented by all QueryProviders and Dashboard
type RuntimeDependencyProvider interface {
	ModTreeItem
	AddRuntimeDependencies([]*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency
}

type WithProvider interface {
	AddWith(with *DashboardWith) hcl.Diagnostics
	GetWiths() []*DashboardWith
	GetWith(string) (*DashboardWith, bool)
}

// QueryProvider must be implemented by resources which have query/sql
type QueryProvider interface {
	RuntimeDependencyProvider
	GetArgs() *QueryArgs
	GetParams() []*ParamDef
	GetSQL() *string
	GetQuery() *Query
	SetArgs(*QueryArgs)
	SetParams([]*ParamDef)
	GetResolvedQuery(*QueryArgs) (*ResolvedQuery, error)
	RequiresExecution(QueryProvider) bool
	ValidateQuery() hcl.Diagnostics
	MergeParentArgs(QueryProvider, QueryProvider) hcl.Diagnostics
	GetQueryProviderImpl() *QueryProviderImpl
	ParamsInheritedFromBase() bool
	ArgsInheritedFromBase() bool
}

type CtyValueProvider interface {
	CtyValue() (cty.Value, error)
}

// ResourceWithMetadata must be implemented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
	SetAnonymous(block *hcl.Block)
	IsAnonymous() bool
	AddReference(ref *ResourceReference)
	GetReferences() []*ResourceReference
}

// DashboardLeafNode must be implemented by resources may be a leaf node in the dashboard execution tree
type DashboardLeafNode interface {
	ModTreeItem
	ResourceWithMetadata
	GetDisplay() string
	GetType() string
	GetWidth() int
}

type ResourceMapsProvider interface {
	GetResourceMaps() *ResourceMaps
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}

// NodeAndEdgeProvider must be implemented by any dashboard leaf node which supports edges and nodes
// (DashboardGraph, DashboardFlow, DashboardHierarchy)
// TODO [node_reuse] add NodeAndEdgeProviderImpl https://github.com/turbot/steampipe/issues/2918
type NodeAndEdgeProvider interface {
	QueryProvider
	WithProvider
	GetEdges() DashboardEdgeList
	SetEdges(DashboardEdgeList)
	GetNodes() DashboardNodeList
	SetNodes(DashboardNodeList)
	AddCategory(category *DashboardCategory) hcl.Diagnostics
	AddChild(child HclResource) hcl.Diagnostics
}
