package ip2whois_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/ip2whois"
)

// sampleWhois returns a deterministic IP2WHOIS v2 lookup response for the given
// domain, mirroring the real API's single-object (non-list) shape.
func sampleWhois(domain string) string {
	return `{
		"domain": "` + domain + `",
		"domain_id": "` + domain + `-id",
		"status": "clientTransferProhibited",
		"create_date": "2000-01-01T00:00:00Z",
		"update_date": "2024-01-01T00:00:00Z",
		"expire_date": "2030-01-01T00:00:00Z",
		"domain_age": 9000,
		"whois_server": "whois.example-registrar.com",
		"registrar": {"iana_id": "123", "name": "Example Registrar", "url": "https://registrar.example"},
		"registrant": {"name": "Reg Ant", "organization": "Acme", "email": "reg@example.com", "country": "US"},
		"admin": {"name": "Ad Min", "email": "admin@example.com", "country": "US"},
		"tech": {"name": "Tech Person", "email": "tech@example.com", "country": "US"},
		"billing": {"name": "Bill Ing", "email": "billing@example.com", "country": "US"},
		"nameservers": ["ns1.example.com", "ns2.example.com"]
	}`
}

// TestReadIteratesDomainsAndAuthenticates is the red-first test: the api key is
// sent as the `key` query parameter, the connector iterates across two domains
// (two separate lookups, like two "pages"), and the whois record is mapped.
func TestReadIteratesDomainsAndAuthenticates(t *testing.T) {
	var (
		mu        sync.Mutex
		sawKeys   []string
		gotDomain []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		sawKeys = append(sawKeys, r.URL.Query().Get("key"))
		domain := r.URL.Query().Get("domain")
		gotDomain = append(gotDomain, domain)
		mu.Unlock()
		_, _ = w.Write([]byte(sampleWhois(domain)))
	}))
	defer srv.Close()

	c := ip2whois.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "domains": "alpha.com,beta.com"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "whois", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per domain)", len(got))
	}
	for _, k := range sawKeys {
		if k != "key_test_123" {
			t.Fatalf("key query = %q, want key_test_123", k)
		}
	}
	sort.Strings(gotDomain)
	if gotDomain[0] != "alpha.com" || gotDomain[1] != "beta.com" {
		t.Fatalf("queried domains = %v, want [alpha.com beta.com]", gotDomain)
	}
	for _, rec := range got {
		if rec["domain"] == nil || rec["registrar_name"] == nil {
			t.Fatalf("record missing domain/registrar_name: %+v", rec)
		}
		if rec["registrant_email"] == nil {
			t.Fatalf("record missing flattened registrant_email: %+v", rec)
		}
	}
}

// TestNameserversStreamFanout verifies the nameservers stream emits one record
// per nameserver across all domains.
func TestNameserversStreamFanout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sampleWhois(r.URL.Query().Get("domain"))))
	}))
	defer srv.Close()

	c := ip2whois.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "domains": "alpha.com,beta.com"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nameservers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// 2 domains x 2 nameservers each.
	if len(got) != 4 {
		t.Fatalf("nameserver records = %d, want 4", len(got))
	}
	for _, rec := range got {
		if rec["domain"] == nil || rec["nameserver"] == nil {
			t.Fatalf("nameserver record malformed: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode yields deterministic records
// without any network access so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := ip2whois.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "whois", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["domain"] == nil {
			t.Fatalf("fixture record missing domain: %+v", rec)
		}
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := ip2whois.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("ip2whois is read-only; Write should be false, got %+v", caps)
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(cat.Streams))
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("ip2whois"); !ok {
		t.Fatal("registry did not resolve ip2whois (self-registration)")
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := ip2whois.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example", "domains": "x.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "whois", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
