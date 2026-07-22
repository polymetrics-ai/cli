package engine

import (
	"context"
	"encoding/json"

	"polymetrics.ai/internal/connectors"
)

// syncMode name constants mirror internal/app/sync_modes.go's MustSyncModeNames
// verbatim (design §B.6). engine cannot import internal/app (app already
// depends on connectors, and PLAN.md forbids editing internal/app/** for this
// task); these are the same five strings, kept in lockstep by
// TestDerivedSyncModesTruthTable against the design doc's truth table rather
// than by a shared import.
const (
	syncModeFullRefreshAppend           = "full_refresh_append"
	syncModeFullRefreshOverwrite        = "full_refresh_overwrite"
	syncModeFullRefreshOverwriteDeduped = "full_refresh_overwrite_deduped"
	syncModeIncrementalAppend           = "incremental_append"
	syncModeIncrementalAppendDeduped    = "incremental_append_deduped"
)

// DerivedSyncModes returns the sync modes a stream supports, derived from its
// bundle-declared shape rather than authored anywhere (design §B.6): the two
// full_refresh modes are always available; the *_deduped variants require
// x-primary-key; the incremental_* modes require an incremental block.
func DerivedSyncModes(s StreamSpec, sch *StreamSchema) []string {
	modes := []string{syncModeFullRefreshAppend, syncModeFullRefreshOverwrite}

	hasPrimaryKey := sch != nil && len(sch.PrimaryKey) > 0
	hasIncremental := s.Incremental != nil

	if hasPrimaryKey {
		modes = append(modes, syncModeFullRefreshOverwriteDeduped)
	}
	if hasIncremental {
		modes = append(modes, syncModeIncrementalAppend)
	}
	if hasPrimaryKey && hasIncremental {
		modes = append(modes, syncModeIncrementalAppendDeduped)
	}
	return modes
}

// Connector adapts a declarative Bundle (+ optional Tier-2 Hooks) to
// connectors.Connector and its optional interfaces. Every method is a thin
// wrapper over the package-level engine functions in read.go/write.go — no
// read/write/check logic is reimplemented here (per the wave0 handoff note).
type Connector struct {
	bundle Bundle
	hooks  Hooks
}

// New returns an engine-backed connectors.Connector for bundle b, dispatching
// to Tier-2 hooks h at their declared extension points (h may be nil when the
// bundle needs no hooks).
func New(b Bundle, h Hooks) *Connector {
	return &Connector{bundle: b, hooks: h}
}

func (c *Connector) Name() string { return c.bundle.Name }

// Metadata synthesizes connectors.Metadata from the bundle's metadata.json,
// matching the legacy hand-written Metadata() shape field-for-field (design
// §C, API-CONTRACT.md §1).
func (c *Connector) Metadata() connectors.Metadata {
	return synthesizeMetadata(c.bundle)
}

// Manifest synthesizes connectors.Manifest from the bundle: streams (with
// derived PK/cursor/sync-modes), write actions, and risk. This is what makes
// engine-backed connectors show up correctly in connectors.ManifestOf without
// any per-connector manifest.go (unlike legacy connectors such as stripe).
func (c *Connector) Manifest() connectors.Manifest {
	return synthesizeManifest(c.bundle)
}

// Definition synthesizes connectors.Definition from the bundle (design §C.1,
// the wave6 target shape already available today via DefinitionProvider).
func (c *Connector) Definition() connectors.Definition {
	return synthesizeDefinition(c.bundle)
}

func (c *Connector) CommandSurface() *connectors.CommandSurface {
	return synthesizeCommandSurface(c.bundle)
}

func (c *Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return Check(ctx, c.bundle, cfg, c.hooks)
}

// Catalog derives connectors.Catalog from the bundle's streams and compiled
// schemas — no network call (matches the "static" shape of legacy Catalog()
// implementations that don't need Check to run first, e.g. stripe/searxng).
func (c *Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(c.bundle.Streams))
	for _, s := range c.bundle.Streams {
		streams = append(streams, legacyStreamOf(c.bundle, s))
	}
	return connectors.Catalog{Connector: c.bundle.Name, Streams: streams}, nil
}

func (c *Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return Read(ctx, c.bundle, req, c.hooks, emit)
}

func (c *Connector) DirectRead(ctx context.Context, req connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	return DirectRead(ctx, c.bundle, req, c.hooks)
}

func (c *Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	return OperationDirectRead(ctx, c.bundle, req, c.hooks)
}

// InitialState satisfies connectors.StatefulReader by delegating to the
// package-level engine.InitialState.
func (c *Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	return InitialState(ctx, c.bundle, stream, cfg)
}

