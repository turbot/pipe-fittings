package parse

import (
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/connection"
)

type ModParseContextOption func(*ModParseContext)

func WithParseFlags(flags ParseModFlag) ModParseContextOption {
	return func(m *ModParseContext) {
		m.Flags = flags
	}
}

func WithListOptions(listOptions filehelpers.ListOptions) ModParseContextOption {
	return func(m *ModParseContext) {
		m.ListOptions = listOptions
	}
}

func WithLateBinding(enabled bool) ModParseContextOption {
	return func(m *ModParseContext) {
		m.supportLateBinding = enabled
	}
}

func WithConnections(connections map[string]connection.PipelingConnection) ModParseContextOption {
	return func(m *ModParseContext) {
		m.PipelingConnections = connections
	}
}
