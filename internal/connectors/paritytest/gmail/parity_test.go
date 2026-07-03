// Package paritytest_gmail is the engine-vs-legacy parity suite for the
// gmail pilot migration (wave1-pilot P-10, PLAN.md; SPEC.md §5.7). This was
// the RED-FIRST assertion (conventions.md/TEST-PLAN.md §5, captured in
// .planning/phases/wave1-pilot/traces/p10-gmail-ledger.md): before
// internal/connectors/defs/gmail existed, engine.Load(defs.FS, "gmail")
// failed with "missing required file metadata.json".
//
// Both the legacy hand-written gmail.Connector (internal/connectors/gmail,
// read-only reference) and the engine-backed connector
// (engine.New(bundle, engine.HooksFor("gmail"))) are driven against the SAME
// httptest Gmail-data server AND the SAME httptest TLS token-exchange
// server (THREAT-MODEL.md Delta 2: the hook requires token_url to be
// https); RAW connectors.Record reflect.DeepEqual equality is the parity
// bar, matching internal/connectors/engine/parity_stripe_test.go and
// paritytest/monday's precedent for a hook-backed pilot
// (engine.New(bundle, engine.HooksFor("monday"))).
package paritytest_gmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/gmail"
	gmailhook "polymetrics.ai/internal/connectors/hooks/gmail" // registers the AuthHook via init(); also gives this test direct access to Hooks.Client for TLS trust
)

// loadGmailBundle resolves the "gmail" bundle from defs.FS via engine.Load,
// the same discovery path TestConformance and every other production caller
// uses.
func loadGmailBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "gmail")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "gmail", err)
	}
	return b
}

// withGmailBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original, mirroring parity_stripe_test.go/parity_searxng_test.go).
func withGmailBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared token-exchange server (both sides authenticate against it) ---

// tokenServer stands in for Google's OAuth token endpoint. It MUST be a TLS
// server: legacy's own token_url default is https, and the gmail AuthHook
// fails closed on a non-https token_url (THREAT-MODEL.md Delta 2) — so a
// plain httptest.Server would make the engine side fail before ever
// resolving an access token, which is not a fair parity comparison. Returns
// the server, the *http.Client that trusts its self-signed cert (wired into
// BOTH connectors' Client field so neither side needs -insecure workarounds
// beyond the shared test client), and a hit counter.
func tokenServer(t *testing.T, accessToken string) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token server: parse form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("token server: grant_type = %q, want refresh_token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, srv.Client(), &hits
}

// gmailRuntimeConfig builds the connectors.RuntimeConfig shared by both
// connectors: base_url/token_url point at the shared httptest servers, and
// the three OAuth secrets are synthetic placeholders (never a real-looking
// credential, per THREAT-MODEL §4).
func gmailRuntimeConfig(baseURL, tokenURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":  baseURL,
		"token_url": tokenURL,
		"scopes":    "https://www.googleapis.com/auth/gmail.readonly",
		"user_id":   "me",
		"page_size": "100",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"client_id":            "client-id-fixture",
			"client_secret":        "client-secret-fixture",
			"client_refresh_token": "refresh-token-fixture",
		},
	}
}

func readAllGmailRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

// normalizeGmailRecord re-encodes r through encoding/json with UseNumber so
// legacy's native Go int64/string map-literal values and the engine's
// json.Number-preserving decode compare equal on numeric fields (mirrors
// parity_stripe_test.go's normalizeRecord).
func normalizeGmailRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode record: %v", err)
	}
	return out
}

func normalizeGmailRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeGmailRecord(t, r)
	}
	return out
}

// --- Gmail data server fixtures (2-page pageToken/nextPageToken cursor) ---

func gmailDataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("pageToken") {
		case "":
			writeJSON(w, `{"messages":[
				{"id":"msg_1","threadId":"thread_1"},
				{"id":"msg_2","threadId":"thread_2"}
			],"nextPageToken":"page2token","resultSizeEstimate":3}`)
		case "page2token":
			writeJSON(w, `{"messages":[{"id":"msg_3","threadId":"thread_3"}],"resultSizeEstimate":3}`)
		default:
			t.Errorf("unexpected pageToken=%q for messages", r.URL.Query().Get("pageToken"))
			writeJSON(w, `{"messages":[]}`)
		}
	})

	mux.HandleFunc("/users/me/threads", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"threads":[
			{"id":"thread_1","snippet":"hello world","historyId":"1001"}
		]}`)
	})

	mux.HandleFunc("/users/me/drafts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"drafts":[
			{"id":"draft_1","message":{"id":"msg_draft_1","threadId":"thread_draft_1"}}
		]}`)
	})

	mux.HandleFunc("/users/me/labels", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"labels":[
			{"id":"INBOX","name":"INBOX","type":"system","messageListVisibility":"show","labelListVisibility":"labelShow","messagesTotal":10,"messagesUnread":2,"threadsTotal":8,"threadsUnread":1}
		]}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// newGmailLegacyConnector builds the legacy connector wired with client (the
