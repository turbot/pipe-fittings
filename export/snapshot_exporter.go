package export

import (
	"context"
	"fmt"
	"io"

	"github.com/turbot/pipe-fittings/v2/steampipeconfig"

	"github.com/turbot/pipe-fittings/v2/constants"
)

type SnapshotExporter struct {
	ExporterBase
}

func (e *SnapshotExporter) Export(_ context.Context, input ExportSourceData, filePath string) error {
	snapshot, ok := input.(*steampipeconfig.SteampipeSnapshot)
	if !ok {
		return fmt.Errorf("SnapshotExporter input must be a SteampipeSnapshot")
	}
	return Write(filePath, NewSnapshotReader(snapshot, false))
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

// NewSnapshotReader returns a reader for the stripped snapshot JSON followed by a newline: the content of a snapshot file.
// io.Copy streams it to the destination one panel at a time.
func NewSnapshotReader(snapshot *steampipeconfig.SteampipeSnapshot, indent bool) *StreamReader {
	return NewStreamReader(func(w io.Writer) error {
		if err := snapshot.WriteStrippedJson(w, indent); err != nil {
			return err
		}
		_, err := io.WriteString(w, "\n")
		return err
	})
}
