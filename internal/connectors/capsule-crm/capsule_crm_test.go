package capsulecrm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics/internal/connectors"
	capsulecrm "polymetrics/internal/connectors/capsule-crm"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Capsule CRM
// connector: Bearer auth, page-number pagination over the "parties" wrapper key
// (perPage short-page stop), and record mapping. Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/parties" {
			http.NotFound(w, r)
			return
		}
		perPage := r.URL.Query().Get("perPage")
		if perPage != "2" {
			t.Errorf("perPage = %q, want 2", perPage)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			// A full page (== perPage) so the connector requests page 2.
			_, _ = w.Write([]byte(`{"parties":[{"id":1,"type":"person","firstName":"Ada","updatedAt":"2026-01-01T00:00:00Z"},{"id":2,"type":"organisation","organisationName":"Acme","updatedAt":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			// A short page (< perPage) so the connector stops.
			_, _ = w.Write([]byte(`{"parties":[{"id":3,"type":"person","firstName":"Grace","updatedAt":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"parties":[]}`))
		}
	}))
	defer srv.Close()

	c := capsulecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"bearer_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "parties", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["first_name"] != "Ada" {
		t.Fatalf("record[0].first_name = %v, want Ada", got[0]["first_name"])
	}
	if got[1]["organisation_name"] != "Acme" {
		t.Fatalf("record[1].organisation_name = %v, want Acme", got[1]["organisation_name"])
	}
}

// TestReadOpportunitiesMapping exercises a second stream end to end so the
// per-stream routing table and mapper are covered.
func TestReadOpportunitiesMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/opportunities" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"opportunities":[{"id":10,"name":"Big Deal","value":{"amount":5000,"currency":"USD"},"milestone":{"id":3,"name":"Bid"},"party":{"id":1},"updatedAt":"2026-02-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := capsulecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"bearer_token": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "opportunities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["name"] != "Big Deal" {
		t.Fatalf("name = %v, want Big Deal", got[0]["name"])
	}
	if got[0]["milestone_id"] == nil {
		t.Fatalf("milestone_id missing: %+v", got[0])
	}
	if got[0]["party_id"] == nil {
		t.Fatalf("party_id missing: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := capsulecrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"parties", "opportunities", "kases"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			n++
			if rec["id"] == nil {
				t.Fatalf("%s fixture record missing id: %+v", stream, rec)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("Read(%s) fixture emitted 0 records", stream)
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := capsulecrm.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := capsulecrm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestMetadataReadOnly(t *testing.T) {
	c := capsulecrm.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = capsulecrm.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("capsule-crm"); !ok {
		t.Fatal("registry did not resolve capsule-crm (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := capsulecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"bearer_token": "tok_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "parties", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should fail SSRF validation")
	}
}

// guard against accidental int parsing regressions in the page param.
var _ = strconv.Itoa
