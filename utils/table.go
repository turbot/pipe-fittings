package utils

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// NewASCIITable returns a table that draws its borders with +, - and | characters.
// Like tablewriter v0, it keeps the spaces at the start and end of a cell (v1 trims them by default).
// Unlike v0, a leading tab is still dropped
func NewASCIITable(w io.Writer, opts ...tablewriter.Option) *tablewriter.Table {
	ascii := tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)}))
	return tablewriter.NewTable(w, append([]tablewriter.Option{ascii, tablewriter.WithTrimSpace(tw.Off)}, opts...)...)
}
