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
	blockType           string
	disableCtySerialise bool
	isTopLevel          bool
}

func NewHclResourceImpl(block *hcl.Block, fullName string) HclResourceImpl {
	// full name has been constructed with the correct short name - which may be a synthetic anonymous block name
	// extract short name from final section of full name
	parts := strings.Split(fullName, ".")
	shortName := parts[len(parts)-1]

	return HclResourceImpl{
		ShortName:       shortName,
		FullName:        fullName,
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		DeclRange:       hclhelpers.BlockRange(block),
		blockType:       block.Type,
	}
}

func NewHclResourceImplNoMod(block *hcl.Block, resourceType, shortName string) HclResourceImpl {
	fullName := fmt.Sprintf("%s.%s", resourceType, shortName)
	return HclResourceImpl{
		ShortName:       shortName,
		FullName:        fullName,
		UnqualifiedName: fmt.Sprintf("%s.%s", resourceType, shortName),
		DeclRange:       hclhelpers.BlockRange(block),
		blockType:       block.Type,
	}
}

func (b *HclResourceImpl) Equals(other *HclResourceImpl) bool {
	if b == nil || other == nil {
		return false
	}

	// Compare FullName
	if b.FullName != other.FullName {
		return false
	}

	// Compare Title (if not nil)
	if (b.Title == nil && other.Title != nil) || (b.Title != nil && other.Title == nil) {
		return false
	}
	if b.Title != nil && other.Title != nil && *b.Title != *other.Title {
		return false
	}

	// Compare ShortName
	if b.ShortName != other.ShortName {
		return false
	}

	// Compare UnqualifiedName
	if b.UnqualifiedName != other.UnqualifiedName {
		return false
	}

	// Compare Description (if not nil)
	if (b.Description == nil && other.Description != nil) || (b.Description != nil && other.Description == nil) {
		return false
	}
	if b.Description != nil && other.Description != nil && *b.Description != *other.Description {
		return false
	}

	// Compare Documentation (if not nil)
	if (b.Documentation == nil && other.Documentation != nil) || (b.Documentation != nil && other.Documentation == nil) {
		return false
	}
	if b.Documentation != nil && other.Documentation != nil && *b.Documentation != *other.Documentation {
		return false
	}

	// Compare Tags
	if len(b.Tags) != len(other.Tags) {
		return false
	}
	for key, value := range b.Tags {
		if otherValue, ok := other.Tags[key]; !ok || value != otherValue {
			return false
		}
	}

	// Compare other fields (blockType, disableCtySerialise, isTopLevel)
	if b.blockType != other.blockType || b.disableCtySerialise != other.disableCtySerialise || b.isTopLevel != other.isTopLevel {
		return false
	}

	return true
}

// Name implements HclResource
// return name in format: '<blocktype>.<shortName>'
func (b *HclResourceImpl) Name() string {
	return b.FullName
}

// GetTitle implements HclResource
func (b *HclResourceImpl) GetTitle() string {
	return typehelpers.SafeString(b.Title)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (b *HclResourceImpl) GetUnqualifiedName() string {
	return b.UnqualifiedName
}

// GetShortName implements HclResource
func (b *HclResourceImpl) GetShortName() string {
	return b.ShortName
}

// OnDecoded implements HclResource
func (b *HclResourceImpl) OnDecoded(block *hcl.Block, _ ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// GetDeclRange implements HclResource
func (b *HclResourceImpl) GetDeclRange() *hcl.Range {
	return &b.DeclRange
}

// BlockType implements HclResource
func (b *HclResourceImpl) BlockType() string {
	return b.blockType
}

// GetDescription implements HclResource
func (b *HclResourceImpl) GetDescription() string {
	return typehelpers.SafeString(b.Description)
}

// GetDocumentation implements HclResource
func (b *HclResourceImpl) GetDocumentation() string {
	return typehelpers.SafeString(b.Documentation)
}

// GetTags implements HclResource
func (b *HclResourceImpl) GetTags() map[string]string {
	if b.Tags != nil {
		return b.Tags
	}
	return map[string]string{}
}

// GetHclResourceImpl implements HclResource
func (b *HclResourceImpl) GetHclResourceImpl() *HclResourceImpl {
	return b
}

// SetTopLevel implements HclResource
func (b *HclResourceImpl) SetTopLevel(isTopLevel bool) {
	b.isTopLevel = isTopLevel
}

// IsTopLevel implements HclResource
func (b *HclResourceImpl) IsTopLevel() bool {
	return b.isTopLevel
}

// CtyValue implements CtyValueProvider
func (b *HclResourceImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return cty_helpers.GetCtyValue(b)
}

// GetBase implements HclResource
func (b *HclResourceImpl) GetBase() HclResource {
	return b.base
}

// GetShowData implements printers.Showable
func (b *HclResourceImpl) GetShowData() *printers.RowData {
	return printers.NewRowData(
		printers.NewFieldValue("Name", b.Name()),
		printers.NewFieldValue("Title", b.GetTitle()),
		printers.NewFieldValue("Description", b.GetDescription()),
		printers.NewFieldValue("Documentation", b.GetDocumentation()),
		printers.NewFieldValue("Tags", b.GetTags()),
	)
}

// GetListData implements printers.Showable
func (b *HclResourceImpl) GetListData() *printers.RowData {
	return printers.NewRowData(
		printers.NewFieldValue("NAME", b.Name()),
	)
}

func (b *HclResourceImpl) setBaseProperties() {
	if b.Title == nil {
		b.Title = b.getBaseImpl().Title
	}
	if b.Description == nil {
		b.Description = b.getBaseImpl().Description
	}

	b.Tags = utils.MergeMaps(b.Tags, b.getBaseImpl().Tags)

}

func (b *HclResourceImpl) getBaseImpl() *HclResourceImpl {
	return b.base.GetHclResourceImpl()
}
