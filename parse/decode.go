package parse

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

var missingVariableErrors = []string{
	// returned when the context variables does not have top level 'type' node (locals/control/etc)
	"Unknown variable",
	// returned when the variables have the type object but a field has not yet been populated
	"Unsupported attribute",
	"Missing map element",
}

func decode(parseCtx *ModParseContext) hcl.Diagnostics {
	utils.LogTime(fmt.Sprintf("decode %s start", parseCtx.CurrentMod.Name()))
	defer utils.LogTime(fmt.Sprintf("decode %s end", parseCtx.CurrentMod.Name()))

	var diags hcl.Diagnostics

	blocks, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocks {
		if block.Type == schema.BlockTypeLocals {
			resources, res := decodeLocalsBlock(block, parseCtx)
			if !res.Success() {
				diags = append(diags, res.Diags...)
				continue
			}
			for _, resource := range resources {
				resourceDiags := addResourceToMod(resource, block, parseCtx)
				diags = append(diags, resourceDiags...)
			}
		} else {
			resource, res := decodeBlock(block, parseCtx)
			diags = append(diags, res.Diags...)
			if !res.Success() || resource == nil {
				continue
			}

			resourceDiags := addResourceToMod(resource, block, parseCtx)
			diags = append(diags, resourceDiags...)
		}
	}

	return diags
}

func addResourceToMod(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) hcl.Diagnostics {
	if !shouldAddToMod(resource, block, parseCtx) {
		return nil
	}
	return parseCtx.CurrentMod.AddResource(resource)

}

func shouldAddToMod(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) bool {
	switch resource.(type) {
	// do not add mods, withs
	case *modconfig.Mod, *modconfig.DashboardWith:
		return false

	case *modconfig.DashboardCategory, *modconfig.DashboardInput:
		// if this is a dashboard category or dashboard input, only add top level blocks
		// this is to allow nested categories/inputs to have the same name as top level categories
		// (nested inputs are added by Dashboard.InitInputs)
		return parseCtx.IsTopLevelBlock(block)
	default:
		return true
	}
}

// special case decode logic for locals
func decodeLocalsBlock(block *hcl.Block, parseCtx *ModParseContext) ([]modconfig.HclResource, *DecodeResult) {
	var resources []modconfig.HclResource
	var res = NewDecodeResult()

	// check name is valid
	diags := ValidateName(block)
	if diags.HasErrors() {
		res.AddDiags(diags)
		return nil, res
	}

	var locals []*modconfig.Local
	locals, res = decodeLocals(block, parseCtx)
	for _, local := range locals {
		resources = append(resources, local)
		handleModDecodeResult(local, res, block, parseCtx)
	}

	return resources, res
}

func decodeBlock(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	var resource modconfig.HclResource
	var res = NewDecodeResult()

	// has this block already been decoded?
	// (this could happen if it is a child block and has been decoded before its parent as part of second decode phase)
	if resource, ok := parseCtx.GetDecodedResourceForBlock(block); ok {
		return resource, res
	}

	// check name is valid
	diags := ValidateName(block)
	if diags.HasErrors() {
		res.AddDiags(diags)
		return nil, res
	}

	// now do the actual decode
	switch {
	case helpers.StringSliceContains(schema.NodeAndEdgeProviderBlocks, block.Type):
		resource, res = decodeNodeAndEdgeProvider(block, parseCtx)
	case helpers.StringSliceContains(schema.QueryProviderBlocks, block.Type):
		resource, res = decodeQueryProvider(block, parseCtx)
	default:
		switch block.Type {
		case schema.BlockTypeMod:
			// decodeMode has slightly different args as this code is shared with ParseModDefinition
			resource, res = decodeMod(block, parseCtx.EvalCtx, parseCtx.CurrentMod)
		case schema.BlockTypeDashboard:
			resource, res = decodeDashboard(block, parseCtx)
		case schema.BlockTypeContainer:
			resource, res = decodeDashboardContainer(block, parseCtx)
		case schema.BlockTypeVariable:
			resource, res = decodeVariable(block, parseCtx)
		case schema.BlockTypeBenchmark:
			resource, res = decodeBenchmark(block, parseCtx)
		case schema.BlockTypePipeline:
			resource, res = decodePipeline(parseCtx.CurrentMod, block, parseCtx)
		case schema.BlockTypeTrigger:
			resource, res = decodeTrigger(parseCtx.CurrentMod, block, parseCtx)
		default:
			// all other blocks are treated the same:
			resource, res = decodeResource(block, parseCtx)
		}
	}

	// Note that an interface value that holds a nil concrete value is itself non-nil.
	if !helpers.IsNil(resource) {
		// handle the result
		// - if there are dependencies, add to run context
		handleModDecodeResult(resource, res, block, parseCtx)
	}

	return resource, res
}

