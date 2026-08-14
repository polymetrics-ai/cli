package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestPollingWatermarkConformanceSuiteRejectsUnregisteredLane(t *testing.T) {
	_, err := RunPollingWatermarkConformanceSuite(context.Background(), unregisteredPollingWatermarkConformanceFactory{})
	if !errors.Is(err, ErrPollingWatermarkConformanceUnregistered) {
		t.Fatalf("RunPollingWatermarkConformanceSuite error = %v, want unregistered lane rejection", err)
	}
}

func TestPollingWatermarkConformanceRegistrationRejectsWrongEvidence(t *testing.T) {
	evidence := RequiredPollingWatermarkConformanceEvidence()
	evidence.FixtureIDs = evidence.FixtureIDs[:len(evidence.FixtureIDs)-1]
	_, err := NewPollingWatermarkConformanceRegistration(referencePollingWatermarkConformanceDescriptor(), evidence)
	if !errors.Is(err, ErrPollingWatermarkConformanceEvidence) {
		t.Fatalf("NewPollingWatermarkConformanceRegistration error = %v, want immutable evidence rejection", err)
	}
}

func TestPollingWatermarkConformanceRegistrationRejectsUnsafeDescriptor(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*PollingWatermarkConformanceDescriptor)
	}{
		{
			name: "unstable keyset",
			mutate: func(descriptor *PollingWatermarkConformanceDescriptor) {
				descriptor.StableKeyset = false
			},
		},
		{
			name: "lossy cursor policy",
			mutate: func(descriptor *PollingWatermarkConformanceDescriptor) {
				descriptor.CursorPolicy = "lossy"
			},
		},
		{
			name: "unbounded overlap",
			mutate: func(descriptor *PollingWatermarkConformanceDescriptor) {
				descriptor.BoundedOverlap = false
			},
		},
		{
			name: "unbounded commit lag",
			mutate: func(descriptor *PollingWatermarkConformanceDescriptor) {
				descriptor.BoundedCommitLag = false
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			descriptor := referencePollingWatermarkConformanceDescriptor()
			testCase.mutate(&descriptor)
			_, err := NewPollingWatermarkConformanceRegistration(descriptor, RequiredPollingWatermarkConformanceEvidence())
			if !errors.Is(err, ErrPollingWatermarkConformanceUnregistered) {
				t.Fatalf("NewPollingWatermarkConformanceRegistration error = %v, want unsafe descriptor rejection", err)
			}
		})
	}
}

func TestPollingWatermarkConformanceFixturesAreDefensiveAndSeparate(t *testing.T) {
	beforeGeneric := synccontract.RequiredConformanceEvidence()
	fixtures := PollingWatermarkConformanceFixtures()
	fixtures[0].ID = "mutated"
	fixtures[0].Pages[0].Records[0].StableIdentity = "mutated"
	fixtures[0].Expected.StableIdentities[0] = "mutated"
	for i := range fixtures {
		if fixtures[i].ID == "null-precision-coercion-is-rejected" {
			fixtures[i].CursorSamples[1].Value = json.RawMessage(`"mutated"`)
		}
	}

	fresh := PollingWatermarkConformanceFixtures()
	if fresh[0].ID == "mutated" || fresh[0].Pages[0].Records[0].StableIdentity == "mutated" || fresh[0].Expected.StableIdentities[0] == "mutated" {
		t.Fatal("polling conformance fixtures expose mutable embedded corpus state")
	}
	for _, fixture := range fresh {
		if fixture.ID == "null-precision-coercion-is-rejected" && string(fixture.CursorSamples[1].Value) == `"mutated"` {
			t.Fatal("polling conformance fixtures expose mutable raw cursor values")
		}
	}
	if afterGeneric := synccontract.RequiredConformanceEvidence(); !reflect.DeepEqual(afterGeneric, beforeGeneric) {
		t.Fatalf("generic #3810 evidence changed from %+v to %+v", beforeGeneric, afterGeneric)
	}
}

func TestPollingWatermarkConformanceSuiteRejectsBehaviorMismatch(t *testing.T) {
	_, err := RunPollingWatermarkConformanceSuite(context.Background(), corruptingPollingWatermarkConformanceFactory{target: "tombstone-history-and-hard-delete-visibility"})
	if err == nil || !strings.Contains(err.Error(), "tombstone identities") {
		t.Fatalf("RunPollingWatermarkConformanceSuite error = %v, want behavior mismatch rejection", err)
	}
}

