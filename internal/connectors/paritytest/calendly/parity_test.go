package calendlyparity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/calendly"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// This file is the wave1-pilot P-4 parity suite for the calendly bundle
// (internal/connectors/defs/calendly): both the legacy hand-written
// calendly.Connector (internal/connectors/calendly, read-only reference) and
// the engine-backed connector built from the bundle are driven live against
// the SAME httptest server, RAW connectors.Record equality, real Calendly
// wire shapes (SPEC wave1-pilot §5.2, conventions.md THE recipe).
//
// Documented parity deviation (conventions.md §5 ledger candidate): legacy
// resolves the org-scoping query param (organization=<uri>) DYNAMICALLY on
// every read by calling GET /users/me first (calendly.go:137-141,
// scopeQuery). The engine's declarative dialect has no mechanism to chain
// one request's response into a later request's query params (read.go's
// buildInitialQuery only resolves config/secrets/cursor templates) — this is
// not expressible in Tier 1 without inventing per-connector Go (which would
// need a StreamHook, out of scope per SPEC wave1-pilot §5.2's plain Tier-1
// specification for calendly). The bundle instead declares a required
// `organization_uri` config value that the operator configures once (the
// exact URI legacy would have discovered via /users/me at read time); every
// subsequent request legacy and the engine send is byte-identical GIVEN the
// same organization URI. This never changes emitted record DATA for any
// input legacy itself would accept: it only changes how the operator
// supplies the (invariant, per-account) organization URI. See docs.md "Known
// limits" and conventions.md §5 ledger.
// loadCalendlyBundle uses engine.Load(defs.FS, "calendly") rather than
// engine.LoadAll(defs.FS) deliberately: SPEC wave1-pilot §6 names both as
// the production discovery path, and Load only descends into the "calendly"
// subtree (fs.Sub) rather than requiring every sibling pilot connector's
// bundle to ALSO be structurally complete at the moment this test runs —
// during DW-1's fully-parallel dispatch, nine other agents are concurrently
// writing their own bundle directories under the same defs.FS root, so
// LoadAll would spuriously fail this package's tests on a sibling's
// in-progress (not-yet-complete) bundle, which is neither this package's
// concern nor something it can observe/control.
func loadCalendlyBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "calendly")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "calendly", err)
	}
	return b
}

func withCalendlyBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

const testOrgURI = "https://api.calendly.com/organizations/ORG1"

// calendlyRuntimeConfig builds the shared RuntimeConfig for both connectors:
// legacy discovers the organization URI itself via /users/me (scopeQuery);
// the engine bundle is fed the identical value directly via config
// (organization_uri — the documented parity deviation above). page_size is
// deliberately left UNSET here (gap-loop cycle-1, REVIEW-B.md finding 2): both
// legacy (calendlyPageSize's default-100 fallback) and the engine bundle
// (spec.json's page_size "default": "100", materialized into RuntimeConfig at
// runtime by the C3 engine mechanism, engine/read.go's
// materializeConfigDefaults) fall back to the identical default of 100 when
// the caller never supplies page_size — proving the common (unset) case works
// without a test-only crutch value. extra carries additional config
// (start_date, an explicit page_size override, etc).
func calendlyRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL, "organization_uri": testOrgURI}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"api_key": "cal_test_deadbeefdeadbeef"},
	}
}

func readAllRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return out
}

// normalizeRecord re-encodes r through encoding/json with UseNumber so
// legacy's native Go types and the engine's json.Number-preserving decode
// compare equal on numeric fields (mirrors parity_stripe_test.go).
func normalizeRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	dec := json.NewDecoder(newByteReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode record: %v", err)
	}
	return out
}

func normalizeRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecord(t, r)
	}
	return out
}

type byteReader struct {
	b []byte
	i int
}

func newByteReader(b []byte) *byteReader { return &byteReader{b: b} }

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, errEOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

var errEOF = jsonEOF{}

