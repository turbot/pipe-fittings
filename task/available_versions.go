package task

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"os"
	//"github.com/turbot/pipe-fittings/plugin"
	"github.com/turbot/pipe-fittings/utils"
)

type AvailableVersionCache struct {
	StructVersion uint32                   `json:"struct_version"`
	CliCache      *CLIVersionCheckResponse `json:"cli_version"`
}

func (av *AvailableVersionCache) asTable() (*tablewriter.Table, error) {
	notificationLines, err := av.buildNotification()
	if err != nil {
		return nil, err
	}
	notificationTable := utils.Map(notificationLines, func(line string) []string {
		return []string{line}
	})

	if len(notificationLines) == 0 {
		return nil, nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to provide the dialog feel
	table.AppendBulk(notificationTable)        // Add Bulk Data

	return table, nil
}

func (av *AvailableVersionCache) buildNotification() ([]string, error) {
	cliLines, err := av.cliNotificationMessage()
	if err != nil {
		return nil, err
	}
	return cliLines, nil
}

func (av *AvailableVersionCache) cliNotificationMessage() ([]string, error) {
	info := av.CliCache
	if info == nil {
		return nil, nil
	}

	if info.NewVersion == "" {
		return nil, nil
	}

	newVersion, err := semver.NewVersion(info.NewVersion)
	if err != nil {
		return nil, err
	}

	if newVersion.GreaterThan(app_specific.AppVersion) {
		var downloadURLColor = color.New(color.FgYellow)
		var notificationLines = []string{
			"",
			fmt.Sprintf("A new version of Steampipe is available! %s → %s", constants.Bold(app_specific.AppVersion.String()), constants.Bold(newVersion)),
			fmt.Sprintf("You can update by downloading from %s", downloadURLColor.Sprint("https://steampipe.io/downloads")),
			"",
		}
		return notificationLines, nil
	}
	return nil, nil
}