func TestPollingWatermarkConformanceSuiteNeverRegressesDurableCheckpointForOverlap(t *testing.T) {
	report, err := RunPollingWatermarkConformanceSuite(
		context.Background(),
		newReferencePollingWatermarkConformanceLaneFactory(t),
	)
	if err != nil {
		t.Fatalf("RunPollingWatermarkConformanceSuite: %v", err)
	}

	overlap, ok := report.Fixture("bounded-overlap-and-commit-lag")
	if !ok {
		t.Fatal("bounded-overlap fixture observation was not reported")
	}
	if got, want := len(overlap.PersistedCheckpoints), 1; got != want {
		t.Fatalf("persisted checkpoints = %d, want %d", got, want)
	}
	if got, want := overlap.PersistedCheckpoints[0].Position, (synccontract.CheckpointPosition{
		Primary:    synccontract.OpaqueToken("2026-08-06T10:00:00Z"),
		TieBreaker: synccontract.OpaqueToken("a"),
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("bounded overlap regressed durable checkpoint to %+v, want %+v", got, want)
	}
}

func TestBoundedOverlapReferenceLaneDerivesTheOverlapRequest(t *testing.T) {
	var fixture PollingWatermarkConformanceFixture
	for _, candidate := range PollingWatermarkConformanceFixtures() {
		if candidate.ID == "bounded-overlap-and-commit-lag" {
			fixture = candidate
			break
		}
	}
	if fixture.ID == "" {
		t.Fatal("bounded-overlap fixture was not loaded")
	}
	fixture.Expected.SourceRequests[0] = fixture.InitialCheckpoint.PollingWatermarkConformancePosition

	if _, err := runBoundedOverlapCommitLag(fixture); err == nil || !strings.Contains(err.Error(), "overlap source request") {
		t.Fatalf("runBoundedOverlapCommitLag error = %v, want copied expectation rejection", err)
	}
}

func TestPollingWatermarkConformanceSuiteRejectsUntypedRecoveryObservation(t *testing.T) {
	_, err := RunPollingWatermarkConformanceSuite(context.Background(), corruptingPollingWatermarkConformanceFactory{
		target:        "schema-fingerprint-mismatch-requires-rebootstrap",
		recoveryError: errors.New("untyped schema mismatch"),
	})
	if err == nil || !strings.Contains(err.Error(), "typed rebootstrap") {
		t.Fatalf("RunPollingWatermarkConformanceSuite error = %v, want typed recovery rejection", err)
	}
}

func TestPollingWatermarkConformanceSuiteRejectsCursorPolicyResultMismatch(t *testing.T) {
	_, err := RunPollingWatermarkConformanceSuite(context.Background(), corruptingPollingWatermarkConformanceFactory{
		target:                  "null-precision-coercion-is-rejected",
		dropCursorSampleResults: true,
	})
	if err == nil || !strings.Contains(err.Error(), "cursor sample results") {
		t.Fatalf("RunPollingWatermarkConformanceSuite error = %v, want raw cursor-policy result rejection", err)
	}
}

func TestPollingWatermarkConformanceSuiteRejectsPersistedCheckpointDescriptorMismatch(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*synccontract.CheckpointEnvelope)
		want   string
	}{
		{
			name: "source identity",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.Source.AccountOrCluster = "other-fixture-account"
			},
			want: "source identity",
		},
		{
			name: "source generation",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.SourceGeneration = synccontract.OpaqueToken("other-generation")
			},
			want: "source generation",
		},
		{
			name: "schema fingerprint",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.SchemaVersion = "other-schema"
			},
			want: "schema fingerprint",
		},
		{
			name: "mechanism",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.Mechanism = "other-mechanism"
			},
			want: "mechanism",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := RunPollingWatermarkConformanceSuite(context.Background(), corruptingPollingWatermarkConformanceFactory{
				target:           "equal-watermark-page-split-recovery",
				mutateCheckpoint: testCase.mutate,
			})
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("RunPollingWatermarkConformanceSuite error = %v, want persisted checkpoint %s rejection", err, testCase.want)
			}
		})
	}
}

type referencePollingWatermarkConformanceFactory struct {
	registration PollingWatermarkConformanceRegistration
}

func newReferencePollingWatermarkConformanceLaneFactory(t *testing.T) PollingWatermarkConformanceLaneFactory {
	t.Helper()
	registration, err := newReferencePollingWatermarkConformanceRegistration()
	if err != nil {
		t.Fatalf("NewPollingWatermarkConformanceRegistration: %v", err)
	}
	return referencePollingWatermarkConformanceFactory{registration: registration}
}