// TLS-trusting client for the shared token server) so the OAuth token
// exchange succeeds against tokenServer's self-signed cert.
func newGmailLegacyConnector(client *http.Client) connectors.Connector {
	return gmail.Connector{Client: client}
}

// newGmailEngineConnector builds the engine-backed connector with the real
// registered AuthHook (mirrors paritytest/monday's
// engine.New(b, engine.HooksFor("monday")) precedent) and a Client override
// is NOT threaded through the hook itself (the hook builds its own
// *http.Client unless Client is set) — the TLS trust is instead supplied via
// hooks.Hooks.Client through a package-level registration override, so this
// helper re-registers a client-carrying Hooks instance for the duration of
// each subtest via hooksWithClient.
func newGmailEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// --- per-stream record parity across all 4 streams ---

func TestParityGmail_StreamRecords(t *testing.T) {
	bundle := loadGmailBundle(t)

	streams := []string{"messages", "threads", "drafts", "labels"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			dataSrv := gmailDataServer(t)
			tokenSrv, tlsClient, _ := tokenServer(t, "tok_"+stream)

			legacy := newGmailLegacyConnector(tlsClient)
			legacyCfg := gmailRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			legacyRecs := readAllGmailRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			hooksHook := newHooksWithClient(tlsClient)
			eng := newGmailEngineConnector(withGmailBaseURL(bundle, dataSrv.URL), hooksHook)
			engCfg := gmailRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			engRecs := readAllGmailRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy gmail emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeGmailRecords(t, engRecs)
			wantNorm := normalizeGmailRecords(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// --- pagination parity: messages 2-page pageToken/nextPageToken ---

func TestParityGmail_MessagesTwoPagePagination(t *testing.T) {
	bundle := loadGmailBundle(t)

	dataSrv := gmailDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_messages")

	legacy := newGmailLegacyConnector(tlsClient)
	legacyRecs := readAllGmailRecords(t, legacy, connectors.ReadRequest{Stream: "messages", Config: gmailRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy messages records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := newGmailEngineConnector(withGmailBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engRecs := readAllGmailRecords(t, eng, connectors.ReadRequest{Stream: "messages", Config: gmailRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine messages records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotIDs := recordIDs(t, engRecs)
	wantIDs := recordIDs(t, legacyRecs)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("messages record id sequence = %v, want %v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"msg_1", "msg_2", "msg_3"}) {
		t.Fatalf("messages record id sequence = %v, want [msg_1 msg_2 msg_3]", gotIDs)
	}
}

func recordIDs(t *testing.T, recs []connectors.Record) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		id, _ := r["id"].(string)
		out[i] = id
	}
	return out
}

// TestParityGmail_LabelsUnpaginatedSinglePage: legacy's paginated=false
// routing table entry (streams.go:28) and this bundle's stream-level
// pagination:{"type":"none"} override must both make exactly ONE request
// for labels, never a page-2 fetch.
func TestParityGmail_LabelsUnpaginatedSinglePage(t *testing.T) {
	bundle := loadGmailBundle(t)

	var legacyHits, engHits int32
	mux := func(hits *int32) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(hits, 1)
			writeJSON(w, `{"labels":[{"id":"INBOX","name":"INBOX","type":"system","messageListVisibility":"show","labelListVisibility":"labelShow","messagesTotal":10,"messagesUnread":2,"threadsTotal":8,"threadsUnread":1}]}`)
		}))
		return srv
	}
	legacySrv := mux(&legacyHits)
	t.Cleanup(legacySrv.Close)
	engSrv := mux(&engHits)
	t.Cleanup(engSrv.Close)

	tokenSrv, tlsClient, _ := tokenServer(t, "tok_labels")

	legacy := newGmailLegacyConnector(tlsClient)
	legacyRecs := readAllGmailRecords(t, legacy, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(legacySrv.URL, tokenSrv.URL, nil)})
	if legacyHits != 1 {
		t.Fatalf("legacy labels request count = %d, want 1 (unpaginated)", legacyHits)
	}

	eng := newGmailEngineConnector(withGmailBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
	engRecs := readAllGmailRecords(t, eng, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(engSrv.URL, tokenSrv.URL, nil)})
	if engHits != 1 {
		t.Fatalf("engine labels request count = %d, want 1 (unpaginated)", engHits)
	}

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("labels record count = %d, want %d (legacy)", len(engRecs), len(legacyRecs))
	}
	gotNorm := normalizeGmailRecords(t, engRecs)
	wantNorm := normalizeGmailRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("labels records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType pins the
// RESOLVED state (gap-loop cycle-1 engine mini-wave item 1, REVIEW-A.md
// adjudication A1: typed computed_fields extraction; formerly
// TestParityGmail_ComputedFieldsStringifyLabelCountFields, which asserted
// the PRE-increment stringified form). messagesTotal/messagesUnread/
// threadsTotal/threadsUnread are all sourced via a BARE single
// "{{ record.messagesTotal }}"-shaped computed_fields template (camelCase
// rename to the schema's snake_case names, no filter, no surrounding
// literal text) — the engine now copies the raw JSON value straight
// through instead of stringifying it via Interpolate, so these 4 fields
// are native json.Number on the engine side too, matching legacy's own
// connsdk UseNumber-decoded json.Number exactly (connsdk/extract.go's
// decodeJSON). RAW type-identical equality, not a coercing/stringifying
// comparison (mirrors paritytest/chargebee's identical A1 finding, now also
// RESOLVED there).
func TestParityGmail_ComputedFieldsPreserveLabelCountFieldsNativeType(t *testing.T) {
	bundle := loadGmailBundle(t)
	srv := gmailDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_stringify_check")

	legacy := newGmailLegacyConnector(tlsClient)
	legacyRecs := readAllGmailRecords(t, legacy, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(srv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) == 0 {
		t.Fatal("legacy emitted zero labels records (test fixture bug)")
	}
	if n, ok := legacyRecs[0]["messages_total"].(json.Number); !ok || n.String() != "10" {
		t.Fatalf("legacy labels[0].messages_total = %#v (%T), want json.Number(\"10\") (test fixture bug: connsdk's UseNumber decode)", legacyRecs[0]["messages_total"], legacyRecs[0]["messages_total"])
	}

	eng := newGmailEngineConnector(withGmailBaseURL(bundle, srv.URL), newHooksWithClient(tlsClient))
	engRecs := readAllGmailRecords(t, eng, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(srv.URL, tokenSrv.URL, nil)})
	if len(engRecs) == 0 {
		t.Fatal("engine emitted zero labels records")
	}
	got, ok := engRecs[0]["messages_total"].(json.Number)
	if !ok {
		t.Fatalf("engine labels[0].messages_total = %#v (%T), want json.Number (native type, typed computed_fields extraction)", engRecs[0]["messages_total"], engRecs[0]["messages_total"])
	}
	if got.String() != "10" {
		t.Fatalf("engine labels[0].messages_total = %q, want %q", got.String(), "10")
	}
}

// --- auth parity: Authorization header after refresh ---

// TestParityGmail_AuthorizationHeaderAfterRefresh asserts BOTH connectors
// send the identical "Authorization: Bearer <access_token>" header (derived
// from the SAME token-exchange response) on the request to the Gmail data
// API — the exact header-parity bar TEST-PLAN.md §1 calls for.
func TestParityGmail_AuthorizationHeaderAfterRefresh(t *testing.T) {
	bundle := loadGmailBundle(t)
	const accessToken = "tok_shared_fixture_value"

	var legacyAuthHeader, engAuthHeader string

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"labels":[]}`)
	}))
	t.Cleanup(legacySrv.Close)

	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"labels":[]}`)
	}))
	t.Cleanup(engSrv.Close)

	tokenSrv, tlsClient, hits := tokenServer(t, accessToken)

	legacy := newGmailLegacyConnector(tlsClient)
	_ = readAllGmailRecords(t, legacy, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(legacySrv.URL, tokenSrv.URL, nil)})

	eng := newGmailEngineConnector(withGmailBaseURL(bundle, engSrv.URL), newHooksWithClient(tlsClient))
	_ = readAllGmailRecords(t, eng, connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(engSrv.URL, tokenSrv.URL, nil)})

	wantHeader := "Bearer " + accessToken
	if legacyAuthHeader != wantHeader {
		t.Fatalf("legacy Authorization header = %q, want %q (test fixture bug)", legacyAuthHeader, wantHeader)
	}
	if engAuthHeader != wantHeader {
		t.Fatalf("engine Authorization header = %q, want %q (legacy, same shared token exchange)", engAuthHeader, wantHeader)
	}
	if *hits != 2 {
		t.Fatalf("token endpoint hits = %d, want 2 (one refresh exchange per connector)", *hits)
	}
}