// Write executes req against the bundle's writes.json actions. A bundle with
// no writes.json declared at all (c.bundle.Writes is nil) returns
// connectors.ErrUnsupportedOperation, matching every other read-only builtin
// connector (Sample, File) rather than surfacing write.go's
// action-not-found error, which is reserved for "writes.json exists but this
// action name is wrong."
func (c *Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if len(c.bundle.Writes) == 0 {
		return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
	}
	return Write(ctx, c.bundle, req, records, c.hooks)
}

// ValidateWrite satisfies connectors.WriteValidator.
func (c *Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if len(c.bundle.Writes) == 0 {
		return connectors.ErrUnsupportedOperation
	}
	return ValidateWrite(ctx, c.bundle, req, records)
}

// DryRunWrite satisfies connectors.DryRunWriter.
func (c *Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if len(c.bundle.Writes) == 0 {
		return connectors.WritePreview{}, connectors.ErrUnsupportedOperation
	}
	return DryRunWrite(ctx, c.bundle, req, records, c.hooks)
}

// Base is embedded by Tier-3 native connectors (design §B.7 Tier 3, e.g.
// native/postgres) to serve identity/metadata/definition from their bundle
// without duplicating the synthesis logic Connector already has. Tier-3
// connectors are NOT declaratively read/written by the engine — they
// implement Check/Catalog/Read/Write themselves and embed Base purely for
// Name/Metadata/Definition.
type Base struct {
	bundle Bundle
}

// NewBase returns a Base backed by bundle b.
func NewBase(b Bundle) Base {
	return Base{bundle: b}
}

func (b Base) Name() string { return b.bundle.Name }

func (b Base) Metadata() connectors.Metadata {
	return synthesizeMetadata(b.bundle)
}

func (b Base) Definition() connectors.Definition {
	return synthesizeDefinition(b.bundle)
}

func (b Base) CommandSurface() *connectors.CommandSurface {
	return synthesizeCommandSurface(b.bundle)
}

// synthesizeMetadata is the single source of truth for bundle -> Metadata,
// shared by Connector.Metadata and Base.Metadata so the two views never
// drift from each other.
func synthesizeMetadata(b Bundle) connectors.Metadata {
	return connectors.Metadata{
		Name:            b.Metadata.Name,
		DisplayName:     b.Metadata.DisplayName,
		IntegrationType: b.Metadata.IntegrationType,
		Description:     b.Metadata.Description,
		Capabilities: connectors.Capabilities{
			Check:   b.Metadata.Capabilities.Check,
			Catalog: true,
			Read:    b.Metadata.Capabilities.Read,
			Write:   b.Metadata.Capabilities.Write,
			Query:   b.Metadata.Capabilities.Query,
		},
	}
}

// synthesizeManifest builds connectors.Manifest from the bundle: legacy
// Stream shape (Fields/PrimaryKey/CursorFields) plus the union of every
// stream's derived sync modes (design §B.6) for Manifest.SyncModes.
func synthesizeManifest(b Bundle) connectors.Manifest {
	streams := make([]connectors.Stream, 0, len(b.Streams))
	writeActions := make([]connectors.WriteActionSpec, 0, len(b.Writes))
	configFields := []connectors.ConfigField{}
	secretFields := []connectors.SecretField{}
	modeSet := map[string]bool{}
	var syncModes []string

	if b.Spec != nil {
		secretSet := map[string]bool{}
		for _, key := range b.Spec.SecretKeys() {
			secretSet[key] = true
		}
		for _, property := range b.Spec.Properties() {
			if secretSet[property] {
				secretFields = append(secretFields, connectors.SecretField{Name: property})
				continue
			}
			configFields = append(configFields, connectors.ConfigField{Name: property})
		}
	}

	for _, s := range b.Streams {
		streams = append(streams, legacyStreamOf(b, s))
		for _, mode := range DerivedSyncModes(s, b.Schemas[s.Name]) {
			if !modeSet[mode] {
				modeSet[mode] = true
				syncModes = append(syncModes, mode)
			}
		}
	}
	// Preserve the canonical mode ordering (matches
	// internal/app/sync_modes.go.MustSyncModeNames) rather than
	// first-seen-per-stream order.
	syncModes = orderCanonicalModes(syncModes)

	for _, a := range b.Writes {
		writeActions = append(writeActions, connectors.WriteActionSpec{
			Name:           a.Name,
			RequiredFields: writeActionRequiredFields(a),
			OptionalFields: writeActionOptionalFields(a),
			Method:         a.Method,
			Path:           a.Path,
			Risk:           a.Risk,
			Confirm:        a.Confirm,
		})
	}

	return connectors.Manifest{
		Metadata:     synthesizeMetadata(b),
		ConfigFields: configFields,
		SecretFields: secretFields,
		Streams:      streams,
		WriteActions: writeActions,
		SyncModes:    syncModes,
		Risk: connectors.RiskSpec{
			Read:     b.Metadata.Risk.Read,
			Write:    b.Metadata.Risk.Write,
			Approval: b.Metadata.Risk.Approval,
		},
	}
}

