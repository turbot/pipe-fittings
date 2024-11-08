package modconfig

import (
	"fmt"
	"golang.org/x/exp/maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/plugin"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

// Mod is a struct representing a Mod resource
type Mod struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// attributes
	Categories []string `cty:"categories" hcl:"categories,optional" json:"categories,omitempty"`
	Color      *string  `cty:"color" hcl:"color" json:"color,omitempty"`
	Icon       *string  `cty:"icon" hcl:"icon" json:"icon,omitempty"`

	// blocks
	Require       *Require   `hcl:"require,block"  json:"-"`
	LegacyRequire *Require   `hcl:"requires,block"  json:"-"`
	OpenGraph     *OpenGraph `hcl:"opengraph,block" json:"open_graph,omitempty"`

	// Dependency attributes - set if this mod is loaded as a dependency

	// the mod version
	Version *DependencyVersion `json:"-"`
	// DependencyPath is the fully qualified mod name including version,
	// which will by the map key in the workspace lock file
	// NOTE: this is the relative path to the mod location from the dependency install dir (.steampipe/mods)
	// e.g. github.com/turbot/steampipe-mod-azure-thrifty@v1.0.0
	// It is populated for dependency mods as part of the mod loading process
	// NOTE: if this mod dependency is a local file dependency, the dependency path will be the file path
	DependencyPath *string `json:"dependency_path,omitempty"`
	// DependencyName return the name of the mod as a dependency, i.e. the mod dependency path, _without_ the version
	// e.g. github.com/turbot/steampipe-mod-azure-thrifty
	DependencyName string `json:"-"`

	// ModPath is the installation location of the mod
	ModPath string `json:"-"`

	// convenient aggregation of all resources
	Resources ModResources `json:"-"`

	// the filepath of the mod.sp/mod.fp/mod.pp file (will be empty for default mod)
	modFilePath string
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	name := fmt.Sprintf("mod.%s", shortName)
	mod := &Mod{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       shortName,
				FullName:        name,
				UnqualifiedName: name,
				DeclRange:       defRange,
				BlockType:       schema.BlockTypeMod,
			},
		},
		ModPath: modPath,
		Require: NewRequire(),
	}
	// call the app specific resource maps constructor to make an empty resource maps
	mod.Resources = NewModResources(mod)
	return mod
}

func (m *Mod) Equals(other *Mod) bool {
	res := m.ShortName == other.ShortName &&
		m.FullName == other.FullName &&
		typehelpers.SafeString(m.Color) == typehelpers.SafeString(other.Color) &&
		typehelpers.SafeString(m.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(m.Documentation) == typehelpers.SafeString(other.Documentation) &&
		typehelpers.SafeString(m.Icon) == typehelpers.SafeString(other.Icon) &&
		typehelpers.SafeString(m.Title) == typehelpers.SafeString(other.Title)
	if !res {
		return res
	}
	// categories
	if m.Categories == nil {
		if other.Categories != nil {
			return false
		}
	} else {
		// we have categories
		if other.Categories == nil {
			return false
		}

		if len(m.Categories) != len(other.Categories) {
			return false
		}
		for i, c := range m.Categories {
			if (other.Categories)[i] != c {
				return false
			}
		}
	}

	// tags
	if len(m.Tags) != len(other.Tags) {
		return false
	}
	for k, v := range m.Tags {
		if otherVal := other.Tags[k]; v != otherVal {
			return false
		}
	}

	// now check the child resources
	return m.Resources.Equals(other.Resources)
}

func (m *Mod) CacheKey() string {
	cacheKey := m.Name()
	if m.Version != nil {
		cacheKey += "." + m.Version.String()
	}

	return cacheKey
}

// CreateDefaultMod creates a default mod created for a workspace with no mod definition
func CreateDefaultMod(modPath string) *Mod {
	m := NewMod(defaultModName, modPath, hcl.Range{})
	folderName := filepath.Base(modPath)
	m.Title = &folderName
	return m
}

// IsDefaultMod returns whether this mod is a default mod created for a workspace with no mod definition
func (m *Mod) IsDefaultMod() bool {
	return m.modFilePath == ""
}

// GetPaths implements ModTreeItem (override base functionality)
func (m *Mod) GetPaths() []NodePath {
	return []NodePath{{m.Name()}}
}

// SetPaths implements ModTreeItem (override base functionality)
func (m *Mod) SetPaths() {}

// OnDecoded implements HclResource
func (m *Mod) OnDecoded(block *hcl.Block, _ ModResourcesProvider) hcl.Diagnostics {
	// handle legacy requires block
	if m.LegacyRequire != nil && !m.LegacyRequire.Empty() {
		// ensure that both 'require' and 'requires' were not set
		for _, b := range block.Body.(*hclsyntax.Body).Blocks {
			if b.Type == schema.BlockTypeRequire {
				return hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Both 'require' and legacy 'requires' blocks are defined",
					Subject:  hclhelpers.BlockRangePointer(block),
				}}
			}
		}
		m.Require = m.LegacyRequire
	}

	// initialise our Require
	if m.Require == nil {
		return nil
	}

	return m.Require.initialise(block)
}

