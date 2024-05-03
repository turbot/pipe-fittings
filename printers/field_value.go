package printers

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/sanitize"
)

type FieldValueOption func(*FieldValue)

// WithListKeyRender sets this field as the key when the item is rendered as a list
// and provides a function to render the key
func WithListKeyRender(render RenderFunc) FieldValueOption {
	return func(m *FieldValue) {
		m.renderOpts.listOpts.listKeyRenderFunc = render
		m.renderOpts.listOpts.isKey = true
	}
}

// WithListKey sets this field as the key when the item is rendered as a list
func WithListKey() FieldValueOption {
	return func(m *FieldValue) {
		m.renderOpts.listOpts.isKey = true
	}
}
func WithRenderValueFunc(render RenderFunc) FieldValueOption {
	return func(m *FieldValue) {
		m.renderOpts.renderValueFunc = render
	}
}

type RenderFunc func(opts sanitize.RenderOptions) string

type FieldValue struct {
	Name       string
	Value      any
	renderOpts FieldRenderOptions
}

func NewFieldValue(name string, value any, opts ...FieldValueOption) FieldValue {
	v := FieldValue{
		Name:  name,
		Value: value,
	}
	for _, opt := range opts {
		opt(&v)
	}
	return v
}

func (f FieldValue) ValueString() string {
	return typehelpers.ToString(f.Value)
}
