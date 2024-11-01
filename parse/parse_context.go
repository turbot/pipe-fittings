package parse

import (
	"fmt"
	"golang.org/x/exp/maps"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type ParseContext struct {
	UnresolvedBlocks map[string]*UnresolvedBlock
	FileData         map[string][]byte

	// the eval context used to decode references in HCL
	EvalCtx *hcl.EvalContext

	Diags hcl.Diagnostics

	RootEvalPath string

	// if set, only decode these blocks
	blockTypes map[string]struct{}
	// if set, exclude these block types
	blockTypeExclusions map[string]struct{}

	DependencyGraph *topsort.Graph
	blocks          hcl.Blocks

	// function used to parse resource property path
	ResourceNameFromDependencyFunc func(propertyPath string) (string, error)
}

func NewParseContext(rootEvalPath string) ParseContext {
	c := ParseContext{
		UnresolvedBlocks: make(map[string]*UnresolvedBlock),
		RootEvalPath:     rootEvalPath,
		// use the default func
		ResourceNameFromDependencyFunc: resourceNameFromDependency,
	}
	// add root node - this will depend on all other nodes
	c.DependencyGraph = c.newDependencyGraph()

	return c
}

func (p *ParseContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	p.blocks = content.Blocks
	p.FileData = fileData
}

func (p *ParseContext) ClearDependencies() {
	p.UnresolvedBlocks = make(map[string]*UnresolvedBlock)
	p.DependencyGraph = p.newDependencyGraph()
}

// AddDependencies is called when a block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (p *ParseContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if p.UnresolvedBlocks[name] != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("duplicate unresolved block name '%s'", name),
			Detail:   fmt.Sprintf("block '%s' already exists. This could mean that there are unresolved duplicate resources,", name),
			Subject:  &block.DefRange,
		})
		return diags
	}

	// store unresolved block
	p.UnresolvedBlocks[name] = NewUnresolvedBlock(block, name, dependencies)

	// store dependency in tree - d
	if !p.DependencyGraph.ContainsNode(name) {
		p.DependencyGraph.AddNode(name)
	}
	// add root dependency
	if err := p.DependencyGraph.AddEdge(RootDependencyNode, name); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to add root dependency to graph",
			Detail:   err.Error(),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
	}

	for _, dep := range dependencies {
		// each dependency object may have multiple traversals
		for _, t := range dep.Traversals {
			dependencyResourceName, err := p.ResourceNameFromDependencyFunc(hclhelpers.TraversalAsString(t))
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to parse dependency",
					Detail:   err.Error(),
					Subject:  hclhelpers.BlockRangePointer(block),
				})
				continue
			}
			if dependencyResourceName == "" {
				continue
			}
			if !p.DependencyGraph.ContainsNode(dependencyResourceName) {
				p.DependencyGraph.AddNode(dependencyResourceName)
			}
			if err := p.DependencyGraph.AddEdge(name, dependencyResourceName); err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to add dependency to graph",
					Detail:   err.Error(),
					Subject:  hclhelpers.BlockRangePointer(block),
				})
			}
		}
	}
	return diags
}

// BlocksToDecode builds a list of blocks to decode, the order of which is determined by the dependency order
func (p *ParseContext) BlocksToDecode() (blocksToDecode hcl.Blocks, _ error) {
	defer func() {
		// apply block inclusions and exclusions (if any)
		blocksToDecode = p.filterBlocks(blocksToDecode)
	}()

	depOrder, err := p.getDependencyOrder()
	if err != nil {
		return nil, err
	}
	if len(depOrder) == 0 {
		return p.blocks, nil
	}

	// NOTE: a block may appear more than once in unresolved blocks
	// if it defines multiple unresolved resources, e.g a locals block

	// make a map of blocks we have already included, keyed by the block def range
	blocksMap := make(map[string]bool)
	for _, name := range depOrder {
		// depOrder is all the blocks required to resolve dependencies.
		// if this one is unparsed, added to list
		block, ok := p.UnresolvedBlocks[name]
		if ok && !blocksMap[block.DeclRange.String()] {
			blocksToDecode = append(blocksToDecode, block.Block)
			// add to map
			blocksMap[block.DeclRange.String()] = true
		}
	}
	return blocksToDecode, nil
}

