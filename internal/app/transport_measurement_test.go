package app

import (
	"testing"
	"time"

	"polymetrics.ai/internal/synctransport"
)

func TestTransportPhaseMeasurementReportsLogicalInputThroughputAndEveryFastPathPhase(t *testing.T) {
	measurement := transportPhaseMeasurement(synctransport.Result{
		RecordsRead: 7, RecordsStaged: 5, RecordsApplied: 5, Pages: 2,
		SourceLogicalBytes: 3_000_000_000, TransformedBytes: 2_000_000_000, ParquetBytes: 800_000_000,
		ExtractElapsed: 4 * time.Second, TransformElapsed: 3 * time.Second, ParquetElapsed: 2 * time.Second,
		ApplyElapsed: time.Second, ReadBackElapsed: 600 * time.Millisecond, IndexConstraintElapsed: 750 * time.Millisecond, PublishElapsed: 500 * time.Millisecond, CheckpointElapsed: 250 * time.Millisecond,
		WallElapsed: 10 * time.Second, PeakCreditBytes: 512 << 20, CreditWaitElapsed: 125 * time.Millisecond,
	})
	if measurement.SourceRecords != 7 || measurement.TransformedRecords != 5 || measurement.CopyAppliedRecords != 5 || measurement.SourceLogicalBytes != 3_000_000_000 || measurement.ParquetBytes != 800_000_000 {
		t.Fatalf("fast-path counters = %#v, want source/transform/COPY/parquet values", measurement)
	}
	if measurement.SourceReadElapsedNanos != (4*time.Second).Nanoseconds() || measurement.TransformElapsedNanos != (3*time.Second).Nanoseconds() || measurement.ParquetCloseElapsedNanos != (2*time.Second).Nanoseconds() || measurement.BinaryCOPYElapsedNanos != (time.Second).Nanoseconds() || measurement.ReadBackElapsedNanos != (600*time.Millisecond).Nanoseconds() || measurement.IndexConstraintBuildElapsedNanos != (750*time.Millisecond).Nanoseconds() || measurement.PublishReceiptElapsedNanos != (500*time.Millisecond).Nanoseconds() || measurement.CheckpointElapsedNanos != (250*time.Millisecond).Nanoseconds() || measurement.CriticalPathElapsedNanos != (10*time.Second).Nanoseconds() {
		t.Fatalf("fast-path phase durations = %#v, want every measured phase", measurement)
	}
	if measurement.InputDecimalMBPerSecond != 300 || measurement.InputMiBPerSecond < 286.10 || measurement.InputMiBPerSecond > 286.11 {
		t.Fatalf("logical input throughput = decimal %.2f MiB %.5f, want 300.00 and ~286.102", measurement.InputDecimalMBPerSecond, measurement.InputMiBPerSecond)
	}
}

func TestTransportPhaseMeasurementHasZeroThroughputWithoutCriticalPath(t *testing.T) {
	measurement := transportPhaseMeasurement(synctransport.Result{SourceLogicalBytes: 3_000_000_000})
	if measurement.InputDecimalMBPerSecond != 0 || measurement.InputMiBPerSecond != 0 {
		t.Fatalf("zero critical path throughput = %#v, want no fabricated rate", measurement)
	}
}
