package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

// --- compile-time interface assertions (design §B.7, API-CONTRACT.md §2) ---

func TestCatalogStaticSchemaMatchesDiscoveredSchemaProjection(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	staticCatalog, err := New(bundle, nil).Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("static Catalog: %v", err)
	}
	if len(staticCatalog.Streams) != 1 {
		t.Fatalf("static Catalog streams = %d, want 1", len(staticCatalog.Streams))
	}

	discovered, err := connectors.StreamFromSchema("widgets", "", fsys["acme/schemas/widgets.json"].Data)
	if err != nil {
		t.Fatalf("StreamFromSchema: %v", err)
	}
	if !reflect.DeepEqual(staticCatalog.Streams[0], discovered) {
		t.Fatalf("static stream = %#v, discovered stream = %#v; catalog consumers must not need separate paths", staticCatalog.Streams[0], discovered)
	}
}

var (
	_ connectors.Connector                   = (*Connector)(nil)
	_ connectors.WriteValidator              = (*Connector)(nil)
	_ connectors.DeclarativeTypedDestination = (*Connector)(nil)
	_ connectors.DryRunWriter                = (*Connector)(nil)
	_ connectors.StatefulReader              = (*Connector)(nil)
	_ connectors.ManifestProvider            = (*Connector)(nil)
	_ connectors.DefinitionProvider          = (*Connector)(nil)
)

func TestDeclarativeTypedDestinationActionDigestIsCanonicalAndDefinitionBound(t *testing.T) {
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://api.example.test", Headers: map[string]string{"X-Provider": "acme"}},
		Writes: []WriteAction{{
			Name: "apply_widget", Method: http.MethodPost, Path: "/widgets", BodyType: "json",
			RecordSchema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"value":{"type":"string"}}}`),
		}},
	}
	first, err := declarativeTypedDestinationActionDigest(bundle, "apply_widget")
	if err != nil {
		t.Fatalf("first action digest: %v", err)
	}
	bundle.Writes[0].RecordSchema = json.RawMessage(`{ "properties": { "value": { "type": "string" }, "id": { "type": "string" } }, "type": "object" }`)
	second, err := declarativeTypedDestinationActionDigest(bundle, "apply_widget")
	if err != nil {
		t.Fatalf("canonical action digest: %v", err)
	}
	if first != second {
		t.Fatalf("equivalent action definitions changed digest: first=%q second=%q", first, second)
	}
	bundle.Writes[0].Path = "/widgets/changed"
	third, err := declarativeTypedDestinationActionDigest(bundle, "apply_widget")
	if err != nil {
		t.Fatalf("changed action digest: %v", err)
	}
	if third == second {
		t.Fatal("changed action definition preserved digest")
	}
}

func TestDeclarativeTypedDestinationActionDigestIncludesEffectiveHTTPDefaults(t *testing.T) {
	compile := func(t *testing.T, baseURL, header string) *Schema {
		t.Helper()
		spec, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"properties":{
				"base_url":{"type":"string","default":` + mustJSONLiteral(t, baseURL) + `},
				"provider_header":{"type":"string","default":` + mustJSONLiteral(t, header) + `}
			}
		}`))
		if err != nil {
			t.Fatalf("CompileSchema: %v", err)
		}
		return spec
	}
	bundle := Bundle{
		Name: "acme",
		Spec: compile(t, "https://api-one.example.test", "one"),
		HTTP: HTTPBase{
			URL: "{{ config.base_url }}", UserAgent: "acme-agent-one",
			Headers: map[string]string{"X-Provider": "{{ config.provider_header }}"},
		},
		Writes: []WriteAction{{
			Name: "apply_widget", Method: http.MethodPost, Path: "/widgets", BodyType: "json",
			RecordSchema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
		}},
	}
	baseline, err := declarativeTypedDestinationActionDigest(bundle, "apply_widget")
	if err != nil {
		t.Fatalf("baseline action digest: %v", err)
	}
	changedAgent := bundle
	changedAgent.HTTP.UserAgent = "acme-agent-two"
	agentDigest, err := declarativeTypedDestinationActionDigest(changedAgent, "apply_widget")
	if err != nil {
		t.Fatalf("user-agent action digest: %v", err)
	}
	if agentDigest == baseline {
		t.Fatal("changed HTTP user agent preserved action digest")
	}
	changedDefaults := bundle
	changedDefaults.Spec = compile(t, "https://api-two.example.test", "two")
	defaultDigest, err := declarativeTypedDestinationActionDigest(changedDefaults, "apply_widget")
	if err != nil {
		t.Fatalf("default action digest: %v", err)
	}
	if defaultDigest == baseline {
		t.Fatal("changed request defaults preserved action digest")
	}
}

