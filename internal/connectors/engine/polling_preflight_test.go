package engine

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestPollingPreflightAdmitsDeclaredPollingBeforeGuardedSourceRead(t *testing.T) {
	fixture := newPollingPreflightFixture(t)
	registry := newPollingPreflightRegistry(t, fixture.source, fixture.apply)

	resolved, err := PollingPreflight(context.Background(), registry, fixture.declaration, fixture.object, synccontract.ModeIncrementalUpsert)
	if err != nil {
		t.Fatalf("PollingPreflight: %v", err)
	}
	if got, want := resolved.Mode, synccontract.ModeIncrementalUpsert; got != want {
		t.Fatalf("resolved mode = %q, want %q", got, want)
	}
	if err := fixture.syncAfterPreflight(resolved); err != nil {
		t.Fatalf("guarded sync: %v", err)
	}
	if got, want := fixture.source.reads, 1; got != want {
		t.Fatalf("source reads = %d, want %d after admitted preflight", got, want)
	}
	if got, want := fixture.apply.prepared, 1; got != want {
		t.Fatalf("target prepares = %d, want %d after admitted preflight", got, want)
	}
	if got, want := fixture.source.emitted, 1; got != want {
		t.Fatalf("source emitted records = %d, want %d after admitted preflight", got, want)
	}
}

func TestPollingPreflightRefusesEachUnsafeDeclarationBeforeSourceIO(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*pollingPreflightFixture)
		mode   synccontract.Mode
		want   string
	}{
		{
			name: "source executor absent",
			mutate: func(f *pollingPreflightFixture) {
				f.registry = newPollingPreflightRegistry(t, nil, f.apply)
			},
			want: `source polling executor "fixture-polling-source-v1" is not registered`,
		},
		{
			name: "apply executor absent",
			mutate: func(f *pollingPreflightFixture) {
				f.registry = newPollingPreflightRegistry(t, f.source, nil)
			},
			want: `target polling executor "fixture-polling-apply-v1" is not registered`,
		},
		{
			name: "missing immutable corpus evidence",
			mutate: func(f *pollingPreflightFixture) {
				f.source.evidence = PollingWatermarkConformanceEvidence{}
			},
			want: "source immutable polling conformance evidence is missing or stale",
		},
		{
			name: "registered source reference mismatch",
			mutate: func(f *pollingPreflightFixture) {
				f.source.reference.ID = "different-polling-source-v1"
			},
			want: "registered source polling executor does not match the declaration",
		},
		{
			name: "registered apply reference mismatch",
			mutate: func(f *pollingPreflightFixture) {
				f.apply.reference.ID = "different-polling-apply-v1"
			},
			want: "registered target polling executor does not match the declaration",
		},
		{
			name: "target immutable corpus evidence missing",
			mutate: func(f *pollingPreflightFixture) {
				f.apply.evidence = PollingWatermarkConformanceEvidence{}
			},
			want: "target immutable polling conformance evidence is missing or stale",
		},
		{
			name: "non lossless cursor codec",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Cursor.Codec = connectors.PollingCursorCodecFloat64
			},
			want: "polling watermark declaration: polling watermark cursor codec must preserve values losslessly",
		},
		{
			name: "no unique tie breaker",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Ordering.TieBreaker.Unique = false
			},
			want: "polling watermark declaration: polling ordering tie_breaker must be unique",
		},
		{
			name: "page checkpoint lacks stable traversal",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Read.StableTraversal = false
			},
			want: "polling watermark declaration: page checkpoints require stable keyset traversal",
		},
		{
			name: "unsafe mutation commit overlap",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Mutation.BoundedOverlap = false
			},
			want: "polling watermark declaration: mutable source requires bounded overlap and commit ordering",
		},
		{
			name: "hard deletes advertised as tombstones",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.DeleteVisibility = connectors.PollingDeleteVisibilityTombstone
			},
			want: "polling watermark declaration: polling watermark cannot advertise tombstones without a cursor-advancing soft delete",
		},
		{
			name: "target strategy is incompatible with mode",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Modes = []synccontract.Mode{synccontract.ModeIncrementalUpsert}
				f.declaration.Target.Strategies = []connectors.PollingApplyStrategy{connectors.PollingApplyStrategyAppend}
			},
			want: `target polling apply does not support sync mode "incremental_upsert"`,
		},
		{
			name: "history lacks transactional retry safe close and insert",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Target.Transaction = connectors.PollingTransactionNone
				f.declaration.Target.RetrySafeCloseAndInsert = false
			},
			want: "polling watermark declaration: history mode requires transaction and retry-safe close-and-insert",
		},
		{
			name: "canonical mode absent from source declaration",
			mode: synccontract.ModeFullAppend,
			want: `polling source does not support sync mode "full_append"`,
		},
		{
			name: "change capture mode is forbidden",
			mutate: func(f *pollingPreflightFixture) {
				f.declaration.Source.Modes = []synccontract.Mode{synccontract.ModeChangeCapture}
			},
			want: "polling watermark declaration: polling watermark cannot declare change_capture mode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newPollingPreflightFixture(t)
			if testCase.mutate != nil {
				testCase.mutate(fixture)
			}
			if fixture.registry == nil {
				fixture.registry = newPollingPreflightRegistry(t, fixture.source, fixture.apply)
			}

			mode := testCase.mode
			if mode == "" {
				mode = synccontract.ModeIncrementalUpsert
			}
			_, err := PollingPreflight(context.Background(), fixture.registry, fixture.declaration, fixture.object, mode)
			if err == nil || err.Error() != testCase.want {
				t.Fatalf("PollingPreflight error = %v, want %q", err, testCase.want)
			}
			if got := fixture.source.reads; got != 0 {
				t.Fatalf("source reads = %d, want 0 after pre-I/O refusal", got)
			}
			if got := fixture.apply.prepared; got != 0 {
				t.Fatalf("target prepares = %d, want 0 after pre-I/O refusal", got)
			}
		})
	}
}

