package backend

import (
	"github.com/turbot/pipe-fittings/queryresult"
)

type RowReader interface {
	Read(columnValues []any, cols []*queryresult.ColumnDef) ([]any, error)
}

func RowReaderFactory(backend DBClientBackendType) RowReader {
	var reader RowReader
	switch backend {
	case PostgresDBClientBackend:
		reader = &PgxRowReader{}
	case SqliteDBClientBackend:
		reader = &SqliteRowReader{}
	default:

	}
	return reader
}
