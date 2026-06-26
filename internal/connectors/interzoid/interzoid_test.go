package interzoid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/interzoid"
)

// TestReadCompanyMatchAuthAndMapping is the red-first test: the Interzoid API
// authenticates with the api_key injected as the `license` query parameter, each
// stream hits a fixed lookup endpoint with the configured input params, and the
// single JSON object at the response root is mapped into one record.
func TestReadCompanyMatchAuthAndMapping(t *testing.T) {
	var sawPath, sawLicense, sawCompany, sawAlgorithm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawLicense = r.URL.Query().Get("license")
		sawCompany = r.URL.Query().Get("company")
		sawAlgorithm = r.URL.Query().Get("algorithm")
		_, _ = w.Write([]byte(`{"Code":"Success","SimKey":"ACME0001","Credits":"4711"}`))
	}))
	defer srv.Close()

	c := interzoid.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":                srv.URL,
			"company":                 "Acme Inc",
			"company_match_algorithm": "model-v4-wide",
		},
		Secrets: map[string]string{"api_key": "lic_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "company_name_matching", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/getcompanymatchadvanced" {
		t.Fatalf("path = %q, want /getcompanymatchadvanced", sawPath)
	}
	if sawLicense != "lic_test_123" {
		t.Fatalf("license = %q, want lic_test_123", sawLicense)
	}
	if sawCompany != "Acme Inc" {
		t.Fatalf("company = %q, want Acme Inc", sawCompany)
	}
	if sawAlgorithm != "model-v4-wide" {
		t.Fatalf("algorithm = %q, want model-v4-wide", sawAlgorithm)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["SimKey"] != "ACME0001" || rec["Code"] != "Success" {
		t.Fatalf("record mapping wrong: %+v", rec)
	}
	if rec["query_company"] != "Acme Inc" {
		t.Fatalf("record should echo query input, got query_company=%v", rec["query_company"])
	}
}

// TestReadAllStreamsHitRightEndpoints walks every published stream and confirms
// each routes to its distinct upstream endpoint with the configured inputs. This
// is the multi-request equivalent of the template's pagination assertion: there
// is no pagination in this lookup API, so coverage is across all four streams.
func TestReadAllStreamsHitRightEndpoints(t *testing.T) {
	paths := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths[r.URL.Path] = r.URL.RawQuery
		if r.URL.Path == "/getorgstandard" {
			_, _ = w.Write([]byte(`{"Code":"Success","Standard":"ACME","Credits":"10"}`))
			return
		}
		_, _ = w.Write([]byte(`{"Code":"Success","SimKey":"K","Credits":"10"}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":                srv.URL,
			"company":                 "Acme",
			"company_match_algorithm": "model-v4-wide",
			"fullname":                "Jane Doe",
			"address":                 "100 Main St",
			"address_match_algorithm": "model-v3-narrow",
			"org":                     "Acme Incorporated",
		},
		Secrets: map[string]string{"api_key": "lic"},
	}
	c := interzoid.New()

	want := map[string]string{
		"company_name_matching":     "/getcompanymatchadvanced",
		"individual_name_matching":  "/getfullnamematch",
		"street_address_matching":   "/getaddressmatchadvanced",
		"standardize_company_names": "/getorgstandard",
	}
	for stream, endpoint := range want {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s): %v", stream, err)
		}
		if n != 1 {
			t.Fatalf("Read(%s) emitted %d records, want 1", stream, n)
		}
		if _, ok := paths[endpoint]; !ok {
			t.Fatalf("stream %s did not hit endpoint %s; saw %v", stream, endpoint, paths)
		}
	}
}

// TestReadFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access so conformance can run without credentials.
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := interzoid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "company_name_matching", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if got[0]["Code"] == nil {
		t.Fatalf("fixture record missing Code: %+v", got[0])
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := interzoid.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "company": "x", "company_match_algorithm": "y"},
		Secrets: map[string]string{"api_key": "lic"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "company_name_matching", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := interzoid.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := interzoid.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{}})
	if err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := interzoid.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) != 4 {
		t.Fatalf("streams = %d, want 4", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
}

func TestReadOnlyCapabilities(t *testing.T) {
	c := interzoid.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read+Check+Catalog", caps)
	}
	if caps.Write {
		t.Fatal("interzoid is read-only; Write must be false")
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = interzoid.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("interzoid"); !ok {
		t.Fatal("registry did not resolve interzoid (self-registration)")
	}
}

// ensure url import is used in case of future query assertions.
var _ = url.Values{}