// AddReference implements ResourceWithMetadata (overridden from ResourceWithMetadataImpl)
func (m *Mod) AddReference(ref *ResourceReference) {
	m.Resources.AddReference(ref)
}

// GetReferences implements ResourceWithMetadata (overridden from ResourceWithMetadataImpl)
func (m *Mod) GetReferences() []*ResourceReference {
	return maps.Values(m.Resources.GetReferences())
}

// GetModResources implements ModResourcesProvider
func (m *Mod) GetModResources() ModResources {
	return m.Resources
}

func (m *Mod) SetModResources(modResources ModResources) {
	m.Resources = modResources
}

func (m *Mod) GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool) {
	return m.Resources.GetResource(parsedName)
}

func (m *Mod) AddModDependencies(modVersions map[string]*ModVersionConstraint) {
	m.Require.AddModDependencies(modVersions)
}

func (m *Mod) RemoveModDependencies(modVersions map[string]*ModVersionConstraint) {
	m.Require.RemoveModDependencies(modVersions)
}

func (m *Mod) RemoveAllModDependencies() {
	m.Require.RemoveAllModDependencies()
}

func (m *Mod) Save() error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	modBody := rootBody.AppendNewBlock("mod", []string{m.ShortName}).Body()
	if m.Title != nil {
		modBody.SetAttributeValue("title", cty.StringVal(*m.Title))
	}
	if m.Description != nil {
		modBody.SetAttributeValue("description", cty.StringVal(*m.Description))
	}
	if m.Color != nil {
		modBody.SetAttributeValue("color", cty.StringVal(*m.Color))
	}
	if m.Documentation != nil {
		modBody.SetAttributeValue("documentation", cty.StringVal(*m.Documentation))
	}
	if m.Icon != nil {
		modBody.SetAttributeValue("icon", cty.StringVal(*m.Icon))
	}
	if len(m.Categories) > 0 {
		categoryValues := make([]cty.Value, len(m.Categories))
		for i, c := range m.Categories {
			categoryValues[i] = cty.StringVal(typehelpers.SafeString(c))
		}
		modBody.SetAttributeValue("categories", cty.ListVal(categoryValues))
	}

	if len(m.Tags) > 0 {
		tagMap := make(map[string]cty.Value, len(m.Tags))
		for k, v := range m.Tags {
			tagMap[k] = cty.StringVal(v)
		}
		modBody.SetAttributeValue("tags", cty.MapVal(tagMap))
	}

	// opengraph
	if opengraph := m.OpenGraph; opengraph != nil {
		opengraphBody := modBody.AppendNewBlock("opengraph", nil).Body()
		if opengraph.Title != nil {
			opengraphBody.SetAttributeValue("title", cty.StringVal(*opengraph.Title))
		}
		if opengraph.Description != nil {
			opengraphBody.SetAttributeValue("description", cty.StringVal(*opengraph.Description))
		}
		if opengraph.Image != nil {
			opengraphBody.SetAttributeValue("image", cty.StringVal(*opengraph.Image))
		}

	}

	// require
	if require := m.Require; require != nil && !require.Empty() {
		requiresBody := modBody.AppendNewBlock("require", nil).Body()

		if require.app != nil && require.app.MinVersionString != "" {
			steampipeRequiresBody := requiresBody.AppendNewBlock(app_specific.AppName, nil).Body()
			steampipeRequiresBody.SetAttributeValue("min_version", cty.StringVal(require.app.MinVersionString))
		}
		if len(require.Plugins) > 0 {
			pluginValues := make([]cty.Value, len(require.Plugins))
			for i, p := range require.Plugins {
				pluginValues[i] = cty.StringVal(typehelpers.SafeString(p))
			}
			requiresBody.SetAttributeValue("plugins", cty.ListVal(pluginValues))
		}
		if len(require.Mods) > 0 {
			for _, m := range require.Mods {
				modBody := requiresBody.AppendNewBlock("mod", []string{m.Name}).Body()
				modBody.SetAttributeValue("version", cty.StringVal(m.VersionString))
			}
		}
	}

	// load existing mod data and remove the mod definitions from it
	return os.WriteFile(app_specific.DefaultModFilePath(m.ModPath), f.Bytes(), 0644) //nolint:gosec // TODO: check file permission
}

