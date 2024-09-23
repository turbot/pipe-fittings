package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/zclconf/go-cty/cty"
)

// DashboardWith is a struct representing a leaf dashboard node
type DashboardWith struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`
}

func NewDashboardWith(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// with blocks cannot be anonymous
	return &DashboardWith{
		QueryProviderImpl: NewQueryProviderImpl(block, mod, shortName),
	}
}

func (w *DashboardWith) Equals(other *DashboardWith) bool {
	diff := w.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (w *DashboardWith) OnDecoded(_ *hcl.Block, _ ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

func (w *DashboardWith) Diff(other *DashboardWith) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: w,
		Name: w.Name(),
	}

	res.queryProviderDiff(w, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (*DashboardWith) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (*DashboardWith) GetDisplay() string {
	return ""
}

// GetType implements DashboardLeafNode
func (*DashboardWith) GetType() string {
	return ""
}

// CtyValue implements CtyValueProvider
func (w *DashboardWith) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(w)
}
