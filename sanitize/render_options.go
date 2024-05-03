package sanitize

import (
	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/pipe-fittings/v2/color"
)

type RenderOptions struct {
	ColorEnabled   bool
	ColorGenerator *color.DynamicColorGenerator
	Verbose        bool
	JsonFormatter  *prettyjson.Formatter
	Indent         int
	// todo not the correct place for this??
	IsList bool
}

func (o RenderOptions) Clone() RenderOptions {

	return RenderOptions{
		ColorEnabled:   o.ColorEnabled,
		ColorGenerator: o.ColorGenerator,
		Verbose:        o.Verbose,
		JsonFormatter:  o.JsonFormatter,
		Indent:         o.Indent,
		IsList:         o.IsList,
	}
}
