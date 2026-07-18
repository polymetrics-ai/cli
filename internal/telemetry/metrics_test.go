package telemetry

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunCountersKeepHotPathLocalUntilFlush(t *testing.T) {
	root := t.TempDir()
	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "counter-locality"}, func(string) {})
	defer Shutdown(context.Background(), handle, func(string) {})
	if !handle.Enabled() {
		t.Fatal("file telemetry disabled, want enabled counters")
	}

	counters := NewRunCounters(ctx)
	if !counters.enabled {
		t.Fatal("NewRunCounters returned disabled counters for enabled telemetry")
	}

	counters.RecordRead()
	counters.RecordTransformed()
	counters.RecordLoaded(2)
	counters.RecordFailed(1)
	counters.RecordBatch()
	if counters.flushedRead != 0 || counters.flushedTransformed != 0 || counters.flushedLoaded != 0 || counters.flushedFailed != 0 || counters.flushedBatches != 0 {
		t.Fatalf("record methods updated flushed counters before batch flush: %+v", counters)
	}

	counters.Flush(ctx)
	if counters.flushedRead != counters.recordsRead || counters.flushedTransformed != counters.recordsTransformed || counters.flushedLoaded != counters.recordsLoaded || counters.flushedFailed != counters.recordsFailed || counters.flushedBatches != counters.batchesFlushed {
		t.Fatalf("flush did not reconcile local/flushed counters: %+v", counters)
	}
}

func TestRunCounterEnabledHotPathAllocations(t *testing.T) {
	root := t.TempDir()
	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "counter-allocs"}, func(string) {})
	defer Shutdown(context.Background(), handle, func(string) {})
	counters := NewRunCounters(ctx)
	if !counters.enabled {
		t.Fatal("NewRunCounters returned disabled counters for enabled telemetry")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		counters.RecordRead()
		counters.RecordTransformed()
		counters.RecordLoaded(1)
	})
	if allocs != 0 {
		t.Fatalf("enabled hot-path counter increments allocated %.2f times per run, want 0", allocs)
	}
}
