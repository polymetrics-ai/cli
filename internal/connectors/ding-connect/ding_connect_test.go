package dingconnect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	dingconnect "polymetrics.ai/internal/connectors/ding-connect"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the DingConnect
// connector: api_key header auth, Skip-offset pagination over Items[], and record
// mapping. The DingConnect API authenticates with an `api_key` header (no prefix)
// and paginates by injecting a numeric Skip offset (page size 100) on every
// request, returning records under the top-level "Items" array.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawSkips []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("api_key")
		if r.URL.Path != "/api/V1/GetCountries" {
			http.NotFound(w, r)
			return
		}
		skip := r.URL.Query().Get("Skip")
		sawSkips = append(sawSkips, skip)
		switch skip {
		case "", "0":
			// A full page (== page size) forces a second request.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Items":[` + fullPage(100, 0) + `]}`))
		case "100":
			// A short page terminates pagination.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Items":[{"CountryIso":"GB","CountryName":"United Kingdom"}]}`))
		default:
			t.Errorf("unexpected Skip=%q", skip)
			_, _ = w.Write([]byte(`{"Items":[]}`))
		}
	}))
	defer srv.Close()

	c := dingconnect.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "countries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "key_test_123" {
		t.Fatalf("api_key header = %q, want key_test_123", sawAuth)
	}
	if len(got) != 101 {
		t.Fatalf("records = %d, want 101 (2 pages)", len(got))
	}
	if len(sawSkips) != 2 {
		t.Fatalf("requests = %d (skips=%v), want 2 pages", len(sawSkips), sawSkips)
	}
	// First request injects Skip on the first page; second advances to 100.
	if sawSkips[0] != "0" || sawSkips[1] != "100" {
		t.Fatalf("skips = %v, want [0 100]", sawSkips)
	}
	for _, rec := range got {
		if rec["CountryIso"] == nil {
			t.Fatalf("record missing CountryIso: %+v", rec)
		}
	}
}

// fullPage renders n country items as JSON objects (comma separated, no
// surrounding brackets) so the handler can splice them into an Items array.
func fullPage(n, base int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		idx := strconv.Itoa(base + i)
		out += `{"CountryIso":"C` + idx + `","CountryName":"Country ` + idx + `"}`
	}
	return out
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records with
// no network access so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := dingconnect.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "providers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["uuid"] == nil {
			t.Fatalf("fixture record missing uuid: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := dingconnect.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := dingconnect.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"countries": false, "currencies": false, "regions": false, "providers": false, "products": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration through the connectors registry.
func TestRegistryResolves(t *testing.T) {
	_ = dingconnect.New() // ensure the package init() ran
	caps := dingconnect.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("ding-connect"); !ok {
		t.Fatal("registry did not resolve ding-connect (self-registration)")
	}
}
