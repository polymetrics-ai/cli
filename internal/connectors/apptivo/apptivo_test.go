package apptivo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/apptivo"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Apptivo
// connector: apiKey/accessKey query auth, OffsetIncrement pagination
// (startIndex/numRecords) over data[], and record mapping. The customers
// endpoint returns 2 full pages then a short final page.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey, sawAccessKey, sawAction string
	var sawStartIndex []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/dao/v6/customers" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawAPIKey = q.Get("apiKey")
		sawAccessKey = q.Get("accessKey")
		sawAction = q.Get("a")
		sawStartIndex = append(sawStartIndex, q.Get("startIndex"))

		// numRecords drives page size; the test server uses 2 to keep the
		// payload small while still exercising multi-page offset paging.
		size, _ := strconv.Atoi(q.Get("numRecords"))
		if size == 0 {
			size = 2
		}
		start, _ := strconv.Atoi(q.Get("startIndex"))
		switch start {
		case 0:
			_, _ = w.Write([]byte(`{"data":[{"customerId":1,"customerName":"Acme","creationDate":"2026-01-01"},{"customerId":2,"customerName":"Globex","creationDate":"2026-01-02"}]}`))
		case 2:
			// Short page (1 < page size of 2) signals the last page.
			_, _ = w.Write([]byte(`{"data":[{"customerId":3,"customerName":"Initech","creationDate":"2026-01-03"}]}`))
		default:
			t.Errorf("unexpected startIndex=%d", start)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := apptivo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "ak_test", "access_key": "secret_test"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "ak_test" {
		t.Fatalf("apiKey = %q, want ak_test", sawAPIKey)
	}
	if sawAccessKey != "secret_test" {
		t.Fatalf("accessKey = %q, want secret_test", sawAccessKey)
	}
	if sawAction != "getAll" {
		t.Fatalf("a = %q, want getAll", sawAction)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	// First request must not inject startIndex (inject_on_first_request:false),
	// the second page must carry startIndex=2.
	if len(sawStartIndex) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(sawStartIndex))
	}
	if sawStartIndex[0] != "" {
		t.Fatalf("first page startIndex = %q, want empty (no inject on first request)", sawStartIndex[0])
	}
	if sawStartIndex[1] != "2" {
		t.Fatalf("second page startIndex = %q, want 2", sawStartIndex[1])
	}
	for _, rec := range got {
		if rec["customerId"] == nil || rec["customerName"] == nil {
			t.Fatalf("record missing customerId/customerName: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call, so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := apptivo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture lead missing id: %+v", got[0])
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := apptivo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := apptivo.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "apptivo" {
		t.Fatalf("catalog connector = %q, want apptivo", cat.Connector)
	}
	want := map[string]string{
		"customers":     "customerId",
		"contacts":      "contactId",
		"leads":         "id",
		"opportunities": "opportunityId",
	}
	got := map[string]string{}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) > 0 {
			got[s.Name] = s.PrimaryKey[0]
		}
	}
	for name, pk := range want {
		if got[name] != pk {
			t.Fatalf("stream %q primary key = %q, want %q", name, got[name], pk)
		}
	}
}

// TestReadIsReadOnly confirms Write is unsupported (Airbyte source is
// full_refresh read-only).
func TestReadOnlyCapabilities(t *testing.T) {
	c := apptivo.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities.Write = true, want false (read-only source)")
	}
}

// TestRegistryResolves confirms the connector self-registers and resolves via
// the shared registry.
func TestRegistryResolves(t *testing.T) {
	_ = apptivo.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("apptivo"); !ok {
		t.Fatal("registry did not resolve apptivo (self-registration)")
	}
}
