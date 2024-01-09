package printers

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"io"
)

// Inspired by Kubernetes

// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, PrintableResource[T], io.Writer) error
}

func GetPrinter[T any](cmd *cobra.Command, opts ...GetPrinterOption) (ResourcePrinter[T], error) {
	cfg := newGetPrinterConfig()
	for _, o := range opts {
		o(cfg)
	}

	format := viper.GetString(constants.ArgOutput)
	key := cmdconfig.CommandFullKey(cmd)

	switch format {
	case constants.OutputFormatPretty, constants.OutputFormatPlain:
		if cfg.commandUsesTable(key) {
			return NewTablePrinter[T]()
		}
		return NewStringPrinter[T]()
	case constants.OutputFormatJSON:
		return NewJsonPrinter[T]()
	case constants.OutputFormatYAML:
		return NewYamlPrinter[T]()
	}
	return nil, fmt.Errorf("unknown output format %q", format)
}
