// Package paritytest_amazonsellerpartner drives the legacy
// internal/connectors/amazon-seller-partner connector and the engine-backed
// connector built from internal/connectors/defs/amazon-seller-partner
// against the SAME httptest SP-API data server AND the SAME httptest LWA
// token-exchange server, asserting RAW connectors.Record reflect.DeepEqual
// record parity (conventions.md's parity-suite pattern, mirrors
// paritytest/gmail's identical "OAuth2 refresh-token-grant AuthHook" shape —
// gmail's AuthHook does the OAuth2 refresh_token grant against Google, this
// one does the LWA refresh_token grant against Amazon's Login with Amazon
// token endpoint, both non-Authorization-Bearer/non-declarative token
// exchanges the engine cannot express without a Tier-2 AuthHook).
//
// This suite is the authoritative substitute this bundle's metadata.json
// conformance.skip_dynamic marker names: conformance's dynamic checks
// synthesize EVERY non-secret spec.json property (including lwa_token_url)
// as the literal string "synthetic-conformance-value", which is not a valid
// URL the AuthHook can POST to — every auth-resolving dynamic check would
// otherwise fail identically and uninformatively on that same synthetic
// non-URL, never actually exercising the read/pagination/incremental logic
// the check is meant to prove. This suite instead wires a REAL (test)
// lwa_token_url and SP-API data server so the full read path — including
// the LWA AuthHook itself — runs for real.
package paritytest_amazonsellerpartner

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
	amazonsellerpartner "polymetrics.ai/internal/connectors/amazon-seller-partner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/amazon-seller-partner" // registers the AuthHook via init()
)

// loadBundle resolves the "amazon-seller-partner" bundle from defs.FS via
// engine.Load, the same discovery path TestConformance and every other
// production caller uses.
func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "amazon-seller-partner")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, %q): %v", "amazon-seller-partner", err)
	}
	return b
}

// withBaseURL returns a shallow copy of b with HTTP.URL pointed at baseURL
// (engine.Bundle is a value type; this never mutates the loaded original,
// mirrors paritytest/gmail's withGmailBaseURL).
func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

// --- shared LWA token-exchange server (both sides authenticate against it) ---

func tokenServer(t *testing.T, accessToken string) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// runtimeConfig builds the connectors.RuntimeConfig shared by both
// connectors: base_url/lwa_token_url point at the shared httptest servers,
// and the three LWA secrets are synthetic placeholders (never a
// real-looking credential, per THREAT-MODEL §4).
func runtimeConfig(baseURL, tokenURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":               baseURL,
		"lwa_token_url":          tokenURL,
		"marketplace_id":         "ATVPDKIKX0DER",
		"replication_start_date": "2020-01-01T00:00:00Z",
		"page_size":              "100",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"lwa_app_id":        "lwa-app-id-fixture",
			"lwa_client_secret": "lwa-client-secret-fixture",
			"refresh_token":     "refresh-token-fixture",
		},
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
// legacy's native Go int64/string map-literal values and the engine's
// json.Number-preserving decode compare equal on numeric fields (mirrors
// parity_stripe_test.go/paritytest_gmail's normalizeGmailRecord).
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

// --- SP-API data server fixtures (NextToken/pagination.nextToken cursors) ---

func dataServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Every record below populates EVERY field the schema/legacy mapRecord
	// declares (matching a real SP-API wire response): legacy's mapRecord
	// functions build a Go map literal that sets every declared key
	// explicitly (nil when the raw API omits it), while the engine's
	// schema-mode projection only copies keys PRESENT in the raw record —
	// an abbreviated fixture that omits a field entirely would therefore
	// diverge (present-with-nil vs entirely-absent) for a reason that has
	// nothing to do with either connector's real behavior against the real
	// API, which always returns these fields. Full fixtures avoid that
	// artifact and are also more realistic.
	mux.HandleFunc("/orders/v0/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("NextToken") {
		case "":
			writeJSON(w, `{"payload":{"Orders":[
				{"AmazonOrderId":"111-1","SellerOrderId":"so-1","PurchaseDate":"2026-01-01T00:00:00Z","LastUpdateDate":"2026-01-01T00:00:00Z","OrderStatus":"Shipped","FulfillmentChannel":"AFN","SalesChannel":"Amazon.com","OrderType":"StandardOrder","NumberOfItemsShipped":1,"NumberOfItemsUnshipped":0,"MarketplaceId":"ATVPDKIKX0DER","OrderTotal":{"CurrencyCode":"USD","Amount":"10.00"},"IsBusinessOrder":false,"IsPrime":true},
				{"AmazonOrderId":"111-2","SellerOrderId":"so-2","PurchaseDate":"2026-01-02T00:00:00Z","LastUpdateDate":"2026-01-02T00:00:00Z","OrderStatus":"Pending","FulfillmentChannel":"MFN","SalesChannel":"Amazon.com","OrderType":"StandardOrder","NumberOfItemsShipped":0,"NumberOfItemsUnshipped":2,"MarketplaceId":"ATVPDKIKX0DER","OrderTotal":{"CurrencyCode":"USD","Amount":"20.00"},"IsBusinessOrder":false,"IsPrime":true}
			],"NextToken":"orders_page2"}}`)
		case "orders_page2":
			writeJSON(w, `{"payload":{"Orders":[
				{"AmazonOrderId":"111-3","SellerOrderId":"so-3","PurchaseDate":"2026-01-03T00:00:00Z","LastUpdateDate":"2026-01-03T00:00:00Z","OrderStatus":"Shipped","FulfillmentChannel":"AFN","SalesChannel":"Amazon.com","OrderType":"StandardOrder","NumberOfItemsShipped":1,"NumberOfItemsUnshipped":0,"MarketplaceId":"ATVPDKIKX0DER","OrderTotal":{"CurrencyCode":"USD","Amount":"30.00"},"IsBusinessOrder":true,"IsPrime":false}
			]}}`)
		default:
			t.Errorf("unexpected NextToken=%q for orders", r.URL.Query().Get("NextToken"))
			writeJSON(w, `{"payload":{"Orders":[]}}`)
		}
	})

	mux.HandleFunc("/fba/inventory/v1/summaries", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("nextToken") {
		case "":
			writeJSON(w, `{"pagination":{"nextToken":"inv_page2"},"payload":{"inventorySummaries":[
				{"sellerSku":"SKU-1","fnSku":"X001","asin":"B001","condition":"NewItem","productName":"Fixture Product 1","totalQuantity":5,"lastUpdatedTime":"2026-01-01T00:00:00Z","inventoryDetails":{"fulfillableQuantity":5}}
			]}}`)
		case "inv_page2":
			writeJSON(w, `{"payload":{"inventorySummaries":[
				{"sellerSku":"SKU-2","fnSku":"X002","asin":"B002","condition":"NewItem","productName":"Fixture Product 2","totalQuantity":9,"lastUpdatedTime":"2026-01-02T00:00:00Z","inventoryDetails":{"fulfillableQuantity":9}}
			]}}`)
		default:
			t.Errorf("unexpected nextToken=%q for inventory", r.URL.Query().Get("nextToken"))
			writeJSON(w, `{"payload":{"inventorySummaries":[]}}`)
		}
	})

	mux.HandleFunc("/finances/v0/financialEventGroups", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("NextToken") {
		case "":
			writeJSON(w, `{"payload":{"FinancialEventGroupList":[
				{"FinancialEventGroupId":"feg-1","ProcessingStatus":"Closed","FundTransferStatus":"Successful","OriginalTotal":{"CurrencyCode":"USD","CurrencyAmount":"100.00"},"ConvertedTotal":{"CurrencyCode":"USD","CurrencyAmount":"100.00"},"FundTransferDate":"2026-01-01T00:00:00Z","TraceId":"trace-1","AccountTail":"1234","BeginningBalance":{"CurrencyCode":"USD","CurrencyAmount":"0.00"},"FinancialEventGroupStart":"2026-01-01T00:00:00Z","FinancialEventGroupEnd":"2026-01-01T00:00:00Z"}
			],"NextToken":"feg_page2"}}`)
		case "feg_page2":
			writeJSON(w, `{"payload":{"FinancialEventGroupList":[
				{"FinancialEventGroupId":"feg-2","ProcessingStatus":"Closed","FundTransferStatus":"Successful","OriginalTotal":{"CurrencyCode":"USD","CurrencyAmount":"200.00"},"ConvertedTotal":{"CurrencyCode":"USD","CurrencyAmount":"200.00"},"FundTransferDate":"2026-01-02T00:00:00Z","TraceId":"trace-2","AccountTail":"1234","BeginningBalance":{"CurrencyCode":"USD","CurrencyAmount":"100.00"},"FinancialEventGroupStart":"2026-01-02T00:00:00Z","FinancialEventGroupEnd":"2026-01-02T00:00:00Z"}
			]}}`)
		default:
			t.Errorf("unexpected NextToken=%q for financial event groups", r.URL.Query().Get("NextToken"))
			writeJSON(w, `{"payload":{"FinancialEventGroupList":[]}}`)
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

func newLegacyConnector() connectors.Connector { return amazonsellerpartner.New() }

func newEngineConnector(b engine.Bundle, h engine.Hooks) connectors.Connector {
	return engine.New(b, h)
}

// --- per-stream record parity across all 3 streams ---

func TestParityAmazonSellerPartner_StreamRecords(t *testing.T) {
	bundle := loadBundle(t)

	streams := []string{"orders", "inventory_summaries", "financial_event_groups"}
	for _, stream := range streams {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			dataSrv := dataServer(t)
			tokenSrv, _ := tokenServer(t, "tok_"+stream)

			legacy := newLegacyConnector()
			legacyCfg := runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: stream, Config: legacyCfg})

			eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), engine.HooksFor("amazon-seller-partner"))
			engCfg := runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: stream, Config: engCfg})

			if len(legacyRecs) == 0 {
				t.Fatalf("legacy amazon-seller-partner emitted zero records for stream %q (test fixture bug)", stream)
			}
			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("record count = %d, want %d (legacy)\nengine:  %+v\nlegacy:  %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}

			gotNorm := normalizeRecords(t, engRecs)
			wantNorm := normalizeRecords(t, legacyRecs)
			for i := range wantNorm {
				if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
					t.Fatalf("stream %q record %d mismatch:\nengine:  %+v\nlegacy:  %+v", stream, i, gotNorm[i], wantNorm[i])
				}
			}
		})
	}
}