func writeActionRequiredFields(action WriteAction) []string {
	fields := appendUniqueStrings(nil, action.PathFields...)
	var schema struct {
		Required []string `json:"required"`
	}
	if len(action.RecordSchema) > 0 && json.Unmarshal(action.RecordSchema, &schema) == nil {
		fields = appendUniqueStrings(fields, schema.Required...)
	}
	return fields
}

func writeActionOptionalFields(action WriteAction) []string {
	required := map[string]bool{}
	for _, field := range writeActionRequiredFields(action) {
		required[field] = true
	}
	var optional []string
	for _, field := range action.BodyFields {
		if !required[field] {
			optional = appendUniqueStrings(optional, field)
		}
	}
	return optional
}

func appendUniqueStrings(base []string, values ...string) []string {
	seen := map[string]bool{}
	for _, value := range base {
		seen[value] = true
	}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		base = append(base, value)
	}
	return base
}

// synthesizeDefinition builds connectors.Definition from the bundle (design
// §C.1). Spec is the compiled spec.json's original bytes when available; a
// bundle built ad hoc in tests without a Spec gets an empty JSON object
// rather than a nil/invalid RawMessage.
func synthesizeDefinition(b Bundle) connectors.Definition {
	streamSummaries := make([]connectors.StreamSummary, 0, len(b.Streams))
	for _, s := range b.Streams {
		sch := b.Schemas[s.Name]
		summary := connectors.StreamSummary{
			Name:      s.Name,
			SyncModes: DerivedSyncModes(s, sch),
		}
		if sch != nil {
			summary.PrimaryKey = sch.PrimaryKey
			summary.CursorField = sch.CursorField
		}
		streamSummaries = append(streamSummaries, summary)
	}

	writeActions := make([]connectors.WriteActionInfo, 0, len(b.Writes))
	for _, a := range b.Writes {
		writeActions = append(writeActions, connectors.WriteActionInfo{
			Name:    a.Name,
			Kind:    a.Kind,
			Method:  a.Method,
			Path:    a.Path,
			Risk:    a.Risk,
			Confirm: a.Confirm,
		})
	}

	return connectors.Definition{
		Name:            b.Metadata.Name,
		DisplayName:     b.Metadata.DisplayName,
		Description:     b.Metadata.Description,
		IntegrationType: b.Metadata.IntegrationType,
		DocsURL:         b.Metadata.DocsURL,
		ReleaseStage:    b.Metadata.ReleaseStage,
		Capabilities:    synthesizeMetadata(b).Capabilities,
		Spec:            specJSON(b),
		Streams:         streamSummaries,
		WriteActions:    writeActions,
		Risk: connectors.RiskSpec{
			Read:     b.Metadata.Risk.Read,
			Write:    b.Metadata.Risk.Write,
			Approval: b.Metadata.Risk.Approval,
		},
	}
}

func synthesizeCommandSurface(b Bundle) *connectors.CommandSurface {
	if b.CLISurface == nil {
		return nil
	}
	surface := b.CLISurface
	out := &connectors.CommandSurface{
		Tagline:     surface.Tagline,
		Usage:       surface.Usage,
		Groups:      make([]connectors.CommandSurfaceGroup, 0, len(surface.Groups)),
		GlobalFlags: make([]connectors.CommandSurfaceFlag, 0, len(surface.GlobalFlags)),
		Commands:    make([]connectors.CommandSurfaceCommand, 0, len(surface.Commands)),
		HelpTopics:  make([]connectors.CommandSurfaceHelpTopic, 0, len(surface.HelpTopics)),
	}
	if surface.SourceCLI != nil {
		out.SourceCLI = &connectors.CommandSurfaceSource{
			Name:      surface.SourceCLI.Name,
			Docs:      surface.SourceCLI.Docs,
			Reference: surface.SourceCLI.Reference,
			Source:    surface.SourceCLI.Source,
		}
	}
	for _, group := range surface.Groups {
		out.Groups = append(out.Groups, connectors.CommandSurfaceGroup{
			ID:       group.ID,
			Title:    group.Title,
			Commands: append([]string(nil), group.Commands...),
		})
	}
	for _, flag := range surface.GlobalFlags {
		out.GlobalFlags = append(out.GlobalFlags, commandSurfaceFlag(flag))
	}
	for _, cmd := range surface.Commands {
		flags := make([]connectors.CommandSurfaceFlag, 0, len(cmd.Flags))
		for _, flag := range cmd.Flags {
			flags = append(flags, commandSurfaceFlag(flag))
		}
		out.Commands = append(out.Commands, connectors.CommandSurfaceCommand{
			Path:          cmd.Path,
			Summary:       cmd.Summary,
			Intent:        cmd.Intent,
			Availability:  cmd.Availability,
			Stream:        cmd.Stream,
			Write:         cmd.Write,
			Operation:     cmd.Operation,
			SourceCLIPath: cmd.SourceCLIPath,
			SourceURL:     cmd.SourceURL,
			Flags:         flags,
			Examples:      append([]string(nil), cmd.Examples...),
			APISurface:    commandSurfaceEndpointRefs(cmd.APISurface),
			OutputPolicy:  cmd.OutputPolicy,
			Risk:          cmd.Risk,
			Approval:      cmd.Approval,
			Notes:         cmd.Notes,
		})
	}
	for _, topic := range surface.HelpTopics {
		out.HelpTopics = append(out.HelpTopics, connectors.CommandSurfaceHelpTopic{
			Name:    topic.Name,
			Summary: topic.Summary,
		})
	}
	return out
}

