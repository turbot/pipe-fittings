package printers

import (
	"context"
	"encoding/json"
	"io"

	"github.com/turbot/pipe-fittings/color"
	"github.com/turbot/pipe-fittings/sanitize"
)

type JsonPrinter[T any] struct {
	Sanitizer *sanitize.Sanitizer
}

func NewJsonPrinter[T any]() (*JsonPrinter[T], error) {
	return &JsonPrinter[T]{
		Sanitizer: sanitize.Instance,
	}, nil
}

func (p JsonPrinter[T]) PrintResource(ctx context.Context, r PrintableResource[T], writer io.Writer) error {
	// marshal
	s, err := json.Marshal(r.GetItems())
	if err != nil {
		return err
	}

	// sanitize
	s = []byte(p.Sanitizer.SanitizeString(string(s)))

	// format
	s, err = color.NewJsonFormatter(true).Format(s)
	if err != nil {
		return err
	}
	_, err = writer.Write(s)
	if err != nil {
		return err
	}

	return nil
}
