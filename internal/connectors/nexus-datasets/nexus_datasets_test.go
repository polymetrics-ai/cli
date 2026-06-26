package nexusdatasets_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	nexusdatasets "polymetrics.ai/internal/connectors/nexus-datasets"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the HMAC
// auth headers (access key id, user id, api key, and an Authorization signature
// header), offset/page pagination across two pages, record mapping from the
// raw_data envelope, and that the records flow through emit.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawAccessKey string
		sawUserID    string
		sawAPIKey    string
		sawAuth      string
		pageCalls    int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAccessKey = r.Header.Get("X-Infor-AccessKeyId")
		sawUserID = r.Header.Get("X-Infor-UserId")
		sawAPIKey = r.Header.Get("X-Infor-ApiKey")
		sawAuth = r.Header.Get("Authorization")
		if !strings.HasPrefix(r.URL.Path, "/datasets/orders") {
			http.NotFound(w, r)
			return
		}
		pageCalls++
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"records":[
				{"id":"rec_1","raw_data":{"order_id":"o1","amount":100},"updated_at":"2026-01-01T00:00:00Z"},
				{"id":"rec_2","raw_data":{"order_id":"o2","amount":200},"updated_at":"2026-01-02T00:00:00Z"}
			]}`))
		case "2":
			_, _ = w.Write([]byte(`{"records":[
				{"id":"rec_3","raw_data":{"order_id":"o3","amount":300},"updated_at":"2026-01-03T00:00:00Z"}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"records":[]}`))
		}
	}))
	defer srv.Close()

	c := nexusdatasets.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":      srv.URL,
			"dataset_name":  "orders",
			"access_key_id": "AKID123",
			"user_id":       "user-1",
			"page_size":     "2",
		},
		Secrets: map[string]string{
			"secret_key": "shhh-secret",
			"api_key":    "data-api-key",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "datasets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAccessKey != "AKID123" {
		t.Fatalf("X-Infor-AccessKeyId = %q, want AKID123", sawAccessKey)
	}
	if sawUserID != "user-1" {
		t.Fatalf("X-Infor-UserId = %q, want user-1", sawUserID)
	}
	if sawAPIKey != "data-api-key" {
		t.Fatalf("X-Infor-ApiKey = %q, want data-api-key", sawAPIKey)
	}
	if !strings.HasPrefix(sawAuth, "InforNexus ") {
		t.Fatalf("Authorization = %q, want HMAC signature prefixed InforNexus", sawAuth)
	}
	if pageCalls < 2 {
		t.Fatalf("pageCalls = %d, want >= 2 (pagination across pages)", pageCalls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		if rec["raw_data"] == nil {
			t.Fatalf("record missing raw_data envelope: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access, so conformance can run without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := nexusdatasets.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture", "dataset_name": "orders"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "datasets", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := nexusdatasets.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogAndMetadata checks the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := nexusdatasets.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("catalog has no streams")
	}
}

// TestRegistryResolution confirms the connector self-registers and resolves via
// NewRegistry().Get with the exact bare hyphenated name.
func TestRegistryResolution(t *testing.T) {
	_ = nexusdatasets.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("nexus-datasets"); !ok {
		t.Fatal("registry did not resolve nexus-datasets (self-registration)")
	}
}
