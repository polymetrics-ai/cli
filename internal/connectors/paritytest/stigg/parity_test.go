package stiggparity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/stigg" // registers the StreamHook via init()
	"polymetrics.ai/internal/connectors/stigg"
)

// This is the migration parity suite for the stigg bundle: stigg is a
// Tier-2 StreamHook connector (docs.md "Overview") -- its GraphQL-over-HTTP
// POST reads are the documented ENGINE_GAP trigger (quarantine.json's
// original finding). Both the legacy hand-written stigg.Connector
// (internal/connectors/stigg, read-only reference) and the engine-backed
// connector (engine.New(bundle, hooks.HooksFor("stigg"))) are driven
// against the SAME httptest.Server; RAW reflect.DeepEqual record equality
// is the parity bar.

func loadStiggBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, _ := engine.LoadAll(defs.FS)
	for _, b := range bundles {
		if b.Name == "stigg" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "stigg", names)
	return engine.Bundle{}
}

func withStiggBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func newStiggEngineConnector(b engine.Bundle) connectors.Connector {
	return engine.New(b, engine.HooksFor("stigg"))
}

func stiggRuntimeConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": baseURL},
		Secrets: map[string]string{"api_key": "test-token-fixture"},
	}
}

func readAllStiggRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func normalizeStiggRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeStiggRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeStiggRecord(t, r)
	}
	return out
}

type graphQLRequest struct {
	Query string `json:"query"`
}

// --- products: GraphQL body + bearer auth parity ---

func stiggProductsServer(t *testing.T, sawAuth *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*sawAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		var body graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"products":[{"id":"prod1","refId":"starter","displayName":"Starter","status":"ACTIVE"},{"id":"prod2","refId":"pro","displayName":"Pro","status":"ACTIVE"}]}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityStigg_ProductsStreamRecordsAndAuth(t *testing.T) {
	bundle := loadStiggBundle(t)

	var legacyAuth string
	legacySrv := stiggProductsServer(t, &legacyAuth)
	legacy := stigg.New()
	legacyRecs := readAllStiggRecords(t, legacy, connectors.ReadRequest{
		Stream: "products",
		Config: stiggRuntimeConfig(legacySrv.URL),
	})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy records = %d, want 2 (test fixture bug)", len(legacyRecs))
	}
	if legacyAuth != "Bearer test-token-fixture" {
		t.Fatalf("legacy Authorization = %q, want bearer token (test fixture bug)", legacyAuth)
	}

	var engAuth string
	engSrv := stiggProductsServer(t, &engAuth)
	eng := newStiggEngineConnector(withStiggBaseURL(bundle, engSrv.URL))
	engRecs := readAllStiggRecords(t, eng, connectors.ReadRequest{
		Stream: "products",
		Config: stiggRuntimeConfig(engSrv.URL),
	})

	if engAuth != "Bearer test-token-fixture" {
		t.Fatalf("engine Authorization = %q, want bearer token (declarative bearer auth)", engAuth)
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeStiggRecords(t, engRecs)
	wantNorm := normalizeStiggRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- plans/customers/subscriptions parity ---

func stiggSimpleServer(t *testing.T, root, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
		_ = root
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityStigg_PlansCustomersSubscriptionsStreams(t *testing.T) {
	bundle := loadStiggBundle(t)

	cases := []struct {
		stream string
		root   string
		body   string
	}{
		{"plans", "plans", `{"data":{"plans":[{"id":"plan1","refId":"starter-monthly","displayName":"Starter Monthly","status":"ACTIVE"}]}}`},
		{"customers", "customers", `{"data":{"customers":[{"id":"cust1","refId":"fixture-customer","displayName":"Fixture Customer","status":"ACTIVE"}]}}`},
		{"subscriptions", "subscriptions", `{"data":{"subscriptions":[{"id":"sub1","refId":"fixture-sub","customerId":"cust1","status":"ACTIVE"}]}}`},
	}

	for _, tc := range cases {
		t.Run(tc.stream, func(t *testing.T) {
			legacySrv := stiggSimpleServer(t, tc.root, tc.body)
			legacy := stigg.New()
			legacyRecs := readAllStiggRecords(t, legacy, connectors.ReadRequest{
				Stream: tc.stream,
				Config: stiggRuntimeConfig(legacySrv.URL),
			})
			if len(legacyRecs) != 1 {
				t.Fatalf("legacy %s records = %d, want 1 (test fixture bug)", tc.stream, len(legacyRecs))
			}

			engSrv := stiggSimpleServer(t, tc.root, tc.body)
			eng := newStiggEngineConnector(withStiggBaseURL(bundle, engSrv.URL))
			engRecs := readAllStiggRecords(t, eng, connectors.ReadRequest{
				Stream: tc.stream,
				Config: stiggRuntimeConfig(engSrv.URL),
			})
			if len(engRecs) != 1 {
				t.Fatalf("engine %s records = %d, want 1", tc.stream, len(engRecs))
			}

			gotNorm := normalizeStiggRecords(t, engRecs)
			wantNorm := normalizeStiggRecords(t, legacyRecs)
			if !reflect.DeepEqual(gotNorm[0], wantNorm[0]) {
				t.Fatalf("%s record mismatch:\nengine:  %+v\nlegacy:  %+v", tc.stream, gotNorm[0], wantNorm[0])
			}
		})
	}
}

// --- Check parity ---

func TestParityStigg_CheckSucceedsAgainstLiveServer(t *testing.T) {
	bundle := loadStiggBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"products":[]}}`))
	}))
	defer srv.Close()

	legacy := stigg.New()
	if err := legacy.Check(context.Background(), stiggRuntimeConfig(srv.URL)); err != nil {
		t.Fatalf("legacy Check: %v", err)
	}

	eng := newStiggEngineConnector(withStiggBaseURL(bundle, srv.URL))
	if err := eng.Check(context.Background(), stiggRuntimeConfig(srv.URL)); err != nil {
		t.Fatalf("engine Check: %v", err)
	}
}

// --- bundle load smoke guard ---

func TestParityStigg_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadStiggBundle(t)

	if len(bundle.Streams) != 4 {
		t.Fatalf("bundle streams = %v, want 4 (products, plans, customers, subscriptions)", bundle.Streams)
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (stigg is read-only)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (stigg has no mutation API)")
	}
}
