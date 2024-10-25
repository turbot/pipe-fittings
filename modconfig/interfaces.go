package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/plugin"
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
	OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics
	GetDeclRange() *hcl.Range
	GetBlockType() string
	GetDescription() string
	GetDocumentation() string
	GetTags() map[string]string
	SetTopLevel(bool)
	IsTopLevel() bool
	GetBase() HclResource
	GetHclResourceImpl() *HclResourceImpl
}

// TODO K make generic??
// type ModI[T ResourceMapsI] interface {
type ModI interface {
	HclResource
	ModTreeItem
	ResourceProvider
	ResourceWithMetadata
	GetDependencyName() string
	GetDependencyPath() *string
	GetModPath() string
	SetFilePath(string)
	GetFilePath() string
	IsDefaultMod() bool
	GetResourceMaps() ResourceMapsI
	SetResourceMaps(maps ResourceMapsI)
	GetInstallCacheKey() string
	AddResource(item HclResource) hcl.Diagnostics
	GetRequire() *Require
	SetRequire(*Require)
	HasDependentMods() bool
	WalkResources(resourceFunc func(item HclResource) (bool, error)) error
	SetDatabase(*string)
	SetSearchPath([]string)
	SetSearchPathPrefix([]string)
	ValidateRequirements(versionMap *plugin.PluginVersionMap) []error
	SetTitle(string)
	BuildResourceTree(mods ModMap) error
	SetDependencyConfigFromPath(dependencyPath string) error
	SetDependencyConfig(*DependencyVersion, *string, string)
	GetModDependency(modName string) *ModVersionConstraint
	RemoveAllModDependencies()
	RemoveModDependencies(mods map[string]*ModVersionConstraint)
	AddModDependencies(mods map[string]*ModVersionConstraint)
	GetVersion() *DependencyVersion
	Save() error
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
	GetDatabase() *string
	GetSearchPath() []string
	GetSearchPathPrefix() []string
	SetDatabase(*string)
	SetSearchPath([]string)
	SetSearchPathPrefix([]string)
}

type ModItem interface {
	GetMod() ModI
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

type ResourceMapsI interface {
	WalkResources(resourceFunc func(item HclResource) (bool, error)) error
	AddResource(item HclResource) hcl.Diagnostics
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
	Equals(other ResourceMapsI) bool
	AddReference(ref *ResourceReference)
	GetReferences() map[string]*ResourceReference
	GetVariables() map[string]*Variable
	GetMods() map[string]ModI
	TopLevelResources() ResourceMapsI
	AddMaps(i ...ResourceMapsI)
}

type ResourceMapsProvider interface {
	GetResourceMaps() ResourceMapsI
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}

type ResourceProvider interface {
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}
