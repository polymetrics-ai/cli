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
	"time"

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

// EmptyResultSource explicitly admits a successful, zero-page read without a
// fabricated checkpoint. The orchestrator keeps rejecting silent zero-page
// executors unless the exact registered source implements this marker.
type EmptyResultSource interface {
	SourceExecutor
	AllowEmptySourceResult()
}

// DestinationExecutor is the narrow destination role. It plans only a
// descriptor-resolved closed strategy and returns #3810's opaque durable
// acknowledgement after its warehouse-mediated workset is durable.
type DestinationExecutor interface {
	TransportExecutorReference() connectors.TransportExecutorReference
	PlanDestination(context.Context, DestinationPlanRequest) (DestinationPlan, error)
	ApplyDestination(context.Context, DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error)
	ReadBackDestination(context.Context, DestinationReadBackRequest) error
}

type TransportExecutionOrigin string

const (
	TransportExecutionOriginSource      TransportExecutionOrigin = "source"
	TransportExecutionOriginDestination TransportExecutionOrigin = "destination"
	TransportExecutionOriginInternal    TransportExecutionOrigin = "internal"
)

type transportExecutionOriginError struct {
	origin TransportExecutionOrigin
	err    error
}

func (e *transportExecutionOriginError) Error() string {
	if e == nil || e.err == nil {
		return "transport execution failed"
	}
	return e.err.Error()
}

