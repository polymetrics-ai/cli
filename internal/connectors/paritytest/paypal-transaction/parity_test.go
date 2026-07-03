// Package paritytest_paypaltransaction is the engine-vs-legacy parity suite
// for the paypal-transaction Tier-2 (AuthHook + StreamHook) migration. Both
// the legacy hand-written paypaltransaction.Connector
// (internal/connectors/paypal-transaction, read-only reference) and the
// engine-backed connector (engine.New(bundle, engine.HooksFor(
// "paypal-transaction"))) are driven against the SAME httptest server (both
// the OAuth2 token endpoint and the data endpoints); RAW connectors.Record
// reflect.DeepEqual-shaped equality (via JSON-normalized comparison, exactly
// like paritytest/twilio) is the parity bar. This is also the authoritative
// live substitute defs/paypal-transaction/metadata.json's bundle-level
// skip_dynamic marker names: conformance's synthetic config can never
// populate a real base_url for the AuthHook's token exchange to reach, so
// this suite is what actually proves both the Basic-credential-in/
// Bearer-out token exchange and the disputes HATEOAS links[] pagination
// against a real server.
package paritytest_paypaltransaction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/paypal-transaction" // registers the AuthHook+StreamHook via init()
	paypaltransaction "polymetrics.ai/internal/connectors/paypal-transaction"
)

const bundleName = "paypal-transaction"

func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, bundleName)
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", bundleName, err)
	}
	return b
}

func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func runtimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":   baseURL,
		"start_date": "2026-01-01T00:00:00Z",
		"end_date":   "2026-01-03T00:00:00Z",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
}

func newLegacyConnector() connectors.Connector { return paypaltransaction.New() }

func newEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
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

func normalizeRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRecord(t, r)
	}
	return out
}

func sortByKey(t *testing.T, recs []map[string]any, key string) {
	t.Helper()
	sort.Slice(recs, func(i, j int) bool {
		return jsonString(recs[i][key]) < jsonString(recs[j][key])
	})
}

func jsonString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// transactionJSONs builds inclusive-range raw {transaction_info:{...}} JSON
// object literals (used to construct a genuinely full-size page so both
// legacy's total_pages check and the engine's short-page-stop paginator
// agree a second page exists).
func transactionJSONs(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		idx := strconv.Itoa(i)
		out = append(out, `{"transaction_info":{"transaction_id":"T`+idx+`","transaction_amount":{"currency_code":"USD","value":"10.00"},"fee_amount":{"currency_code":"USD","value":"-0.50"},"transaction_status":"S","transaction_event_code":"T0006","transaction_initiation_date":"2026-01-01T00:00:00Z","transaction_updated_date":"2026-01-01T00:05:00Z","paypal_account_id":"ACC1"}}`)
	}
	return out
}

// productJSONs builds inclusive-range raw product JSON object literals.
func productJSONs(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		idx := strconv.Itoa(i)
		out = append(out, `{"id":"PROD-`+idx+`","name":"Product `+idx+`","description":"d`+idx+`","type":"SERVICE","category":"SOFTWARE","create_time":"2026-01-01T00:00:00Z"}`)
	}
	return out
}

