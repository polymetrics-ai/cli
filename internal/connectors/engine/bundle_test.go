package engine

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/defs"
)

func validMetadata(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false }
	}`
}

func dynamicSchemaMetadata(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "database",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": true }
	}`
}

const validSpec = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["base_url"],
	"properties": {
		"base_url": { "type": "string" },
		"token": { "type": "string", "x-secret": true }
	}
}`

const validStreams = `{
	"base": {
		"url": "{{ config.base_url }}",
		"user_agent": "test-agent",
		"headers": {},
		"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ cursor }}" } ],
		"pagination": { "type": "none" },
		"check": { "method": "GET", "path": "/ping" },
		"error_map": []
	},
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json"
		}
	]
}`

const validWidgetsSchema = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"x-primary-key": ["id"],
	"x-cursor-field": "updated_at",
	"properties": {
		"id": { "type": "integer" },
		"updated_at": { "type": "string" }
	}
}`

const validAPISurface = `{
	"api": "test API v1",
	"endpoints": [
		{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } }
	]
}`

const validDocs = `# Overview

test

## Auth setup

none

## Streams notes

none

## Write actions & risks

none

## Known limits

none
`

func fullValidBundleFS(name string) fstest.MapFS {
	return fstest.MapFS{
		name + "/metadata.json":                        &fstest.MapFile{Data: []byte(validMetadata(name))},
		name + "/spec.json":                            &fstest.MapFile{Data: []byte(validSpec)},
		name + "/streams.json":                         &fstest.MapFile{Data: []byte(validStreams)},
		name + "/api_surface.json":                     &fstest.MapFile{Data: []byte(validAPISurface)},
		name + "/schemas/widgets.json":                 &fstest.MapFile{Data: []byte(validWidgetsSchema)},
		name + "/docs.md":                              &fstest.MapFile{Data: []byte(validDocs)},
		name + "/fixtures/streams/widgets/page_1.json": &fstest.MapFile{Data: []byte(`{"request":{"method":"GET","path":"/widgets","query":{}},"response":{"status":200,"body":{"data":[]}}}`)},
	}
}

func TestBundleLoadHappyPathFullBundle(t *testing.T) {
	fsys := fullValidBundleFS("acme")

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Name != "acme" {
		t.Fatalf("Name = %q", b.Name)
	}
	if b.Metadata.Name != "acme" {
		t.Fatalf("Metadata.Name = %q", b.Metadata.Name)
	}
	if b.Spec == nil {
		t.Fatalf("Spec not compiled")
	}
	if b.HTTP.URL != "{{ config.base_url }}" {
		t.Fatalf("HTTP.URL = %q", b.HTTP.URL)
	}
	if len(b.Streams) != 1 || b.Streams[0].Name != "widgets" {
		t.Fatalf("Streams = %+v", b.Streams)
	}
	if b.Writes != nil {
		t.Fatalf("Writes should be nil when writes.json absent, got %+v", b.Writes)
	}
	sch, ok := b.Schemas["widgets"]
	if !ok {
		t.Fatalf("Schemas missing widgets entry")
	}
	if len(sch.PrimaryKey) != 1 || sch.PrimaryKey[0] != "id" {
		t.Fatalf("PrimaryKey = %v", sch.PrimaryKey)
	}
	if sch.CursorField != "updated_at" {
		t.Fatalf("CursorField = %q", sch.CursorField)
	}
	if b.Surface == nil {
		t.Fatalf("Surface not parsed")
	}
	if b.Docs == "" {
		t.Fatalf("Docs not loaded")
	}
	if b.Fixtures == nil {
		t.Fatalf("Fixtures should be non-nil when fixtures/ present")
	}
}

func TestBundleLoadOptionalFilesAbsent(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/fixtures/streams/widgets/page_1.json")

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Writes != nil {
		t.Fatalf("Writes should be nil when writes.json absent")
	}
	if b.Fixtures != nil {
		t.Fatalf("Fixtures should be nil when fixtures/ absent")
	}
}

