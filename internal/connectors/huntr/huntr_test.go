package huntr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/huntr"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Huntr
// connector: Bearer auth, Huntr's next-cursor pagination over data[] (cursor
// carried in the `next` query param, stop when the response `next` is empty),
// and record mapping. Red until internal/connectors/huntr exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/members" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("next") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"mem_1","email":"a@example.com","createdAt":1700000000},{"id":"mem_2","email":"b@example.com","createdAt":1700000100}],"next":"cursor_2"}`))
		case "cursor_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"mem_3","email":"c@example.com","createdAt":1700000200}],"next":null}`))
		default:
			t.Errorf("unexpected next=%q", r.URL.Query().Get("next"))
			_, _ = w.Write([]byte(`{"data":[],"next":null}`))
		}
	}))
	defer srv.Close()

	c := huntr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "hk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer hk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer hk_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["email"] != "a@example.com" {
		t.Fatalf("record mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies the deterministic fixture path emits records
// without any network access so credential-free conformance passes.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := huntr.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCheckRequiresSecret verifies non-fixture Check rejects a missing api_key
// without making a network call.
func TestCheckRequiresSecret(t *testing.T) {
	c := huntr.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{}})
	if err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := huntr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"members": false, "candidates": false, "activities": false, "notes": false, "actions": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = huntr.New() // ensure init ran
	c := huntr.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("huntr"); !ok {
		t.Fatal("registry did not resolve huntr (self-registration)")
	}
}