func decodeMod(block *hcl.Block, evalCtx *hcl.EvalContext, mod *modconfig.Mod) (*modconfig.Mod, *DecodeResult) {
	res := NewDecodeResult()

	// decode the database attribute separately
	// do a partial decode using a schema containing just database - use to pull out all other body content in the remain block
	databaseContent, remain, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: schema.AttributeTypeDatabase},
		}})
	res.HandleDecodeDiags(diags)

	// decode the body
	moreDiags := DecodeHclBody(remain, evalCtx, mod, mod)
	res.HandleDecodeDiags(moreDiags)

	connectionString, searchPath, searchPathPrefix, moreDiags := resolveConnectionString(databaseContent, evalCtx)
	res.HandleDecodeDiags(moreDiags)

	// if connection string or search path was specified (by the mod referencing a connection), set them
	if connectionString != nil {
		mod.Database = connectionString
	}
	if searchPath != nil {
		mod.SearchPath = searchPath
	}
	if searchPathPrefix != nil {
		mod.SearchPathPrefix = searchPathPrefix
	}

	return mod, res

}

func DecodeRequire(block *hcl.Block, evalCtx *hcl.EvalContext) (*modconfig.Require, hcl.Diagnostics) {
	require := modconfig.NewRequire()
	// set ranges
	require.DeclRange = hclhelpers.BlockRange(block)
	require.TypeRange = block.TypeRange
	// decode
	diags := gohcl.DecodeBody(block.Body, evalCtx, require)
	return require, diags
}

// generic decode function for any resource we do not have custom decode logic for
func decodeResource(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := NewDecodeResult()
	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.HandleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	diags = DecodeHclBody(block.Body, parseCtx.EvalCtx, parseCtx, resource)
	if len(diags) > 0 {
		res.HandleDecodeDiags(diags)
	}
	return resource, res
}

// return a shell resource for the given block
func resourceForBlock(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, hcl.Diagnostics) {
	var resource modconfig.HclResource
	// parseCtx already contains the current mod
	mod := parseCtx.CurrentMod
	blockName := parseCtx.DetermineBlockName(block)

	factoryFuncs := map[string]func(*hcl.Block, *modconfig.Mod, string) modconfig.HclResource{
		// for block type mod, just use the current mod
		schema.BlockTypeMod:       func(*hcl.Block, *modconfig.Mod, string) modconfig.HclResource { return mod },
		schema.BlockTypeQuery:     modconfig.NewQuery,
		schema.BlockTypeControl:   modconfig.NewControl,
		schema.BlockTypeBenchmark: modconfig.NewBenchmark,
		schema.BlockTypeDashboard: modconfig.NewDashboard,
		schema.BlockTypeContainer: modconfig.NewDashboardContainer,
		schema.BlockTypeChart:     modconfig.NewDashboardChart,
		schema.BlockTypeCard:      modconfig.NewDashboardCard,
		schema.BlockTypeFlow:      modconfig.NewDashboardFlow,
		schema.BlockTypeGraph:     modconfig.NewDashboardGraph,
		schema.BlockTypeHierarchy: modconfig.NewDashboardHierarchy,
		schema.BlockTypeImage:     modconfig.NewDashboardImage,
		schema.BlockTypeInput:     modconfig.NewDashboardInput,
		schema.BlockTypeTable:     modconfig.NewDashboardTable,
		schema.BlockTypeText:      modconfig.NewDashboardText,
		schema.BlockTypeNode:      modconfig.NewDashboardNode,
		schema.BlockTypeEdge:      modconfig.NewDashboardEdge,
		schema.BlockTypeCategory:  modconfig.NewDashboardCategory,
		schema.BlockTypeWith:      modconfig.NewDashboardWith,
	}

	factoryFunc, ok := factoryFuncs[block.Type]
	if !ok {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("resourceForBlock called for unsupported block type %s", block.Type),
			Subject:  hclhelpers.BlockRangePointer(block),
		},
		}
	}
	resource = factoryFunc(block, mod, blockName)
	return resource, nil
}