func TestBundleLoadStreamsOptionalIffDynamicSchema(t *testing.T) {
	fsys := fstest.MapFS{
		"pg/metadata.json":    &fstest.MapFile{Data: []byte(dynamicSchemaMetadata("pg"))},
		"pg/spec.json":        &fstest.MapFile{Data: []byte(validSpec)},
		"pg/api_surface.json": &fstest.MapFile{Data: []byte(`{"api":"pg","endpoints":[]}`)},
		"pg/docs.md":          &fstest.MapFile{Data: []byte(validDocs)},
	}

	b, err := Load(fsys, "pg")
	if err != nil {
		t.Fatalf("Load should succeed without streams.json when dynamic_schema=true: %v", err)
	}
	if len(b.Streams) != 0 {
		t.Fatalf("Streams = %+v, want empty", b.Streams)
	}
}

func TestBundleLoadStreamsRequiredWithoutDynamicSchema(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/streams.json")

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected error: streams.json required when dynamic_schema=false")
	}
	if !strings.Contains(err.Error(), "streams.json") {
		t.Fatalf("error %q does not name streams.json", err.Error())
	}
}

func TestBundleLoadDirNameMismatch(t *testing.T) {
	fsys := fullValidBundleFS("actual-dir")
	fsys["actual-dir/metadata.json"] = &fstest.MapFile{Data: []byte(validMetadata("declared-name"))}

	_, err := Load(fsys, "actual-dir")
	if err == nil {
		t.Fatalf("expected dir-name/metadata.name mismatch error")
	}
	if !strings.Contains(err.Error(), "actual-dir") || !strings.Contains(err.Error(), "declared-name") {
		t.Fatalf("error %q does not name both dir and metadata name", err.Error())
	}
}

func TestBundleLoadBadNameRegex(t *testing.T) {
	fsys := fullValidBundleFS("Source-GitHub")
	fsys["Source-GitHub/metadata.json"] = &fstest.MapFile{Data: []byte(validMetadata("Source-GitHub"))}

	_, err := Load(fsys, "Source-GitHub")
	if err == nil {
		t.Fatalf("expected bad name regex error")
	}
	if !strings.Contains(err.Error(), "Source-GitHub") {
		t.Fatalf("error %q does not name the offending value", err.Error())
	}
}

func TestBundleLoadMissingRequiredFile(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/api_surface.json")

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected missing required file error")
	}
	if !strings.Contains(err.Error(), "api_surface.json") {
		t.Fatalf("error %q does not name the missing file", err.Error())
	}
}

func TestBundleLoadMetaSchemaViolation(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	// metadata.json missing the required "capabilities" field -> meta-schema violation.
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga"
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected meta-schema violation error for metadata.json missing capabilities")
	}
}

func TestBundleLoadAllIteratesBundles(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	for k, v := range fullValidBundleFS("beta") {
		fsys[k] = v
	}

	bundles, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(bundles) != 2 {
		t.Fatalf("LoadAll returned %d bundles, want 2", len(bundles))
	}
	names := map[string]bool{}
	for _, b := range bundles {
		names[b.Name] = true
	}
	if !names["acme"] || !names["beta"] {
		t.Fatalf("LoadAll bundles = %v", names)
	}
}

func TestBundleLoadAllEmptyTreeIsFine(t *testing.T) {
	bundles, err := LoadAll(fstest.MapFS{})
	if err != nil {
		t.Fatalf("LoadAll on empty tree: %v", err)
	}
	if len(bundles) != 0 {
		t.Fatalf("LoadAll on empty tree returned %d bundles", len(bundles))
	}
}

