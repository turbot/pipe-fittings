package flowpipe

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

type Mod struct {
	*modconfig.ModBase[*FlowpipeResourceMaps]
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	return &Mod{
		ModBase: modconfig.NewMod[*FlowpipeResourceMaps](shortName, modPath, defRange),
	}
}
