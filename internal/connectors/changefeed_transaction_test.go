package connectors

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/synccontract"
)

func TestDurableChangefeedTransactionOrdersDeliveryBeforeCheckpoint(t *testing.T) {
	steps := []string{}
	sink := &changefeedTransactionSink{steps: &steps}
	store := &changefeedTransactionStore{steps: &steps}
	transaction, err := NewDurableChangefeedTransaction(sink, store)
	if err != nil {
		t.Fatal(err)
	}
	candidate := changefeedTransactionCheckpoint(time.Now().UTC().Add(-time.Second))
	events := []CDCEvent{{Operation: "insert", Record: Record{"id": 7}}}
	if err := transaction.Commit(context.Background(), candidate, events, func(CDCEvent) error {
		steps = append(steps, "emit")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"write", "checkpoint", "emit"}; !reflect.DeepEqual(steps, want) {
		t.Fatalf("transaction steps = %v, want %v", steps, want)
	}
	if store.checkpoint.CommittedAt == nil {
		t.Fatal("checkpoint was persisted without a committed timestamp")
	}
	if candidate.CommittedAt != nil {
		t.Fatal("transaction mutated the checkpoint candidate")
	}
}

func TestDurableChangefeedTransactionDoesNotPersistFailedWrite(t *testing.T) {
	steps := []string{}
	sink := &changefeedTransactionSink{steps: &steps, writeErr: errors.New("downstream unavailable")}
	store := &changefeedTransactionStore{steps: &steps}
	transaction, err := NewDurableChangefeedTransaction(sink, store)
	if err != nil {
		t.Fatal(err)
	}
	err = transaction.Commit(context.Background(), changefeedTransactionCheckpoint(time.Now().UTC().Add(-time.Second)), nil, func(CDCEvent) error {
		steps = append(steps, "emit")
		return nil
	})
	if !errors.Is(err, sink.writeErr) {
		t.Fatalf("Commit() error = %v, want write failure", err)
	}
	if want := []string{"write"}; !reflect.DeepEqual(steps, want) {
		t.Fatalf("transaction steps = %v, want %v", steps, want)
	}
}

type changefeedTransactionSink struct {
	steps    *[]string
	writeErr error
}

func (s *changefeedTransactionSink) Name() string { return "warehouse" }

func (s *changefeedTransactionSink) CommitChangefeedTransaction(_ context.Context, _ synccontract.CheckpointEnvelope, _ []CDCEvent) error {
	*s.steps = append(*s.steps, "write")
	return s.writeErr
}

type changefeedTransactionStore struct {
	steps      *[]string
	checkpoint synccontract.CheckpointEnvelope
}

func (s *changefeedTransactionStore) PersistDurableChangefeedCheckpoint(_ context.Context, checkpoint synccontract.CheckpointEnvelope) error {
	*s.steps = append(*s.steps, "checkpoint")
	s.checkpoint = checkpoint.Clone()
	return nil
}

func changefeedTransactionCheckpoint(observedAt time.Time) synccontract.CheckpointEnvelope {
	return synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source: synccontract.SourceIdentity{
			Engine:           "postgres",
			AccountOrCluster: "cluster-a",
			ObjectScope:      "public.events",
		},
		Mechanism: "logical_replication",
		SnapshotBarrier: &synccontract.SnapshotBarrier{
			Kind:  "postgres_logical_slot",
			Token: synccontract.OpaqueToken("0/10"),
		},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("0/20"),
			TieBreaker: synccontract.OpaqueToken("0/18"),
		},
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken("1\npublication"),
		SchemaVersion:    "postgres-cdc-v1",
		ProtocolVersion:  "pgoutput-v1",
		Dedupe: synccontract.DedupeIdentity{
			Kind:  "postgres_transaction_end_lsn",
			Value: synccontract.OpaqueToken("0/20"),
		},
		DedupeWindow: synccontract.DedupeWindow{
			Kind:  "postgres_lsn_interval",
			Start: synccontract.OpaqueToken("0/10"),
			End:   synccontract.OpaqueToken("0/20"),
		},
		ObservedAt: observedAt,
	}
}