func (e *transportExecutionOriginError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func TransportExecutionOriginOf(err error) (TransportExecutionOrigin, bool) {
	var tagged *transportExecutionOriginError
	if !errors.As(err, &tagged) || tagged == nil {
		return "", false
	}
	return tagged.origin, true
}

func tagTransportExecutionError(origin TransportExecutionOrigin, err error) error {
	if err == nil {
		return nil
	}
	if _, tagged := TransportExecutionOriginOf(err); tagged {
		return err
	}
	return &transportExecutionOriginError{origin: origin, err: err}
}

type DestinationApplyOutputError struct {
	err    error
	output json.RawMessage
}

func NewDestinationApplyOutputError(err error, output json.RawMessage) error {
	if err == nil {
		return nil
	}
	return &DestinationApplyOutputError{err: err, output: append(json.RawMessage(nil), output...)}
}

func (e *DestinationApplyOutputError) Error() string {
	if e == nil || e.err == nil {
		return "destination apply failed"
	}
	return e.err.Error()
}

func (e *DestinationApplyOutputError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func DestinationApplyOutput(err error) (json.RawMessage, bool) {
	var outputErr *DestinationApplyOutputError
	if !errors.As(err, &outputErr) || outputErr == nil || len(outputErr.output) == 0 {
		return nil, false
	}
	return append(json.RawMessage(nil), outputErr.output...), true
}

// FullOverwriteDestination is the optional run-scoped destination protocol for
// the canonical replace mode. It keeps the whole replacement lifecycle behind
// a destination-neutral port: the orchestrator stages and admits each bounded
// source unit, while the destination owns its private shadow, receipt, and
// publish implementation. No database handle, SQL string, or source payload
// crosses this boundary.
//
// A destination that does not implement this port cannot receive a
// full_overwrite transport run. Applying the generic per-page destination
// method for that mode would make replacement semantics depend on page size.
type FullOverwriteDestination interface {
	DestinationExecutor
	BeginFullOverwrite(context.Context, FullOverwriteRunRequest) (FullOverwriteRun, error)
}

// FullOverwriteRunRequest binds a run-scoped replacement to the same sealed
// plan, connection ownership, and approval evidence as its bounded pages.
// It intentionally has no provider cursor, target identifier, record, or
// database-specific type.
type FullOverwriteRunRequest struct {
	ConnectionID  string
	Generation    int64
	Plan          DestinationPlan
	Runtime       connectors.RuntimeConfig
	Source        connectors.Connector
	SourceRuntime connectors.RuntimeConfig
	Binding       DestinationBinding
	Stream        string
	Mode          synccontract.Mode
	BatchSize     int
	// TransformPlanJSON is canonical closed configuration supplied by App. It
	// contains no engine syntax; destinations parse only their shared contract.
	TransformPlanJSON string
	TransformPlanHash string
	Approval          DestinationApproval `json:"-"`
}

// FullOverwriteRun owns one destination-private shadow-and-publish lifecycle.
// ApplyFullOverwrite accepts one independently reopened warehouse workset at a
// time, so source pagination never turns into a whole-run in-memory payload.
// PublishFullOverwrite produces the sole durable acknowledgement; checkpoint
// advancement is performed by the generic orchestrator only after its
// ReadBackFullOverwrite confirmation. AbortFullOverwrite is idempotent cleanup
// for every pre-publication exit path.
type FullOverwriteRun interface {
	ApplyFullOverwrite(context.Context, DestinationApplyRequest) error
	PublishFullOverwrite(context.Context, FullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error)
	ReadBackFullOverwrite(context.Context, synccontract.DownstreamAcknowledgement) error
	AbortFullOverwrite(context.Context) error
}

// FullOverwritePublicationRequest is payload-free aggregate evidence supplied
// only after source emission completed. LastCheckpoint remains source-owned;
// the destination cannot replace it, and the orchestrator remains responsible
// for committing it after receipt reconciliation.
type FullOverwritePublicationRequest struct {
	LastCheckpoint *synccontract.CheckpointEnvelope
	Pages          int
	Records        int
	Tombstones     int
}

// ManagedTargetApprovalDestination marks a destination whose structural
// transport binding is authorized by App's closed managed-target plan/preview
// lifecycle. The marker keeps provider identity out of shared App dispatch.
type ManagedTargetApprovalDestination interface {
	DestinationExecutor
	ManagedTargetApprovalDestination()
}

// DefinitionOwnedApprovalDestination marks a destination whose plan, preview,
// approval, and workset binding are derived from a persisted connection and
// the selected connector definition. It exposes no caller-selected action or
// provider request surface.
type DefinitionOwnedApprovalDestination interface {
	DestinationExecutor
	DefinitionOwnedApprovalDestination()
}

// SourceRequest is the fixed source invocation context. It has no generic
// request URL, SQL text, command, action, or caller-authored payload.
type SourceRequest struct {
	Connector connectors.Connector
	Runtime   connectors.RuntimeConfig
	Stream    string
	// CursorField is the persisted stream-scoped cursor selected during
	// connection configuration. It is structural source state, never a raw
	// provider query fragment; native sources validate it against their live
	// catalog before they construct a page request.
	CursorField string
	Mode        synccontract.Mode
	BatchSize   int
	PrimaryKey  []string
	Resume      synccontract.ResumeExpectation
	Checkpoint  *synccontract.CheckpointEnvelope
	// UnitDeadline bounds one provider page fetch when the registered source
	// supports page-aware deadlines. It never bounds the whole transport run.
	UnitDeadline time.Duration `json:"-"`
	// RecordExtraction reports an individual provider page-fetch duration to
	// the orchestration measurement. It carries neither a record nor a route.
	RecordExtraction func(time.Duration) `json:"-"`
}

// SourcePage carries a bounded provider payload separately from #3810's
// tombstones and checkpoint candidate. The orchestrator never annotates the
// provider record map with `_polymetrics_*` fields.
type SourcePage struct {
	Records             []connectors.Record
	Tombstones          []synccontract.Tombstone
	CandidateCheckpoint synccontract.CheckpointEnvelope
	// DeferCheckpoint stages and applies this page without publishing its
	// candidate as resumable state. PostgreSQL bootstrap uses it for every
	// non-final snapshot page because one slot barrier covers the whole snapshot.
	DeferCheckpoint bool
}

// WarehouseStage is the required mediator. The source never invokes a
// destination executor directly: it places a bounded page into this typed
// stage, and the destination consumes the workset returned from that stage.
// A real durable implementation belongs to the corresponding warehouse/apply
// foundation; #3864 exercises this seam only with fakes.
type WarehouseStage interface {
	Stage(context.Context, WarehouseStageRequest) (WarehouseReceipt, error)
	Reopen(context.Context, WarehouseReceipt) (WarehouseWorkset, error)
}

// RetirableWarehouseStage owns bounded transient worksets and can remove an
// exact receipt only after its checkpoint is durable. It is intentionally
// optional: durable stages that retain artifacts for longer reconciliation do
// not need to implement it, while implementations that do must never accept a
// caller-provided path or a receipt that they do not own.
type RetirableWarehouseStage interface {
	WarehouseStage
	Retire(context.Context, WarehouseReceipt) error
}

type WarehouseStageRequest struct {
	// ConnectionID is the opaque project-owned identity of the connection that
	// owns the staged artifact. It is not a display name or credential value.
	ConnectionID string
	// Generation binds a receipt to the source generation that produced it.
	Generation      int64
	SourceName      string
	DestinationName string
	Stream          string
	Mode            synccontract.Mode
	Page            SourcePage
}

// WarehouseReceipt is the immutable, record-free identity of a durable staged
// page. A destination must reopen it through the stage; it must never receive
// the source-owned page directly.
type WarehouseReceipt struct {
	ID               string
	Owner            string
	Generation       int64
	Stream           string
	Mode             synccontract.Mode
	CheckpointSHA256 string
	TombstonesSHA256 string
	ManifestSHA256   string
	ContentSHA256    string
	ParquetSHA256    string
	Records          int
	Tombstones       int
}

func (r WarehouseReceipt) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("warehouse stage returned an empty receipt ID")
	}
	if strings.TrimSpace(r.Owner) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no owner", r.ID)
	}
	if r.Generation <= 0 {
		return fmt.Errorf("warehouse stage receipt %q has invalid generation", r.ID)
	}
	if strings.TrimSpace(r.Stream) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no stream", r.ID)
	}
	if err := r.Mode.Validate(); err != nil {
		return fmt.Errorf("warehouse stage receipt %q mode: %w", r.ID, err)
	}
	if strings.TrimSpace(r.CheckpointSHA256) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no checkpoint identity", r.ID)
	}
	if strings.TrimSpace(r.TombstonesSHA256) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no tombstone identity", r.ID)
	}
	if strings.TrimSpace(r.ManifestSHA256) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no manifest identity", r.ID)
	}
	if strings.TrimSpace(r.ContentSHA256) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no content identity", r.ID)
	}
	if strings.TrimSpace(r.ParquetSHA256) == "" {
		return fmt.Errorf("warehouse stage receipt %q has no parquet identity", r.ID)
	}
	if r.Records < 0 || r.Tombstones < 0 {
		return fmt.Errorf("warehouse stage receipt %q has negative bounded counts", r.ID)
	}
	return nil
}

