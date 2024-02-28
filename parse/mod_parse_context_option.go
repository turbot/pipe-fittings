package parse

import filehelpers "github.com/turbot/go-kit/files"

// use options pattern
type ModParseContextOption func(*ModParseContext)

func WithParseFlags(flags ParseModFlag) ModParseContextOption {
	return func(m *ModParseContext) {
		m.Flags = flags
	}
}
func WithListOptions(listOptions *filehelpers.ListOptions) ModParseContextOption {
	return func(m *ModParseContext) {
		m.ListOptions = listOptions
	}
}
