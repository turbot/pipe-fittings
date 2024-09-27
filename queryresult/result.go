package queryresult

import (
	"time"
)

type RowResult struct {
	Data  []interface{}
	Error error
}
type TimingMetadata struct {
	Duration time.Duration
}

// TimingContainer is an interface that allows us to parameterize the Result struct
// it must handle case of a query returning a stream of timing data OR a timing data struct directly
// the GetTiming func is used to populate the timing data in the JSON output (as we cannot serialise a stream!)
type TimingContainer interface {
	GetTiming() any
}

type Result[T TimingContainer] struct {
	RowChan chan *RowResult
	Cols    []*ColumnDef
	Timing  T
}

func NewResult[T TimingContainer](cols []*ColumnDef, emptyTiming T) *Result[T] {
	c := make(chan *RowResult)
	return &Result[T]{
		RowChan: c,
		Cols:    cols,
		Timing:  emptyTiming,
	}
}

// IsExportSourceData implements ExportSourceData
func (*Result[T]) IsExportSourceData() {}

// Close closes the row channel
func (r *Result[T]) Close() {
	close(r.RowChan)
}

func (r *Result[T]) StreamRow(rowResult []interface{}) {
	r.RowChan <- &RowResult{Data: rowResult}
}
func (r *Result[T]) StreamError(err error) {
	r.RowChan <- &RowResult{Error: err}
}

type SyncQueryResult struct {
	Rows   []interface{}
	Cols   []*ColumnDef
	Timing any
}