type jsonEOF struct{}

func (jsonEOF) Error() string { return "EOF" }

// usersMeHandler serves the real Calendly /users/me wire shape: every field
// legacy's calendlyUserRecord mapper reads is present (Calendly's actual API
// always returns the full user object; a field genuinely absent from the
// wire response would make legacy emit an explicit nil for it via its Go
// map's zero-value read, whereas the engine's schema projection would omit
// the key entirely — a divergence that has nothing to do with THIS bundle's
// authoring and everything to do with the raw fixture completeness, so the
// fixture here matches Calendly's real always-fully-populated response).
func usersMeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resource":{
			"uri":"https://api.calendly.com/users/U1",
			"name":"Ada",
			"email":"ada@example.com",
			"slug":"ada",
			"timezone":"UTC",
			"current_organization":"` + testOrgURI + `",
			"created_at":"2026-01-01T00:00:00Z",
			"updated_at":"2026-01-01T00:00:00Z"
		}}`))
	}
}

// --- per-stream record parity across all 5 streams ---

func calendlyStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())

	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"collection":[
			{"uri":"https://api.calendly.com/scheduled_events/E1","name":"Intro","status":"active","start_time":"2026-01-01T10:00:00Z","end_time":"2026-01-01T10:30:00Z","event_type":"https://api.calendly.com/event_types/ET1","created_at":"2026-01-01T09:00:00Z","updated_at":"2026-01-01T09:00:00Z"}
		],"pagination":{"count":1,"next_page":null}}`)
	})

	mux.HandleFunc("/event_types", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"collection":[
			{"uri":"https://api.calendly.com/event_types/ET1","name":"15 Minute Meeting","slug":"15min","active":true,"duration":15,"kind":"solo","scheduling_url":"https://calendly.com/fixture/15min","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}
		],"pagination":{"count":1,"next_page":null}}`)
	})

	mux.HandleFunc("/organization_memberships", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"collection":[
			{"uri":"https://api.calendly.com/organization_memberships/OM1","role":"admin","user":{"uri":"https://api.calendly.com/users/U1","name":"Ada","email":"ada@example.com"},"organization":"`+testOrgURI+`","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}
		],"pagination":{"count":1,"next_page":null}}`)
	})

	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"collection":[
			{"uri":"https://api.calendly.com/groups/G1","name":"Sales","organization":"`+testOrgURI+`","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}
		],"pagination":{"count":1,"next_page":null}}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func TestParityCalendly_StreamRecords(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	streams := []string{"scheduled_events", "event_types", "organization_memberships", "groups"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := calendlyStreamServer(t)

			legacy := calendly.New()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: calendlyRuntimeConfig(srv.URL, nil)})

			eng := engine.New(withCalendlyBaseURL(bundle, srv.URL), nil)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: calendlyRuntimeConfig(srv.URL, nil)})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy calendly emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeRecords(t, engRecs)
			wantNorm := normalizeRecords(t, legacyRecs)
			for i := range wantNorm {
				got := gotNorm[i]
				want := wantNorm[i]
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, got, want)
				}
			}
		})
	}
}

