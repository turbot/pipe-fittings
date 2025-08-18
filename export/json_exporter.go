package export

import (
	"context"
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/pipe-fittings/v2/queryresult"
	constants2 "github.com/turbot/pipe-helpers/constants"
)

type JsonExporter struct {
	ExporterBase
}

// Export processes the query result and writes JSON output to a file
func (e *JsonExporter) Export(ctx context.Context, input ExportSourceData, filePath string) error {
	result, ok := input.(*queryresult.Result[*queryresult.QueryTimingMetadata])
	if !ok {
		return fmt.Errorf("JsonExporter input must be a queryresult.Result")
	}

	// Generate JSON output
	jsonString, _, _ := querydisplay.BuildJSON(ctx, result)

	// Write to file
	return Write(filePath, strings.NewReader(jsonString))
}

func (e *JsonExporter) FileExtension() string {
	return constants2.JsonExtension
}

func (e *JsonExporter) Name() string {
	return constants2.OutputFormatJSON
}
