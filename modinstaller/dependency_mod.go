package modinstaller

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

// DependencyMod is a mod which has been installed as a dependency
type DependencyMod struct {
	InstalledVersion *versionmap.InstalledModVersion
	Mod              *modconfig.Mod
}