func decodeLocals(block *hcl.Block, parseCtx *ModParseContext) ([]*modconfig.Local, *DecodeResult) {
	res := NewDecodeResult()
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		res.Diags = diags
		return nil, res
	}

	// build list of locals
	locals := make([]*modconfig.Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			res.Diags = append(res.Diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				Detail:   badIdentifierDetail,
				Subject:  &attr.NameRange,
			})
			continue
		}
		// try to evaluate expression
		val, diags := attr.Expr.Value(parseCtx.EvalCtx)
		// handle any resulting diags, which may specify dependencies
		res.HandleDecodeDiags(diags)

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range, parseCtx.CurrentMod))
	}
	return locals, res
}

func decodeVariable(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Variable, *DecodeResult) {
	res := NewDecodeResult()

	var variable *modconfig.Variable
	content, diags := block.Body.Content(VariableBlockSchema)
	res.HandleDecodeDiags(diags)

	v, diags := DecodeVariableBlock(block, content, parseCtx)
	res.HandleDecodeDiags(diags)

	if res.Success() {
		variable = modconfig.NewVariable(v, parseCtx.CurrentMod)
	} else {
		slog.Error("decodeVariable failed", "diags", res.Diags)
		return nil, res
	}
	// if a type property was specified, extract type string from the hcl source
	if attr, exists := content.Attributes[schema.AttributeTypeType]; exists {
		src := parseCtx.FileData[attr.Expr.Range().Filename]
		variable.TypeString = extractExpressionString(attr.Expr, src)
	}

	diags = decodeProperty(content, "tags", &variable.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "tags", &variable.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	return variable, res

}

func decodeQueryProvider(block *hcl.Block, parseCtx *ModParseContext) (modconfig.QueryProvider, *DecodeResult) {
	res := NewDecodeResult()
	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.HandleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	// decode the database attribute separately
	// do a partial decode using a schema containing just database - use to pull out all other body content in the remain block
	databaseContent, remain, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: schema.AttributeTypeDatabase},
		}})

	res.HandleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'resource' to populate all properties that can be automatically decoded
	diags = DecodeHclBody(remain, parseCtx.EvalCtx, parseCtx, resource)
	res.HandleDecodeDiags(diags)

	// decode 'with',args and params blocks
	res.Merge(decodeQueryProviderBlocks(block, remain.(*hclsyntax.Body), resource, parseCtx))

	// resolve the connection string and (if set) search path
	qp := resource.(modconfig.QueryProvider)
	connectionString, searchPath, searchPathPrefix, diags := resolveConnectionString(databaseContent, parseCtx.EvalCtx)
	if connectionString != nil {
		qp.SetDatabase(connectionString)
	}
	if searchPath != nil {
		qp.SetSearchPath(searchPath)
	}
	if searchPathPrefix != nil {
		qp.SetSearchPathPrefix(searchPathPrefix)
	}
	res.HandleDecodeDiags(diags)

	return qp, res
}

func resolveConnectionString(content *hcl.BodyContent, evalCtx *hcl.EvalContext) (cs *string, searchPath, searchPathPrefix []string, diags hcl.Diagnostics) {
	var connectionString string
	attr, exists := content.Attributes[schema.AttributeTypeDatabase]
	if !exists {
		return nil, searchPath, searchPathPrefix, diags
	}

	var dbValue cty.Value
	diags = gohcl.DecodeExpression(attr.Expr, evalCtx, &dbValue)

	if diags.HasErrors() {
		// use decode result to handle any dependencies
		res := NewDecodeResult()
		res.HandleDecodeDiags(diags)
		diags = res.Diags
		// if there are other errors, return them
		if diags.HasErrors() {
			return nil, searchPath, searchPathPrefix, res.Diags
		}
		// so there is a dependency error - if it is for a connection, return the connection name as the connection string
		for _, dep := range res.Depends {
			for _, traversal := range dep.Traversals {
				depName := hclhelpers.TraversalAsString(traversal)
				if strings.HasPrefix(depName, "connection.") {
					return &depName, searchPath, searchPathPrefix, diags
				}
			}
		}
		// if we get here, there is a dependency error but it is not for a connection
		// return the original diags for the calling code to handle
		return nil, searchPath, searchPathPrefix, diags
	}
	// check if this is a connection string or a connection
	if dbValue.Type() == cty.String {
		connectionString = dbValue.AsString()
	} else {
		// if this is a temporary connection, ignore (this will only occur during the variable parsing phase)
		if dbValue.Type().HasAttribute("temporary") {
			return nil, searchPath, searchPathPrefix, diags
		}

		c, err := app_specific_connection.CtyValueToConnection(dbValue)
		if err != nil {
			return nil, searchPath, searchPathPrefix, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  err.Error(),
					Subject:  attr.Range.Ptr(),
				}}
		}

		// the connection type must support connection strings
		if conn, ok := c.(connection.ConnectionStringProvider); ok {
			connectionString = conn.GetConnectionString()
		} else {
			slog.Warn("connection does not support connection string", "db", c)
			return nil, searchPath, searchPathPrefix, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "invalid connection reference - only connections which implement GetConnectionString() are supported",
				}}
		}
		if conn, ok := c.(connection.SearchPathProvider); ok {
			searchPath = conn.GetSearchPath()
			searchPathPrefix = conn.GetSearchPathPrefix()
		}
	}

	return &connectionString, searchPath, searchPathPrefix, diags
}