func TestPollingPreflightCoversCorpusCursorDomainAndEmptyPage(t *testing.T) {
	tests := []struct {
		name   string
		cursor connectors.PollingCursor
		want   string
	}{
		{
			name:   "null watermark",
			cursor: connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, AllowsNull: true},
			want:   "polling watermark declaration: polling watermark cursor cannot allow null",
		},
		{
			name:   "nanosecond timestamp boundary",
			cursor: connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"},
		},
		{
			name:   "large numeric tie breaker",
			cursor: connectors.PollingCursor{Codec: connectors.PollingCursorCodecDecimal, Type: connectors.PollingCursorTypeInteger, Precision: "exact"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newPollingPreflightFixture(t)
			fixture.declaration.Source.Cursor = testCase.cursor
			resolved, err := PollingPreflight(context.Background(), fixture.registry, fixture.declaration, fixture.object, synccontract.ModeIncrementalUpsert)
			if testCase.want != "" {
				if err == nil || err.Error() != testCase.want {
					t.Fatalf("PollingPreflight error = %v, want %q", err, testCase.want)
				}
				if got := fixture.source.reads; got != 0 {
					t.Fatalf("source reads = %d, want 0 for rejected cursor", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("PollingPreflight: %v", err)
			}
			if got, want := resolved.Declaration.Source.Cursor, testCase.cursor; got != want {
				t.Fatalf("resolved cursor = %+v, want exact corpus-domain value %+v", got, want)
			}
			fixture.emptyResult = true
			if err := fixture.syncAfterPreflight(resolved); err != nil {
				t.Fatalf("guarded empty-page sync: %v", err)
			}
			if got, want := fixture.source.emptyPages, 1; got != want {
				t.Fatalf("empty pages = %d, want %d; empty source result must be observable", got, want)
			}
			if got := fixture.source.emitted; got != 0 {
				t.Fatalf("empty source result emitted %d records, want 0", got)
			}
		})
	}
}

func TestPollingModeEligibilitySweepsEveryImplementedPollingModeThroughRuntimePreflight(t *testing.T) {
	fixture := newPollingPreflightFixture(t)
	fixture.declaration.Source.Modes = []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
	}
	fixture.declaration.Target.Strategies = []connectors.PollingApplyStrategy{
		connectors.PollingApplyStrategyReplace,
		connectors.PollingApplyStrategyAppend,
		connectors.PollingApplyStrategyMerge,
		connectors.PollingApplyStrategyDedupe,
		connectors.PollingApplyStrategyDedupeHistory,
	}

	eligibility := PollingModeEligibilityOf(context.Background(), fixture.registry, fixture.declaration, fixture.object)
	if got, want := len(eligibility), len(fixture.declaration.Source.Modes); got != want {
		t.Fatalf("eligibility rows = %d, want one runtime-preflight row per declared mode (%d)", got, want)
	}
	for _, row := range eligibility {
		if row.Status != "implemented" || row.Reason != "" {
			t.Fatalf("eligibility for %q = %+v, want runtime-admitted implementation", row.Mode, row)
		}
	}
	if fixture.source.reads != 0 || fixture.apply.prepared != 0 {
		t.Fatalf("eligibility sweep triggered source I/O: reads=%d prepares=%d", fixture.source.reads, fixture.apply.prepared)
	}

	fixture.source.evidence = PollingWatermarkConformanceEvidence{}
	eligibility = PollingModeEligibilityOf(context.Background(), fixture.registry, fixture.declaration, fixture.object)
	for _, row := range eligibility {
		if row.Status != "blocked" || row.Reason != "source immutable polling conformance evidence is missing or stale" {
			t.Fatalf("eligibility for stale %q = %+v, want exact runtime-preflight refusal", row.Mode, row)
		}
	}
	if fixture.source.reads != 0 || fixture.apply.prepared != 0 {
		t.Fatalf("refused eligibility sweep triggered source I/O: reads=%d prepares=%d", fixture.source.reads, fixture.apply.prepared)
	}
}

type pollingPreflightFixture struct {
	declaration *connectors.PollingWatermarkDescriptor
	object      connectors.PollingCatalogObject
	source      *pollingPreflightSource
	apply       *pollingPreflightApply
	registry    *PollingPreflightRegistry
	emptyResult bool
}

