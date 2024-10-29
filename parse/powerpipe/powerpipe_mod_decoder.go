package flowpipe

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/modconfig/powerpipe"
	"github.com/turbot/pipe-fittings/parse"

	"github.com/turbot/pipe-fittings/schema"
)

type PowerpipeModDecoder struct {
	parse.DecoderImpl
}

func NewPowerpipeModDecoder(opts ...parse.DecoderOption) parse.Decoder {
	d := &PowerpipeModDecoder{
		DecoderImpl: parse.NewDecoderImpl(),
	}
	for _, block := range schema.NodeAndEdgeProviderBlocks {
		d.DecodeFuncs[block] = d.decodeNodeAndEdgeProvider
	}
	for _, blockType := range schema.QueryProviderBlocks {
		d.DecodeFuncs[blockType] = d.decodeQueryProvider
	}
	d.DecodeFuncs[schema.BlockTypeDashboard] = d.decodeDashboard
	d.DecodeFuncs[schema.BlockTypeContainer] = d.decodeDashboardContainer
	d.DecodeFuncs[schema.BlockTypeBenchmark] = d.decodeBenchmark
	// apply options
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *PowerpipeModDecoder) decodeNodeAndEdgeProvider(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {
	res := parse.NewDecodeResult()

	// get shell resource
	resource, diags := parse.ResourceForBlock(block, parseCtx)
	res.HandleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	nodeAndEdgeProvider, ok := resource.(powerpipe.NodeAndEdgeProvider)
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
	diags = parse.DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	// decode sql args and params
	res.Merge(d.decodeQueryProviderBlocks(block, body, resource, parseCtx))

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := d.decodeNodeAndEdgeProviderBlocks(body, nodeAndEdgeProvider, parseCtx)
		res.Merge(blocksRes)
	}

	return resource, res
}

func (d *PowerpipeModDecoder) decodeNodeAndEdgeProviderBlocks(content *hclsyntax.Body, nodeAndEdgeProvider powerpipe.NodeAndEdgeProvider, parseCtx *parse.ModParseContext) *parse.DecodeResult {
	var res = parse.NewDecodeResult()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		switch block.Type {
		case schema.BlockTypeCategory:
			// decode block
			category, blockRes := d.DecodeBlock(block, parseCtx)
			res.Merge(blockRes)
			if !blockRes.Success() {
				continue
			}

			// add the category to the nodeAndEdgeProvider
			res.AddDiags(nodeAndEdgeProvider.AddCategory(category.(*powerpipe.DashboardCategory)))

			// DO NOT add the category to the mod

		case schema.BlockTypeNode, schema.BlockTypeEdge:
			child, childRes := d.decodeQueryProvider(block, parseCtx)

			// TACTICAL if child has any runtime dependencies, claim them
			// this is to ensure if this resource is used as base, we can be correctly identified
			// as the publisher of the runtime dependencies
			for _, r := range child.(powerpipe.QueryProvider).GetRuntimeDependencies() {
				r.Provider = nodeAndEdgeProvider
			}

			// populate metadata, set references and call OnDecoded
			parse.HandleModDecodeResult(child, childRes, block, parseCtx)
			res.Merge(childRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddChild(child)
				res.AddDiags(moreDiags)
			}
		case schema.BlockTypeWith:
			with, withRes := d.DecodeBlock(block, parseCtx)
			res.Merge(withRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddWith(with.(*powerpipe.DashboardWith))
				res.AddDiags(moreDiags)
			}
		}

	}

	return res
}

func (d *PowerpipeModDecoder) decodeQueryProvider(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {
	res := parse.NewDecodeResult()
	// get shell resource
	resource, diags := parse.ResourceForBlock(block, parseCtx)
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
	diags = parse.DecodeHclBody(remain, parseCtx.EvalCtx, parseCtx, resource)
	res.HandleDecodeDiags(diags)

	// decode 'with',args and params blocks
	res.Merge(d.decodeQueryProviderBlocks(block, remain.(*hclsyntax.Body), resource, parseCtx))

	// resolve the connection string and (if set) search path
	qp := resource.(powerpipe.QueryProvider)
	connectionString, searchPath, searchPathPrefix, diags := parse.ResolveConnectionString(databaseContent, parseCtx.EvalCtx)
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

func (d *PowerpipeModDecoder) decodeQueryProviderBlocks(block *hcl.Block, content *hclsyntax.Body, resource modconfig.HclResource, parseCtx *parse.ModParseContext) *parse.DecodeResult {
	var diags hcl.Diagnostics
	res := parse.NewDecodeResult()
	queryProvider, ok := resource.(powerpipe.QueryProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a QueryProvider", block.Type))
	}

	if attr, exists := content.Attributes[schema.AttributeTypeArgs]; exists {
		args, runtimeDependencies, diags := parse.DecodeArgs(attr.AsHCLAttribute(), parseCtx.EvalCtx, queryProvider)
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
			paramDef, runtimeDependencies, moreDiags := parse.DecodeParam(block, parseCtx)
			if !moreDiags.HasErrors() {
				params = append(params, paramDef)
				queryProvider.AddRuntimeDependencies(runtimeDependencies)
				// add and references contained in the param block to the control refs
				moreDiags = parse.AddReferences(resource, block, parseCtx)
			}
			diags = append(diags, moreDiags...)
		}
	}

	queryProvider.SetParams(params)
	res.HandleDecodeDiags(diags)
	return res
}

func (d *PowerpipeModDecoder) decodeDashboard(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {
	res := parse.NewDecodeResult()
	dashboard := powerpipe.NewDashboard(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*powerpipe.Dashboard)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.HandleDecodeDiags(diags)

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = parse.DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, dashboard)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	if dashboard.Base != nil && len(dashboard.Base.ChildNames) > 0 {
		supportedChildren := []string{schema.BlockTypeContainer, schema.BlockTypeChart, schema.BlockTypeCard, schema.BlockTypeFlow, schema.BlockTypeGraph, schema.BlockTypeHierarchy, schema.BlockTypeImage, schema.BlockTypeInput, schema.BlockTypeTable, schema.BlockTypeText}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := parse.ResolveChildrenFromNames(dashboard.Base.ChildNames, block, supportedChildren, parseCtx)
		dashboard.Base.Children = children
	}
	if !res.Success() {
		return dashboard, res
	}

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := d.decodeDashboardBlocks(body, dashboard, parseCtx)
		res.Merge(blocksRes)
	}

	return dashboard, res
}