func mustJSONLiteral(t *testing.T, value string) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(raw)
}

// Base itself is NOT asserted against connectors.Connector or
// connectors.ManifestProvider: per API-CONTRACT.md §2, Tier-3 natives that
// embed it supply Check/Catalog/Read/Write themselves and are not required to
// also provide a legacy Manifest(). tier3FakeConnector below is the
// compile-time proof that Base + those four methods together satisfy
// connectors.Connector.
var (
	_ connectors.DefinitionProvider                     = Base{}
	_ connectors.OperationDirectReadPreflighter         = Base{}
	_ connectors.OperationDirectWritePreflighter        = Base{}
	_ connectors.OperationDirectWriteBindingPreflighter = Base{}
	_ connectors.OperationStructuredJSONBodyPreflighter = Base{}
)

// --- test fixtures ---

func widgetsRecordSchema(t *testing.T, primaryKey, cursorField string) *StreamSchema {
	t.Helper()
	extra := ""
	if primaryKey != "" {
		extra += `,"x-primary-key":["` + primaryKey + `"]`
	}
	if cursorField != "" {
		extra += `,"x-cursor-field":"` + cursorField + `"`
	}
	raw := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"updated_at": {"type": "string"}
		}` + extra + `
	}`)
	sch, err := CompileSchema(raw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	return &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}
}

func minimalSpecSchema(t *testing.T) *Schema {
	t.Helper()
	sch, err := CompileSchema([]byte(`{"type":"object","required":["account_id","api_key"],"properties":{"account_id":{"type":"string"},"api_key":{"type":"string","x-secret":true}}}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	return sch
}

func newConnectorTestBundle(t *testing.T, srv *httptest.Server) Bundle {
	t.Helper()
	return Bundle{
		Name: "acme",
		Metadata: Metadata{
			Name:            "acme",
			DisplayName:     "Acme",
			Description:     "Acme widgets API.",
			IntegrationType: "api",
			DocsURL:         "https://example.com/docs",
			ReleaseStage:    "beta",
			Capabilities:    Capabilities{Check: true, Read: true, Write: true},
			Risk: RiskSpec{
				Read:     "read-only REST calls against the configured account",
				Write:    "mutates widget records",
				Approval: "reverse ETL writes require plan preview and approval token",
			},
		},
		Spec: minimalSpecSchema(t),
		HTTP: HTTPBase{URL: srv.URL},
		Streams: []StreamSpec{
			{Name: "widgets", Path: "/widgets", Records: RecordsSpec{Path: "data"}},
		},
		Schemas: map[string]*StreamSchema{
			"widgets": widgetsRecordSchema(t, "id", "updated_at"),
		},
	}
}

// --- Metadata()/Manifest() synthesis ---

func TestConnectorMetadataSynthesizedFromBundle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)

	if got := c.Name(); got != "acme" {
		t.Fatalf("Name() = %q, want acme", got)
	}
	md := c.Metadata()
	if md.Name != "acme" || md.DisplayName != "Acme" {
		t.Fatalf("Metadata() name/display = %q/%q, want acme/Acme", md.Name, md.DisplayName)
	}
	if md.IntegrationType != "api" {
		t.Fatalf("Metadata().IntegrationType = %q, want api", md.IntegrationType)
	}
	if md.Description != "Acme widgets API." {
		t.Fatalf("Metadata().Description = %q, want bundle description", md.Description)
	}
	if !md.Capabilities.Check || !md.Capabilities.Read || !md.Capabilities.Write {
		t.Fatalf("Metadata().Capabilities = %+v, want check/read/write true", md.Capabilities)
	}
}