// dataServer builds a server serving BOTH the OAuth2 token endpoint (HTTP
// Basic client_id:client_secret in, Bearer access_token out) and the 4 data
// endpoints (transactions 2-page page_number with a full 100-record first
// page, balances single-page, products 2-page page_number with a full
// 20-record first page (page_size=20), disputes 2-page HATEOAS links[]) —
// mirrors legacy's own TestReadPaginatesAndAuthenticates fixture shapes,
// extended to cover every stream and sized so pagination genuinely
// continues on both the legacy and engine sides.
func dataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); !ok {
			t.Errorf("token request missing HTTP Basic auth")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"A123","token_type":"Bearer","expires_in":3600}`))
	})

	mux.HandleFunc("/v1/reporting/transactions", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer A123" {
			t.Errorf("transactions request Authorization = %q, want Bearer A123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			// A FULL 100-record page (the configured page_size) so both
			// legacy's total_pages check AND the engine's short-page-stop
			// paginator agree there is a second page — proving genuine
			// pagination parity, not just the documented total_pages-vs-
			// short-page-stop deviation's edge case.
			_, _ = w.Write([]byte(`{"transaction_details":[` + strings.Join(transactionJSONs(1, 100), ",") + `],"page":1,"total_pages":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"transaction_details":[` + strings.Join(transactionJSONs(101, 101), ",") + `],"page":2,"total_pages":2}`))
		default:
			t.Errorf("unexpected page=%q for transactions", page)
		}
	})

	mux.HandleFunc("/v1/reporting/balances", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balances":[{"currency":"USD","primary":true,"total_balance":{"currency_code":"USD","value":"100.00"},"available_balance":{"currency_code":"USD","value":"90.00"},"withheld_balance":{"currency_code":"USD","value":"10.00"}},{"currency":"EUR","primary":false,"total_balance":{"currency_code":"EUR","value":"50.00"},"available_balance":{"currency_code":"EUR","value":"45.00"},"withheld_balance":{"currency_code":"EUR","value":"5.00"}}],"account_id":"ACC1"}`))
	})

	mux.HandleFunc("/v1/catalogs/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			// A FULL 20-record page (products' page_size override) for the
			// same reason as transactions above.
			_, _ = w.Write([]byte(`{"products":[` + strings.Join(productJSONs(1, 20), ",") + `],"total_pages":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"products":[` + strings.Join(productJSONs(21, 21), ",") + `],"total_pages":2}`))
		default:
			t.Errorf("unexpected page=%q for products", page)
		}
	})

	mux.HandleFunc("/v1/customer/disputes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.RawQuery, "next_page_token") {
			_, _ = w.Write([]byte(`{"items":[{"dispute_id":"PP-D-2","reason":"UNAUTHORIZED","status":"RESOLVED","dispute_state":"RESOLVED","dispute_amount":{"currency_code":"USD","value":"5.00"},"create_time":"2026-01-02T00:00:00Z","update_time":"2026-01-02T00:05:00Z"}],"links":[]}`))
			return
		}
		nextHref := "http://" + r.Host + "/v1/customer/disputes?page_size=50&next_page_token=abc"
		_, _ = w.Write([]byte(`{"items":[{"dispute_id":"PP-D-1","reason":"MERCHANDISE_OR_SERVICE_NOT_RECEIVED","status":"RESOLVED","dispute_state":"RESOLVED","dispute_amount":{"currency_code":"USD","value":"12.31"},"create_time":"2026-01-01T00:00:00Z","update_time":"2026-01-01T00:05:00Z"}],"links":[{"href":"` + nextHref + `","rel":"next","method":"GET"}]}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// --- per-stream record parity across all 4 streams ---

func TestParityPaypalTransaction_StreamRecords(t *testing.T) {
	bundle := loadBundle(t)

	cases := []struct {
		stream  string
		sortKey string
	}{
		{"transactions", "transaction_id"},
		{"balances", "currency"},
		{"products", "id"},
		{"disputes", "dispute_id"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.stream, func(t *testing.T) {
			srv := dataServer(t)

			legacy := newLegacyConnector()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: tc.stream, Config: runtimeConfig(srv.URL, nil)})

			eng := newEngineConnector(withBaseURL(bundle, srv.URL), engine.HooksFor(bundleName))
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: tc.stream, Config: runtimeConfig(srv.URL, nil)})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy paypal-transaction emitted zero records for stream %q (test fixture bug)", tc.stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			normLegacy := normalizeRecords(t, legacyRecs)
			normEng := normalizeRecords(t, engRecs)
			sortByKey(t, normLegacy, tc.sortKey)
			sortByKey(t, normEng, tc.sortKey)

			if !reflect.DeepEqual(normEng, normLegacy) {
				t.Fatalf("records differ:\nengine: %+v\nlegacy: %+v", normEng, normLegacy)
			}
		})
	}
}

// --- pagination parity: transactions 2-page page_number ---

func TestParityPaypalTransaction_TransactionsTwoPagePagination(t *testing.T) {
	bundle := loadBundle(t)
	srv := dataServer(t)

	legacy := newLegacyConnector()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "transactions", Config: runtimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 101 {
		t.Fatalf("legacy transactions records = %d, want 101 (2 pages, 100 + 1)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, srv.URL), engine.HooksFor(bundleName))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "transactions", Config: runtimeConfig(srv.URL, nil)})
	if len(engRecs) != 101 {
		t.Fatalf("engine transactions records = %d, want 101 (2 pages, 100 + 1)", len(engRecs))
	}
}

// --- pagination parity: disputes 2-page HATEOAS links[] ---

func TestParityPaypalTransaction_DisputesTwoPageHATEOASPagination(t *testing.T) {
	bundle := loadBundle(t)
	srv := dataServer(t)

	legacy := newLegacyConnector()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "disputes", Config: runtimeConfig(srv.URL, nil)})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy disputes records = %d, want 2 (2 HATEOAS pages)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, srv.URL), engine.HooksFor(bundleName))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "disputes", Config: runtimeConfig(srv.URL, nil)})
	if len(engRecs) != 2 {
		t.Fatalf("engine disputes records = %d, want 2 (2 HATEOAS pages)", len(engRecs))
	}
	if engRecs[0]["dispute_id"] != "PP-D-1" || engRecs[1]["dispute_id"] != "PP-D-2" {
		t.Fatalf("engine disputes order/mapping wrong: %+v", engRecs)
	}
}

// --- auth parity: HTTP Basic client-credentials token exchange ---

// tokenAuthServer builds a fresh server whose /v1/oauth2/token handler
// records whether the request carried HTTP Basic auth, for one connector
// under test at a time (avoids any cross-request marker-string plumbing).
func tokenAuthServer(t *testing.T, sawBasic *bool) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); ok {
			*sawBasic = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"A123","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/reporting/balances", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balances":[{"currency":"USD","total_balance":{"currency_code":"USD","value":"1.00"}}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestParityPaypalTransaction_TokenExchangeUsesHTTPBasic(t *testing.T) {
	bundle := loadBundle(t)

	var legacySawBasic bool
	legacySrv := tokenAuthServer(t, &legacySawBasic)
	legacy := newLegacyConnector()
	if err := legacy.Check(context.Background(), runtimeConfig(legacySrv.URL, nil)); err != nil {
		t.Fatalf("legacy Check: %v", err)
	}
	if !legacySawBasic {
		t.Fatal("legacy token exchange did not use HTTP Basic")
	}

	var engineSawBasic bool
	engineSrv := tokenAuthServer(t, &engineSawBasic)
	eng := newEngineConnector(withBaseURL(bundle, engineSrv.URL), engine.HooksFor(bundleName))
	if err := eng.Check(context.Background(), runtimeConfig(engineSrv.URL, nil)); err != nil {
		t.Fatalf("engine Check: %v", err)
	}
	if !engineSawBasic {
		t.Fatal("engine token exchange did not use HTTP Basic")
	}
}
