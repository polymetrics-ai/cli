package synccontract

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrDownstreamAcknowledgementRequired prevents state advancement after a
// read or a merely attempted downstream write.
var ErrDownstreamAcknowledgementRequired = errors.New("durable downstream acknowledgement is required before checkpoint commit")

// DownstreamAcknowledgement is supplied only after the destination has made
// the batch durable according to its own native protocol.
type DownstreamAcknowledgement struct {
	Sink           string    `json:"sink"`
	Durable        bool      `json:"durable"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
}

func (a DownstreamAcknowledgement) validate() error {
	if !a.Durable {
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
