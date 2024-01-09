package printers

import "github.com/turbot/go-kit/helpers"

type GetPrinterOption func(*GetPrinterConfig)

type GetPrinterConfig struct {
	tableCommands map[string]struct{}
}

func (c GetPrinterConfig) commandUsesTable(key string) bool {
	_, usesTable := c.tableCommands[key]
	return usesTable
}

func newGetPrinterConfig() *GetPrinterConfig {
	return &GetPrinterConfig{
		tableCommands: make(map[string]struct{}),
	}
}

func WithTableCommands(tableCommands []string) GetPrinterOption {
	return func(m *GetPrinterConfig) {
		m.tableCommands = helpers.SliceToLookup(tableCommands)
	}
}
