package mondayparity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/monday" // registers the StreamHook/CheckHook via init()
	"polymetrics.ai/internal/connectors/monday"
)

// This file is the migration parity suite for the monday bundle (PLAN.md
// P-8, wave1-pilot): monday is the StreamHook pilot (SPEC §5.5) — its
// GraphQL POST reads with in-body pagination are a Tier-2 trigger
// (StreamSpec.Body is deliberately unwired in engine/read.go this wave).
// Both the legacy hand-written monday.Connector (internal/connectors/monday,
// read-only reference) and the engine-backed connector
// (engine.New(bundle, hooks.HooksFor("monday"))) are driven against the SAME
// httptest.Server; RAW reflect.DeepEqual record equality is the parity bar.
// This is the red-first test: it fails to even compile/load until
// defs/monday and hooks/monday exist.

func loadMondayBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("engine.LoadAll(defs.FS): %v", err)
	}
	for _, b := range bundles {
		if b.Name == "monday" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "monday", names)
	return engine.Bundle{}
}

func withMondayBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func newMondayEngineConnector(b engine.Bundle) connectors.Connector {
	return engine.New(b, engine.HooksFor("monday"))
}

// mondayRuntimeConfig builds a RuntimeConfig that authenticates BOTH sides
// identically despite their different secret-key conventions: legacy's
// mondaySecret (monday.go:407-422) reads the catalog's dotted
// "credentials.api_token" form (falling back to a bare "api_token"), while
// this bundle's spec.json declares a flat "api_token" property (the
// dialect's Vars.Secrets lookup is a literal key match, no dotted-path
// walk for secrets) — both keys are set to the SAME token value so
// TestParityMonday_BoardsStreamRecordsAndAuth's Authorization-header
// assertion is genuinely comparing the same credential on both sides, not
// an artifact of one side silently falling back to "no auth".
func mondayRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"credentials.api_token": "tok_abc123", "api_token": "tok_abc123"},
	}
}

func readAllMondayRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeMondayRecord re-encodes r through encoding/json so both
// connectors compare on canonical JSON shape rather than incidental Go type
// identity (e.g. string vs json.Number for ids that legacy stringifies).
func normalizeMondayRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	return out
}

func normalizeMondayRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeMondayRecord(t, r)
	}
	return out
}

// graphqlRequest is the minimal shape of a monday.com GraphQL POST body.
type graphqlRequest struct {
	Query string `json:"query"`
}

// --- boards: page-number pagination + raw-token auth parity ---