// WarehouseWorkset is stage-owned output. CandidateCheckpoint is included for
// audit visibility, but the orchestrator commits the original source candidate
// so a stage cannot silently substitute a source position.
type WarehouseWorkset struct {
	ID string
	// SourceParquet is the immutable stage-owned materialization reopened for
	// this receipt. Destinations may consume it only together with the receipt
	// hashes; callers cannot supply it on the command surface.
	SourceParquet       string
	Records             []connectors.Record
	Tombstones          []synccontract.Tombstone
	CandidateCheckpoint synccontract.CheckpointEnvelope
}

// DestinationBinding is structural application identity supplied by App from
// persisted connection state. It contains no display name or credential.
type DestinationBinding struct {
	WorkspaceID       string
	SourceConnectorID string
	ConnectionID      string
	StreamID          string
	PrimaryKey        []string
}

type DestinationPlanRequest struct {
	Connector         connectors.Connector
	Runtime           connectors.RuntimeConfig
	Source            connectors.Connector
	SourceRuntime     connectors.RuntimeConfig
	Binding           DestinationBinding
	Stream            string
	Mode              synccontract.Mode
	BatchSize         int
	TransformPlanJSON string
	TransformPlanHash string
	ApplyStrategy     connectors.DestinationApplyStrategy
	Approval          DestinationApproval `json:"-"`
}

type DestinationPlan struct {
	ApplyStrategy          connectors.DestinationApplyStrategy
	TransformPlanHash      string
	ActionDefinitionSHA256 string
}

