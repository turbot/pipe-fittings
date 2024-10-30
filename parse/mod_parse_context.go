package parse

import (
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/inputvars"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/versionmap"
	"github.com/turbot/terraform-components/terraform"
	"github.com/zclconf/go-cty/cty"
)

const RootDependencyNode = "RootDependencyNode"

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
)

/*
	ReferenceTypeValueMap is the raw data used to build the evaluation context

When resolving hcl references like :
- query.q1
- var.v1
- mod1.query.my_query.sql

ReferenceTypeValueMap is keyed  by resource type, then by resource name
*/
type ReferenceTypeValueMap map[string]map[string]cty.Value

type ModParseContext struct {
	ParseContext

	// the mod which is currently being parsed
	CurrentMod *modconfig.Mod
	// the workspace lock data
	WorkspaceLock *versionmap.WorkspaceLock

	Flags       ParseModFlag
	ListOptions filehelpers.ListOptions

	// Variables are populated in an initial parse pass top we store them on the run context
	// so we can set them on the mod when we do the main parse

	// Variables is a tree of maps of the variables in the current mod and child dependency mods
	Variables *modconfig.ModVariableMap

	PipelingConnections map[string]connection.PipelingConnection

	ParentParseCtx *ModParseContext

	// stack of parent resources for the currently parsed block
	// (unqualified name)
	parents []string

	// map of resource children, keyed by parent unqualified name
	blockChildMap map[string][]string

	// map of top  level blocks, for easy checking
	topLevelBlocks map[*hcl.Block]struct{}
	// map of block names, keyed by a hash of the blopck
	blockNameMap map[string]string
	// map of ReferenceTypeValueMaps keyed by mod name
	// NOTE: all values from root mod are keyed with "local"
	referenceValues map[string]ReferenceTypeValueMap

	// a map of just the top level dependencies of the CurrentMod, keyed my full mod DependencyName (with no version)
	topLevelDependencyMods modconfig.ModMap
	// if we are loading dependency mod, this contains the details
	DependencyConfig *ModDependencyConfig
	modResources     modconfig.ModResources
	// map of late binding variable values
	// - this is added to the eval context if includeLateBindingResourcesInEvalContext is true
	lateBindingVars map[string]cty.Value

	// do we support late binding resources?
	supportLateBinding bool
	// if connections are early binding, this map contains the connection values
	connectionValueMap map[string]cty.Value

	// tactical: should temporary connections be added to the reference values?
	// this is a temporary solution until the 2 methods of determining runtime dependencies are merged
	includeLateBindingResourcesInEvalContext bool

	// mutex to control access to topLevelDependencyMods and modResources when asyncronously adding dependency mods
	depLock         sync.Mutex
	configValueMaps map[string]map[string]cty.Value
	decoderOptions  []DecoderOption
}

func NewModParseContext(workspaceLock *versionmap.WorkspaceLock, rootEvalPath string, opts ...ModParseContextOption) (*ModParseContext, error) {
	parseContext := NewParseContext(rootEvalPath)
	c := &ModParseContext{
		ParseContext: parseContext,

		WorkspaceLock: workspaceLock,

		topLevelDependencyMods: make(modconfig.ModMap),
		blockChildMap:          make(map[string][]string),
		blockNameMap:           make(map[string]string),
		// initialise reference maps - even though we later overwrite them
		referenceValues: map[string]ReferenceTypeValueMap{
			"local": make(ReferenceTypeValueMap),
		},
		lateBindingVars: make(map[string]cty.Value),
		// default to supporting late binding
		supportLateBinding: true,
		configValueMaps:    make(map[string]map[string]cty.Value),
	}

	// apply options
	for _, opt := range opts {
		opt(c)
	}
	// add root node - this will depend on all other nodes
	c.DependencyGraph = c.newDependencyGraph()

	// if we DO NOT support late binding resources, we need to build the connection value map now
	if !c.supportLateBinding {
		if err := c.buildConnectionValueMap(); err != nil {
			return nil, err
		}
	}
	c.RebuildEvalContext()

	return c, nil
}

