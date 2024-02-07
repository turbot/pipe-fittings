package printers

import (
	"github.com/turbot/pipe-fittings/sanitize"
)

type FieldRenderOptions struct {
	// a function implementing custom rendering logic to display the value
	RenderValueFunc func(opts sanitize.RenderOptions) string
	// a function implementing custom rendering logic to display the key AND value
	RenderKeyValueFunc func(opts sanitize.RenderOptions) string
	Indent             int
}

func newFieldRenderOptions(f FieldValue) FieldRenderOptions {
	return FieldRenderOptions{
		RenderValueFunc:    f.RenderValueFunc,
		RenderKeyValueFunc: f.RenderKeyValueFunc,
		Indent:             f.Indent,
	}
}