func decodeQueryProviderBlocks(block *hcl.Block, content *hclsyntax.Body, resource modconfig.HclResource, parseCtx *ModParseContext) *DecodeResult {
	var diags hcl.Diagnostics
	res := NewDecodeResult()
	queryProvider, ok := resource.(modconfig.QueryProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a QueryProvider", block.Type))
	}

	if attr, exists := content.Attributes[schema.AttributeTypeArgs]; exists {
		args, runtimeDependencies, diags := decodeArgs(attr.AsHCLAttribute(), parseCtx.EvalCtx, queryProvider)
		if diags.HasErrors() {
			// handle dependencies
			res.HandleDecodeDiags(diags)
		} else {
			queryProvider.SetArgs(args)
			queryProvider.AddRuntimeDependencies(runtimeDependencies)
		}
	}

	var params []*modconfig.ParamDef
	for _, b := range content.Blocks {
		block = b.AsHCLBlock()
		switch block.Type {
		case schema.BlockTypeParam:
			paramDef, runtimeDependencies, moreDiags := decodeParam(block, parseCtx)
			if !moreDiags.HasErrors() {
				params = append(params, paramDef)
				queryProvider.AddRuntimeDependencies(runtimeDependencies)
				// add and references contained in the param block to the control refs
				moreDiags = AddReferences(resource, block, parseCtx)
			}
			diags = append(diags, moreDiags...)
		}
	}

	queryProvider.SetParams(params)
	res.HandleDecodeDiags(diags)
	return res
}

func decodeNodeAndEdgeProvider(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := NewDecodeResult()

	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.HandleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	nodeAndEdgeProvider, ok := resource.(modconfig.NodeAndEdgeProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a NodeAndEdgeProvider", block.Type))
	}

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.HandleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'resource' to populate all properties that can be automatically decoded
	diags = DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	// decode sql args and params
	res.Merge(decodeQueryProviderBlocks(block, body, resource, parseCtx))

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeNodeAndEdgeProviderBlocks(body, nodeAndEdgeProvider, parseCtx)
		res.Merge(blocksRes)
	}

	return resource, res
}

func decodeNodeAndEdgeProviderBlocks(content *hclsyntax.Body, nodeAndEdgeProvider modconfig.NodeAndEdgeProvider, parseCtx *ModParseContext) *DecodeResult {
	var res = NewDecodeResult()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		switch block.Type {
		case schema.BlockTypeCategory:
			// decode block
			category, blockRes := decodeBlock(block, parseCtx)
			res.Merge(blockRes)
			if !blockRes.Success() {
				continue
			}

			// add the category to the nodeAndEdgeProvider
			res.AddDiags(nodeAndEdgeProvider.AddCategory(category.(*modconfig.DashboardCategory)))

			// DO NOT add the category to the mod

		case schema.BlockTypeNode, schema.BlockTypeEdge:
			child, childRes := decodeQueryProvider(block, parseCtx)

			// TACTICAL if child has any runtime dependencies, claim them
			// this is to ensure if this resource is used as base, we can be correctly identified
			// as the publisher of the runtime dependencies
			for _, r := range child.GetRuntimeDependencies() {
				r.Provider = nodeAndEdgeProvider
			}

			// populate metadata, set references and call OnDecoded
			handleModDecodeResult(child, childRes, block, parseCtx)
			res.Merge(childRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddChild(child)
				res.AddDiags(moreDiags)
			}
		case schema.BlockTypeWith:
			with, withRes := decodeBlock(block, parseCtx)
			res.Merge(withRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddWith(with.(*modconfig.DashboardWith))
				res.AddDiags(moreDiags)
			}
		}

	}

	return res
}

