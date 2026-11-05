package parse

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/v2/schema"
)

// ResourceTarget is a resource that a mod load is for. See WithResourceTargets.
type ResourceTarget struct {
	// Mod is the short name of the mod that defines the resource. If it is empty, the resource can be in any mod.
	Mod string
	// Name is the resource type and name, for example "benchmark.cis_v300"
	Name string
}

// isResourceType returns whether blocks of the type are resources that can be referenced by name,
// and so can be left out of a load that does not need them
func isResourceType(blockType string) bool {
	switch blockType {
	case schema.BlockTypeMod, schema.BlockTypeVariable, schema.BlockTypeLocals:
		return false
	}
	for _, b := range WorkspaceBlockSchema.Blocks {
		if b.Type == blockType {
			return true
		}
	}
	return false
}

// resourceBlockName returns the name of a top level resource block, for example "control.c1",
// or false if the block is not a resource
func resourceBlockName(block *hcl.Block) (string, bool) {
	if len(block.Labels) == 0 || !isResourceType(block.Type) {
		return "", false
	}
	return block.Type + "." + block.Labels[0], true
}

// neededResources returns the names of the resource blocks in content that are needed for the given resources:
// the resources themselves, and every resource they reference, directly or indirectly.
// The blocks of mod, variable and locals are always decoded, so the resources referenced by locals are also needed.
//
// required are the names that must be defined in content. optional are names that are needed if content defines them.
// modName is the short name of the mod. depRefs are the resources it references in other mods, keyed by mod name.
//
// It returns false if it cannot tell which resources are needed: a required name or a reference is not a top
// level block in content (for example a nested block, or a query from a .sql file), or a block is not HCL native syntax.
// In that case, every resource must be decoded.
func neededResources(content *hcl.BodyContent, modName string, required, optional []string) (needed map[string]struct{}, depRefs map[string][]string, ok bool) {
	blocks := make(map[string]*hcl.Block)
	var localsBlocks []*hcl.Block
	for _, block := range content.Blocks {
		if name, isResource := resourceBlockName(block); isResource {
			blocks[name] = block
		} else if block.Type == schema.BlockTypeLocals {
			localsBlocks = append(localsBlocks, block)
		}
	}

	needed = make(map[string]struct{})
	depRefs = make(map[string][]string)
	var queue []string
	need := func(name string) bool {
		if _, exists := blocks[name]; !exists {
			return false
		}
		if _, seen := needed[name]; !seen {
			needed[name] = struct{}{}
			queue = append(queue, name)
		}
		return true
	}
	// addReferences needs every local resource that the block references, and records its references to other mods
	addReferences := func(block *hcl.Block) bool {
		refs, ok := blockReferences(block, modName)
		if !ok {
			return false
		}
		for _, ref := range refs {
			if ref.mod != "" {
				depRefs[ref.mod] = append(depRefs[ref.mod], ref.name)
			} else if !need(ref.name) {
				return false
			}
		}
		return true
	}

	for _, name := range required {
		if !need(name) {
			return nil, nil, false
		}
	}
	for _, name := range optional {
		need(name)
	}
	for _, block := range localsBlocks {
		if !addReferences(block) {
			return nil, nil, false
		}
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if !addReferences(blocks[name]) {
			return nil, nil, false
		}
	}
	return needed, depRefs, true
}

// resourceReference is a reference to a resource. mod is empty for a resource in the same mod.
type resourceReference struct {
	mod  string
	name string
}

// blockReferences returns the resources that the block and its nested blocks reference,
// or false if the block is not HCL native syntax
func blockReferences(block *hcl.Block, modName string) ([]resourceReference, bool) {
	body, ok := block.Body.(*hclsyntax.Body)
	if !ok {
		return nil, false
	}
	var refs []resourceReference
	var walk func(body *hclsyntax.Body)
	walk = func(body *hclsyntax.Body) {
		for _, attr := range body.Attributes {
			for _, traversal := range attr.Expr.Variables() {
				if ref, isResource := traversalReference(traversal, modName); isResource {
					refs = append(refs, ref)
				}
			}
		}
		for _, nested := range body.Blocks {
			walk(nested.Body)
		}
	}
	walk(body)
	return refs, true
}

// traversalReference returns the resource that a traversal refers to, if it refers to one:
// "control.c1.title" refers to control.c1 in the same mod, as does "<modName>.control.c1".
// "aws_compliance.control.c1" refers to control.c1 in mod aws_compliance.
// Other traversals, such as "var.region", "local.tags" or "self.input.x", do not refer to resources.
func traversalReference(traversal hcl.Traversal, modName string) (resourceReference, bool) {
	steps := []string{traversal.RootName()}
	for _, step := range traversal[1:] {
		attr, isAttr := step.(hcl.TraverseAttr)
		if !isAttr || len(steps) == 3 {
			break
		}
		steps = append(steps, attr.Name)
	}
	switch {
	case len(steps) >= 2 && isResourceType(steps[0]):
		return resourceReference{name: steps[0] + "." + steps[1]}, true
	case len(steps) == 3 && !slices.Contains([]string{"var", "local", "self"}, steps[0]) && isResourceType(steps[1]):
		ref := resourceReference{name: steps[1] + "." + steps[2]}
		if steps[0] != modName {
			ref.mod = steps[0]
		}
		return ref, true
	}
	return resourceReference{}, false
}
