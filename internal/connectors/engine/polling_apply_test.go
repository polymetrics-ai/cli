package engine

import (
	"context"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestApplyPollingPageDispatchesOnlyARegisteredBoundedTargetAndReturnsDurableAcknowledgement(t *testing.T) {
	fixture := newPollingPreflightFixture(t)
	fixture.declaration.Target.MaxBatchBytes = 256
	apply := &recordingPollingPageApply{
		reference: fixture.declaration.Target.Executor,
		evidence:  RequiredPollingWatermarkConformanceEvidence(),
	}
	registry := NewPollingPreflightRegistry()
	if err := registry.RegisterSource(fixture.source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterApply(apply); err != nil {
		t.Fatal(err)
	}
	resolved, err := PollingPreflight(context.Background(), registry, fixture.declaration, fixture.object, synccontract.ModeIncrementalUpsert)
	if err != nil {
		t.Fatalf("PollingPreflight() error = %v", err)
	}

	page := PollingApplyPage{Records: []PollingApplyRecord{{
		Record: connectors.Record{"id": int64(7), "value": "newer"},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("2026-08-15T00:00:00.000000001Z"),
			TieBreaker: synccontract.OpaqueToken("7"),
		},
	}}}
	acknowledgement, err := ApplyPollingPage(context.Background(), resolved, page)
	if err != nil {
		t.Fatalf("ApplyPollingPage() error = %v", err)
	}
	if acknowledgement.Sink != "fixture-polling-target" || acknowledgement.AcknowledgedAt.IsZero() {
		t.Fatalf("ApplyPollingPage() acknowledgement = %#v, want durable target acknowledgement", acknowledgement)
	}
	if apply.calls != 1 {
		t.Fatalf("target apply calls = %d, want one", apply.calls)
	}
	if got := apply.page.Records[0].Record["value"]; got != "newer" {
		t.Fatalf("target record value = %#v, want sealed mapped source value", got)
	}

	apply.page.Records[0].Record["value"] = "target mutation"
	if got := page.Records[0].Record["value"]; got != "newer" {
		t.Fatalf("ApplyPollingPage leaked a mutable record map: got %#v, want original value", got)
	}
}

func TestApplyPollingPageRefusesUnsafeInputBeforeTargetMutation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*connectors.PollingWatermarkDescriptor, *recordingPollingPageApply, *PollingPreflightRegistry)
		page   PollingApplyPage
	}{
		{
			name: "unregistered apply cannot receive page",
			mutate: func(_ *connectors.PollingWatermarkDescriptor, _ *recordingPollingPageApply, registry *PollingPreflightRegistry) {
				registry.applies = map[connectors.TransportExecutorReference]PollingPreflightApplyExecutor{}
			},
		},
		{
			name: "record limit",
			mutate: func(declaration *connectors.PollingWatermarkDescriptor, _ *recordingPollingPageApply, _ *PollingPreflightRegistry) {
				declaration.Target.MaxBatchRecords = 1
			},
			page: pollingApplyTestPage(2, 8),
		},
		{
			name: "byte limit",
			mutate: func(declaration *connectors.PollingWatermarkDescriptor, _ *recordingPollingPageApply, _ *PollingPreflightRegistry) {
				declaration.Target.MaxBatchBytes = 1
			},
			page: pollingApplyTestPage(1, 32),
		},
		{
			name: "cancellation",
			page: pollingApplyTestPage(1, 8),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newPollingPreflightFixture(t)
			fixture.declaration.Target.MaxBatchBytes = 256
			apply := &recordingPollingPageApply{reference: fixture.declaration.Target.Executor, evidence: RequiredPollingWatermarkConformanceEvidence()}
			registry := NewPollingPreflightRegistry()
			if err := registry.RegisterSource(fixture.source); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterApply(apply); err != nil {
				t.Fatal(err)
			}
			if tt.mutate != nil {
				tt.mutate(fixture.declaration, apply, registry)
			}
			resolved, err := PollingPreflight(context.Background(), registry, fixture.declaration, fixture.object, synccontract.ModeIncrementalUpsert)
			if tt.name == "unregistered apply cannot receive page" {
				if err == nil {
					t.Fatal("PollingPreflight() succeeded with an unregistered target")
				}
				if apply.calls != 0 {
					t.Fatalf("target calls = %d, want zero after preflight refusal", apply.calls)
				}
				return
			}
			if err != nil {
				t.Fatalf("PollingPreflight() error = %v", err)
			}

			ctx := context.Background()
			if tt.name == "cancellation" {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			if _, err := ApplyPollingPage(ctx, resolved, tt.page); err == nil {
				t.Fatal("ApplyPollingPage() succeeded, want pre-mutation refusal")
			}
			if apply.calls != 0 {
				t.Fatalf("target calls = %d, want zero after unsafe page refusal", apply.calls)
			}
		})
	}
}

func pollingApplyTestPage(records, valueBytes int) PollingApplyPage {
	page := PollingApplyPage{Records: make([]PollingApplyRecord, records)}
	for index := range page.Records {
		page.Records[index] = PollingApplyRecord{
			Record: connectors.Record{"id": int64(index + 1), "value": string(make([]byte, valueBytes))},
			Position: synccontract.CheckpointPosition{
				Primary:    synccontract.OpaqueToken("2026-08-15T00:00:00.000000001Z"),
				TieBreaker: synccontract.OpaqueToken{byte(index + 1)},
			},
		}
	}
	return page
}

type recordingPollingPageApply struct {
	reference connectors.TransportExecutorReference
	evidence  PollingWatermarkConformanceEvidence
	calls     int
	page      PollingApplyPage
}

func (a *recordingPollingPageApply) PollingApplyExecutorReference() connectors.TransportExecutorReference {
	return a.reference
}

func (a *recordingPollingPageApply) PollingApplyConformanceEvidence() PollingWatermarkConformanceEvidence {
	return a.evidence
}

func (a *recordingPollingPageApply) ApplyPollingPage(_ context.Context, _ ResolvedPollingWatermark, page PollingApplyPage) (synccontract.DownstreamAcknowledgement, error) {
	a.calls++
	a.page = page
	return synccontract.NewDurableDownstreamAcknowledgement("fixture-polling-target", time.Now().UTC())
}