func TestConnectorManifestSynthesizedFromBundleSpotFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	m := c.Manifest()

	if m.Metadata.Name != "acme" {
		t.Fatalf("Manifest().Metadata.Name = %q, want acme", m.Metadata.Name)
	}
	if len(m.ConfigFields) != 1 || m.ConfigFields[0].Name != "account_id" || !m.ConfigFields[0].Required {
		t.Fatalf("Manifest().ConfigFields = %#v, want required account_id", m.ConfigFields)
	}
	if len(m.SecretFields) != 1 || m.SecretFields[0].Name != "api_key" || !m.SecretFields[0].Required {
		t.Fatalf("Manifest().SecretFields = %#v, want required api_key", m.SecretFields)
	}
	if len(m.Streams) != 1 || m.Streams[0].Name != "widgets" {
		t.Fatalf("Manifest().Streams = %+v, want one stream named widgets", m.Streams)
	}
	if len(m.Streams[0].PrimaryKey) != 1 || m.Streams[0].PrimaryKey[0] != "id" {
		t.Fatalf("Manifest().Streams[0].PrimaryKey = %v, want [id]", m.Streams[0].PrimaryKey)
	}
	if len(m.Streams[0].CursorFields) != 1 || m.Streams[0].CursorFields[0] != "updated_at" {
		t.Fatalf("Manifest().Streams[0].CursorFields = %v, want [updated_at]", m.Streams[0].CursorFields)
	}
	wantModes := []string{"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped"}
	assertStringSliceEqual(t, m.SyncModes, wantModes)
	if m.Risk.Read == "" {
		t.Fatalf("Manifest().Risk.Read is empty, want a synthesized risk string")
	}
}

func TestCommandSurfaceAddsSharedApprovalStdinMarkerForWrites(t *testing.T) {
	for _, tc := range []struct {
		name        string
		globalFlags []CLIFlag
		intent      string
		wantCount   int
		wantSummary string
	}{
		{
			name:        "reverse ETL derives marker",
			intent:      "reverse_etl",
			wantCount:   1,
			wantSummary: "Read the approval token as one bounded line from standard input.",
		},
		{
			name:      "direct write derives marker",
			intent:    "direct_write",
			wantCount: 1,
		},
		{
			name: "declared marker is not duplicated",
			globalFlags: []CLIFlag{{
				Name:    "approval-token-stdin",
				Type:    "boolean",
				Summary: "bundle declaration",
			}},
			intent:      "reverse_etl",
			wantCount:   1,
			wantSummary: "bundle declaration",
		},
		{
			name:      "read command has no marker",
			intent:    "direct_read",
			wantCount: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			surface := synthesizeCommandSurface(Bundle{CLISurface: &CLISurface{
				GlobalFlags: tc.globalFlags,
				Commands:    []CLICommand{{Path: "widgets apply", Intent: tc.intent}},
			}})
			if surface == nil {
				t.Fatal("synthesizeCommandSurface returned nil")
			}
			count := 0
			for _, flag := range surface.GlobalFlags {
				if flag.Name != "approval-token-stdin" {
					continue
				}
				count++
				if tc.wantSummary != "" && flag.Summary != tc.wantSummary {
					t.Fatalf("approval stdin summary = %q, want %q", flag.Summary, tc.wantSummary)
				}
			}
			if count != tc.wantCount {
				t.Fatalf("approval stdin marker count = %d, want %d", count, tc.wantCount)
			}
		})
	}
}