// TestParityGmail_TokenEndpointFailureSurfacesAsAuthError asserts a token
// endpoint failure surfaces as an error on BOTH sides (never a silent
// unauthenticated request to the Gmail data API) — TEST-PLAN.md §1's gmail
// row: "token-endpoint failure surfaces as auth error, not silent unauth
// request".
func TestParityGmail_TokenEndpointFailureSurfacesAsAuthError(t *testing.T) {
	bundle := loadGmailBundle(t)

	var dataHits int32
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&dataHits, 1)
		writeJSON(w, `{"labels":[]}`)
	}))
	t.Cleanup(dataSrv.Close)

	failingTokenSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}))
	t.Cleanup(failingTokenSrv.Close)
	tlsClient := failingTokenSrv.Client()

	legacy := newGmailLegacyConnector(tlsClient)
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read succeeded despite a failing token endpoint, want an error (test fixture bug)")
	}

	eng := newGmailEngineConnector(withGmailBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "labels", Config: gmailRuntimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read succeeded despite a failing token endpoint, want an error")
	}

	if dataHits != 0 {
		t.Fatalf("Gmail data API received %d requests despite a failed token exchange, want 0 (no silent unauthenticated fallback)", dataHits)
	}
}

// --- write parity: legacy stays read-only; the engine bundle now writes (Pass B) ---

