package postgres

import (
	"errors"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/jackc/pgx/v5"
)

var errPostgresArrowCopyValueInvalid = errors.New("PostgreSQL Arrow binary COPY value is invalid")

// postgresArrowCopyFromSource is the reusable pgx binary-COPY source for one
// transformed Arrow record. It reuses one []any value vector and creates no
// Go map or struct per output row. pgx.Conn.CopyFrom chooses PostgreSQL's
// binary COPY protocol; this adapter intentionally has no INSERT fallback.
type postgresArrowCopyFromSource struct {
	record arrow.Record
	row    int
	values []any
	err    error
}

func newPostgresArrowCopyFromSource(record arrow.Record) *postgresArrowCopyFromSource {
	return &postgresArrowCopyFromSource{record: record, row: -1, values: make([]any, int(record.NumCols()))}
}

func (s *postgresArrowCopyFromSource) Next() bool {
	if s == nil || s.record == nil || s.err != nil {
		return false
	}
	s.row++
	return s.row < int(s.record.NumRows())
}

func (s *postgresArrowCopyFromSource) Values() ([]any, error) {
	if s == nil || s.record == nil || s.row < 0 || s.row >= int(s.record.NumRows()) {
		return nil, errPostgresArrowCopyValueInvalid
	}
	for index := range s.values {
		value, err := postgresArrowCopyValue(s.record.Column(index), s.row)
		if err != nil {
			s.err = err
			return nil, err
		}
		s.values[index] = value
	}
	return s.values, nil
}

func (s *postgresArrowCopyFromSource) Err() error {
	if s == nil {
		return errPostgresArrowCopyValueInvalid
	}
	return s.err
}

func postgresArrowCopyValue(column arrow.Array, row int) (any, error) {
	if column == nil || row < 0 || row >= column.Len() {
		return nil, errPostgresArrowCopyValueInvalid
	}
	if column.IsNull(row) {
		return nil, nil
	}
	switch typed := column.(type) {
	case *array.Int64:
		return typed.Value(row), nil
	case *array.String:
		return typed.Value(row), nil
	case *array.Date32:
		return time.Unix(0, 0).UTC().AddDate(0, 0, int(typed.Value(row))), nil
	case *array.Timestamp:
		timestampType, ok := typed.DataType().(*arrow.TimestampType)
		if !ok {
			return nil, errPostgresArrowCopyValueInvalid
		}
		return typed.Value(row).ToTime(timestampType.Unit), nil
	default:
		return nil, errPostgresArrowCopyValueInvalid
	}
}

var _ pgx.CopyFromSource = (*postgresArrowCopyFromSource)(nil)