func NewChildModParseContext(parent *ModParseContext, modVersion *versionmap.ResolvedVersionConstraint, rootEvalPath string) (*ModParseContext, error) {
	// create a child run context
	child, err := NewModParseContext(parent.WorkspaceLock, rootEvalPath,
		WithParseFlags(parent.Flags),
		WithListOptions(parent.ListOptions),
		WithLateBinding(parent.supportLateBinding),
		WithConnections(parent.PipelingConnections),
		WithDecoderOptions(parent.decoderOptions...),
		WithConfigValueMap(parent.configValueMaps))

	if err != nil {
		return nil, err
	}
	// copy our block types and exclusions
	child.blockTypes = parent.blockTypes
	child.blockTypeExclusions = parent.blockTypeExclusions

	// set the child's parent
	child.ParentParseCtx = parent
	// set the dependency config
	child.DependencyConfig = NewDependencyConfig(modVersion)
	// set variables if parent has any
	if parent.Variables != nil {
		childVars, ok := parent.Variables.DependencyVariables[modVersion.Name]
		if ok {
			child.Variables = childVars
			child.Variables.PopulatePublicVariables()
			child.AddVariablesToEvalContext()
		}
	}
	child.connectionValueMap = parent.connectionValueMap

	// ensure to inherit the value of includeLateBindingResourcesInEvalContext
	child.includeLateBindingResourcesInEvalContext = parent.includeLateBindingResourcesInEvalContext

	// if this is a filepath dependency, we need to exclude hidden files underneath the target filepath
	// (so we ignore any .steampipe or .powerpipe folders under the mod folder)
	if modVersion.FilePath != "" {
		child.ListOptions.Exclude = []string{
			fmt.Sprintf("%s/.*", modVersion.FilePath),
			fmt.Sprintf("%s/.*/**", modVersion.FilePath),
		}
	} else {
		// otherwise if this is a normal dependency modify the ListOptions to ensure we include hidden files - these are excluded by default
		child.ListOptions.Exclude = nil

	}
	return child, nil
}

func (m *ModParseContext) EnsureWorkspaceLock(mod *modconfig.Mod) error {
	// if the mod has dependencies, there must a workspace lock object in the run context
	// (mod MUST be the workspace mod, not a dependency, as we would hit this error as soon as we parse it)
	if mod.HasDependentMods() && (m.WorkspaceLock.Empty() || m.WorkspaceLock.Incomplete()) {
		// logger := fplog.Logger(m.RunCtx)
		// logger.Error("mod has dependencies but no workspace lock file found", "mod", mod.Name(), "m.HasDependentMods()", mod.HasDependentMods(), "m.WorkspaceLock.Empty()", m.WorkspaceLock.Empty(), "m.WorkspaceLock.Incomplete()", m.WorkspaceLock.Incomplete())
		return perr.BadRequestWithTypeAndMessage(perr.ErrorCodeDependencyFailure, "not all dependencies are installed - run '"+app_specific.AppName+" mod install'")
	}

	return nil
}

func (m *ModParseContext) PushParent(parent modconfig.ModTreeItem) {
	m.parents = append(m.parents, parent.GetUnqualifiedName())
}

func (m *ModParseContext) PopParent() string {
	n := len(m.parents) - 1
	res := m.parents[n]
	m.parents = m.parents[:n]
	return res
}

func (m *ModParseContext) PeekParent() string {
	if len(m.parents) == 0 {
		return m.CurrentMod.Name()
	}
	return m.parents[len(m.parents)-1]
}

