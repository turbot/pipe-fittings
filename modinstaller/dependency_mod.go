package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

// DependencyMod is a mod which has been installed as a dependency
// enrich the mod with commit hash and ref
type DependencyMod struct {
	Constraint *versionmap.ResolvedVersionConstraint
	Mod        *modconfig.Mod
}