func (m *Mod) HasDependentMods() bool {
	return m.Require != nil && len(m.Require.Mods) > 0
}

func (m *Mod) GetModDependency(modName string) *ModVersionConstraint {
	if m.Require == nil {
		return nil
	}
	return m.Require.GetModDependency(modName)
}

func (m *Mod) WalkResources(resourceFunc func(item HclResource) (bool, error)) error {
	return m.Resources.WalkResources(resourceFunc)
}

func (m *Mod) SetFilePath(modFilePath string) {
	m.modFilePath = modFilePath
}
func (m *Mod) GetFilePath() string {
	return m.modFilePath
}

// ValidateRequirements validates that the current steampipe CLI and the installed plugins is compatible with the mod
func (m *Mod) ValidateRequirements(pluginVersionMap *plugin.PluginVersionMap) []error {
	var validationErrors []error
	if err := m.validateAppVersion(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// if we have a plugin map, validate required plugins
	if pluginVersionMap != nil {
		validationErrors = append(validationErrors, m.validatePluginVersions(*pluginVersionMap)...)
	}
	return validationErrors
}

func (m *Mod) FilePath() string {
	return m.modFilePath
}

func (m *Mod) validateAppVersion() error {
	if m.Require == nil {
		return nil
	}
	return m.Require.validateAppVersion(m.Name())
}

func (m *Mod) validatePluginVersions(availablePlugins plugin.PluginVersionMap) []error {
	if m.Require == nil {
		return nil
	}

	return m.Require.validatePluginVersions(m.GetInstallCacheKey(), availablePlugins)
}

// CtyValue implements CtyValueProvider
func (m *Mod) CtyValue() (cty.Value, error) {
	return cty_helpers.GetCtyValue(m)
}

// GetInstallCacheKey returns the key used to find this mod in a workspace lock InstallCache
func (m *Mod) GetInstallCacheKey() string {
	// if the ModDependencyPath is set, this is a dependency mod - use that
	if m.DependencyPath != nil {
		return *m.DependencyPath
	}
	// otherwise use the short name
	return m.ShortName
}

// SetDependencyConfigFromPath sets DependencyPath, DependencyName and Version
func (m *Mod) SetDependencyConfigFromPath(dependencyPath string) error {
	// parse the dependency path to get the dependency name and version
	dependencyName, dependencyVersion, err := ParseModDependencyPath(dependencyPath)
	if err != nil {
		return err
	}
	m.DependencyPath = &dependencyPath
	m.DependencyName = dependencyName
	m.Version = dependencyVersion
	return nil
}

func (m *Mod) SetDependencyConfig(dependencyVersion *DependencyVersion, dependencyPath *string, dependencyName string) {
	m.DependencyPath = dependencyPath
	m.DependencyName = dependencyName
	m.Version = dependencyVersion
}

// RequireHasUnresolvedArgs returns whether the mod has any mod requirements which have unresolved args
// (this could be because the arg refers to a variable, meanin gwe need an additional parse phase
// to resolve the arg values)
func (m *Mod) RequireHasUnresolvedArgs() bool {
	if m.Require == nil {
		return false
	}
	for _, m := range m.Require.Mods {
		for _, a := range m.Args {
			if !a.IsKnown() {
				return true
			}
		}
	}
	return false
}

// GetConnectionDependsOn determines whether the mod depends on a connection
// if so the Database field will be a ConnectionDependency struct
func (m *Mod) GetConnectionDependsOn() []string {
	if m.Database != nil {
		if csp, ok := (m.Database).(connection.ConnectionDependency); ok {
			return []string{csp.GetConnectionString()}
		}
	}
	return nil
}

func (m *Mod) GetDefaultConnectionString(evalContext *hcl.EvalContext) (connection.ConnectionStringProvider, error) {
	if m.Database != nil {

		if cd, ok := (m.Database).(connection.ConnectionDependency); ok {
			return app_specific_connection.ConnectionStringFromConnectionName(evalContext, cd.GetConnectionString())
		}

		return m.Database, nil
	}
	// if no database is set on mod, use the default steampipe connection
	return connection.NewConnectionString(constants.DefaultSteampipeConnectionString), nil
}