func mondayBoardsServer(t *testing.T, sawAuth, sawVersion *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*sawAuth = r.Header.Get("Authorization")
		*sawVersion = r.Header.Get("API-Version")
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if strings.TrimSuffix(r.URL.Path, "/") != "/v2" {
			http.NotFound(w, r)
			return
		}
		var body graphqlRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "page: 1"):
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","name":"Board One","state":"active","board_kind":"public","description":"d1","type":"board","updated_at":"2026-01-01T00:00:00Z","workspace_id":"ws_1"},{"id":"2","name":"Board Two","state":"active","board_kind":"private","description":"d2","type":"board","updated_at":"2026-01-02T00:00:00Z","workspace_id":"ws_1"}]}}`))
		case strings.Contains(body.Query, "page: 2"):
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"3","name":"Board Three","state":"archived","board_kind":"public","description":"d3","type":"board","updated_at":"2026-01-03T00:00:00Z","workspace_id":"ws_1"}]}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"boards":[]}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityMonday_BoardsStreamRecordsAndAuth(t *testing.T) {
	bundle := loadMondayBundle(t)

	var legacyAuth, legacyVersion string
	legacySrv := mondayBoardsServer(t, &legacyAuth, &legacyVersion)
	legacy := monday.New()
	legacyRecs := readAllMondayRecords(t, legacy, connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(legacySrv.URL+"/v2", map[string]string{"page_size": "2"}),
	})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy records = %d, want 3 (test fixture bug)", len(legacyRecs))
	}
	if legacyAuth != "tok_abc123" {
		t.Fatalf("legacy Authorization = %q, want raw token (test fixture bug)", legacyAuth)
	}

	var engAuth, engVersion string
	engSrv := mondayBoardsServer(t, &engAuth, &engVersion)
	eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL+"/v2"))
	engRecs := readAllMondayRecords(t, eng, connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(engSrv.URL+"/v2", map[string]string{"page_size": "2"}),
	})

	if engAuth != "tok_abc123" {
		t.Fatalf("engine Authorization = %q, want raw token tok_abc123 (no Bearer prefix, legacy parity)", engAuth)
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeMondayRecords(t, engRecs)
	wantNorm := normalizeMondayRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- items: cursor-based next_items_page pagination parity ---

func mondayItemsServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "next_items_page"):
			if !strings.Contains(body.Query, "CUR_2") {
				t.Errorf("next_items_page query missing cursor: %q", body.Query)
			}
			_, _ = w.Write([]byte(`{"data":{"next_items_page":{"cursor":null,"items":[{"id":"i3","name":"Item Three","state":"active","created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-03T00:00:00Z","group":{"id":"grp_1","title":"Group One"},"board":{"id":"1","name":"Board One"}}]}}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"boards":[{"id":"1","items_page":{"cursor":"CUR_2","items":[{"id":"i1","name":"Item One","state":"active","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","group":{"id":"grp_1","title":"Group One"},"board":{"id":"1","name":"Board One"}},{"id":"i2","name":"Item Two","state":"active","created_at":"2026-01-02T00:00:00Z","updated_at":"2026-01-02T00:00:00Z","group":{"id":"grp_1","title":"Group One"},"board":{"id":"1","name":"Board One"}}]}}]}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityMonday_ItemsCursorPagination(t *testing.T) {
	bundle := loadMondayBundle(t)

	legacySrv := mondayItemsServer(t)
	legacy := monday.New()
	legacyRecs := readAllMondayRecords(t, legacy, connectors.ReadRequest{
		Stream: "items",
		Config: mondayRuntimeConfig(legacySrv.URL, map[string]string{"page_size": "2"}),
	})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy items = %d, want 3 (test fixture bug)", len(legacyRecs))
	}

	engSrv := mondayItemsServer(t)
	eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL))
	engRecs := readAllMondayRecords(t, eng, connectors.ReadRequest{
		Stream: "items",
		Config: mondayRuntimeConfig(engSrv.URL, map[string]string{"page_size": "2"}),
	})
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("items = %d, want %d (legacy, cursor pagination)", len(engRecs), len(legacyRecs))
	}

	gotNorm := normalizeMondayRecords(t, engRecs)
	wantNorm := normalizeMondayRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("item %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- users/teams/tags: simple page-number streams, one page each ---

func mondaySimpleServer(t *testing.T, root, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphqlRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(req.Query, root) {
			t.Errorf("query did not target %s: %q", root, req.Query)
		}
		switch {
		case strings.Contains(req.Query, "page: 1"):
			_, _ = w.Write([]byte(body))
		default:
			_, _ = w.Write([]byte(`{"data":{"` + root + `":[]}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityMonday_UsersTeamsTagsStreams(t *testing.T) {
	bundle := loadMondayBundle(t)

	cases := []struct {
		stream string
		root   string
		body   string
	}{
		{"users", "users", `{"data":{"users":[{"id":"u1","name":"User One","email":"user1@example.com","enabled":true,"is_admin":false,"is_guest":false,"is_pending":false,"created_at":"2026-01-01T00:00:00Z"}]}}`},
		{"teams", "teams", `{"data":{"teams":[{"id":"t1","name":"Team One","picture_url":"https://example.com/t1.png"}]}}`},
		{"tags", "tags", `{"data":{"tags":[{"id":"tag1","name":"tag-one","color":"blue"}]}}`},
	}

	for _, tc := range cases {
		t.Run(tc.stream, func(t *testing.T) {
			legacySrv := mondaySimpleServer(t, tc.root, tc.body)
			legacy := monday.New()
			legacyRecs := readAllMondayRecords(t, legacy, connectors.ReadRequest{
				Stream: tc.stream,
				Config: mondayRuntimeConfig(legacySrv.URL, nil),
			})
			if len(legacyRecs) != 1 {
				t.Fatalf("legacy %s records = %d, want 1 (test fixture bug)", tc.stream, len(legacyRecs))
			}

			engSrv := mondaySimpleServer(t, tc.root, tc.body)
			eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL))
			engRecs := readAllMondayRecords(t, eng, connectors.ReadRequest{
				Stream: tc.stream,
				Config: mondayRuntimeConfig(engSrv.URL, nil),
			})
			if len(engRecs) != 1 {
				t.Fatalf("engine %s records = %d, want 1", tc.stream, len(engRecs))
			}

			gotNorm := normalizeMondayRecords(t, engRecs)
			wantNorm := normalizeMondayRecords(t, legacyRecs)
			if !reflect.DeepEqual(gotNorm[0], wantNorm[0]) {
				t.Fatalf("%s record mismatch:\nengine:  %+v\nlegacy:  %+v", tc.stream, gotNorm[0], wantNorm[0])
			}
		})
	}
}

// --- error-path parity: GraphQL errors surfaced as a real error, not an
// empty read (monday returns HTTP 200 even for query errors). ---

func TestParityMonday_GraphQLErrorSurfacesAsError(t *testing.T) {
	bundle := loadMondayBundle(t)

	errBody := `{"errors":[{"message":"Some field arguments are invalid"}]}`
	newErrServer := func() *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(errBody))
		}))
		t.Cleanup(srv.Close)
		return srv
	}

	legacySrv := newErrServer()
	legacy := monday.New()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(legacySrv.URL, nil),
	}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read did not error on a GraphQL error envelope (test fixture bug)")
	}

	engSrv := newErrServer()
	eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(engSrv.URL, nil),
	}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read did not error on a GraphQL error envelope, want an error (StreamHook must surface monday's HTTP-200 GraphQL errors)")
	}
}

