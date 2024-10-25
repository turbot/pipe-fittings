package modconfig

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/plugin"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/sperr"
)

// Require is a struct representing mod dependencies
type Require struct {
	Plugins                          []*plugin.PluginVersion `hcl:"plugin,block"`
	DeprecatedSteampipeVersionString string                  `hcl:"steampipe,optional"`

	// one of these may be set
	Flowpipe  *AppRequire `hcl:"flowpipe,block"`
	Steampipe *AppRequire `hcl:"steampipe,block"`
	Powerpipe *AppRequire `hcl:"powerpipe,block"`

	Mods []*ModVersionConstraint `hcl:"mod,block"`
	// map keyed by name [and alias]
	ModMap map[string]*ModVersionConstraint
	// range of the require block body
	DeclRange hcl.Range
	// range of the require block type
	TypeRange hcl.Range

	// the app require block - one of Flowpipe/Steampipe/Powerpipe
	// reference it as 'app' to avoid having to check all the blocks
	app *AppRequire
}

func NewRequire() *Require {
	return &Require{
		ModMap: make(map[string]*ModVersionConstraint),
	}
}

func (r *Require) Clone() *Require {
	require := NewRequire()
	require.Flowpipe = r.Flowpipe
	require.Steampipe = r.Steampipe
	require.Powerpipe = r.Powerpipe
	require.app = r.app

	require.Plugins = r.Plugins
	require.Mods = r.Mods
	require.DeclRange = r.DeclRange
	require.TypeRange = r.TypeRange

	// we need to shallow copy the map
	// if we don't, when the other one gets
	// modified - this one gets as well
	for k, mvc := range r.ModMap {
		require.ModMap[k] = mvc
	}
	return require
}

func (r *Require) initialise(modBlock *hcl.Block) hcl.Diagnostics {
	// This will actually be called twice - once when we load the mod definition,
	// and again when we load the mod resources (and set the mod metadata, references etc)
	// If we have already initialised, return (we can tell by checking the DeclRange)
	if !r.DeclRange.Empty() {
		return nil
	}

	var diags hcl.Diagnostics
	// handle deprecated properties
	moreDiags := r.handleDeprecations()
	diags = append(diags, moreDiags...)

	requireBlock, moreDiags := FindRequireBlock(modBlock)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return diags
	}
	if requireBlock == nil {
		// nothing else to populate
		return nil
	}
	// set our Ranges
	r.DeclRange = hclhelpers.BlockRange(requireBlock)
	r.TypeRange = requireBlock.TypeRange

	// initialise the app block (powerpipe/steampipe/flowpipe)
	diags = append(diags, r.initialiseAppRequire()...)
	if diags.HasErrors() {
		return diags
	}
	return r.InitialiseConstraints(requireBlock)

}

func (r *Require) InitialiseConstraints(requireBlock *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// build maps of plugin and mod blocks
	pluginBlockMap := hclhelpers.BlocksToMap(hclhelpers.FindChildBlocks(requireBlock, schema.BlockTypePlugin))
	modBlockMap := hclhelpers.BlocksToMap(hclhelpers.FindChildBlocks(requireBlock, schema.BlockTypeMod))

	if r.app != nil {
		moreDiags := r.app.initialise(requireBlock)
		diags = append(diags, moreDiags...)
	}

	for _, p := range r.Plugins {
		moreDiags := p.Initialise(pluginBlockMap[p.RawName])
		diags = append(diags, moreDiags...)
	}
	for _, m := range r.Mods {
		moreDiags := m.Initialise(modBlockMap[m.Name])
		diags = append(diags, moreDiags...)
		if !diags.HasErrors() {
			// key map entry by name [and alias]
			r.ModMap[m.Name] = m
		}
	}
	return diags
}

func (r *Require) handleDeprecations() hcl.Diagnostics {
	var diags hcl.Diagnostics
	// the 'steampipe' property is deprecated and replace with a steampipe block
	if r.DeprecatedSteampipeVersionString != "" {
		// if there is both a steampipe block and property, fail
		if r.app != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Both 'steampipe' block and deprecated 'steampipe' property are set",
				Subject:  &r.DeclRange,
			})
		} else {
			r.app = &AppRequire{MinVersionString: r.DeprecatedSteampipeVersionString}
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Property 'steampipe' is deprecated for mod require block - use a steampipe block instead",
				Subject:  &r.DeclRange,
			},
			)
		}
	}
	return diags
}

func (r *Require) validateAppVersion(modName string) error {
	if appVersionConstraint := r.AppVersionConstraint(); appVersionConstraint != nil {
		if !appVersionConstraint.Check(app_specific.AppVersion) {
			return fmt.Errorf("%s version %s does not satisfy %s which requires version %s", app_specific.AppName, app_specific.AppVersion.String(), modName, r.app.MinVersionString)
		}
	}
	return nil
}

