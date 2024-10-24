package parse

import (
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/modconfig"
)

type ModParseContextOption[T modconfig.ResourceMapsI] func(*ModParseContext[T])

func WithParseFlags[T modconfig.ResourceMapsI](flags ParseModFlag) ModParseContextOption[T] {
	return func(m *ModParseContext[T]) {
		m.Flags = flags
	}
}

func WithListOptions[T modconfig.ResourceMapsI](listOptions filehelpers.ListOptions) ModParseContextOption[T] {
	return func(m *ModParseContext[T]) {
		m.ListOptions = listOptions
	}
}

func WithLateBinding[T modconfig.ResourceMapsI](enabled bool) ModParseContextOption[T] {
	return func(m *ModParseContext[T]) {
		m.supportLateBinding = enabled
	}
}

func WithConnections[T modconfig.ResourceMapsI](connections map[string]connection.PipelingConnection) ModParseContextOption[T] {
	return func(m *ModParseContext[T]) {
		m.PipelingConnections = connections
	}
}
