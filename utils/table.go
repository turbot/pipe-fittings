package utils

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// NewASCIITable returns a table that draws its borders with +, - and | characters
func NewASCIITable(w io.Writer, opts ...tablewriter.Option) *tablewriter.Table {
	ascii := tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)}))
	return tablewriter.NewTable(w, append([]tablewriter.Option{ascii}, opts...)...)
}