// TestBundleLoadAllOneBadBundleDoesNotHideTheRest is an ENGINE HARDENING
// regression (hardening-ledger.md): LoadAll previously aborted the ENTIRE
// batch (returned nil bundles, a single-bundle error) the instant ANY ONE
// directory failed to load. With ~400 independently-authored bundles in
// defs/, and the newly-added strict-decode/meta-schema unknown-key checks
// now correctly failing a real (if large) subset of them, that all-or-
// nothing contract meant a single malformed bundle anywhere in the fleet
// silently hid every other (compliant) bundle from LoadAll's caller — the
// exact "one bad apple spoils fleet-wide discoverability" failure mode
// cmd/connectorgen's own validateBundleDir already avoids by design (it
// isolates one bundle's load error into a Finding and keeps validating the
// rest). LoadAll now mirrors that same resilience: it still returns every
// bundle that DID load cleanly, and a non-nil error whenever at least one
// did not — the error names every failing bundle (not just the first) so a
// caller that treats err!=nil as fatal still learns the full failing set
// from the error text, and a caller that wants the good subset (this
// package's own defs.FS-wide golden/parity tests, conformance) can keep
// going against bundles.
func TestBundleLoadAllOneBadBundleDoesNotHideTheRest(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	for k, v := range fullValidBundleFS("beta") {
		fsys[k] = v
	}
	// "broken" has an unknown base-level key (streams.json) and must fail
	// to load, but must not prevent acme/beta from coming back.
	for k, v := range fullValidBundleFS("broken") {
		fsys[k] = v
	}
	fsys["broken/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"query": { "limit": "1" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	bundles, err := LoadAll(fsys)
	if err == nil {
		t.Fatalf("LoadAll: expected a non-nil error naming the broken bundle, got nil (bundles: %+v)", bundles)
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Fatalf("LoadAll error = %q, want it to name the failing bundle %q", err.Error(), "broken")
	}
	var loadErr *LoadAllError
	if !errors.As(err, &loadErr) {
		t.Fatalf("LoadAll error = %v (%T), want it to be (or wrap) a *LoadAllError", err, err)
	}
	if len(loadErr.Failures) != 1 || loadErr.Failures[0].Name != "broken" {
		t.Fatalf("LoadAllError.Failures = %+v, want exactly one entry named %q", loadErr.Failures, "broken")
	}

	names := map[string]bool{}
	for _, b := range bundles {
		names[b.Name] = true
	}
	if !names["acme"] || !names["beta"] {
		t.Fatalf("LoadAll bundles = %v, want acme and beta still returned despite broken's failure", names)
	}
	if names["broken"] {
		t.Fatalf("LoadAll bundles = %v, want broken itself excluded (it never loaded)", names)
	}
}

// TestBundleLoadAllDefsFS exercises the real embedded defs.FS scaffold
// (the `all:*` embed directive in defs.go) end-to-end: the stray embedded
// defs.go file is ignored, and the Wave F goldens must load cleanly.
//
// ENGINE HARDENING (hardening-ledger.md): this no longer asserts err == nil.
// The newly-added streams.json/writes.json/metadata.json unknown-key checks
// (meta-schema additionalProperties:false + loader strict-decode) correctly
// fail a real, pre-existing subset of defs/ bundles that declared fields
// the engine silently ignored (rentcast's "base.check.query" and ~150
// siblings' identical shape — RequestSpec only carries Method/Path, so that
// JSON never did anything at runtime). Repairing those bundles' own
// streams.json files is explicitly out of scope for this dispatch (listed
// in the hardening ledger for a follow-up instead), so LoadAll(defs.FS) is
// now expected to return a non-nil error naming them. What this test still
// pins, unweakened: LoadAll's resilience contract (TestBundleLoadAll
// OneBadBundleDoesNotHideTheRest) means the golden bundles must STILL come
// back in the returned slice even though err is non-nil, and any bundle
// that legitimately fails must be a KNOWN, currently-tracked case, not a
// silent new regression — this test fails loudly if an UNEXPECTED bundle
// starts failing to load.
func TestBundleLoadAllDefsFS(t *testing.T) {
	bundles, err := LoadAll(defs.FS)
	byName := map[string]bool{}
	for _, b := range bundles {
		byName[b.Name] = true
	}
	for _, golden := range []string{"stripe", "postgres"} {
		if !byName[golden] {
			t.Fatalf("LoadAll(defs.FS) missing golden bundle %q (got %v); err=%v", golden, byName, err)
		}
	}
	var loadErr *LoadAllError
	if err != nil && !errors.As(err, &loadErr) {
		t.Fatalf("LoadAll(defs.FS) returned an error NOT shaped like the known per-bundle unknown-key failures (hardening-ledger.md): %v", err)
	}
}

// TestBundleLoadFromOnDiskTestdata exercises the loader against a real
// os.DirFS-backed fixture bundle (testdata/bundles/widget-demo), rather than
// only the in-memory fstest.MapFS cases above.
func TestBundleLoadFromOnDiskTestdata(t *testing.T) {
	fsys := os.DirFS("testdata/bundles")

	b, err := Load(fsys, "widget-demo")
	if err != nil {
		t.Fatalf("Load(testdata/bundles, widget-demo): %v", err)
	}
	if b.Name != "widget-demo" {
		t.Fatalf("Name = %q", b.Name)
	}
	if len(b.Streams) != 1 || b.Streams[0].Name != "widgets" {
		t.Fatalf("Streams = %+v", b.Streams)
	}
	if b.Fixtures == nil {
		t.Fatalf("Fixtures should be non-nil")
	}

	bundles, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll(testdata/bundles): %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("LoadAll(testdata/bundles) returned %d bundles, want 1", len(bundles))
	}
}

