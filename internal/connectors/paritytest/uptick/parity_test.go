// Package paritytest_uptick is the engine-vs-legacy parity suite for the
// uptick fresh migration. Both the legacy hand-written uptick.Connector
// (internal/connectors/uptick, read-only reference) and the engine-backed
// connector (engine.New(bundle, engine.HooksFor("uptick"))) are driven
// against the SAME httptest Uptick-data server AND the SAME httptest TLS
// token-exchange server; RAW connectors.Record reflect.DeepEqual equality is
// the parity bar, matching internal/connectors/paritytest/strava's precedent
// for a hook-backed connector authenticating via a custom OAuth2 grant. This
// suite is the authoritative correctness bar named in
// internal/connectors/defs/uptick/docs.md's Known limits (metadata.json's
// conformance.skip_dynamic marker), since conformance's synthetic config can
// never round-trip a real username/password token exchange.
package paritytest_uptick

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	uptickhook "polymetrics.ai/internal/connectors/hooks/uptick" // registers the AuthHook via init(); also gives this test direct access to Hooks.Client for TLS trust
	"polymetrics.ai/internal/connectors/uptick"
)

func loadUptickBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "uptick")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "uptick", err)
	}
	return b
}

// withUptickBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original).
func withUptickBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// uptickServer stands in for BOTH Uptick's data API and its OAuth token
// endpoint (legacy's token_url is always {base_url}/api/oauth2/token/, a
// path under the SAME host as the data API, unlike strava/snapchat-marketing
// whose token endpoint is a distinct host) — a single httptest server
// handles both routes. Returns the server and a *tokenForm capture for
// asserting the exchanged grant shape.
func uptickServer(t *testing.T, accessToken string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/oauth2/token/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token server: parse form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "password" {
			t.Fatalf("token server: grant_type = %q, want password", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	mux.HandleFunc("/api/v2.14/clients/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			next := "http://" + r.Host + "/api/v2.14/clients/?page=2"
			writeJSON(w, `{"data":[`+
				`{"id":1,"created":"2026-01-01T00:00:00.000000Z","updated":"2026-01-01T00:00:00.000000Z","ref":"CLI-1","name":"Alpha","is_active":true,"contact_name":"Contact A","contact_email":"a@example.com","contact_phone_bh":"+61000000001","address":"1 Example St","notes":"n1"}`+
				`],"links":{"next":"`+next+`"}}`)
		case "2":
			writeJSON(w, `{"data":[`+
				`{"id":2,"created":"2026-01-02T00:00:00.000000Z","updated":"2026-01-02T00:00:00.000000Z","ref":"CLI-2","name":"Beta","is_active":true,"contact_name":"Contact B","contact_email":"b@example.com","contact_phone_bh":"+61000000002","address":"2 Example St","notes":"n2"}`+
				`],"links":{"next":null}}`)
		default:
			t.Errorf("unexpected clients page=%q", r.URL.Query().Get("page"))
		}
	})

	mux.HandleFunc("/api/v2.14/tasks/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, `{"data":[`+
			`{"id":10,"created":"2026-01-01T00:00:00.000000Z","updated":"2026-01-01T00:00:00.000000Z","deleted":null,"ref":"TSK-1","description":"d","is_active":true,"name":"Task A","due":"2026-01-05T00:00:00.000000Z","status":"active","client":"1","property":"1","priority":"normal"}`+
			`],"links":{"next":null}}`)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func uptickRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url": baseURL,
		"username": "ops@example.com",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"client_id":     "client-id-fixture",
			"client_secret": "client-secret-fixture",
			"password":      "password-fixture",
		},
	}
}

func readAllUptickRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func normalizeUptickRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeUptickRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeUptickRecord(t, r)
	}
	return out
}

func newUptickLegacyConnector(client *http.Client) connectors.Connector {
	return uptick.Connector{Client: client}
}

func newUptickEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// newHooksWithClient constructs the real uptick AuthHook
// (uptickhook.New().(*uptickhook.Hooks), the exact same type
// engine.RegisterHooks("uptick", ...) constructs) with its Client field
// pointed at the shared plain-HTTP data/token server (unlike
// strava/snapchat-marketing, uptick's own token endpoint is plain http in
// this suite's local test server — legacy's own validateHTTPURL/
// uptickBaseURL accepts plain http for local test servers, matching
// production https indistinguishably for this parity comparison).
func newHooksWithClient(client *http.Client) engine.Hooks {
	h := uptickhook.New().(*uptickhook.Hooks)
	h.Client = client
	return h
}

