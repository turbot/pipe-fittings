package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/utils"
)

// BuildResourceTree builds the control tree structure by setting the parent property for each control and benchmark
// NOTE: this also builds the sorted benchmark list
func (m *Mod) BuildResourceTree(loadedDependencyMods ModMap) (err error) {
	utils.LogTime(fmt.Sprintf("BuildResourceTree %s start", m.Name()))
	defer utils.LogTime(fmt.Sprintf("BuildResourceTree %s end", m.Name()))
	defer func() {
		if err == nil {
			err = m.validateResourceTree()
		}
	}()

	// build lookup of Children and parents
	childrenLookup, err := m.getChildParentsLookup()
	if err != nil {
		return err
	}

	if err := m.addResourcesIntoTree(m, childrenLookup); err != nil {
		return err
	}

	if !m.HasDependentMods() {
		return nil
	}
	// add dependent mods into tree
	for _, requiredMod := range m.Require.Mods {
		// find this mod in installed dependency mods
		depMod, ok := loadedDependencyMods[requiredMod.Name]
		if !ok {
			return fmt.Errorf("dependency mod %s is not loaded", requiredMod.Name)
		}

		if err := m.addResourcesIntoTree(depMod, childrenLookup); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mod) getChildParentsLookup() (map[string][]ModTreeItem, error) {
	// build lookup of all Children
	childrenLookup := make(map[string][]ModTreeItem)
	resourceFunc := func(parent HclResource) (bool, error) {
		if treeItem, ok := parent.(ModTreeItem); ok {
			for _, child := range treeItem.GetChildren() {
				childrenLookup[child.Name()] = append(childrenLookup[child.Name()], treeItem)
			}
		}
		// continue walking
		return true, nil
	}
	err := m.Resources.WalkResources(resourceFunc)
	if err != nil {
		return nil, err
	}
	return childrenLookup, nil
}

// add all resource in sourceMod into _our_ resource tree
func (m *Mod) addResourcesIntoTree(sourceMod *Mod, childParentLookup map[string][]ModTreeItem) error {
	utils.LogTime(fmt.Sprintf("addResourcesIntoTree %s source %s start", m.Name(), sourceMod.Name()))
	defer utils.LogTime(fmt.Sprintf("addResourcesIntoTree %s source %s end", m.Name(), sourceMod.Name()))

	var leafNodes []ModTreeItem
	var err error

	resourceFunc := func(item HclResource) (bool, error) {
		// skip mods
		if _, ok := item.(*Mod); ok {
			return true, nil
		}

		if treeItem, ok := item.(ModTreeItem); ok {
			// NOTE: add resource into _our_ resource tree, i.e. mod 'm'
			if err = m.addItemIntoResourceTree(treeItem, childParentLookup); err != nil {
				// stop walking
				return false, err
			}
			if len(treeItem.GetChildren()) == 0 {
				leafNodes = append(leafNodes, treeItem)
			}
		}
		// continue walking
		return true, nil
	}

	// iterate through all resources in source mod
	err = sourceMod.WalkResources(resourceFunc)
	if err != nil {
		return err
	}

	// now initialise all Paths properties
	for _, l := range leafNodes {
		l.SetPaths()
	}

	return nil
}

func (m *Mod) addItemIntoResourceTree(item ModTreeItem, childParentLookup map[string][]ModTreeItem) error {
	parents := childParentLookup[item.Name()]
	if len(parents) == 0 {
		parents = []ModTreeItem{m}
	}
	for _, p := range parents {
		// if we are the parent, add as a child
		if err := item.AddParent(p); err != nil {
			return err
		}
		if p == m {
			m.Children = append(m.Children, item)
		}
	}

	return nil
}

// check whether a resource with the same name has already been added to the mod
// (it is possible to add the same resource to a mod more than once as the parent resource
// may have dependency errors and so be decoded again)
func CheckForDuplicate(existing, new HclResource) hcl.Diagnostics {
	if existing.GetDeclRange().String() == new.GetDeclRange().String() {
		// decl range is the same - this is the same resource - allowable
		return nil
	}
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Mod defines more than one resource named '%s'", new.Name()),
		Detail:   fmt.Sprintf("\n- %s\n- %s", existing.GetDeclRange(), new.GetDeclRange()),
	}}
}

func (m *Mod) AddResource(item HclResource) hcl.Diagnostics {
	return m.Resources.AddResource(item)
}
