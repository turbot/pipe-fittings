package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/schema"
)

type FlowpipeRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r *FlowpipeRequire) initialise(requireBlock *hcl.Block) hcl.Diagnostics {
	// find the steampipe block
	flowpipeBlock := hclhelpers.FindFirstChildBlock(requireBlock, schema.BlockTypeSteampipe)
	if flowpipeBlock == nil {
		// can happen if there is a legacy property - just use the parent block
		flowpipeBlock = requireBlock
	}
	// set DeclRange
	r.DeclRange = hclhelpers.BlockRange(flowpipeBlock)

	if r.MinVersionString == "" {
		return nil
	}

	// convert min version into constraint (including prereleases)
	minVersion, err := semver.NewVersion(strings.TrimPrefix(r.MinVersionString, "v"))
	if err == nil {
		r.Constraint, err = semver.NewConstraint(fmt.Sprintf(">=%s-0", minVersion))
	}
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid required flowpipe version %s", r.MinVersionString),
				Subject:  &r.DeclRange,
			}}
	}
	return nil
}
