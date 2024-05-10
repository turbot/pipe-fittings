package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardImage is a struct representing a leaf dashboard node
type DashboardImage struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Src *string `cty:"src" hcl:"src" column:"src,string"  json:"src,omitempty" snapshot:"src"`
	Alt *string `cty:"alt" hcl:"alt" column:"alt,string"  json:"alt,omitempty" snapshot:"alt"`

	// these properties are JSON serialised by the parent LeafRun
	Width   *int    `cty:"width" hcl:"width" column:"width,string"  json:"width,omitempty" `
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`

	Base *DashboardImage `hcl:"base" json:"-"`
}

func NewDashboardImage(block *hcl.Block, mod *Mod, shortName string) HclResource {
	i := &DashboardImage{
		QueryProviderImpl: NewQueryProviderImpl(block, mod, shortName),
	}
	i.SetAnonymous(block)
	return i
}

func (i *DashboardImage) Equals(other *DashboardImage) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (i *DashboardImage) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	i.setBaseProperties()
	return i.QueryProviderImpl.OnDecoded(block, resourceMapProvider)
}

func (i *DashboardImage) Diff(other *DashboardImage) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}
	if !utils.SafeStringsEqual(i.Src, other.Src) {
		res.AddPropertyDiff("Src")
	}

	if !utils.SafeStringsEqual(i.Alt, other.Alt) {
		res.AddPropertyDiff("Alt")
	}

	res.populateChildDiffs(i, other)
	res.queryProviderDiff(i, other)
	res.dashboardLeafNodeDiff(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardImage) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetDisplay implements DashboardLeafNode
func (i *DashboardImage) GetDisplay() string {
	return typehelpers.SafeString(i.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardImage) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (*DashboardImage) GetType() string {
	return ""
}

// ValidateQuery implements QueryProvider
func (i *DashboardImage) ValidateQuery() hcl.Diagnostics {
	// query is optional - nothing to do
	return nil
}

// CtyValue implements CtyValueProvider
func (i *DashboardImage) CtyValue() (cty.Value, error) {
	return GetCtyValue(i)
}

func (i *DashboardImage) setBaseProperties() {
	if i.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	i.base = i.Base
	// call into parent nested struct setBaseProperties
	i.QueryProviderImpl.setBaseProperties()

	if i.Src == nil {
		i.Src = i.Base.Src
	}

	if i.Alt == nil {
		i.Alt = i.Base.Alt
	}

	if i.Width == nil {
		i.Width = i.Base.Width
	}

	if i.Display == nil {
		i.Display = i.Base.Display
	}
}

// GetShowData implements printers.Showable
func (i *DashboardImage) GetShowData() *printers.RowData {
	res := printers.NewRowData(
		printers.NewFieldValue("Width", i.Width),
		printers.NewFieldValue("Display", i.Display),
		printers.NewFieldValue("Src", i.Src),
		printers.NewFieldValue("Alt", i.Alt),
	)
	// merge fields from base, putting base fields first
	res.Merge(i.QueryProviderImpl.GetShowData())
	return res
}
