package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"

	"polymetrics.ai/internal/connectors/database"
)

func TestTransformArrowRecordEvaluatesClosedProjectionAndFilter(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"upper":"status"},"target":"status","type":"string"}],"where":{"not_equal":[{"mod":["id",2]},0]}}`))
	if err != nil {
		t.Fatal(err)
	}
	input := transformArrowInput(t, []int64{1, 2, 3}, []string{"new", "ignored", "done"})
	defer input.Release()
	output, err := database.TransformArrowRecord(context.Background(), plan, input)
	if err != nil {
		t.Fatalf("TransformArrowRecord() error = %v", err)
	}
	defer output.Release()
	if !output.Next() {
		t.Fatalf("TransformArrowRecord() yielded no output: %v", output.Err())
	}
	record := output.Record()
	if got, want := record.NumRows(), int64(2); got != want {
		t.Fatalf("output rows = %d, want %d", got, want)
	}
	ids, ok := record.Column(0).(*array.Int64)
	if !ok || ids.Value(0) != 1 || ids.Value(1) != 3 {
		t.Fatalf("output typed event ids = %T %v, want [1 3]", record.Column(0), record.Column(0))
	}
	statuses, ok := record.Column(1).(*array.String)
	if !ok || statuses.Value(0) != "NEW" || statuses.Value(1) != "DONE" {
		t.Fatalf("output typed statuses = %T %v, want [NEW DONE]", record.Column(1), record.Column(1))
	}
	if output.Next() {
		t.Fatal("TransformArrowRecord() returned an unexpected second batch")
	}
	if err := output.Err(); err != nil {
		t.Fatalf("TransformArrowRecord() iterator error = %v", err)
	}
}

func TestTransformArrowRecordRefusesMissingInputFieldBeforeDuckDBIO(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"expr":{"upper":"status"},"target":"status","type":"string"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).Append(1)
	input := builder.NewRecord()
	builder.Release()
	defer input.Release()
	if _, err := database.TransformArrowRecord(context.Background(), plan, input); !errors.Is(err, database.ErrTransformPlanInvalid) {
		t.Fatalf("TransformArrowRecord(missing field) error = %T %v, want ErrTransformPlanInvalid", err, err)
	}
}

func TestTransformArrowRecordPreservesZeroRowBatch(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	input := transformArrowInput(t, nil, nil)
	defer input.Release()
	output, err := database.TransformArrowRecord(context.Background(), plan, input)
	if err != nil {
		t.Fatalf("TransformArrowRecord(zero rows) error = %v", err)
	}
	defer output.Release()
	if output.Next() && output.Record().NumRows() != 0 {
		t.Fatalf("zero-row transform yielded %d rows", output.Record().NumRows())
	}
	if err := output.Err(); err != nil {
		t.Fatalf("zero-row transform iterator error = %v", err)
	}
}

// A transport run constructs one transformer and reuses it for every bounded
// Arrow range. Opening DuckDB per range would turn a 5 GB test into a process
// startup benchmark instead of measuring vector transformation.
func TestArrowTransformerReusesOneDuckDBConnectionForMultipleBatches(t *testing.T) {
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"id","type":"int64"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	transformer, err := database.NewArrowTransformer(context.Background(), plan)
	if err != nil {
		t.Fatalf("NewArrowTransformer() = %v", err)
	}
	defer func() {
		if err := transformer.Close(); err != nil {
			t.Fatalf("Close() = %v", err)
		}
	}()
	for _, value := range []int64{41, 42} {
		input := transformArrowInput(t, []int64{value}, []string{"x"})
		output, err := transformer.Transform(context.Background(), input)
		input.Release()
		if err != nil {
			t.Fatalf("Transform(%d) = %v", value, err)
		}
		if !output.Next() || output.Record().Column(0).(*array.Int64).Value(0) != value {
			output.Release()
			t.Fatalf("Transform(%d) output = %#v, want retained transformer result", value, output.Record())
		}
		output.Release()
	}
}

func transformArrowInput(t *testing.T, ids []int64, statuses []string) arrow.Record {
	t.Helper()
	if len(ids) != len(statuses) {
		t.Fatalf("input ids/statuses length = %d/%d", len(ids), len(statuses))
	}
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}, {Name: "status", Type: arrow.BinaryTypes.String}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).AppendValues(ids, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues(statuses, nil)
	record := builder.NewRecord()
	builder.Release()
	return record
}
