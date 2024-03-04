package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

// ResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type ResourceMaps struct {
	// the parent mod
	Mod *Mod

	Benchmarks            map[string]*Benchmark
	Controls              map[string]*Control
	Dashboards            map[string]*Dashboard
	DashboardCategories   map[string]*DashboardCategory
	DashboardCards        map[string]*DashboardCard
	DashboardCharts       map[string]*DashboardChart
	DashboardContainers   map[string]*DashboardContainer
	DashboardEdges        map[string]*DashboardEdge
	DashboardFlows        map[string]*DashboardFlow
	DashboardGraphs       map[string]*DashboardGraph
	DashboardHierarchies  map[string]*DashboardHierarchy
	DashboardImages       map[string]*DashboardImage
	DashboardInputs       map[string]map[string]*DashboardInput
	DashboardTables       map[string]*DashboardTable
	DashboardTexts        map[string]*DashboardText
	DashboardNodes        map[string]*DashboardNode
	GlobalDashboardInputs map[string]*DashboardInput
	Locals                map[string]*Local
	Variables             map[string]*Variable
	// all mods (including deps)
	Mods       map[string]*Mod
	Queries    map[string]*Query
	References map[string]*ResourceReference
	// map of snapshot paths, keyed by snapshot name
	Snapshots map[string]string

	// flowpipe
	Pipelines map[string]*Pipeline
	Triggers  map[string]*Trigger
}

func NewModResources(mod *Mod) *ResourceMaps {
	res := emptyModResources()
	res.Mod = mod
	res.Mods[mod.GetInstallCacheKey()] = mod
	return res
}

func NewSourceSnapshotModResources(snapshotPaths []string) *ResourceMaps {
	res := emptyModResources()
	res.AddSnapshots(snapshotPaths)
	return res
}

func emptyModResources() *ResourceMaps {
	return &ResourceMaps{
		Controls:              make(map[string]*Control),
		Benchmarks:            make(map[string]*Benchmark),
		Dashboards:            make(map[string]*Dashboard),
		DashboardCards:        make(map[string]*DashboardCard),
		DashboardCharts:       make(map[string]*DashboardChart),
		DashboardContainers:   make(map[string]*DashboardContainer),
		DashboardEdges:        make(map[string]*DashboardEdge),
		DashboardFlows:        make(map[string]*DashboardFlow),
		DashboardGraphs:       make(map[string]*DashboardGraph),
		DashboardHierarchies:  make(map[string]*DashboardHierarchy),
		DashboardImages:       make(map[string]*DashboardImage),
		DashboardInputs:       make(map[string]map[string]*DashboardInput),
		DashboardTables:       make(map[string]*DashboardTable),
		DashboardTexts:        make(map[string]*DashboardText),
		DashboardNodes:        make(map[string]*DashboardNode),
		DashboardCategories:   make(map[string]*DashboardCategory),
		GlobalDashboardInputs: make(map[string]*DashboardInput),
		Locals:                make(map[string]*Local),
		Mods:                  make(map[string]*Mod),
		Queries:               make(map[string]*Query),
		References:            make(map[string]*ResourceReference),
		Snapshots:             make(map[string]string),
		Variables:             make(map[string]*Variable),

		// Flowpipe
		Pipelines: make(map[string]*Pipeline),
		Triggers:  make(map[string]*Trigger),
	}
}

// QueryProviders returns a slice of all QueryProviders
func (m *ResourceMaps) QueryProviders() []QueryProvider {
	res := make([]QueryProvider, m.queryProviderCount())
	idx := 0
	f := func(item HclResource) (bool, error) {
		if queryProvider, ok := item.(QueryProvider); ok {
			res[idx] = queryProvider
			idx++
		}
		return true, nil
	}

	// resource func does not return an error
	_ = m.WalkResources(f)

	return res
}

// TopLevelResources returns a new ResourceMaps containing only top level resources (i.e. no dependencies)
func (m *ResourceMaps) TopLevelResources() *ResourceMaps {
	res := NewModResources(m.Mod)

	f := func(item HclResource) (bool, error) {
		if modItem, ok := item.(ModItem); ok {
			if mod := modItem.GetMod(); mod != nil && mod.FullName == m.Mod.FullName {
				// the only error we expect is a duplicate item error - ignore
				_ = res.AddResource(item)
			}
		}
		return true, nil
	}

	// resource func does not return an error
	_ = m.WalkResources(f)

	return res
}

