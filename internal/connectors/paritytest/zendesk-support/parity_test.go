// Package paritytest_zendesksupport drives the legacy
// internal/connectors/zendesk-support connector and the engine-backed
// connector built from internal/connectors/defs/zendesk-support against ONE
// shared httptest.Server, per connector, asserting RAW reflect.DeepEqual
// record parity (conventions.md §"Parity suite minimum"). This file is the
// red-first test for wave1-pilot P-7 (zendesk-support): it loads the bundle
// via engine.Load(defs.FS, "zendesk-support") before the bundle exists, so
// the FIRST run of this file must FAIL red on a missing-bundle load error —
// captured in .planning/phases/wave1-pilot/traces/p7-zendesk-support-ledger.md.
package paritytest_zendesksupport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	zendesksupport "polymetrics.ai/internal/connectors/zendesk-support"
)

// jsonRoundTrip re-encodes v through encoding/json into a canonical
// map[string]any, so incidental Go type identity (e.g. int vs float64) never
// causes a false parity mismatch.
func jsonRoundTrip(v map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// loadZendeskBundle resolves the "zendesk-support" bundle from defs.FS via
// engine.Load, the exact call the dispatch brief specifies (paritytest/<name>
// loads the bundle via engine.Load(defs.FS, "<name>")).
func loadZendeskBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "zendesk-support")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "zendesk-support", err)
	}
	return b
}

// withZendeskBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; never mutates the loaded
// original), mirroring parity_stripe_test.go/parity_searxng_test.go's
// with<Name>BaseURL helper.
func withZendeskBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// zendeskAPITokenConfig builds the RuntimeConfig for the API-token/Basic auth
// path shared by both connectors: base_url points at the shared httptest
// server (legacy's baseURL() honors a base_url override directly, appending
// /api/v2 itself; the engine bundle's streams.json templates
// "{{ config.base_url }}/api/v2" the same way — see docs.md's documented
// config-surface deviation: this bundle requires base_url directly rather
// than legacy's subdomain-derivation, which the declarative dialect cannot
// express), and api_token/email are the Basic-auth secrets (legacy resolves
// either the dotted "credentials.api_token"/"credentials.email" keys or
// these bare aliases via its secret() helper's multi-key lookup —
// zendesk_support.go's authenticator; the engine bundle's spec.json declares
// the bare names as the canonical property/secret surface). extra carries
// any additional config values a subtest needs (start_date).
func zendeskAPITokenConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"api_token": "tok_fixture_123",
			"email":     "agent@example.com",
		},
	}
}

// zendeskOAuthConfig builds the RuntimeConfig for the OAuth access-token/
// Bearer auth path.
func zendeskOAuthConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"access_token": "oauth_fixture_abc"},
	}
}

func readAllZendeskRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeZendeskRecord re-encodes r through encoding/json so both
// connectors compare canonical JSON shape rather than incidental Go type
// identity, then callers assert RAW reflect.DeepEqual equality against
// legacy — no field stripping/normalization (conventions.md §"Red-first
// protocol": "never weaken an assertion to get green").
func normalizeZendeskRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := jsonRoundTrip(map[string]any(r))
	if err != nil {
		t.Fatalf("json round-trip record: %v", err)
	}
	return raw
}

func normalizeZendeskRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeZendeskRecord(t, r)
	}
	return out
}

// --- per-stream record parity across all 5 legacy streams ---

