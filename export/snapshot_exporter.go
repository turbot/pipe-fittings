package export

import (
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"strings"

	"github.com/turbot/pipe-fittings/constants"
)

type SnapshotExporter struct {
	ExporterBase
}

func (e *SnapshotExporter) Export(_ context.Context, input ExportSourceData, filePath string) error {
	snapshot, ok := input.(*steampipeconfig.SteampipeSnapshot)
	if !ok {
		return fmt.Errorf("SnapshotExporter input must be a SteampipeSnapshot")
	}
	snapshotBytes, err := snapshot.AsStrippedJson(false)
	if err != nil {
		return err
	}

	res := strings.NewReader(fmt.Sprintf("%s\n", string(snapshotBytes)))

	return Write(filePath, res)
}

func (e *SnapshotExporter) FileExtension() string {
	return constants.SnapshotExtension
}

func (e *SnapshotExporter) Name() string {
	return constants.OutputFormatSnapshot
}

func (*SnapshotExporter) Alias() string {
	return "pps"
}