// VariableValueCtyMap converts a map of variables to a map of the underlying cty value
// Note: if the variable type is a late binding type (i.e. PipelingConnection), DO NOT add to map
func VariableValueCtyMap(variables map[string]*modconfig.Variable, supportLateBinding bool) (ret, lateBindingVars, lateBindingVarDeps map[string]cty.Value) {
	lateBindingVarDeps = make(map[string]cty.Value)
	ret = make(map[string]cty.Value, len(variables))
	lateBindingVars = make(map[string]cty.Value, len(variables))
	for k, v := range variables {
		if supportLateBinding && v.IsLateBinding() {
			// if the variable is a late binding variable, build a cty value containing all referenced connections
			resourceNames, ok := ConnectionNamesValueFromCtyValue(v.Value)
			if ok {
				// add to late binding vars map
				lateBindingVars[v.GetShortName()] = v.Value
				lateBindingVarDeps[v.GetShortName()] = resourceNames
			}
		} else {
			ret[k] = v.Value
		}
	}
	return ret, lateBindingVars, lateBindingVarDeps
}

// AddInputVariableValues adds evaluated variables to the run context.
// This function is called for the root run context after loading all input variables
func (m *ModParseContext) AddInputVariableValues(inputVariables *modconfig.ModVariableMap) {
	utils.LogTime("AddInputVariableValues")
	defer utils.LogTime("AddInputVariableValues end")
	// store the variables
	m.Variables = inputVariables

	// now add variables into eval context
	m.AddVariablesToEvalContext()
}

func (m *ModParseContext) AddVariablesToEvalContext() {
	m.addRootVariablesToReferenceMap()
	m.addDependencyVariablesToReferenceMap()
	m.RebuildEvalContext()
}

// addRootVariablesToReferenceMap sets the Variables property
// and adds the variables to the referenceValues map (used to build the eval context)
func (m *ModParseContext) addRootVariablesToReferenceMap() {

	variables := m.Variables.RootVariables
	// write local variables directly into referenceValues map
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	varCtyMap, lateBindingVars, lateBindingVarDeps := VariableValueCtyMap(variables, m.supportLateBinding)
	m.referenceValues["local"]["var"] = varCtyMap
	// store the late binding vars in case we need to add them (to parse pipeline params for example)
	maps.Copy(m.lateBindingVars, lateBindingVars)
	// add late binding variables deps to reference values
	m.addLateBindingVariablesToReferenceValues(m.referenceValues["local"], lateBindingVarDeps)

}

// addDependencyVariablesToReferenceMap adds the dependency variables to the referenceValues map
// (used to build the eval context)
func (m *ModParseContext) addDependencyVariablesToReferenceMap() {
	// retrieve the resolved dependency versions for the parent mod
	resolvedVersions := m.WorkspaceLock.InstallCache[m.Variables.Mod.GetInstallCacheKey()]

	for depModName, depVars := range m.Variables.DependencyVariables {
		alias := resolvedVersions[depModName].Alias
		if m.referenceValues[alias] == nil {
			m.referenceValues[alias] = make(ReferenceTypeValueMap)
		}

		varCtyMap, lateBindingVars, lateBindingVarDeps := VariableValueCtyMap(depVars.RootVariables, m.supportLateBinding)
		m.referenceValues[alias]["var"] = varCtyMap
		if m.lateBindingVars == nil {
			m.lateBindingVars = make(map[string]cty.Value)
		}
		// store the late binding vars in case we need to add them (to parse pipeline params for example)
		maps.Copy(m.lateBindingVars, lateBindingVars)
		// add late binding variables deps to reference values
		m.addLateBindingVariablesToReferenceValues(m.referenceValues["local"], lateBindingVarDeps)

	}
}

