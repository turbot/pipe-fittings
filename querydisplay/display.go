package querydisplay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	pconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/queryresult"
	pqueryresult "github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/utils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ShowOutput displays the output using the proper formatter as applicable
func ShowOutput[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) (rowCount, rowErrors int) {

	outputFormat := viper.GetString(pconstants.ArgOutput)
	switch outputFormat {
	case constants.OutputFormatJSON:
		rowCount, rowErrors = displayJSON(ctx, result)
	case constants.OutputFormatCSV:
		rowCount, rowErrors = displayCSV(ctx, result)
	case constants.OutputFormatLine:
		rowCount, rowErrors = displayLine(ctx, result)
	case constants.OutputFormatTable:
		displayTable(ctx, result)
	}

	return rowCount, rowErrors
}

type ShowWrappedTableOptions struct {
	AutoMerge        bool
	HideEmptyColumns bool
	Truncate         bool
	OutputMirror     io.Writer
}

func ShowWrappedTable(headers []string, rows [][]string, opts *ShowWrappedTableOptions) {
	if opts == nil {
		opts = &ShowWrappedTableOptions{}
	}
	t := table.NewWriter()

	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault
	if opts.OutputMirror == nil {
		t.SetOutputMirror(os.Stdout)
	} else {
		t.SetOutputMirror(opts.OutputMirror)
	}

	rowConfig := table.RowConfig{AutoMerge: opts.AutoMerge}
	colConfigs, headerRow := getColumnSettings(headers, rows, opts)

	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(headerRow)

	for _, row := range rows {
		rowObj := table.Row{}
		for _, col := range row {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj, rowConfig)
	}
	t.Render()
}

// show table using text/tabwriter
func ShowTable(headers []string, rows [][]string, opts *ShowWrappedTableOptions) {
	if opts == nil {
		opts = &ShowWrappedTableOptions{}
	}

	// Use the provided OutputMirror or default to os.Stdout
	writer := opts.OutputMirror
	if writer == nil {
		writer = os.Stdout
	}

	// Create a tabwriter
	tw := tabwriter.NewWriter(writer, 1, 1, 4, ' ', 0)

	// Prepare the headers
	headerLine := strings.Join(headers, "\t")
	fmt.Fprintln(tw, headerLine) //nolint:forbidigo // ui output

	// Process rows
	for _, row := range rows {
		// Join the row into a tab-delimited string
		rowLine := strings.Join(row, "\t")
		fmt.Fprintln(tw, rowLine) //nolint:forbidigo // ui output
	}

	// Flush the writer to render the table
	if err := tw.Flush(); err != nil {
		fmt.Printf("Error flushing tabwriter: %v\n", err) //nolint:forbidigo // ui output
	}
}

func GetMaxCols() int {
	colsAvailable, _, _ := gows.GetWinSize()
	// check if STEAMPIPE_DISPLAY_WIDTH env variable is set
	if viper.IsSet(pconstants.ArgDisplayWidth) {
		colsAvailable = viper.GetInt(pconstants.ArgDisplayWidth)
	}
	return colsAvailable
}

// calculate and returns column configuration based on header and row content
func getColumnSettings(headers []string, rows [][]string, opts *ShowWrappedTableOptions) ([]table.ColumnConfig, table.Row) {
	colConfigs := make([]table.ColumnConfig, len(headers))
	headerRow := make(table.Row, len(headers))

	sumOfAllCols := 0

	// account for the spaces around the value of a column and separators
	spaceAccounting := (len(headers) * 3) + 1

	for idx, colName := range headers {
		headerRow[idx] = colName

		// get the maximum len of strings in this column
		maxLen := getTerminalColumnsRequiredForString(colName)
		colHasValue := false
		for _, row := range rows {
			colVal := row[idx]
			if !colHasValue && len(colVal) > 0 {
				// the !colHasValue is necessary in the condition,
				// otherwise, even after being set, we will keep
				// evaluating the length
				colHasValue = true
			}

			// get the maximum line length of the value
			colLen := getTerminalColumnsRequiredForString(colVal)
			if colLen > maxLen {
				maxLen = colLen
			}
		}
		colConfigs[idx] = table.ColumnConfig{
			Name:     colName,
			Number:   idx + 1,
			WidthMax: maxLen,
			WidthMin: maxLen,
		}
		if opts.HideEmptyColumns && !colHasValue {
			colConfigs[idx].Hidden = true
		}
		sumOfAllCols += maxLen
	}

	// now that all columns are set to the widths that they need,
	// set the last one to occupy as much as is available - no more - no less
	sumOfRest := sumOfAllCols - colConfigs[len(colConfigs)-1].WidthMax
	// get the max cols width
	maxCols := GetMaxCols()
	if sumOfAllCols > maxCols {
		colConfigs[len(colConfigs)-1].WidthMax = maxCols - sumOfRest - spaceAccounting
		colConfigs[len(colConfigs)-1].WidthMin = maxCols - sumOfRest - spaceAccounting
		if opts.Truncate {
			colConfigs[len(colConfigs)-1].WidthMaxEnforcer = helpers.TruncateString
		}
	}

	return colConfigs, headerRow
}