func (m *ResourceMaps) Equals(other *ResourceMaps) bool {
	//TODO use cmp.Equals or similar
	if other == nil {
		return false
	}

	for name, query := range m.Queries {
		if otherQuery, ok := other.Queries[name]; !ok {
			return false
		} else if !query.Equals(otherQuery) {
			return false
		}
	}
	for name := range other.Queries {
		if _, ok := m.Queries[name]; !ok {
			return false
		}
	}

	for name, control := range m.Controls {
		if otherControl, ok := other.Controls[name]; !ok {
			return false
		} else if !control.Equals(otherControl) {
			return false
		}
	}
	for name := range other.Controls {
		if _, ok := m.Controls[name]; !ok {
			return false
		}
	}

	for name, benchmark := range m.Benchmarks {
		if otherBenchmark, ok := other.Benchmarks[name]; !ok {
			return false
		} else if !benchmark.Equals(otherBenchmark) {
			return false
		}
	}
	for name := range other.Benchmarks {
		if _, ok := m.Benchmarks[name]; !ok {
			return false
		}
	}

	for name, variable := range m.Variables {
		if otherVariable, ok := other.Variables[name]; !ok {
			return false
		} else if !variable.Equals(otherVariable) {
			return false
		}
	}
	for name := range other.Variables {
		if _, ok := m.Variables[name]; !ok {
			return false
		}
	}

	for name, pipeline := range m.Pipelines {
		if otherPipeline, ok := other.Pipelines[name]; !ok {
			return false
		} else if !pipeline.Equals(otherPipeline) {
			return false
		}
	}
	for name := range other.Pipelines {
		if _, ok := m.Pipelines[name]; !ok {
			return false
		}
	}

	// TODO: do we need integration & notifier here?

	for name, trigger := range m.Triggers {
		if otherTrigger, ok := other.Triggers[name]; !ok {
			return false
		} else if !trigger.Equals(otherTrigger) {
			return false
		}
	}
	for name := range other.Triggers {
		if _, ok := m.Triggers[name]; !ok {
			return false
		}
	}

	for name, dashboard := range m.Dashboards {
		if otherDashboard, ok := other.Dashboards[name]; !ok {
			return false
		} else if !dashboard.Equals(otherDashboard) {
			return false
		}
	}
	for name := range other.Dashboards {
		if _, ok := m.Dashboards[name]; !ok {
			return false
		}
	}

	for name, container := range m.DashboardContainers {
		if otherContainer, ok := other.DashboardContainers[name]; !ok {
			return false
		} else if !container.Equals(otherContainer) {
			return false
		}
	}
	for name := range other.DashboardContainers {
		if _, ok := m.DashboardContainers[name]; !ok {
			return false
		}
	}

	for name, cards := range m.DashboardCards {
		if otherCard, ok := other.DashboardCards[name]; !ok {
			return false
		} else if !cards.Equals(otherCard) {
			return false
		}
	}
	for name := range other.DashboardCards {
		if _, ok := m.DashboardCards[name]; !ok {
			return false
		}
	}

	for name, charts := range m.DashboardCharts {
		if otherChart, ok := other.DashboardCharts[name]; !ok {
			return false
		} else if !charts.Equals(otherChart) {
			return false
		}
	}
	for name := range other.DashboardCharts {
		if _, ok := m.DashboardCharts[name]; !ok {
			return false
		}
	}

	for name, flows := range m.DashboardFlows {
		if otherFlow, ok := other.DashboardFlows[name]; !ok {
			return false
		} else if !flows.Equals(otherFlow) {
			return false
		}
	}
	for name := range other.DashboardFlows {
		if _, ok := m.DashboardFlows[name]; !ok {
			return false
		}
	}

	for name, flows := range m.DashboardGraphs {
		if otherFlow, ok := other.DashboardGraphs[name]; !ok {
			return false
		} else if !flows.Equals(otherFlow) {
			return false
		}
	}
	for name := range other.DashboardGraphs {
		if _, ok := m.DashboardGraphs[name]; !ok {
			return false
		}
	}

	for name, hierarchies := range m.DashboardHierarchies {
		if otherHierarchy, ok := other.DashboardHierarchies[name]; !ok {
			return false
		} else if !hierarchies.Equals(otherHierarchy) {
			return false
		}
	}

	for name := range other.DashboardNodes {
		if _, ok := m.DashboardNodes[name]; !ok {
			return false
		}
	}

	for name := range other.DashboardEdges {
		if _, ok := m.DashboardEdges[name]; !ok {
			return false
		}
	}
	for name := range other.DashboardCategories {
		if _, ok := m.DashboardCategories[name]; !ok {
			return false
		}
	}

	for name, images := range m.DashboardImages {
		if otherImage, ok := other.DashboardImages[name]; !ok {
			return false
		} else if !images.Equals(otherImage) {
			return false
		}
	}
	for name := range other.DashboardImages {
		if _, ok := m.DashboardImages[name]; !ok {
			return false
		}
	}

	for name, input := range m.GlobalDashboardInputs {
		if otherInput, ok := other.GlobalDashboardInputs[name]; !ok {
			return false
		} else if !input.Equals(otherInput) {
			return false
		}
	}
	for name := range other.DashboardInputs {
		if _, ok := m.DashboardInputs[name]; !ok {
			return false
		}
	}

	for dashboardName, inputsForDashboard := range m.DashboardInputs {
		if otherInputsForDashboard, ok := other.DashboardInputs[dashboardName]; !ok {
			return false
		} else {

			for name, input := range inputsForDashboard {
				if otherInput, ok := otherInputsForDashboard[name]; !ok {
					return false
				} else if !input.Equals(otherInput) {
					return false
				}
			}
			for name := range otherInputsForDashboard {
				if _, ok := inputsForDashboard[name]; !ok {
					return false
				}
			}

		}
	}
	for name := range other.DashboardInputs {
		if _, ok := m.DashboardInputs[name]; !ok {
			return false
		}
	}

	for name, table := range m.DashboardTables {
		if otherTable, ok := other.DashboardTables[name]; !ok {
			return false
		} else if !table.Equals(otherTable) {
			return false
		}
	}
	for name, category := range m.DashboardCategories {
		if otherCategory, ok := other.DashboardCategories[name]; !ok {
			return false
		} else if !category.Equals(otherCategory) {
			return false
		}
	}
	for name := range other.DashboardTables {
		if _, ok := m.DashboardTables[name]; !ok {
			return false
		}
	}

	for name, text := range m.DashboardTexts {
		if otherText, ok := other.DashboardTexts[name]; !ok {
			return false
		} else if !text.Equals(otherText) {
			return false
		}
	}
	for name := range other.DashboardTexts {
		if _, ok := m.DashboardTexts[name]; !ok {
			return false
		}
	}

	for name, reference := range m.References {
		if otherReference, ok := other.References[name]; !ok {
			return false
		} else if !reference.Equals(otherReference) {
			return false
		}
	}

	for name := range other.References {
		if _, ok := m.References[name]; !ok {
			return false
		}
	}

	for name := range other.Locals {
		if _, ok := m.Locals[name]; !ok {
			return false
		}
	}
	return true
}

