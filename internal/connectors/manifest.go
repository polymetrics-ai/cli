package connectors

import (
	"context"
	"strings"

	"polymetrics.ai/internal/synccontract"
)

type ConfigField struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
}

type SecretField struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Required     bool   `json:"required,omitempty"`
	RequiredWhen string `json:"required_when,omitempty"`
}

type PaginationSpec struct {
	Type           string `json:"type,omitempty"`
	PageSizeField  string `json:"page_size_field,omitempty"`
	PageLimitField string `json:"page_limit_field,omitempty"`
	DelayField     string `json:"delay_field,omitempty"`
	DefaultLimit   string `json:"default_limit,omitempty"`
}

type RiskSpec struct {
	Read     string `json:"read,omitempty"`
	Write    string `json:"write,omitempty"`
	Mutation string `json:"mutation,omitempty"`
	Approval string `json:"approval,omitempty"`
}

type WriteActionSpec struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	RequiredFields []string `json:"required_fields,omitempty"`
	// RequiredAnyFields carries an either/or requirement that RequiredFields
	// cannot express: each inner slice is one group of fields that together
	// satisfy the constraint, and a write must supply every field of at least
	// one group. amazon-sqs set_queue_attributes is the shape it exists for —
	// `attributes`, OR `attribute_name` together with `attribute_value`.
	//
	// It is a separate field because RequiredFields is a list of FIELD NAMES
	// that `pm connectors inspect --json` publishes; folding a synthesized
	// sentence into it would hand a machine consumer a string that is not a
	// field.
	RequiredAnyFields [][]string `json:"required_any_fields,omitempty"`
	OptionalFields    []string   `json:"optional_fields,omitempty"`
	Method            string     `json:"method,omitempty"`
	Path              string     `json:"path,omitempty"`
	RedactFields      []string   `json:"redact_fields,omitempty"`
	Risk              string     `json:"risk,omitempty"`
	// Batchable mirrors the bundle's "batchable" declaration. nil means the
	// action never declared it and is therefore batchable; see IsBatchable for
	// why this is a pointer rather than a bool.
	Batchable *bool  `json:"batchable,omitempty"`
	Confirm   string `json:"confirm,omitempty"`
	// AllowsUnchanged is true only when the connector declaration admits a
	// successful no-change terminal outcome for this exact action. It is
	// intentionally distinct from the HTTP method: an arbitrary DELETE must
	// not manufacture a complete reverse-ETL acknowledgement.
	AllowsUnchanged bool `json:"allows_unchanged,omitempty"`
}

// BinaryUploadSource is the declaration-derived boundary a public
// binary_upload command must satisfy. It intentionally exposes neither a URL
// nor a request body: the command runner may only bind its named source field
// to the existing approval-bound write lifecycle.
type BinaryUploadSource struct {
	Field             string   `json:"field"`
	MaxBytes          int64    `json:"max_bytes"`
	AllowedMediaTypes []string `json:"allowed_media_types"`
}

// ConfirmationForWriteAction normalizes manifest metadata into the closed
// runtime policy. DELETE is destructive by construction, so omission cannot
// downgrade it. An unknown non-empty legacy declaration also fails closed.
func ConfirmationForWriteAction(action WriteActionSpec) WriteConfirmation {
	if strings.EqualFold(strings.TrimSpace(action.Method), "DELETE") {
		return WriteConfirmation{Kind: ConfirmationKindDestructive}
	}
	confirmation, err := ParseWriteConfirmation(action.Confirm)
	if err != nil && strings.TrimSpace(action.Confirm) != "" {
		return WriteConfirmation{Kind: ConfirmationKindDestructive}
	}
	return confirmation
}

// IsBatchable reports whether the action may run from a bulk reverse ETL plan.
// Only an explicit false says no, so connectors that never declare the field —
// which is all of them today — stay batchable.
func (s WriteActionSpec) IsBatchable() bool {
	return s.Batchable == nil || *s.Batchable
}

type AuthModeSpec struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	ConfigFields []string `json:"config_fields,omitempty"`
	SecretFields []string `json:"secret_fields,omitempty"`
	Read         bool     `json:"read"`
	Write        bool     `json:"write"`
}

