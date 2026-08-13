package synctransport

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestOrchestratorReadsBackDestinationReceiptBeforeCheckpointCAS(t *testing.T) {
	for _, tt := range []struct {
		name        string
		readBackErr error
		wantOrder   []string
	}{
		{
			name:      "read-back precedes checkpoint CAS",
			wantOrder: []string{"stage", "reopen", "apply", "read-back", "checkpoint-cas"},
		},
		{
			name:        "read-back failure preserves checkpoint",
			readBackErr: errors.New("independent provider read-back failed"),
			wantOrder:   []string{"stage", "reopen", "apply", "read-back"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "api")
			order := []string{}
			stage := &receiptOrderingWarehouseStage{order: &order}
			destination := &receiptOrderingDestination{
				reference:   pair.destinationExecutor.reference,
				sink:        pair.destination.Name(),
				order:       &order,
				readBackErr: tt.readBackErr,
			}
			registry := NewRegistry(pair.verifier)
			if err := registry.RegisterSource(pair.sourceExecutor); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterDestination(destination); err != nil {
				t.Fatal(err)
			}

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      "records",
				Mode:        synccontract.ModeFullAppend,
				BatchSize:   1,
				Stage:       stage,
				Commit: func(synccontract.CheckpointEnvelope) error {
					order = append(order, "checkpoint-cas")
					return nil
				},
			})
			if tt.readBackErr == nil && err != nil {
				t.Fatalf("Run() = %v", err)
			}
			if tt.readBackErr != nil && !errors.Is(err, tt.readBackErr) {
				t.Fatalf("Run() error = %v, want read-back error %v", err, tt.readBackErr)
			}
			if !reflect.DeepEqual(order, tt.wantOrder) {
				t.Fatalf("transport order = %v, want %v", order, tt.wantOrder)
			}
		})
	}
}

type receiptOrderingWarehouseStage struct {
	order *[]string
	page  SourcePage
}

func (s *receiptOrderingWarehouseStage) Stage(_ context.Context, request WarehouseStageRequest) (WarehouseReceipt, error) {
	*s.order = append(*s.order, "stage")
	s.page = request.Page
	return WarehouseReceipt{
		ID:               "stage-receipt",
		Owner:            "connection-owner",
		Generation:       1,
		Stream:           request.Stream,
		Mode:             request.Mode,
		CheckpointSHA256: "checkpoint",
		TombstonesSHA256: "tombstones",
		ManifestSHA256:   "manifest",
		ContentSHA256:    "content",
		ParquetSHA256:    "parquet",
	}, nil
}

func (s *receiptOrderingWarehouseStage) Reopen(_ context.Context, receipt WarehouseReceipt) (WarehouseWorkset, error) {
	*s.order = append(*s.order, "reopen")
	return WarehouseWorkset{
		ID:                  receipt.ID,
		Records:             s.page.Records,
		Tombstones:          s.page.Tombstones,
		CandidateCheckpoint: s.page.CandidateCheckpoint,
	}, nil
}

type receiptOrderingDestination struct {
	reference   connectors.TransportExecutorReference
	sink        string
	order       *[]string
	readBackErr error
}

func (d *receiptOrderingDestination) TransportExecutorReference() connectors.TransportExecutorReference {
	return d.reference
}

func (d *receiptOrderingDestination) PlanDestination(_ context.Context, request DestinationPlanRequest) (DestinationPlan, error) {
	return DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}

func (d *receiptOrderingDestination) ApplyDestination(_ context.Context, _ DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	*d.order = append(*d.order, "apply")
	return synccontract.NewDurableDownstreamAcknowledgement(d.sink, time.Now().UTC())
}

func (d *receiptOrderingDestination) ReadBackDestination(_ context.Context, _ DestinationReadBackRequest) error {
	*d.order = append(*d.order, "read-back")
	return d.readBackErr
}