func decodeDashboard(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Dashboard, *DecodeResult) {
	res := NewDecodeResult()
	dashboard := modconfig.NewDashboard(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.Dashboard)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.HandleDecodeDiags(diags)

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, dashboard)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	if dashboard.Base != nil && len(dashboard.Base.ChildNames) > 0 {
		supportedChildren := []string{schema.BlockTypeContainer, schema.BlockTypeChart, schema.BlockTypeCard, schema.BlockTypeFlow, schema.BlockTypeGraph, schema.BlockTypeHierarchy, schema.BlockTypeImage, schema.BlockTypeInput, schema.BlockTypeTable, schema.BlockTypeText}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(dashboard.Base.ChildNames, block, supportedChildren, parseCtx)
		dashboard.Base.SetChildren(children)
	}
	if !res.Success() {
		return dashboard, res
	}

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeDashboardBlocks(body, dashboard, parseCtx)
		res.Merge(blocksRes)
	}

	return dashboard, res
}

func decodeDashboardBlocks(content *hclsyntax.Body, dashboard *modconfig.Dashboard, parseCtx *ModParseContext) *DecodeResult {
	var res = NewDecodeResult()
	// set dashboard as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboard)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()

		// decode block
		resource, blockRes := decodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// we expect either inputs or child report nodes
		// add the resource to the mod
		res.AddDiags(addResourceToMod(resource, block, parseCtx))
		// add to the dashboard children
		// (we expect this cast to always succeed)
		if child, ok := resource.(modconfig.ModTreeItem); ok {
			res.AddDiags(dashboard.AddChild(child))
		}

	}

	moreDiags := dashboard.InitInputs()
	res.AddDiags(moreDiags)

	return res
}

func decodeDashboardContainer(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.DashboardContainer, *DecodeResult) {
	res := NewDecodeResult()
	container := modconfig.NewDashboardContainer(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.DashboardContainer)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.HandleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, container)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeDashboardContainerBlocks(body, container, parseCtx)
		res.Merge(blocksRes)
	}

	return container, res
}

func decodeDashboardContainerBlocks(content *hclsyntax.Body, dashboardContainer *modconfig.DashboardContainer, parseCtx *ModParseContext) *DecodeResult {
	var res = NewDecodeResult()

	// set container as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboardContainer)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		resource, blockRes := decodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// special handling for inputs
		if b.Type == schema.BlockTypeInput {
			input := resource.(*modconfig.DashboardInput)
			dashboardContainer.Inputs = append(dashboardContainer.Inputs, input)
			dashboardContainer.AddChild(input)
			// the input will be added to the mod by the parent dashboard

		} else {
			// for all other children, add to mod and children
			res.AddDiags(addResourceToMod(resource, block, parseCtx))
			if child, ok := resource.(modconfig.ModTreeItem); ok {
				dashboardContainer.AddChild(child)
			}
		}
	}

	return res
}

