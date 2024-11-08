package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/printers"
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
	GetFullName() string
	OnDecoded(*hcl.Block, ModResourcesProvider) hcl.Diagnostics
	GetDeclRange() *hcl.Range
	GetBlockType() string
	GetDescription() string
	GetDocumentation() string
	GetTags() map[string]string
	SetTopLevel(bool)
	IsTopLevel() bool
	GetBase() HclResource
	GetNestedStructs() []CtyValueProvider
	GetHclResourceImpl() *HclResourceImpl
}

// ModTreeItem must be implemented by elements of the mod resource hierarchy
// i.e. Control, Benchmark, Dashboard
type ModTreeItem interface {
	HclResource
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
	GetDatabase() connection.ConnectionStringProvider
	GetSearchPath() []string
	GetSearchPathPrefix() []string
	SetDatabase(connection.ConnectionStringProvider)
	SetSearchPath([]string)
	SetSearchPathPrefix([]string)
}

type ModItem interface {
	GetMod() *Mod
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
	GetResourceWithMetadataRemain() hcl.Body
}

type ModResources interface {
	WalkResources(resourceFunc func(item HclResource) (bool, error)) error
	AddResource(item HclResource) hcl.Diagnostics
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
	Equals(other ModResources) bool
	AddReference(ref *ResourceReference)
	GetReferences() map[string]*ResourceReference
	GetVariables() map[string]*Variable
	GetMods() map[string]*Mod
	TopLevelResources() ModResources
	AddMaps(i ...ModResources)
}

type ModResourcesProvider interface {
	GetModResources() ModResources
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}

type ResourceProvider interface {
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}