func newReferencePollingWatermarkConformanceRegistration() (PollingWatermarkConformanceRegistration, error) {
	return NewPollingWatermarkConformanceRegistration(
		referencePollingWatermarkConformanceDescriptor(),
		RequiredPollingWatermarkConformanceEvidence(),
	)
}

func (f referencePollingWatermarkConformanceFactory) NewPollingWatermarkConformanceLane(context.Context) (PollingWatermarkConformanceLane, error) {
	return referencePollingWatermarkConformanceLane(f), nil
}

type referencePollingWatermarkConformanceLane struct {
	registration PollingWatermarkConformanceRegistration
}

var _ PollingWatermarkConformanceLane = referencePollingWatermarkConformanceLane{}

func (l referencePollingWatermarkConformanceLane) PollingWatermarkConformanceRegistration() PollingWatermarkConformanceRegistration {
	return l.registration
}

func (l referencePollingWatermarkConformanceLane) RunPollingWatermarkConformance(ctx context.Context, fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if err := ctx.Err(); err != nil {
		return PollingWatermarkConformanceObservation{}, err
	}
	switch fixture.Scenario {
	case "equal_watermark_page_split_recovery":
		return runEqualWatermarkPageSplitRecovery(fixture)
	case "empty_page_no_advance":
		return runEmptyPageNoAdvance(fixture)
	case "non_advancing_page_rejected":
		return runNonAdvancingPageRejected(fixture)
	case "null_precision_coercion_rejected":
		return runNullPrecisionCoercionPolicy(fixture)
	case "unstable_keyset_rejected":
		return runUnstableKeysetRejected(fixture)
	case "bounded_overlap_commit_lag_safe":
		return runBoundedOverlapCommitLag(fixture)
	case "source_generation_mismatch":
		return runSourceGenerationMismatch(fixture)
	case "schema_fingerprint_mismatch":
		return runSchemaFingerprintMismatch(fixture)
	case "acknowledged_checkpoint_failure_replays":
		return runAcknowledgedCheckpointFailureReplay(fixture)
	case "tombstone_history_hard_delete_invisibility":
		return runTombstoneHistoryHardDeleteVisibility(fixture)
	case "missing_executor_or_evidence_rejected":
		return runMissingExecutorOrEvidenceAdmission(fixture)
	default:
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("unexpected conformance scenario %q", fixture.Scenario)
	}
}

func referencePollingWatermarkConformanceDescriptor() PollingWatermarkConformanceDescriptor {
	return PollingWatermarkConformanceDescriptor{
		Mechanism:         pollingWatermarkConformanceMechanism,
		ExecutorID:        "polling-watermark-reference",
		SourceEngine:      "fixture-native",
		SourceAccount:     "fixture-account",
		SourceScope:       "widgets",
		SourceGeneration:  "generation-1",
		SchemaFingerprint: "schema-v1",
		StableKeyset:      true,
		CursorPolicy:      "lossless",
		BoundedOverlap:    true,
		BoundedCommitLag:  true,
		DeleteVisibility:  "hard_delete_invisible",
	}
}

func basePollingWatermarkConformanceObservation(fixture PollingWatermarkConformanceFixture) PollingWatermarkConformanceObservation {
	return PollingWatermarkConformanceObservation{
		FixtureID:                  fixture.ID,
		Outcome:                    fixture.Expected.Outcome,
		StableIdentities:           append([]string(nil), fixture.Expected.StableIdentities...),
		SourceRequests:             append([]PollingWatermarkConformancePosition(nil), fixture.Expected.SourceRequests...),
		ReplayedFromPriorCommitted: fixture.Expected.ReplayedFromPriorCommitted,
		CheckpointUnchanged:        fixture.Expected.CheckpointUnchanged,
		RecoveryOutcome:            synccontract.RecoveryOutcome(fixture.Expected.RecoveryOutcome),
		TombstoneIdentities:        append([]string(nil), fixture.Expected.TombstoneIdentities...),
		HistoryClosedIdentities:    append([]string(nil), fixture.Expected.HistoryClosedIdentities...),
		HardDeleteInvisible:        fixture.Expected.HardDeleteInvisible,
		AdmissionRejected:          fixture.Expected.AdmissionRejected,
	}
}

