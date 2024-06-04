package modinstaller

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/xlab/treeprint"
)

const (
	VerbInstalled   = "Installed"
	VerbUninstalled = "Uninstalled"
	VerbUpgraded    = "Upgraded"
	VerbDowngraded  = "Downgraded"
	VerbPruned      = "Pruned"
)

var dryRunVerbs = map[string]string{
	VerbInstalled:   "Would install",
	VerbUninstalled: "Would uninstall",
	VerbUpgraded:    "Would upgrade",
	VerbDowngraded:  "Would downgrade",
	VerbPruned:      "Would prune",
}

func getVerb(verb string) string {
	if viper.GetBool(constants.ArgDryRun) {
		verb = dryRunVerbs[verb]
	}
	return verb
}

func BuildInstallSummary(installData *InstallData) string {

	// for now treat an install as update - we only install deps which are in the mod.sp but missing in the mod folder
	installCount, installedTreeString := getInstallationResultString(installData.Installed)
	uninstallCount, uninstalledTreeString := getInstallationResultString(installData.Uninstalled)
	upgradeCount, upgradeTreeString := getInstallationResultString(installData.Upgraded)
	downgradeCount, downgradeTreeString := getInstallationResultString(installData.Downgraded)

	var installString, upgradeString, downgradeString, uninstallString string
	if installCount > 0 {
		verb := getVerb(VerbInstalled)
		installString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, installCount, utils.Pluralize("mod", installCount), installedTreeString)
	}
	if uninstallCount > 0 {
		verb := getVerb(VerbUninstalled)
		uninstallString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, uninstallCount, utils.Pluralize("mod", uninstallCount), uninstalledTreeString)
	}
	if upgradeCount > 0 {
		verb := getVerb(VerbUpgraded)
		upgradeString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, upgradeCount, utils.Pluralize("mod", upgradeCount), upgradeTreeString)
	}
	if downgradeCount > 0 {
		verb := getVerb(VerbDowngraded)
		downgradeString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, downgradeCount, utils.Pluralize("mod", downgradeCount), downgradeTreeString)
	}

	if installCount+uninstallCount+upgradeCount+downgradeCount == 0 {
		if len(installData.Lock.InstallCache) == 0 {
			return "No mods are installed"
		}
		return "All targetted mods are up to date"
	}
	return fmt.Sprintf("%s%s%s%s", installString, upgradeString, downgradeString, uninstallString)
}

func getInstallationResultString(paths [][]string) (int, string) {
	count := len(paths)
	if count == 0 {
		return count, ""
	}
	// build tree
	var tree treeprint.Tree
	nodeMap := map[string]treeprint.Tree{}
	for _, path := range paths {

		var parentNode treeprint.Tree
		pathString := ""

		for _, segment := range path {

			if pathString != "" {
				pathString += "/"
			}
			pathString += segment

			// do we have a node for this path already?
			node, ok := nodeMap[pathString]
			if !ok {
				if parentNode == nil {
					tree = treeprint.NewWithRoot(segment)
					node = tree
				} else {
					node = parentNode.AddBranch(segment)
				}
				nodeMap[pathString] = node
			}
			parentNode = node
		}
	}

	treeString := tree.String()

	return count, treeString
}

func BuildUninstallSummary(installData *InstallData) string {
	uninstallCount, uninstalledTreeString := getInstallationResultString(installData.Uninstalled)
	if uninstallCount == 0 {
		return "Nothing uninstalled"
	}
	verb := getVerb(VerbUninstalled)
	return fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, uninstallCount, utils.Pluralize("mod", uninstallCount), uninstalledTreeString)
}
