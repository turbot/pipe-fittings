package powerpipe

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

type Mod struct {
	*modconfig.ModBase[*PowerpipeResourceMaps]
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	return &Mod{
		ModBase: modconfig.NewMod[*PowerpipeResourceMaps](shortName, modPath, defRange),
	}
}
