package printers

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/viper"
	color2 "github.com/turbot/pipe-helpers/color"
	constants2 "github.com/turbot/pipe-helpers/constants"
	sanitize2 "github.com/turbot/pipe-helpers/sanitize"
)

type StringPrinter[T any] struct {
	colorGenerator *color2.DynamicColorGenerator
	Sanitizer      *sanitize2.Sanitizer
}

func NewStringPrinter[T any]() (*StringPrinter[T], error) {
	colorGenerator, err := color2.NewDynamicColorGenerator(0, 16)
	if err != nil {
		return nil, err
	}

	p := &StringPrinter[T]{
		colorGenerator: colorGenerator,
		Sanitizer:      sanitize2.NullSanitizer,
	}
	return p, nil
}

func (p StringPrinter[T]) PrintResource(_ context.Context, r PrintableResource[T], writer io.Writer) error {
	items := r.GetItems()
	enableColor := viper.GetString(constants2.ArgOutput) == constants2.OutputFormatPretty
	for _, item := range items {
		if item, isSanitizedStringer := any(item).(sanitize2.SanitizedStringer); isSanitizedStringer {
			colorOpts := sanitize2.RenderOptions{
				ColorGenerator: p.colorGenerator,
				ColorEnabled:   enableColor,
				Verbose:        viper.GetBool(constants2.ArgVerbose),
				JsonFormatter:  color2.NewJsonFormatter(!enableColor),
			}

			var str string
			if p.Sanitizer != nil {
				str = item.String(p.Sanitizer, colorOpts)
			} else {
				str = item.String(sanitize2.Instance, colorOpts)
			}

			if _, err := writer.Write([]byte(str)); err != nil {
				return fmt.Errorf("error printing resource")
			}
		}

		// TODO just call String() and manually sanitize
	}
	return nil
}
