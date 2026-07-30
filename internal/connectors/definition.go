package connectors

import "encoding/json"

// Definition is the unified connector descriptor introduced by architecture
// v2 (design doc §C.1). In wave0 it coexists with Metadata/Manifest, added
// purely additively alongside the existing Connector/ManifestProvider
// interfaces; wave6 folds Metadata/ManifestProvider into it and joins
// Definition() to the core Connector interface. See API-CONTRACT.md §1.
type Definition struct {
	Name               string                  `json:"name"`
	DisplayName        string                  `json:"display_name"`
	Description        string                  `json:"description,omitempty"`
	IntegrationType    string                  `json:"integration_type"`
	DocsURL            string                  `json:"docs_url,omitempty"`
	ReleaseStage       string                  `json:"release_stage"`
	Capabilities       Capabilities            `json:"capabilities"`
	Spec               json.RawMessage         `json:"spec"`
	Streams            []StreamSummary         `json:"streams"`
	WriteActions       []WriteActionInfo       `json:"write_actions,omitempty"`
	ProviderOperations []ProviderOperationInfo `json:"provider_operations,omitempty"`
	Risk               RiskSpec                `json:"risk"`
	Icon               *ConnectorIcon          `json:"icon,omitempty"`
}

// StreamSummary is one Definition.Streams entry. SyncModes is always DERIVED
// (design §B.6, engine.DerivedSyncModes) — it is never hand-authored in a
// bundle and must never be trusted as an independent source of truth.
type StreamSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	PrimaryKey  []string `json:"primary_key,omitempty"`
	CursorField string   `json:"cursor_field,omitempty"`
	SyncModes   []string `json:"sync_modes"`
}

// WriteActionInfo is one Definition.WriteActions entry.
type WriteActionInfo struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Risk    string `json:"risk"`
	Confirm string `json:"confirm,omitempty"`
}

// ProviderOperationInfo is a serialized, metadata-only description of a
// provider-native search/query operation. It carries enough typed contract data
// for docs, inspection, conformance fixtures, and future executor selection
// without exposing a raw SQL/GraphQL/HTTP escape hatch.
type ProviderOperationInfo struct {
	ID             string                  `json:"id"`
	Kind           string                  `json:"kind"`
	Summary        string                  `json:"summary"`
	Description    string                  `json:"description,omitempty"`
	Risk           string                  `json:"risk,omitempty"`
	Approval       string                  `json:"approval,omitempty"`
	OutputPolicy   string                  `json:"output_policy"`
	RequestSchema  json.RawMessage         `json:"request_schema"`
	ResponseSchema json.RawMessage         `json:"response_schema"`
	Bounds         ProviderOperationBounds `json:"bounds"`
	Pagination     *ProviderPaginationInfo `json:"pagination,omitempty"`
	Fixture        *ProviderFixtureInfo    `json:"fixture,omitempty"`
}

type ProviderOperationBounds struct {
	DefaultLimit int `json:"default_limit"`
	MaxLimit     int `json:"max_limit"`
	MaxPages     int `json:"max_pages"`
	MaxBytes     int `json:"max_bytes"`
}

type ProviderPaginationInfo struct {
	Type                string `json:"type"`
	CursorRequestField  string `json:"cursor_request_field,omitempty"`
	CursorResponseField string `json:"cursor_response_field,omitempty"`
	PageRequestField    string `json:"page_request_field,omitempty"`
	PageSizeField       string `json:"page_size_field,omitempty"`
	OffsetRequestField  string `json:"offset_request_field,omitempty"`
	LimitRequestField   string `json:"limit_request_field,omitempty"`
	ItemsResponseField  string `json:"items_response_field,omitempty"`
	HasMoreField        string `json:"has_more_field,omitempty"`
}

type ProviderFixtureInfo struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

// DefinitionProvider is implemented by engine-backed and Tier-3 connectors in
// wave0; the method joins the core Connector interface in wave6 (design
// §C.1). Callers that need a Definition today (e.g. certify, a future CLI
// surface) must type-assert for this interface rather than assume every
// Connector implements it, mirroring the existing ManifestProvider pattern.
type DefinitionProvider interface {
	Definition() Definition
}

func DefinitionOf(c Connector) (Definition, bool) {
	provider, ok := c.(DefinitionProvider)
	if !ok {
		return Definition{}, false
	}
	def := provider.Definition()
	def.Icon = MetadataWithIcon(c.Metadata()).Icon
	return def, true
}

func streamSummariesFromManifest(manifest Manifest) []StreamSummary {
	out := make([]StreamSummary, 0, len(manifest.Streams))
	for _, stream := range manifest.Streams {
		summary := StreamSummary{
			Name:       stream.Name,
			PrimaryKey: stream.PrimaryKey,
			SyncModes:  append([]string(nil), manifest.SyncModes...),
		}
		if len(stream.CursorFields) > 0 {
			summary.CursorField = stream.CursorFields[0]
		}
		out = append(out, summary)
	}
	return out
}

func writeActionInfosFromManifest(manifest Manifest) []WriteActionInfo {
	out := make([]WriteActionInfo, 0, len(manifest.WriteActions))
	for _, action := range manifest.WriteActions {
		out = append(out, WriteActionInfo{
			Name:   action.Name,
			Method: action.Method,
			Path:   action.Path,
			Risk:   action.Risk,
		})
	}
	return out
}
