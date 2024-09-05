package modconfig

import (
	"fmt"

	typehelpers "github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig/var_config"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// TODO check DescriptionSet - still required?

// Variable is a struct representing a Variable resource
type Variable struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Default cty.Value         `column:"default_value,jsonb" json:"-"`
	Type    cty.Type          `column:"var_type,string" json:"-"`
	Tags    map[string]string `column:"tags,jsonb" cty:"tags" hcl:"tags,optional" json:"tags,omitempty"`
	Enum    cty.Value         `json:"-"`

	// TypeString (json: type) is currently showing the HCL Type: string or list(string)
	// current type = list(list(string))
	//
	// future type = ["list",["list","string"]]
	TypeString string `json:"type"`

	// This is the strategic field, where the HCL string representation of cty.Type is stored
	// for example: list(list(string))
	TypeHclString string `json:"type_string"`

	DefaultGo any    `json:"value_default"`
	ValueGo   any    `json:"value"`
	EnumGo    []any  `json:"enum"`
	ModName   string `json:"mod_name"`

	// set after value resolution `column:"value,jsonb"`
	Value                      cty.Value                      `column:"value,jsonb" json:"-"`
	ValueSourceType            string                         `column:"value_source,string" json:"-"`
	ValueSourceFileName        string                         `column:"value_source_file_name,string" json:"-"`
	ValueSourceStartLineNumber int                            `column:"value_source_start_line_number,integer" json:"-"`
	ValueSourceEndLineNumber   int                            `column:"value_source_end_line_number,integer" json:"-"`
	ParsingMode                var_config.VariableParsingMode `json:"-"`
}

func NewVariable(v *var_config.Variable, mod *Mod) *Variable {
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
				blockType:       schema.BlockTypeVariable,
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

		TypeHclString: hclhelpers.CtyTypeToHclType(v.Type, v.Default.Type()), // strategic, this where the HCL string representation of cty.Type is stored
	}

	// deprecated, we will change this "type" to cty.Type's json serialisation later
	res.TypeString = res.TypeHclString

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
func (v *Variable) OnDecoded(block *hcl.Block, _ ResourceMapsProvider) hcl.Diagnostics {
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
		v.TypeHclString = hclhelpers.CtyTypeToHclType(value.Type())
		v.TypeString = v.TypeHclString
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

func (v *Variable) Diff(other *Variable) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: v,
		Name: v.Name(),
	}

	if !utils.SafeStringsEqual(v.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(v.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(v, other)
	return res
}

// CtyValue implements CtyValueProvider
func (v *Variable) CtyValue() (cty.Value, error) {
	return GetCtyValue(v)
}
