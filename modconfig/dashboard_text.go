package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardText is a struct representing a leaf dashboard node
type DashboardText struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Value   *string `cty:"value" hcl:"value" column:"value,string" snapshot:"value"  json:"value,omitempty"`
	Width   *int    `cty:"width" hcl:"width" column:"width,string"  json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,string"  json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base *DashboardText `hcl:"base" json:"-"`
	Mod  *Mod           `cty:"mod" json:"-"`
}

func NewDashboardText(block *hcl.Block, mod *Mod, shortName string) HclResource {
	t := &DashboardText{
		ModTreeItemImpl: NewModTreeItemImpl(block, mod, shortName),
	}
	t.SetAnonymous(block)
	return t
}

func (t *DashboardText) Equals(other *DashboardText) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (t *DashboardText) OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics {
	t.setBaseProperties()
	return nil
}

func (t *DashboardText) Diff(other *DashboardText) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}

	if !utils.SafeStringsEqual(t.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(t.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(t, other)
	res.dashboardLeafNodeDiff(t, other)
	return res
}

// GetWidth implements DashboardLeafNode
func (t *DashboardText) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetDisplay implements DashboardLeafNode
func (t *DashboardText) GetDisplay() string {
	return typehelpers.SafeString(t.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardText) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (t *DashboardText) GetType() string {
	return typehelpers.SafeString(t.Type)
}

// CtyValue implements CtyValueProvider
func (t *DashboardText) CtyValue() (cty.Value, error) {
	return GetCtyValue(t)
}

func (t *DashboardText) setBaseProperties() {
	if t.Base == nil {
		return
	}
	if t.Title == nil {
		t.Title = t.Base.Title
	}
	if t.Type == nil {
		t.Type = t.Base.Type
	}
	if t.Display == nil {
		t.Display = t.Base.Display
	}
	if t.Value == nil {
		t.Value = t.Base.Value
	}
	if t.Width == nil {
		t.Width = t.Base.Width
	}
}
