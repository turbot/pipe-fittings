package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ModTreeItemImpl struct {
	HclResourceImpl
	// required to allow partial decoding
	ModTreeItemRemain hcl.Body `hcl:",remain" json:"-"`

	Mod              *Mod    `cty:"mod" json:"-"`
	ConnectionString *string `cty:"connection_string" hcl:"connection_string" json:"-"`

	Paths []NodePath `column:"path,jsonb" json:"-"`

	// TODO DO WE EVER HAVE MULTIPLE PARENTS
	parents  []ModTreeItem
	children []ModTreeItem
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

// GetConnectionString implements ConnectionStringItem, ModTreeItem
func (b *ModTreeItemImpl) GetConnectionString() *string {
	if b.ConnectionString != nil {
		return b.ConnectionString
	}
	if len(b.parents) > 0 {
		return b.parents[0].GetConnectionString()
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
