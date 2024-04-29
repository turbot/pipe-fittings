package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

// DependencyMod is a mod which has been installed as a dependency
// enrich the mod with commit hash and ref
type DependencyMod struct {
	*modconfig.Mod
	Commit string
	GitRef string
}