// AddModResources is used to add mod resources to the eval context
func (m *ModParseContext) AddModResources(mod *modconfig.Mod) hcl.Diagnostics {
	if len(m.UnresolvedBlocks) > 0 {
		// should never happen
		panic("calling AddModResources on ModParseContext but there are unresolved blocks from a previous parse")
	}

	var diags hcl.Diagnostics
	moreDiags := m.storeResourceInReferenceValueMap(mod)
	diags = append(diags, moreDiags...)

	// do not add variables (as they have already been added)
	// if the resource is for a dependency mod, do not add locals
	shouldAdd := func(item modconfig.HclResource) bool {
		if item.GetBlockType() == schema.BlockTypeMod ||
			item.GetBlockType() == schema.BlockTypeVariable ||
			item.GetBlockType() == schema.BlockTypeLocals && item.(modconfig.ModItem).GetMod().GetShortName() != m.CurrentMod.ShortName {
			return false
		}
		return true
	}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// add all mod resources (except those excluded) into cty map
		if shouldAdd(item) {
			moreDiags := m.storeResourceInReferenceValueMap(item)
			diags = append(diags, moreDiags...)
		}
		// continue walking
		return true, nil
	}
	err := mod.WalkResources(resourceFunc)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "error walking mod resources",
			Detail:   err.Error(),
		})
		return diags
	}

	// rebuild the eval context
	m.RebuildEvalContext()
	return diags
}

func (m *ModParseContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	// put blocks into map as well
	m.topLevelBlocks = make(map[*hcl.Block]struct{}, len(m.blocks))
	for _, b := range content.Blocks {
		m.topLevelBlocks[b] = struct{}{}
	}
	m.ParseContext.SetDecodeContent(content, fileData)
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (m *ModParseContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	// TACTICAL if this is NOT a top level block, add the parent name to the block name
	// this is needed to avoid circular dependency errors if a nested block references
	// a top level block with the same name
	if !m.IsTopLevelBlock(block) {
		name = fmt.Sprintf("%s.%s", m.PeekParent(), name)
	}
	return m.ParseContext.AddDependencies(block, name, dependencies)
}

// ShouldCreateDefaultMod returns whether the flag is set to create a default mod if no mod definition exists
func (m *ModParseContext) ShouldCreateDefaultMod() bool {
	return m.Flags&CreateDefaultMod == CreateDefaultMod
}

// AddResource stores this resource as a variable to be added to the eval context.
func (m *ModParseContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := m.storeResourceInReferenceValueMap(resource)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// rebuild the eval context
	m.RebuildEvalContext()

	return nil
}

// GetMod finds the mod with given short name, looking only in first level dependencies
// this is used to resolve resource references
// specifically when the 'children' property of dashboards and benchmarks refers to resource in a dependency mod
func (m *ModParseContext) GetMod(modShortName string) *modconfig.Mod {
	if modShortName == m.CurrentMod.ShortName {
		return m.CurrentMod
	}
	// we need to iterate through dependency mods of the current mod
	key := m.CurrentMod.GetInstallCacheKey()
	deps := m.WorkspaceLock.InstallCache[key]
	for _, dep := range deps {
		depMod, ok := m.topLevelDependencyMods[dep.Name]
		if ok && depMod.ShortName == modShortName {
			return depMod
		}
	}
	return nil
}

func (m *ModParseContext) GetModResources() modconfig.ModResources {
	if m.modResources != nil {
		return m.modResources
	}

	m.setModResources()
	return m.modResources
}

func (m *ModParseContext) setModResources() {
	utils.LogTime(fmt.Sprintf("ModParseContext.setModResources %p", m))
	defer utils.LogTime(fmt.Sprintf("ModParseContext.setModResources %p end", m))

	// get a map of top level loaded dep mods
	deps := m.GetTopLevelDependencyMods()

	// use the current mod as the base resource map
	sourceModResources := make([]modconfig.ModResources, 0, len(deps)+1)

	sourceModResources = append(sourceModResources, m.CurrentMod.GetModResources())

	// merge in the top level resources of the dependency mods
	for _, dep := range deps {
		sourceModResources = append(sourceModResources, dep.GetModResources().TopLevelResources())
	}

	m.modResources = modconfig.NewModResources(m.CurrentMod, sourceModResources...)
}

func (m *ModParseContext) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return m.GetModResources().GetResource(parsedName)
}

