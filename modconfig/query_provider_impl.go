package modconfig

import (
	"fmt"
	"github.com/turbot/pipe-fittings/printers"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type QueryProviderImpl struct {
	RuntimeDependencyProviderImpl
	QueryProviderRemain hcl.Body `hcl:",remain" json:"-"`

	SQL       *string     `cty:"sql" hcl:"sql" column:"sql,string" json:"sql,omitempty"`
	Query     *Query      `cty:"query" hcl:"query" json:"-"`
	Args      *QueryArgs  `cty:"args" column:"args,jsonb" json:"args,omitempty"`
	Params    []*ParamDef `cty:"params" column:"params,jsonb" json:"params,omitempty"`
	QueryName *string     `column:"query,string" json:"query,omitempty"`

	//nolint:unused // TODO: unused function
	withs               []*DashboardWith
	disableCtySerialise bool
	// flags to indicate if params and args were inherited from base resource
	argsInheritedFromBase   bool
	paramsInheritedFromBase bool
}

func NewQueryProviderImpl(block *hcl.Block, mod *Mod, shortName string) QueryProviderImpl {
	return QueryProviderImpl{
		RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
			ModTreeItemImpl: NewModTreeItemImpl(block, mod, shortName),
		},
	}
}

// GetParams implements QueryProvider
func (q *QueryProviderImpl) GetParams() []*ParamDef {
	return q.Params
}

// GetArgs implements QueryProvider
func (q *QueryProviderImpl) GetArgs() *QueryArgs {
	return q.Args

}

// GetSQL implements QueryProvider
func (q *QueryProviderImpl) GetSQL() *string {
	return q.SQL
}

// GetQuery implements QueryProvider
func (q *QueryProviderImpl) GetQuery() *Query {
	return q.Query
}

// SetArgs implements QueryProvider
func (q *QueryProviderImpl) SetArgs(args *QueryArgs) {
	q.Args = args
}

// SetParams implements QueryProvider
func (q *QueryProviderImpl) SetParams(params []*ParamDef) {
	q.Params = params
}

// ValidateQuery implements QueryProvider
// returns an error if neither sql or query are set
// it is overridden by resource types for which sql is optional
func (q *QueryProviderImpl) ValidateQuery() hcl.Diagnostics {
	var diags hcl.Diagnostics
	// Top level resources (with the exceptions of controls and queries) are never executed directly,
	// only used as base for a nested resource.
	// Therefore only nested resources, controls and queries MUST have sql or a query defined
	queryRequired := !q.IsTopLevel() ||
		helpers.StringSliceContains([]string{schema.BlockTypeQuery, schema.BlockTypeControl}, q.BlockType())

	if !queryRequired {
		return nil
	}

	if queryRequired && q.Query == nil && q.SQL == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s does not define a query or SQL", q.Name()),
			Subject:  q.GetDeclRange(),
		})
	}
	return diags
}

// RequiresExecution implements QueryProvider
func (q *QueryProviderImpl) RequiresExecution(queryProvider QueryProvider) bool {
	return queryProvider.GetQuery() != nil || queryProvider.GetSQL() != nil
}

// GetResolvedQuery return the SQL and args to run the query
func (q *QueryProviderImpl) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	argsArray, err := ResolveArgs(q, runtimeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve args for %s: %s", q.Name(), err.Error())
	}
	sql := typehelpers.SafeString(q.GetSQL())
	// we expect there to be sql on the query provider, NOT a Query
	if sql == "" {
		return nil, fmt.Errorf("getResolvedQuery faiuled - no sql set for '%s'", q.Name())
	}

	return &ResolvedQuery{
		Name:       q.Name(),
		ExecuteSQL: sql,
		RawSQL:     sql,
		Args:       argsArray,
	}, nil
}

// MergeParentArgs merges our args with our parent args (ours take precedence)
func (q *QueryProviderImpl) MergeParentArgs(queryProvider QueryProvider, parent QueryProvider) (diags hcl.Diagnostics) {
	parentArgs := parent.GetArgs()
	if parentArgs == nil {
		return nil
	}

	args, err := parentArgs.Merge(queryProvider.GetArgs(), parent)
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  parent.(HclResource).GetDeclRange(),
		}}
	}

	queryProvider.SetArgs(args)
	return nil
}

// GetQueryProviderImpl implements QueryProvider
func (q *QueryProviderImpl) GetQueryProviderImpl() *QueryProviderImpl {
	return q
}

// ParamsInheritedFromBase implements QueryProvider
// determine whether our params were inherited from base resource
func (q *QueryProviderImpl) ParamsInheritedFromBase() bool {
	return q.paramsInheritedFromBase
}

// ArgsInheritedFromBase implements QueryProvider
// determine whether our args were inherited from base resource
func (q *QueryProviderImpl) ArgsInheritedFromBase() bool {
	return q.argsInheritedFromBase
}

// CtyValue implements CtyValueProvider
func (q *QueryProviderImpl) CtyValue() (cty.Value, error) {
	if q.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(q)
}

func (q *QueryProviderImpl) setBaseProperties() {
	q.RuntimeDependencyProviderImpl.setBaseProperties()
	if q.SQL == nil {
		q.SQL = q.getBaseImpl().SQL
	}
	if q.Query == nil {
		q.Query = q.getBaseImpl().Query
	}
	if q.Args == nil {
		q.Args = q.getBaseImpl().Args
		q.argsInheritedFromBase = true
	}
	if q.Params == nil {
		q.Params = q.getBaseImpl().Params
		q.paramsInheritedFromBase = true
	}
}

func (q *QueryProviderImpl) getBaseImpl() *QueryProviderImpl {
	return q.base.(QueryProvider).GetQueryProviderImpl()
}

func (q *QueryProviderImpl) OnDecoded(block *hcl.Block, _ ResourceMapsProvider) hcl.Diagnostics {
	q.populateQueryName()

	return nil
}

func (q *QueryProviderImpl) populateQueryName() {
	if q.Query != nil {
		q.QueryName = &q.Query.FullName
	}
}

// GetShowData implements printers.Showable
func (q *QueryProviderImpl) GetShowData() *printers.RowData {

	res := printers.NewRowData(
		printers.FieldValue{Name: "SQL", Value: q.SQL},
		printers.FieldValue{Name: "Query", Value: q.Query},
		printers.FieldValue{Name: "Args", Value: q.Args},
		printers.FieldValue{Name: "Params", Value: q.Params},
	)
	res.Merge(q.RuntimeDependencyProviderImpl.GetShowData())
	return res
}
