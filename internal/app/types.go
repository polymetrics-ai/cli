package app

import (
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
}

type StreamState struct {
	Connection          string                           `json:"connection"`
	Stream              string                           `json:"stream"`
	Checkpoint          *synccontract.CheckpointEnvelope `json:"checkpoint,omitempty"`
	GenerationID        int64                            `json:"generation_id"`
	LastSuccessfulRunID string                           `json:"last_successful_run_id,omitempty"`
	RecordsLoaded       int                              `json:"records_loaded,omitempty"`
	UpdatedAt           time.Time                        `json:"updated_at"`
}

type CreateConnectionRequest struct {
	Name        string                  `json:"name"`
	Source      EndpointConfig          `json:"source"`
	Destination EndpointConfig          `json:"destination"`
	Streams     map[string]StreamConfig `json:"streams"`
}

type Connection struct {
	// ID is the opaque generated identifier used as a warehouse path
	// component. Name is a display value and never becomes a path.
	ID          string                  `json:"id,omitempty"`
	Name        string                  `json:"name"`
	Source      EndpointConfig          `json:"source"`
	Destination EndpointConfig          `json:"destination"`
	Streams     map[string]StreamConfig `json:"streams"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
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
	Connection          string                            `json:"connection"`
	Stream              string                            `json:"stream"`
	BatchSize           int                               `json:"batch_size,omitempty"`
	DestinationApproval synctransport.DestinationApproval `json:"-"`
}

type Run struct {
	ID                 string            `json:"id"`
	Type               string            `json:"type"`
	Connection         string            `json:"connection,omitempty"`
	Stream             string            `json:"stream,omitempty"`
	Status             string            `json:"status"`
	RecordsRead        int               `json:"records_read"`
	RecordsTransformed int               `json:"records_transformed"`
	RecordsLoaded      int               `json:"records_loaded"`
	RecordsFailed      int               `json:"records_failed"`
	BatchCount         int               `json:"batch_count,omitempty"`
	Checkpoint         map[string]string `json:"checkpoint,omitempty"`
	Error              string            `json:"error,omitempty"`
	StartedAt          time.Time         `json:"started_at"`
	CompletedAt        time.Time         `json:"completed_at,omitempty"`
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
	ConnectorCommandOperation  string                       `json:"connector_command_operation,omitempty"`
	ConnectorCommandPathParams map[string]string            `json:"connector_command_path_params,omitempty"`
	ConnectorCommandQuery      map[string]string            `json:"connector_command_query,omitempty"`
	ConnectorCommandRecord     connectors.Record            `json:"connector_command_record,omitempty"`
	PayloadIdentity            []PayloadIdentity            `json:"payload_identity,omitempty"`
	ConfirmationChallenge      string                       `json:"confirmation_challenge,omitempty"`
	ConfirmationPolicy         connectors.WriteConfirmation `json:"confirmation,omitempty"`
	RedactFields               []string                     `json:"redact_fields,omitempty"`
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
	// TransportConnectionID and TransportBindingSHA256 are only used by the
	// closed GitHub issue-label walking slice. They bind a pre-run approval to
	// one connection configuration; neither field is caller-selectable write
	// input and neither contains an approval token or credential material.
	TransportConnectionID  string    `json:"transport_connection_id,omitempty"`
	TransportBindingSHA256 string    `json:"transport_binding_sha256,omitempty"`
	TransportForwardPlanID string    `json:"transport_forward_plan_id,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	ExpiresAt              time.Time `json:"expires_at"`
}

type RunReverseETLRequest struct {
	PlanID        string                       `json:"plan_id"`
	ApprovalToken string                       `json:"-"`
	Confirmation  connectors.WriteConfirmation `json:"-"`
	// WithheldFlags carries the command flags an operator re-supplies for
	// fields the plan withheld from disk. It is never persisted.
	WithheldFlags map[string][]string `json:"-"`
}

type ReverseRun struct {
	ID               string `json:"id"`
	PlanID           string `json:"plan_id"`
	Status           string `json:"status"`
	RecordsStaged    int    `json:"records_staged"`
	RecordsSucceeded int    `json:"records_succeeded"`
	RecordsFailed    int    `json:"records_failed"`
	Error            string `json:"error,omitempty"`
	// OperationDirectWrite is populated only for a successful direct_write
	// command. Its body is decoded according to the operation output policy.
	OperationDirectWrite *connectors.OperationDirectWriteResult `json:"operation_direct_write,omitempty"`
	StartedAt            time.Time                              `json:"started_at"`
	CompletedAt          time.Time                              `json:"completed_at,omitempty"`
}