func (p *ParseContext) filterBlocks(blocks hcl.Blocks) hcl.Blocks {
	var res hcl.Blocks

	for _, block := range blocks {
		if p.shouldIncludeBlock(block) {
			res = append(res, block)
		}
	}
	return res
}

func (p *ParseContext) shouldIncludeBlock(block *hcl.Block) bool {
	// if inclusions are set, only include these block types
	if len(p.blockTypes) > 0 {
		if _, ok := p.blockTypes[block.Type]; !ok {
			return false
		}
	}
	// if exclusions are set, apply them
	if len(p.blockTypeExclusions) > 0 {
		if _, ok := p.blockTypeExclusions[block.Type]; ok {
			return false
		}
	}
	return true
}

// EvalComplete returns whether all elements in the dependency tree fully evaluated
func (p *ParseContext) EvalComplete() bool {
	return len(p.UnresolvedBlocks) == 0
}

func (p *ParseContext) FormatDependencies() string {
	// first get the dependency order
	dependencyOrder, err := p.getDependencyOrder()
	if err != nil {
		return err.Error()
	}
	// build array of dependency strings - processes dependencies in reverse order for presentation reasons
	numDeps := len(dependencyOrder)
	depStrings := make([]string, numDeps)
	for i := 0; i < len(dependencyOrder); i++ {
		srcIdx := len(dependencyOrder) - i - 1
		resourceName := dependencyOrder[srcIdx]
		// find dependency
		dep, ok := p.UnresolvedBlocks[resourceName]

		if ok {
			depStrings[i] = dep.String()
		} else {
			// this could happen if there is a dependency on a missing item
			depStrings[i] = fmt.Sprintf("  MISSING: %s", resourceName)
		}
	}

	return helpers.Tabify(strings.Join(depStrings, "\n"), "   ")
}

func (p *ParseContext) newDependencyGraph() *topsort.Graph {
	dependencyGraph := topsort.NewGraph()
	// add root node - this will depend on all other nodes
	dependencyGraph.AddNode(RootDependencyNode)
	return dependencyGraph
}

// return the optimal run order required to resolve dependencies

func (p *ParseContext) getDependencyOrder() ([]string, error) {
	rawDeps, err := p.DependencyGraph.TopSort(RootDependencyNode)
	if err != nil {
		return nil, err
	}

	// now remove the variable names and dedupe
	var deps = map[string]struct{}{}
	for _, d := range rawDeps {
		if d == RootDependencyNode {
			continue
		}

		dep, err := p.ResourceNameFromDependencyFunc(d)
		if err != nil {
			return nil, err
		}
		deps[dep] = struct{}{}

	}
	return maps.Keys(deps), nil
}

// eval functions

func (p *ParseContext) BuildEvalContext(variables map[string]cty.Value) {
	// create evaluation context
	p.EvalCtx = &hcl.EvalContext{
		Variables: variables,
		// use the RootEvalPath as the file root for functions
		Functions: funcs.ContextFunctions(p.RootEvalPath),
	}
}

func (p *ParseContext) SetBlockTypes(blockTypes ...string) {
	p.blockTypes = make(map[string]struct{}, len(blockTypes))
	for _, t := range blockTypes {
		p.blockTypes[t] = struct{}{}
	}
}

func (m *ParseContext) SetBlockTypeExclusions(blockTypes ...string) {
	m.blockTypeExclusions = make(map[string]struct{}, len(blockTypes))
	for _, t := range blockTypes {
		m.blockTypeExclusions[t] = struct{}{}
	}
}

// default resourceNameFromDependency func
func resourceNameFromDependency(propertyPath string) (string, error) {
	parsedPropertyPath, err := modconfig.ParseResourcePropertyPath(propertyPath)

	if err != nil {
		return "", err
	}
	if parsedPropertyPath == nil {
		return "", nil
	}

	// 'd' may be a property path - when storing dependencies we only care about the resource names
	dependencyResourceName := parsedPropertyPath.ToResourceName()
	return dependencyResourceName, nil
}
