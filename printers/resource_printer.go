package printers

import (
	"context"
	"fmt"
	"io"
	"strings"

	constants2 "github.com/turbot/pipe-helpers/constants"
	"github.com/turbot/pipe-helpers/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Inspired by Kubernetes

// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, PrintableResource[T], io.Writer) error
}

func GetPrinter[T any](cmd *cobra.Command) (ResourcePrinter[T], error) {
	f := viper.GetString(constants2.ArgOutput)
	key := utils.CommandFullKey(cmd)
	cmdType := strings.Split(key, ".")[len(strings.Split(key, "."))-1]
	switch f {
	case constants2.OutputFormatPretty, constants2.OutputFormatPlain:
		switch cmdType {
		case "list":
			return NewTablePrinter[T]()
		case "show":
			var empty T
			if IsShowable(empty) {
				return NewShowPrinter[T]()
			}
			return NewStringPrinter[T]()

		default:
			return NewStringPrinter[T]()
		}
	case constants2.OutputFormatJSON:
		return NewJsonPrinter[T]()
	case constants2.OutputFormatYAML:
		return NewYamlPrinter[T]()
	}
	return nil, fmt.Errorf("unknown output format %q", f)
}