func TestCommandSurfaceProjectsOperationParameterByteCaps(t *testing.T) {
	bundle := Bundle{
		Operations: []OperationSpec{{
			ID: "acme.widgets.get", Kind: "rest_read",
			REST: &RESTOperationSpec{Parameters: []OperationParameter{
				{Name: "id", In: "path", MaxBytes: 128},
				{Name: "filter", In: "query"},
			}},
		}},
		CLISurface: &CLISurface{Commands: []CLICommand{{
			Path: "widgets get", Intent: "direct_read", Operation: "acme.widgets.get",
			Flags: []CLIFlag{
				{Name: "id", Type: "string", MapsTo: "path.id"},
				{Name: "filter", Type: "string", MapsTo: "query.filter"},
				{Name: "body", Type: "string", MapsTo: "body.name"},
			},
		}}},
	}
	surface := synthesizeCommandSurface(bundle)
	if surface == nil || len(surface.Commands) != 1 || len(surface.Commands[0].Flags) != 3 {
		t.Fatalf("command surface = %#v", surface)
	}
	flags := surface.Commands[0].Flags
	if flags[0].MaxBytes != 128 || flags[1].MaxBytes != defaultOperationParameterMaxBytes || flags[2].MaxBytes != 0 {
		t.Fatalf("projected max bytes = [%d %d %d]", flags[0].MaxBytes, flags[1].MaxBytes, flags[2].MaxBytes)
	}
	for index, flag := range flags[:2] {
		if flag.MaxBytesOrigin != "pm_policy" || flag.MaxBytesPolicyVersion != OperationParameterExecutionPolicyVersion {
			t.Fatalf("flag[%d] execution provenance = (%q, %q)", index, flag.MaxBytesOrigin, flag.MaxBytesPolicyVersion)
		}
	}
	if flags[2].MaxBytesOrigin != "" || flags[2].MaxBytesPolicyVersion != "" {
		t.Fatalf("unbounded body flag execution provenance = (%q, %q)", flags[2].MaxBytesOrigin, flags[2].MaxBytesPolicyVersion)
	}
}

func TestCommandSurfaceProjectsDeferredFoundationGap(t *testing.T) {
	surface := synthesizeCommandSurface(Bundle{CLISurface: &CLISurface{Commands: []CLICommand{{
		Path: "widgets archive", Intent: "binary_upload", Availability: "deferred",
		Foundation: &CommandFoundation{ID: "binary_upload_foundation", Reason: "binary request binding is not available"},
	}}}})
	if surface == nil || len(surface.Commands) != 1 {
		t.Fatalf("command surface = %#v", surface)
	}
	got := surface.Commands[0].Foundation
	if got == nil || got.ID != "binary_upload_foundation" || got.Reason != "binary request binding is not available" {
		t.Fatalf("foundation gap = %#v", got)
	}
}

// --- Definition() ---

func TestConnectorDefinitionSynthesizedFromBundle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)
	b.Writes = []WriteAction{
		{Name: "update_widget", Kind: "update", Method: "PATCH", Path: "/widgets/{{ record.id }}", Risk: "medium"},
	}

	c := New(b, nil)
	def := c.Definition()

	if def.Name != "acme" || def.DisplayName != "Acme" {
		t.Fatalf("Definition() name/display = %q/%q, want acme/Acme", def.Name, def.DisplayName)
	}
	if def.IntegrationType != "api" || def.ReleaseStage != "beta" || def.DocsURL != "https://example.com/docs" {
		t.Fatalf("Definition() = %+v, want bundle metadata carried over verbatim", def)
	}
	var specDecoded map[string]any
	if err := json.Unmarshal(def.Spec, &specDecoded); err != nil {
		t.Fatalf("Definition().Spec is not valid JSON: %v (%s)", err, def.Spec)
	}
	if len(def.Streams) != 1 || def.Streams[0].Name != "widgets" {
		t.Fatalf("Definition().Streams = %+v, want one stream named widgets", def.Streams)
	}
	if len(def.WriteActions) != 1 || def.WriteActions[0].Name != "update_widget" {
		t.Fatalf("Definition().WriteActions = %+v, want one action named update_widget", def.WriteActions)
	}
	if def.WriteActions[0].Method != "PATCH" || def.WriteActions[0].Risk != "medium" {
		t.Fatalf("Definition().WriteActions[0] = %+v, want method=PATCH risk=medium", def.WriteActions[0])
	}
}

// --- F5 (REVIEW.md flag): Definition.Spec must be the bundle's VERBATIM
// spec.json bytes, not a lossy reconstruction (every property flattened to
// {"type":"string"}, dropping enum/default/required/descriptions/non-string
// types). ---