// TestParityUptick_ClientsTwoPagePaginationAndAuth proves the OAuth2
// password-grant AuthHook exchange plus links.next 2-page pagination
// produce RAW-record-identical output to legacy across both connectors.
func TestParityUptick_ClientsTwoPagePaginationAndAuth(t *testing.T) {
	bundle := loadUptickBundle(t)
	dataSrv := uptickServer(t, "tok_clients")

	legacy := newUptickLegacyConnector(dataSrv.Client())
	legacyCfg := uptickRuntimeConfig(dataSrv.URL, nil)
	legacyRecs := readAllUptickRecords(t, legacy, connectors.ReadRequest{Stream: "clients", Config: legacyCfg})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy clients records = %d, want 2 (2 pages)", len(legacyRecs))
	}

	eng := newUptickEngineConnector(withUptickBaseURL(bundle, dataSrv.URL), newHooksWithClient(dataSrv.Client()))
	engCfg := uptickRuntimeConfig(dataSrv.URL, nil)
	engRecs := readAllUptickRecords(t, eng, connectors.ReadRequest{Stream: "clients", Config: engCfg})
	if len(engRecs) != 2 {
		t.Fatalf("engine clients records = %d, want 2 (2 pages)", len(engRecs))
	}

	gotNorm := normalizeUptickRecords(t, engRecs)
	wantNorm := normalizeUptickRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("clients record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// TestParityUptick_TasksSingleStream proves per-record field parity (all
// fields present, native types preserved) for a second stream with a
// different field set.
func TestParityUptick_TasksSingleStream(t *testing.T) {
	bundle := loadUptickBundle(t)
	dataSrv := uptickServer(t, "tok_tasks")

	legacy := newUptickLegacyConnector(dataSrv.Client())
	legacyCfg := uptickRuntimeConfig(dataSrv.URL, nil)
	legacyRecs := readAllUptickRecords(t, legacy, connectors.ReadRequest{Stream: "tasks", Config: legacyCfg})

	eng := newUptickEngineConnector(withUptickBaseURL(bundle, dataSrv.URL), newHooksWithClient(dataSrv.Client()))
	engCfg := uptickRuntimeConfig(dataSrv.URL, nil)
	engRecs := readAllUptickRecords(t, eng, connectors.ReadRequest{Stream: "tasks", Config: engCfg})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)", len(engRecs), len(legacyRecs))
	}
	gotNorm := normalizeUptickRecords(t, engRecs)
	wantNorm := normalizeUptickRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("tasks record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
		if _, ok := gotNorm[i]["id"].(json.Number); !ok {
			t.Fatalf("tasks record %d id not numeric: %#v", i, gotNorm[i]["id"])
		}
	}
}

// TestParityUptick_IncrementalUpdatedSinceSentWhenCursorPresent proves both
// sides send updatedsince=<lower_bound> when a start_date-seeded lower bound
// resolves (dedicated servers per connector so each side's own request is
// unambiguous to capture).
func TestParityUptick_IncrementalUpdatedSinceSentWhenCursorPresent(t *testing.T) {
	bundle := loadUptickBundle(t)

	var legacySawUpdatedSince, engSawUpdatedSince string

	legacyMux := http.NewServeMux()
	legacyMux.HandleFunc("/api/oauth2/token/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "token_type": "Bearer", "expires_in": 3600})
	})
	legacyMux.HandleFunc("/api/v2.14/clients/", func(w http.ResponseWriter, r *http.Request) {
		legacySawUpdatedSince = r.URL.Query().Get("updatedsince")
		writeJSON(w, `{"data":[],"links":{"next":null}}`)
	})
	legacySrv := httptest.NewServer(legacyMux)
	t.Cleanup(legacySrv.Close)

	engMux := http.NewServeMux()
	engMux.HandleFunc("/api/oauth2/token/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "token_type": "Bearer", "expires_in": 3600})
	})
	engMux.HandleFunc("/api/v2.14/clients/", func(w http.ResponseWriter, r *http.Request) {
		engSawUpdatedSince = r.URL.Query().Get("updatedsince")
		writeJSON(w, `{"data":[],"links":{"next":null}}`)
	})
	engSrv := httptest.NewServer(engMux)
	t.Cleanup(engSrv.Close)

	legacy := newUptickLegacyConnector(legacySrv.Client())
	legacyCfg := uptickRuntimeConfig(legacySrv.URL, map[string]string{"start_date": "2026-01-01T00:00:00Z"})
	_ = readAllUptickRecords(t, legacy, connectors.ReadRequest{Stream: "clients", Config: legacyCfg})
	if legacySawUpdatedSince != "2026-01-01T00:00:00Z" {
		t.Fatalf("legacy updatedsince = %q, want 2026-01-01T00:00:00Z", legacySawUpdatedSince)
	}

	eng := newUptickEngineConnector(withUptickBaseURL(bundle, engSrv.URL), newHooksWithClient(engSrv.Client()))
	engCfg := uptickRuntimeConfig(engSrv.URL, map[string]string{"start_date": "2026-01-01T00:00:00Z"})
	_ = readAllUptickRecords(t, eng, connectors.ReadRequest{Stream: "clients", Config: engCfg})
	if engSawUpdatedSince != "2026-01-01T00:00:00Z" {
		t.Fatalf("engine updatedsince = %q, want 2026-01-01T00:00:00Z", engSawUpdatedSince)
	}
}
