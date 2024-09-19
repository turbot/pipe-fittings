package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"log/slog"
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
		// does this support a connection string
		type connectionStringProvider interface {
			GetConnectionString() string
		}
		if conn, ok := c.(connectionStringProvider); ok {
			results[schema.AttributeTypeDatabase] = utils.ToStringPointer(conn.GetConnectionString())
		} else {
			slog.Warn("connection does not support connection string", "db", c)
		}

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
		case schema.AttributeTypeSql:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}
		case schema.AttributeTypeDatabase:
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
				slog.Warn("db", "db", goVals)
				//p.Database = goVals
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
	shortName := value.GetAttr("short_name").AsString()
	connectionType := value.GetAttr("type").AsString()
	var declRange hclhelpers.Range
	err = gocty.FromCtyValue(value.GetAttr("decl_range"), &declRange)
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}

	// now instantiate an empty connection of the correct type
	conn, err := app_specific_connection.InstantiateConnection(connectionType, shortName, declRange.HclRange())
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}

	// now decode the cty value into the connection

	// TODO why is env in cty?
	// we already decoded the bvase fields, so remove from the value
	baseFields := []string{"type", "short_name", "decl_range", "env"}
	originalMap := value.AsValueMap()
	// Remove the base fields that belong to the nested struct (ConnectionImpl)
	// create new cty values, one for the base
	for _, field := range baseFields {
		delete(originalMap, field)
	}
	connectionValue := cty.ObjectVal(originalMap)

	err = gocty.FromCtyValue(connectionValue, &conn)
	if err != nil {
		return nil, perr.BadRequestWithMessage("unable to decode connection: " + err.Error())
	}

	return conn, nil
}
