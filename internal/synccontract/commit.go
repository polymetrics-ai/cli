package synccontract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ErrDownstreamAcknowledgementRequired prevents state advancement after a
// read or a merely attempted downstream write.
var ErrDownstreamAcknowledgementRequired = errors.New("durable downstream acknowledgement is required before checkpoint commit")

// ErrDurableETLDestinationRequired identifies an ETL admission failure before
// a generic destination can produce a checkpointed write.
var ErrDurableETLDestinationRequired = errors.New("checkpointed sync requires a durable destination acknowledgement")

// DownstreamAcknowledgement is supplied only after the destination has made
// the batch durable according to its own native protocol.
type DownstreamAcknowledgement struct {
	Sink           string    `json:"sink"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
	durable        bool
}

// DurableCheckpointCommitment proves that a checkpoint candidate was committed
// after a validated durable downstream acknowledgement.
type DurableCheckpointCommitment struct {
	checkpoint CheckpointEnvelope
	valid      bool
}

// DurableETLDestination supplies an acknowledgement only after its writes are
// durable enough for a checkpoint to advance.
type DurableETLDestination interface {
	AcknowledgeETLDurability(context.Context, string) (DownstreamAcknowledgement, error)
}

// DestinationDurabilityAdmissionError reports a destination that cannot
// truthfully acknowledge a checkpointed ETL write.
type DestinationDurabilityAdmissionError struct {
	Destination string
}

func (e *DestinationDurabilityAdmissionError) Error() string {
	guidance := "migrate this connection to a destination with durable checkpoint acknowledgement"
	if e == nil || strings.TrimSpace(e.Destination) == "" {
		return "checkpointed sync cannot start because the destination cannot report durable writes; " + guidance
	}
	return fmt.Sprintf("destination %q cannot run checkpointed sync because it cannot report durable writes; %s", e.Destination, guidance)
}

func (e *DestinationDurabilityAdmissionError) Unwrap() error {
	return ErrDurableETLDestinationRequired
}

// NewDurableDownstreamAcknowledgement constructs evidence that a destination
// has made a run durable. Callers cannot manufacture valid evidence with a
// struct literal because its durable marker is intentionally unexported.
func NewDurableDownstreamAcknowledgement(sink string, acknowledgedAt time.Time) (DownstreamAcknowledgement, error) {
	acknowledgement := DownstreamAcknowledgement{
		Sink:           sink,
		AcknowledgedAt: acknowledgedAt,
		durable:        true,
	}
	if err := acknowledgement.validate(); err != nil {
		return DownstreamAcknowledgement{}, err
	}
	return acknowledgement, nil
}

func (a DownstreamAcknowledgement) validate() error {
	if !a.durable {
		return ErrDownstreamAcknowledgementRequired
	}
	if strings.TrimSpace(a.Sink) == "" || a.AcknowledgedAt.IsZero() {
		return fmt.Errorf("%w: sink and acknowledgement timestamp are required", ErrDownstreamAcknowledgementRequired)
	}
	return nil
}

// CommitAfterDownstreamAcknowledgement stamps and passes a checkpoint to its
// committer only after an explicit durable downstream acknowledgement. The
// candidate itself remains unmodified, so a failed commit cannot make an
// in-memory source position appear durable.
func CommitAfterDownstreamAcknowledgement(candidate CheckpointEnvelope, acknowledgement DownstreamAcknowledgement, commit func(CheckpointEnvelope) error) error {
	if err := acknowledgement.validate(); err != nil {
		return err
	}
	if commit == nil {
		return fmt.Errorf("checkpoint committer is required")
	}
	if err := candidate.Validate(); err != nil {
		return err
	}
	if acknowledgement.AcknowledgedAt.Before(candidate.ObservedAt) {
		return fmt.Errorf("durable acknowledgement cannot precede checkpoint observation")
	}
	committed := candidate.Clone()
	acknowledgedAt := acknowledgement.AcknowledgedAt
	committed.CommittedAt = &acknowledgedAt
	return commit(committed)
}

// CommitDurableCheckpointAfterDownstreamAcknowledgement returns a commitment
// only after the checkpoint persistence callback has succeeded.
func CommitDurableCheckpointAfterDownstreamAcknowledgement(candidate CheckpointEnvelope, acknowledgement DownstreamAcknowledgement, commit func(CheckpointEnvelope) error) (DurableCheckpointCommitment, error) {
	var commitment DurableCheckpointCommitment
	if err := CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(committed CheckpointEnvelope) error {
		if err := commit(committed); err != nil {
			return err
		}
		commitment = DurableCheckpointCommitment{
			checkpoint: committed.Clone(),
			valid:      true,
		}
		return nil
	}); err != nil {
		return DurableCheckpointCommitment{}, err
	}
	return commitment, nil
}

// ValidateCandidate verifies that the commitment is bound to candidate.
func (c DurableCheckpointCommitment) ValidateCandidate(candidate CheckpointEnvelope) error {
	if !c.valid {
		return fmt.Errorf("%w: durable checkpoint commitment is required", ErrDownstreamAcknowledgementRequired)
	}
	if candidate.CommittedAt != nil {
		return fmt.Errorf("%w: checkpoint candidate must be uncommitted", ErrDownstreamAcknowledgementRequired)
	}
	if err := candidate.Validate(); err != nil {
		return fmt.Errorf("%w: checkpoint candidate is invalid: %v", ErrDownstreamAcknowledgementRequired, err)
	}
	if err := c.checkpoint.Validate(); err != nil {
		return fmt.Errorf("%w: committed checkpoint is invalid: %v", ErrDownstreamAcknowledgementRequired, err)
	}
	if c.checkpoint.CommittedAt == nil {
		return fmt.Errorf("%w: committed checkpoint timestamp is required", ErrDownstreamAcknowledgementRequired)
	}

	expected := candidate.Clone()
	actual := c.checkpoint.Clone()
	expected.CommittedAt = nil
	actual.CommittedAt = nil
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("%w: commitment does not match checkpoint candidate", ErrDownstreamAcknowledgementRequired)
	}
	return nil
}
