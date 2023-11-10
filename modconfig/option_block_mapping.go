package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
)

type OptionsBlockFactory = func(*hcl.Block) (options.Options, hcl.Diagnostics)

// SteampipeOptionsBlockMapping is an OptionsBlockFactory used to map global steampipe options
// TODO look at deprecations
func SteampipeOptionsBlockMapping(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	switch block.Type {

	case options.ConnectionBlock:
		return new(options.Connection), nil
	case options.DatabaseBlock:
		return new(options.Database), nil
	case options.TerminalBlock:
		return new(options.Terminal), nil
	case options.GeneralBlock:
		return new(options.General), nil
	case options.QueryBlock:
		return new(options.Query), nil
	case options.CheckBlock:
		return new(options.Check), nil
	case options.DashboardBlock:
		return new(options.GlobalDashboard), nil
	case options.PluginBlock:
		return new(options.Plugin), nil
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Type),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return nil, diags
	}
}