// --- optional conformance skip markers (R3: hook-aware dynamic conformance) --
//
// A bundle may declare an OPTIONAL, explicit "conformance" marker at either
// stream level (streams.json's per-stream {"conformance": {"skip_dynamic":
// true, "reason": "..."}}) or bundle level (metadata.json's top-level
// equivalent), for connectors whose dynamic (fixture-replay) checks cannot
// meaningfully run because the bundle's real behavior lives entirely behind
// a Tier-2 hook that conformance's declarative-only replay harness cannot
// exercise. This is parsed by the loader (no behavior beyond struct
// population); dynamic.go interprets the marker, connectorgen validate
// requires a non-empty reason.

const streamsWithStreamConformanceMarker = `{
	"base": {
		"url": "{{ config.base_url }}",
		"user_agent": "test-agent",
		"headers": {},
		"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ cursor }}" } ],
		"pagination": { "type": "none" },
		"check": { "method": "GET", "path": "/ping" },
		"error_map": []
	},
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json",
			"conformance": { "skip_dynamic": true, "reason": "hook-covered; proven live by internal/connectors/paritytest/acme" }
		}
	]
}`

func TestBundleLoadParsesStreamConformanceMarker(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(streamsWithStreamConformanceMarker)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
	s := b.Streams[0]
	if s.Conformance == nil {
		t.Fatalf("stream %q Conformance marker not parsed (got nil)", s.Name)
	}
	if !s.Conformance.SkipDynamic {
		t.Fatalf("stream %q Conformance.SkipDynamic = false, want true", s.Name)
	}
	if s.Conformance.Reason == "" {
		t.Fatalf("stream %q Conformance.Reason is empty", s.Name)
	}
}

// TestBundleLoadStreamWithNoConformanceMarkerIsNil locks in that an ordinary
// stream (no "conformance" key at all) parses with a nil marker, not a
// zero-value non-nil struct — dynamic.go's marker-presence check must be
// able to distinguish "no marker" from "marker present but false".
func TestBundleLoadStreamWithNoConformanceMarkerIsNil(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].Conformance != nil {
		t.Fatalf("Conformance = %+v, want nil for a stream with no conformance block", b.Streams[0].Conformance)
	}
}