// RebuildEvalContext the eval context from the cached reference values
func (m *ModParseContext) RebuildEvalContext() {
	// convert reference values to cty objects
	variables := make(map[string]cty.Value)

	// now for each mod add all the values
	for mod, modMap := range m.referenceValues {
		// TODO: this code is from steampipe, looks like there's a special treatment if the mod is named "local"?
		if mod == "local" {
			for k, v := range modMap {
				variables[k] = cty.ObjectVal(v)
			}
			continue
		}

		// mod map is map[string]map[string]cty.Value
		// for each element (i.e. map[string]cty.Value) convert to cty object
		refTypeMap := make(map[string]cty.Value)
		// TODO: this code is from steampipe, looks like there's a special treatment if the mod is named "local"?
		if mod == "local" {
			for k, v := range modMap {
				variables[k] = cty.ObjectVal(v)
			}
		} else {
			for refType, typeValueMap := range modMap {
				refTypeMap[refType] = cty.ObjectVal(typeValueMap)
			}
		}
		// now convert the referenceValues itself to a cty object
		variables[mod] = cty.ObjectVal(refTypeMap)
	}

	// add in any config value maps	(these will be values of global config items which may be referred to -
	// e.g. Flowpipe adds Notifiers)
	for name, valueMap := range m.configValueMaps {
		variables[name] = cty.ObjectVal(valueMap)
	}

	if !m.supportLateBinding && len(m.PipelingConnections) > 0 {
		variables[schema.BlockTypeConnection] = cty.ObjectVal(m.connectionValueMap)
	}
	// should we include connections
	if m.supportLateBinding && m.includeLateBindingResourcesInEvalContext {
		if len(m.PipelingConnections) > 0 {
			connMap := BuildTemporaryConnectionMapForEvalContext(m.PipelingConnections)
			variables[schema.BlockTypeConnection] = cty.ObjectVal(connMap)
		}

		if len(m.lateBindingVars) > 0 {
			var vars map[string]cty.Value
			if currentVars, gotVars := variables[schema.AttributeVar]; gotVars {
				vars = currentVars.AsValueMap()
			}
			if vars == nil {
				vars = make(map[string]cty.Value)
			}
			for k, v := range m.lateBindingVars {
				vars[k] = v
			}
			// convert back to cty object
			variables[schema.AttributeVar] = cty.ObjectVal(vars)
		}
	}
	// rebuild the eval context
	m.ParseContext.BuildEvalContext(variables)
}

// store the resource as a cty value in the reference valuemap
func (m *ModParseContext) storeResourceInReferenceValueMap(resource modconfig.HclResource) hcl.Diagnostics {
	// add resource to variable map
	ctyValue, diags := m.getResourceCtyValue(resource)
	if diags.HasErrors() {
		return diags
	}

	// add into the reference value map
	if diags := m.addReferenceValue(resource, ctyValue); diags.HasErrors() {
		return diags
	}

	// remove this resource from unparsed blocks
	n := resource.Name()
	delete(m.UnresolvedBlocks, n)

	return nil
}

// convert a HclResource into a cty value, taking into account nested structs
func (m *ModParseContext) getResourceCtyValue(resource modconfig.HclResource) (cty.Value, hcl.Diagnostics) {
	ctyValue, err := resource.(modconfig.CtyValueProvider).CtyValue()
	if err != nil {
		return cty.Zero, m.errToCtyValueDiags(resource, err)
	}
	// if this is a value map, merge in the values of base structs
	// if it is NOT a value map, the resource must have overridden CtyValue so do not merge base structs
	if ctyValue.Type().FriendlyName() != "object" {
		return ctyValue, nil
	}
	// TODO [node_reuse] fetch nested structs and serialise automatically https://github.com/turbot/steampipe/issues/2924
	valueMap := ctyValue.AsValueMap()
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}
	// get all nested structs (i.e. HclResourceImpl, ModTreeItemImpl and QueryProviderImpl - if this resource contains them)
	nestedStructs := resource.GetNestedStructs()
	for _, base := range nestedStructs {
		if err := m.mergeResourceCtyValue(base, valueMap); err != nil {
			return cty.Zero, m.errToCtyValueDiags(resource, err)
		}
	}

	return cty.ObjectVal(valueMap), nil
}

