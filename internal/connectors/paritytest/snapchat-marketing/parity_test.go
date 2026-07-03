// Package paritytest_snapchatmarketing is the engine-vs-legacy parity suite
// for the snapchat-marketing fresh migration. Both the legacy hand-written
// snapchatmarketing.Connector (internal/connectors/snapchat-marketing,
// read-only reference) and the engine-backed connector
// (engine.New(bundle, engine.HooksFor("snapchat-marketing"))) are driven
// against the SAME httptest Snapchat-data server AND the SAME httptest TLS
// token-exchange server; RAW connectors.Record reflect.DeepEqual equality is
// the parity bar, matching internal/connectors/paritytest/strava's precedent
// for a hook-backed connector authenticating via an OAuth2 refresh-token
// grant. This suite is the authoritative correctness bar named in
// internal/connectors/defs/snapchat-marketing/docs.md's Known limits
// (metadata.json's conformance.skip_dynamic marker), since conformance's
// synthetic config can never round-trip a real refresh-token exchange.
package paritytest_snapchatmarketing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	snapchathook "polymetrics.ai/internal/connectors/hooks/snapchat-marketing" // registers the AuthHook via init(); also gives this test direct access to Hooks.Client for TLS trust
	snapchatmarketing "polymetrics.ai/internal/connectors/snapchat-marketing"
)

func loadSnapchatBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "snapchat-marketing")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "snapchat-marketing", err)
	}
	return b
}

// withSnapchatBaseURL returns a shallow copy of b with HTTP.URL pointed at
// baseURL (engine.Bundle is a value type; this never mutates the loaded
// original).
func withSnapchatBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared token-exchange server (both sides authenticate against it) ---

// tokenServer stands in for Snapchat's OAuth token endpoint. It MUST be a
// TLS server: legacy's own token_url default is https, and a fair parity
// comparison requires both sides talking to the identical endpoint. Returns
// the server, the *http.Client that trusts its self-signed cert (wired into
// BOTH connectors' Client field so neither side needs -insecure workarounds),
// and a hit counter.
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

func snapchatRuntimeConfig(baseURL, tokenURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":  baseURL,
		"token_url": tokenURL,
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"client_id":     "client-id-fixture",
			"client_secret": "client-secret-fixture",
			"refresh_token": "refresh-token-fixture",
		},
	}
}

func readAllSnapchatRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func normalizeSnapchatRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeSnapchatRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeSnapchatRecord(t, r)
	}
	return out
}

func newSnapchatLegacyConnector(client *http.Client) connectors.Connector {
	return snapchatmarketing.Connector{Client: client}
}

func newSnapchatEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// newHooksWithClient constructs the real snapchat-marketing AuthHook
// (snapchathook.New().(*snapchathook.Hooks), the exact same type
// engine.RegisterHooks("snapchat-marketing", ...) constructs) but overrides
// its exported Client field to trust the shared tokenServer's self-signed
// TLS certificate.
func newHooksWithClient(client *http.Client) engine.Hooks {
	h := snapchathook.New().(*snapchathook.Hooks)
	h.Client = client
	return h
}

// --- Snapchat data server: organizations (no fan-out) + campaigns (fan-out
// over ad_account_ids), both envelope-wrapped, both 2-page next_link ---

func snapchatDataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/organizations", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("cursor") {
		case "":
			next := "http://" + r.Host + "/organizations?cursor=PAGE2"
			writeJSON(w, `{"request_status":"SUCCESS","organizations":[`+
				`{"sub_request_status":"SUCCESS","organization":{"id":"org1","name":"Org One","type":"ENTERPRISE","country":"US","address_line_1":"1 Main St","locality":"LA","administrative_district_level_1":"CA","postal_code":"90001","created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-01T00:00:00.000Z"}}`+
				`],"paging":{"next_link":"`+next+`"}}`)
		case "PAGE2":
			writeJSON(w, `{"request_status":"SUCCESS","organizations":[`+
				`{"sub_request_status":"SUCCESS","organization":{"id":"org2","name":"Org Two","type":"ENTERPRISE","country":"US","address_line_1":"2 Main St","locality":"SF","administrative_district_level_1":"CA","postal_code":"94101","created_at":"2026-01-02T00:00:00.000Z","updated_at":"2026-01-02T00:00:00.000Z"}}`+
				`]}`)
		default:
			t.Errorf("unexpected organizations cursor=%q", r.URL.Query().Get("cursor"))
		}
	})

	mux.HandleFunc("/adaccounts/ACC1/campaigns", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("cursor") {
		case "":
			next := "http://" + r.Host + "/adaccounts/ACC1/campaigns?cursor=PAGE2"
			writeJSON(w, `{"request_status":"SUCCESS","campaigns":[`+
				`{"sub_request_status":"SUCCESS","campaign":{"id":"c1","name":"Camp One","status":"ACTIVE","ad_account_id":"ACC1","objective":"WEB_CONVERSION","start_time":"2026-01-01T00:00:00.000Z","end_time":"2026-02-01T00:00:00.000Z","daily_budget_micro":1000000,"lifetime_spend_cap_micro":5000000,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-01T00:00:00.000Z"}}`+
				`],"paging":{"next_link":"`+next+`"}}`)
		case "PAGE2":
			writeJSON(w, `{"request_status":"SUCCESS","campaigns":[`+
				`{"sub_request_status":"SUCCESS","campaign":{"id":"c2","name":"Camp Two","status":"PAUSED","ad_account_id":"ACC1","objective":"APP_INSTALLS","start_time":"2026-01-05T00:00:00.000Z","end_time":"2026-02-05T00:00:00.000Z","daily_budget_micro":2000000,"lifetime_spend_cap_micro":8000000,"created_at":"2026-01-05T00:00:00.000Z","updated_at":"2026-01-05T00:00:00.000Z"}}`+
				`]}`)
		default:
			t.Errorf("unexpected campaigns cursor=%q", r.URL.Query().Get("cursor"))
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// TestParitySnapchatMarketing_OrganizationsEnvelopeUnwrapAndPagination
// proves the engine's schema-projection + computed_fields envelope-unwrap
// shape produces RAW-record-identical output to legacy's unwrapEnvelopes +
// mapRecord for a non-fan-out, 2-page next_link stream.
func TestParitySnapchatMarketing_OrganizationsEnvelopeUnwrapAndPagination(t *testing.T) {
	bundle := loadSnapchatBundle(t)
	dataSrv := snapchatDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_organizations")

	legacy := newSnapchatLegacyConnector(tlsClient)
	legacyCfg := snapchatRuntimeConfig(dataSrv.URL+"/", tokenSrv.URL, nil)
	legacyRecs := readAllSnapchatRecords(t, legacy, connectors.ReadRequest{Stream: "organizations", Config: legacyCfg})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy organizations records = %d, want 2 (2 pages)", len(legacyRecs))
	}

	eng := newSnapchatEngineConnector(withSnapchatBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engCfg := snapchatRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
	engRecs := readAllSnapchatRecords(t, eng, connectors.ReadRequest{Stream: "organizations", Config: engCfg})
	if len(engRecs) != 2 {
		t.Fatalf("engine organizations records = %d, want 2 (2 pages)", len(engRecs))
	}

	gotNorm := normalizeSnapchatRecords(t, engRecs)
	wantNorm := normalizeSnapchatRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("organizations record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
		if _, wrapped := gotNorm[i]["organization"]; wrapped {
			t.Fatalf("engine record %d still wrapped in organization envelope: %+v", i, gotNorm[i])
		}
	}
}

// TestParitySnapchatMarketing_CampaignsFanOutOverAdAccountIDs proves the
// engine's fan_out (config_key: ad_account_ids, path_var) reproduces
// legacy's streamPaths "adaccount" scope: one full sub-sequence per
// configured ad account id, with envelope unwrap and 2-page pagination
// exercised inside that sub-sequence.
func TestParitySnapchatMarketing_CampaignsFanOutOverAdAccountIDs(t *testing.T) {
	bundle := loadSnapchatBundle(t)
	dataSrv := snapchatDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_campaigns")

	legacy := newSnapchatLegacyConnector(tlsClient)
	legacyCfg := snapchatRuntimeConfig(dataSrv.URL+"/", tokenSrv.URL, map[string]string{"ad_account_ids": "ACC1"})
	legacyRecs := readAllSnapchatRecords(t, legacy, connectors.ReadRequest{Stream: "campaigns", Config: legacyCfg})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy campaigns records = %d, want 2 (2 pages)", len(legacyRecs))
	}

	eng := newSnapchatEngineConnector(withSnapchatBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engCfg := snapchatRuntimeConfig(dataSrv.URL, tokenSrv.URL, map[string]string{"ad_account_ids": "ACC1"})
	engRecs := readAllSnapchatRecords(t, eng, connectors.ReadRequest{Stream: "campaigns", Config: engCfg})
	if len(engRecs) != 2 {
		t.Fatalf("engine campaigns records = %d, want 2 (2 pages)", len(engRecs))
	}

	gotNorm := normalizeSnapchatRecords(t, engRecs)
	wantNorm := normalizeSnapchatRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("campaigns record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
		// daily_budget_micro/lifetime_spend_cap_micro must stay numeric
		// (typed computed_fields extraction), not stringified.
		if _, ok := gotNorm[i]["daily_budget_micro"].(json.Number); !ok {
			t.Fatalf("campaigns record %d daily_budget_micro not numeric: %#v", i, gotNorm[i]["daily_budget_micro"])
		}
	}

	if legacyRecs[0]["ad_account_id"] != "ACC1" || engRecs[0]["ad_account_id"] != "ACC1" {
		t.Fatalf("campaigns record missing ad_account_id=ACC1: legacy=%v engine=%v", legacyRecs[0]["ad_account_id"], engRecs[0]["ad_account_id"])
	}
}

// TestParitySnapchatMarketing_MissingAdAccountIDs proves the DOCUMENTED
// parity deviation (docs.md's Known limits): legacy's streamPaths hard-errors
// when ad_account_ids is unset (snapchat_marketing.go:249), while the
// engine's fan_out.ids_from.config_key resolves an empty/absent config value
// to a zero-length id list (read.go's splitTrimmedCSV) and silently emits
// ZERO records (no sub-sequence ever runs) rather than erroring. This never
// changes emitted-record DATA for any legacy-ACCEPTED input (both sides
// require ad_account_ids to emit any campaigns/adsquads/ads records at all);
// it only changes the empty-config-value failure mode from a hard error to
// an empty read, which is ACCEPTABLE per conventions.md §5's meta-rule.
func TestParitySnapchatMarketing_MissingAdAccountIDs(t *testing.T) {
	bundle := loadSnapchatBundle(t)
	dataSrv := snapchatDataServer(t)
	tokenSrv, tlsClient, _ := tokenServer(t, "tok_missing")

	legacy := newSnapchatLegacyConnector(tlsClient)
	legacyCfg := snapchatRuntimeConfig(dataSrv.URL+"/", tokenSrv.URL, nil)
	if err := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: legacyCfg}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("legacy Read(campaigns) with no ad_account_ids: want error, got nil")
	}

	eng := newSnapchatEngineConnector(withSnapchatBaseURL(bundle, dataSrv.URL), newHooksWithClient(tlsClient))
	engCfg := snapchatRuntimeConfig(dataSrv.URL, tokenSrv.URL, nil)
	engRecs := readAllSnapchatRecords(t, eng, connectors.ReadRequest{Stream: "campaigns", Config: engCfg})
	if len(engRecs) != 0 {
		t.Fatalf("engine Read(campaigns) with no ad_account_ids: want 0 records (documented deviation), got %d", len(engRecs))
	}
}
