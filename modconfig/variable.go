package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Variable is a struct representing a Variable resource
type Variable struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Default cty.Value         ` json:"-"`
	Type    cty.Type          ` json:"-"`
	Tags    map[string]string `cty:"tags" hcl:"tags,optional" json:"tags,omitempty"`
	Enum    cty.Value         `json:"-"`

	// TypeString shows the type as specified in the hcl file
	// or if no type is specified, it derives a type string from the variable value
	TypeString    string         `json:"type_string"`
	DefaultGo     any            `json:"value_default"`
	ValueGo       any            `json:"value"`
	EnumGo        []any          `json:"enum,omitempty"`
	ModName       string         `json:"mod_name"`
	Subtype       hcl.Expression `json:"-"`
	SubtypeString string         `json:"subtype_string,omitempty"`

	// set after value resolution `column:"value,jsonb"`
	Value                      cty.Value           `json:"-"`
	ValueSourceType            string              `json:"-"`
	ValueSourceFileName        string              `json:"-"`
	ValueSourceStartLineNumber int                 `json:"-"`
	ValueSourceEndLineNumber   int                 `json:"-"`
	ParsingMode                VariableParsingMode `json:"-"`
	Format                     string              `json:"-"`
}

func NewVariable(v *RawVariable, mod *Mod) *Variable {
	var defaultGo interface{} = nil
	if !v.Default.IsNull() {
		defaultGo, _ = hclhelpers.CtyToGo(v.Default)
	}
	fullName := fmt.Sprintf("%s.var.%s", mod.ShortName, v.Name)
	res := &Variable{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       v.Name,
				Description:     &v.Description,
				FullName:        fullName,
				DeclRange:       v.DeclRange,
				UnqualifiedName: fmt.Sprintf("var.%s", v.Name),
				BlockType:       schema.BlockTypeVariable,
			},
			Mod: mod,
		},
		Default:   v.Default,
		DefaultGo: defaultGo,
		// initialise the value to the default - may be set later
		Value:   v.Default,
		ValueGo: defaultGo,

		Type:        v.Type,
		ParsingMode: v.ParsingMode,
		ModName:     mod.ShortName,
		Enum:        v.Enum,
		EnumGo:      v.EnumGo,
		Format:      v.Format,
	}

	if v.Title != "" {
		res.Title = &v.Title
	}

	// if no type is set and a default _is_ set, use default to set the type
	if res.Type.Equals(cty.DynamicPseudoType) && !res.Default.IsNull() {
		res.Type = res.Default.Type()
	}
	return res
}

func (v *Variable) Equals(other *Variable) bool {

	if v.Enum.Equals(other.Enum) == cty.False {
		return false
	}

	return v.ShortName == other.ShortName &&
		v.FullName == other.FullName &&
		typehelpers.SafeString(v.Description) == typehelpers.SafeString(other.Description) &&
		v.Default.RawEquals(other.Default) &&
		v.Value.RawEquals(other.Value)
}

// OnDecoded implements HclResource
func (v *Variable) OnDecoded(block *hcl.Block, _ ModResourcesProvider) hcl.Diagnostics {
	return nil
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

func (v *Variable) SetInputValue(value cty.Value, sourceType string, sourceRange tfdiags.SourceRange) error {
	// if the value type is a tuple with no elem type, and we have a type, set the variable to have our type
	if value.Type().Equals(cty.Tuple(nil)) && !v.Type.Equals(cty.DynamicPseudoType) {
		var err error
		value, err = convert.Convert(value, v.Type)
		if err != nil {
			return err
		}
	}

	v.Value = value
	v.ValueSourceType = sourceType
	v.ValueSourceFileName = sourceRange.Filename
	v.ValueSourceStartLineNumber = sourceRange.Start.Line
	v.ValueSourceEndLineNumber = sourceRange.End.Line
	v.ValueGo, _ = hclhelpers.CtyToGo(value)

	// if type string is not set, derive from the type of value
	if v.TypeString == "" {
		v.TypeString = hclhelpers.CtyTypeToHclType(value.Type())
	}

	if v.Enum != cty.NilVal {
		// check that the value is in the enum
		valid, err := hclhelpers.ValidateSettingWithEnum(v.Value, v.Enum)
		if err != nil {
			return err
		}
		if !valid {
			return perr.BadRequestWithMessage(fmt.Sprintf("value %s not in enum", v.ValueGo))
		}
	}

	return nil
}

func (v *Variable) Diff(other *Variable) *ModTreeItemDiffs {
	res := &ModTreeItemDiffs{
		Item: v,
		Name: v.Name(),
	}

	if !utils.SafeStringsEqual(v.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(v.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.PopulateChildDiffs(v, other)
	return res
}

// CtyValue implements CtyValueProvider
func (v *Variable) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(v)
}

// IsLateBinding returns true if the variable has a type which is late binding, i.e. the value is resolved at run time
// rather than at parse time.
// These variables are not added to the eval context, but instead are resolved at execution time
func (v *Variable) IsLateBinding() bool {
	return IsLateBindingType(v.Type)
}

func (p *Variable) IsConnectionType() bool {
	return IsConnectionType(p.Type)
}