func commandSurfaceEndpointRefs(refs []CLISurfaceEndpointRef) []connectors.CommandSurfaceEndpointRef {
	out := make([]connectors.CommandSurfaceEndpointRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, connectors.CommandSurfaceEndpointRef{Method: ref.Method, Path: ref.Path})
	}
	return out
}

func commandSurfaceFlag(flag CLIFlag) connectors.CommandSurfaceFlag {
	return connectors.CommandSurfaceFlag{
		Name:    flag.Name,
		Type:    flag.Type,
		Summary: flag.Summary,
		Values:  append([]string(nil), flag.Values...),
		MapsTo:  flag.MapsTo,
	}
}

// specJSON returns the bundle's spec.json VERBATIM (F5, REVIEW.md fix): a
// bundle loaded via Load/LoadAll always has RawSpec populated with the exact
// bytes read from disk, so Definition.Spec now serves types, enums,
// defaults, required, and descriptions byte-for-byte instead of a lossy
// reconstruction. A bundle with no RawSpec at all (an ad hoc test bundle
// that only ever set Spec directly, never went through Load) falls back to
// reconstructing a JSON object from the compiled *Schema's Properties()/
// SecretKeys() — every property flattened to {"type":"string"} (+x-secret) —
// preserving prior behavior for that construction path. A bundle with
// NEITHER RawSpec NOR Spec gets an empty JSON object.
func specJSON(b Bundle) []byte {
	if len(b.RawSpec) > 0 {
		return []byte(b.RawSpec)
	}
	if b.Spec == nil {
		return []byte("{}")
	}
	secrets := make(map[string]bool, len(b.Spec.SecretKeys()))
	for _, name := range b.Spec.SecretKeys() {
		secrets[name] = true
	}
	properties := make(map[string]any, len(b.Spec.Properties()))
	for _, name := range b.Spec.Properties() {
		prop := map[string]any{"type": "string"}
		if secrets[name] {
			prop["x-secret"] = true
		}
		properties[name] = prop
	}
	doc := map[string]any{"type": "object", "properties": properties}
	raw, err := json.Marshal(doc)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

// legacyStreamOf builds the legacy connectors.Stream shape for one
// StreamSpec, reusing its compiled schema for PrimaryKey/CursorFields. Field
// types are not derivable from the compiled schema (Schema only exposes
// property names, not per-property JSON types), so Fields carries names only
// (Type left as the zero value) — no wave0 consumer inspects Field.Type.
func legacyStreamOf(b Bundle, s StreamSpec) connectors.Stream {
	sch := b.Schemas[s.Name]
	stream := connectors.Stream{Name: s.Name}
	if sch == nil {
		return stream
	}
	stream.PrimaryKey = sch.PrimaryKey
	if sch.CursorField != "" {
		stream.CursorFields = []string{sch.CursorField}
	}
	for _, name := range sch.Properties() {
		stream.Fields = append(stream.Fields, connectors.Field{Name: name})
	}
	return stream
}

// canonicalModeOrder is internal/app/sync_modes.go's MustSyncModeNames order.
var canonicalModeOrder = []string{
	syncModeFullRefreshAppend,
	syncModeFullRefreshOverwrite,
	syncModeFullRefreshOverwriteDeduped,
	syncModeIncrementalAppend,
	syncModeIncrementalAppendDeduped,
}

// orderCanonicalModes returns the subset of canonicalModeOrder present in
// modes, in canonical order.
func orderCanonicalModes(modes []string) []string {
	present := make(map[string]bool, len(modes))
	for _, m := range modes {
		present[m] = true
	}
	out := make([]string, 0, len(modes))
	for _, m := range canonicalModeOrder {
		if present[m] {
			out = append(out, m)
		}
	}
	return out
}