func decodeBenchmark(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Benchmark, *DecodeResult) {
	res := NewDecodeResult()
	benchmark := modconfig.NewBenchmark(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.Benchmark)
	content, diags := block.Body.Content(BenchmarkBlockSchema)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "children", &benchmark.ChildNames, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "description", &benchmark.Description, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "documentation", &benchmark.Documentation, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "tags", &benchmark.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "title", &benchmark.Title, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "type", &benchmark.Type, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = decodeProperty(content, "display", &benchmark.Display, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	// now add children
	if res.Success() {
		supportedChildren := []string{schema.BlockTypeBenchmark, schema.BlockTypeControl}
		children, diags := resolveChildrenFromNames(benchmark.ChildNames.StringList(), block, supportedChildren, parseCtx)
		res.HandleDecodeDiags(diags)

		// now set children and child name strings
		benchmark.SetChildren(children)
		benchmark.ChildNameStrings = getChildNameStringsFromModTreeItem(children)
	}

	diags = decodeProperty(content, "base", &benchmark.Base, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)
	if benchmark.Base != nil && len(benchmark.Base.ChildNames) > 0 {
		supportedChildren := []string{schema.BlockTypeBenchmark, schema.BlockTypeControl}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(benchmark.Base.ChildNameStrings, block, supportedChildren, parseCtx)
		benchmark.Base.SetChildren(children)
	}
	diags = decodeProperty(content, "width", &benchmark.Width, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)
	return benchmark, res
}

func decodeProperty(content *hcl.BodyContent, property string, dest interface{}, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if attr, ok := content.Attributes[property]; ok {
		diags = gohcl.DecodeExpression(attr.Expr, evalCtx, dest)
	}
	return diags
}

// handleModDecodeResult
// if decode was successful:
// - generate and set resource metadata
// - add resource to ModParseContext (which adds it to the mod)handleModDecodeResult
func handleModDecodeResult(resource modconfig.HclResource, res *DecodeResult, block *hcl.Block, parseCtx *ModParseContext) {
	if !res.Success() {
		if len(res.Depends) > 0 {
			moreDiags := parseCtx.AddDependencies(block, resource.GetUnqualifiedName(), res.Depends)
			res.AddDiags(moreDiags)
		}
		return
	}
	// set whether this is a top level resource
	resource.SetTopLevel(parseCtx.IsTopLevelBlock(block))

	// call post decode hook
	// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
	moreDiags := resource.OnDecoded(block, parseCtx)
	res.AddDiags(moreDiags)

	// add references
	moreDiags = AddReferences(resource, block, parseCtx)
	res.AddDiags(moreDiags)

	// validate the resource
	moreDiags = validateResource(resource)
	res.AddDiags(moreDiags)
	// if we failed validation, return
	if !res.Success() {
		return
	}

	// if resource is NOT anonymous, and this is a TOP LEVEL BLOCK, add into the run context
	// NOTE: we can only reference resources defined in a top level block
	if !resourceIsAnonymous(resource) && resource.IsTopLevel() {
		moreDiags = parseCtx.AddResource(resource)
		res.AddDiags(moreDiags)
	}

	// if resource supports metadata, save it
	if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
		moreDiags = AddResourceMetadata(resourceWithMetadata, resource.GetHclResourceImpl().DeclRange, parseCtx)
		res.AddDiags(moreDiags)
	}
}

func resourceIsAnonymous(resource modconfig.HclResource) bool {
	// (if a resource anonymous it must support ResourceWithMetadata)
	resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata)
	anonymousResource := ok && resourceWithMetadata.IsAnonymous()
	return anonymousResource
}

func AddResourceMetadata(resourceWithMetadata modconfig.ResourceWithMetadata, srcRange hcl.Range, parseCtx *ModParseContext) hcl.Diagnostics {
	metadata, err := GetMetadataForParsedResource(resourceWithMetadata.Name(), srcRange, parseCtx.FileData, parseCtx.CurrentMod)
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &srcRange,
		}}
	}
	//  set on resource
	resourceWithMetadata.SetMetadata(metadata)
	return nil
}

func ValidateName(block *hcl.Block) hcl.Diagnostics {
	if len(block.Labels) == 0 {
		return nil
	}

	if !hclsyntax.ValidIdentifier(block.Labels[0]) {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		}}
	}
	return nil
}

// Validate all blocks and attributes are supported
// We use partial decoding so that we can automatically decode as many properties as possible
// and only manually decode properties requiring special logic.
// The problem is the partial decode does not return errors for invalid attributes/blocks, so we must implement our own
func validateHcl(blockType string, body *hclsyntax.Body, schema *hcl.BodySchema) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// identify any blocks specified by hcl tags
	var supportedBlocks = make(map[string]struct{})
	var supportedAttributes = make(map[string]struct{})
	for _, b := range schema.Blocks {
		supportedBlocks[b.Type] = struct{}{}
	}
	for _, b := range schema.Attributes {
		supportedAttributes[b.Name] = struct{}{}
	}

	// now check for invalid blocks
	for _, block := range body.Blocks {
		if _, ok := supportedBlocks[block.Type]; !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf(`Unsupported block type: Blocks of type '%s' are not expected here.`, block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}
	for _, attribute := range body.Attributes {
		if _, ok := supportedAttributes[attribute.Name]; !ok {
			// special case code for deprecated properties
			subject := attribute.Range()
			if isDeprecated(attribute, blockType) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf(`Deprecated attribute: '%s' is deprecated for '%s' blocks and will be ignored.`, attribute.Name, blockType),
					Subject:  &subject,
				})
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf(`Unsupported attribute: '%s' not expected here.`, attribute.Name),
					Subject:  &subject,
				})
			}
		}
	}

	return diags
}

func isDeprecated(attribute *hclsyntax.Attribute, blockType string) bool {
	switch attribute.Name {
	case "search_path", "search_path_prefix":
		return blockType == schema.BlockTypeQuery || blockType == schema.BlockTypeControl
	default:
		return false
	}
}