// TestParityGmail_WriteUnsupportedOnBothSides was renamed in spirit by the
// Pass B full-surface expansion but keeps its original name for git-blame
// continuity: legacy internal/connectors/gmail remains permanently read-only
// (gmail.go:191-192's ErrUnsupportedOperation is never going to change,
// since the legacy package is frozen reference code until wave6's registry
// flip), while the ENGINE-BACKED bundle now declares capabilities.write:
// true and 35 write actions (writes.json). This is an intentional,
// documented capability divergence between legacy and the migrated bundle
// (docs.md "Overview"/"Write actions & risks") — the two sides are no
// longer expected to agree on write support, only legacy's own read-only
// behavior is pinned here.
func TestParityGmail_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadGmailBundle(t)

	legacy := gmail.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug; legacy internal/connectors/gmail is frozen read-only reference code)")
	}

	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true (Pass B full-surface expansion declares 35 write actions)")
	}
	if len(bundle.Writes) == 0 {
		t.Fatal("bundle write actions = 0, want 35 (Pass B full-surface expansion writes.json)")
	}
	if len(bundle.Writes) != 35 {
		t.Fatalf("bundle write actions = %d, want 35", len(bundle.Writes))
	}
}

// --- manifest-surface parity ---

// TestParityGmail_ManifestSurface pins the ENGINE bundle's own stream/write
// manifest surface directly (Pass B full-surface expansion): legacy's
// Catalog() only ever described the original wave1-pilot 4-stream,
// read-only surface and is no longer a meaningful parity oracle for the
// bundle's now-much-larger surface (10 streams, 35 write actions) — legacy
// itself was never extended to describe history/filters/send_as/delegates/
// forwarding_addresses/profile or any write action, since it is frozen
// reference code. The 4 original streams' primary keys are still asserted
// against legacy's Catalog() below for continuity; the 6 new streams and
// all 35 write actions are asserted directly against the bundle.
func TestParityGmail_ManifestSurface(t *testing.T) {
	bundle := loadGmailBundle(t)

	legacyCatalog, err := gmail.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor("gmail"))
	engManifest := connectors.ManifestOf(eng)

	gotStreams := gmailManifestStreamSurface(engManifest.Streams)
	wantStreams := []gmailStreamSurface{
		{Name: "delegates", PrimaryKey: []string{"delegate_email"}},
		{Name: "drafts", PrimaryKey: []string{"id"}},
		{Name: "filters", PrimaryKey: []string{"id"}},
		{Name: "forwarding_addresses", PrimaryKey: []string{"forwarding_email"}},
		{Name: "history", PrimaryKey: []string{"id"}},
		{Name: "labels", PrimaryKey: []string{"id"}},
		{Name: "messages", PrimaryKey: []string{"id"}},
		{Name: "profile", PrimaryKey: []string{"email_address"}},
		{Name: "send_as", PrimaryKey: []string{"send_as_email"}},
		{Name: "threads", PrimaryKey: []string{"id"}},
	}
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v", gotStreams, wantStreams)
	}

	// The 4 original wave1-pilot streams' primary keys still match legacy's
	// own Catalog() exactly (continuity check, not a full-surface one).
	legacyStreams := gmailManifestStreamSurface(legacyCatalog.Streams)
	legacyByName := make(map[string]gmailStreamSurface, len(legacyStreams))
	for _, s := range legacyStreams {
		legacyByName[s.Name] = s
	}
	for _, name := range []string{"drafts", "labels", "messages", "threads"} {
		legacyStream, ok := legacyByName[name]
		if !ok {
			t.Fatalf("legacy Catalog() missing expected original stream %q", name)
		}
		var engStream gmailStreamSurface
		for _, s := range gotStreams {
			if s.Name == name {
				engStream = s
			}
		}
		if !reflect.DeepEqual(engStream, legacyStream) {
			t.Fatalf("stream %q surface = %+v, want %+v (legacy Catalog())", name, engStream, legacyStream)
		}
	}

	if len(engManifest.WriteActions) != 35 {
		t.Fatalf("engine write actions = %d, want 35 (Pass B full-surface expansion)", len(engManifest.WriteActions))
	}
}

