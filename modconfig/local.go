package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// Local is a struct representing a Local resource
type Local struct {
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Value cty.Value
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	fullName := fmt.Sprintf("%s.local.%s", mod.ShortName, name)
	// create a fake block to pass to NewHclResourceImpl
	b := &hcl.Block{Body: &hclsyntax.Body{SrcRange: declRange}}

	l := &Local{
		Value: val,
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: NewHclResourceImpl(b, fullName, WithDisableCtySerialise()),
		},
	}
	l.Mod = mod
	return l
}

// CtyValue implements CtyValueProvider
func (l *Local) CtyValue() (cty.Value, error) {
	return l.Value, nil
}

func (l *Local) Diff(other *Local) *ModTreeItemDiffs {
	res := &ModTreeItemDiffs{
		Item: l,
		Name: l.Name(),
	}

	if !utils.SafeStringsEqual(l.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(l.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.PopulateChildDiffs(l, other)
	return res
}
