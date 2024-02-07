package printers

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
)

// Inspired by Kubernetes

// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, PrintableResource[T], io.Writer) error
}

func GetPrinter[T any](cmd *cobra.Command) (ResourcePrinter[T], error) {
	f := viper.GetString(constants.ArgOutput)
	key := commandFullKey(cmd)
	cmdType := strings.Split(key, ".")[len(strings.Split(key, "."))-1]
	switch f {
	case constants.OutputFormatPretty, constants.OutputFormatPlain:
		switch cmdType {
		case "list":
			return NewTablePrinter[T]()
		default:
			return NewStringPrinter[T]()
		}
	case constants.OutputFormatJSON:
		return NewJsonPrinter[T]()
	case constants.OutputFormatYAML:
		return NewYamlPrinter[T]()
	}
	return nil, fmt.Errorf("unknown output format %q", f)
}

func commandFullKey(cmd *cobra.Command) string {
	var parents []string
	parents = append(parents, cmd.Name())
	cmd.VisitParents(func(parent *cobra.Command) {
		parents = append(parents, parent.Name())
	})

	slices.Reverse(parents)

	return strings.Join(parents, ".")
}
