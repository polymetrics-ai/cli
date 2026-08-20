package synccontract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	// Output is connector-owned typed-action output captured after the target
	// reports durable success; callers cannot supply it through the transport API.
	Output  json.RawMessage `json:"output,omitempty"`
	durable bool
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

// WithOutput attaches validated destination evidence to a durable
// acknowledgement. The durable marker remains private, so this cannot
// manufacture a checkpoint authority from an untrusted result payload.
func (a DownstreamAcknowledgement) WithOutput(output json.RawMessage) (DownstreamAcknowledgement, error) {
	if err := a.validate(); err != nil {
		return DownstreamAcknowledgement{}, err
	}
	if len(output) == 0 || !json.Valid(output) {
		return DownstreamAcknowledgement{}, fmt.Errorf("durable acknowledgement output must be valid JSON")
	}
	a.Output = append(json.RawMessage(nil), output...)
	return a, nil
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
