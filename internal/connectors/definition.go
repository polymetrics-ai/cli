package connectors

import (
	"encoding/json"

	"polymetrics.ai/internal/synccontract"
)

// Definition is the unified connector descriptor introduced by architecture
// v2 (design doc §C.1). In wave0 it coexists with Metadata/Manifest, added
// purely additively alongside the existing Connector/ManifestProvider
// interfaces; wave6 folds Metadata/ManifestProvider into it and joins
// Definition() to the core Connector interface. See API-CONTRACT.md §1.
type Definition struct {
	Name             string                      `json:"name"`
	DisplayName      string                      `json:"display_name"`
	Description      string                      `json:"description,omitempty"`
	IntegrationType  string                      `json:"integration_type"`
	DocsURL          string                      `json:"docs_url,omitempty"`
	ReleaseStage     string                      `json:"release_stage"`
	Capabilities     Capabilities                `json:"capabilities"`
	Changefeed       *ChangefeedDescriptor       `json:"changefeed,omitempty"`
	PollingWatermark *PollingWatermarkDescriptor `json:"polling_watermark,omitempty"`
	SyncTransport    *SyncTransportDescriptor    `json:"sync_transport,omitempty"`
	EnabledContract  *EnabledConnectorContract   `json:"enabled_connector_contract,omitempty"`
	Spec             json.RawMessage             `json:"spec"`
	Streams          []StreamSummary             `json:"streams"`
	WriteActions     []WriteActionInfo           `json:"write_actions,omitempty"`
	Risk             RiskSpec                    `json:"risk"`
	Icon             *ConnectorIcon              `json:"icon,omitempty"`
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
	Name             string                  `json:"name"`
	Kind             string                  `json:"kind"`
	Method           string                  `json:"method"`
	Path             string                  `json:"path"`
	Risk             string                  `json:"risk"`
	TransportBinding *TransportActionBinding `json:"transport_binding,omitempty"`
	// Batchable mirrors the bundle's "batchable" declaration; nil means
	// undeclared and therefore batchable. See WriteActionSpec.IsBatchable.
	Batchable *bool  `json:"batchable,omitempty"`
	Confirm   string `json:"confirm,omitempty"`
}

const (
	TransportCapabilityIssueLabel = "issue_label"
	TransportActionRoleApply      = "apply"
	TransportActionRoleReplace    = "replace"
	TransportActionRoleCleanup    = "cleanup"
	TransportInputTargetIssue     = "target_issue"
	TransportInputLabel           = "label"
	TransportInputShapeScalar     = "scalar"
	TransportInputShapeList       = "singleton_array"
)

// TransportActionBinding declares how a closed transport capability selects a
// write action and constructs its provider record.
type TransportActionBinding struct {
	Capability string                  `json:"capability"`
	Role       string                  `json:"role"`
	Modes      []synccontract.Mode     `json:"modes"`
	Inputs     []TransportInputBinding `json:"inputs"`
}

// TransportInputBinding maps one typed transport input into an action record.
type TransportInputBinding struct {
	Input string `json:"input"`
	Field string `json:"field"`
	Shape string `json:"shape"`
}

// Clone returns an independent copy of b.
func (b *TransportActionBinding) Clone() *TransportActionBinding {
	if b == nil {
		return nil
	}
	copied := *b
	copied.Modes = append([]synccontract.Mode(nil), b.Modes...)
	copied.Inputs = append([]TransportInputBinding(nil), b.Inputs...)
	return &copied
}

// IsBatchable reports whether the action may run from a bulk reverse ETL plan.
func (i WriteActionInfo) IsBatchable() bool {
	return i.Batchable == nil || *i.Batchable
}

// DefinitionProvider is implemented by engine-backed and Tier-3 connectors in
// wave0; the method joins the core Connector interface in wave6 (design
// §C.1). Callers that need a Definition today (e.g. certify and connector
// inspection JSON) must type-assert for this interface rather than assume every
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
	if def.Changefeed != nil {
		def.Changefeed = def.Changefeed.Clone()
	}
	if def.PollingWatermark != nil {
		def.PollingWatermark = def.PollingWatermark.Clone()
	}
	if def.SyncTransport != nil {
		def.SyncTransport = def.SyncTransport.Clone()
	}
	if def.EnabledContract != nil {
		def.EnabledContract = def.EnabledContract.Clone()
	}
	def.WriteActions = cloneWriteActionInfos(def.WriteActions)
	def.Capabilities.CDC = HasImplementedChangefeed(c, def.Changefeed)
	def.Icon = MetadataWithIcon(c.Metadata()).Icon
	return def, true
}

// MetadataOf returns the public metadata projection for c. CDC is derived
// from the same descriptor/executor rule as DefinitionOf; legacy metadata and
// CDCReader method presence never make it true on their own.
func MetadataOf(c Connector) Metadata {
	metadata := MetadataWithIcon(c.Metadata())
	if def, ok := DefinitionOf(c); ok {
		metadata.Capabilities.CDC = def.Capabilities.CDC
	} else {
		metadata.Capabilities.CDC = false
	}
	return metadata
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

// cloneBoolPtr copies the pointed-to value so a projection cannot alias — and
// therefore mutate — the source spec's policy flag.
func cloneBoolPtr(v *bool) *bool {
	if v == nil {
		return nil
	}
	copied := *v
	return &copied
}

func cloneWriteActionInfos(actions []WriteActionInfo) []WriteActionInfo {
	if actions == nil {
		return nil
	}
	copied := make([]WriteActionInfo, len(actions))
	for i, action := range actions {
		copied[i] = action
		copied[i].Batchable = cloneBoolPtr(action.Batchable)
		copied[i].TransportBinding = action.TransportBinding.Clone()
	}
	return copied
}

func writeActionInfosFromManifest(manifest Manifest) []WriteActionInfo {
	out := make([]WriteActionInfo, 0, len(manifest.WriteActions))
	for _, action := range manifest.WriteActions {
		out = append(out, WriteActionInfo{
			Name:      action.Name,
			Method:    action.Method,
			Path:      action.Path,
			Risk:      action.Risk,
			Batchable: cloneBoolPtr(action.Batchable),
		})
	}
	return out
}
