package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type HclResourceImpl struct {
	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`
	// FullName is: <modShortName>.<blockType>.<shortName> if there is a mod
	// and <blockType>.<shortName> if there is no mod
	FullName  string  `cty:"name" json:"qualified_name,omitempty"`
	Title     *string `cty:"title" hcl:"title"  json:"title,omitempty"`
	ShortName string  `cty:"short_name" hcl:"name,label" json:"-"`
	// UnqualifiedName is the <blockType>.<shortName>
	UnqualifiedName string            `cty:"unqualified_name" json:"-"`
	Description     *string           `cty:"description" hcl:"description" json:"description,omitempty"`
	Documentation   *string           `cty:"documentation" hcl:"documentation" json:"documentation,omitempty"`
	DeclRange       hcl.Range         `json:"-"` // No corresponding cty tag, so using "-"
	Tags            map[string]string `cty:"tags" hcl:"tags,optional" json:"tags,omitempty"`
	// TODO can we move this out of here?
	MaxConcurrency *int `cty:"max_concurrency" hcl:"max_concurrency,optional" json:"max_concurrency,omitempty"`

	base                HclResource
	BlockType           string `json:"-"`
	disableCtySerialise bool
	isTopLevel          bool
}

// options pattern
type HclResourceImplOption func(*HclResourceImpl)

func WithDisableCtySerialise() HclResourceImplOption {
	return func(b *HclResourceImpl) {
		b.disableCtySerialise = true
	}
}

func NewHclResourceImpl(block *hcl.Block, fullName string, opts ...HclResourceImplOption) HclResourceImpl {
	// full name has been constructed with the correct short name - which may be a synthetic anonymous block name
	// extract short name from final section of full name
	parts := strings.Split(fullName, ".")
	shortName := parts[len(parts)-1]

	res := HclResourceImpl{
		ShortName:       shortName,
		FullName:        fullName,
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		DeclRange:       hclhelpers.BlockRange(block),
		BlockType:       block.Type,
	}

	for _, opt := range opts {
		opt(&res)
	}
	return res
}

func (h *HclResourceImpl) Equals(other *HclResourceImpl) bool {
	if h == nil || other == nil {
		return false
	}

	// Compare FullName
	if h.FullName != other.FullName {
		return false
	}

	// Compare Title (if not nil)
	if (h.Title == nil && other.Title != nil) || (h.Title != nil && other.Title == nil) {
		return false
	}
	if h.Title != nil && other.Title != nil && *h.Title != *other.Title {
		return false
	}

	// Compare ShortName
	if h.ShortName != other.ShortName {
		return false
	}

	// Compare UnqualifiedName
	if h.UnqualifiedName != other.UnqualifiedName {
		return false
	}

	// Compare Description (if not nil)
	if (h.Description == nil && other.Description != nil) || (h.Description != nil && other.Description == nil) {
		return false
	}
	if h.Description != nil && other.Description != nil && *h.Description != *other.Description {
		return false
	}

	// Compare Documentation (if not nil)
	if (h.Documentation == nil && other.Documentation != nil) || (h.Documentation != nil && other.Documentation == nil) {
		return false
	}
	if h.Documentation != nil && other.Documentation != nil && *h.Documentation != *other.Documentation {
		return false
	}

	// Compare Tags
	if len(h.Tags) != len(other.Tags) {
		return false
	}
	for key, value := range h.Tags {
		if otherValue, ok := other.Tags[key]; !ok || value != otherValue {
			return false
		}
	}

	// Compare other fields (blockType, disableCtySerialise, isTopLevel)
	if h.BlockType != other.BlockType || h.disableCtySerialise != other.disableCtySerialise || h.isTopLevel != other.isTopLevel {
		return false
	}

	return true
}

// Name implements HclResource
// return name in format: '<blocktype>.<shortName>'
func (h *HclResourceImpl) Name() string {
	return h.FullName
}

// GetTitle implements HclResource
func (h *HclResourceImpl) GetTitle() string {
	return typehelpers.SafeString(h.Title)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (h *HclResourceImpl) GetUnqualifiedName() string {
	return h.UnqualifiedName
}

// GetShortName implements HclResource
func (h *HclResourceImpl) GetShortName() string {
	return h.ShortName
}

// GetFullName implements *Mod
func (h *HclResourceImpl) GetFullName() string {
	return h.FullName
}

// OnDecoded implements HclResource
func (h *HclResourceImpl) OnDecoded(block *hcl.Block, _ ModResourcesProvider) hcl.Diagnostics {
	return nil
}

// GetDeclRange implements HclResource
func (h *HclResourceImpl) GetDeclRange() *hcl.Range {
	return &h.DeclRange
}

// GetBlockType implements HclResource
func (h *HclResourceImpl) GetBlockType() string {
	return h.BlockType
}

// GetDescription implements HclResource
func (h *HclResourceImpl) GetDescription() string {
	return typehelpers.SafeString(h.Description)
}

// GetDocumentation implements HclResource
func (h *HclResourceImpl) GetDocumentation() string {
	return typehelpers.SafeString(h.Documentation)
}

// GetTags implements HclResource
func (h *HclResourceImpl) GetTags() map[string]string {
	if h.Tags != nil {
		return h.Tags
	}
	return map[string]string{}
}

// GetHclResourceImpl implements HclResource
func (h *HclResourceImpl) GetHclResourceImpl() *HclResourceImpl {
	return h
}

// SetTopLevel implements HclResource
func (h *HclResourceImpl) SetTopLevel(isTopLevel bool) {
	h.isTopLevel = isTopLevel
}

// IsTopLevel implements HclResource
func (h *HclResourceImpl) IsTopLevel() bool {
	return h.isTopLevel
}

// CtyValue implements CtyValueProvider
func (h *HclResourceImpl) CtyValue() (cty.Value, error) {
	if h.disableCtySerialise {
		return cty.Zero, nil
	}
	return cty_helpers.GetCtyValue(h)
}

func (h *HclResourceImpl) SetBase(base HclResource) {
	h.base = base
	h.SetBaseProperties()
}

// GetBase implements HclResource
func (h *HclResourceImpl) GetBase() HclResource {
	return h.base
}

// GetShowData implements printers.Showable
func (h *HclResourceImpl) GetShowData() *printers.RowData {
	return printers.NewRowData(
		printers.NewFieldValue("Name", h.Name()),
		printers.NewFieldValue("Title", h.GetTitle()),
		printers.NewFieldValue("Description", h.GetDescription()),
		printers.NewFieldValue("Documentation", h.GetDocumentation()),
		printers.NewFieldValue("Tags", h.GetTags()),
	)
}

// GetListData implements printers.Showable
func (h *HclResourceImpl) GetListData() *printers.RowData {
	return printers.NewRowData(
		printers.NewFieldValue("NAME", h.Name()),
	)
}

func (h *HclResourceImpl) SetBaseProperties() {
	if h.Title == nil {
		h.Title = h.getBaseImpl().Title
	}
	if h.Description == nil {
		h.Description = h.getBaseImpl().Description
	}

	h.Tags = utils.MergeMaps(h.Tags, h.getBaseImpl().Tags)

}

func (h *HclResourceImpl) getBaseImpl() *HclResourceImpl {
	return h.base.GetHclResourceImpl()
}

func (h *HclResourceImpl) GetNestedStructs() []CtyValueProvider {
	// return all nested structs - this is used to get the nested structs for the cty serialisation
	// we return ourselves
	return []CtyValueProvider{h}
}
