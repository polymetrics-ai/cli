package connectors

import "context"

type ConfigField struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
}

type SecretField struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
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
	OptionalFields []string `json:"optional_fields,omitempty"`
	Method         string   `json:"method,omitempty"`
	Path           string   `json:"path,omitempty"`
	Risk           string   `json:"risk,omitempty"`
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
	if provider, ok := c.(ManifestProvider); ok {
		return manifestWithIcon(provider.Manifest())
	}
	return manifestWithIcon(Manifest{
		Metadata: c.Metadata(),
		Risk: RiskSpec{
			Read:     "connector-specific",
			Write:    "connector-specific",
			Approval: "external mutations require preview and approval",
		},
	})
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
	return []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	}
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
		Streams:              []Stream{{Name: "tables", Description: "Local JSONL warehouse tables."}},
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
