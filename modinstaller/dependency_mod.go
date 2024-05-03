package modinstaller

import (
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/versionmap"
)

// DependencyMod is a mod which has been installed as a dependency
type DependencyMod struct {
	InstalledVersion *versionmap.InstalledModVersion
	Mod              *modconfig.Mod
}
