package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// PollingApplyExecutor is the closed, registered native target boundary. It
// receives a preflight-resolved descriptor and a bounded source page; it has
// no SQL, relation, HTTP, shell, or caller-selected strategy input.
type PollingApplyExecutor interface {
	PollingPreflightApplyExecutor
	ApplyPollingPage(context.Context, ResolvedPollingWatermark, PollingApplyPage) (synccontract.DownstreamAcknowledgement, error)
}

// PollingApplyRecord keeps one source record paired with the source-owned
// ordering tuple that controls replay and late-event handling at the target.
// A target must never replace it with extraction-time ordering.
type PollingApplyRecord struct {
	Record   connectors.Record
	Position synccontract.CheckpointPosition
}

// PollingApplyPage is the sole data-bearing request accepted by a native
// polling target. Records and tombstones are distinct so absence cannot be
// interpreted as a target deletion.
type PollingApplyPage struct {
	Records    []PollingApplyRecord
	Tombstones []synccontract.Tombstone
}

// ApplyPollingPage validates the descriptor-owned record and byte bounds,
// clones the source payload, and dispatches only to the exact executor that
// was registered by successful preflight. Its acknowledgement is returned
// only from the target implementation after that target has proved durability.
func ApplyPollingPage(ctx context.Context, resolved ResolvedPollingWatermark, page PollingApplyPage) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("polling apply context is required")
	}
	if err := ctx.Err(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if resolved.Declaration == nil || resolved.Apply == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("resolved polling target apply is required")
	}
	if err := resolved.Declaration.Validate(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("resolved polling watermark declaration: %w", err)
	}
	if err := validatePollingApplyMode(resolved.Declaration.Target, resolved.Mode); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	apply, ok := resolved.Apply.(PollingApplyExecutor)
	if !ok || isNilPollingPreflightExecutor(apply) {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("resolved polling target executor cannot apply pages")
	}
	if apply.PollingApplyExecutorReference() != resolved.Declaration.Target.Executor {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("resolved polling target executor does not match the declaration")
	}
	clone, err := clonePollingApplyPage(page, resolved.Declaration.Target)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := ctx.Err(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	acknowledgement, err := apply.ApplyPollingPage(ctx, resolved, clone)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if acknowledgement.Sink == "" || acknowledgement.AcknowledgedAt.IsZero() {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("polling target apply returned an invalid durable acknowledgement")
	}
	return acknowledgement, nil
}

func clonePollingApplyPage(page PollingApplyPage, target connectors.PollingApplyDescriptor) (PollingApplyPage, error) {
	if len(page.Records) == 0 && len(page.Tombstones) == 0 {
		return PollingApplyPage{}, fmt.Errorf("polling apply page is empty")
	}
	if len(page.Records)+len(page.Tombstones) > target.MaxBatchRecords {
		return PollingApplyPage{}, fmt.Errorf("polling apply page exceeds declared record limit")
	}
	clone := PollingApplyPage{
		Records:    make([]PollingApplyRecord, len(page.Records)),
		Tombstones: make([]synccontract.Tombstone, len(page.Tombstones)),
	}
	bytes := 0
	for index, record := range page.Records {
		if len(record.Record) == 0 || len(record.Position.Primary) == 0 || len(record.Position.TieBreaker) == 0 {
			return PollingApplyPage{}, fmt.Errorf("polling apply record requires a source ordering tuple")
		}
		clonedRecord, err := clonePollingApplyRecord(record.Record)
		if err != nil {
			return PollingApplyPage{}, err
		}
		encoded, err := json.Marshal(clonedRecord)
		if err != nil {
			return PollingApplyPage{}, fmt.Errorf("polling apply record is not bounded serializable data")
		}
		bytes += len(encoded)
		if bytes > target.MaxBatchBytes {
			return PollingApplyPage{}, fmt.Errorf("polling apply page exceeds declared byte limit")
		}
		clone.Records[index] = PollingApplyRecord{Record: clonedRecord, Position: record.Position.Clone()}
	}
	for index, tombstone := range page.Tombstones {
		if err := tombstone.Validate(); err != nil {
			return PollingApplyPage{}, fmt.Errorf("polling apply tombstone: %w", err)
		}
		encoded, err := json.Marshal(tombstone)
		if err != nil {
			return PollingApplyPage{}, fmt.Errorf("polling apply tombstone is not bounded serializable data")
		}
		bytes += len(encoded)
		if bytes > target.MaxBatchBytes {
			return PollingApplyPage{}, fmt.Errorf("polling apply page exceeds declared byte limit")
		}
		clone.Tombstones[index] = tombstone.Clone()
	}
	return clone, nil
}

func clonePollingApplyRecord(record connectors.Record) (connectors.Record, error) {
	clone := make(connectors.Record, len(record))
	for key, value := range record {
		cloned, err := clonePollingApplyValue(value)
		if err != nil {
			return nil, err
		}
		clone[key] = cloned
	}
	return clone, nil
}

func clonePollingApplyValue(value any) (any, error) {
	switch typed := value.(type) {
	case nil, bool, string, int8, int16, int32, int64, uint8, uint16, uint32, uint64, time.Time, [16]byte:
		return typed, nil
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return nil, fmt.Errorf("polling apply record contains a non-finite float")
		}
		return typed, nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, fmt.Errorf("polling apply record contains a non-finite float")
		}
		return typed, nil
	case []byte:
		return append([]byte(nil), typed...), nil
	case json.RawMessage:
		if !json.Valid(typed) {
			return nil, fmt.Errorf("polling apply record contains invalid JSON")
		}
		return append(json.RawMessage(nil), typed...), nil
	case connectors.Record:
		return clonePollingApplyRecord(typed)
	case map[string]any:
		return clonePollingApplyRecord(connectors.Record(typed))
	case []any:
		clone := make([]any, len(typed))
		for index := range typed {
			cloned, err := clonePollingApplyValue(typed[index])
			if err != nil {
				return nil, err
			}
			clone[index] = cloned
		}
		return clone, nil
	default:
		if value == nil || (reflect.ValueOf(value).Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil()) {
			return nil, fmt.Errorf("polling apply record contains an unsupported nil value")
		}
		return nil, fmt.Errorf("polling apply record contains unsupported value type %T", value)
	}
}
