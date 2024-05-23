package modconfig

import (
	"fmt"
	"github.com/turbot/pipe-fittings/app_specific"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
)

type AppRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r *AppRequire) initialise(requireBlock *hcl.Block) hcl.Diagnostics {
	// find the steampipe block
	appBlock := hclhelpers.FindFirstChildBlock(requireBlock, app_specific.AppName)
	if appBlock == nil {
		// can happen if there is a legacy property - just use the parent block
		appBlock = requireBlock
	}
	// set DeclRange
	r.DeclRange = hclhelpers.BlockRange(appBlock)

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
				Summary:  fmt.Sprintf("invalid required %s version %s", app_specific.AppName, r.MinVersionString),
				Subject:  &r.DeclRange,
			}}
	}
	return nil
}