// TestParityCalendly_UsersSingleObject exercises the users/me single_object
// stream separately: it has no collection envelope and no org scoping.
func TestParityCalendly_UsersSingleObject(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	srv := calendlyStreamServer(t)

	legacy := calendly.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "users", Config: calendlyRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy users records = %d, want 1 (single_object)", len(legacyRecs))
	}

	eng := engine.New(withCalendlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "users", Config: calendlyRuntimeConfig(srv.URL, nil)})
	if len(engRecs) != 1 {
		t.Fatalf("engine users records = %d, want 1 (single_object)", len(engRecs))
	}

	gotNorm := normalizeRecords(t, engRecs)[0]
	wantNorm := normalizeRecords(t, legacyRecs)[0]
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("users record mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// TestParityCalendly_UsersStreamPaginationExplicitlyNone asserts the "users"
// single_object stream (gap-loop cycle-1, REVIEW-B.md finding 4) declares an
// explicit stream-level "pagination": {"type": "none"} override rather than
// silently inheriting base.pagination's next_url paginator. Before the fix,
// /users/me had no declared override, so it inherited the next_url paginator
// applied to every other (collection) stream — harmless today only because
// /users/me's response has no pagination.next_page envelope to trigger it,
// but a stream-level declaration WRONG for the stream's real (single-object,
// unpaginated) shape; a stream-level Pagination entirely replaces base's
// wholesale (bundle.go's StreamSpec.Pagination doc comment: "overrides
// base"), so this also protects against any future base.pagination change
// silently reaching a single_object stream it was never meant to apply to.
func TestParityCalendly_UsersStreamPaginationExplicitlyNone(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	var usersStream *engine.StreamSpec
	for i := range bundle.Streams {
		if bundle.Streams[i].Name == "users" {
			usersStream = &bundle.Streams[i]
			break
		}
	}
	if usersStream == nil {
		t.Fatal("bundle has no \"users\" stream")
	}
	if usersStream.Pagination == nil {
		t.Fatal("users stream has no stream-level pagination override, want explicit {type: none}")
	}
	if usersStream.Pagination.Type != "none" {
		t.Fatalf("users stream pagination.type = %q, want %q", usersStream.Pagination.Type, "none")
	}
}

// --- next_url 2-page pagination parity (scheduled_events) ---

func calendlyPaginatedServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())

	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page_token") == "" {
			next := "http://" + r.Host + "/scheduled_events?organization=" + r.URL.Query().Get("organization") + "&count=" + r.URL.Query().Get("count") + "&page_token=TOK2"
			writeJSON(w, `{"collection":[
				{"uri":"https://api.calendly.com/scheduled_events/E1","name":"Intro","status":"active","start_time":"2026-01-01T10:00:00Z","end_time":"2026-01-01T10:30:00Z","event_type":"https://api.calendly.com/event_types/ET1","created_at":"2026-01-01T09:00:00Z","updated_at":"2026-01-01T09:00:00Z"},
				{"uri":"https://api.calendly.com/scheduled_events/E2","name":"Demo","status":"active","start_time":"2026-01-02T10:00:00Z","end_time":"2026-01-02T10:30:00Z","event_type":"https://api.calendly.com/event_types/ET1","created_at":"2026-01-02T09:00:00Z","updated_at":"2026-01-02T09:00:00Z"}
			],"pagination":{"count":2,"next_page":"`+next+`"}}`)
			return
		}
		writeJSON(w, `{"collection":[
			{"uri":"https://api.calendly.com/scheduled_events/E3","name":"Review","status":"active","start_time":"2026-01-03T10:00:00Z","end_time":"2026-01-03T10:30:00Z","event_type":"https://api.calendly.com/event_types/ET1","created_at":"2026-01-03T09:00:00Z","updated_at":"2026-01-03T09:00:00Z"}
		],"pagination":{"count":1,"next_page":null}}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestParityCalendly_ScheduledEventsTwoPagePagination(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	srv := calendlyPaginatedServer(t)

	legacy := calendly.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy scheduled_events records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := engine.New(withCalendlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(srv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine scheduled_events records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotURIs := recordURIs(t, engRecs)
	wantURIs := recordURIs(t, legacyRecs)
	if !reflect.DeepEqual(gotURIs, wantURIs) {
		t.Fatalf("scheduled_events record uri sequence = %v, want %v", gotURIs, wantURIs)
	}
	wantSeq := []string{
		"https://api.calendly.com/scheduled_events/E1",
		"https://api.calendly.com/scheduled_events/E2",
		"https://api.calendly.com/scheduled_events/E3",
	}
	if !reflect.DeepEqual(gotURIs, wantSeq) {
		t.Fatalf("scheduled_events record uri sequence = %v, want %v", gotURIs, wantSeq)
	}
}

func recordURIs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		uri, _ := r["uri"].(string)
		out[i] = uri
	}
	return out
}

// TestParityCalendly_NextPageNullTerminates asserts a null (absent) next_page
// stops pagination after exactly one request on both sides — the same
// "empty/absent next URL is a benign stop" semantics engine/paginate.go's
// nextURL.Next documents.
func TestParityCalendly_NextPageNullTerminates(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	var requestCount int
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())
	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		writeJSON(w, `{"collection":[{"uri":"https://api.calendly.com/scheduled_events/E1","name":"Intro","status":"active","start_time":"2026-01-01T10:00:00Z"}],"pagination":{"count":1,"next_page":null}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	eng := engine.New(withCalendlyBaseURL(bundle, srv.URL), nil)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(srv.URL, nil)})
	if len(engRecs) != 1 {
		t.Fatalf("engine scheduled_events records = %d, want 1", len(engRecs))
	}
	if requestCount != 1 {
		t.Fatalf("scheduled_events request count = %d, want 1 (null next_page must terminate immediately)", requestCount)
	}
}

