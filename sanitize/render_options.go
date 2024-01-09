package sanitize

import (
	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/pipe-fittings/color"
)

type RenderOptions struct {
	ColorEnabled   bool
	ColorGenerator *color.DynamicColorGenerator
	Verbose        bool
	JsonFormatter  *prettyjson.Formatter
}
