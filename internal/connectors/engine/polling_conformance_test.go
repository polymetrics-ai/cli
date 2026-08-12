package engine

import (
	"context"
	"reflect"
	"testing"

	"polymetrics.ai/internal/synccontract"
)

// TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture is the first
// #3856 RED. It deliberately refers to the missing immutable corpus and
// no-skip runner rather than adapting the legacy scalar-cursor test helpers.
func TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture(t *testing.T) {
	report, err := RunPollingWatermarkConformanceSuite(
		context.Background(),
		newReferencePollingWatermarkConformanceLaneFactory(t),
	)
	if err != nil {
		t.Fatalf("RunPollingWatermarkConformanceSuite: %v", err)
	}

	wantIDs := RequiredPollingWatermarkConformanceEvidence().FixtureIDs
	if !reflect.DeepEqual(report.ExecutedFixtureIDs, wantIDs) {
		t.Fatalf("executed fixture IDs = %v, want immutable corpus IDs %v", report.ExecutedFixtureIDs, wantIDs)
	}
	if runner := reflect.TypeOf(RunPollingWatermarkConformanceSuite); runner.NumIn() != 2 || runner.IsVariadic() {
		t.Fatalf("runner signature = %v, want only context and lane factory (no skip/filter input)", runner)
	}

	equalWatermark, ok := report.Fixture("equal-watermark-page-split-recovery")
	if !ok {
		t.Fatal("equal-watermark fixture observation was not reported")
	}
	if got, want := equalWatermark.StableIdentities, []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stable identities = %v, want %v after equal-watermark page replay", got, want)
	}
	if got, want := equalWatermark.SourceRequests, []PollingWatermarkConformancePosition{
		{Watermark: "2026-08-06T10:00:00Z", TieBreaker: "a"},
		{Watermark: "2026-08-06T10:00:00Z", TieBreaker: "b"},
		{Watermark: "2026-08-06T10:00:00Z", TieBreaker: "b"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("source requests = %+v, want composite page/restart tuple requests %+v", got, want)
	}
	if !equalWatermark.ReplayedFromPriorCommitted {
		t.Fatal("acknowledgement followed by persistence failure did not restart from the prior committed envelope")
	}
	if got := len(equalWatermark.PersistedCheckpoints); got != 2 {
		t.Fatalf("persisted checkpoints = %d, want prior committed checkpoint plus recovered checkpoint", got)
	}
	if got := equalWatermark.PersistedCheckpoints[0].Position; !reflect.DeepEqual(got, synccontract.CheckpointPosition{
		Primary:    synccontract.OpaqueToken("2026-08-06T10:00:00Z"),
		TieBreaker: synccontract.OpaqueToken("b"),
	}) {
		t.Fatalf("prior committed position = %+v, want composite tuple through b", got)
	}
	if got := equalWatermark.PersistedCheckpoints[1].Position; !reflect.DeepEqual(got, synccontract.CheckpointPosition{
		Primary:    synccontract.OpaqueToken("2026-08-06T10:00:00Z"),
		TieBreaker: synccontract.OpaqueToken("c"),
	}) {
		t.Fatalf("recovered committed position = %+v, want composite tuple through c", got)
	}
}
