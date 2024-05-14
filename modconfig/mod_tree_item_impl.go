package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
)

type ModTreeItemImpl struct {
	HclResourceImpl
	// required to allow partial decoding
	ModTreeItemRemain hcl.Body `hcl:",remain" json:"-"`

	Mod              *Mod     `cty:"mod" json:"-"`
	Database         *string  `cty:"database" hcl:"database" json:"database,omitempty"`
	SearchPath       []string `cty:"search_path" hcl:"search_path,optional" json:"search_path,omitempty"`
	SearchPathPrefix []string `cty:"search_path_prefix" hcl:"search_path_prefix,optional" json:"search_path_prefix,omitempty"`

	Paths []NodePath `json:"path,omitempty"`

	// node may have multiple parents
	// use a map to avoid dupes
	parents  map[string]ModTreeItem
	children []ModTreeItem
}

func NewModTreeItemImpl(block *hcl.Block, mod *Mod, shortName string) ModTreeItemImpl {
	return ModTreeItemImpl{
		HclResourceImpl: NewHclResourceImpl(block, mod, shortName),
		Mod:             mod,
		parents:         make(map[string]ModTreeItem),
	}
}

// AddParent implements ModTreeItem
func (b *ModTreeItemImpl) AddParent(parent ModTreeItem) error {
	// lazily create the map
	if b.parents == nil {
		b.parents = make(map[string]ModTreeItem)
	}
	b.parents[parent.Name()] = parent
	return nil
}

// GetParents implements ModTreeItem
func (b *ModTreeItemImpl) GetParents() []ModTreeItem {
	// lazily create the map
	if b.parents == nil {
		b.parents = make(map[string]ModTreeItem)
	}
	return maps.Values(b.parents)
}

// GetChildren implements ModTreeItem
func (b *ModTreeItemImpl) GetChildren() []ModTreeItem {
	return b.children
}
func (b *ModTreeItemImpl) GetPaths() []NodePath {
	// lazy load
	if len(b.Paths) == 0 {
		b.SetPaths()
	}
	return b.Paths
}

// SetPaths implements ModTreeItem
func (b *ModTreeItemImpl) SetPaths() {
	for _, parent := range b.parents {
		for _, parentPath := range parent.GetPaths() {
			b.Paths = append(b.Paths, append(parentPath, b.FullName))
		}
	}
}

// GetMod implements ModItem, ModTreeItem
func (b *ModTreeItemImpl) GetMod() *Mod {
	return b.Mod
}

// GetDatabase implements DatabaseItem
func (b *ModTreeItemImpl) GetDatabase() *string {
	if b.Database != nil {
		return b.Database
	}
	if len(b.parents) > 0 {
		return b.GetParents()[0].GetDatabase()
	}
	return nil
}

// GetSearchPath implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPath() []string {
	if len(b.SearchPath) != 0 {
		return b.SearchPath
	}
	if len(b.parents) > 0 {
		return b.GetParents()[0].GetSearchPath()
	}

	return nil
}

// GetSearchPathPrefix implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPathPrefix() []string {
	if len(b.SearchPathPrefix) != 0 {
		return b.SearchPathPrefix
	}
	if len(b.GetParents()) > 0 {
		return b.GetParents()[0].GetSearchPathPrefix()
	}
	return nil
}

// GetModTreeItemImpl implements ModTreeItem
func (b *ModTreeItemImpl) GetModTreeItemImpl() *ModTreeItemImpl {
	return b
}

// CtyValue implements CtyValueProvider
func (b *ModTreeItemImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}

// GetShowData implements printers.Showable
func (b *ModTreeItemImpl) GetShowData() *printers.RowData {
	var name = b.ShortName
	// if this is a dependency resource, use the full name
	if b.IsDependencyResource() {
		name = b.Name()
	}
	res := printers.NewRowData(
		// override name to take parents into account - merge will handle this and ignore the base name
		printers.NewFieldValue("Name", name),
	)
	if b.Mod != nil {
		res.AddField(printers.NewFieldValue("Mod", b.Mod.ShortName))
	}
	res.AddField(printers.NewFieldValue("Database", b.Database))

	// merge fields from base, putting base fields first
	res.Merge(b.HclResourceImpl.GetShowData())
	return res
}

// GetListData implements printers.Listable
func (b *ModTreeItemImpl) GetListData() *printers.RowData {
	var name = b.ShortName
	if b.IsDependencyResource() {
		name = b.Name()
	}
	res := printers.NewRowData()
	if b.Mod != nil {
		res.AddField(printers.NewFieldValue("MOD", b.Mod.ShortName))
	}

	res.AddField(printers.NewFieldValue("NAME", name))
	// NOTE - do not merge the base fields here, which only includes NAME, as we want to override the order of the fields
	//res.Merge(b.HclResourceImpl.GetListData())

	return res
}

func (b *ModTreeItemImpl) IsDependencyResource() bool {
	return b.GetMod().DependencyPath != nil
}