// DestinationApproval carries only the ephemeral result of a separately
// prepared PM plan -> preview -> approval lifecycle. It is intentionally
// non-serializable: warehouse receipts, runtime configuration, destination
// plans, and evidence artifacts never retain the operator token.
type DestinationApproval struct {
	PlanID                 string                            `json:"-"`
	ApprovalToken          string                            `json:"-"`
	Confirmation           connectors.WriteConfirmation      `json:"-"`
	Evidence               *connectors.WriteApprovalEvidence `json:"-"`
	Target                 connectors.WriteApprovalTarget    `json:"-"`
	PreviewDigest          string                            `json:"-"`
	ActionDefinitionSHA256 string                            `json:"-"`
	// AuthorizeNextUnit rechecks a standing authorization immediately before a
	// staged batch can cause a destination side effect. It is in-memory only:
	// receipts and checkpoints retain no token or authorization callback.
	AuthorizeNextUnit func(context.Context) error `json:"-"`
	// IssueWriteEvidence returns a fresh, destination-scoped evidence value for
	// the next declared typed write. The adapter alone supplies the concrete
	// prepared request to the connector engine; callers cannot use this hook to
	// choose a route, action, body, or approval target. A fresh value is needed
	// because the engine consumes evidence after each provider mutation while a
	// durable authorization may admit several bounded worksets.
	IssueWriteEvidence func(context.Context) (*connectors.WriteApprovalEvidence, error) `json:"-"`
}

type DestinationApplyRequest struct {
	// ConnectionID pins this apply to the connection that owns Receipt. The
	// destination must not infer ownership from caller-supplied paths.
	ConnectionID string
	Plan         DestinationPlan
	// Receipt is the immutable handle that produced Workset. It is passed by
	// the orchestrator after Reopen, not supplied by a destination caller.
	Receipt WarehouseReceipt
	Workset WarehouseWorkset
	// Runtime is supplied per call so a registered adapter never keeps
	// credential material after the request returns.
	Runtime       connectors.RuntimeConfig
	Source        connectors.Connector
	SourceRuntime connectors.RuntimeConfig
	Destination   connectors.Connector
	Binding       DestinationBinding
	Stream        string
	Mode          synccontract.Mode
	BatchSize     int
	// Approval was obtained before the transport run. Planning consumes its
	// single-use grant; Apply must revalidate its binding and expiry after the
	// warehouse workset is independently reopened and before each mutation.
	Approval DestinationApproval `json:"-"`
}

// DestinationReadBackRequest carries the exact durable acknowledgement and
// reopened workset that the destination must independently verify before the
// checkpoint CAS is allowed to advance.
type DestinationReadBackRequest struct {
	Plan            DestinationPlan
	Workset         WarehouseWorkset
	Acknowledgement synccontract.DownstreamAcknowledgement
	// Runtime is the same per-call endpoint configuration used by Apply.
	Runtime       connectors.RuntimeConfig
	Source        connectors.Connector
	SourceRuntime connectors.RuntimeConfig
	Destination   connectors.Connector
	Binding       DestinationBinding
	Stream        string
	Mode          synccontract.Mode
}

// PreflightRequest contains only identities and closed declarations needed to
// prove dispatch. It has no source payload and performs no provider I/O.
type PreflightRequest struct {
	Source            connectors.Connector
	Destination       connectors.Connector
	Stream            string
	Mode              synccontract.Mode
	DestinationAction string
}

// RunRequest wires a resolved connection into one shared orchestrator.
type RunRequest struct {
	ConnectionID       string
	Generation         int64
	Source             connectors.Connector
	SourceRuntime      connectors.RuntimeConfig
	Destination        connectors.Connector
	DestinationRuntime connectors.RuntimeConfig
	DestinationBinding DestinationBinding
	Stream             string
	// DestinationAction is the stable definition-owned action identity stored
	// on the connection stream. It is never caller-supplied execution input.
	DestinationAction string
	CursorField       string
	Mode              synccontract.Mode
	BatchSize         int
	// MaxInFlightBatches bounds the ordered Arrow full-overwrite pipeline.
	// Zero keeps the programmatic legacy serial behavior; the CLI/app selects
	// the user-facing default of two for an admitted fast path.
	MaxInFlightBatches int
	TransformPlanJSON  string
	TransformPlanHash  string
	// FastSegments is the connector-neutral versioned Parquet segment store
	// used only by the Arrow transformed full-overwrite path. Legacy transports
	// retain Stage and their JSONL-backed warehouse behavior unchanged.
	FastSegments FastSegmentStore
	// ByteCreditCapacity bounds retained Arrow payload bytes. Zero selects the
	// 512 MiB fast-path default; it is never a run deadline.
	ByteCreditCapacity int64
	Resume             synccontract.ResumeExpectation
	Checkpoint         *synccontract.CheckpointEnvelope
	// UnitDeadline bounds a single retryable provider-page fetch or destination
	// apply/read-back unit. Zero selects the conservative default; it is never
	// a deadline for the full source-to-destination run.
	UnitDeadline time.Duration
	// Approval is intentionally carried only in memory from App.RunETL to the
	// destination apply boundary. It is never written into a stage artifact.
	Approval DestinationApproval `json:"-"`
	Stage    WarehouseStage
	Commit   func(synccontract.CheckpointEnvelope) error
}

