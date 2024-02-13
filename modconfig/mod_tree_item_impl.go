package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/zclconf/go-cty/cty"
)

type ModTreeItemImpl struct {
	HclResourceImpl
	// required to allow partial decoding
	ModTreeItemRemain hcl.Body `hcl:",remain" json:"-"`

	Mod              *Mod     `cty:"mod" json:"-"`
	Database         *string  `cty:"database" hcl:"database" json:"database,omitempty"`
	SearchPath       []string `cty:"search_path" hcl:"search_path,optional" json:"search_path,omitempty"`
	SearchPathPrefix []string `cty:"search_path_prefix" hcl:"search_path_prefix,optional" json:"search_path_prefix,omitempty"`

	Paths []NodePath `column:"path,jsonb" json:"path,omitempty"`

	// TODO DO WE EVER HAVE MULTIPLE PARENTS
	parents  []ModTreeItem
	children []ModTreeItem
}

func NewModTreeItemImpl(block *hcl.Block, mod *Mod, shortName string) ModTreeItemImpl {
	return ModTreeItemImpl{
		HclResourceImpl: NewHclResourceImpl(block, mod, shortName),
		Mod:             mod,
	}
}

// AddParent implements ModTreeItem
func (b *ModTreeItemImpl) AddParent(parent ModTreeItem) error {
	b.parents = append(b.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (b *ModTreeItemImpl) GetParents() []ModTreeItem {
	return b.parents
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
		return b.parents[0].GetDatabase()
	}
	return nil
}

// GetSearchPath implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPath() []string {
	if len(b.SearchPath) != 0 {
		return b.SearchPath
	}
	if len(b.parents) > 0 {
		return b.parents[0].GetSearchPath()
	}
	return nil
}

// GetSearchPathPrefix implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPathPrefix() []string {
	if len(b.SearchPathPrefix) != 0 {
		return b.SearchPathPrefix
	}
	if len(b.parents) > 0 {
		return b.parents[0].GetSearchPathPrefix()
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
	// override name to take parents into account
	var name = b.ShortName
	if b.parents != nil {
		name = b.Name()

	}
	res := printers.NewRowData(
		printers.FieldValue{Name: "Name", Value: name},
		printers.FieldValue{Name: "Mod", Value: b.Mod.ShortName},
		printers.FieldValue{Name: "Database", Value: b.Database},
	)
	res.Merge(b.HclResourceImpl.GetShowData())
	return res
}

// GetListData implements printers.Listable
func (b *ModTreeItemImpl) GetListData() *printers.RowData {
	var name = b.ShortName
	if b.parents != nil {
		name = b.Name()
	}
	res := printers.NewRowData(
		printers.FieldValue{Name: "NAME", Value: name},
		printers.FieldValue{Name: "MOD", Value: b.Mod.ShortName},
	)
	res.Merge(b.HclResourceImpl.GetListData())
	return res
}
