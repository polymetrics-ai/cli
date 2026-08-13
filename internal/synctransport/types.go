// Package synctransport owns the consumer-facing, closed source/destination
// dispatch seam. It deliberately contains no provider protocol, database
// driver, generic HTTP/SQL/shell transport, or second sync-mode vocabulary.
package synctransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// ConformanceVerification asks an authority outside a descriptor or executor
// whether an evidence reference was accepted. The source and destination
// executors cannot self-admit by returning fixture IDs or a digest.
type ConformanceVerification struct {
	Role          connectors.TransportRole
	ConnectorName string
	Executor      connectors.TransportExecutorReference
	Evidence      connectors.ConformanceEvidenceReference
}

// ConformanceVerifier is intentionally small and read-only. A future #3810
// evidence runner can implement it without moving checkpoint semantics into
// this package. Until then the default verifier keeps real transport admission
// closed.
type ConformanceVerifier interface {
	VerifyTransportConformance(ConformanceVerification) error
}

type unavailableConformanceVerifier struct{}

func (unavailableConformanceVerifier) VerifyTransportConformance(ConformanceVerification) error {
	return fmt.Errorf("external transport conformance verification is unavailable")
}

// SourceExecutor is the narrow source role. It emits #3810-owned page
// candidates; the orchestrator neither parses their opaque positions nor
// invents tombstone/history/recovery semantics.
type SourceExecutor interface {
	TransportExecutorReference() connectors.TransportExecutorReference
	ReadTransport(context.Context, SourceRequest, func(SourcePage) error) error
}

// DestinationExecutor is the narrow destination role. It plans only a
// descriptor-resolved closed strategy and returns #3810's opaque durable
// acknowledgement after its warehouse-mediated workset is durable.
type DestinationExecutor interface {
	TransportExecutorReference() connectors.TransportExecutorReference
	PlanDestination(context.Context, DestinationPlanRequest) (DestinationPlan, error)
	ApplyDestination(context.Context, DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error)
}

// SourceRequest is the fixed source invocation context. It has no generic
// request URL, SQL text, command, action, or caller-authored payload.
type SourceRequest struct {
	Connector  connectors.Connector
	Runtime    connectors.RuntimeConfig
	Stream     string
	Mode       synccontract.Mode
	BatchSize  int
	Resume     synccontract.ResumeExpectation
	Checkpoint *synccontract.CheckpointEnvelope
}

// SourcePage carries a bounded provider payload separately from #3810's
// tombstones and checkpoint candidate. The orchestrator never annotates the
// provider record map with `_polymetrics_*` fields.
type SourcePage struct {
	Records             []connectors.Record
	Tombstones          []synccontract.Tombstone
	CandidateCheckpoint synccontract.CheckpointEnvelope
}

// WarehouseStage is the required mediator. The source never invokes a
// destination executor directly: it places a bounded page into this typed
// stage, and the destination consumes the workset returned from that stage.
// A real durable implementation belongs to the corresponding warehouse/apply
// foundation; #3864 exercises this seam only with fakes.
type WarehouseStage interface {
	Stage(context.Context, WarehouseStageRequest) (WarehouseWorkset, error)
}

type WarehouseStageRequest struct {
	SourceName      string
	DestinationName string
	Stream          string
	Mode            synccontract.Mode
	Page            SourcePage
}

// WarehouseWorkset is stage-owned output. CandidateCheckpoint is included for
// audit visibility, but the orchestrator commits the original source candidate
// so a stage cannot silently substitute a source position.
type WarehouseWorkset struct {
	ID                  string
	Records             []connectors.Record
	Tombstones          []synccontract.Tombstone
	CandidateCheckpoint synccontract.CheckpointEnvelope
}

type DestinationPlanRequest struct {
	Connector     connectors.Connector
	Runtime       connectors.RuntimeConfig
	Stream        string
	Mode          synccontract.Mode
	ApplyStrategy connectors.DestinationApplyStrategy
}

type DestinationPlan struct {
	ApplyStrategy connectors.DestinationApplyStrategy
}

type DestinationApplyRequest struct {
	Plan    DestinationPlan
	Workset WarehouseWorkset
}

// PreflightRequest contains only identities and closed declarations needed to
// prove dispatch. It has no source payload and performs no provider I/O.
type PreflightRequest struct {
	Source      connectors.Connector
	Destination connectors.Connector
	Stream      string
	Mode        synccontract.Mode
}

// RunRequest wires a resolved connection into one shared orchestrator.
type RunRequest struct {
	Source             connectors.Connector
	SourceRuntime      connectors.RuntimeConfig
	Destination        connectors.Connector
	DestinationRuntime connectors.RuntimeConfig
	Stream             string
	Mode               synccontract.Mode
	BatchSize          int
	Resume             synccontract.ResumeExpectation
	Checkpoint         *synccontract.CheckpointEnvelope
	Stage              WarehouseStage
	Commit             func(synccontract.CheckpointEnvelope) error
}

type Result struct {
	RecordsRead         int
	RecordsStaged       int
	RecordsApplied      int
	Pages               int
	CommittedCheckpoint *synccontract.CheckpointEnvelope
}

func (r RunRequest) preflightRequest() PreflightRequest {
	return PreflightRequest{Source: r.Source, Destination: r.Destination, Stream: r.Stream, Mode: r.Mode}
}