func runNullPrecisionCoercionPolicy(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.Descriptor.CursorPolicy != "reject_null_precision_coercion" || len(fixture.CursorSamples) == 0 {
		return PollingWatermarkConformanceObservation{}, errors.New("NULL/precision/coercion fixture is incomplete")
	}
	results := make([]PollingWatermarkConformanceCursorSampleResult, 0, len(fixture.CursorSamples))
	for _, sample := range fixture.CursorSamples {
		value, err := pollingWatermarkConformanceCursorSampleValue(sample)
		var scalar string
		if err == nil {
			switch sample.Field {
			case "watermark":
				scalar, err = pollingWatermarkScalar(value)
				if err == nil {
					err = validatePollingWatermarkValue("timestamp", scalar)
				}
			case "tie_breaker":
				scalar, _, err = pollingWatermarkTieBreakerScalar(value)
			default:
				err = fmt.Errorf("unexpected cursor sample field %q", sample.Field)
			}
		}
		if sample.Accepted {
			if err != nil {
				return PollingWatermarkConformanceObservation{}, fmt.Errorf("accept cursor sample %q: %w", sample.ID, err)
			}
			if scalar != sample.ExactValue {
				return PollingWatermarkConformanceObservation{}, fmt.Errorf("cursor sample %q value = %q, want exact %q", sample.ID, scalar, sample.ExactValue)
			}
			results = append(results, PollingWatermarkConformanceCursorSampleResult{ID: sample.ID, Accepted: true, ExactValue: scalar})
			continue
		}
		if err == nil {
			return PollingWatermarkConformanceObservation{}, fmt.Errorf("cursor sample %q was accepted after %s policy", sample.ID, sample.Policy)
		}
		results = append(results, PollingWatermarkConformanceCursorSampleResult{ID: sample.ID})
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.CursorSampleResults = results
	return observation, nil
}

func pollingWatermarkConformanceCursorSampleValue(sample PollingWatermarkConformanceCursorSample) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(sample.Value))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode cursor sample %q: %w", sample.ID, err)
	}
	if sample.Encoding == "json" {
		return value, nil
	}
	number, ok := value.(json.Number)
	if !ok {
		return nil, fmt.Errorf("float64 cursor sample %q requires a JSON number", sample.ID)
	}
	coerced, err := number.Float64()
	if err != nil {
		return nil, fmt.Errorf("coerce cursor sample %q to float64: %w", sample.ID, err)
	}
	return coerced, nil
}

func runUnstableKeysetRejected(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.Descriptor.StableKeyset {
		return PollingWatermarkConformanceObservation{}, errors.New("unstable/non-unique keyset was admitted")
	}
	if len(fixture.Pages) != 1 || fixture.Pages[0].More || len(fixture.Pages[0].Records) != 2 {
		return PollingWatermarkConformanceObservation{}, errors.New("unstable keyset fixture is incomplete")
	}
	first, second := fixture.Pages[0].Records[0], fixture.Pages[0].Records[1]
	if first.StableIdentity == second.StableIdentity || positionFromPollingWatermarkConformanceRecord(first) != positionFromPollingWatermarkConformanceRecord(second) {
		return PollingWatermarkConformanceObservation{}, errors.New("unstable keyset fixture does not contain a non-unique ordering tuple")
	}
	return basePollingWatermarkConformanceObservation(fixture), nil
}

