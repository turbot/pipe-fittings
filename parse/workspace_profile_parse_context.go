package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type WorkspaceProfileParseContext[T modconfig.WorkspaceProfile] struct {
	ParseContext
	workspaceProfiles map[string]T
	valueMap          map[string]cty.Value
}

func NewWorkspaceProfileParseContext[T modconfig.WorkspaceProfile](rootEvalPath string) *WorkspaceProfileParseContext[T] {
	parseContext := NewParseContext(rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done
	//parseContext.BlockTypes = []string{schema.BlockTypeWorkspaceProfile}
	c := &WorkspaceProfileParseContext[T]{
		ParseContext:      parseContext,
		workspaceProfiles: make(map[string]T),
		valueMap:          make(map[string]cty.Value),
	}

	c.buildEvalContext()

	return c
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (c *WorkspaceProfileParseContext[T]) AddResource(workspaceProfile T) hcl.Diagnostics {
	profileName := workspaceProfile.ShortName()
	ctyVal, err := workspaceProfile.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert workspaceProfile '%s' to its cty value", profileName),
			Detail:   err.Error(),
			Subject:  workspaceProfile.GetDeclRange(),
		}}
	}

	c.workspaceProfiles[profileName] = workspaceProfile
	c.valueMap[workspaceProfile.ShortName()] = ctyVal

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, profileName)

	c.buildEvalContext()

	return nil
}

func (c *WorkspaceProfileParseContext[T]) buildEvalContext() {
	// rebuild the eval context
	// build a map with a single key - workspace
	vars := map[string]cty.Value{
		"workspace": cty.ObjectVal(c.valueMap),
	}
	c.ParseContext.BuildEvalContext(vars)

}