func (r RunRequest) validateExecution() error {
	if r.BatchSize <= 0 {
		return fmt.Errorf("transport batch size must be positive")
	}
	return nil
}

// validateDispatchDependencies runs after closed registry preflight so an
// absent stage cannot hide a missing executor, invalid mode, or unsafe
// acknowledgement declaration. None of these checks can cause source I/O.
func (r RunRequest) validateDispatchDependencies() error {
	if isNilInterface(r.Stage) {
		return fmt.Errorf("warehouse stage is required for transport dispatch")
	}
	if r.Commit == nil {
		return fmt.Errorf("checkpoint committer is required for transport dispatch")
	}
	return nil
}

func cloneSourceRequest(request SourceRequest) SourceRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.Resume.SourceGeneration = append(synccontract.OpaqueToken(nil), request.Resume.SourceGeneration...)
	if request.Checkpoint != nil {
		checkpoint := request.Checkpoint.Clone()
		clone.Checkpoint = &checkpoint
	}
	return clone
}

func cloneDestinationPlanRequest(request DestinationPlanRequest) DestinationPlanRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	return clone
}

func cloneSourcePage(page SourcePage) (SourcePage, error) {
	clone := page
	records, err := cloneRecords(page.Records)
	if err != nil {
		return SourcePage{}, fmt.Errorf("clone page records: %w", err)
	}
	clone.Records = records
	clone.Tombstones = cloneTombstones(page.Tombstones)
	clone.CandidateCheckpoint = page.CandidateCheckpoint.Clone()
	return clone, nil
}

func cloneWarehouseWorkset(workset WarehouseWorkset) (WarehouseWorkset, error) {
	clone := workset
	records, err := cloneRecords(workset.Records)
	if err != nil {
		return WarehouseWorkset{}, fmt.Errorf("clone workset records: %w", err)
	}
	clone.Records = records
	clone.Tombstones = cloneTombstones(workset.Tombstones)
	clone.CandidateCheckpoint = workset.CandidateCheckpoint.Clone()
	return clone, nil
}

func cloneRuntimeConfig(runtime connectors.RuntimeConfig) connectors.RuntimeConfig {
	clone := runtime
	clone.Config = cloneStringMap(runtime.Config)
	clone.Secrets = cloneStringMap(runtime.Secrets)
	clone.ApprovedPayloadSHA256 = cloneStringMap(runtime.ApprovedPayloadSHA256)
	if runtime.ResolvedCatalog != nil {
		catalog := *runtime.ResolvedCatalog
		catalog.Streams = append([]connectors.Stream(nil), runtime.ResolvedCatalog.Streams...)
		clone.ResolvedCatalog = &catalog
	}
	return clone
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func cloneRecords(records []connectors.Record) ([]connectors.Record, error) {
	if records == nil {
		return nil, nil
	}
	clone := make([]connectors.Record, len(records))
	for index, record := range records {
		clonedRecord, err := cloneRecord(record)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", index, err)
		}
		clone[index] = clonedRecord
	}
	return clone, nil
}

var errUnsupportedTransportRecordValue = errors.New("unsupported transport record value")

// cloneRecord copies the closed JSON-like record vocabulary accepted by
// transport. A stage or destination therefore cannot mutate a nested provider
// field through the workset it receives, and an unrecognized mutable value
// cannot silently cross either boundary by alias.
func cloneRecord(record connectors.Record) (connectors.Record, error) {
	clone := make(connectors.Record, len(record))
	for key, value := range record {
		clonedValue, err := cloneRecordValue(value)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", key, err)
		}
		clone[key] = clonedValue
	}
	return clone, nil
}

func cloneRecordValue(value any) (any, error) {
	switch typed := value.(type) {
	case connectors.Record:
		return cloneRecord(typed)
	case map[string]any:
		clone := make(map[string]any, len(typed))
		for key, nested := range typed {
			clonedValue, err := cloneRecordValue(nested)
			if err != nil {
				return nil, fmt.Errorf("map field %q: %w", key, err)
			}
			clone[key] = clonedValue
		}
		return clone, nil
	case map[string]string:
		return cloneStringMap(typed), nil
	case []any:
		clone := make([]any, len(typed))
		for index, nested := range typed {
			clonedValue, err := cloneRecordValue(nested)
			if err != nil {
				return nil, fmt.Errorf("list item %d: %w", index, err)
			}
			clone[index] = clonedValue
		}
		return clone, nil
	case []connectors.Record:
		return cloneRecords(typed)
	case json.RawMessage:
		return append(json.RawMessage(nil), typed...), nil
	case []byte:
		return append([]byte(nil), typed...), nil
	case nil, bool, string, json.Number,
		float32, float64,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return value, nil
	default:
		return nil, fmt.Errorf("%w: %T", errUnsupportedTransportRecordValue, value)
	}
}

func cloneTombstones(tombstones []synccontract.Tombstone) []synccontract.Tombstone {
	if tombstones == nil {
		return nil
	}
	clone := make([]synccontract.Tombstone, len(tombstones))
	for index, tombstone := range tombstones {
		clone[index] = tombstone.Clone()
	}
	return clone
}

func validateTransportName(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("transport %s is required", label)
	}
	return nil
}