type gmailStreamSurface struct {
	Name       string
	PrimaryKey []string
}

func gmailManifestStreamSurface(streams []connectors.Stream) []gmailStreamSurface {
	out := make([]gmailStreamSurface, len(streams))
	for i, s := range streams {
		out[i] = gmailStreamSurface{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

// TestParityGmail_BundleLoadsAndValidates pins the Pass B full-surface
// expansion's bundle shape: 10 streams (up from the wave1-pilot's original
// 4), capabilities.write: true, and 35 write actions. Only the "history"
// stream declares an incremental block — messages/threads/drafts/labels are
// still full_refresh-only (legacy's own doc comment, streams.go:31-34,
// documents no publishable cursor field on THOSE 4 list endpoints; history
// is a genuinely different, cursor-bearing endpoint Gmail added specifically
// for incremental sync, see docs.md Streams notes).
func TestParityGmail_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadGmailBundle(t)

	wantStreams := []string{
		"delegates", "drafts", "filters", "forwarding_addresses", "history",
		"labels", "messages", "profile", "send_as", "threads",
	}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 35 {
		t.Fatalf("bundle write actions = %d, want 35 (Pass B full-surface expansion)", len(bundle.Writes))
	}
	if !bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = false, want true (Pass B full-surface expansion)")
	}
	wantIncremental := map[string]bool{"history": true}
	for _, s := range bundle.Streams {
		hasIncremental := s.Incremental != nil
		if hasIncremental != wantIncremental[s.Name] {
			t.Errorf("stream %q incremental block present = %v, want %v", s.Name, hasIncremental, wantIncremental[s.Name])
		}
	}
}

// --- AuthSpec shape guard ---

// TestParityGmail_AuthSpecIsSoleCustomCandidate locks in SPEC.md §5.7's "no
// roster swap needed" decision: legacy has no alternate auth path, so the
// bundle declares exactly one auth candidate (mode custom, hook gmail), not
// a when-gated fallback list like github's bearer-or-app_jwt resolution.
func TestParityGmail_AuthSpecIsSoleCustomCandidate(t *testing.T) {
	bundle := loadGmailBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1 (no alternate auth path exists in legacy gmail)", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "custom" || spec.Hook != "gmail" {
		t.Fatalf("auth spec = %+v, want mode=custom hook=gmail", spec)
	}
}

// --- helper: a Hooks instance carrying the shared TLS-trusting client -----

// newHooksWithClient constructs the real gmail AuthHook
// (gmailhook.New().(*gmailhook.Hooks), the exact same type
// engine.RegisterHooks("gmail", ...) constructs) but overrides its
// exported Client field to trust the shared tokenServer's self-signed TLS
// certificate — mirrors how engine.HooksFor("gmail") behaves in production
// EXCEPT for the test certificate trust, which production never needs (a
// real Google token endpoint has a publicly trusted cert).
func newHooksWithClient(client *http.Client) engine.Hooks {
	h := gmailhook.New().(*gmailhook.Hooks)
	h.Client = client
	return h
}
