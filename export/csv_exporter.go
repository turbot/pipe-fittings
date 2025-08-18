package export

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/pipe-fittings/v2/queryresult"
	constants2 "github.com/turbot/pipe-helpers/constants"
)

type CsvExporter struct {
	ExporterBase
}

// Export processes the query result and writes CSV output to a file
func (e *CsvExporter) Export(ctx context.Context, input ExportSourceData, filePath string) error {
	result, ok := input.(*queryresult.Result[*queryresult.QueryTimingMetadata])
	if !ok {
		return fmt.Errorf("CsvExporter input must be a queryresult.Result")
	}

	// Generate csv output
	jsonString, _, _ := querydisplay.BuildCSV(ctx, result)

	// Write to file
	return Write(filePath, strings.NewReader(jsonString))
}

func (e *CsvExporter) FileExtension() string {
	return constants2.CsvExtension
}

func (e *CsvExporter) Name() string {
	return constants2.OutputFormatCSV
}
