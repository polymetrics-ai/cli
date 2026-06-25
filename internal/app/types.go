package app

import (
	"time"

	"polymetrics.ai/internal/connectors"
)

type AddCredentialRequest struct {
	Name      string            `json:"name"`
	Connector string            `json:"connector"`
	Config    map[string]string `json:"config"`
	Secrets   map[string]string `json:"-"`
}

type CredentialMeta struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Connector       string            `json:"connector"`
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
	SyncMode         string   `json:"sync_mode"`
	CursorField      string   `json:"cursor_field,omitempty"`
	PrimaryKey       []string `json:"primary_key,omitempty"`
	DestinationTable string   `json:"destination_table,omitempty"`
}

type StreamState struct {
	Connection          string    `json:"connection"`
	Stream              string    `json:"stream"`
	Cursor              string    `json:"cursor,omitempty"`
	GenerationID        int64     `json:"generation_id"`
	LastSuccessfulRunID string    `json:"last_successful_run_id,omitempty"`
	RecordsLoaded       int       `json:"records_loaded,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreateConnectionRequest struct {
	Name        string                  `json:"name"`
	Source      EndpointConfig          `json:"source"`
	Destination EndpointConfig          `json:"destination"`
	Streams     map[string]StreamConfig `json:"streams"`
}

type Connection struct {
	Name        string                  `json:"name"`
	Source      EndpointConfig          `json:"source"`
	Destination EndpointConfig          `json:"destination"`
	Streams     map[string]StreamConfig `json:"streams"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type CatalogSnapshot struct {
	Connection string             `json:"connection"`
	Catalog    connectors.Catalog `json:"catalog"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type RunETLRequest struct {
	Connection string `json:"connection"`
	Stream     string `json:"stream"`
	BatchSize  int    `json:"batch_size,omitempty"`
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
	Limit int    `json:"limit"`
}

type PlanReverseETLRequest struct {
	Name                  string            `json:"name"`
	SourceTable           string            `json:"source_table"`
	DestinationConnector  string            `json:"destination_connector"`
	DestinationCredential string            `json:"destination_credential"`
	DestinationConfig     map[string]string `json:"destination_config,omitempty"`
	Action                string            `json:"action"`
	Mappings              map[string]string `json:"mappings"`
	Limit                 int               `json:"limit,omitempty"`
}

type ReversePlan struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	Status                string              `json:"status"`
	SourceTable           string              `json:"source_table"`
	DestinationConnector  string              `json:"destination_connector"`
	DestinationCredential string              `json:"destination_credential"`
	DestinationConfig     map[string]string   `json:"destination_config,omitempty"`
	Action                string              `json:"action"`
	Mappings              map[string]string   `json:"mappings"`
	RecordCount           int                 `json:"record_count"`
	Sample                []connectors.Record `json:"sample,omitempty"`
	PlanHash              string              `json:"plan_hash"`
	ApprovalTokenHash     string              `json:"approval_token_hash,omitempty"`
	ApprovalToken         string              `json:"approval_token,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	ExpiresAt             time.Time           `json:"expires_at"`
}

type RunReverseETLRequest struct {
	PlanID        string `json:"plan_id"`
	ApprovalToken string `json:"-"`
}

type ReverseRun struct {
	ID               string    `json:"id"`
	PlanID           string    `json:"plan_id"`
	Status           string    `json:"status"`
	RecordsStaged    int       `json:"records_staged"`
	RecordsSucceeded int       `json:"records_succeeded"`
	RecordsFailed    int       `json:"records_failed"`
	Error            string    `json:"error,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
}