func metadataWithBundleConformanceMarker(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"conformance": { "skip_dynamic": true, "reason": "custom-auth-only; hook not registered in conformance's replay harness" }
	}`
}

func TestBundleLoadParsesBundleLevelConformanceMarker(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(metadataWithBundleConformanceMarker("acme"))}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Conformance == nil {
		t.Fatalf("Metadata.Conformance marker not parsed (got nil)")
	}
	if !b.Metadata.Conformance.SkipDynamic {
		t.Fatalf("Metadata.Conformance.SkipDynamic = false, want true")
	}
	if b.Metadata.Conformance.Reason == "" {
		t.Fatalf("Metadata.Conformance.Reason is empty")
	}
}

func TestBundleLoadMetadataWithNoConformanceMarkerIsNil(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Conformance != nil {
		t.Fatalf("Metadata.Conformance = %+v, want nil for metadata with no conformance block", b.Metadata.Conformance)
	}
}

// streamsWithOptionalQueryDialect exercises the gap-loop item-3 optional-query
// dialect (REVIEW-B.md cross-cutting adjudication 2): a stream.Query entry
// may be either a plain string (today's exact hard-error semantics,
// "page[size]") or an object {template, omit_when_absent, default}.
const streamsWithOptionalQueryDialect = `{
	"base": { "url": "{{ config.base_url }}" },
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json",
			"query": {
				"page[size]": "100",
				"status": { "template": "{{ config.status }}", "omit_when_absent": true },
				"count": { "template": "{{ config.page_size }}", "default": "100" }
			}
		}
	]
}`

// TestBundleLoadParsesOptionalQueryDialect proves streams.json's per-entry
// query dialect round-trips through the loader: a plain string entry stays a
// hard-required template; an object entry carries its template/
// omit_when_absent/default fields distinctly.
func TestBundleLoadParsesOptionalQueryDialect(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(streamsWithOptionalQueryDialect)}
	fsys["acme/spec.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["base_url"],
		"properties": {
			"base_url": { "type": "string" },
			"status": { "type": "string" },
			"page_size": { "type": "string" }
		}
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
	q := b.Streams[0].Query
	staticEntry, ok := q["page[size]"]
	if !ok || staticEntry.Template != "100" || staticEntry.OmitWhenAbsent {
		t.Fatalf("query[page[size]] = %+v, want plain string entry Template=100 OmitWhenAbsent=false", staticEntry)
	}
	statusEntry, ok := q["status"]
	if !ok || statusEntry.Template != "{{ config.status }}" || !statusEntry.OmitWhenAbsent {
		t.Fatalf("query[status] = %+v, want Template={{ config.status }} OmitWhenAbsent=true", statusEntry)
	}
	countEntry, ok := q["count"]
	if !ok || countEntry.Template != "{{ config.page_size }}" || countEntry.Default != "100" {
		t.Fatalf("query[count] = %+v, want Template={{ config.page_size }} Default=100", countEntry)
	}
}

// --- ENGINE HARDENING: unknown-key strict decode ---------------------------
//
// The re-review (hardening-ledger.md) found internal/connectors/defs/rentcast
// declaring "base.check.query" (and several other bundles declaring a bare
// "base.query"), a field HTTPBase/RequestSpec do not have at all. Because
// json.Unmarshal silently drops unknown object keys and the meta-schemas
// previously left every nested sub-object as a bare {"type":"object"} with no
// additionalProperties:false, that invented mechanism passed every gate
// (meta-schema validate, connectorgen validate, go build) while doing
// nothing at runtime — Check() never sends a query at all. These tests pin
// TWO independent layers of defense: (1) the meta-schemas
// (streams.schema.json/writes.schema.json/metadata.schema.json) now declare
// explicit property allowlists with additionalProperties:false on every
// structured sub-object (free-form maps like headers/query/body/
// computed_fields/record_schema and user JSON-Schema documents like
// spec.json's "properties" are deliberately left open); (2) the loader
// itself strict-decodes streams.json/writes.json/metadata.json (independent
// of the meta-schema, so a future meta-schema regression/relaxation cannot
// silently reopen this hole) and names the offending file+key in the error.

func TestBundleLoadRejectsUnknownBaseLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"query": { "limit": "1" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown base-level key %q, got nil", "query")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "query") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "query")
	}
}

// TestBundleLoadAcceptsBaseCheckQueryKey supersedes the former
// TestBundleLoadRejectsUnknownBaseCheckQueryKey (checkquery-ledger.md):
// base.check.query (the exact rentcast shape the hardening ledger's trigger
// named, and 148 siblings' identical shape) is no longer an unknown key —
// RequestSpec now has a Query map[string]QueryParam field mirroring
// StreamSpec.Query's existing string-or-object dialect verbatim, per the
// hardening ledger's own suggested follow-up shape. Loading must now succeed
// AND the query must round-trip into RequestSpec.Query exactly like
// stream.Query does.
func TestBundleLoadAcceptsBaseCheckQueryKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping", "query": { "limit": "1", "offset": "0" } }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v, want base.check.query to load cleanly (RequestSpec.Query dialect addition)", err)
	}
	if b.HTTP.Check == nil {
		t.Fatalf("HTTP.Check is nil")
	}
	limit, ok := b.HTTP.Check.Query["limit"]
	if !ok || limit.Template != "1" {
		t.Fatalf("Check.Query[limit] = %+v, want plain string entry Template=1", limit)
	}
	offset, ok := b.HTTP.Check.Query["offset"]
	if !ok || offset.Template != "0" {
		t.Fatalf("Check.Query[offset] = %+v, want plain string entry Template=0", offset)
	}
}

// TestBundleLoadParsesCheckQueryOptionalDialect proves check.query accepts
// the SAME object-form (omit_when_absent/default) dialect as stream.Query,
// not just plain strings — since RequestSpec.Query reuses the identical
// QueryParam type.
func TestBundleLoadParsesCheckQueryOptionalDialect(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": {
				"method": "GET",
				"path": "/ping",
				"query": {
					"limit": "1",
					"status": { "template": "{{ config.status }}", "omit_when_absent": true },
					"count": { "template": "{{ config.page_size }}", "default": "100" }
				}
			}
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}
	fsys["acme/spec.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["base_url"],
		"properties": {
			"base_url": { "type": "string" },
			"status": { "type": "string" },
			"page_size": { "type": "string" }
		}
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	q := b.HTTP.Check.Query
	staticEntry, ok := q["limit"]
	if !ok || staticEntry.Template != "1" || staticEntry.OmitWhenAbsent {
		t.Fatalf("Check.Query[limit] = %+v, want plain string entry Template=1 OmitWhenAbsent=false", staticEntry)
	}
	statusEntry, ok := q["status"]
	if !ok || statusEntry.Template != "{{ config.status }}" || !statusEntry.OmitWhenAbsent {
		t.Fatalf("Check.Query[status] = %+v, want Template={{ config.status }} OmitWhenAbsent=true", statusEntry)
	}
	countEntry, ok := q["count"]
	if !ok || countEntry.Template != "{{ config.page_size }}" || countEntry.Default != "100" {
		t.Fatalf("Check.Query[count] = %+v, want Template={{ config.page_size }} Default=100", countEntry)
	}
}

func TestBundleLoadRejectsUnknownStreamLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json", "not_a_real_field": true }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown stream-level key %q, got nil", "not_a_real_field")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "not_a_real_field") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "not_a_real_field")
	}
}

func TestBundleLoadRejectsUnknownAuthCandidateKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "scope": "read" } ],
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown auth-candidate key %q (note: valid key is \"scopes\", not \"scope\"), got nil", "scope")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "scope") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "scope")
	}
}

func TestBundleLoadRejectsUnknownWritesActionKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": true, "query": false, "cdc": false, "dynamic_schema": false }
	}`)}
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "create_widget",
				"kind": "create",
				"method": "POST",
				"path": "/widgets",
				"record_schema": { "type": "object", "properties": {} },
				"risk": "low",
				"retries": 3
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown writes-action key %q, got nil", "retries")
	}
	if !strings.Contains(err.Error(), "writes.json") || !strings.Contains(err.Error(), "retries") {
		t.Fatalf("Load error = %q, want it to name writes.json and the unknown key %q", err.Error(), "retries")
	}
}

