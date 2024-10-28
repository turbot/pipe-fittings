package parse

import (
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

type ModDependencyConfig struct {
	ModDependency  *versionmap.ResolvedVersionConstraint
	DependencyPath *string
}

func (c ModDependencyConfig) SetModProperties(mod *modconfig.Mod) {
	mod.SetDependencyConfig(&c.ModDependency.DependencyVersion, c.DependencyPath, c.ModDependency.Name)
}

func NewDependencyConfig(modDependency *versionmap.ResolvedVersionConstraint) *ModDependencyConfig {
	var d string
	switch {
	case modDependency.Branch != "":
		d = fmt.Sprintf("%s#%s", modDependency.Name, modDependency.Branch)
	case modDependency.FilePath != "":
		d = modDependency.Name
	case modDependency.Version != nil:
		d = fmt.Sprintf("%s@v%s", modDependency.Name, modDependency.Version.String())
	}
	return &ModDependencyConfig{
		DependencyPath: &d,
		ModDependency:  modDependency,
	}
}