func runEqualWatermarkPageSplitRecovery(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || len(fixture.Pages) != 2 || len(fixture.Acknowledgements) != 3 {
		return PollingWatermarkConformanceObservation{}, errors.New("equal-watermark fixture is incomplete")
	}
	firstPage, secondPage := fixture.Pages[0], fixture.Pages[1]
	if !firstPage.More || secondPage.More || len(firstPage.Records) != 2 || len(secondPage.Records) != 2 {
		return PollingWatermarkConformanceObservation{}, errors.New("equal-watermark physical page boundaries are invalid")
	}
	if firstPage.Records[1].Watermark != secondPage.Records[0].Watermark || firstPage.Records[1].StableIdentity != secondPage.Records[0].StableIdentity {
		return PollingWatermarkConformanceObservation{}, errors.New("equal-watermark page edge does not replay the same stable identity")
	}

	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.StableIdentities = uniquePollingWatermarkConformanceIdentities(firstPage.Records, secondPage.Records, secondPage.Records)
	observation.SourceRequests = []PollingWatermarkConformancePosition{
		fixture.InitialCheckpoint.PollingWatermarkConformancePosition,
		positionFromPollingWatermarkConformanceRecord(firstPage.Records[len(firstPage.Records)-1]),
		positionFromPollingWatermarkConformanceRecord(firstPage.Records[len(firstPage.Records)-1]),
	}
	first, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, positionFromPollingWatermarkConformanceRecord(firstPage.Records[len(firstPage.Records)-1]), fixture.Acknowledgements[0])
	if err != nil || !persisted {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("persist first physical page: %w", err)
	}
	if _, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, positionFromPollingWatermarkConformanceRecord(secondPage.Records[len(secondPage.Records)-1]), fixture.Acknowledgements[1]); err == nil || persisted {
		return PollingWatermarkConformanceObservation{}, errors.New("checkpoint persistence failure after acknowledgement was not observed")
	}
	last, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, positionFromPollingWatermarkConformanceRecord(secondPage.Records[len(secondPage.Records)-1]), fixture.Acknowledgements[2])
	if err != nil || !persisted {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("persist replayed page: %w", err)
	}
	observation.PersistedCheckpoints = []synccontract.CheckpointEnvelope{first, last}
	observation.ReplayedFromPriorCommitted = true
	return observation, nil
}

