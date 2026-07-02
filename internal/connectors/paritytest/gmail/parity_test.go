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
			if stream == "labels" {
				// Documented, ACCEPTABLE-per-precedent deviation (matches
				// chargebee's identical finding): computed_fields' camelCase
				// -> snake_case rename (messagesTotal -> messages_total, etc.)
				// resolves through engine.Interpolate, which always returns a
				// Go string regardless of the raw JSON value's real type
				// (engine/interpolate.go's stringify) — so these 4 count
				// fields compare as their STRING form on both sides rather
				// than raw type-identical. See docs.md's Known limits and
				// this connector's ledger entry.
				for _, m := range []map[string]any{gotNorm[0], wantNorm[0]} {
					for _, k := range []string{"messages_total", "messages_unread", "threads_total", "threads_unread"} {
						m[k] = stringifyCountField(m[k])
					}
				}
			}
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// stringifyCountField renders a label count field (messages_total etc.) to
// its string form regardless of whether it arrived as a json.Number
// (legacy) or an already-stringified value (engine computed_fields) — see
// TestParityGmail_StreamRecords/labels's comment and docs.md Known limits.
func stringifyCountField(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		raw, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(raw)
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
	// See TestParityGmail_StreamRecords/labels: the 4 count fields are a
	// documented computed_fields stringification deviation, not raw-type
	// identical.
	for _, m := range append(gotNorm, wantNorm...) {
		for _, k := range []string{"messages_total", "messages_unread", "threads_total", "threads_unread"} {
			m[k] = stringifyCountField(m[k])
		}
	}
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("labels records mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm, wantNorm)
	}
}

// TestParityGmail_ComputedFieldsStringifyLabelCountFields locks in the
// documented parity deviation (docs.md Known limits; conventions.md §5
// candidate ledger entry): computed_fields' camelCase->snake_case rename is
// the ONLY Tier-1 way to project messagesTotal/messagesUnread/
// threadsTotal/threadsUnread under their schema-declared snake_case names,
// and engine.Interpolate (which every computed_fields template resolves
// through) always returns a Go string — so the engine emits these 4 fields
// as strings, while legacy emits them as connsdk's UseNumber-decoded
// json.Number (connsdk/extract.go's decodeJSON). This never changes the
// DATA a consumer would read (the string "10" carries the identical
// information as json.Number "10"), but it IS a real type-shape change, so
// it is asserted explicitly here rather than silently absorbed by a
// coercing equality helper (mirrors
// paritytest/chargebee's identical TestParityChargebee_
// ComputedFieldsStringifyNumericAndBooleanFields finding).
func TestParityGmail_ComputedFieldsStringifyLabelCountFields(t *testing.T) {
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
	got, ok := engRecs[0]["messages_total"].(string)
	if !ok {
		t.Fatalf("engine labels[0].messages_total = %#v (%T), want string (computed_fields always stringifies — engine/interpolate.go's resolveExpr/stringify)", engRecs[0]["messages_total"], engRecs[0]["messages_total"])
	}
	if got != "10" {
		t.Fatalf("engine labels[0].messages_total = %q, want %q (same numeric VALUE as legacy's json.Number, different Go type)", got, "10")
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

// --- write parity: both sides reject writes (read-only connector) ---

func TestParityGmail_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadGmailBundle(t)

	legacy := gmail.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor("gmail"))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (gmail bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (gmail is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (gmail is read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ---

func TestParityGmail_ManifestSurface(t *testing.T) {
	bundle := loadGmailBundle(t)

	legacyCatalog, err := gmail.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor("gmail"))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := gmailManifestStreamSurface(legacyCatalog.Streams)
	gotStreams := gmailManifestStreamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (gmail is read-only)", engManifest.WriteActions)
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

func TestParityGmail_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadGmailBundle(t)

	wantStreams := []string{"drafts", "labels", "messages", "threads"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (gmail is read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (gmail has no mutation API)")
	}
	for _, s := range bundle.Streams {
		if s.Incremental != nil {
			t.Errorf("stream %q declares an incremental block, want none (legacy publishes no cursor field, streams.go:31-34)", s.Name)
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