// merge the cty value of the given interface into valueMap
// (note: this mutates valueMap)
func (m *ModParseContext) mergeResourceCtyValue(resource modconfig.CtyValueProvider, valueMap map[string]cty.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in mergeResourceCtyValue: %s", helpers.ToError(r).Error())
		}
	}()
	ctyValue, err := resource.CtyValue()
	if err != nil {
		return err
	}
	if ctyValue == cty.Zero {
		return nil
	}
	// merge results
	for k, v := range ctyValue.AsValueMap() {
		valueMap[k] = v
	}
	return nil
}

func (m *ModParseContext) errToCtyValueDiags(resource modconfig.HclResource, err error) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("failed to convert resource '%s' to its cty value", resource.Name()),
		Detail:   err.Error(),
		Subject:  resource.GetDeclRange(),
	}}
}

func (m *ModParseContext) addReferenceValue(resource modconfig.HclResource, value cty.Value) hcl.Diagnostics {
	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to parse resource name %s", resource.Name()),
			Detail:   err.Error(),
			Subject:  resource.GetDeclRange(),
		}}
	}

	// TODO validate mod name clashes
	// TODO mod reserved names
	// TODO handle aliases

	key := parsedName.Name
	typeString := parsedName.ItemType

	// most resources will have a mod property - use this if available
	var mod *modconfig.Mod
	if modTreeItem, ok := resource.(modconfig.ModItem); ok {
		mod = modTreeItem.GetMod()
	}
	// fall back to current mod
	if mod == nil {
		mod = m.CurrentMod
	}

	modName := mod.ShortName
	if mod.ModPath == m.RootEvalPath {
		modName = "local"
	}
	variablesForMod, ok := m.referenceValues[modName]
	// do we have a map of reference values for this dep mod?
	if !ok {
		// no - create one
		variablesForMod = make(ReferenceTypeValueMap)
		m.referenceValues[modName] = variablesForMod
	}
	// do we have a map of reference values for this type
	variablesForType, ok := variablesForMod[typeString]
	if !ok {
		// no - create one
		variablesForType = make(map[string]cty.Value)
	}

	// some flowpipe items has 4 parts on it:
	// mod_name.integration.slack.my_slack_app
	//
	// this means the name needs to be broken into 2 parts to allow automatic references/hcl expresion valuation to work

	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		variablesForSubType := variablesForType[parts[0]]
		if variablesForSubType == cty.NilVal {
			variablesForSubType = cty.ObjectVal(map[string]cty.Value{})
		}

		mapValue := variablesForSubType.AsValueMap()
		if mapValue == nil {
			mapValue = make(map[string]cty.Value)
		}

		mapValue[parts[1]] = value

		variablesForSubType = cty.ObjectVal(mapValue)
		variablesForType[parts[0]] = variablesForSubType
	}

	// DO NOT update the cached cty values if the value already exists
	// this can happen in the case of variables where we initialise the context with values read from file
	// or passed on the command line,	// does this item exist in the map
	if _, ok := variablesForType[key]; !ok {
		variablesForType[key] = value
		variablesForMod[typeString] = variablesForType
		m.referenceValues[modName] = variablesForMod
	}

	return nil
}

func (m *ModParseContext) IsTopLevelBlock(block *hcl.Block) bool {
	_, isTopLevel := m.topLevelBlocks[block]
	return isTopLevel
}

func (m *ModParseContext) AddLoadedDependencyMod(mod *modconfig.Mod) {
	// lock the depLock as this is called async
	m.depLock.Lock()
	defer m.depLock.Unlock()

	m.topLevelDependencyMods[mod.DependencyName] = mod
	m.modResources.AddMaps(mod.GetModResources().TopLevelResources())
}

