package parse

import (
	"fmt"
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
	UnresolvedBlocks map[string]*unresolvedBlock
	FileData         map[string][]byte

	// the eval context used to decode references in HCL
	EvalCtx *hcl.EvalContext

	Diags hcl.Diagnostics

	RootEvalPath string

	// if set, only decode these blocks
	blockTypes map[string]struct{}
	// if set, exclude these block types
	blockTypeExclusions map[string]struct{}

	dependencyGraph *topsort.Graph
	blocks          hcl.Blocks
}

func NewParseContext(rootEvalPath string) ParseContext {
	c := ParseContext{
		UnresolvedBlocks: make(map[string]*unresolvedBlock),
		RootEvalPath:     rootEvalPath,
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()

	return c
}

func (r *ParseContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	r.blocks = content.Blocks
	r.FileData = fileData
}

func (r *ParseContext) ClearDependencies() {
	r.UnresolvedBlocks = make(map[string]*unresolvedBlock)
	r.dependencyGraph = r.newDependencyGraph()
}

// AddDependencies is called when a block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (r *ParseContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if r.UnresolvedBlocks[name] != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("duplicate unresolved block name '%s'", name),
			Detail:   fmt.Sprintf("block '%s' already exists. This could mean that there are unresolved duplicate resources,", name),
			Subject:  &block.DefRange,
		})
		return diags
	}

	// store unresolved block
	r.UnresolvedBlocks[name] = newUnresolvedBlock(block, name, dependencies)

	// store dependency in tree - d
	if !r.dependencyGraph.ContainsNode(name) {
		r.dependencyGraph.AddNode(name)
	}
	// add root dependency
	if err := r.dependencyGraph.AddEdge(rootDependencyNode, name); err != nil {
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
			parsedPropertyPath, err := modconfig.ParseResourcePropertyPath(hclhelpers.TraversalAsString(t))

			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to parse dependency",
					Detail:   err.Error(),
					Subject:  hclhelpers.BlockRangePointer(block),
				})
				continue
			}
			if parsedPropertyPath == nil {
				continue
			}

			// 'd' may be a property path - when storing dependencies we only care about the resource names
			dependencyResourceName := parsedPropertyPath.ToResourceName()
			if !r.dependencyGraph.ContainsNode(dependencyResourceName) {
				r.dependencyGraph.AddNode(dependencyResourceName)
			}
			if err := r.dependencyGraph.AddEdge(name, dependencyResourceName); err != nil {
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
func (r *ParseContext) BlocksToDecode() (blocksToDecode hcl.Blocks, _ error) {
	defer func() {
		// apply block inclusions and exclusions (if any)
		blocksToDecode = r.filterBlocks(blocksToDecode)
	}()

	depOrder, err := r.getDependencyOrder()
	if err != nil {
		return nil, err
	}
	if len(depOrder) == 0 {
		return r.blocks, nil
	}

	// NOTE: a block may appear more than once in unresolved blocks
	// if it defines multiple unresolved resources, e.g a locals block

	// make a map of blocks we have already included, keyed by the block def range
	blocksMap := make(map[string]bool)
	for _, name := range depOrder {
		// depOrder is all the blocks required to resolve dependencies.
		// if this one is unparsed, added to list
		block, ok := r.UnresolvedBlocks[name]
		if ok && !blocksMap[block.DeclRange.String()] {
			blocksToDecode = append(blocksToDecode, block.Block)
			// add to map
			blocksMap[block.DeclRange.String()] = true
		}
	}
	return blocksToDecode, nil
}
func (r *ParseContext) filterBlocks(blocks hcl.Blocks) hcl.Blocks {
	var res hcl.Blocks

	for _, block := range blocks {
		if r.shouldIncludeBlock(block) {
			res = append(res, block)
		}
	}
	return res
}

func (r *ParseContext) shouldIncludeBlock(block *hcl.Block) bool {
	// if inclusions are set, only include these block types
	if len(r.blockTypes) > 0 {
		if _, ok := r.blockTypes[block.Type]; !ok {
			return false
		}
	}
	// if exclusions are set, apply them
	if len(r.blockTypeExclusions) > 0 {
		if _, ok := r.blockTypeExclusions[block.Type]; ok {
			return false
		}
	}
	return true
}

// EvalComplete returns whether all elements in the dependency tree fully evaluated
func (r *ParseContext) EvalComplete() bool {
	return len(r.UnresolvedBlocks) == 0
}

func (r *ParseContext) FormatDependencies() string {
	// first get the dependency order
	dependencyOrder, err := r.getDependencyOrder()
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
		dep, ok := r.UnresolvedBlocks[resourceName]

		if ok {
			depStrings[i] = dep.String()
		} else {
			// this could happen if there is a dependency on a missing item
			depStrings[i] = fmt.Sprintf("  MISSING: %s", resourceName)
		}
	}

	return helpers.Tabify(strings.Join(depStrings, "\n"), "   ")
}

func (r *ParseContext) newDependencyGraph() *topsort.Graph {
	dependencyGraph := topsort.NewGraph()
	// add root node - this will depend on all other nodes
	dependencyGraph.AddNode(rootDependencyNode)
	return dependencyGraph
}

// return the optimal run order required to resolve dependencies

func (r *ParseContext) getDependencyOrder() ([]string, error) {
	rawDeps, err := r.dependencyGraph.TopSort(rootDependencyNode)
	if err != nil {
		return nil, err
	}

	// now remove the variable names and dedupe
	var deps []string
	for _, d := range rawDeps {
		if d == rootDependencyNode {
			continue
		}

		propertyPath, err := modconfig.ParseResourcePropertyPath(d)
		if err != nil {
			return nil, err
		}
		dep := modconfig.BuildModResourceName(propertyPath.ItemType, propertyPath.Name)
		if !helpers.StringSliceContains(deps, dep) {
			deps = append(deps, dep)
		}
	}
	return deps, nil
}

// eval functions

func (r *ParseContext) BuildEvalContext(variables map[string]cty.Value) {
	// create evaluation context
	r.EvalCtx = &hcl.EvalContext{
		Variables: variables,
		// use the mod path as the file root for functions
		Functions: funcs.ContextFunctions(r.RootEvalPath),
	}
}