type Manifest struct {
	Metadata             Metadata          `json:"metadata"`
	ConfigFields         []ConfigField     `json:"config_fields,omitempty"`
	SecretFields         []SecretField     `json:"secret_fields,omitempty"`
	AuthModes            []AuthModeSpec    `json:"auth_modes,omitempty"`
	Streams              []Stream          `json:"streams,omitempty"`
	WriteActions         []WriteActionSpec `json:"write_actions,omitempty"`
	SyncModes            []string          `json:"sync_modes,omitempty"`
	SourceSyncModes      []string          `json:"source_sync_modes,omitempty"`
	DestinationSyncModes []string          `json:"destination_sync_modes,omitempty"`
	Pagination           PaginationSpec    `json:"pagination,omitempty"`
	Risk                 RiskSpec          `json:"risk"`
}

type ManifestProvider interface {
	Manifest() Manifest
}

func ManifestOf(c Connector) Manifest {
	var manifest Manifest
	if provider, ok := c.(ManifestProvider); ok {
		manifest = provider.Manifest()
	} else {
		manifest = Manifest{
			Metadata: c.Metadata(),
			Risk: RiskSpec{
				Read:     "connector-specific",
				Write:    "connector-specific",
				Approval: "external mutations require preview and approval",
			},
		}
	}
	manifest.Metadata.Capabilities.CDC = MetadataOf(c).Capabilities.CDC
	return manifestWithIcon(manifest)
}

func (r *Registry) Manifest(name string) (Manifest, bool) {
	c, ok := r.Get(name)
	if !ok {
		return Manifest{}, false
	}
	return ManifestOf(c), true
}

func (r *Registry) ListManifests() []Manifest {
	list := r.List()
	out := make([]Manifest, 0, len(list))
	for _, item := range list {
		c, ok := r.Get(item.Name)
		if !ok {
			continue
		}
		out = append(out, ManifestOf(c))
	}
	return out
}

func allSyncModes() []string {
	return synccontract.PublicModeNames()
}

func readSourceSyncModes() []string {
	return []string{"full_refresh", "incremental"}
}

func warehouseDestinationSyncModes() []string {
	return []string{"append", "overwrite", "append_dedup", "overwrite_dedup"}
}

func (s Sample) Manifest() Manifest {
	catalog, _ := s.Catalog(context.Background(), RuntimeConfig{})
	return Manifest{
		Metadata:             s.Metadata(),
		Streams:              catalog.Streams,
		SyncModes:            allSyncModes(),
		SourceSyncModes:      readSourceSyncModes(),
		DestinationSyncModes: warehouseDestinationSyncModes(),
		Risk: RiskSpec{
			Read:     "local deterministic sample data",
			Write:    "unsupported",
			Mutation: "none",
			Approval: "not required for reads",
		},
	}
}

func (f File) Manifest() Manifest {
	return Manifest{
		Metadata: f.Metadata(),
		ConfigFields: []ConfigField{
			{Name: "path", Description: "Local JSONL or CSV file path.", Required: true},
			{Name: "stream", Description: "Optional stream name override."},
		},
		Streams:              []Stream{{Name: "file", Description: "Local file stream from configured path."}},
		SyncModes:            allSyncModes(),
		SourceSyncModes:      readSourceSyncModes(),
		DestinationSyncModes: warehouseDestinationSyncModes(),
		Risk: RiskSpec{
			Read:     "local file read",
			Write:    "unsupported",
			Mutation: "none",
			Approval: "not required for reads",
		},
	}
}

func (w Warehouse) Manifest() Manifest {
	return Manifest{
		Metadata: w.Metadata(),
		ConfigFields: []ConfigField{
			{Name: "path", Description: "Local warehouse directory.", Required: false},
		},
		Streams:              []Stream{{Name: "tables", Description: "Local Parquet warehouse tables."}},
		SyncModes:            allSyncModes(),
		SourceSyncModes:      readSourceSyncModes(),
		DestinationSyncModes: warehouseDestinationSyncModes(),
		Risk: RiskSpec{
			Read:     "local warehouse read",
			Write:    "local file write",
			Mutation: "local dependency-free warehouse writes",
			Approval: "not required for ETL destination writes; reverse ETL still requires approval",
		},
	}
}

func (o Outbox) Manifest() Manifest {
	return Manifest{
		Metadata: o.Metadata(),
		ConfigFields: []ConfigField{
			{Name: "path", Description: "Local outbox directory.", Required: false},
		},
		Streams: []Stream{{Name: "records", Description: "Reverse ETL outbox records."}},
		Risk: RiskSpec{
			Read:     "unsupported",
			Write:    "local file write",
			Mutation: "reverse ETL receipt writes",
			Approval: "reverse ETL plan approval required before writes",
		},
	}
}