func richSpecJSONFullBundleFS(name string) fstest.MapFS {
	richSpec := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["base_url"],
		"properties": {
			"base_url": { "type": "string", "description": "Base URL." },
			"port": { "type": "integer", "default": 5432 },
			"sslmode": { "type": "string", "enum": ["disable", "require", "verify-full"], "default": "require" },
			"token": { "type": "string", "x-secret": true }
		}
	}`
	fsys := fullValidBundleFS(name)
	fsys[name+"/spec.json"] = &fstest.MapFile{Data: []byte(richSpec)}
	return fsys
}

func TestDefinitionSpecByteEqualsBundleSpecJSON(t *testing.T) {
	fsys := richSpecJSONFullBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	c := New(b, nil)
	def := c.Definition()

	wantRaw := fsys["acme/spec.json"].Data
	var want, got map[string]any
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("unmarshal want spec.json: %v", err)
	}
	if err := json.Unmarshal(def.Spec, &got); err != nil {
		t.Fatalf("Definition().Spec is not valid JSON: %v (%s)", err, def.Spec)
	}

	// Canonical re-marshal comparison (tolerates whitespace-only differences,
	// not semantic ones): every field (type, enum, default, required,
	// description, x-secret) must survive verbatim — the F5 bug flattened all
	// of this to {"type":"string"}.
	wantCanon, _ := json.Marshal(want)
	gotCanon, _ := json.Marshal(got)
	if string(gotCanon) != string(wantCanon) {
		t.Fatalf("Definition().Spec = %s, want byte-equivalent to bundle spec.json %s", gotCanon, wantCanon)
	}

	// Spot-check the specific fields the lossy reconstruction dropped.
	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Definition().Spec properties = %+v, want an object", got["properties"])
	}
	port, ok := props["port"].(map[string]any)
	if !ok || port["type"] != "integer" {
		t.Fatalf("Definition().Spec properties.port = %+v, want type=integer preserved (not flattened to string)", props["port"])
	}
	if port["default"] == nil {
		t.Fatalf("Definition().Spec properties.port = %+v, want default preserved", props["port"])
	}
	sslmode, ok := props["sslmode"].(map[string]any)
	if !ok || sslmode["enum"] == nil {
		t.Fatalf("Definition().Spec properties.sslmode = %+v, want enum preserved", props["sslmode"])
	}
	if got["required"] == nil {
		t.Fatalf("Definition().Spec = %+v, want top-level required[] preserved", got)
	}
}

// TestConnectorDefinitionSpecFallsBackForAdHocBundleWithNoRawSpec locks in
// backward compatibility: a bundle built ad hoc in a test (never loaded via
// Load, so RawSpec was never populated) still gets a valid Definition().Spec
// derived from whatever *Schema it DID set — the existing
// TestConnectorDefinitionSynthesizedFromBundle behavior, unchanged.
func TestConnectorDefinitionSpecFallsBackForAdHocBundleWithNoRawSpec(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv) // Spec set via minimalSpecSchema, RawSpec never set

	c := New(b, nil)
	def := c.Definition()

	var specDecoded map[string]any
	if err := json.Unmarshal(def.Spec, &specDecoded); err != nil {
		t.Fatalf("Definition().Spec is not valid JSON: %v (%s)", err, def.Spec)
	}
	props, ok := specDecoded["properties"].(map[string]any)
	if !ok || props["api_key"] == nil {
		t.Fatalf("Definition().Spec = %+v, want reconstructed properties.api_key (fallback path for a bundle with no RawSpec)", specDecoded)
	}
}

// --- derived sync modes truth table (design §B.6) ---

func TestDerivedSyncModesTruthTable(t *testing.T) {
	baseStream := StreamSpec{Name: "widgets"}

	cases := []struct {
		name        string
		primaryKey  string
		cursorField string
		incremental *IncrementalSpec
		want        []string
	}{
		{
			name: "neither pk nor incremental",
			want: []string{"full_refresh_append", "full_refresh_overwrite"},
		},
		{
			name:       "primary key alone does not admit cursor-required compatibility modes",
			primaryKey: "id",
			want:       []string{"full_refresh_append", "full_refresh_overwrite"},
		},
		{
			name:        "incremental cursor without schema cursor advertises incremental modes",
			primaryKey:  "id",
			incremental: &IncrementalSpec{CursorField: "updated_at"},
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
				"incremental_append", "incremental_append_deduped",
			},
		},
		{
			name:        "cursor field without incremental block retains only full refresh modes",
			primaryKey:  "id",
			cursorField: "updated_at",
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
			},
		},
		{
			name:        "incremental only adds incremental_append (no dedup, no pk)",
			cursorField: "updated_at",
			incremental: &IncrementalSpec{CursorField: "updated_at"},
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "incremental_append",
			},
		},
		{
			name:        "both pk and incremental add every mode",
			primaryKey:  "id",
			cursorField: "updated_at",
			incremental: &IncrementalSpec{CursorField: "updated_at"},
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
				"incremental_append", "incremental_append_deduped",
			},
		},
		{
			name:        "mismatched incremental and schema cursors do not advertise incremental modes",
			primaryKey:  "id",
			cursorField: "updated_at",
			incremental: &IncrementalSpec{CursorField: "created_at"},
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
			},
		},
		{
			name:        "empty incremental cursor retains schema-backed incremental modes",
			primaryKey:  "id",
			cursorField: "updated_at",
			incremental: &IncrementalSpec{},
			want: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
				"incremental_append", "incremental_append_deduped",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stream := baseStream
			stream.Incremental = tc.incremental
			var sch *StreamSchema
			if tc.primaryKey != "" || tc.cursorField != "" {
				sch = widgetsRecordSchema(t, tc.primaryKey, tc.cursorField)
			} else {
				sch = widgetsRecordSchema(t, "", "")
			}
			got := DerivedSyncModes(stream, sch)
			assertStringSliceEqual(t, got, tc.want)
		})
	}
}

func TestDerivedSyncModesNilSchemaIsNeitherCase(t *testing.T) {
	got := DerivedSyncModes(StreamSpec{Name: "widgets"}, nil)
	assertStringSliceEqual(t, got, []string{"full_refresh_append", "full_refresh_overwrite"})
}

func TestConnectorIncrementalCursorWithoutSchemaFieldProjectsEffectiveCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)
	b.Streams[0].Incremental = &IncrementalSpec{CursorField: "updated_at"}
	b.Schemas["widgets"] = widgetsRecordSchema(t, "id", "")
	b.Schemas["widgets"].Raw = json.RawMessage(`{
		"type": "object",
		"x-primary-key": ["id"],
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"updated_at": {"type": "string"}
		}
	}`)

	c := New(b, nil)
	catalog, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(catalog.Streams) != 1 || !reflect.DeepEqual(catalog.Streams[0].CursorFields, []string{"updated_at"}) {
		t.Fatalf("Catalog().Streams = %+v, want incremental cursor projection", catalog.Streams)
	}

	definition := c.Definition()
	if len(definition.Streams) != 1 || definition.Streams[0].CursorField != "updated_at" {
		t.Fatalf("Definition().Streams = %+v, want incremental cursor projection", definition.Streams)
	}
	assertStringSliceEqual(t, definition.Streams[0].SyncModes, []string{
		"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
		"incremental_append", "incremental_append_deduped",
	})
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice = %v, want %v", got, want)
		}
	}
}

// --- Write without writes.json -> ErrUnsupportedOperation ---

func TestConnectorWriteWithoutWritesJSONReturnsErrUnsupportedOperation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv) // no Writes set

	c := New(b, nil)
	_, err := c.Write(context.Background(), connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "1"}})
	if err == nil {
		t.Fatalf("Write() error = nil, want ErrUnsupportedOperation")
	}
	if !errorsIs(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write() error = %v, want wrapping connectors.ErrUnsupportedOperation", err)
	}
}

func TestConnectorValidateWriteWithoutWritesJSONReturnsErrUnsupportedOperation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "1"}})
	if !errorsIs(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("ValidateWrite() error = %v, want wrapping connectors.ErrUnsupportedOperation", err)
	}
}

func TestConnectorDryRunWriteWithoutWritesJSONReturnsErrUnsupportedOperation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	_, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "1"}})
	if !errorsIs(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("DryRunWrite() error = %v, want wrapping connectors.ErrUnsupportedOperation", err)
	}
}

func errorsIs(err, target error) bool {
	for err != nil {
		if err == target { //nolint:errorlint // simple identity walk mirroring errors.Is without importing it twice here
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// --- Write WITH writes.json succeeds through to engine.Write ---

func TestConnectorWriteWithWritesJSONDelegatesToEngineWrite(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)
	b.Writes = []WriteAction{
		{Name: "update_widget", Kind: "update", Method: "PATCH", Path: "/widgets/{{ record.id }}", PathFields: []string{"id"}},
	}

	c := New(b, nil)
	result, err := c.Write(context.Background(), connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{{"id": "42", "name": "gizmo"}})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("Write() RecordsWritten = %d, want 1", result.RecordsWritten)
	}
	if gotMethod != "PATCH" || gotPath != "/widgets/42" {
		t.Fatalf("request = %s %s, want PATCH /widgets/42", gotMethod, gotPath)
	}
}

// --- Check/Catalog/Read delegate to package-level engine functions ---

func TestConnectorCheckDelegatesToEngineCheck(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)
	b.HTTP.Check = &RequestSpec{Method: "GET", Path: "/status"}

	c := New(b, nil)
	if err := c.Check(context.Background(), connectors.RuntimeConfig{}); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !called {
		t.Fatalf("Check() did not call the declarative check endpoint")
	}
}

func TestConnectorCatalogReflectsBundleStreams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	catalog, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog() error = %v", err)
	}
	if catalog.Connector != "acme" {
		t.Fatalf("Catalog().Connector = %q, want acme", catalog.Connector)
	}
	if len(catalog.Streams) != 1 || catalog.Streams[0].Name != "widgets" {
		t.Fatalf("Catalog().Streams = %+v, want one stream named widgets", catalog.Streams)
	}
	if len(catalog.Streams[0].PrimaryKey) != 1 || catalog.Streams[0].PrimaryKey[0] != "id" {
		t.Fatalf("Catalog().Streams[0].PrimaryKey = %v, want [id]", catalog.Streams[0].PrimaryKey)
	}
}

func TestConnectorReadDelegatesToEngineRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"a"},{"id":"2","name":"b"}]}`))
	}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	var records []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "widgets"}, func(r connectors.Record) error {
		records = append(records, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("Read() emitted %d records, want 2", len(records))
	}
}

