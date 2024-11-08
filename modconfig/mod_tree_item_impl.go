package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
)

type ModTreeItemImpl struct {
	HclResourceImpl
	// required to allow partial decoding
	ModTreeItemRemain hcl.Body `hcl:",remain" json:"-"`

	// auto cty serialisation fails with an NRE for mod struct so we manually serialise
	Mod              *Mod                                `cty:"-" json:"-"`
	DatabaseString   *string                             `cty:"database" hcl:"database" json:"database,omitempty"`
	Database         connection.ConnectionStringProvider `cty:"-" json:"-"`
	SearchPath       []string                            `cty:"search_path" hcl:"search_path,optional" json:"search_path,omitempty"`
	SearchPathPrefix []string                            `cty:"search_path_prefix" hcl:"search_path_prefix,optional" json:"search_path_prefix,omitempty"`

	Paths            []NodePath    `json:"path,omitempty"`
	Children         []ModTreeItem `json:"-"`
	ChildNameStrings []string      `cty:"child_name_strings" json:"children,omitempty"`

	// node may have multiple parents
	// use a map to avoid dupes
	parents map[string]ModTreeItem
}

func NewModTreeItemImpl(block *hcl.Block, mod *Mod, shortName string) ModTreeItemImpl {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	return ModTreeItemImpl{
		HclResourceImpl: NewHclResourceImpl(block, fullName),
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
	return b.Children
}

func (b *ModTreeItemImpl) SetChildren(children []ModTreeItem) {
	b.Children = children
}

func (b *ModTreeItemImpl) AddChild(children ...ModTreeItem) {
	b.Children = append(b.Children, children...)
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
func (b *ModTreeItemImpl) GetDatabase() connection.ConnectionStringProvider {
	if b.Database != nil {
		return b.Database
	}

	// if we have a parent, ask for its database
	// (stop when we get to the mod - the mod database property has lower precedence)
	if len(b.parents) > 0 {
		if parent := b.GetParents()[0]; parent.GetBlockType() != schema.BlockTypeMod {
			return parent.GetDatabase()
		}
	}

	return nil
}

// GetSearchPath implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPath() []string {
	if len(b.SearchPath) != 0 {
		return b.SearchPath
	}
	// if we have a parent, ask for its search path
	// (stop when we get to the mod - the mod database property has lower precedence)
	if len(b.parents) > 0 {
		if parent := b.GetParents()[0]; parent.GetBlockType() != schema.BlockTypeMod {
			return parent.GetSearchPath()
		}
	}

	return nil
}

// GetSearchPathPrefix implements DatabaseItem
func (b *ModTreeItemImpl) GetSearchPathPrefix() []string {
	if len(b.SearchPathPrefix) != 0 {
		return b.SearchPathPrefix
	}
	// if we have a parent, ask for its search path prefix
	// (stop when we get to the mod - the mod database property has lower precedence)
	if len(b.parents) > 0 {
		if parent := b.GetParents()[0]; parent.GetBlockType() != schema.BlockTypeMod {
			return parent.GetSearchPath()
		}
	}

	return nil
}

// SetDatabase implements DatabaseItem
func (b *ModTreeItemImpl) SetDatabase(database connection.ConnectionStringProvider) {
	b.Database = database
}

// SetSearchPath implements DatabaseItem
func (b *ModTreeItemImpl) SetSearchPath(searchPath []string) {
	b.SearchPath = searchPath
}

// SetSearchPathPrefix implements DatabaseItem
func (b *ModTreeItemImpl) SetSearchPathPrefix(searchPathPrefix []string) {
	b.SearchPathPrefix = searchPathPrefix
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
	val, err := cty_helpers.GetCtyValue(b)
	if err != nil {
		return cty.NilVal, err
	}
	vm := val.AsValueMap()
	if b.Mod != nil {
		vm["mod"], err = b.Mod.CtyValue()
		if err != nil {
			return cty.NilVal, err
		}
	}
	return cty.ObjectVal(vm), nil
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

func (b *ModTreeItemImpl) GetNestedStructs() []CtyValueProvider {
	// return all nested structs - this is used to get the nested structs for the cty serialisation
	// we return ourselves and our base structs
	return append([]CtyValueProvider{b}, b.HclResourceImpl.GetNestedStructs()...)
}
