package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/pipe-fittings/v2/app_specific"
)

type timeLog struct {
	Time       time.Time
	Interval   time.Duration
	Cumulative time.Duration
	Operation  string
}

var Timing []timeLog

func shouldProfile() bool {
	return strings.ToUpper(os.Getenv(app_specific.EnvProfile)) == "TRUE"
}

func LogTime(operation string) {
	if !shouldProfile() {
		return
	}
	lastTimelogIdx := len(Timing) - 1
	var elapsed time.Duration
	var cumulative time.Duration
	if lastTimelogIdx >= 0 {
		elapsed = time.Since(Timing[lastTimelogIdx].Time)
		cumulative = time.Since(Timing[0].Time)

	}
	Timing = append(Timing, timeLog{time.Now(), elapsed, cumulative, operation})
}

func DisplayProfileData(op io.Writer) {
	if shouldProfile() {
		fmt.Fprint(op, "Timing\n") //nolint:forbidigo // TODO: better way to print out? Or maybe this is acceptable

		var data [][]string
		for _, logEntry := range Timing {
			var itemData []string
			itemData = append(itemData, logEntry.Operation)
			if logEntry.Interval > 300*time.Millisecond {
				itemData = append(itemData, aurora.Bold(aurora.BrightRed(logEntry.Interval.String())).String())
			} else {
				itemData = append(itemData, logEntry.Interval.String())
			}
			itemData = append(itemData, logEntry.Cumulative.String())
			data = append(data, itemData)
		}
		table := tablewriter.NewWriter(op)
		table.SetHeader([]string{"Operation", "Elapsed", "Cumulative"})
		table.SetBorder(true)
		table.SetReflowDuringAutoWrap(false)
		table.SetAutoWrapText(false)
		table.AppendBulk(data)
		table.Render()
	}
}

func DisplayProfileDataJsonl(op io.Writer) {
	if shouldProfile() {
		// Create a new JSON encoder
		encoder := json.NewEncoder(op)

		for _, logEntry := range Timing {
			var itemData []string
			itemData = append(itemData, logEntry.Operation)
			itemData = append(itemData, logEntry.Interval.String())
			itemData = append(itemData, logEntry.Cumulative.String())

			if err := encoder.Encode(itemData); err != nil {
				panic(err)
			}
		}
	}
}
