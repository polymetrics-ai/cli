package connectors

import "encoding/json"

// Definition is the unified connector descriptor introduced by architecture
// v2 (design doc §C.1). In wave0 it coexists with Metadata/Manifest, added
// purely additively alongside the existing Connector/ManifestProvider
// interfaces; wave6 folds Metadata/ManifestProvider into it and joins
// Definition() to the core Connector interface. See API-CONTRACT.md §1.
type Definition struct {
	Name            string            `json:"name"`
	DisplayName     string            `json:"display_name"`
	Description     string            `json:"description,omitempty"`
	IntegrationType string            `json:"integration_type"`
	DocsURL         string            `json:"docs_url,omitempty"`
	ReleaseStage    string            `json:"release_stage"`
	Capabilities    Capabilities      `json:"capabilities"`
	Spec            json.RawMessage   `json:"spec"`
	Streams         []StreamSummary   `json:"streams"`
	WriteActions    []WriteActionInfo `json:"write_actions,omitempty"`
	Risk            RiskSpec          `json:"risk"`
	Icon            *ConnectorIcon    `json:"icon,omitempty"`
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

// DefinitionProvider is implemented by engine-backed and Tier-3 connectors in
// wave0; the method joins the core Connector interface in wave6 (design
// §C.1). Callers that need a Definition today (e.g. certify, a future CLI
// surface) must type-assert for this interface rather than assume every
// Connector implements it, mirroring the existing ManifestProvider pattern.
type DefinitionProvider interface {
	Definition() Definition
}