func newPollingPreflightFixture(t *testing.T) *pollingPreflightFixture {
	t.Helper()
	declaration := &connectors.PollingWatermarkDescriptor{
		Status: connectors.PollingWatermarkStatusImplemented,
		Source: connectors.PollingWatermarkSourceDescriptor{
			Executor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "fixture-polling-source-v1"},
			Object:   connectors.PollingCatalogObjectSelector{Kind: connectors.PollingCatalogObjectRelation},
			Read: connectors.PollingReadProtocol{
				Kind:            connectors.PollingReadProtocolKeyset,
				MaxPageSize:     100,
				MaxPages:        2,
				MaxRequests:     2,
				StableTraversal: true,
				Predicate:       connectors.PollingKeysetPredicateLexicographic,
			},
			Snapshot: connectors.PollingSnapshotBarrier{Kind: connectors.PollingSnapshotBarrierTransaction},
			Cursor:   connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"},
			Ordering: connectors.PollingOrderingTuple{
				Watermark:  connectors.PollingOrderingField{CatalogField: "updated_at", Ascending: true},
				TieBreaker: connectors.PollingOrderingField{CatalogField: "id", Ascending: true, Unique: true},
			},
			Mutation:         connectors.PollingMutationPolicy{Mutable: true, CommitOrdered: true, BoundedOverlap: true},
			Identity:         connectors.PollingSourceIdentity{Engine: "fixture-native", AccountScope: "fixture-account", ObjectScope: "widgets"},
			Schema:           connectors.PollingSchemaCompatibilityExactFingerprint,
			DeleteVisibility: connectors.PollingDeleteVisibilityHardDeleteInvisible,
			Modes:            []synccontract.Mode{synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupeHistory},
		},
		Target: connectors.PollingApplyDescriptor{
			Executor:                connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "fixture-polling-apply-v1"},
			MaxBatchRecords:         100,
			MaxBatchBytes:           1 << 20,
			Staging:                 connectors.PollingStagingReplaceSupported,
			StableKeyMapping:        []string{"id"},
			ConditionalOrderFence:   true,
			Transaction:             connectors.PollingTransactionRequired,
			PartialResult:           connectors.PollingPartialResultRollback,
			RetrySafeCloseAndInsert: true,
			ValidityWindow:          connectors.PollingValidityWindowSupported,
			Strategies:              []connectors.PollingApplyStrategy{connectors.PollingApplyStrategyMerge, connectors.PollingApplyStrategyDedupeHistory},
		},
	}
	source := &pollingPreflightSource{reference: declaration.Source.Executor, evidence: RequiredPollingWatermarkConformanceEvidence()}
	apply := &pollingPreflightApply{reference: declaration.Target.Executor, evidence: RequiredPollingWatermarkConformanceEvidence()}
	fixture := &pollingPreflightFixture{
		declaration: declaration,
		object: connectors.PollingCatalogObject{
			Kind:      connectors.PollingCatalogObjectRelation,
			NameParts: []string{"public", "widgets"},
			Columns:   []string{"id", "updated_at"},
		},
		source: source,
		apply:  apply,
	}
	fixture.registry = newPollingPreflightRegistry(t, source, apply)
	return fixture
}

func newPollingPreflightRegistry(t *testing.T, source *pollingPreflightSource, apply *pollingPreflightApply) *PollingPreflightRegistry {
	t.Helper()
	registry := NewPollingPreflightRegistry()
	if source != nil {
		if err := registry.RegisterSource(source); err != nil {
			t.Fatalf("RegisterSource: %v", err)
		}
	}
	if apply != nil {
		if err := registry.RegisterApply(apply); err != nil {
			t.Fatalf("RegisterApply: %v", err)
		}
	}
	return registry
}

func (f *pollingPreflightFixture) syncAfterPreflight(resolved ResolvedPollingWatermark) error {
	if resolved.Source != f.source || resolved.Apply != f.apply {
		return context.Canceled
	}
	f.source.reads++
	if f.emptyResult {
		f.source.emptyPages++
	} else {
		f.source.emitted++
	}
	f.apply.prepared++
	return nil
}

type pollingPreflightSource struct {
	reference  connectors.TransportExecutorReference
	evidence   PollingWatermarkConformanceEvidence
	reads      int
	emptyPages int
	emitted    int
}

func (s *pollingPreflightSource) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return s.reference
}

func (s *pollingPreflightSource) PollingSourceConformanceEvidence() PollingWatermarkConformanceEvidence {
	return s.evidence
}

type pollingPreflightApply struct {
	reference connectors.TransportExecutorReference
	evidence  PollingWatermarkConformanceEvidence
	prepared  int
}

func (a *pollingPreflightApply) PollingApplyExecutorReference() connectors.TransportExecutorReference {
	return a.reference
}

func (a *pollingPreflightApply) PollingApplyConformanceEvidence() PollingWatermarkConformanceEvidence {
	return a.evidence
}