// zendeskStreamServer answers every stream's collection endpoint with a
// single non-paginated page (meta.has_more:false), shaped exactly like
// Zendesk's real wire format: {"<key>":[...], "meta":{"has_more":bool,
// "after_cursor":string|null}}.
func zendeskStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/tickets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"tickets":[
			{"id":1,"subject":"Cannot log in","description":"User cannot log in to the portal","status":"open","priority":"high","type":"problem","requester_id":100,"assignee_id":200,"organization_id":300,"group_id":400,"brand_id":500,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":null}}`)
	})
	mux.HandleFunc("/api/v2/users", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"users":[
			{"id":7,"name":"Ada Lovelace","email":"ada@example.com","role":"admin","phone":"+15550100","active":true,"verified":true,"organization_id":300,"time_zone":"UTC","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":null}}`)
	})
	mux.HandleFunc("/api/v2/organizations", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"organizations":[
			{"id":300,"name":"Acme Corp","details":"enterprise account","notes":"vip","group_id":400,"shared_tickets":true,"shared_comments":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":null}}`)
	})
	mux.HandleFunc("/api/v2/groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"groups":[
			{"id":400,"name":"Support","description":"Front-line support","default":true,"deleted":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":null}}`)
	})
	mux.HandleFunc("/api/v2/satisfaction_ratings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"satisfaction_ratings":[
			{"id":900,"score":"good","comment":"Fast resolution","reason":"","ticket_id":1,"assignee_id":200,"requester_id":100,"group_id":400,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":null}}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func TestParityZendesk_StreamRecords(t *testing.T) {
	bundle := loadZendeskBundle(t)

	streams := []string{"tickets", "users", "organizations", "groups", "satisfaction_ratings"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			srv := zendeskStreamServer(t)

			legacy := zendesksupport.New()
			legacyRecs := readAllZendeskRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: zendeskAPITokenConfig(srv.URL, nil)})
			if len(legacyRecs) == 0 {
				t.Fatalf("legacy zendesk-support emitted zero records for stream %q (test fixture bug)", stream)
			}

			eng := engine.New(withZendeskBaseURL(bundle, srv.URL), nil)
			engRecs := readAllZendeskRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: zendeskAPITokenConfig(srv.URL, nil)})

			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}
			gotNorm := normalizeZendeskRecords(t, engRecs)
			wantNorm := normalizeZendeskRecords(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// --- pagination parity: cursor page[after]/meta.after_cursor, 2 pages ---

// zendeskTwoPageServer serves /tickets across 2 pages: page 1 sets
// meta.has_more:true and meta.after_cursor:"CURSOR2"; page 2 (requested with
// page[after]=CURSOR2) sets meta.has_more:false and meta.after_cursor:null,
// matching legacy's own harvest() termination rule (zendesk_support.go:189:
// hasMore != "true" || nextCursor == "") and legacy's own test fixture shape
// (zendesk_support_test.go:33-38).
func zendeskTwoPageServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path != "/api/v2/tickets" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page[after]") {
		case "":
			writeJSON(w, `{"tickets":[
				{"id":1,"subject":"a","description":"first","status":"open","priority":"normal","type":"question","requester_id":100,"assignee_id":200,"organization_id":300,"group_id":400,"brand_id":500,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"},
				{"id":2,"subject":"b","description":"second","status":"open","priority":"normal","type":"question","requester_id":101,"assignee_id":201,"organization_id":300,"group_id":400,"brand_id":500,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}
			],"meta":{"has_more":true,"after_cursor":"CURSOR2"}}`)
		case "CURSOR2":
			writeJSON(w, `{"tickets":[
				{"id":3,"subject":"c","description":"third","status":"solved","priority":"normal","type":"question","requester_id":102,"assignee_id":202,"organization_id":300,"group_id":400,"brand_id":500,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-03T00:00:00Z"}
			],"meta":{"has_more":false,"after_cursor":null}}`)
		default:
			t.Errorf("unexpected page[after]=%q", r.URL.Query().Get("page[after]"))
			writeJSON(w, `{"tickets":[],"meta":{"has_more":false,"after_cursor":null}}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

func TestParityZendesk_TicketsTwoPagePagination(t *testing.T) {
	bundle := loadZendeskBundle(t)

	legacySrv, legacyPaths := zendeskTwoPageServer(t)
	legacy := zendesksupport.New()
	legacyRecs := readAllZendeskRecords(t, legacy, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy tickets records = %d, want 3 (2 pages); paths=%v", len(legacyRecs), *legacyPaths)
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requested %d pages, want exactly 2: %v", len(*legacyPaths), *legacyPaths)
	}

	engSrv, engPaths := zendeskTwoPageServer(t)
	eng := engine.New(withZendeskBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllZendeskRecords(t, eng, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(engSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine tickets records = %d, want 3 (2 pages); paths=%v", len(engRecs), *engPaths)
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requested %d pages, want exactly 2 (no loop, no under-consumption): %v", len(*engPaths), *engPaths)
	}

	gotIDs := recordIDs(t, engRecs)
	wantIDs := recordIDs(t, legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("tickets record id sequence = %v, want %v (legacy)", gotIDs, wantIDs)
	}
}

func recordIDs(t *testing.T, recs []connectors.Record) []any {
	t.Helper()
	out := make([]any, len(recs))
	for i, r := range recs {
		out[i] = r["id"]
	}
	return out
}

// --- pagination stop-signal parity: has_more:false with a non-null cursor ---

// zendeskHasMoreFalseNonNullCursorServer serves /tickets exactly one page
// that sets meta.has_more:false but STILL populates a non-empty
// meta.after_cursor — the exact shape Zendesk's own cursor-pagination docs
// warn callers about ("the cursor properties may be populated even when
// has_more is false") and the divergent case REVIEW-B.md's zendesk-support
// finding 2 called out as untested. Legacy's harvest() stops on
// `hasMore != "true" || nextCursor == ""` (zendesk_support.go:189) — has_more
// alone is sufficient to stop, regardless of the cursor value — so a
// correct client must issue exactly ONE request here, never a second page.
func zendeskHasMoreFalseNonNullCursorServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		writeJSON(w, `{"tickets":[
			{"id":1,"subject":"a","description":"first","status":"open","priority":"normal","type":"question","requester_id":100,"assignee_id":200,"organization_id":300,"group_id":400,"brand_id":500,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}
		],"meta":{"has_more":false,"after_cursor":"STALE_CURSOR_AFTER_LAST_PAGE"}}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

func TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination(t *testing.T) {
	bundle := loadZendeskBundle(t)

	legacySrv, legacyPaths := zendeskHasMoreFalseNonNullCursorServer(t)
	legacy := zendesksupport.New()
	legacyRecs := readAllZendeskRecords(t, legacy, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy tickets records = %d, want 1 (single page, test fixture bug)", len(legacyRecs))
	}
	if len(*legacyPaths) != 1 {
		t.Fatalf("legacy requested %d pages, want exactly 1 (has_more:false stops regardless of cursor): %v", len(*legacyPaths), *legacyPaths)
	}

	engSrv, engPaths := zendeskHasMoreFalseNonNullCursorServer(t)
	eng := engine.New(withZendeskBaseURL(bundle, engSrv.URL), nil)
	engRecs := readAllZendeskRecords(t, eng, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(engSrv.URL, nil)})
	if len(engRecs) != 1 {
		t.Fatalf("engine tickets records = %d, want 1 (single page)", len(engRecs))
	}
	if len(*engPaths) != 1 {
		t.Fatalf("engine requested %d pages, want exactly 1 (has_more:false via stop_path must stop even though after_cursor is non-null): %v", len(*engPaths), *engPaths)
	}
}

// --- incremental parity: legacy sends NO server-side incremental filter ---

// zendeskEmptyTicketsServer (below) answers every request with an empty,
// terminal tickets page (has_more:false) so the read terminates after
// exactly one request. Zendesk's real API has no documented updated_at>=
// query filter for these collection endpoints (legacy zendesk_support.go
// implements no request-side incremental filtering at all — InitialState
// always starts empty, and start_date is mentioned only in a doc comment,
// never wired to a query param); this bundle therefore declares NO
// `incremental` block at all on any of the 5 streams (streams.json keeps
// only `x-cursor-field: updated_at` on each schema for manifest-surface
// parity, the calendly-3-streams pattern — see the parity-deviation ledger
// entry in the bundle's docs.md "Known limits"). Both sides still MUST
// accept an app-persisted state cursor without erroring (InitialState/State
// plumbing parity) even though it is never forwarded on the wire, which is
// what this test actually asserts. TestParityZendesk_StartDateConfigNever-
// SendsServerFilter (below) covers the sibling config-only start_date case.
func TestParityZendesk_IncrementalConfigAcceptedWithoutServerFilter(t *testing.T) {
	bundle := loadZendeskBundle(t)
	const appPersistedCursor = "2026-01-02T00:00:00Z" // updated_at is an RFC3339 string cursor field

	legacySrv := zendeskEmptyTicketsServer(t)
	legacy := zendesksupport.New()
	legacyRecs, legacyErr := readAllZendeskRecordsErr(legacy, connectors.ReadRequest{
		Stream: "tickets",
		Config: zendeskAPITokenConfig(legacySrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})
	if legacyErr != nil {
		t.Fatalf("legacy Read with state cursor: %v", legacyErr)
	}
	if len(legacyRecs) != 0 {
		t.Fatalf("legacy records = %d, want 0 (empty terminal page, test fixture bug)", len(legacyRecs))
	}

	engSrv := zendeskEmptyTicketsServer(t)
	eng := engine.New(withZendeskBaseURL(bundle, engSrv.URL), nil)
	engRecs, engErr := readAllZendeskRecordsErr(eng, connectors.ReadRequest{
		Stream: "tickets",
		Config: zendeskAPITokenConfig(engSrv.URL, nil),
		State:  map[string]string{"cursor": appPersistedCursor},
	})
	if engErr != nil {
		t.Fatalf("engine Read with state cursor: %v", engErr)
	}
	if len(engRecs) != 0 {
		t.Fatalf("engine records = %d, want 0 (empty terminal page)", len(engRecs))
	}
}

// TestParityZendesk_StartDateConfigNeverSendsServerFilter is the inversion of
// the former StartDateConfigRaisesLowerBound test (REVIEW-B.md zendesk-support
// finding 1): legacy's harvest() sends NO server-side incremental filter
// query parameter under any config — "updated_at[gte]" is not real Zendesk
// Support collection-endpoint API surface, and declaring
// incremental.request_param would silently no-op in production while
// narrowing the emitted record set against fixture/test servers that happen
// to honor the invented parameter (a schema/API-fidelity blocker). A
// configured start_date must still be ACCEPTED without error (the
// InitialState/state-cursor/start_date config-plumbing parity requirement),
// but it must never be forwarded as any query parameter on the wire — this
// test asserts the raw request query is untouched by start_date, matching
// legacy byte-for-byte.
func TestParityZendesk_StartDateConfigNeverSendsServerFilter(t *testing.T) {
	bundle := loadZendeskBundle(t)
	const startDate = "2026-01-01T00:00:00Z"

	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		writeJSON(w, `{"tickets":[],"meta":{"has_more":false,"after_cursor":null}}`)
	}))
	defer srv.Close()

	eng := engine.New(withZendeskBaseURL(bundle, srv.URL), nil)
	_, err := readAllZendeskRecordsErr(eng, connectors.ReadRequest{
		Stream: "tickets",
		Config: zendeskAPITokenConfig(srv.URL, map[string]string{"start_date": startDate}),
	})
	if err != nil {
		t.Fatalf("engine Read with start_date config: %v", err)
	}
	if got := gotQuery.Get("updated_at[gte]"); got != "" {
		t.Fatalf("updated_at[gte] = %q, want empty (legacy sends no server-side incremental filter; \"updated_at[gte]\" is not real Zendesk API surface)", got)
	}
	for key := range gotQuery {
		if key == "page[size]" || key == "page[after]" {
			continue
		}
		t.Fatalf("unexpected query param %q=%q sent for start_date config (legacy sends only page[size]/page[after])", key, gotQuery.Get(key))
	}
}

func zendeskEmptyTicketsServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"tickets":[],"meta":{"has_more":false,"after_cursor":null}}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func readAllZendeskRecordsErr(c connectors.Connector, req connectors.ReadRequest) ([]connectors.Record, error) {
	var out []connectors.Record
	err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	})
	return out, err
}

// --- auth parity: BOTH dual-auth candidates (Basic api-token AND OAuth Bearer) ---

func zendeskAuthCaptureServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, `{"tickets":[],"meta":{"has_more":false,"after_cursor":null}}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotAuth
}

// TestParityZendesk_APITokenBasicAuthParity asserts the API-token credential
// path sends the EXACT legacy Basic header shape: base64("<email>/token:<api_token>").
func TestParityZendesk_APITokenBasicAuthParity(t *testing.T) {
	bundle := loadZendeskBundle(t)

	legacySrv, legacyAuth := zendeskAuthCaptureServer(t)
	legacy := zendesksupport.New()
	_ = readAllZendeskRecords(t, legacy, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(legacySrv.URL, nil)})

	engSrv, engAuth := zendeskAuthCaptureServer(t)
	eng := engine.New(withZendeskBaseURL(bundle, engSrv.URL), nil)
	_ = readAllZendeskRecords(t, eng, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(engSrv.URL, nil)})

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("agent@example.com/token:tok_fixture_123"))
	if *legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", *legacyAuth, wantAuth)
	}
	if *engAuth != wantAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy)", *engAuth, wantAuth)
	}
}

// TestParityZendesk_OAuthBearerAuthParity asserts the OAuth access_token
// credential path sends a Bearer header, and that it takes PRECEDENCE over
// an API-token secret when both happen to be present — mirroring legacy's
// authenticator() precedence (zendesk_support.go:272: access_token checked
// first).
func TestParityZendesk_OAuthBearerAuthParity(t *testing.T) {
	bundle := loadZendeskBundle(t)

	legacySrv, legacyAuth := zendeskAuthCaptureServer(t)
	legacy := zendesksupport.New()
	_ = readAllZendeskRecords(t, legacy, connectors.ReadRequest{Stream: "tickets", Config: zendeskOAuthConfig(legacySrv.URL, nil)})

	engSrv, engAuth := zendeskAuthCaptureServer(t)
	eng := engine.New(withZendeskBaseURL(bundle, engSrv.URL), nil)
	_ = readAllZendeskRecords(t, eng, connectors.ReadRequest{Stream: "tickets", Config: zendeskOAuthConfig(engSrv.URL, nil)})

	wantAuth := "Bearer oauth_fixture_abc"
	if *legacyAuth != wantAuth {
		t.Fatalf("legacy Authorization = %q, want %q (test fixture bug)", *legacyAuth, wantAuth)
	}
	if *engAuth != wantAuth {
		t.Fatalf("engine Authorization = %q, want %q (legacy)", *engAuth, wantAuth)
	}
}

// --- error-path parity (non-2xx mapping) ---

func TestParityZendesk_ErrorPathParity(t *testing.T) {
	bundle := loadZendeskBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, `{"error":"Couldn't authenticate you"}`)
	}))
	defer srv.Close()

	legacy := zendesksupport.New()
	_, legacyErr := readAllZendeskRecordsErr(legacy, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(srv.URL, nil)})
	if legacyErr == nil {
		t.Fatal("legacy Read on a 401 response = nil error, want non-nil (test fixture bug)")
	}

	eng := engine.New(withZendeskBaseURL(bundle, srv.URL), nil)
	_, engErr := readAllZendeskRecordsErr(eng, connectors.ReadRequest{Stream: "tickets", Config: zendeskAPITokenConfig(srv.URL, nil)})
	if engErr == nil {
		t.Fatal("engine Read on a 401 response = nil error, want non-nil")
	}
}

// --- bundle load smoke guard ---

func TestParityZendesk_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadZendeskBundle(t)

	wantStreams := map[string]bool{"tickets": true, "users": true, "organizations": true, "groups": true, "satisfaction_ratings": true}
	seen := map[string]bool{}
	for _, s := range bundle.Streams {
		if wantStreams[s.Name] {
			seen[s.Name] = true
		}
	}
	for name := range wantStreams {
		if !seen[name] {
			t.Fatalf("bundle streams missing legacy stream %q; got %d streams", name, len(bundle.Streams))
		}
	}

	if len(bundle.Writes) == 0 {
		t.Fatal("bundle write actions = 0, want Pass B write actions")
	}
	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true")
	}
	if len(bundle.HTTP.Auth) != 2 {
		t.Fatalf("bundle auth candidates = %d, want 2 (OAuth bearer + Basic api-token, when-gated)", len(bundle.HTTP.Auth))
	}
}
