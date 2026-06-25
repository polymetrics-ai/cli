package carequalitycommission_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	carequalitycommission "polymetrics/internal/connectors/care-quality-commission"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the CQC connector:
// it asserts the Ocp-Apim-Subscription-Key header is sent, that page-increment
// pagination walks two pages of the locations stream (perPage threshold), and
// that records are mapped. Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("Ocp-Apim-Subscription-Key")
		if r.URL.Path != "/locations" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("perPage"); got != "2" {
			t.Errorf("perPage = %q, want 2", got)
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"page":1,"perPage":2,"totalPages":2,"total":3,"locations":[{"locationId":"1-101","locationName":"Alpha Care Home","postalCode":"AB1 2CD"},{"locationId":"1-102","locationName":"Beta Clinic","postalCode":"EF3 4GH"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"page":2,"perPage":2,"totalPages":2,"total":3,"locations":[{"locationId":"1-103","locationName":"Gamma Hospice","postalCode":"IJ5 6KL"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"page":99,"perPage":2,"totalPages":2,"total":3,"locations":[]}`))
		}
	}))
	defer srv.Close()

	c := carequalitycommission.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "primary-key-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "primary-key-123" {
		t.Fatalf("Ocp-Apim-Subscription-Key = %q, want primary-key-123", sawKey)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("requested %d pages (%v), want 2", len(pagesSeen), pagesSeen)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["locationId"] != "1-101" || got[0]["locationName"] != "Alpha Care Home" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["locationId"] != "1-103" {
		t.Fatalf("last record mapped wrong: %+v", got[2])
	}
}

// TestProvidersStreamMapsRecords confirms the providers stream maps the provider
// id/name fields and stops after a short page (single page, perPage not reached).
func TestProvidersStreamMapsRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/providers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"page":1,"perPage":1000,"totalPages":1,"total":1,"providers":[{"providerId":"1-201","providerName":"Acme Healthcare Ltd"}]}`))
	}))
	defer srv.Close()

	c := carequalitycommission.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "primary-key-123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "providers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["providerId"] != "1-201" || got[0]["providerName"] != "Acme Healthcare Ltd" {
		t.Fatalf("provider mapped wrong: %+v", got[0])
	}
}

// TestInspectionAreasUnpaginated confirms the inspection_areas stream reads the
// inspectionAreas array (no paginator) and maps its records.
func TestInspectionAreasUnpaginated(t *testing.T) {
	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/inspection-areas" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"inspectionAreas":[{"inspectionAreaId":"IA1","inspectionAreaName":"Safe","inspectionAreaType":"key_question","status":"active"},{"inspectionAreaId":"IA2","inspectionAreaName":"Effective","inspectionAreaType":"key_question","status":"active"}]}`))
	}))
	defer srv.Close()

	c := carequalitycommission.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "primary-key-123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "inspection_areas", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("requested %d times, want 1 (unpaginated)", requestCount)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["inspectionAreaId"] != "IA1" || got[0]["inspectionAreaName"] != "Safe" {
		t.Fatalf("inspection area mapped wrong: %+v", got[0])
	}
}

// TestFixtureMode confirms credential-free fixture reads emit deterministic
// records for the conformance harness.
func TestFixtureMode(t *testing.T) {
	c := carequalitycommission.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"locations", "providers", "inspection_areas"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
	}
	// Check short-circuits without network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata confirms the published catalog and read-only metadata.
func TestCatalogAndMetadata(t *testing.T) {
	c := carequalitycommission.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("CQC connector is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"locations": false, "providers": false, "inspection_areas": false}
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

// TestRegistryResolution asserts the connector self-registers and resolves via
// the shared registry under its bare system name.
func TestRegistryResolution(t *testing.T) {
	_ = carequalitycommission.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("care-quality-commission"); !ok {
		t.Fatal("registry did not resolve care-quality-commission (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF risk.
func TestBaseURLValidation(t *testing.T) {
	c := carequalitycommission.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url")
	}
}