// validatePluginVersions validates that for every plugin requirement there's at least one plugin installed
func (r *Require) validatePluginVersions(modName string, plugins plugin.PluginVersionMap) []error {
	if len(r.Plugins) == 0 {
		return nil
	}
	// if this is a steampipe backend and there is no plugin map, it must be a pre-0.22 version which does not return plugin versions
	if plugins.Backend == constants.SteampipeBackendName && plugins.AvailablePlugins == nil {
		slog.Warn("Mod plugin requirements cannot be validated. Steampipe backend does not provide plugin version information. Upgrade Steampipe to enable plugin version validation.", "mod", modName)
		return nil
	}

	var validationErrors []error
	for _, requiredPlugin := range r.Plugins {
		if err := r.searchInstalledPluginForRequirement(modName, requiredPlugin, plugins); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	return validationErrors
}

// searchInstalledPluginForRequirement returns plugin validation errors if no plugin is found which satisfies
// the mod requirement. If plugin is found nil error is returned.
func (r *Require) searchInstalledPluginForRequirement(modName string, requirement *plugin.PluginVersion, plugins plugin.PluginVersionMap) error {
	for installedName, installed := range plugins.AvailablePlugins {
		org, name, _ := ociinstaller.NewImageRef(installedName).GetOrgNameAndStream()
		if org != requirement.Org || name != requirement.Name {
			// no point checking - different plugin
			continue
		}
		// if org and name matches but the plugin is built locally, return without any validation error
		if installed.IsLocal() {
			return nil
		}
		// if org and name matches, check whether the version constraint is satisfied
		if requirement.Constraint.Check(installed.Semver()) {
			// constraint is satisfied
			return nil
		}
	}
	// validation failed - return error
	return sperr.New("%s backend '%s' does not provide a plugin which satisfies requirement '%s@%s' - required by '%s'", plugins.Backend, plugins.Database, requirement.RawName, requirement.MinVersionString, modName)
}

// AddModDependencies adds all the mod in newModVersions to our list of mods, using the following logic
// - if a mod with same name, [alias] and constraint exists, it is not added
// - if a mod with same name [and alias] and different constraint exist, it is replaced
func (r *Require) AddModDependencies(newModVersions map[string]*ModVersionConstraint) {
	// rebuild the Mods array

	// first rebuild the mod map
	for name, newVersion := range newModVersions {
		r.ModMap[name] = newVersion
	}

	// now update the mod array from the map
	var newMods = make([]*ModVersionConstraint, len(r.ModMap))
	idx := 0
	for _, requiredVersion := range r.ModMap {
		newMods[idx] = requiredVersion
		idx++
	}
	// sort by name
	sort.Sort(ModVersionConstraintCollection(newMods))
	// write back
	r.Mods = newMods
}

func (r *Require) RemoveModDependencies(versions map[string]*ModVersionConstraint) {
	// first rebuild the mod map
	for name := range versions {
		delete(r.ModMap, name)
	}
	// now update the mod array from the map
	var newMods = make([]*ModVersionConstraint, len(r.ModMap))
	idx := 0
	for _, requiredVersion := range r.ModMap {
		newMods[idx] = requiredVersion
		idx++
	}
	// sort by name
	sort.Sort(ModVersionConstraintCollection(newMods))
	// write back
	r.Mods = newMods
}

func (r *Require) RemoveAllModDependencies() {
	r.Mods = nil
}

func (r *Require) GetModDependency(name string /*,alias string*/) *ModVersionConstraint {
	return r.ModMap[name]
}

func (r *Require) ContainsMod(requiredModVersion *ModVersionConstraint) bool {
	if c := r.GetModDependency(requiredModVersion.Name); c != nil {
		return c.Equals(requiredModVersion)
	}
	return false
}

func (r *Require) Empty() bool {
	if r == nil {
		return true
	}

	return r.AppVersionConstraint() == nil && len(r.Mods) == 0 && len(r.Plugins) == 0
}

func (r *Require) AppVersionConstraint() *semver.Constraints {
	if r.app == nil {
		return nil
	}
	return r.app.Constraint
}

func (r *Require) initialiseAppRequire() hcl.Diagnostics {
	// ensure the app block matching the app name is set
	if r.Flowpipe != nil {
		if app_specific.AppName == "flowpipe" {
			r.app = r.Flowpipe
		} else {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `blocks of type "flowpipe" are not expected here`,
				Subject:  &r.DeclRange,
			}}
		}
	}
	if r.Steampipe != nil {
		if app_specific.AppName == "steampipe" {
			r.app = r.Steampipe
		} else if app_specific.AppName != "powerpipe" {
			// powerpipe does not error for steampipe block, but ignores it
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `blocks of type "steampipe"" are not expected here`,
				Subject:  &r.DeclRange,
			}}
		}
	}
	if r.Powerpipe != nil {
		if app_specific.AppName == "powerpipe" {
			r.app = r.Powerpipe
		} else {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `blocks of type "powerpipe" are not expected here`,
				Subject:  &r.DeclRange,
			}}
		}
	}
	return nil
}

// FindRequireBlock finds the require block under the given mod block
func FindRequireBlock(modBlock *hcl.Block) (*hcl.Block, hcl.Diagnostics) {
	requireBlock := hclhelpers.FindFirstChildBlock(modBlock, schema.BlockTypeRequire)
	if requireBlock == nil {
		// was this the legacy 'requires' block?
		requireBlock = hclhelpers.FindFirstChildBlock(modBlock, schema.BlockTypeLegacyRequires)
	}
	return requireBlock, nil
}
