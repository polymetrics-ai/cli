package synccontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Operation is the stable operation type carried by a change envelope.
type Operation string

const (
	OperationInsert     Operation = "insert"
	OperationUpdate     Operation = "update"
	OperationDelete     Operation = "delete"
	OperationTruncate   Operation = "truncate"
	OperationInvalidate Operation = "invalidate"
)

// DeleteImage identifies the source image available for a row delete.
type DeleteImage string

const (
	DeleteImageKeyOnly     DeleteImage = "key_only"
	DeleteImageBefore      DeleteImage = "before_image"
	DeleteImageUnavailable DeleteImage = "unavailable"
)

// Tombstone is the delete/change envelope shared by all sync mechanisms.
// Record payloads stay JSON values so the contract neither assumes a SQL row
// nor rewrites source fields.
type Tombstone struct {
	Operation   Operation          `json:"operation"`
	EventID     OpaqueToken        `json:"event_id"`
	Key         json.RawMessage    `json:"key"`
	DeleteImage DeleteImage        `json:"delete_image"`
	Before      json.RawMessage    `json:"before,omitempty"`
	Position    CheckpointPosition `json:"position"`
}

// Clone returns a defensive copy of opaque and raw JSON values.
func (t Tombstone) Clone() Tombstone {
	clone := t
	clone.EventID = cloneToken(t.EventID)
	clone.Key = append(json.RawMessage(nil), t.Key...)
	clone.Before = append(json.RawMessage(nil), t.Before...)
	clone.Position = t.Position.Clone()
	return clone
}

// Validate requires enough information for deterministic, idempotent delete
// handling without imposing a provider-specific record schema.
func (t Tombstone) Validate() error {
	if len(t.EventID) == 0 {
		return fmt.Errorf("tombstone event identity is required")
	}
	if err := t.Position.validateOrdered("tombstone"); err != nil {
		return err
	}
	switch t.Operation {
	case OperationDelete:
		if !validTombstoneJSON(t.Key) {
			return fmt.Errorf("tombstone key must be valid JSON")
		}
		switch t.DeleteImage {
		case DeleteImageKeyOnly:
			if len(t.Before) != 0 {
				return fmt.Errorf("key-only tombstone cannot include a before image")
			}
		case DeleteImageBefore:
			if !validTombstoneJSON(t.Before) {
				return fmt.Errorf("before-image tombstone requires valid before JSON")
			}
		case DeleteImageUnavailable:
			if len(t.Before) != 0 {
				return fmt.Errorf("unavailable-image tombstone cannot include a before image")
			}
		default:
			return fmt.Errorf("unsupported tombstone delete image %q", t.DeleteImage)
		}
	case OperationTruncate, OperationInvalidate:
		if len(t.Key) != 0 || t.DeleteImage != "" || len(t.Before) != 0 {
			return fmt.Errorf("%s tombstone cannot carry row-delete fields", t.Operation)
		}
	default:
		return fmt.Errorf("unsupported tombstone operation %q", t.Operation)
	}
	return nil
}

func validTombstoneJSON(value json.RawMessage) bool {
	return len(value) != 0 && json.Valid(value) && !bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}

// HistoryDeleteAction is deliberately closed: a history target may close its
// validity window, but it may not physically remove the target row.
type HistoryDeleteAction string

const (
	HistoryDeleteCloseValidityWindow  HistoryDeleteAction = "close_validity_window"
	HistoryDeletePhysicalTargetDelete HistoryDeleteAction = "physical_target_delete"
)

const (
	// HistoryValidFromColumn is set when a history version becomes current.
	// Its matching end boundary is set when a tombstone closes that version.
	HistoryValidFromColumn = "_valid_from"
	// HistoryValidToColumn is the close-boundary column that a history target
	// must set when applying a tombstone.
	HistoryValidToColumn = "_valid_to"
	// HistoryIsCurrentColumn is cleared when a tombstone closes the current
	// validity window.
	HistoryIsCurrentColumn = "_is_current"
)

// Validate rejects physical deletes as non-conforming for history targets.
func (a HistoryDeleteAction) Validate() error {
	if a != HistoryDeleteCloseValidityWindow {
		return fmt.Errorf("history delete action %q is non-conforming; use %q", a, HistoryDeleteCloseValidityWindow)
	}
	return nil
}

// HistoryWindowClose describes the only permissible target mutation for a
// tombstone in history mode: set HistoryValidToColumn and
// HistoryIsCurrentColumn=false.
type HistoryWindowClose struct {
	Action    HistoryDeleteAction `json:"action"`
	EventID   OpaqueToken         `json:"event_id"`
	Key       json.RawMessage     `json:"key"`
	ValidTo   time.Time           `json:"valid_to"`
	IsCurrent bool                `json:"is_current"`
}

// CloseHistoryWindow converts a valid tombstone into its history-target
// mutation. It does not expose an option that performs a physical delete.
func CloseHistoryWindow(tombstone Tombstone, validTo time.Time) (HistoryWindowClose, error) {
	if tombstone.Operation != OperationDelete {
		return HistoryWindowClose{}, fmt.Errorf("history window close requires a row delete")
	}
	if err := tombstone.Validate(); err != nil {
		return HistoryWindowClose{}, err
	}
	if validTo.IsZero() {
		return HistoryWindowClose{}, fmt.Errorf("history validity close timestamp is required")
	}
	return HistoryWindowClose{
		Action:    HistoryDeleteCloseValidityWindow,
		EventID:   cloneToken(tombstone.EventID),
		Key:       append(json.RawMessage(nil), tombstone.Key...),
		ValidTo:   validTo,
		IsCurrent: false,
	}, nil
}
