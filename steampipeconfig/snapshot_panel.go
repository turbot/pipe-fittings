package steampipeconfig

// SnapshotPanel is an interface implemented by all nodes which are to be included in the Snapshot Panels map
// this consists of all 'Run' types - LeafRun, DashboardRun, etc.
type SnapshotPanel interface {
	IsSnapshotPanel()
}

// SnapshotRowsPanel is a snapshot panel whose data rows can be written one at a time.
// SteampipeSnapshot.WriteStrippedJson then converts one row at a time, not the whole panel.
//
// The panel must marshal to the same JSON as PanelWithoutRows, with a "rows" array added to its "data" object.
type SnapshotRowsPanel interface {
	SnapshotPanel
	// PanelWithoutRows returns a value that marshals to the panel JSON without "rows" in its "data" object
	PanelWithoutRows() any
	// EachRow calls f with each data row, in order, and stops at the first error
	EachRow(f func(row any) error) error
}