// GetTopLevelDependencyMods build a mod map of top level loaded dependencies, keyed by mod name
func (m *ModParseContext) GetTopLevelDependencyMods() modconfig.ModMap {
	return m.topLevelDependencyMods
}

func (m *ModParseContext) SetCurrentMod(mod *modconfig.Mod) error {
	m.CurrentMod = mod
	// populate the resource maps
	m.setModResources()
	// now we have the mod, load any arg values from the mod require - these will be passed to dependency mods
	return m.loadModRequireArgs()
}

// when reloading a mod dependency tree to resolve require args values, this function is called after each mod is loaded
// to load the require arg values and update the variable values
func (m *ModParseContext) loadModRequireArgs() error {
	//if we have not loaded variable definitions yet, do not load require args
	if m.Variables == nil {
		return nil
	}

	depModVarValues, err := inputvars.CollectVariableValuesFromModRequire(m.CurrentMod, m.WorkspaceLock)
	if err != nil {
		return err
	}
	if len(depModVarValues) == 0 {
		return nil
	}
	// if any mod require args have an unknown value, we have failed to resolve them - raise an error
	if err := m.validateModRequireValues(depModVarValues); err != nil {
		return err
	}

	// now update the variables map with the input values
	err = inputvars.SetVariableValues(depModVarValues, m.Variables)
	if err != nil {
		return err
	}

	// now add  overridden variables into eval context - in case the root mod references any dependency variable values
	m.AddVariablesToEvalContext()

	return nil
}

func (m *ModParseContext) validateModRequireValues(depModVarValues terraform.InputValues) error {
	if len(depModVarValues) == 0 {
		return nil
	}
	var missingVarExpressions []string
	requireBlock := m.getModRequireBlock()
	if requireBlock == nil {
		return fmt.Errorf("require args extracted but no require block found for %s", m.CurrentMod.Name())
	}

	for k, v := range depModVarValues {
		// if we successfully resolved this value, continue
		if v.Value.IsKnown() {
			continue
		}
		parsedVarName, err := modconfig.ParseResourceName(k)
		if err != nil {
			return err
		}

		// re-parse the require block manually to extract the range and unresolved arg value expression
		var errorString string
		errorString, err = m.getErrorStringForUnresolvedArg(parsedVarName, requireBlock)
		if err != nil {
			// if there was an error retrieving details, return less specific error string
			errorString = fmt.Sprintf("\"%s\"  (%s %s)", k, m.CurrentMod.Name(), m.CurrentMod.GetDeclRange().Filename)
		}

		missingVarExpressions = append(missingVarExpressions, errorString)
	}

	if errorCount := len(missingVarExpressions); errorCount > 0 {
		if errorCount == 1 {
			return fmt.Errorf("failed to resolve dependency mod argument value: %s", missingVarExpressions[0])
		}

		return fmt.Errorf("failed to resolve %d dependency mod arguments %s:\n\t%s", errorCount, utils.Pluralize("value", errorCount), strings.Join(missingVarExpressions, "\n\t"))
	}
	return nil
}