// --- pagination parity: orders 2-page NextToken/payload.NextToken cursor ---

func TestParityAmazonSellerPartner_OrdersTwoPagePagination(t *testing.T) {
	bundle := loadBundle(t)

	dataSrv := dataServer(t)
	tokenSrv, _ := tokenServer(t, "tok_orders")

	legacy := newLegacyConnector()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy orders records = %d, want 3 (2 pages)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), engine.HooksFor("amazon-seller-partner"))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine orders records = %d, want 3 (2 pages)", len(engRecs))
	}

	gotIDs := recordIDs(t, engRecs, "AmazonOrderId")
	wantIDs := recordIDs(t, legacyRecs, "AmazonOrderId")
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("orders record id sequence = %v, want %v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(gotIDs, []string{"111-1", "111-2", "111-3"}) {
		t.Fatalf("orders record id sequence = %v, want [111-1 111-2 111-3]", gotIDs)
	}
}

// TestParityAmazonSellerPartner_InventoryDistinctTokenPath asserts the FBA
// inventory stream uses the distinct pagination.nextToken body path (vs
// payload.NextToken for orders/financial_event_groups), matching legacy's
// distinct tokenPath per endpoint exactly.
func TestParityAmazonSellerPartner_InventoryDistinctTokenPath(t *testing.T) {
	bundle := loadBundle(t)

	dataSrv := dataServer(t)
	tokenSrv, _ := tokenServer(t, "tok_inventory")

	legacy := newLegacyConnector()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "inventory_summaries", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy inventory records = %d, want 2 (2 pages)", len(legacyRecs))
	}

	eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), engine.HooksFor("amazon-seller-partner"))
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "inventory_summaries", Config: runtimeConfig(dataSrv.URL, tokenSrv.URL, nil)})
	if len(engRecs) != 2 {
		t.Fatalf("engine inventory records = %d, want 2 (2 pages)", len(engRecs))
	}

	gotSKUs := recordIDs(t, engRecs, "sellerSku")
	wantSKUs := recordIDs(t, legacyRecs, "sellerSku")
	if !reflect.DeepEqual(gotSKUs, wantSKUs) {
		t.Fatalf("inventory sellerSku sequence = %v, want %v", gotSKUs, wantSKUs)
	}
}

