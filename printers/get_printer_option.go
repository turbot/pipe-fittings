package printers

import "github.com/turbot/go-kit/helpers"

type GetPrinterOption func(*GetPrinterConfig)

type GetPrinterConfig struct {
	stringPrinterCommands map[string]struct{}
}

func (c GetPrinterConfig) commandUsesStringPrinter(key string) bool {
	_, usesTable := c.stringPrinterCommands[key]
	return usesTable
}

func newGetPrinterConfig() *GetPrinterConfig {
	return &GetPrinterConfig{
		stringPrinterCommands: make(map[string]struct{}),
	}
}

func WithStringPrinterCommands(stringPrinterCommands []string) GetPrinterOption {
	return func(m *GetPrinterConfig) {
		m.stringPrinterCommands = helpers.SliceToLookup(stringPrinterCommands)
	}
}