func (m *ModParseContext) getErrorStringForUnresolvedArg(parsedVarName *modconfig.ParsedResourceName, requireBlock *hclsyntax.Block) (_ string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	// which mod and variable is this is this for
	modShortName := parsedVarName.Mod
	varName := parsedVarName.Name
	var modDependencyName string
	// determine the mod dependency name as that is how it will be keyed in the require map
	for depName, modVersion := range m.WorkspaceLock.InstallCache[m.CurrentMod.GetInstallCacheKey()] {
		if modVersion.Alias == modShortName {
			modDependencyName = depName
			break
		}
	}

	// iterate through require blocks looking for mod blocks
	for _, b := range requireBlock.Body.Blocks {
		// only interested in mod blocks
		if b.Type != "mod" {
			continue
		}
		// if this is not the mod we're looking for, continue
		if b.Labels[0] != modDependencyName {
			continue
		}
		// now find the failed arg
		argsAttr, ok := b.Body.Attributes["args"]
		if !ok {
			return "", fmt.Errorf("no args block found for %s", modDependencyName)
		}
		// iterate over args looking for the correctly named item
		for _, a := range argsAttr.Expr.(*hclsyntax.ObjectConsExpr).Items {
			thisVarName, err := a.KeyExpr.Value(&hcl.EvalContext{})
			if err != nil {
				return "", err
			}

			// is this the var we are looking for?
			if thisVarName.AsString() != varName {
				continue
			}

			// this is the var, get the value expression
			expr, ok := a.ValueExpr.(*hclsyntax.ScopeTraversalExpr)
			if !ok {
				return "", fmt.Errorf("failed to get args details for %s", parsedVarName.ToResourceName())
			}
			// ok we have the expression - build the error string
			exprString := hclhelpers.TraversalAsString(expr.Traversal)
			r := expr.Range()
			sourceRange := fmt.Sprintf("%s:%d", r.Filename, r.Start.Line)
			res := fmt.Sprintf("\"%s = %s\" (%s %s)",
				parsedVarName.ToResourceName(),
				exprString,
				m.CurrentMod.Name(),
				sourceRange)
			return res, nil

		}
	}
	return "", fmt.Errorf("failed to get args details for %s", parsedVarName.ToResourceName())
}

func (m *ModParseContext) getModRequireBlock() *hclsyntax.Block {
	for _, b := range m.CurrentMod.GetResourceWithMetadataRemain().(*hclsyntax.Body).Blocks {
		if b.Type == schema.BlockTypeRequire {
			return b
		}
	}
	return nil

}

// LoadVariablesOnly returns whether we are ONLY loading variables
func (m *ModParseContext) LoadVariablesOnly() bool {
	if len(m.blockTypes) != 1 {
		return false
	}
	_, ok := m.blockTypes[schema.BlockTypeVariable]
	return ok
}

func (m *ModParseContext) SetBlockTypeExclusions(blockTypes ...string) {
	m.blockTypeExclusions = make(map[string]struct{}, len(blockTypes))
	for _, t := range blockTypes {
		m.blockTypeExclusions[t] = struct{}{}
	}
}

// SetIncludeLateBindingResources sets whether connections be included in the eval context
// and rebuilds the eval context
func (m *ModParseContext) SetIncludeLateBindingResources(include bool) {
	// this is only relevant if we support late binding resources
	if !m.supportLateBinding {
		return
	}
	m.includeLateBindingResourcesInEvalContext = include
	m.RebuildEvalContext()
}

func (m *ModParseContext) addLateBindingVariablesToReferenceValues(targetMap ReferenceTypeValueMap, varNames map[string]cty.Value) {
	if !m.supportLateBinding {
		return
	}
	// do we already have a map for late binding variables?
	if targetMap[constants.LateBindingVarsKey] == nil {
		targetMap[constants.LateBindingVarsKey] = map[string]cty.Value{}
	}

	maps.Copy(targetMap[constants.LateBindingVarsKey], varNames)
}

func (m *ModParseContext) buildConnectionValueMap() error {
	connectionMap := map[string]cty.Value{}
	for _, conn := range m.PipelingConnections {
		connType := conn.GetConnectionType()
		shortName := conn.GetShortName()
		// get map for type

		typeMapVal, ok := connectionMap[connType]
		var typeMap map[string]cty.Value
		if ok {
			typeMap = typeMapVal.AsValueMap()
		} else {
			typeMap = map[string]cty.Value{}
		}

		// add the connection
		ctyVal, err := conn.CtyValue()
		if err != nil {
			return err
		}
		typeMap[shortName] = ctyVal
		connectionMap[connType] = cty.ObjectVal(typeMap)
	}

	m.connectionValueMap = connectionMap
	return nil
}
