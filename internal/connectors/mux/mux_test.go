package mux_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mux"
)

// TestReadPaginatesAndAuthenticates is the red-first test: HTTP Basic auth
// (username/password), Mux page-number pagination over data[], and record
// mapping. The server returns a full page (limit records) on page 1 to trigger a
// second request, then a short page on page 2 to stop.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/video/v1/assets" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// limit defaults to 2 in the test config, so a 2-record page is "full".
			_, _ = w.Write([]byte(`{"data":[{"id":"asset_1","status":"ready","created_at":"1700000000"},{"id":"asset_2","status":"ready","created_at":"1700000100"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"asset_3","status":"preparing","created_at":"1700000200"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := mux.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "token_id_123", "page_size": "2"},
		Secrets: map[string]string{"password": "secret_key_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "assets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("token_id_123:secret_key_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access (required for credential-free conformance).
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := mux.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "live_streams", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode ensures Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := mux.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog covers the core streams.
func TestCatalogStreams(t *testing.T) {
	c := mux.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"assets": false, "live_streams": false, "uploads": false, "signing_keys": false}
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

// TestRegistration confirms self-registration via NewRegistry().Get.
func TestRegistration(t *testing.T) {
	_ = mux.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("mux"); !ok {
		t.Fatal("registry did not resolve mux (self-registration)")
	}
	caps := mux.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
