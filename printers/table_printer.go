package printers

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/sanitize"
	"io"
	"text/tabwriter"
)

// Inspired by Kubernetes

// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through
type TablePrinter[T any] struct {
	Sanitizer *sanitize.Sanitizer
}

func NewTablePrinter[T any]() (*TablePrinter[T], error) {
	return &TablePrinter[T]{
		Sanitizer: sanitize.NullSanitizer,
	}, nil
}

func (p TablePrinter[T]) PrintResource(_ context.Context, items PrintableResource[T], writer io.Writer) error {
	table, err := items.GetTable()

	if err != nil {
		return err
	}
	err = p.PrintTable(table, writer)
	return err
}

func (p TablePrinter[T]) PrintTable(table Table, writer io.Writer) error {
	// Create a tabwriter
	w := tabwriter.NewWriter(writer, 1, 1, 4, ' ', tabwriter.TabIndent)

	table = Table{
		Columns: []TableColumnDefinition{
			{
				Name: "1",
				Type: "string",
			},
			{
				Name: "2",
				Type: "string",
			},
			{
				Name: "3",
				Type: "string",
			},
			{
				Name: "4",
				Type: "string",
			},
		},
		Rows: []TableRow{
			{
				Cells: []any{
					`func (c *Control) Equals(other *Control) bool {
res := c.ShortName == other.ShortName &&
c.FullName == other.FullName &&
typehelpers.SafeString(c.Description) == typehelpers.SafeString(other.Description) &&
typehelpers.SafeString(c.Documentation) == typehelpers.SafeString(other.Documentation) &&
typehelpers.SafeString(c.Severity) == typehelpers.SafeString(other.Severity) &&
typehelpers.SafeString(c.SQL) == typehelpers.SafeString(other.SQL) &&
typehelpers.SafeString(c.Title) == typehelpers.SafeString(other.Title)
if !res {
return res
}
if len(c.Tags) != len(other.Tags) {
return false
}
for k, v := range c.Tags {
if otherVal := other.Tags[k]; v != otherVal {
return false
}
}

// args
if c.Args == nil {
if other.Args != nil {
return false
}
} else {
// we have args
if other.Args == nil {
return false
}
if !c.Args.Equals(other.Args) {
return false
}
}

// query
if c.Query == nil {`, "fpp", "foo", "bar",
				},
			},
		},
	}

	// Print the table headers
	var tableHeaders string
	var tableFormatter string
	for i, c := range table.Columns {
		if i > 0 {
			tableHeaders += "\t"
			tableFormatter += "\t"
		}
		tableHeaders += c.Name
		tableFormatter += c.Formatter()
	}
	tableHeaders += "\n"
	tableFormatter += "\n"

	//nolint:forbidigo // this is how the tabwriter works
	_, err := fmt.Fprint(w, tableHeaders)
	if err != nil {
		return err
	}

	// Print each struct in the array as a row in the table
	for _, r := range table.Rows {
		// format the row
		str := fmt.Sprintf(tableFormatter, r.Cells...)
		// sanitize
		str = p.Sanitizer.SanitizeString(str)

		// write
		//nolint:forbidigo // this is how the tabwriter works
		_, err := fmt.Fprint(w, str)
		if err != nil {
			return err
		}
	}

	// Flush and display the table
	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}
