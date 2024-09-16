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

type Result[T any] struct {
	RowChan chan *RowResult
	Cols    []*ColumnDef
	Timing  T
}

func NewResult[T any](cols []*ColumnDef, emptyTiming T) *Result[T] {
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

type SyncQueryResult[T any] struct {
	Rows   []interface{}
	Cols   []*ColumnDef
	Timing T
}
