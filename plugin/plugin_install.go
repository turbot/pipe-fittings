package plugin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/pipe-fittings/utils"
)

type PluginInstallReports []*PluginInstallReport

// making the type compatible with sort.Interface so that we can use the sort package utilities
func (i PluginInstallReports) Len() int                 { return len(i) }
func (i PluginInstallReports) Swap(lIdx, rIdx int)      { i[lIdx], i[rIdx] = i[rIdx], i[lIdx] }
func (i PluginInstallReports) Less(lIdx, rIdx int) bool { return i[lIdx].Plugin < i[rIdx].Plugin }

type PluginInstallReport struct {
	Skipped        bool
	Plugin         string
	SkipReason     string
	DocURL         string
	Version        string
	IsUpdateReport bool
}

func (i *PluginInstallReport) skipString() string {
	ref := ociinstaller.NewImageRef(i.Plugin)
	_, name, constraint := ref.GetOrgNameAndStream()

	return fmt.Sprintf("Plugin:   %s\nReason:   %s", fmt.Sprintf("%s@%s", name, constraint), i.SkipReason)
}

func (i *PluginInstallReport) installString() string {
	thisReport := []string{}
	if i.IsUpdateReport {
		thisReport = append(
			thisReport,
			fmt.Sprintf("Updated plugin: %s%s", constants.Bold(i.Plugin), i.Version),
		)
		if len(i.DocURL) > 0 {
			thisReport = append(
				thisReport,
				fmt.Sprintf("Documentation:  %s", i.DocURL),
			)
		}
	} else {
		thisReport = append(
			thisReport,
			fmt.Sprintf("Installed plugin: %s%s", constants.Bold(i.Plugin), i.Version),
		)
		if len(i.DocURL) > 0 {
			thisReport = append(
				thisReport,
				fmt.Sprintf("Documentation:    %s", i.DocURL),
			)
		}
	}

	return strings.Join(thisReport, "\n")
}

func (i *PluginInstallReport) String() string {
	if !i.Skipped {
		return i.installString()
	} else {
		return i.skipString()
	}
}

// PrintInstallReports Prints out the installation reports onto the console
func PrintInstallReports(reports PluginInstallReports, isUpdateReport bool) {
	installedOrUpdated := PluginInstallReports{}
	canBeInstalled := PluginInstallReports{}
	canBeUpdated := PluginInstallReports{}
	notFound := PluginInstallReports{}

	for _, report := range reports {
		report.IsUpdateReport = isUpdateReport
		if !report.Skipped {
			installedOrUpdated = append(installedOrUpdated, report)
		} else if report.SkipReason == constants.InstallMessagePluginNotInstalled {
			canBeInstalled = append(canBeInstalled, report)
		} else if report.SkipReason == constants.InstallMessagePluginAlreadyInstalled {
			canBeUpdated = append(canBeUpdated, report)
		} else if report.SkipReason == constants.InstallMessagePluginNotFound {
			notFound = append(notFound, report)
		}
	}

	// sort the report
	sort.Stable(reports)
	// sort the individual chunks
	sort.Stable(installedOrUpdated)
	sort.Stable(canBeInstalled)
	sort.Stable(canBeUpdated)
	sort.Stable(notFound)

	if len(installedOrUpdated) > 0 {
		//nolint:forbidigo // acceptable
		fmt.Println()
		asString := []string{}
		for _, report := range installedOrUpdated {
			asString = append(asString, report.installString())
		}
		//nolint:forbidigo // acceptable
		fmt.Println(strings.Join(asString, "\n\n"))
	}

	if len(installedOrUpdated) < len(reports) {
		installSkipReports := []string{}
		for _, report := range reports {
			showReport := true
			if report.SkipReason == constants.InstallMessagePluginAlreadyInstalled || report.SkipReason == constants.InstallMessagePluginLatestAlreadyInstalled {
				showReport = false
			}
			if report.Skipped && showReport {
				installSkipReports = append(installSkipReports, report.skipString())
			}
		}

		skipCount := len(installSkipReports)
		if (len(installSkipReports)) > 0 {
			//nolint:forbidigo // acceptable
			fmt.Printf(
				"\nSkipped the following %s:\n\n%s\n",
				utils.Pluralize("plugin", skipCount),
				strings.Join(installSkipReports, "\n\n"),
			)
		}

		if len(canBeInstalled) > 0 {
			pluginList := []string{}
			for _, r := range canBeInstalled {
				pluginList = append(pluginList, r.Plugin)
			}
			//nolint:forbidigo // acceptable
			fmt.Println()
			//nolint:forbidigo // acceptable
			fmt.Printf(
				"To install %s which %s not installed, please run %s\n",
				utils.Pluralize("plugin", len(canBeInstalled)),
				utils.Pluralize("is", len(canBeInstalled)),
				constants.Bold(fmt.Sprintf(
					"steampipe plugin install %s",
					strings.Join(pluginList, " "),
				)),
			)
		}

		if len(canBeUpdated) > 0 {
			pluginList := []string{}
			for _, r := range canBeUpdated {
				pluginList = append(pluginList, r.Plugin)
			}
			//nolint:forbidigo // acceptable
			fmt.Println()
			//nolint:forbidigo // acceptable
			fmt.Printf(
				"To update %s %s: %s\nTo update all plugins: %s",
				utils.Pluralize("this", len(pluginList)),
				utils.Pluralize("plugin", len(pluginList)),
				constants.Bold(fmt.Sprintf("steampipe plugin update %s", strings.Join(pluginList, " "))),
				constants.Bold(fmt.Sprintln("steampipe plugin update --all")),
			)
		}
	}
}