// --- incremental parity: start_date-raised lower bound (min_start_time) ---

// incrementalCaptureServer answers every scheduled_events request with an
// empty collection (so the read terminates after exactly one request) and
// records the min_start_time query value it observed.
func incrementalCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())
	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("min_start_time")
		writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &got
}

func TestParityCalendly_IncrementalMinStartTimeFromStartDate(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	const startDate = "2026-01-01T00:00:00Z"

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := calendly.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(legacySrv.URL, map[string]string{"start_date": startDate}),
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(engSrv.URL, map[string]string{"start_date": startDate}),
	})

	if *legacyGot != startDate {
		t.Fatalf("legacy min_start_time = %q, want %q (test fixture bug)", *legacyGot, startDate)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine min_start_time = %q, want %q (legacy)", *engGot, *legacyGot)
	}
}

// TestParityCalendly_IncrementalMinStartTimeFromStateCursor feeds both sides
// the SAME app-persisted state cursor (an RFC3339 string — scheduled_events'
// cursor field, start_time, is itself an RFC3339 string on the wire, so
// internal/app/sync_modes.go's recordCursor->toComparableString persists it
// verbatim as RFC3339, unlike a numeric cursor field's digit-string shape;
// the B1 lesson still applies generically but this stream's honest
// persisted shape IS RFC3339) and asserts both forward it identically as
// min_start_time (param_format rfc3339, the default: sent verbatim).
func TestParityCalendly_IncrementalMinStartTimeFromStateCursor(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	const appPersistedCursor = "2026-02-01T00:00:00Z"

	legacySrv, legacyGot := incrementalCaptureServer(t)
	legacy := calendly.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(legacySrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	engSrv, engGot := incrementalCaptureServer(t)
	eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(engSrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})

	if *legacyGot != appPersistedCursor {
		t.Fatalf("legacy min_start_time = %q, want %q (test fixture bug)", *legacyGot, appPersistedCursor)
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine min_start_time = %q, want %q (legacy, same app-persisted cursor)", *engGot, *legacyGot)
	}
}