func (d *PowerpipeModDecoder) decodeDashboardBlocks(content *hclsyntax.Body, dashboard *powerpipe.Dashboard, parseCtx *parse.ModParseContext) *parse.DecodeResult {
	var res = parse.NewDecodeResult()
	// set dashboard as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboard)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()

		// decode block
		resource, blockRes := d.DecodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// we expect either inputs or child report nodes
		// add the resource to the mod
		res.AddDiags(parse.AddResourceToMod(resource, block, parseCtx))
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

func (d *PowerpipeModDecoder) decodeDashboardContainer(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {
	res := parse.NewDecodeResult()
	container := powerpipe.NewDashboardContainer(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*powerpipe.DashboardContainer)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.HandleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = parse.DecodeHclBody(body, parseCtx.EvalCtx, parseCtx, container)
	// handle any resulting diags, which may specify dependencies
	res.HandleDecodeDiags(diags)

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := d.decodeDashboardContainerBlocks(body, container, parseCtx)
		res.Merge(blocksRes)
	}

	return container, res
}

func (d *PowerpipeModDecoder) decodeDashboardContainerBlocks(content *hclsyntax.Body, dashboardContainer *powerpipe.DashboardContainer, parseCtx *parse.ModParseContext) *parse.DecodeResult {
	var res = parse.NewDecodeResult()

	// set container as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboardContainer)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		resource, blockRes := d.DecodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// special handling for inputs
		if b.Type == schema.BlockTypeInput {
			input := resource.(*powerpipe.DashboardInput)
			dashboardContainer.Inputs = append(dashboardContainer.Inputs, input)
			dashboardContainer.AddChild(input)
			// the input will be added to the mod by the parent dashboard

		} else {
			// for all other children, add to mod and children
			res.AddDiags(parse.AddResourceToMod(resource, block, parseCtx))
			if child, ok := resource.(modconfig.ModTreeItem); ok {
				dashboardContainer.AddChild(child)
			}
		}
	}

	return res
}

func (d *PowerpipeModDecoder) decodeBenchmark(block *hcl.Block, parseCtx *parse.ModParseContext) (modconfig.HclResource, *parse.DecodeResult) {
	res := parse.NewDecodeResult()
	benchmark := powerpipe.NewBenchmark(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*powerpipe.Benchmark)
	content, diags := block.Body.Content(parse.BenchmarkBlockSchema)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "children", &benchmark.ChildNames, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "description", &benchmark.Description, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "documentation", &benchmark.Documentation, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "tags", &benchmark.Tags, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "title", &benchmark.Title, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "type", &benchmark.Type, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	diags = parse.DecodeProperty(content, "display", &benchmark.Display, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)

	// now add children
	if res.Success() {
		supportedChildren := []string{schema.BlockTypeBenchmark, schema.BlockTypeControl}
		children, diags := parse.ResolveChildrenFromNames(benchmark.ChildNames.StringList(), block, supportedChildren, parseCtx)
		res.HandleDecodeDiags(diags)

		// now set children and child name strings
		benchmark.Children = children
		benchmark.ChildNameStrings = parse.GetChildNameStringsFromModTreeItem(children)
	}

	diags = parse.DecodeProperty(content, "base", &benchmark.Base, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)
	if benchmark.Base != nil && len(benchmark.Base.ChildNames) > 0 {
		supportedChildren := []string{schema.BlockTypeBenchmark, schema.BlockTypeControl}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := parse.ResolveChildrenFromNames(benchmark.Base.ChildNameStrings, block, supportedChildren, parseCtx)
		benchmark.Children = children
	}
	diags = parse.DecodeProperty(content, "width", &benchmark.Width, parseCtx.EvalCtx)
	res.HandleDecodeDiags(diags)
	return benchmark, res
}