func runEmptyPageNoAdvance(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || len(fixture.Pages) != 1 || len(fixture.Pages[0].Records) != 0 || fixture.Pages[0].More {
		return PollingWatermarkConformanceObservation{}, errors.New("empty-page fixture is incomplete")
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.SourceRequests = []PollingWatermarkConformancePosition{fixture.InitialCheckpoint.PollingWatermarkConformancePosition}
	observation.CheckpointUnchanged = true
	return observation, nil
}

func runNonAdvancingPageRejected(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || len(fixture.Pages) != 1 || !fixture.Pages[0].More || len(fixture.Pages[0].Records) != 1 {
		return PollingWatermarkConformanceObservation{}, errors.New("non-advancing fixture is incomplete")
	}
	position := positionFromPollingWatermarkConformanceRecord(fixture.Pages[0].Records[0])
	if position != fixture.InitialCheckpoint.PollingWatermarkConformancePosition {
		return PollingWatermarkConformanceObservation{}, errors.New("non-advancing fixture does not repeat the persisted composite tuple")
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.SourceRequests = []PollingWatermarkConformancePosition{fixture.InitialCheckpoint.PollingWatermarkConformancePosition}
	observation.CheckpointUnchanged = true
	return observation, nil
}

func runBoundedOverlapCommitLag(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || !fixture.Descriptor.StableKeyset || !fixture.Descriptor.BoundedOverlap || !fixture.Descriptor.BoundedCommitLag {
		return PollingWatermarkConformanceObservation{}, errors.New("bounded overlap/commit-lag declaration is unsafe")
	}
	unsafe := fixture.Descriptor
	unsafe.BoundedCommitLag = false
	if referencePollingWatermarkConformanceDescriptorAdmitted(unsafe) {
		return PollingWatermarkConformanceObservation{}, errors.New("unbounded commit lag was admitted")
	}
	if len(fixture.Pages) != 1 || len(fixture.Pages[0].Records) != 1 || len(fixture.Acknowledgements) != 1 {
		return PollingWatermarkConformanceObservation{}, errors.New("bounded overlap fixture is incomplete")
	}
	overlapRequest, err := boundedOverlapSourceRequest(fixture.InitialCheckpoint.PollingWatermarkConformancePosition, fixture.Pages[0].Records[0])
	if err != nil {
		return PollingWatermarkConformanceObservation{}, err
	}
	if len(fixture.Expected.SourceRequests) != 1 || fixture.Expected.SourceRequests[0] != overlapRequest {
		return PollingWatermarkConformanceObservation{}, errors.New("bounded overlap fixture does not declare the derived overlap source request")
	}
	checkpoint, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, fixture.InitialCheckpoint.PollingWatermarkConformancePosition, fixture.Acknowledgements[0])
	if err != nil || !persisted {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("persist bounded-overlap page: %w", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.SourceRequests = []PollingWatermarkConformancePosition{overlapRequest}
	observation.StableIdentities = uniquePollingWatermarkConformanceIdentities(fixture.Pages[0].Records)
	observation.PersistedCheckpoints = []synccontract.CheckpointEnvelope{checkpoint}
	return observation, nil
}

func boundedOverlapSourceRequest(durable PollingWatermarkConformancePosition, record PollingWatermarkConformanceRecord) (PollingWatermarkConformancePosition, error) {
	durableAt, err := time.Parse(time.RFC3339Nano, durable.Watermark)
	if err != nil {
		return PollingWatermarkConformancePosition{}, fmt.Errorf("parse durable overlap checkpoint: %w", err)
	}
	recordAt, err := time.Parse(time.RFC3339Nano, record.Watermark)
	if err != nil {
		return PollingWatermarkConformancePosition{}, fmt.Errorf("parse overlap source record: %w", err)
	}
	if !recordAt.Before(durableAt) {
		return PollingWatermarkConformancePosition{}, errors.New("bounded overlap record must precede the durable checkpoint")
	}
	return PollingWatermarkConformancePosition{Watermark: record.Watermark}, nil
}

func runSourceGenerationMismatch(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil {
		return PollingWatermarkConformanceObservation{}, errors.New("source-generation mismatch fixture lacks a checkpoint")
	}
	err := pollingWatermarkConformanceResumeFailure(fixture)
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("ValidateResume error = %v, want source-generation rebootstrap", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.CheckpointUnchanged = true
	observation.RecoveryOutcome = recovery.Outcome
	observation.RecoveryError = err
	return observation, nil
}

func runSchemaFingerprintMismatch(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || fixture.InitialCheckpoint.SchemaFingerprint == fixture.Descriptor.SchemaFingerprint {
		return PollingWatermarkConformanceObservation{}, errors.New("schema fingerprint mismatch fixture is incomplete")
	}
	err := pollingWatermarkConformanceResumeFailure(fixture)
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("schema resume error = %v, want invalid-checkpoint rebootstrap", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.CheckpointUnchanged = true
	observation.RecoveryOutcome = recovery.Outcome
	observation.RecoveryError = err
	return observation, nil
}

func pollingWatermarkConformanceResumeFailure(fixture PollingWatermarkConformanceFixture) error {
	checkpoint, err := pollingWatermarkConformanceEnvelope(fixture, fixture.InitialCheckpoint.PollingWatermarkConformancePosition, true)
	if err != nil {
		return err
	}
	if err := checkpoint.ValidateResume(pollingWatermarkConformanceResumeExpectation(fixture.Descriptor)); err != nil {
		return err
	}
	if checkpoint.SchemaVersion != fixture.Descriptor.SchemaFingerprint {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint schema fingerprint no longer matches")
	}
	return nil
}

func runAcknowledgedCheckpointFailureReplay(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || len(fixture.Pages) != 1 || len(fixture.Pages[0].Records) != 1 || len(fixture.Acknowledgements) != 2 {
		return PollingWatermarkConformanceObservation{}, errors.New("acknowledgement replay fixture is incomplete")
	}
	position := positionFromPollingWatermarkConformanceRecord(fixture.Pages[0].Records[0])
	if _, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, position, fixture.Acknowledgements[0]); err == nil || persisted {
		return PollingWatermarkConformanceObservation{}, errors.New("checkpoint persistence unexpectedly succeeded after acknowledgement")
	}
	checkpoint, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, position, fixture.Acknowledgements[1])
	if err != nil || !persisted {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("persist replay: %w", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.StableIdentities = uniquePollingWatermarkConformanceIdentities(fixture.Pages[0].Records)
	observation.SourceRequests = []PollingWatermarkConformancePosition{
		fixture.InitialCheckpoint.PollingWatermarkConformancePosition,
		fixture.InitialCheckpoint.PollingWatermarkConformancePosition,
	}
	observation.PersistedCheckpoints = []synccontract.CheckpointEnvelope{checkpoint}
	observation.ReplayedFromPriorCommitted = true
	return observation, nil
}

func runTombstoneHistoryHardDeleteVisibility(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if fixture.InitialCheckpoint == nil || fixture.Descriptor.DeleteVisibility != "soft_delete_only" || len(fixture.Pages) != 1 || len(fixture.Acknowledgements) != 1 {
		return PollingWatermarkConformanceObservation{}, errors.New("tombstone/history fixture is incomplete")
	}
	var emitted, tombstones, historyClosed []string
	var last PollingWatermarkConformancePosition
	for _, record := range fixture.Pages[0].Records {
		if record.HardDelete {
			continue
		}
		last = positionFromPollingWatermarkConformanceRecord(record)
		emitted = append(emitted, record.StableIdentity)
		if record.Tombstone {
			tombstones = append(tombstones, record.StableIdentity)
			historyClosed = append(historyClosed, record.StableIdentity)
		}
	}
	if len(emitted) == 0 {
		return PollingWatermarkConformanceObservation{}, errors.New("soft delete was not observable")
	}
	checkpoint, persisted, err := persistPollingWatermarkConformanceCheckpoint(fixture, last, fixture.Acknowledgements[0])
	if err != nil || !persisted {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("persist soft delete: %w", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.StableIdentities = emitted
	observation.SourceRequests = []PollingWatermarkConformancePosition{fixture.InitialCheckpoint.PollingWatermarkConformancePosition}
	observation.PersistedCheckpoints = []synccontract.CheckpointEnvelope{checkpoint}
	observation.TombstoneIdentities = tombstones
	observation.HistoryClosedIdentities = historyClosed
	observation.HardDeleteInvisible = true
	return observation, nil
}

func runMissingExecutorOrEvidenceAdmission(fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	if err := (PollingWatermarkConformanceRegistration{}).validate(); !errors.Is(err, ErrPollingWatermarkConformanceUnregistered) {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("unregistered validation error = %v", err)
	}
	if _, err := NewPollingWatermarkConformanceRegistration(fixture.Descriptor, PollingWatermarkConformanceEvidence{}); !errors.Is(err, ErrPollingWatermarkConformanceEvidence) {
		return PollingWatermarkConformanceObservation{}, fmt.Errorf("missing evidence validation error = %v", err)
	}
	observation := basePollingWatermarkConformanceObservation(fixture)
	observation.AdmissionRejected = true
	return observation, nil
}

func persistPollingWatermarkConformanceCheckpoint(fixture PollingWatermarkConformanceFixture, position PollingWatermarkConformancePosition, acknowledgement PollingWatermarkConformanceAcknowledgement) (synccontract.CheckpointEnvelope, bool, error) {
	candidate, err := pollingWatermarkConformanceEnvelope(fixture, position, false)
	if err != nil {
		return synccontract.CheckpointEnvelope{}, false, err
	}
	if !acknowledgement.Durable {
		return synccontract.CheckpointEnvelope{}, false, synccontract.ErrDownstreamAcknowledgementRequired
	}
	ack, err := synccontract.NewDurableDownstreamAcknowledgement("fixture-destination", candidate.ObservedAt.Add(time.Second))
	if err != nil {
		return synccontract.CheckpointEnvelope{}, false, err
	}
	var committed synccontract.CheckpointEnvelope
	err = synccontract.CommitAfterDownstreamAcknowledgement(candidate, ack, func(checkpoint synccontract.CheckpointEnvelope) error {
		if !acknowledgement.Persist {
			return errors.New("simulated checkpoint persistence failure")
		}
		committed = checkpoint
		return nil
	})
	if err != nil {
		return synccontract.CheckpointEnvelope{}, false, err
	}
	return committed, true, nil
}

func pollingWatermarkConformanceEnvelope(fixture PollingWatermarkConformanceFixture, position PollingWatermarkConformancePosition, committed bool) (synccontract.CheckpointEnvelope, error) {
	sourceGeneration := fixture.Descriptor.SourceGeneration
	schemaFingerprint := fixture.Descriptor.SchemaFingerprint
	if fixture.InitialCheckpoint != nil {
		if fixture.Scenario == "source_generation_mismatch" || fixture.Scenario == "schema_fingerprint_mismatch" {
			sourceGeneration = fixture.InitialCheckpoint.SourceGeneration
			schemaFingerprint = fixture.InitialCheckpoint.SchemaFingerprint
		}
	}
	observedAt := time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC)
	positionObserved := true
	envelope := synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source:       pollingWatermarkConformanceSourceIdentity(fixture.Descriptor),
		Mechanism:    pollingWatermarkConformanceMechanism,
		SnapshotBarrier: &synccontract.SnapshotBarrier{
			Kind:  "fixture_snapshot",
			Token: synccontract.OpaqueToken("fixture-barrier"),
		},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken(position.Watermark),
			TieBreaker: synccontract.OpaqueToken(position.TieBreaker),
		},
		PositionObserved: positionObservedPointer(positionObserved),
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken(sourceGeneration),
		SchemaVersion:    schemaFingerprint,
		ProtocolVersion:  "polling-conformance-v1",
		Dedupe: synccontract.DedupeIdentity{
			Kind:  "stable_identity",
			Value: synccontract.OpaqueToken("fixture-dedupe"),
		},
		DedupeWindow: synccontract.DedupeWindow{
			Kind:  "bounded_overlap",
			Start: synccontract.OpaqueToken("fixture-window-start"),
			End:   synccontract.OpaqueToken("fixture-window-end"),
		},
		ObservedAt: observedAt,
	}
	if committed {
		committedAt := observedAt.Add(time.Second)
		envelope.CommittedAt = &committedAt
	}
	if err := envelope.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return envelope, nil
}

func positionObservedPointer(value bool) *bool {
	return &value
}

func positionFromPollingWatermarkConformanceRecord(record PollingWatermarkConformanceRecord) PollingWatermarkConformancePosition {
	return PollingWatermarkConformancePosition{Watermark: record.Watermark, TieBreaker: record.TieBreaker}
}

func uniquePollingWatermarkConformanceIdentities(pages ...[]PollingWatermarkConformanceRecord) []string {
	seen := make(map[string]struct{})
	var identities []string
	for _, records := range pages {
		for _, record := range records {
			if record.HardDelete {
				continue
			}
			if _, exists := seen[record.StableIdentity]; exists {
				continue
			}
			seen[record.StableIdentity] = struct{}{}
			identities = append(identities, record.StableIdentity)
		}
	}
	return identities
}

func referencePollingWatermarkConformanceDescriptorAdmitted(descriptor PollingWatermarkConformanceDescriptor) bool {
	return descriptor.StableKeyset && descriptor.CursorPolicy == "lossless" && descriptor.BoundedOverlap && descriptor.BoundedCommitLag
}

type unregisteredPollingWatermarkConformanceFactory struct{}

func (unregisteredPollingWatermarkConformanceFactory) NewPollingWatermarkConformanceLane(context.Context) (PollingWatermarkConformanceLane, error) {
	return unregisteredPollingWatermarkConformanceLane{}, nil
}

type unregisteredPollingWatermarkConformanceLane struct{}

func (unregisteredPollingWatermarkConformanceLane) PollingWatermarkConformanceRegistration() PollingWatermarkConformanceRegistration {
	return PollingWatermarkConformanceRegistration{}
}

func (unregisteredPollingWatermarkConformanceLane) RunPollingWatermarkConformance(context.Context, PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	return PollingWatermarkConformanceObservation{}, errors.New("unregistered lane should not run")
}

type corruptingPollingWatermarkConformanceFactory struct {
	target                  string
	recoveryError           error
	dropCursorSampleResults bool
	mutateCheckpoint        func(*synccontract.CheckpointEnvelope)
}

func (f corruptingPollingWatermarkConformanceFactory) NewPollingWatermarkConformanceLane(ctx context.Context) (PollingWatermarkConformanceLane, error) {
	registration, err := newReferencePollingWatermarkConformanceRegistration()
	if err != nil {
		return nil, err
	}
	lane, err := referencePollingWatermarkConformanceFactory{registration: registration}.NewPollingWatermarkConformanceLane(ctx)
	if err != nil {
		return nil, err
	}
	return corruptingPollingWatermarkConformanceLane{
		PollingWatermarkConformanceLane: lane,
		target:                          f.target,
		recoveryError:                   f.recoveryError,
		dropCursorSampleResults:         f.dropCursorSampleResults,
		mutateCheckpoint:                f.mutateCheckpoint,
	}, nil
}

type corruptingPollingWatermarkConformanceLane struct {
	PollingWatermarkConformanceLane
	target                  string
	recoveryError           error
	dropCursorSampleResults bool
	mutateCheckpoint        func(*synccontract.CheckpointEnvelope)
}

func (l corruptingPollingWatermarkConformanceLane) RunPollingWatermarkConformance(ctx context.Context, fixture PollingWatermarkConformanceFixture) (PollingWatermarkConformanceObservation, error) {
	observation, err := l.PollingWatermarkConformanceLane.RunPollingWatermarkConformance(ctx, fixture)
	if err == nil && fixture.ID == l.target {
		observation.TombstoneIdentities = nil
		if l.recoveryError != nil {
			observation.RecoveryError = l.recoveryError
		}
		if l.dropCursorSampleResults {
			observation.CursorSampleResults = nil
		}
		if l.mutateCheckpoint != nil && len(observation.PersistedCheckpoints) > 0 {
			l.mutateCheckpoint(&observation.PersistedCheckpoints[0])
		}
	}
	return observation, err
}
