package modinstaller

import (
	"os"
	"path/filepath"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/versionmap"
)

func (i *ModInstaller) Prune() (versionmap.DependencyVersionListMap, error) {
	unusedMods := i.installData.Lock.GetUnreferencedMods()
	// now delete any mod folders which are not in the lock file
	for name, versions := range unusedMods {
		for _, version := range versions {
			depPath := i.getDependencyDestPath(modconfig.BuildModDependencyPath(name, version))
			if err := i.deleteDependencyItem(depPath); err != nil {
				return nil, err
			}
		}
	}

	return unusedMods, nil
}

func (i *ModInstaller) deleteDependencyItem(depPath string) error {
	if err := os.RemoveAll(depPath); err != nil {
		return err
	}
	return i.deleteEmptyFolderTree(filepath.Dir(depPath))

}

func (i *ModInstaller) deleteEmptyFolderTree(folderPath string) error {
	// if the parent folder is empty, delete it
	err := os.Remove(folderPath)
	if err == nil {
		parent := filepath.Dir(folderPath)
		return i.deleteEmptyFolderTree(parent)
	}
	return nil
}
