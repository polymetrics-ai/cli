package onehms_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	onehms "polymetrics/internal/connectors/100ms"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth with the
// management token, 100ms data[]/last cursor pagination (next page requested via
// ?start=<last>), and record mapping over two pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/rooms" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"room_1","name":"alpha","enabled":true},{"id":"room_2","name":"beta","enabled":true}],"last":"room_2"}`))
		case "room_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"room_3","name":"gamma","enabled":false}],"last":""}`))
		default:
			t.Errorf("unexpected start=%q", r.URL.Query().Get("start"))
			_, _ = w.Write([]byte(`{"data":[],"last":""}`))
		}
	}))
	defer srv.Close()

	c := onehms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"management_token": "mgmt_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "rooms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer mgmt_test_123" {
		t.Fatalf("Authorization = %q, want Bearer mgmt_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["name"] != "alpha" {
		t.Fatalf("first room name = %v, want alpha", got[0]["name"])
	}
}

// TestSessionsPaginate exercises a second stream/endpoint to confirm the routing
// table and the data[]/last loop are stream-agnostic.
func TestSessionsPaginate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"sess_1","room_id":"room_1","active":true}],"last":"sess_1"}`))
		case "sess_1":
			_, _ = w.Write([]byte(`{"data":[{"id":"sess_2","room_id":"room_2","active":false}],"last":""}`))
		default:
			_, _ = w.Write([]byte(`{"data":[],"last":""}`))
		}
	}))
	defer srv.Close()

	c := onehms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"management_token": "mgmt_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read sessions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("session records = %d, want 2", len(got))
	}
	if got[0]["room_id"] != "room_1" {
		t.Fatalf("session room_id = %v, want room_1", got[0]["room_id"])
	}
}

// TestFixtureModeNoNetwork verifies credential-free conformance: fixture mode
// emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := onehms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "rooms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := onehms.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"rooms": false, "sessions": false, "recordings": false, "templates": false}
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
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = onehms.New() // ensure init ran
	c := onehms.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("100ms"); !ok {
		t.Fatal("registry did not resolve 100ms (self-registration)")
	}
}

// TestMissingSecretRejected confirms a live read without the management token errors.
func TestMissingSecretRejected(t *testing.T) {
	c := onehms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.100ms.live/v2"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "rooms", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without management_token should error")
	}
}
