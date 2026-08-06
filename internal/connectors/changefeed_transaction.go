package connectors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/synccontract"
)

var ErrDurableChangefeedTransactionRequired = errors.New("native CDC requires a durable changefeed transaction")

// DurableChangefeedSink commits one source transaction durably downstream.
type DurableChangefeedSink interface {
	Name() string
	CommitChangefeedTransaction(context.Context, synccontract.CheckpointEnvelope, []CDCEvent) error
}

// DurableChangefeedCheckpointStore persists a committed source checkpoint.
type DurableChangefeedCheckpointStore interface {
	PersistDurableChangefeedCheckpoint(context.Context, synccontract.CheckpointEnvelope) error
}

// DurableChangefeedTransaction coordinates downstream delivery and checkpoint persistence.
type DurableChangefeedTransaction struct {
	sink            DurableChangefeedSink
	checkpointStore DurableChangefeedCheckpointStore
}

// NewDurableChangefeedTransaction builds the boundary used by native CDC executors.
func NewDurableChangefeedTransaction(sink DurableChangefeedSink, checkpointStore DurableChangefeedCheckpointStore) (*DurableChangefeedTransaction, error) {
	if sink == nil {
		return nil, ErrDurableChangefeedTransactionRequired
	}
	if strings.TrimSpace(sink.Name()) == "" {
		return nil, fmt.Errorf("%w: downstream sink name is required", ErrDurableChangefeedTransactionRequired)
	}
	if checkpointStore == nil {
		return nil, fmt.Errorf("%w: checkpoint store is required", ErrDurableChangefeedTransactionRequired)
	}
	return &DurableChangefeedTransaction{sink: sink, checkpointStore: checkpointStore}, nil
}

// Commit writes a source transaction, persists its checkpoint, then publishes its events.
func (t *DurableChangefeedTransaction) Commit(ctx context.Context, candidate synccontract.CheckpointEnvelope, events []CDCEvent, emit func(CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if t == nil || t.sink == nil || t.checkpointStore == nil {
		return ErrDurableChangefeedTransactionRequired
	}
	if emit == nil {
		return errors.New("CDC transaction requires an event observer")
	}
	if candidate.CommittedAt != nil {
		return errors.New("CDC transaction requires an uncommitted checkpoint candidate")
	}
	if err := candidate.Validate(); err != nil {
		return fmt.Errorf("validate CDC checkpoint candidate: %w", err)
	}
	if err := t.sink.CommitChangefeedTransaction(ctx, candidate.Clone(), events); err != nil {
		return fmt.Errorf("write CDC transaction: %w", err)
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(t.sink.Name(), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("record CDC transaction durability: %w", err)
	}
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(committed synccontract.CheckpointEnvelope) error {
		return t.checkpointStore.PersistDurableChangefeedCheckpoint(ctx, committed)
	}); err != nil {
		return fmt.Errorf("persist durable CDC checkpoint: %w", err)
	}
	for _, event := range events {
		if err := emit(event); err != nil {
			return err
		}
	}
	return nil
}