// GetResource tries to find a resource with the given name in the ResourceMaps
// NOTE: this does NOT support inputs, which are NOT uniquely named in a mod
func (m *ResourceMaps) GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool) {
	modName := parsedName.Mod
	if modName == "" {
		modName = m.Mod.ShortName
	}
	longName := fmt.Sprintf("%s.%s.%s", modName, parsedName.ItemType, parsedName.Name)

	// NOTE: we could use WalkResources, but this is quicker

	switch parsedName.ItemType {
	case schema.BlockTypeBenchmark:
		resource, found = m.Benchmarks[longName]
	case schema.BlockTypeControl:
		resource, found = m.Controls[longName]
	case schema.BlockTypeDashboard:
		resource, found = m.Dashboards[longName]
	case schema.BlockTypeCard:
		resource, found = m.DashboardCards[longName]
	case schema.BlockTypeCategory:
		resource, found = m.DashboardCategories[longName]
	case schema.BlockTypeChart:
		resource, found = m.DashboardCharts[longName]
	case schema.BlockTypeContainer:
		resource, found = m.DashboardContainers[longName]
	case schema.BlockTypeEdge:
		resource, found = m.DashboardEdges[longName]
	case schema.BlockTypeFlow:
		resource, found = m.DashboardFlows[longName]
	case schema.BlockTypeGraph:
		resource, found = m.DashboardGraphs[longName]
	case schema.BlockTypeHierarchy:
		resource, found = m.DashboardHierarchies[longName]
	case schema.BlockTypeImage:
		resource, found = m.DashboardImages[longName]
	case schema.BlockTypeNode:
		resource, found = m.DashboardNodes[longName]
	case schema.BlockTypeTable:
		resource, found = m.DashboardTables[longName]
	case schema.BlockTypeText:
		resource, found = m.DashboardTexts[longName]
	case schema.BlockTypeInput:
		// this function only supports global inputs
		// if the input has a parent dashboard, you must use GetDashboardInput
		resource, found = m.GlobalDashboardInputs[longName]
	case schema.BlockTypeQuery:
		resource, found = m.Queries[longName]
	// note the special case for variables - "var" rather than "variable"
	case schema.AttributeVar:
		resource, found = m.Variables[longName]
	case schema.BlockTypePipeline:
		resource, found = m.Pipelines[longName]
	case schema.BlockTypeTrigger:
		resource, found = m.Triggers[longName]
	case schema.BlockTypeMod:
		for _, mod := range m.Mods {
			if mod.ShortName == parsedName.Name {
				resource = mod
				found = true
				break
			}
		}

	}
	return resource, found
}

