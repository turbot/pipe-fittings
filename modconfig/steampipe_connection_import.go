package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	gokit "github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/utils"
)

// The definition of a single ConnectionImport
type ConnectionImport struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	FileName        string `json:"file_name"`
	StartLineNumber int    `json:"start_line_number"`
	EndLineNumber   int    `json:"end_line_number"`

	Source      *string  `json:"source" cty:"source" hcl:"source"`
	Connections []string `json:"connections" cty:"connections" hcl:"connections,optional"`
	Prefix      *string  `json:"prefix" cty:"prefix" hcl:"prefix,optional"`
}

func (c ConnectionImport) Equals(other ConnectionImport) bool {

	return utils.PtrEqual(c.Source, other.Source) &&
		gokit.StringSliceEqualIgnoreOrder(c.Connections, other.Connections) &&
		utils.PtrEqual(c.Prefix, other.Prefix)
}

func (c *ConnectionImport) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	c.FileName = fileName
	c.StartLineNumber = startLineNumber
	c.EndLineNumber = endLineNumber
}

func (c *ConnectionImport) GetSource() *string {
	return c.Source
}

func (c *ConnectionImport) GetPrefix() *string {
	return c.Prefix
}

func (c *ConnectionImport) GetConnections() []string {
	return c.Connections
}

func NewConnectionImport(block *hcl.Block) *ConnectionImport {

	connectionImportName := block.Labels[0]

	return &ConnectionImport{
		HclResourceImpl: HclResourceImpl{
			FullName:        connectionImportName,
			ShortName:       connectionImportName,
			UnqualifiedName: connectionImportName,
			DeclRange:       block.DefRange,
		},
		FileName:        block.DefRange.Filename,
		StartLineNumber: block.DefRange.Start.Line,
		EndLineNumber:   block.DefRange.End.Line,
	}
}
