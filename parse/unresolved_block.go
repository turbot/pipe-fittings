package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

type UnresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	DeclRange    hcl.Range
	Dependencies map[string]*modconfig.ResourceDependency
}

func NewUnresolvedBlock(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) *UnresolvedBlock {
	return &UnresolvedBlock{
		Name:         name,
		Block:        block,
		Dependencies: dependencies,
		DeclRange:    hclhelpers.BlockRange(block),
	}
}

func (b UnresolvedBlock) String() string {
	depStrings := make([]string, len(b.Dependencies))
	idx := 0
	for _, dep := range b.Dependencies {
		depStrings[idx] = fmt.Sprintf(`%s -> %s`, b.Name, dep.String())
		idx++
	}
	return strings.Join(depStrings, "\n")
}