func (m *ResourceMaps) PopulateReferences() {
	// only populate references if introspection is enabled
	switch viper.GetString(constants.ArgIntrospection) {
	case constants.IntrospectionInfo:
		m.References = make(map[string]*ResourceReference)

		resourceFunc := func(resource HclResource) (bool, error) {
			if resourceWithMetadata, ok := resource.(ResourceWithMetadata); ok {
				for _, ref := range resourceWithMetadata.GetReferences() {
					m.References[ref.String()] = ref
				}

				// if this resource is a RuntimeDependencyProvider, add references from any 'withs'
				if nep, ok := resource.(NodeAndEdgeProvider); ok {
					m.populateNodeEdgeProviderRefs(nep)
				} else if rdp, ok := resource.(RuntimeDependencyProvider); ok {
					m.populateWithRefs(resource.GetUnqualifiedName(), rdp, getWithRoot(rdp))
				}
			}

			// continue walking
			return true, nil
		}
		// resource func does not return an error
		_ = m.WalkResources(resourceFunc)
	}
}

// populate references for any nodes/edges which have reference a 'with'
func (m *ResourceMaps) populateNodeEdgeProviderRefs(nep NodeAndEdgeProvider) {
	var withRoots = map[string]WithProvider{}
	for _, n := range nep.GetNodes() {
		// lazy populate with-root
		// (build map keyed by parent
		// - in theory if we inherit some nodes from base, they may have different parents)
		parent := n.GetParents()[0]
		if withRoots[parent.Name()] == nil && len(n.GetRuntimeDependencies()) > 0 {
			withRoots[parent.Name()] = getWithRoot(n)
		}
		m.populateWithRefs(nep.GetUnqualifiedName(), n, withRoots[parent.Name()])
	}
	for _, e := range nep.GetEdges() {
		// lazy populate with root
		parent := e.GetParents()[0]
		if withRoots[parent.Name()] == nil && len(e.GetRuntimeDependencies()) > 0 {
			withRoots[parent.Name()] = getWithRoot(e)
		}

		m.populateWithRefs(nep.GetUnqualifiedName(), e, withRoots[parent.Name()])
	}
}

// populate references for any 'with' blocks referenced by the RuntimeDependencyProvider
func (m *ResourceMaps) populateWithRefs(name string, rdp RuntimeDependencyProvider, withRoot WithProvider) {
	// unexpected but behave nicely
	if withRoot == nil {
		return
	}
	for _, r := range rdp.GetRuntimeDependencies() {
		if r.PropertyPath.ItemType == schema.BlockTypeWith {
			// find the with
			w, ok := withRoot.GetWith(r.PropertyPath.ToResourceName())
			if ok {
				for _, withRef := range w.References {
					// build a new reference changing the 'from' to the NodeAndEdgeProvider
					ref := withRef.CloneWithNewFrom(name)
					m.References[ref.String()] = ref
				}
			}
		}
	}
}

// search up the tree to find the root resource which will host any referenced 'withs'
// this will either be a dashboard ot a NodeEdgeProvider
func getWithRoot(rdp RuntimeDependencyProvider) WithProvider {
	var withRoot, _ = rdp.(WithProvider)
	// get the root resource which 'owns' any withs
	// (if our parent is the Mod, we are the root resource, otherwise traverse up until we find the mod
	parent := rdp.GetParents()[0]

	for parent.BlockType() != schema.BlockTypeMod {
		if wp, ok := parent.(WithProvider); ok {
			withRoot = wp
		}
		parent = parent.GetParents()[0]
	}
	return withRoot
}