// --- Check parity: monday's `query { me { id } }` GraphQL POST check ---

func TestParityMonday_CheckSendsMeQuery(t *testing.T) {
	bundle := loadMondayBundle(t)

	newCheckServer := func(sawQuery *string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body graphqlRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			*sawQuery = body.Query
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"me":{"id":"1"}}}`))
		}))
		t.Cleanup(srv.Close)
		return srv
	}

	var legacyQuery string
	legacySrv := newCheckServer(&legacyQuery)
	legacy := monday.New()
	if err := legacy.Check(context.Background(), mondayRuntimeConfig(legacySrv.URL, nil)); err != nil {
		t.Fatalf("legacy Check: %v", err)
	}
	if !strings.Contains(legacyQuery, "me") {
		t.Fatalf("legacy check query = %q, want a query containing me (test fixture bug)", legacyQuery)
	}

	var engQuery string
	engSrv := newCheckServer(&engQuery)
	eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL))
	if err := eng.Check(context.Background(), mondayRuntimeConfig(engSrv.URL, nil)); err != nil {
		t.Fatalf("engine Check: %v", err)
	}
	if !strings.Contains(engQuery, "me") {
		t.Fatalf("engine check query = %q, want a query containing me (CheckHook must port legacy's `query { me { id } }`)", engQuery)
	}
}

// --- API-Version optional header parity ---

func TestParityMonday_APIVersionHeaderOptional(t *testing.T) {
	bundle := loadMondayBundle(t)

	var legacyVersion, engVersion string
	legacySrv := mondayBoardsServer(t, new(string), &legacyVersion)
	legacy := monday.New()
	_ = readAllMondayRecords(t, legacy, connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(legacySrv.URL+"/v2", map[string]string{"api_version": "2024-01"}),
	})
	if legacyVersion != "2024-01" {
		t.Fatalf("legacy API-Version = %q, want 2024-01 (test fixture bug)", legacyVersion)
	}

	engSrv := mondayBoardsServer(t, new(string), &engVersion)
	eng := newMondayEngineConnector(withMondayBaseURL(bundle, engSrv.URL+"/v2"))
	_ = readAllMondayRecords(t, eng, connectors.ReadRequest{
		Stream: "boards",
		Config: mondayRuntimeConfig(engSrv.URL+"/v2", map[string]string{"api_version": "2024-01"}),
	})
	if engVersion != "2024-01" {
		t.Fatalf("engine API-Version = %q, want 2024-01 (legacy parity)", engVersion)
	}
}

// --- bundle load smoke guard ---

func TestParityMonday_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadMondayBundle(t)

	wantStreams := []string{"boards", "items", "tags", "teams", "users"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (monday is read-only)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (monday has no mutation API)")
	}
}
