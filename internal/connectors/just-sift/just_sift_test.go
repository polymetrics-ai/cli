package justsift_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	justsift "polymetrics.ai/internal/connectors/just-sift"
)

// TestReadPeoplePaginatesAndAuthenticates is the red-first test: Bearer auth,
// JustSift page-increment pagination over data[], and record mapping for the
// peoples stream. Red until internal/connectors/just-sift exists.
func TestReadPeoplePaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/search/people" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"p1","displayName":"Ada"},{"id":"p2","displayName":"Grace"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"p3","displayName":"Katherine"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := justsift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "peoples", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		if rec["connector"] != "just-sift" {
			t.Fatalf("record connector = %v, want just-sift", rec["connector"])
		}
	}
}

// TestReadFieldsCursorPaginates exercises the link-cursor pagination for the
// fields stream (links.next), confirming records aggregate across 2 pages.
func TestReadFieldsCursorPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fields/person" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("cursor") == "" {
			_, _ = w.Write([]byte(`{"data":[{"id":"f1","displayName":"Name"}],"links":{"next":"cursor=abc"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"f2","displayName":"Email"}],"links":{}}`))
	}))
	defer srv.Close()

	c := justsift.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "fields", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
}

// TestFixtureModeNoNetwork confirms credential-free fixture reads work for
// conformance without a server.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := justsift.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "peoples", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read produced no records")
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := justsift.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = justsift.New() // ensure init ran
	c := justsift.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("just-sift"); !ok {
		t.Fatal("registry did not resolve just-sift (self-registration)")
	}
}
