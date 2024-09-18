package modconfig

import (
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type PipelineStepQuery struct {
	PipelineStepBase
	Database         *string
	ConnectionString *string       `json:"database"`
	Sql              *string       `json:"sql"`
	Args             []interface{} `json:"args"`
}

func (p *PipelineStepQuery) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || p != nil && helpers.IsNil(iOther) {
		return false
	}

	other, ok := iOther.(*PipelineStepQuery)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if len(p.Args) != len(other.Args) {
		return false
	}
	for i := range p.Args {
		if p.Args[i] != other.Args[i] {
			return false
		}
	}

	return utils.PtrEqual(p.Database, other.Database) &&
		utils.PtrEqual(p.Sql, other.Sql)
}

func (p *PipelineStepQuery) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var diags hcl.Diagnostics
	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	// sql
	results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeSql, p.Sql)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}

	// database
	if databaseExpression, ok := p.UnresolvedAttributes[schema.AttributeTypeDatabase]; ok {
		// attribute needs resolving, this case may happen if we specify the entire option as an attribute
		var connValue cty.Value
		diags := gohcl.DecodeExpression(databaseExpression, evalContext, &connValue)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}
		c, err := CtyValueToConnection(connValue)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to resolve connection attribute: " + err.Error())
		}
		results[schema.AttributeTypeDatabase] = c

	} else {
		// database
		results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeDatabase, p.Database)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}
	}

	//// attribute needs resolving, this case may happen if we specify the entire option as an attribute
	//var opts cty.Value
	//diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeOptions], evalContext, &opts)
	//if diags.HasErrors() {
	//	return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	//}
	//resolvedOpts, err = CtyValueToPipelineStepInputOptionList(opts)
	//if err != nil {
	//	return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse options attribute: " + err.Error())
	//}
	//results, diags = simpleTypeInputFromAttribute(p.GetUnresolvedAttributes(), results, evalContext, schema.AttributeTypeDatabase, p.Database)
	//if diags.HasErrors() {
	//	return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	//}

	if _, ok := results[schema.AttributeTypeDatabase]; !ok {
		return nil, perr.BadRequestWithMessage(p.Name + ": database must be supplied")
	}

	if p.UnresolvedAttributes[schema.AttributeTypeArgs] != nil {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeArgs], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}

		mapValue, err := hclhelpers.CtyToGoInterfaceSlice(args)
		if err != nil {
			return nil, perr.BadRequestWithMessage(p.Name + ": unable to parse args attribute to an array " + err.Error())
		}
		results[schema.AttributeTypeArgs] = mapValue

	} else if p.Args != nil {
		results[schema.AttributeTypeArgs] = p.Args
	}

	return results, nil
}

func (p *PipelineStepQuery) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSql, schema.AttributeTypeDatabase:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeArgs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVals, err2 := hclhelpers.CtyToGoInterfaceSlice(val)
				if err2 != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeArgs + "' attribute to Go values",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Args = goVals
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Query Step '" + attr.Name + "'",
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepQuery) Validate() hcl.Diagnostics {
	// validate the base attributes
	diags := p.ValidateBaseAttributes()
	return diags
}

func CtyValueToConnection(value cty.Value) (_ connection.PipelingConnection, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = perr.BadRequestWithMessage("unable to decode connection: " + r.(string))
		}
	}()
	// get the name and extract the block type
	name := value.GetAttr("name")

	parts := strings.Split(name.AsString(), ".")
	if len(parts) != 2 {
		return nil, perr.BadRequestWithMessage("connection name must be in the format <type>.<name>")
	}
	connectionType := parts[0]
	shortName := parts[1]

	// now instantiate an empty connection of the correct type
	// TODO KAI can we get range from cty
	conn, err := app_specific_connection.InstantiateConnection(connectionType, shortName, hcl.Range{})
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}
	// now decode the cty value into the connection
	err = gocty.FromCtyValue(value, &conn)
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}
	return conn, nil
}

//	var output []PipelineStepInputOption
//
//	opts := value.AsValueSlice()
//
//	for _, opt := range opts {
//		valueMap := opt.AsValueMap()
//
//		isValid := false
//		option := PipelineStepInputOption{
//			UnresolvedAttributes: make(map[string]hcl.Expression),
//		}
//
//		for k, v := range valueMap {
//			switch k {
//			case schema.AttributeTypeValue:
//				if !v.IsNull() {
//					isValid = true
//					val := v.AsString()
//					option.Value = &val
//				}
//			case schema.AttributeTypeLabel:
//				if !v.IsNull() {
//					label := v.AsString()
//					option.Label = &label
//				}
//			case schema.AttributeTypeSelected:
//				if !v.IsNull() && v.Type() == cty.Bool {
//					isSelected := v.True()
//					option.Selected = &isSelected
//				}
//			case schema.AttributeTypeStyle:
//				if !v.IsNull() {
//					s := v.AsString()
//					option.Style = &s
//				}
//			default:
//				return nil, perr.BadRequestWithMessage(k + " is not a valid attribute for input options")
//			}
//		}
//
//		if isValid {
//			output = append(output, option)
//		} else {
//			return nil, perr.BadRequestWithMessage("input options must declare a value")
//		}
//	}
//
//	return output, nil
//}