// TestParityCalendly_MinStartTimeOnlyAppliesToScheduledEvents asserts the
// non-scheduled_events streams never receive min_start_time even when
// start_date is configured — legacy only sets it for the scheduled_events
// resource (calendly.go:158's `if endpoint.resource == "scheduled_events"`
// guard); other streams do have incremental cursor fields (updated_at) but
// no server-side request_param wired for them (client_filtered would be the
// alternative, and legacy does NOT client-filter these streams either — it
// forwards no filter param at all and relies on nothing, i.e. it is a
// full-refresh-in-practice cursor field for those three streams: SPEC
// wave1-pilot doesn't call this out explicitly, but the legacy source is
// unambiguous and this test proves the engine bundle matches it exactly).
func TestParityCalendly_MinStartTimeOnlyAppliesToScheduledEvents(t *testing.T) {
	bundle := loadCalendlyBundle(t)
	const startDate = "2026-01-01T00:00:00Z"

	for _, stream := range []string{"event_types", "organization_memberships", "groups"} {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			var legacyGotMinStart, engGotMinStart string
			legacyMux := http.NewServeMux()
			legacyMux.HandleFunc("/users/me", usersMeHandler())
			legacyMux.HandleFunc("/"+stream, func(w http.ResponseWriter, r *http.Request) {
				legacyGotMinStart = r.URL.Query().Get("min_start_time")
				writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
			})
			legacySrv := httptest.NewServer(legacyMux)
			t.Cleanup(legacySrv.Close)

			legacy := calendly.New()
			_ = readAllRecords(t, legacy, connectors.ReadRequest{
				Stream: stream,
				Config: calendlyRuntimeConfig(legacySrv.URL, map[string]string{"start_date": startDate}),
			})

			engMux := http.NewServeMux()
			engMux.HandleFunc("/users/me", usersMeHandler())
			engMux.HandleFunc("/"+stream, func(w http.ResponseWriter, r *http.Request) {
				engGotMinStart = r.URL.Query().Get("min_start_time")
				writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
			})
			engSrv := httptest.NewServer(engMux)
			t.Cleanup(engSrv.Close)

			eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
			_ = readAllRecords(t, eng, connectors.ReadRequest{
				Stream: stream,
				Config: calendlyRuntimeConfig(engSrv.URL, map[string]string{"start_date": startDate}),
			})

			if legacyGotMinStart != "" {
				t.Fatalf("legacy %s min_start_time = %q, want empty (test fixture bug)", stream, legacyGotMinStart)
			}
			if engGotMinStart != legacyGotMinStart {
				t.Fatalf("engine %s min_start_time = %q, want %q (legacy)", stream, engGotMinStart, legacyGotMinStart)
			}
		})
	}
}

// --- page_size default parity (gap-loop cycle-1, REVIEW-B.md finding 2) ---

// countCaptureServer answers every scheduled_events request with an empty
// collection (single request, then stop) and records the "count" query
// value it observed (the wire param page_size templates into).
func countCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var got string
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())
	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("count")
		writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &got
}

// TestParityCalendly_PageSizeDefaultsTo100WhenUnset asserts that leaving
// page_size UNSET (an input legacy's calendlyPageSize accepts and defaults to
// 100 for, calendly.go:363-376) no longer hard-errors the engine bundle and
// sends the identical "count=100" both sides send. Before the gap-loop fix,
// stream.Query's "count": "{{ config.page_size }}" template had no
// absent-key tolerance, so an unset page_size hard-errored every
// organization-scoped read even though page_size was NOT in spec.json's
// required[] — an undocumented accepted-input regression (REVIEW-B.md
// finding 2). spec.json's page_size now declares a JSON Schema "default":
// "100" annotation, materialized into RuntimeConfig.Config for any genuinely
// absent key by the C3 engine mechanism (engine/read.go's
// materializeConfigDefaults) before query templating ever runs — the exact
// mechanism REVIEW-B.md's adjudication 2 and the gap-loop plan's C3 decision
// specify, restoring legacy's default-100 behavior via a single mechanism
// rather than a static "100" literal.
func TestParityCalendly_PageSizeDefaultsTo100WhenUnset(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	legacySrv, legacyGot := countCaptureServer(t)
	legacy := calendly.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(legacySrv.URL, nil), // page_size deliberately unset
	})

	engSrv, engGot := countCaptureServer(t)
	eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(engSrv.URL, nil), // page_size deliberately unset
	})

	if *legacyGot != "100" {
		t.Fatalf("legacy count = %q, want %q (test fixture bug)", *legacyGot, "100")
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine count = %q, want %q (legacy default-100)", *engGot, *legacyGot)
	}
}

