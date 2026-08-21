package app

import (
	"encoding/json"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

type AddCredentialRequest struct {
	Name           string            `json:"name"`
	Connector      string            `json:"connector"`
	Config         map[string]string `json:"config"`
	Secrets        map[string]string `json:"-"`
	ProviderFamily string            `json:"provider_family,omitempty"`
	AuthProfile    string            `json:"auth_profile,omitempty"`
	LinkCredential string            `json:"-"`
}

type CredentialCoordinationDeclarationError struct {
	err error
}

func (e *CredentialCoordinationDeclarationError) Error() string { return e.err.Error() }

func (e *CredentialCoordinationDeclarationError) Unwrap() error { return e.err }

type CredentialLinkValidationError struct {
	err error
}

func (e *CredentialLinkValidationError) Error() string { return e.err.Error() }

func (e *CredentialLinkValidationError) Unwrap() error { return e.err }

// ApprovalConsumptionUncertainError stops a reverse plan when its single-use
// approval may have been consumed but the durability result is uncertain.
// Callers must create a new plan rather than retrying the old approval.
type ApprovalConsumptionUncertainError struct {
	PlanID     string
	ConsumedAt time.Time
	err        error
}

// AuthorizationScopeChangedError reports the exact durable authorization
// property that differs from the scope approved by the operator. It is typed
// so callers can require a new plan/preview/proceed without parsing text.
type AuthorizationScopeChangedError struct {
	Reference string
	Property  string
}

func (e *AuthorizationScopeChangedError) Error() string {
	if e == nil {
		return "authorization scope changed"
	}
	return fmt.Sprintf("authorization %q scope changed: %s requires re-approval", e.Reference, e.Property)
}

// AuthorizationRevokedError reports an explicit per-connection authorization
// revocation before a destination dispatch can occur.
type AuthorizationRevokedError struct{ Reference string }

func (e *AuthorizationRevokedError) Error() string {
	if e == nil {
		return "authorization has been revoked"
	}
	return fmt.Sprintf("authorization %q has been revoked", e.Reference)
}

// AuthorizationExpiredError reports a standing authorization whose approved
// scope expiry has elapsed before a destination dispatch can occur.
type AuthorizationExpiredError struct{ Reference string }

func (e *AuthorizationExpiredError) Error() string {
	if e == nil {
		return "authorization has expired"
	}
	return fmt.Sprintf("authorization %q has expired", e.Reference)
}

// AuthorizationTokenReplayError reports use of the one-time plan token after
// it created a durable authorization record.
type AuthorizationTokenReplayError struct{ Reference string }

func (e *AuthorizationTokenReplayError) Error() string {
	if e == nil {
		return "authorization token has already been consumed"
	}
	return fmt.Sprintf("authorization token for %q has already been consumed", e.Reference)
}

func (e *ApprovalConsumptionUncertainError) Error() string {
	if e == nil {
		return "reverse plan approval consumption transition is uncertain; create a new reverse plan and obtain a fresh preview and approval before retrying"
	}
	return fmt.Sprintf("reverse plan %q approval consumption transition is uncertain; create a new reverse plan and obtain a fresh preview and approval before retrying", e.PlanID)
}

func (e *ApprovalConsumptionUncertainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type CredentialMeta struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Connector       string            `json:"connector"`
	ProviderFamily  string            `json:"provider_family,omitempty"`
	AuthProfile     string            `json:"auth_profile,omitempty"`
	Config          map[string]string `json:"config,omitempty"`
	SecretFields    []string          `json:"secret_fields,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastValidatedAt time.Time         `json:"last_validated_at,omitempty"`
}

type EndpointConfig struct {
	Connector  string            `json:"connector"`
	Credential string            `json:"credential"`
	Config     map[string]string `json:"config,omitempty"`
}

type StreamConfig struct {
	// StreamID is allocated and persisted once when the stream is attached to a
	// connection. It is structural identity for managed destinations; map keys,
	// display names, and destination tables remain mutable configuration.
	StreamID            string   `json:"stream_id,omitempty"`
	SyncMode            string   `json:"sync_mode"`
	LegacyCompatibility bool     `json:"legacy_compatibility,omitempty"`
	CursorField         string   `json:"cursor_field,omitempty"`
	PrimaryKey          []string `json:"primary_key,omitempty"`
	DestinationTable    string   `json:"destination_table,omitempty"`
	// TransformPlan is normalized TransformPlanV1 JSON. It never stores a
	// source filename, arbitrary SQL, or raw user formatting; TransformPlanHash
	// binds the persisted closed form into later plans and approvals.
	TransformPlan     string `json:"transform_plan,omitempty"`
	TransformPlanHash string `json:"transform_plan_hash,omitempty"`
	// DestinationAction is a stable action name from the selected destination
	// definition. It is required only when that destination declares more than
	// one action for this stream mode; execution receives no action override.
	DestinationAction string `json:"destination_action,omitempty"`
}

type StreamState struct {
	Connection   string                           `json:"connection"`
	Stream       string                           `json:"stream"`
	Checkpoint   *synccontract.CheckpointEnvelope `json:"checkpoint,omitempty"`
	GenerationID int64                            `json:"generation_id"`
	// ActiveWorkID and ActiveWorkFence form one durable, connection-and-stream
	// scoped work lease. They are present only while a source/stage/destination
	// run owns the stream; effects and checkpoint commits renew the same fence
	// before touching I/O, and terminal completion clears the work ID without
	// rewinding the monotonic fence.
	ActiveWorkID         string     `json:"active_work_id,omitempty"`
	ActiveWorkFence      int64      `json:"active_work_fence,omitempty"`
	ActiveWorkLeaseUntil *time.Time `json:"active_work_lease_until,omitempty"`
	LastSuccessfulRunID  string     `json:"last_successful_run_id,omitempty"`
	RecordsLoaded        int        `json:"records_loaded,omitempty"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateConnectionRequest struct {
	Name              string                  `json:"name"`
	Source            EndpointConfig          `json:"source"`
	Destination       EndpointConfig          `json:"destination"`
	Streams           map[string]StreamConfig `json:"streams"`
	TargetCopyWorkers int                     `json:"target_copy_workers,omitempty"`
}

type Connection struct {
	// ID is the opaque generated identifier used as a warehouse path
	// component. Name is a display value and never becomes a path.
	ID                string                  `json:"id,omitempty"`
	Name              string                  `json:"name"`
	Source            EndpointConfig          `json:"source"`
	Destination       EndpointConfig          `json:"destination"`
	Streams           map[string]StreamConfig `json:"streams"`
	TargetCopyWorkers int                     `json:"target_copy_workers,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

func cloneEndpointConfig(config EndpointConfig) EndpointConfig {
	clone := config
	clone.Config = cloneStringMap(config.Config)
	return clone
}

func cloneStreamConfig(config StreamConfig) StreamConfig {
	clone := config
	clone.PrimaryKey = append([]string(nil), config.PrimaryKey...)
	return clone
}

func cloneStreamConfigs(configs map[string]StreamConfig) map[string]StreamConfig {
	if configs == nil {
		return nil
	}
	clone := make(map[string]StreamConfig, len(configs))
	for name, config := range configs {
		clone[name] = cloneStreamConfig(config)
	}
	return clone
}

func cloneCreateConnectionRequest(req CreateConnectionRequest) CreateConnectionRequest {
	clone := req
	clone.Source = cloneEndpointConfig(req.Source)
	clone.Destination = cloneEndpointConfig(req.Destination)
	clone.Streams = cloneStreamConfigs(req.Streams)
	return clone
}

func cloneConnection(connection Connection) Connection {
	clone := connection
	clone.Source = cloneEndpointConfig(connection.Source)
	clone.Destination = cloneEndpointConfig(connection.Destination)
	clone.Streams = cloneStreamConfigs(connection.Streams)
	return clone
}

type CatalogSnapshot struct {
	Connection string             `json:"connection"`
	Catalog    connectors.Catalog `json:"catalog"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// catalogReference is the compact state.json index for an account catalog.
// AccountKey is an opaque CoordinationIdentity projection; it is never a
// credential value and is intentionally kept out of CatalogSnapshot and CLI
// output. The referenced file contains schemas only, never this key.
type catalogReference struct {
	Connector  string                   `json:"connector"`
	AccountKey connectors.AuthCohortKey `json:"account_key"`
	File       string                   `json:"file"`
}

type RunETLRequest struct {
	Connection string `json:"connection"`
	Stream     string `json:"stream"`
	BatchSize  int    `json:"batch_size,omitempty"`
	// MaxInFlightBatches is an optional ordered Arrow full-overwrite pipeline
	// bound. Zero means the caller did not select the CLI/app capability
	// control; admitted fast paths choose their documented default of two.
	MaxInFlightBatches           int                               `json:"max_in_flight_batches,omitempty"`
	DestinationApproval          synctransport.DestinationApproval `json:"-"`
	rateParkingResumeCheckpoint  *synccontract.CheckpointEnvelope
	rateParkingRearmAttemptRunID string
}

type Run struct {
	ID                                string            `json:"id"`
	Type                              string            `json:"type"`
	Connection                        string            `json:"connection,omitempty"`
	Stream                            string            `json:"stream,omitempty"`
	Status                            string            `json:"status"`
	RecordsRead                       int               `json:"records_read"`
	RecordsTransformed                int               `json:"records_transformed"`
	RecordsLoaded                     int               `json:"records_loaded"`
	RecordsFailed                     int               `json:"records_failed"`
	BatchSize                         int               `json:"batch_size,omitempty"`
	BatchCount                        int               `json:"batch_count,omitempty"`
	Checkpoint                        map[string]string `json:"checkpoint,omitempty"`
	DeclarativeTypedDestinationPlanID string            `json:"declarative_typed_destination_plan_id,omitempty"`
	RateParkingRearmAttemptRunID      string            `json:"rate_parking_rearm_attempt_run_id,omitempty"`
	// TransportPhaseMeasurement is emitted with the terminal run transition on
	// closed source -> warehouse -> destination transports. It deliberately
	// contains counts and elapsed times only, never records, paths, tokens, or
	// connector configuration.
	TransportPhaseMeasurement *TransportPhaseMeasurement `json:"transport_phase_measurement,omitempty"`
	// DestinationResults retains each completed declarative typed destination
	// action's full provider result.
	DestinationResults []json.RawMessage `json:"destination_results,omitempty"`
	Error              string            `json:"error,omitempty"`
	StartedAt          time.Time         `json:"started_at"`
	CompletedAt        time.Time         `json:"completed_at,omitempty"`
}

type TransportPhaseMeasurement struct {
	// Legacy fields remain for existing non-fast transport readers.
	ExtractedRecords         int   `json:"extracted_records"`
	WarehouseParquetRecords  int   `json:"warehouse_parquet_records"`
	PostgreSQLAppliedRecords int   `json:"postgresql_applied_records"`
	ExtractElapsedNanos      int64 `json:"extract_elapsed_ns"`
	WarehouseElapsedNanos    int64 `json:"warehouse_elapsed_ns"`
	PostgreSQLElapsedNanos   int64 `json:"postgresql_elapsed_ns"`

	// Generic fast-path counters report input as logical source bytes (the
	// source Arrow buffer bytes before transformation), never Parquet, pgwire,
	// or target-storage bytes. Phase intervals may overlap; critical_path_ns is
	// the run wall clock used for both throughput rates.
	SourceRecords                    int     `json:"source_records"`
	TransformedRecords               int     `json:"transformed_records"`
	CopyAppliedRecords               int     `json:"copy_applied_records"`
	SourceLogicalBytes               int64   `json:"source_logical_bytes"`
	TransformedLogicalBytes          int64   `json:"transformed_logical_bytes"`
	ParquetBytes                     int64   `json:"parquet_bytes"`
	SourceReadElapsedNanos           int64   `json:"source_read_elapsed_ns"`
	TransformElapsedNanos            int64   `json:"transform_elapsed_ns"`
	ParquetCloseElapsedNanos         int64   `json:"parquet_close_fsync_elapsed_ns"`
	BinaryCOPYElapsedNanos           int64   `json:"binary_copy_elapsed_ns"`
	IndexConstraintBuildElapsedNanos int64   `json:"index_constraint_build_elapsed_ns"`
	PublishReceiptElapsedNanos       int64   `json:"publish_receipt_elapsed_ns"`
	CheckpointElapsedNanos           int64   `json:"checkpoint_elapsed_ns"`
	CriticalPathElapsedNanos         int64   `json:"critical_path_elapsed_ns"`
	PeakCreditBytes                  int64   `json:"peak_credit_bytes"`
	ByteCreditWaitElapsedNanos       int64   `json:"byte_credit_wait_elapsed_ns"`
	InputDecimalMBPerSecond          float64 `json:"input_decimal_mb_per_second"`
	InputMiBPerSecond                float64 `json:"input_mib_per_second"`
}

type QueryTableRequest struct {
	Table string `json:"table"`
	// Connection scopes the read to one connection's tables. It is required
	// only when more than one connection materializes the same table name.
	Connection string `json:"connection,omitempty"`
	Limit      int    `json:"limit"`
}

// ActionSourceReadRequest identifies the warehouse table an action step reads.
// Connection selects one owner's materialization; the `_unattributed` sentinel
// selects a root-owned table. An empty selector preserves typed ambiguity when
// several owners materialize the same name.
type ActionSourceReadRequest struct {
	Table      string `json:"table"`
	Connection string `json:"connection,omitempty"`
}

type QuerySQLOrigin uint8

const (
	QuerySQLOriginGeneric QuerySQLOrigin = iota
	QuerySQLOriginFlow
)

// QuerySQLRequest describes a read-only analytical query over the local
// warehouse. Connection scopes every table view available to the query, so a
// flow or other caller cannot silently resolve a same-named table from another
// connection. UnattributedConnection selects only root-owned tables.
type QuerySQLRequest struct {
	SQL        string         `json:"sql"`
	Connection string         `json:"connection,omitempty"`
	Limit      int            `json:"limit"`
	Origin     QuerySQLOrigin `json:"-"`

	// sameOwnerCaseEquivalentDestinationCollisions is an App-owned immutable
	// configuration snapshot. It stays private so callers cannot manufacture a
	// warehouse policy; QuerySQL adds it beside the resolver snapshot used by
	// DuckDB for this one request.
	sameOwnerCaseEquivalentDestinationCollisions []warehouseDestinationCollision
}

type PlanReverseETLRequest struct {
	Name        string `json:"name"`
	SourceTable string `json:"source_table"`
	// SourceConnection scopes the source table to one connection. It is
	// required only when several connections materialize the same table name.
	SourceConnection      string            `json:"source_connection,omitempty"`
	DestinationConnector  string            `json:"destination_connector"`
	DestinationCredential string            `json:"destination_credential"`
	DestinationConfig     map[string]string `json:"destination_config,omitempty"`
	Action                string            `json:"action"`
	Mappings              map[string]string `json:"mappings"`
	Limit                 int               `json:"limit,omitempty"`
}

type PlanConnectorCommandRequest struct {
	Name       string              `json:"name"`
	Connector  string              `json:"connector"`
	Credential string              `json:"credential"`
	Config     map[string]string   `json:"config,omitempty"`
	Path       []string            `json:"path"`
	Flags      map[string][]string `json:"flags,omitempty"`
	Preview    bool                `json:"preview,omitempty"`
}

type PayloadIdentity struct {
	RecordIndex     int    `json:"record_index"`
	Field           string `json:"field"`
	PathHash        string `json:"path_hash"`
	ContentSHA256   string `json:"content_sha256"`
	SizeBytes       int64  `json:"size_bytes"`
	ModTimeUnixNano int64  `json:"mod_time_unix_nano"`
}

type ReversePlan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Mode        string `json:"mode,omitempty"`
	SourceTable string `json:"source_table"`
	// SourceConnection is the connection the plan was built against, resolved
	// once at plan time. Preview and run scope their reads by it so a plan
	// keeps resolving to the same table after another connection materializes
	// one of the same name.
	SourceConnection      string            `json:"source_connection,omitempty"`
	DestinationConnector  string            `json:"destination_connector"`
	DestinationCredential string            `json:"destination_credential"`
	DestinationConfig     map[string]string `json:"destination_config,omitempty"`
	Action                string            `json:"action"`
	Mappings              map[string]string `json:"mappings"`
	ConnectorCommand      string            `json:"connector_command,omitempty"`
	ConnectorCommandPath  []string          `json:"connector_command_path,omitempty"`
	// ConnectorCommandOperation identifies a direct_write operation. When it is
	// empty, the plan retains the existing writes.json action path.
	ConnectorCommandOperation  string            `json:"connector_command_operation,omitempty"`
	ConnectorCommandPathParams map[string]string `json:"connector_command_path_params,omitempty"`
	ConnectorCommandQuery      map[string]string `json:"connector_command_query,omitempty"`
	// ConnectorCommandHeaders is the closed, declaration-owned request-header
	// input for a direct_write plan. It participates in the plan hash and the
	// engine preview digest; CLI presentation clears it like the body record.
	ConnectorCommandHeaders      map[string]string            `json:"connector_command_headers,omitempty"`
	ConnectorCommandHeaderValues map[string][]string          `json:"connector_command_header_values,omitempty"`
	ConnectorCommandRecord       connectors.Record            `json:"connector_command_record,omitempty"`
	PayloadIdentity              []PayloadIdentity            `json:"payload_identity,omitempty"`
	ConfirmationChallenge        string                       `json:"confirmation_challenge,omitempty"`
	ConfirmationPolicy           connectors.WriteConfirmation `json:"confirmation,omitempty"`
	RedactFields                 []string                     `json:"redact_fields,omitempty"`
	// WithheldFields names the record fields this plan actually removed before
	// persisting, which is a subset of RedactFields: a declared field the
	// operator never supplied was never present and is never owed back. Only
	// these have to be re-supplied to preview or run the plan.
	WithheldFields      []string                       `json:"withheld_fields,omitempty"`
	RecordCount         int                            `json:"record_count"`
	Sample              []connectors.Record            `json:"sample,omitempty"`
	PlanHash            string                         `json:"plan_hash"`
	PlanSeal            *connectors.WritePlanSeal      `json:"plan_seal,omitempty"`
	PreviewDigest       string                         `json:"preview_digest,omitempty"`
	PreviewedAt         time.Time                      `json:"previewed_at,omitempty"`
	ApprovalTokenHash   string                         `json:"approval_token_hash,omitempty"`
	ApprovalGrant       *connectors.WriteApprovalGrant `json:"approval_grant,omitempty"`
	ApprovalToken       string                         `json:"approval_token,omitempty"`
	ApprovalConsumedAt  time.Time                      `json:"approval_consumed_at,omitempty"`
	ApprovalUncertainAt time.Time                      `json:"approval_consumption_uncertain_at,omitempty"`
	// AuthorizationReference identifies the durable, non-secret scope record
	// minted by the single-use approval. It is safe to persist in a schedule or
	// print as a reference; the scope record never carries the token.
	AuthorizationReference string `json:"authorization_reference,omitempty"`
	// TransportConnectionID and TransportBindingSHA256 are used by closed
	// definition-selected transport writes. They bind a pre-run approval to
	// one connection configuration; neither field is caller-selectable write
	// input and neither contains an approval token or credential material.
	TransportConnectionID           string `json:"transport_connection_id,omitempty"`
	TransportStream                 string `json:"transport_stream,omitempty"`
	TransportBindingSHA256          string `json:"transport_binding_sha256,omitempty"`
	TransportActionDefinitionSHA256 string `json:"transport_action_definition_sha256,omitempty"`
	TransportForwardPlanID          string `json:"transport_forward_plan_id,omitempty"`
	// AuthorizationLifetime is a bounded day-scale lifetime requested when a
	// PostgreSQL managed-target transport plan is created. It is included in
	// the sealed plan hash before its single-use approval token is issued.
	AuthorizationLifetime time.Duration `json:"authorization_lifetime_ns,omitempty"`
	// TargetCopyWorkers and TargetCopyWorkerMaximum are safe, persisted
	// connection-policy evidence for the target's bounded immutable COPY
	// capacity. They carry neither a credential nor a target identifier.
	TargetCopyWorkers       int       `json:"target_copy_workers,omitempty"`
	TargetCopyWorkerMaximum int       `json:"target_copy_worker_maximum,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	ExpiresAt               time.Time `json:"expires_at"`
}

type RunReverseETLRequest struct {
	PlanID        string                       `json:"plan_id"`
	ApprovalToken string                       `json:"-"`
	Confirmation  connectors.WriteConfirmation `json:"-"`
	// WithheldFlags carries the command flags an operator re-supplies for
	// fields the plan withheld from disk. It is never persisted.
	WithheldFlags map[string][]string `json:"-"`
}

// FlowActionExecutionRequest is the fully resolved, connector-backed action
// request emitted by a flow step. Records are supplied by the flow engine only
// after its connection-scoped source read; they are never persisted here.
// AuthorizationReference is the durable, content-free record minted by an
// earlier plan → preview → approval → execute lifecycle.
type FlowActionExecutionRequest struct {
	FlowName               string
	StepID                 string
	RunID                  string
	ManifestDigest         string
	SourceTable            string
	SourceConnection       string
	DestinationTable       string
	DestinationConnector   string
	DestinationCredential  string
	DestinationConfig      map[string]string
	Action                 string
	Mappings               map[string]string
	AuthorizationReference string
	ReadBackStream         string
	Records                []connectors.Record
}

// PreparedFlowAction is an in-memory, payload-bound execution prepared from a
// standing authorization. Its exported fields are safe opaque evidence, not
// authority. Request, mapped payload, and preview remain process-private.
type PreparedFlowAction struct {
	Identity string
	FiringID string

	request        FlowActionExecutionRequest
	mappedRecords  []connectors.Record
	preview        connectors.WritePreview
	sealedIdentity string
	scopeIdentity  string
}

// FlowActionExecutionResult contains only observable delivery accounting and
// an opaque receipt identifier. It deliberately excludes source records,
// payload content, credentials, and destination configuration.
type FlowActionExecutionResult struct {
	RecordsAttempted          int
	RecordsSucceeded          int
	RecordsFailed             int
	ReceiptID                 string
	PreparedExecutionIdentity string
	FiringID                  string
}

// FlowActionReceipt is durable evidence that a connector acknowledged a flow
// action and its configured target stream was read back successfully. It is
// recorded only after both events, so it is safe for flow checkpointing.
type FlowActionReceipt struct {
	ID                        string    `json:"id"`
	RunID                     string    `json:"run_id"`
	FiringID                  string    `json:"firing_id"`
	PreparedExecutionIdentity string    `json:"prepared_execution_identity"`
	FlowName                  string    `json:"flow_name"`
	StepID                    string    `json:"step_id"`
	AuthorizationReference    string    `json:"authorization_reference"`
	DestinationConnector      string    `json:"destination_connector"`
	Action                    string    `json:"action"`
	AcknowledgedAt            time.Time `json:"acknowledged_at"`
	ReadBackAt                time.Time `json:"read_back_at"`
}

type ReverseRun struct {
	ID               string `json:"id"`
	PlanID           string `json:"plan_id"`
	Status           string `json:"status"`
	RecordsStaged    int    `json:"records_staged"`
	RecordsSucceeded int    `json:"records_succeeded"`
	RecordsFailed    int    `json:"records_failed"`
	Error            string `json:"error,omitempty"`
	// DestinationResult retains the complete typed write result for regular
	// reverse ETL.
	DestinationResult    json.RawMessage                        `json:"destination_result,omitempty"`
	OperationDirectWrite *connectors.OperationDirectWriteResult `json:"operation_direct_write,omitempty"`
	StartedAt            time.Time                              `json:"started_at"`
	CompletedAt          time.Time                              `json:"completed_at,omitempty"`
}