func (m *ResourceMaps) Empty() bool {
	return len(m.Mods)+
		len(m.Queries)+
		len(m.Controls)+
		len(m.Benchmarks)+
		len(m.Variables)+
		len(m.Dashboards)+
		len(m.DashboardContainers)+
		len(m.DashboardCards)+
		len(m.DashboardCharts)+
		len(m.DashboardFlows)+
		len(m.DashboardGraphs)+
		len(m.DashboardHierarchies)+
		len(m.DashboardNodes)+
		len(m.DashboardEdges)+
		len(m.DashboardCategories)+
		len(m.DashboardImages)+
		len(m.DashboardInputs)+
		len(m.DashboardTables)+
		len(m.DashboardTexts)+
		len(m.References) == 0
}

// this is used to create an optimized ResourceMaps containing only the queries which will be run
//
//nolint:unused // TODO: check this unused property
func (m *ResourceMaps) addControlOrQuery(provider QueryProvider) {
	switch p := provider.(type) {
	case *Query:
		if p != nil {
			m.Queries[p.FullName] = p
		}
	case *Control:
		if p != nil {
			m.Controls[p.FullName] = p
		}
	}
}

// WalkResources calls resourceFunc for every resource in the mod
// if any resourceFunc returns false or an error, return immediately
func (m *ResourceMaps) WalkResources(resourceFunc func(item HclResource) (bool, error)) error {
	for _, r := range m.Mods {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Benchmarks {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Controls {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Dashboards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCards {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCategories {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardCharts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardContainers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardEdges {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardFlows {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardGraphs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardHierarchies {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardImages {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, inputsForDashboard := range m.DashboardInputs {
		for _, r := range inputsForDashboard {
			if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
				return err
			}
		}
	}
	for _, r := range m.DashboardNodes {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardTables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.DashboardTexts {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.GlobalDashboardInputs {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Locals {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Queries {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	// we cannot walk source snapshots as they are not a HclResource
	for _, r := range m.Variables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	for _, r := range m.Pipelines {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	for _, r := range m.Triggers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	return nil
}

func (m *ResourceMaps) AddResource(item HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		if existing, ok := m.Queries[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Queries[name] = r

	case *Control:
		name := r.Name()
		if existing, ok := m.Controls[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Controls[name] = r

	case *Benchmark:
		name := r.Name()
		if existing, ok := m.Benchmarks[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Benchmarks[name] = r

	case *Dashboard:
		name := r.Name()
		if existing, ok := m.Dashboards[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Dashboards[name] = r

	case *DashboardContainer:
		name := r.Name()
		if existing, ok := m.DashboardContainers[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardContainers[name] = r

	case *DashboardCard:
		name := r.Name()
		if existing, ok := m.DashboardCards[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		} else {
			m.DashboardCards[name] = r
		}

	case *DashboardChart:
		name := r.Name()
		if existing, ok := m.DashboardCharts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardCharts[name] = r

	case *DashboardFlow:
		name := r.Name()
		if existing, ok := m.DashboardFlows[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardFlows[name] = r

	case *DashboardGraph:
		name := r.Name()
		if existing, ok := m.DashboardGraphs[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardGraphs[name] = r

	case *DashboardHierarchy:
		name := r.Name()
		if existing, ok := m.DashboardHierarchies[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardHierarchies[name] = r

	case *DashboardNode:
		name := r.Name()
		if existing, ok := m.DashboardNodes[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardNodes[name] = r

	case *DashboardEdge:
		name := r.Name()
		if existing, ok := m.DashboardEdges[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardEdges[name] = r

	case *DashboardCategory:
		name := r.Name()
		if existing, ok := m.DashboardCategories[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardCategories[name] = r

	case *DashboardImage:
		name := r.Name()
		if existing, ok := m.DashboardImages[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardImages[name] = r

	case *DashboardInput:
		// if input has a dashboard asssigned, add to DashboardInputs
		name := r.Name()
		if dashboardName := r.DashboardName; dashboardName != "" {
			inputsForDashboard := m.DashboardInputs[dashboardName]
			if inputsForDashboard == nil {
				inputsForDashboard = make(map[string]*DashboardInput)
				m.DashboardInputs[dashboardName] = inputsForDashboard
			}
			// no need to check for dupes as we have already checked before adding the input to th m od
			inputsForDashboard[name] = r
			break
		}

		// so Dashboard Input must be global
		if existing, ok := m.GlobalDashboardInputs[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.GlobalDashboardInputs[name] = r

	case *DashboardTable:
		name := r.Name()
		if existing, ok := m.DashboardTables[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardTables[name] = r

	case *DashboardText:
		name := r.Name()
		if existing, ok := m.DashboardTexts[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.DashboardTexts[name] = r

	case *Variable:
		name := r.Name()
		if existing, ok := m.Variables[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Variables[name] = r

	case *Local:
		name := r.Name()
		if existing, ok := m.Locals[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Locals[name] = r

	case *Pipeline:
		name := r.Name()
		if existing, ok := m.Pipelines[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Pipelines[name] = r

	case *Trigger:
		name := r.Name()
		if existing, ok := m.Triggers[name]; ok {
			diags = append(diags, checkForDuplicate(existing, item)...)
			break
		}
		m.Triggers[name] = r
	}

	return diags
}

func (m *ResourceMaps) AddSnapshots(snapshotPaths []string) {
	for _, snapshotPath := range snapshotPaths {
		snapshotName := fmt.Sprintf("snapshot.%s", utils.FilenameNoExtension(snapshotPath))
		m.Snapshots[snapshotName] = snapshotPath
	}
}

func (m *ResourceMaps) Merge(others []*ResourceMaps) *ResourceMaps {
	res := NewModResources(m.Mod)
	sourceMaps := append([]*ResourceMaps{m}, others...)

	for _, source := range sourceMaps {
		for k, v := range source.Benchmarks {
			res.Benchmarks[k] = v
		}
		for k, v := range source.Controls {
			res.Controls[k] = v
		}
		for k, v := range source.Dashboards {
			res.Dashboards[k] = v
		}
		for k, v := range source.DashboardContainers {
			res.DashboardContainers[k] = v
		}
		for k, v := range source.DashboardCards {
			res.DashboardCards[k] = v
		}
		for k, v := range source.DashboardCategories {
			res.DashboardCategories[k] = v
		}
		for k, v := range source.DashboardCharts {
			res.DashboardCharts[k] = v
		}
		for k, v := range source.DashboardEdges {
			res.DashboardEdges[k] = v
		}
		for k, v := range source.DashboardFlows {
			res.DashboardFlows[k] = v
		}
		for k, v := range source.DashboardGraphs {
			res.DashboardGraphs[k] = v
		}
		for k, v := range source.DashboardHierarchies {
			res.DashboardHierarchies[k] = v
		}
		for k, v := range source.DashboardNodes {
			res.DashboardNodes[k] = v
		}
		for k, v := range source.DashboardImages {
			res.DashboardImages[k] = v
		}
		for k, v := range source.DashboardInputs {
			res.DashboardInputs[k] = v
		}
		for k, v := range source.DashboardTables {
			res.DashboardTables[k] = v
		}
		for k, v := range source.DashboardTexts {
			res.DashboardTexts[k] = v
		}
		for k, v := range source.GlobalDashboardInputs {
			res.GlobalDashboardInputs[k] = v
		}
		for k, v := range source.Locals {
			res.Locals[k] = v
		}
		for k, v := range source.Mods {
			res.Mods[k] = v
		}
		for k, v := range source.Queries {
			res.Queries[k] = v
		}
		for k, v := range source.Snapshots {
			res.Snapshots[k] = v
		}
		for k, v := range source.Pipelines {
			res.Pipelines[k] = v
		}
		for k, v := range source.Triggers {
			res.Triggers[k] = v
		}
		for k, v := range source.Variables {
			// TODO check why this was necessary and test variables thoroughly
			// NOTE: only include variables from root mod  - we add in the others separately
			//if v.Mod.FullName == m.Mod.FullName {
			res.Variables[k] = v
			//}
		}
	}

	return res
}

func (m *ResourceMaps) queryProviderCount() int {
	numDashboardInputs := 0
	for _, inputs := range m.DashboardInputs {
		numDashboardInputs += len(inputs)
	}

	numItems :=
		len(m.Controls) +
			len(m.DashboardCards) +
			len(m.DashboardCharts) +
			len(m.DashboardEdges) +
			len(m.DashboardFlows) +
			len(m.DashboardGraphs) +
			len(m.DashboardHierarchies) +
			len(m.DashboardImages) +
			numDashboardInputs +
			len(m.DashboardNodes) +
			len(m.DashboardTables) +
			len(m.GlobalDashboardInputs) +
			len(m.Queries)
	return numItems
}