// TestParityCalendly_PageSizeExplicitOverride asserts an explicitly-configured
// page_size still overrides the spec.json default on both sides (C3 defaults
// only fill a genuinely ABSENT key — schema.go's materializeConfigDefaults
// never overrides a key already present in cfg.Config).
func TestParityCalendly_PageSizeExplicitOverride(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	legacySrv, legacyGot := countCaptureServer(t)
	legacy := calendly.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(legacySrv.URL, map[string]string{"page_size": "50"}),
	})

	engSrv, engGot := countCaptureServer(t)
	eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{
		Stream: "scheduled_events",
		Config: calendlyRuntimeConfig(engSrv.URL, map[string]string{"page_size": "50"}),
	})

	if *legacyGot != "50" {
		t.Fatalf("legacy count = %q, want %q (test fixture bug)", *legacyGot, "50")
	}
	if *engGot != *legacyGot {
		t.Fatalf("engine count = %q, want %q (legacy override)", *engGot, *legacyGot)
	}
}

// --- auth header parity ---

func TestParityCalendly_BearerAuthHeaderByteIdentical(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	var legacyAuth, engAuth string
	legacyMux := http.NewServeMux()
	legacyMux.HandleFunc("/users/me", usersMeHandler())
	legacyMux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		legacyAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
	})
	legacySrv := httptest.NewServer(legacyMux)
	t.Cleanup(legacySrv.Close)

	legacy := calendly.New()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(legacySrv.URL, nil)})

	engMux := http.NewServeMux()
	engMux.HandleFunc("/users/me", usersMeHandler())
	engMux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		engAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"collection":[],"pagination":{"count":0,"next_page":null}}`)
	})
	engSrv := httptest.NewServer(engMux)
	t.Cleanup(engSrv.Close)

	eng := engine.New(withCalendlyBaseURL(bundle, engSrv.URL), nil)
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(engSrv.URL, nil)})

	if legacyAuth != "Bearer cal_test_deadbeefdeadbeef" {
		t.Fatalf("legacy Authorization = %q, want Bearer cal_test_deadbeefdeadbeef (test fixture bug)", legacyAuth)
	}
	if engAuth != legacyAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy)", engAuth, legacyAuth)
	}
}

// --- error-path parity: non-2xx mapping ---

func TestParityCalendly_Non2xxErrorPath(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", usersMeHandler())
	mux.HandleFunc("/scheduled_events", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	legacy := calendly.New()
	var legacyErr error
	func() {
		legacyErr = legacy.Read(context.Background(), connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(srv.URL, nil)}, func(connectors.Record) error { return nil })
	}()
	if legacyErr == nil {
		t.Fatal("legacy Read: want error on 401, got nil")
	}

	eng := engine.New(withCalendlyBaseURL(bundle, srv.URL), nil)
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "scheduled_events", Config: calendlyRuntimeConfig(srv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read: want error on 401, got nil")
	}
}

// --- manifest-surface parity ---

// TestParityCalendly_ManifestSurface compares stream NAMES, CURSOR FIELDS, and
// PRIMARY KEYS against legacy's Catalog() (calendly.go:89, backed by
// calendlyStreams() in streams.go) rather than connectors.ManifestOf: legacy
// calendly has no dedicated manifest.go (unlike stripe), so
// ManifestOf(calendly.New()) falls through to the generic zero-Streams path —
// comparing against that would be vacuous. Catalog() is legacy's real,
// meaningful published stream surface. Primary keys ARE compared now
// (gap-loop cycle-1 REVIEW-B.md finding 1/adjudication 1): legacy's `id`
// (idFromURI(uri)) is restored via the `last_path_segment` computed_fields
// filter and `x-primary-key: ["id"]`, so both sides now publish the
// identical `["id"]` primary key for every stream.
func TestParityCalendly_ManifestSurface(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	legacyCatalog, err := calendly.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, nil)
	engManifest := connectors.ManifestOf(eng)

	wantStreams := manifestStreamSurface(legacyCatalog.Streams)
	gotStreams := manifestStreamSurface(engManifest.Streams)
	if missing := missingCalendlyStreamSurface(gotStreams, wantStreams); len(missing) != 0 {
		t.Fatalf("engine manifest missing legacy stream surface entries %+v; got %+v", missing, gotStreams)
	}

	legacyByName := make(map[string][]string, len(legacyCatalog.Streams))
	for _, s := range legacyCatalog.Streams {
		legacyByName[s.Name] = s.PrimaryKey
	}
	for _, s := range engManifest.Streams {
		wantPK, ok := legacyByName[s.Name]
		if !ok {
			continue
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("engine legacy stream %q missing primary key", s.Name)
		}
		if !reflect.DeepEqual(s.PrimaryKey, wantPK) {
			t.Fatalf("engine legacy stream %q primary key = %v, want %v", s.Name, s.PrimaryKey, wantPK)
		}
	}
}

type streamSurface struct {
	Name         string
	CursorFields []string
}

func manifestStreamSurface(streams []connectors.Stream) []streamSurface {
	out := make([]streamSurface, len(streams))
	for i, s := range streams {
		cf := append([]string{}, s.CursorFields...)
		sort.Strings(cf)
		out[i] = streamSurface{Name: s.Name, CursorFields: cf}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func missingCalendlyStreamSurface(got, want []streamSurface) []streamSurface {
	byName := make(map[string]streamSurface, len(got))
	for _, s := range got {
		byName[s.Name] = s
	}
	var missing []streamSurface
	for _, s := range want {
		if gotS, ok := byName[s.Name]; !ok || !reflect.DeepEqual(gotS.CursorFields, s.CursorFields) {
			missing = append(missing, s)
		}
	}
	return missing
}

func missingStrings(got, want []string) []string {
	seen := make(map[string]bool, len(got))
	for _, s := range got {
		seen[s] = true
	}
	var missing []string
	for _, s := range want {
		if !seen[s] {
			missing = append(missing, s)
		}
	}
	return missing
}

// TestParityCalendly_BundleLoadsAndValidates is a fast-failing smoke guard:
// the bundle must load via engine.LoadAll(defs.FS), keep the 5 legacy streams
// by name, and declare the Pass B write actions.
func TestParityCalendly_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	wantStreams := []string{"event_types", "groups", "organization_memberships", "scheduled_events", "users"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if missing := missingStrings(gotStreams, wantStreams); len(missing) != 0 {
		t.Fatalf("bundle streams missing legacy streams %v; got %v", missing, gotStreams)
	}

	if len(bundle.Writes) == 0 {
		t.Fatal("bundle write actions = 0, want Pass B write actions")
	}
	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true")
	}
}

// TestParityCalendly_SpecHasNoDeadConfigKeys guards against the F6 dead-key
// regression fixed by the gap-loop (REVIEW-B.md finding 3): spec.json
// previously declared "max_pages" and "mode", neither consumed by any
// template or engine mechanism ("mode" described a legacy-only fixture
// affordance the bundle explicitly does not have; "max_pages" implied a wired
// page cap that did not exist). Both are deleted; this test fails loudly if
// either dead key (or a new undeclared/unwired one following the same defect
// class) reappears in spec.json's declared properties.
func TestParityCalendly_SpecHasNoDeadConfigKeys(t *testing.T) {
	bundle := loadCalendlyBundle(t)

	wantKeys := []string{"api_key", "base_url", "organization_uri", "page_size", "routing_form_uri", "start_date", "user_uri"}
	gotKeys := append([]string{}, bundle.Spec.Properties()...)
	sort.Strings(gotKeys)
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("spec.json declared properties = %v, want %v (no dead max_pages/mode keys)", gotKeys, wantKeys)
	}
}