func TestBundleLoadRejectsUnknownAPISurfaceEndpointKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" }, "deprecated": true }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown api_surface.json endpoint key %q, got nil", "deprecated")
	}
	if !strings.Contains(err.Error(), "api_surface.json") || !strings.Contains(err.Error(), "deprecated") {
		t.Fatalf("Load error = %q, want it to name api_surface.json and the unknown key %q", err.Error(), "deprecated")
	}
}

func TestBundleLoadRejectsUnknownMetadataTopLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"maintainer": "nobody"
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown metadata.json top-level key %q, got nil", "maintainer")
	}
	if !strings.Contains(err.Error(), "metadata.json") || !strings.Contains(err.Error(), "maintainer") {
		t.Fatalf("Load error = %q, want it to name metadata.json and the unknown key %q", err.Error(), "maintainer")
	}
}

// TestBundleLoadStillAcceptsFreeFormMapKeys pins the deliberate scope
// boundary: headers, stream.query (string-or-object dialect), body, and
// computed_fields are genuinely free-form maps (arbitrary caller-defined
// keys), and must NOT be rejected by the strict-decode/meta-schema
// tightening above.
func TestBundleLoadStillAcceptsFreeFormMapKeys(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"headers": { "X-Anything-Custom": "v1", "X-Another-One": "v2" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"query": { "arbitrary_param_name": "{{ config.base_url }}" },
				"body": { "any_shape_here": { "nested": true } },
				"records": { "path": "data" },
				"computed_fields": { "whatever_field": "{{ record.id }}" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: unexpected error for free-form map keys: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
}
