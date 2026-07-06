package serpstatparity_test

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
	_ "polymetrics.ai/internal/connectors/hooks/serpstat" // registers the StreamHook via init()
	"polymetrics.ai/internal/connectors/serpstat"
)

// This is the migration parity suite for the serpstat bundle: serpstat is a
// Tier-2 StreamHook connector (docs.md "Overview") -- its JSON-RPC-over-HTTP
// POST reads with in-body pagination are the documented ENGINE_GAP trigger
// (quarantine.json's original finding). Both the legacy hand-written
// serpstat.Connector (internal/connectors/serpstat, read-only reference) and
// the engine-backed connector (engine.New(bundle, hooks.HooksFor("serpstat")))
// are driven against the SAME httptest.Server; RAW reflect.DeepEqual record
// equality is the parity bar.

func loadSerpstatBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, _ := engine.LoadAll(defs.FS)
	for _, b := range bundles {
		if b.Name == "serpstat" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "serpstat", names)
	return engine.Bundle{}
}

func withSerpstatBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func newSerpstatEngineConnector(b engine.Bundle) connectors.Connector {
	return engine.New(b, engine.HooksFor("serpstat"))
}

func serpstatRuntimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  cfg,
		Secrets: map[string]string{"api_key": "test-token-fixture"},
	}
}

func readAllSerpstatRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
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

func normalizeSerpstatRecord(t *testing.T, r connectors.Record) map[string]any {
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

func normalizeSerpstatRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeSerpstatRecord(t, r)
	}
	return out
}

type jsonRPCBody struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// --- domain_keywords: JSON-RPC body + query-param token auth parity ---

func serpstatKeywordsServer(t *testing.T, sawToken *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*sawToken = r.URL.Query().Get("token")
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body jsonRPCBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch int(asFloat(body.Params["page"])) {
		case 1:
			_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"k1","position":1,"url":"https://example.com/1"},{"keyword":"k2","position":2,"url":"https://example.com/2"}]}}`))
		case 2:
			_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"k3","position":3,"url":"https://example.com/3"}]}}`))
		default:
			_, _ = w.Write([]byte(`{"result":{"data":[]}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func asFloat(v any) float64 {
	f, _ := v.(float64)
	return f
}

func TestParitySerpstat_DomainKeywordsStreamRecordsAndAuth(t *testing.T) {
	bundle := loadSerpstatBundle(t)

	var legacyToken string
	legacySrv := serpstatKeywordsServer(t, &legacyToken)
	legacy := serpstat.New()
	legacyRecs := readAllSerpstatRecords(t, legacy, connectors.ReadRequest{
		Stream: "domain_keywords",
		Config: serpstatRuntimeConfig(legacySrv.URL, map[string]string{"page_size": "2", "pages_to_fetch": "0", "domain": "example.com"}),
	})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy records = %d, want 3 (test fixture bug)", len(legacyRecs))
	}
	if legacyToken != "test-token-fixture" {
		t.Fatalf("legacy token query param = %q, want test-token-fixture (test fixture bug)", legacyToken)
	}

	var engToken string
	engSrv := serpstatKeywordsServer(t, &engToken)
	eng := newSerpstatEngineConnector(withSerpstatBaseURL(bundle, engSrv.URL))
	engRecs := readAllSerpstatRecords(t, eng, connectors.ReadRequest{
		Stream: "domain_keywords",
		Config: serpstatRuntimeConfig(engSrv.URL, map[string]string{"page_size": "2", "pages_to_fetch": "0", "domain": "example.com"}),
	})

	if engToken != "test-token-fixture" {
		t.Fatalf("engine token query param = %q, want test-token-fixture (declarative api_key_query auth)", engToken)
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeSerpstatRecords(t, engRecs)
	wantNorm := normalizeSerpstatRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

// --- domain_competitors parity ---

func serpstatCompetitorsServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body jsonRPCBody
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		if body.Method != "SerpstatDomainProcedure.getCompetitors" {
			t.Errorf("method = %q, want getCompetitors procedure", body.Method)
		}
		_, _ = w.Write([]byte(`{"result":{"data":[{"domain":"competitor.example.com","visibility":5.5}]}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParitySerpstat_DomainCompetitorsStream(t *testing.T) {
	bundle := loadSerpstatBundle(t)

	legacySrv := serpstatCompetitorsServer(t)
	legacy := serpstat.New()
	legacyRecs := readAllSerpstatRecords(t, legacy, connectors.ReadRequest{
		Stream: "domain_competitors",
		Config: serpstatRuntimeConfig(legacySrv.URL, map[string]string{"pages_to_fetch": "1"}),
	})
	if len(legacyRecs) != 1 {
		t.Fatalf("legacy records = %d, want 1 (test fixture bug)", len(legacyRecs))
	}

	engSrv := serpstatCompetitorsServer(t)
	eng := newSerpstatEngineConnector(withSerpstatBaseURL(bundle, engSrv.URL))
	engRecs := readAllSerpstatRecords(t, eng, connectors.ReadRequest{
		Stream: "domain_competitors",
		Config: serpstatRuntimeConfig(engSrv.URL, map[string]string{"pages_to_fetch": "1"}),
	})
	if len(engRecs) != 1 {
		t.Fatalf("engine records = %d, want 1", len(engRecs))
	}

	gotNorm := normalizeSerpstatRecords(t, engRecs)
	wantNorm := normalizeSerpstatRecords(t, legacyRecs)
	if !reflect.DeepEqual(gotNorm[0], wantNorm[0]) {
		t.Fatalf("record mismatch:\nengine:  %+v\nlegacy:  %+v", gotNorm[0], wantNorm[0])
	}
}

// --- Check parity ---

func TestParitySerpstat_CheckSucceedsAgainstLiveServer(t *testing.T) {
	bundle := loadSerpstatBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"data":[]}}`))
	}))
	defer srv.Close()

	legacy := serpstat.New()
	if err := legacy.Check(context.Background(), serpstatRuntimeConfig(srv.URL, nil)); err != nil {
		t.Fatalf("legacy Check: %v", err)
	}

	eng := newSerpstatEngineConnector(withSerpstatBaseURL(bundle, srv.URL))
	if err := eng.Check(context.Background(), serpstatRuntimeConfig(srv.URL, nil)); err != nil {
		t.Fatalf("engine Check: %v", err)
	}
}

// --- bundle load smoke guard ---

func TestParitySerpstat_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadSerpstatBundle(t)

	wantStreams := map[string]bool{"domain_keywords": true, "domain_competitors": true, "domain_urls": true}
	if len(bundle.Streams) != len(wantStreams) {
		t.Fatalf("bundle streams = %v, want %d (%v)", bundle.Streams, len(wantStreams), wantStreams)
	}
	for _, s := range bundle.Streams {
		if !wantStreams[s.Name] {
			t.Fatalf("unexpected bundle stream %q", s.Name)
		}
	}
	if len(bundle.Writes) != 0 {
		t.Fatalf("bundle write actions = %v, want none (serpstat is read-only)", bundle.Writes)
	}
	if bundle.Metadata.Capabilities.Write {
		t.Fatal("bundle metadata.capabilities.write = true, want false (serpstat has no mutation API)")
	}
}