func recordIDs(t *testing.T, recs []connectors.Record, field string) []string {
	t.Helper()
	out := make([]string, len(recs))
	for i, r := range recs {
		id, _ := r[field].(string)
		out[i] = id
	}
	return out
}

// --- auth parity: x-amz-access-token header after LWA refresh ---

// TestParityAmazonSellerPartner_AccessTokenHeaderAfterRefresh asserts BOTH
// connectors send the identical "x-amz-access-token: <access_token>" header
// (derived from the SAME LWA token-exchange response) on the request to the
// SP-API data endpoint — NEVER Authorization: Bearer, which is a different
// header SP-API does not use.
func TestParityAmazonSellerPartner_AccessTokenHeaderAfterRefresh(t *testing.T) {
	bundle := loadBundle(t)
	const accessToken = "Atza|tok_shared_fixture_value"

	var legacyAccessToken, engAccessToken, legacyAuthHeader, engAuthHeader string

	legacySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyAccessToken = r.Header.Get("x-amz-access-token")
		legacyAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"payload":{"Orders":[]}}`)
	}))
	t.Cleanup(legacySrv.Close)

	engSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engAccessToken = r.Header.Get("x-amz-access-token")
		engAuthHeader = r.Header.Get("Authorization")
		writeJSON(w, `{"payload":{"Orders":[]}}`)
	}))
	t.Cleanup(engSrv.Close)

	tokenSrv, hits := tokenServer(t, accessToken)

	legacy := newLegacyConnector()
	_ = readAllRecords(t, legacy, connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(legacySrv.URL, tokenSrv.URL, nil)})

	eng := newEngineConnector(withBaseURL(bundle, engSrv.URL), engine.HooksFor("amazon-seller-partner"))
	_ = readAllRecords(t, eng, connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(engSrv.URL, tokenSrv.URL, nil)})

	if legacyAccessToken != accessToken {
		t.Fatalf("legacy x-amz-access-token = %q, want %q (test fixture bug)", legacyAccessToken, accessToken)
	}
	if engAccessToken != accessToken {
		t.Fatalf("engine x-amz-access-token = %q, want %q (legacy, same shared token exchange)", engAccessToken, accessToken)
	}
	if legacyAuthHeader != "" {
		t.Fatalf("legacy Authorization header = %q, want empty (SP-API uses x-amz-access-token, not Bearer) (test fixture bug)", legacyAuthHeader)
	}
	if engAuthHeader != "" {
		t.Fatalf("engine Authorization header = %q, want empty (SP-API uses x-amz-access-token, not Bearer)", engAuthHeader)
	}
	if *hits != 2 {
		t.Fatalf("token endpoint hits = %d, want 2 (one refresh exchange per connector)", *hits)
	}
}

// TestParityAmazonSellerPartner_TokenEndpointFailureSurfacesAsAuthError
// asserts a token endpoint failure surfaces as an error on BOTH sides
// (never a silent unauthenticated request to the SP-API data endpoint).
func TestParityAmazonSellerPartner_TokenEndpointFailureSurfacesAsAuthError(t *testing.T) {
	bundle := loadBundle(t)

	var dataHits int32
	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&dataHits, 1)
		writeJSON(w, `{"payload":{"Orders":[]}}`)
	}))
	t.Cleanup(dataSrv.Close)

	failingTokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}))
	t.Cleanup(failingTokenSrv.Close)

	legacy := newLegacyConnector()
	legacyErr := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if legacyErr == nil {
		t.Fatal("legacy Read succeeded despite a failing token endpoint, want an error (test fixture bug)")
	}

	eng := newEngineConnector(withBaseURL(bundle, dataSrv.URL), engine.HooksFor("amazon-seller-partner"))
	engErr := eng.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: runtimeConfig(dataSrv.URL, failingTokenSrv.URL, nil)}, func(connectors.Record) error { return nil })
	if engErr == nil {
		t.Fatal("engine Read succeeded despite a failing token endpoint, want an error")
	}

	if dataHits != 0 {
		t.Fatalf("SP-API data endpoint received %d requests despite a failed token exchange, want 0 (no silent unauthenticated fallback)", dataHits)
	}
}

// --- write parity: both sides reject writes (read-only connector) ---

func TestParityAmazonSellerPartner_WriteUnsupportedOnBothSides(t *testing.T) {
	bundle := loadBundle(t)

	legacy := amazonsellerpartner.New()
	if _, err := legacy.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("legacy Write succeeded, want ErrUnsupportedOperation (test fixture bug)")
	}

	eng := engine.New(bundle, engine.HooksFor("amazon-seller-partner"))
	if _, err := eng.Write(context.Background(), connectors.WriteRequest{}, nil); err == nil {
		t.Fatal("engine Write succeeded, want an error (bundle declares capabilities.write: false, no writes.json)")
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (amazon-seller-partner is read-only)")
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (read-only, no writes.json)", bundle.Writes)
	}
}

// --- manifest-surface parity ---

func TestParityAmazonSellerPartner_ManifestSurface(t *testing.T) {
	bundle := loadBundle(t)

	legacyCatalog, err := amazonsellerpartner.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	eng := engine.New(bundle, engine.HooksFor("amazon-seller-partner"))
	engManifest := connectors.ManifestOf(eng)

	wantStreams := streamSurface(legacyCatalog.Streams)
	gotStreams := streamSurface(engManifest.Streams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("stream surface = %+v, want %+v (legacy Catalog())", gotStreams, wantStreams)
	}

	if len(engManifest.WriteActions) != 0 {
		t.Fatalf("engine write actions = %v, want none (amazon-seller-partner is read-only)", engManifest.WriteActions)
	}
}

type streamSurfaceEntry struct {
	Name       string
	PrimaryKey []string
}

func streamSurface(streams []connectors.Stream) []streamSurfaceEntry {
	out := make([]streamSurfaceEntry, len(streams))
	for i, s := range streams {
		out[i] = streamSurfaceEntry{Name: s.Name, PrimaryKey: append([]string{}, s.PrimaryKey...)}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// --- bundle load smoke guard ---

func TestParityAmazonSellerPartner_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadBundle(t)

	wantStreams := []string{"financial_event_groups", "inventory_summaries", "orders"}
	gotStreams := make([]string, 0, len(bundle.Streams))
	for _, s := range bundle.Streams {
		gotStreams = append(gotStreams, s.Name)
	}
	sort.Strings(gotStreams)
	if !reflect.DeepEqual(gotStreams, wantStreams) {
		t.Fatalf("bundle streams = %v, want %v", gotStreams, wantStreams)
	}

	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (read-only, no writes.json)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (amazon-seller-partner has no mutation API)")
	}
	for _, s := range bundle.Streams {
		if s.Incremental == nil {
			t.Errorf("stream %q declares no incremental block, want one (legacy publishes a CursorFields hint and applies it via LastUpdatedAfter/startDateTime/FinancialEventGroupStartedAfter)", s.Name)
		}
	}
}

// --- AuthSpec shape guard ---

// TestParityAmazonSellerPartner_AuthSpecIsSoleCustomCandidate locks in "no
// roster swap needed": legacy has no alternate auth path (LWA is the only
// credential shape SP-API supports), so the bundle declares exactly one auth
// candidate (mode custom, hook amazon-seller-partner), not a when-gated
// fallback list like github's bearer-or-app_jwt resolution.
func TestParityAmazonSellerPartner_AuthSpecIsSoleCustomCandidate(t *testing.T) {
	bundle := loadBundle(t)

	if len(bundle.HTTP.Auth) != 1 {
		t.Fatalf("len(bundle.HTTP.Auth) = %d, want 1 (no alternate auth path exists in legacy amazon-seller-partner)", len(bundle.HTTP.Auth))
	}
	spec := bundle.HTTP.Auth[0]
	if spec.Mode != "custom" || spec.Hook != "amazon-seller-partner" {
		t.Fatalf("auth spec = %+v, want mode=custom hook=amazon-seller-partner", spec)
	}
}
