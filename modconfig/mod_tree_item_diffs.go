package modconfig

import (
	"github.com/turbot/go-kit/helpers"
)

// ModTreeItemDiffs is a struct representing the differences between 2 DashboardTreeItems (of same type)
type ModTreeItemDiffs struct {
	Name              string
	Item              ModTreeItem
	ChangedProperties []string
	AddedItems        []string
	RemovedItems      []string
}

func (d *ModTreeItemDiffs) AddPropertyDiff(propertyName string) {
	if !helpers.StringSliceContains(d.ChangedProperties, propertyName) {
		d.ChangedProperties = append(d.ChangedProperties, propertyName)
	}
}

func (d *ModTreeItemDiffs) AddAddedItem(name string) {
	d.AddedItems = append(d.AddedItems, name)
}

func (d *ModTreeItemDiffs) AddRemovedItem(name string) {
	d.RemovedItems = append(d.RemovedItems, name)
}

func (d *ModTreeItemDiffs) PopulateChildDiffs(old ModTreeItem, new ModTreeItem) {
	// build map of child names
	oldChildMap := make(map[string]ModTreeItem)
	newChildMap := make(map[string]ModTreeItem)

	oldChildren := old.GetChildren()
	newChildren := new.GetChildren()

	for i, child := range oldChildren {
		// check for child ordering
		if i < len(newChildren) && newChildren[i].Name() != child.Name() {
			d.AddPropertyDiff("Children")
		}
		oldChildMap[child.Name()] = child
	}
	for _, child := range newChildren {
		newChildMap[child.Name()] = child
	}

	for childName /*prevChild*/ := range oldChildMap {
		if _ /*child*/, existInNew := newChildMap[childName]; !existInNew {
			d.AddRemovedItem(childName)
		}
		//else {
		// TODO INCOMPLETE sort out referencing https://github.com/turbot/pipe-fittings/issues/614
		// so this resource exists on old and new

		//// TACTICAL
		//// some child resources are not added to the mod but we must consider them for the diff
		//var childDiff = &ModTreeItemDiffs{}
		//switch t := child.(type) {
		//case *dashboard.DashboardWith:
		//	childDiff = t.Diff(prevChild.(*dashboard.DashboardWith))
		//case *dashboard.DashboardNode:
		//	childDiff = t.Diff(prevChild.(*dashboard.DashboardNode))
		//case *dashboard.DashboardEdge:
		//	childDiff = t.Diff(prevChild.(*dashboard.DashboardEdge))
		//}
		//if childDiff.HasChanges() {
		//	d.AddPropertyDiff("Children")
		//}

		//}
	}
	for childName := range newChildMap {
		if _, existsInOld := oldChildMap[childName]; !existsInOld {
			d.AddAddedItem(childName)
		}
	}
}

func (d *ModTreeItemDiffs) HasChanges() bool {
	return len(d.ChangedProperties)+
		len(d.AddedItems)+
		len(d.RemovedItems) > 0
}

func (d *ModTreeItemDiffs) Merge(other *ModTreeItemDiffs) {
	for _, added := range other.AddedItems {
		d.AddAddedItem(added)
	}
	for _, removed := range other.RemovedItems {
		d.AddRemovedItem(removed)
	}
	for _, changed := range other.ChangedProperties {
		d.AddPropertyDiff(changed)
	}
}