func TestConnectorInitialStateDelegatesToEngineInitialState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	b := newConnectorTestBundle(t, srv)

	c := New(b, nil)
	state, err := c.InitialState(context.Background(), "widgets", connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("InitialState() error = %v", err)
	}
	if state["stream"] != "widgets" {
		t.Fatalf("InitialState()[stream] = %q, want widgets", state["stream"])
	}
}

// --- engine.Base serves Definition() for a Tier-3 fake ---

// tier3FakeConnector is the minimal Tier-3 shape: it embeds engine.Base for
// Name/Metadata/Definition and supplies its own Check/Catalog/Read/Write, per
// design §B.7 Tier 3 (native/postgres follows this exact pattern in a later
// wave).
type tier3FakeConnector struct {
	Base
}

func (tier3FakeConnector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return nil
}

func (f tier3FakeConnector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: f.Name()}, nil
}

func (tier3FakeConnector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return nil
}

func (tier3FakeConnector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

var _ connectors.Connector = tier3FakeConnector{}

func TestBaseServesDefinitionForTier3Fake(t *testing.T) {
	b := Bundle{
		Name: "postgres-fake",
		Metadata: Metadata{
			Name:            "postgres-fake",
			DisplayName:     "Postgres (fake)",
			IntegrationType: "database",
			ReleaseStage:    "ga",
			Capabilities:    Capabilities{Check: true, Read: true, CDC: true, DynamicSchema: true},
		},
		Spec: minimalSpecSchema(t),
	}

	fake := tier3FakeConnector{Base: NewBase(b)}

	if got := fake.Name(); got != "postgres-fake" {
		t.Fatalf("Name() = %q, want postgres-fake", got)
	}
	md := fake.Metadata()
	if md.Name != "postgres-fake" || md.IntegrationType != "database" {
		t.Fatalf("Metadata() = %+v, want name=postgres-fake integration_type=database", md)
	}

	def := fake.Definition()
	if def.Name != "postgres-fake" || def.DisplayName != "Postgres (fake)" {
		t.Fatalf("Definition() = %+v, want name=postgres-fake display_name=%q", def, "Postgres (fake)")
	}
	if def.IntegrationType != "database" || def.ReleaseStage != "ga" {
		t.Fatalf("Definition() = %+v, want integration_type=database release_stage=ga", def)
	}
	if len(def.Streams) != 0 {
		t.Fatalf("Definition().Streams = %+v, want empty for a dynamic-schema bundle with no streams.json", def.Streams)
	}
}
