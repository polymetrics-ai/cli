package warehouse_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

func TestWriteArrowSegmentClosesAndReopensTypedParquet(t *testing.T) {
	ctx := context.Background()
	location := segmentManifestLocation(t)
	record := warehouseArrowRecord(t, []int64{11, 12}, []string{"first", "second"})
	defer record.Release()
	segment, err := warehouse.WriteArrowSegment(ctx, location, "events", warehouse.ArrowSegmentRequest{
		ID: "arrow-001", TransformPlanHash: segmentHash("transform"), SourceLogicalBytes: 128, SourceRows: 3,
	}, record)
	if err != nil {
		t.Fatalf("WriteArrowSegment() error = %v", err)
	}
	if segment.Manifest.TransformedRows != 2 || segment.Manifest.SourceRows != 3 || segment.Manifest.ParquetBytes < 1 {
		t.Fatalf("segment accounting = %#v", segment.Manifest)
	}
	if got, want := filepath.Base(segment.Path), "arrow-001.parquet"; got != want {
		t.Fatalf("segment path = %q, want %q", got, want)
	}
	reopened, err := warehouse.OpenArrowSegment(ctx, location, "events", "arrow-001")
	if err != nil {
		t.Fatalf("OpenArrowSegment() error = %v", err)
	}
	if reopened.Manifest != segment.Manifest || reopened.Path != segment.Path {
		t.Fatalf("reopened segment = %#v, want %#v", reopened, segment)
	}
}

func TestWriteArrowSegmentRefusesInvalidAccountingBeforePublishing(t *testing.T) {
	ctx := context.Background()
	location := segmentManifestLocation(t)
	record := warehouseArrowRecord(t, []int64{1}, []string{"one"})
	defer record.Release()
	_, err := warehouse.WriteArrowSegment(ctx, location, "events", warehouse.ArrowSegmentRequest{ID: "invalid", SourceRows: 0}, record)
	if !errors.Is(err, warehouse.ErrTransportSegmentManifestInvalid) {
		t.Fatalf("WriteArrowSegment() error = %T %v, want ErrTransportSegmentManifestInvalid", err, err)
	}
	dir, err := location.SegmentDir("events")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "invalid.parquet")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("invalid segment created a parquet side effect: %v", err)
	}
}

func TestWriteArrowSegmentSupportsZeroRows(t *testing.T) {
	ctx := context.Background()
	location := segmentManifestLocation(t)
	record := warehouseArrowRecord(t, nil, nil)
	defer record.Release()
	segment, err := warehouse.WriteArrowSegment(ctx, location, "events", warehouse.ArrowSegmentRequest{ID: "zero", SourceRows: 0}, record)
	if err != nil {
		t.Fatalf("WriteArrowSegment(zero rows) error = %v", err)
	}
	// A zero-row variable-width Arrow column still has its valid four-byte
	// offset buffer, so zero rows—not zero retained schema/buffer bytes—is the
	// meaningful empty-stream assertion.
	if segment.Manifest.SourceLogicalBytes != 0 || segment.Manifest.TransformedRows != 0 || segment.Manifest.TransformedBytes < 0 {
		t.Fatalf("zero segment accounting = %#v", segment.Manifest)
	}
	if _, err := warehouse.OpenArrowSegment(ctx, location, "events", "zero"); err != nil {
		t.Fatalf("OpenArrowSegment(zero rows) error = %v", err)
	}
}

func TestWriteArrowSegmentAcceptsDuckDBTransformedDateAndStringRecord(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"date":"updated_at"},"target":"event_date","type":"date"},{"expr":{"cast":{"multiply":["amount",100]}},"target":"amount_cents","type":"int64","rounding":"exact"},{"expr":{"upper":"status"},"target":"status","type":"string"}],"where":{"not_equal":[{"mod":["id",2]},0]}}`))
	if err != nil {
		t.Fatal(err)
	}
	input := warehouseDuckDBTransformInput(t)
	defer input.Release()
	transformer, err := database.NewArrowTransformer(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transformer.Close() }()
	output, err := transformer.Transform(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Release()
	if !output.Next() {
		t.Fatalf("DuckDB transform yielded no record: %v", output.Err())
	}
	if _, err := warehouse.WriteArrowSegment(context.Background(), segmentManifestLocation(t), "events", warehouse.ArrowSegmentRequest{ID: "duckdb-date", TransformPlanHash: plan.Hash(), SourceLogicalBytes: 64, SourceRows: 1}, output.Record()); err != nil {
		t.Fatalf("WriteArrowSegment(DuckDB record) = %v", err)
	}
}

func warehouseArrowRecord(t *testing.T, ids []int64, names []string) arrow.Record {
	t.Helper()
	if len(ids) != len(names) {
		t.Fatalf("ids/names length = %d/%d", len(ids), len(names))
	}
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}, {Name: "name", Type: arrow.BinaryTypes.String}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).AppendValues(ids, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues(names, nil)
	record := builder.NewRecord()
	builder.Release()
	return record
}

func warehouseDuckDBTransformInput(t *testing.T) arrow.Record {
	t.Helper()
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}, {Name: "amount", Type: arrow.PrimitiveTypes.Int64}, {Name: "status", Type: arrow.BinaryTypes.String}, {Name: "updated_at", Type: arrow.FixedWidthTypes.Timestamp_ns}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).Append(1)
	builder.Field(1).(*array.Int64Builder).Append(11)
	builder.Field(2).(*array.StringBuilder).Append("new")
	builder.Field(3).(*array.TimestampBuilder).Append(arrow.Timestamp(1_754_000_000_000_000_000))
	record := builder.NewRecord()
	builder.Release()
	return record
}