type Result struct {
	RecordsRead      int
	RecordsStaged    int
	RecordsApplied   int
	Pages            int
	ExtractElapsed   time.Duration
	TransformElapsed time.Duration
	StageElapsed     time.Duration
	ParquetElapsed   time.Duration
	ApplyElapsed     time.Duration
	// IndexConstraintElapsed is destination schema/index work performed once
	// for the private full-overwrite shadow. It is distinct from binary COPY so
	// a high COPY rate cannot hide a slow target build.
	IndexConstraintElapsed time.Duration
	PublishElapsed         time.Duration
	CheckpointElapsed      time.Duration
	WallElapsed            time.Duration
	SourceLogicalBytes     int64
	TransformedBytes       int64
	ParquetBytes           int64
	PeakCreditBytes        int64
	CreditWaitElapsed      time.Duration
	// DestinationResults retain every provider-returned response field, key,
	// value, receipt, status, body, occurrence ID, and credential-equal byte
	// verbatim. They remain opaque to the transport core: mapping and provider
	// protocol stay connector-owned. Only system-generated diagnostics, plans,
	// logs, and errors are rendered secret-safely.
	DestinationResults  []json.RawMessage
	CommittedCheckpoint *synccontract.CheckpointEnvelope
}

const defaultTransportUnitDeadline = time.Minute

func (r RunRequest) preflightRequest() PreflightRequest {
	return PreflightRequest{Source: r.Source, Destination: r.Destination, Stream: r.Stream, Mode: r.Mode, DestinationAction: r.DestinationAction}
}

func (r RunRequest) validateExecution() error {
	if r.BatchSize <= 0 {
		return fmt.Errorf("transport batch size must be positive")
	}
	if r.UnitDeadline < 0 {
		return fmt.Errorf("transport unit deadline must not be negative")
	}
	if r.ByteCreditCapacity < 0 {
		return fmt.Errorf("transport byte credit capacity must not be negative")
	}
	if r.MaxInFlightBatches < 0 || r.MaxInFlightBatches > 8 {
		return fmt.Errorf("transport max in-flight batches must be zero or between 1 and 8")
	}
	return nil
}

func (r RunRequest) unitDeadline() time.Duration {
	if r.UnitDeadline > 0 {
		return r.UnitDeadline
	}
	return defaultTransportUnitDeadline
}

// validateDispatchDependencies runs after closed registry preflight so an
// absent stage cannot hide a missing executor, invalid mode, or unsafe
// acknowledgement declaration. None of these checks can cause source I/O.
func (r RunRequest) validateDispatchDependencies() error {
	if r.Commit == nil {
		return fmt.Errorf("checkpoint committer is required for transport dispatch")
	}
	return nil
}

func (r RunRequest) validateLegacyDispatchDependencies() error {
	if isNilInterface(r.Stage) {
		return fmt.Errorf("warehouse stage is required for transport dispatch")
	}
	return nil
}

func cloneSourceRequest(request SourceRequest) SourceRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.PrimaryKey = append([]string(nil), request.PrimaryKey...)
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
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	return clone
}

func cloneDestinationReadBackRequest(request DestinationReadBackRequest) (DestinationReadBackRequest, error) {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	workset, err := cloneWarehouseWorkset(request.Workset)
	if err != nil {
		return DestinationReadBackRequest{}, err
	}
	clone.Workset = workset
	return clone, nil
}

func cloneDestinationApplyRequest(request DestinationApplyRequest) (DestinationApplyRequest, error) {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	workset, err := cloneWarehouseWorkset(request.Workset)
	if err != nil {
		return DestinationApplyRequest{}, err
	}
	clone.Workset = workset
	return clone, nil
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