// getTerminalColumnsRequiredForString returns the length of the longest line in the string
func getTerminalColumnsRequiredForString(str string) int {
	colsRequired := 0
	scanner := bufio.NewScanner(bytes.NewBufferString(str))
	for scanner.Scan() {
		line := scanner.Text()
		runeCount := utf8.RuneCountInString(line)
		if runeCount > colsRequired {
			colsRequired = runeCount
		}
	}
	return colsRequired
}

type jsonOutput struct {
	Columns  []pqueryresult.ColumnDef `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
	Metadata any                      `json:"metadata,omitempty"`
}

func newJSONOutput() *jsonOutput {
	return &jsonOutput{
		Rows: make([]map[string]interface{}, 0),
	}
}

func displayJSON[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) (rowCount, rowErrors int) {
	jsonOutput := newJSONOutput()

	// add column defs to the JSON output
	for _, col := range result.Cols {
		// create a new column def, converting the data type to lowercase
		c := pqueryresult.ColumnDef{
			Name:         col.Name,
			OriginalName: col.OriginalName,
			DataType:     strings.ToLower(col.DataType),
		}
		// add to the column def array
		jsonOutput.Columns = append(jsonOutput.Columns, c)
	}

	// define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		record := map[string]interface{}{}
		for idx, col := range result.Cols {
			value, _ := ParseJSONOutputColumnValue(row[idx], col)
			// get the column def
			c := jsonOutput.Columns[idx]
			// add the value under the unique column name
			record[c.Name] = value
		}
		jsonOutput.Rows = append(jsonOutput.Rows, record)
	}

	// call this function for each row
	count, err := IterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return 0, rowErrors
	}

	// now we have iterated the rows, get the timing
	if viper.IsSet(constants.ArgTiming) {
		jsonOutput.Metadata = result.Timing.GetTiming()
	}

	// display the JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(jsonOutput); err != nil {
		//nolint:forbidigo // acceptable
		fmt.Print("Error displaying result as JSON", err)
		return 0, 0
	}
	return count, rowErrors
}

func displayCSV[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) (rowCount, rowErrors int) {

	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(viper.GetString(pconstants.ArgSeparator))[0]

	if viper.GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(columnNames(result.Cols))
	}

	// print the data as it comes
	// define function display each csv row
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		rowAsString, _ := ColumnValuesAsString(row, result.Cols, WithNullString(""))
		_ = csvWriter.Write(rowAsString)
	}

	// call this function for each row
	count, err := IterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return 0, rowErrors
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		error_helpers.ShowErrorWithMessage(ctx, csvWriter.Error(), "unable to print csv")
	}

	return count, rowErrors
}

func displayLine[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) (rowCount, rowErrors int) {

	maxColNameLength, rowErrors := 0, 0
	for _, col := range result.Cols {
		thisLength := utf8.RuneCountInString(col.Name)
		if thisLength > maxColNameLength {
			maxColNameLength = thisLength
		}
	}
	itemIdx := 0

	// define a function to display each row
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		recordAsString, _ := ColumnValuesAsString(row, result.Cols)
		requiredTerminalColumnsForValuesOfRecord := 0
		for _, colValue := range recordAsString {
			colRequired := getTerminalColumnsRequiredForString(colValue)
			if requiredTerminalColumnsForValuesOfRecord < colRequired {
				requiredTerminalColumnsForValuesOfRecord = colRequired
			}
		}

		lineFormat := fmt.Sprintf("%%-%ds | %%s\n", maxColNameLength)
		multiLineFormat := fmt.Sprintf("%%-%ds | %%-%ds", maxColNameLength, requiredTerminalColumnsForValuesOfRecord)

		fmt.Printf("-[ RECORD %-2d ]%s\n", itemIdx+1, strings.Repeat("-", 75)) //nolint:forbidigo // intentional use of fmt

		// get the column names (this takes into account the original name)
		columnNames := columnNames(result.Cols)
		for idx, column := range recordAsString {
			lines := strings.Split(column, "\n")
			if len(lines) == 1 {
				//nolint:forbidigo // acceptable
				fmt.Printf(lineFormat, columnNames[idx], lines[0])
			} else {
				for lineIdx, line := range lines {
					if lineIdx == 0 {
						// the first line
						//nolint:forbidigo // acceptable
						fmt.Printf(multiLineFormat, columnNames[idx], line)
					} else {
						// next lines
						//nolint:forbidigo // acceptable
						fmt.Printf(multiLineFormat, "", line)
					}

					// is this not the last line of value?
					if lineIdx < len(lines)-1 {
						//nolint:forbidigo // acceptable
						fmt.Printf(" +\n")
					} else {
						//nolint:forbidigo // acceptable
						fmt.Printf("\n")
					}

				}
			}
		}
		itemIdx++

	}

	// call this function for each row
	var err error
	rowCount, err = IterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return 0, rowErrors
	}

	return rowCount, rowErrors
}

// tables are buffered before rendering so limit the display to the first 10000 rows
const maxTableDisplayRows = 10000

func displayTable[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) (rowCount, rowErrors int) {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	// the table
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault

	var colConfigs []table.ColumnConfig
	headers := make(table.Row, len(result.Cols))

	// get the column names (this takes into account the original name)
	columnNames := columnNames(result.Cols)
	for idx, columnName := range columnNames {
		headers[idx] = columnName
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     columnName,
			Number:   idx + 1,
			WidthMax: constants.MaxColumnWidth,
		})
	}

	t.SetColumnConfigs(colConfigs)
	if viper.GetBool(pconstants.ArgHeader) {
		t.AppendHeader(headers)
	}

	// how may rows are we displaying (max 10000)
	displayRowCount := 0

	// define a function to execute for each row
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		if displayRowCount >= maxTableDisplayRows {
			return
		}
		displayRowCount++

		rowAsString, _ := ColumnValuesAsString(row, result.Cols)
		rowObj := table.Row{}
		for _, col := range rowAsString {
			// trim out non-displayable code-points in string
			// exfept white-spaces
			col = strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) || unicode.IsGraphic(r) {
					// return if this is a white space character
					return r
				}
				return -1
			}, col)
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
	}

	// iterate each row, adding each to the table
	count, err := IterateResults(result, rowFunc)
	if err != nil {
		// display the error
		//nolint:forbidigo // acceptable
		fmt.Println()
		error_helpers.ShowError(ctx, err)
		rowErrors++
		//nolint:forbidigo // acceptable
		fmt.Println()
	}
	// write out the table to the buffer
	t.Render()

	// page out the table
	ShowPaged(ctx, outbuf.String())

	status := fmt.Sprintf("%s rows", utils.HumanizeNumber(count))
	if displayRowCount >= maxTableDisplayRows {
		status += fmt.Sprintf(" (%s shown)", utils.HumanizeNumber(maxTableDisplayRows))
	}
	//nolint:forbidigo // acceptable
	fmt.Println(status)

	return count, rowErrors
}

type displayResultsFunc[T queryresult.TimingContainer] func(row []interface{}, result *queryresult.Result[T])

// call func displayResult for each row of results
func IterateResults[T queryresult.TimingContainer](result *queryresult.Result[T], displayResult displayResultsFunc[T]) (int, error) {
	count := 0
	for row := range result.RowChan {
		if row == nil {
			return count, nil
		}
		if row.Error != nil {
			return count, row.Error
		}
		displayResult(row.Data, result)
		count++
	}
	// we will not get here
	return count, nil
}

// DisplayErrorTiming shows the time taken for the query to fail
func DisplayErrorTiming(t time.Time) {
	elapsed := time.Since(t)
	var sb strings.Builder
	// large numbers should be formatted with commas
	p := message.NewPrinter(language.English)

	milliseconds := float64(elapsed.Microseconds()) / 1000
	seconds := elapsed.Seconds()
	if seconds < 0.5 {
		sb.WriteString(p.Sprintf("\nTime: %dms.", int64(milliseconds)))
	} else {
		sb.WriteString(p.Sprintf("\nTime: %.1fs.", seconds))
	}
	//nolint:forbidigo // acceptable
	fmt.Println(sb.String())
}
