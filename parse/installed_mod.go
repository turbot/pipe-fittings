package parse

import (
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

type InstalledMod struct {
	Mod     *modconfig.Mod
	Version *semver.Version
}
