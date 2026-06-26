package defillama_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defillama"
)

// TestReadProtocolsPaginates is the red-first test for the DefiLlama connector.
// DefiLlama's public API has no authentication, so the connector sends no
// Authorization header. The /protocols endpoint returns a top-level JSON array;
// the connector pages through it client-side with limit/offset (the API itself
// returns the full list, so the connector slices it to keep payloads bounded and
// to prove the multi-page read loop). This test serves two pages of the sliced
// array and asserts every record is emitted and mapped.
func TestReadProtocolsPaginates(t *testing.T) {
	var sawAuth string
	var pathHits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		pathHits = append(pathHits, r.URL.Path)
		if r.URL.Path != "/protocols" {
			http.NotFound(w, r)
			return
		}
		// A six-element array; the connector reads it with page_size=3 so it must
		// fetch twice (offset 0 and offset 3) and stop on the short final page.
		all := []string{
			`{"id":"1","name":"Aave","slug":"aave","category":"Lending","chain":"Ethereum","tvl":1000.5,"mcap":2000}`,
			`{"id":"2","name":"Curve","slug":"curve","category":"Dexes","chain":"Ethereum","tvl":900.25,"mcap":1500}`,
			`{"id":"3","name":"Lido","slug":"lido","category":"Liquid Staking","chain":"Ethereum","tvl":800,"mcap":1200}`,
			`{"id":"4","name":"Uniswap","slug":"uniswap","category":"Dexes","chain":"Ethereum","tvl":700,"mcap":1100}`,
			`{"id":"5","name":"MakerDAO","slug":"makerdao","category":"CDP","chain":"Ethereum","tvl":600,"mcap":1000}`,
			`{"id":"6","name":"Compound","slug":"compound","category":"Lending","chain":"Ethereum","tvl":500,"mcap":900}`,
		}
		offset := 0
		if v := r.URL.Query().Get("offset"); v != "" {
			offset, _ = strconv.Atoi(v)
		}
		limit := len(all)
		if v := r.URL.Query().Get("limit"); v != "" {
			limit, _ = strconv.Atoi(v)
		}
		end := offset + limit
		if offset > len(all) {
			offset = len(all)
		}
		if end > len(all) {
			end = len(all)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("["))
		for i := offset; i < end; i++ {
			if i > offset {
				_, _ = w.Write([]byte(","))
			}
			_, _ = w.Write([]byte(all[i]))
		}
		_, _ = w.Write([]byte("]"))
	}))
	defer srv.Close()

	c := defillama.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "page_size": "3"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "protocols", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization = %q, want empty (DefiLlama is unauthenticated)", sawAuth)
	}
	if len(pathHits) < 2 {
		t.Fatalf("expected at least 2 page requests, got %d (%v)", len(pathHits), pathHits)
	}
	if len(got) != 6 {
		t.Fatalf("records = %d, want 6 (2 pages of 3)", len(got))
	}
	if got[0]["name"] != "Aave" || got[0]["category"] != "Lending" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[0]["slug"] != "aave" || got[5]["name"] != "Compound" {
		t.Fatalf("record mapping incomplete: first=%+v last=%+v", got[0], got[5])
	}
}

// TestReadStablecoinsNestedPath verifies the connector extracts records from a
// nested JSON path (stablecoins live under "peggedAssets", not the root) and
// that per-stream host base overrides are honored.
func TestReadStablecoinsNestedPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stablecoins" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"peggedAssets":[{"id":"1","name":"Tether","symbol":"USDT","pegType":"peggedUSD","pegMechanism":"fiat-backed","circulating":{"peggedUSD":1000},"price":1.0},{"id":"2","name":"USD Coin","symbol":"USDC","pegType":"peggedUSD","pegMechanism":"fiat-backed","circulating":{"peggedUSD":500},"price":1.0}]}`))
	}))
	defer srv.Close()

	c := defillama.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"stablecoins_base_url": srv.URL},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stablecoins", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["symbol"] != "USDT" || got[1]["symbol"] != "USDC" {
		t.Fatalf("stablecoin mapping wrong: %+v / %+v", got[0], got[1])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so conformance can run without configuration.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := defillama.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"protocols", "chains", "stablecoins", "dexs", "fees"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) < 2 {
			t.Fatalf("Read(%s) fixture: got %d records, want >= 2", stream, len(got))
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := defillama.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog lists the core streams.
func TestCatalogStreams(t *testing.T) {
	c := defillama.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"protocols": false, "chains": false, "stablecoins": false, "dexs": false, "fees": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = defillama.New() // ensure init ran
	caps := defillama.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (DefiLlama is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("defillama"); !ok {
		t.Fatal("registry did not resolve defillama (self-registration)")
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base URLs.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := defillama.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "file:///etc/passwd"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "protocols", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
