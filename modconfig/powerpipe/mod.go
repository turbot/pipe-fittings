package powerpipe

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

type Mod struct {
	*modconfig.ModBase[*PowerpipeResourceMaps]
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	m := &Mod{
		ModBase: modconfig.NewModBase[*PowerpipeResourceMaps](shortName, modPath, defRange),
	}

	m.ResourceMaps = NewPowerpipeResourceMaps(m).(*PowerpipeResourceMaps)
	return m
}
