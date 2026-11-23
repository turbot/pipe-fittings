package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/pkg/twwidth"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/utils"
)

type timeLog struct {
	Time       time.Time
	Interval   time.Duration
	Cumulative time.Duration
	Operation  string
}

var timing []timeLog

func shouldProfile() bool {
	return strings.ToUpper(os.Getenv(app_specific.EnvProfile)) == "TRUE"
}

func LogTime(operation string) {
	if !shouldProfile() {
		return
	}
	lastTimelogIdx := len(timing) - 1
	var elapsed time.Duration
	var cumulative time.Duration
	if lastTimelogIdx >= 0 {
		cumulative = time.Since(timing[0].Time)
		elapsed = time.Since(timing[lastTimelogIdx].Time)
	}
	timing = append(timing, timeLog{time.Now(), elapsed, cumulative, operation})
}

func ClearProfileData() {
	timing = []timeLog{}
}

func DisplayProfileData(minTime time.Duration) {
	if shouldProfile() {

		minString := fmt.Sprintf("< %s", minTime.String())
		var data [][]string
		for _, logEntry := range timing {
			var itemData []string
			itemData = append(itemData, logEntry.Operation)

			timeStr := logEntry.Time.Format(time.StampMilli)
			itemData = append(itemData, timeStr)

			intervalStr := logEntry.Interval.String()
			if logEntry.Interval < minTime {
				intervalStr = minString
			}
			itemData = append(itemData, intervalStr)

			cumulativeStr := logEntry.Cumulative.String()
			if logEntry.Cumulative < minTime {
				cumulativeStr = minString
			}
			itemData = append(itemData, cumulativeStr)

			data = append(data, itemData)
		}

		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		displayTable(writer, data)
		_ = writer.Flush()

		// weird hoops we have to jump through to get this printing vial hclog
		s := ""
		for _, b := range b.Bytes() {
			if b == '\n' {
				log.Printf("[WARN] %s\n", s)
				s = ""
			} else {
				s += string(b)
			}
		}

		ClearProfileData()
	}

}

func displayTable(out io.Writer, data [][]string) {
	// The Operation column wraps at 50 characters, as with tablewriter v0 (the v1 limit includes the 2 padding spaces).
	// v1 collapses runs of spaces in every cell of a column that has a limit, so the limit is set only when an
	// Operation name needs it, and never on the Time column (StampMilli puts two spaces before days 1-9).
	// When it is set, runs of spaces in that column collapse.
	table := utils.NewASCIITable(out, tablewriter.WithRowAutoWrap(tw.WrapNormal))
	for _, row := range data {
		if len(row) > 0 && twwidth.Width(row[0]) > 50 {
			table.Configure(func(cfg *tablewriter.Config) {
				cfg.Row.ColMaxWidths.PerColumn = tw.Mapper[int, int]{0: 52}
			})
			break
		}
	}
	table.Header([]string{"Operation", "Time", "Elapsed", "Cumulative"})
	_ = table.Bulk(data)
	_ = table.Render()
}
