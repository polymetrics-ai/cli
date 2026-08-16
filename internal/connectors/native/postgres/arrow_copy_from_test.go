package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

func TestPostgresArrowCopyFromSourceReusesTypedRowVector(t *testing.T) {
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}, {Name: "label", Type: arrow.BinaryTypes.String}, {Name: "event_date", Type: arrow.PrimitiveTypes.Date32}, {Name: "updated_at", Type: arrow.FixedWidthTypes.Timestamp_us}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"ONE", "TWO"}, nil)
	builder.Field(2).(*array.Date32Builder).AppendValues([]arrow.Date32{0, 1}, nil)
	builder.Field(3).(*array.TimestampBuilder).AppendValues([]arrow.Timestamp{0, arrow.Timestamp((24 * time.Hour).Microseconds())}, nil)
	record := builder.NewRecord()
	builder.Release()
	defer record.Release()

	source := newPostgresArrowCopyFromSource(record)
	if !source.Next() {
		t.Fatal("Next() = false, want first transformed Arrow row")
	}
	first, err := source.Values()
	if err != nil || first[0] != int64(1) || first[1] != "ONE" || !first[2].(time.Time).Equal(time.Unix(0, 0).UTC()) || !first[3].(time.Time).Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("first binary COPY values = %#v, %v; want typed transformed row", first, err)
	}
	if !source.Next() {
		t.Fatal("Next() = false, want second transformed Arrow row")
	}
	second, err := source.Values()
	if err != nil || second[0] != int64(2) || second[1] != "TWO" || !second[2].(time.Time).Equal(time.Unix(0, 0).UTC().AddDate(0, 0, 1)) || !second[3].(time.Time).Equal(time.Unix(0, 0).UTC().AddDate(0, 0, 1)) {
		t.Fatalf("second binary COPY values = %#v, %v; want typed transformed row", second, err)
	}
	if &first[0] != &second[0] {
		t.Fatal("COPY Values() allocated a new row vector, want one reusable vector")
	}
}

func TestPostgresArrowCopyFromSourceRefusesUnsupportedOutputType(t *testing.T) {
	schema := arrow.NewSchema([]arrow.Field{{Name: "unsupported", Type: arrow.PrimitiveTypes.Float64}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Float64Builder).Append(1.5)
	record := builder.NewRecord()
	builder.Release()
	defer record.Release()
	source := newPostgresArrowCopyFromSource(record)
	if !source.Next() {
		t.Fatal("Next() = false, want source row")
	}
	if _, err := source.Values(); !errors.Is(err, errPostgresArrowCopyValueInvalid) || !errors.Is(source.Err(), errPostgresArrowCopyValueInvalid) {
		t.Fatalf("Values() unsupported type error = %T %v, want errPostgresArrowCopyValueInvalid", err, err)
	}
}

func TestPostgresArrowCopyFromSourceHandlesZeroRowsAndNulls(t *testing.T) {
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	empty := builder.NewRecord()
	if newPostgresArrowCopyFromSource(empty).Next() {
		empty.Release()
		builder.Release()
		t.Fatal("zero-row COPY source advanced")
	}
	empty.Release()
	builder.Field(0).(*array.Int64Builder).AppendNull()
	nullRecord := builder.NewRecord()
	builder.Release()
	defer nullRecord.Release()
	source := newPostgresArrowCopyFromSource(nullRecord)
	if !source.Next() {
		t.Fatal("Next() = false, want nullable row")
	}
	values, err := source.Values()
	if err != nil || values[0] != nil {
		t.Fatalf("nullable COPY value = %#v, %v; want nil", values, err)
	}
}
