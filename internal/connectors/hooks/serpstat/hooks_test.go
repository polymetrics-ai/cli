// Package serpstat's hooks_test.go proves hooks.go's StreamHook ports
// legacy internal/connectors/serpstat/serpstat.go's JSON-RPC body
// construction, in-body page-number pagination, and result.data record
// extraction byte-for-byte.
package serpstat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

type jsonRPCRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("serpstat")
	if h == nil {
		t.Fatal("engine.HooksFor(\"serpstat\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "serpstat" {
		t.Fatalf("ConnectorName() = %q, want serpstat", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("registered hooks implement AuthHook, want none (auth is fully declarative via api_key_query)")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

func TestReadStream_EmptyStreamNameDefaultsToDomainKeywords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body jsonRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Method != "SerpstatDomainProcedure.getKeywords" {
			t.Errorf("method = %q, want getKeywords procedure", body.Method)
		}
		_, _ = w.Write([]byte(`{"result":{"data":[]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false for empty stream name, want true (defaults to domain_keywords)")
	}
}

// --- JSON-RPC body shape + pagination ---

func TestReadStream_DomainKeywordsJSONRPCBodyAndPagination(t *testing.T) {
	var requests []jsonRPCRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		requests = append(requests, body)
		w.Header().Set("Content-Type", "application/json")
		switch body.Params["page"].(float64) {
		case 1:
			_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"k1","position":1,"url":"https://example.com/1"},{"keyword":"k2","position":2,"url":"https://example.com/2"}]}}`))
		case 2:
			_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"k3","position":3,"url":"https://example.com/3"}]}}`))
		default:
			t.Errorf("unexpected page requested: %v", body.Params["page"])
		}
	}))
	defer srv.Close()

	req := connectors.ReadRequest{
		Stream: "domain_keywords",
		Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "2", "pages_to_fetch": "0", "domain": "example.com", "region_id": "g_us"}},
	}

	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_keywords"}, req, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (short-page stop after page 2)", len(got))
	}
	if len(requests) != 2 {
		t.Fatalf("requests = %d, want 2 (stopped on short page)", len(requests))
	}

	first := requests[0]
	if first.ID != 1 {
		t.Fatalf("first request id = %d, want 1 (page number)", first.ID)
	}
	if first.Method != "SerpstatDomainProcedure.getKeywords" {
		t.Fatalf("method = %q, want getKeywords procedure", first.Method)
	}
	if first.Params["domain"] != "example.com" {
		t.Fatalf("params.domain = %v, want example.com", first.Params["domain"])
	}
	if first.Params["se"] != "g_us" {
		t.Fatalf("params.se = %v, want g_us", first.Params["se"])
	}
	if first.Params["size"].(float64) != 2 {
		t.Fatalf("params.size = %v, want 2", first.Params["size"])
	}

	second := requests[1]
	if second.ID != 2 {
		t.Fatalf("second request id = %d, want 2", second.ID)
	}
}

func TestReadStream_DomainCompetitorsUsesGetCompetitorsProcedure(t *testing.T) {
	var sawMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body jsonRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawMethod = body.Method
		_, _ = w.Write([]byte(`{"result":{"data":[{"domain":"competitor.example.com","visibility":5.5}]}}`))
	}))
	defer srv.Close()

	req := connectors.ReadRequest{Stream: "domain_competitors", Config: connectors.RuntimeConfig{Config: map[string]string{"pages_to_fetch": "1"}}}
	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_competitors"}, req, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawMethod != "SerpstatDomainProcedure.getCompetitors" {
		t.Fatalf("method = %q, want getCompetitors procedure", sawMethod)
	}
	if len(got) != 1 || got[0]["domain"] != "competitor.example.com" {
		t.Fatalf("records = %+v, want one competitor record", got)
	}
}

func TestReadStream_DomainUrlsUsesGetDomainUrlsProcedure(t *testing.T) {
	var sawMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body jsonRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawMethod = body.Method
		_, _ = w.Write([]byte(`{"result":{"data":[{"url":"https://example.com/","keywords":42}]}}`))
	}))
	defer srv.Close()

	req := connectors.ReadRequest{Stream: "domain_urls", Config: connectors.RuntimeConfig{Config: map[string]string{"pages_to_fetch": "1"}}}
	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_urls"}, req, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawMethod != "SerpstatDomainProcedure.getDomainUrls" {
		t.Fatalf("method = %q, want getDomainUrls procedure", sawMethod)
	}
	if len(got) != 1 || got[0]["url"] != "https://example.com/" {
		t.Fatalf("records = %+v, want one url record", got)
	}
}

// --- pages_to_fetch bounds ---

func TestReadStream_PagesToFetchCapsRequestCount(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		// Every page returns a FULL page (never short), so only
		// pages_to_fetch bounds the loop -- proving the cap is a real
		// hard stop, not just the pre-existing short-page stop.
		_, _ = w.Write([]byte(`{"result":{"data":[{"keyword":"k","position":1,"url":"https://example.com"}]}}`))
	}))
	defer srv.Close()

	req := connectors.ReadRequest{Stream: "domain_keywords", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "1", "pages_to_fetch": "2"}}}
	h := Hooks{}
	if _, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_keywords"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if hits != 2 {
		t.Fatalf("requests = %d, want 2 (pages_to_fetch cap)", hits)
	}
}

func TestReadStream_InvalidPagesToFetchIsError(t *testing.T) {
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"pages_to_fetch": "-1"}}}
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_keywords"}, req, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for a negative pages_to_fetch")
	}
}

func TestReadStream_InvalidPageSizeIsError(t *testing.T) {
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "0"}}}
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_keywords"}, req, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for an out-of-range page_size")
	}
}

// --- defaults ---

func TestReadStream_DefaultsDomainAndRegionWhenUnset(t *testing.T) {
	var sawParams map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body jsonRPCRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawParams = body.Params
		_, _ = w.Write([]byte(`{"result":{"data":[]}}`))
	}))
	defer srv.Close()

	req := connectors.ReadRequest{Stream: "domain_keywords", Config: connectors.RuntimeConfig{}}
	h := Hooks{}
	if _, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "domain_keywords"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if sawParams["domain"] != "serpstat.com" {
		t.Fatalf("params.domain = %v, want default serpstat.com", sawParams["domain"])
	}
	if sawParams["se"] != "g_us" {
		t.Fatalf("params.se = %v, want default g_us", sawParams["se"])
	}
}

// --- ctx cancellation ---

func TestReadStream_HonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":{"data":[]}}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := Hooks{}
	_, err := h.ReadStream(ctx, engine.StreamSpec{Name: "domain_keywords"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream(cancelled ctx) error = nil, want a cancellation error")
	}
}
