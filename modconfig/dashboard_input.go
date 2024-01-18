package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardInput is a struct representing a leaf dashboard node
type DashboardInput struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	DashboardName string                  `column:"dashboard,string" json:"dashboard,omitempty"`
	Label         *string                 `cty:"label" hcl:"label" column:"label,string" json:"label,omitempty"`
	Placeholder   *string                 `cty:"placeholder" hcl:"placeholder" column:"placeholder,string" json:"placeholder,omitempty"`
	Options       []*DashboardInputOption `cty:"options" hcl:"option,block" json:"options,omitempty" snapshot:"options"`
	// tactical - exists purely so we can put "unqualified_name" in the snbapshot panel for the input
	// TODO remove when input names are refactored https://github.com/turbot/steampipe/issues/2863
	InputName string `cty:"input_name" json:"unqualified_name" snapshot:"unqualified_name"`

	// these properties are JSON serialised by the parent LeafRun
	Width     *int            `cty:"width" hcl:"width" column:"width,string"  json:"width,omitempty"`
	Type      *string         `cty:"type" hcl:"type" column:"type,string"  json:"type,omitempty"`
	Display   *string         `cty:"display" hcl:"display" json:"display,omitempty"`
	Base      *DashboardInput `hcl:"base" json:"-"`
	dashboard *Dashboard
}

func NewDashboardInput(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// input cannot be anonymous
	i := &DashboardInput{
		QueryProviderImpl: NewQueryProviderImpl(block, mod, shortName),
	}

	// tactical set input name
	i.InputName = i.UnqualifiedName

	return i
}

// TODO remove https://github.com/turbot/steampipe/issues/2864
func (i *DashboardInput) Clone() *DashboardInput {
	return &DashboardInput{
		ResourceWithMetadataImpl: i.ResourceWithMetadataImpl,
		QueryProviderImpl:        i.QueryProviderImpl,
		Width:                    i.Width,
		Type:                     i.Type,
		Label:                    i.Label,
		Placeholder:              i.Placeholder,
		Display:                  i.Display,
		Options:                  i.Options,
		InputName:                i.InputName,
		dashboard:                i.dashboard,
	}
}

func (i *DashboardInput) Equals(other *DashboardInput) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (i *DashboardInput) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	i.setBaseProperties()
	return i.QueryProviderImpl.OnDecoded(block, resourceMapProvider)
}

func (i *DashboardInput) Diff(other *DashboardInput) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}

	if !utils.SafeStringsEqual(i.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(i.Label, other.Label) {
		res.AddPropertyDiff("Instance")
	}

	if !utils.SafeStringsEqual(i.Placeholder, other.Placeholder) {
		res.AddPropertyDiff("Placeholder")
	}

	if len(i.Options) != len(other.Options) {
		res.AddPropertyDiff("Options")
	} else {
		for idx, o := range i.Options {
			if !other.Options[idx].Equals(o) {
				res.AddPropertyDiff("Options")
			}
		}
	}

	res.populateChildDiffs(i, other)
	res.queryProviderDiff(i, other)
	res.dashboardLeafNodeDiff(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardInput) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetDisplay implements DashboardLeafNode
func (i *DashboardInput) GetDisplay() string {
	return typehelpers.SafeString(i.Display)
}

// GetType implements DashboardLeafNode
func (i *DashboardInput) GetType() string {
	return typehelpers.SafeString(i.Type)
}

// SetDashboard sets the parent dashboard container
func (i *DashboardInput) SetDashboard(dashboard *Dashboard) {
	i.dashboard = dashboard
	i.DashboardName = dashboard.Name()
}

// ValidateQuery implements QueryProvider
func (i *DashboardInput) ValidateQuery() hcl.Diagnostics {
	// inputs with placeholder or options, or text type do not need a query
	if i.Placeholder != nil ||
		len(i.Options) > 0 ||
		typehelpers.SafeString(i.Type) == "text" {
		return nil
	}

	return i.QueryProviderImpl.ValidateQuery()
}

// DependsOnInput returns whether this input has a runtime dependency on the given input¬
func (i *DashboardInput) DependsOnInput(changedInputName string) bool {
	for _, r := range i.runtimeDependencies {
		if r.SourceResourceName() == changedInputName {
			return true
		}
	}
	return false
}

// CtyValue implements CtyValueProvider
func (i *DashboardInput) CtyValue() (cty.Value, error) {
	return GetCtyValue(i)
}

func (i *DashboardInput) setBaseProperties() {
	if i.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	i.base = i.Base
	// call into parent nested struct setBaseProperties
	i.QueryProviderImpl.setBaseProperties()

	if i.Type == nil {
		i.Type = i.Base.Type
	}

	if i.Display == nil {
		i.Display = i.Base.Display
	}

	if i.Label == nil {
		i.Label = i.Base.Label
	}

	if i.Placeholder == nil {
		i.Placeholder = i.Base.Placeholder
	}

	if i.Width == nil {
		i.Width = i.Base.Width
	}

	if i.Options == nil {
		i.Options = i.Base.Options
	}
}
