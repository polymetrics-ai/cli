package coassemble_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/coassemble"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Coassemble
// connector: the custom COASSEMBLE-V1-SHA256 Authorization header, page-increment
// pagination over the root array (page/length params), and record mapping. Red
// until internal/connectors/coassemble exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/headless/courses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			// A full page (length=2) signals there may be more.
			_, _ = w.Write([]byte(`[{"id":1,"title":"Onboarding","active":true},{"id":2,"title":"Safety","active":false}]`))
		case "2":
			// A short page (fewer than length) ends pagination.
			_, _ = w.Write([]byte(`[{"id":3,"title":"Compliance","active":true}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := coassemble.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"user_id": "uid_1", "user_token": "tok_1"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "COASSEMBLE-V1-SHA256 UserId=uid_1, UserToken=tok_1"
	if sawAuth != want {
		t.Fatalf("Authorization = %q, want %q", sawAuth, want)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms the credential-free fixture path emits deterministic
// records so the conformance harness can run without live API keys.
func TestReadFixtureMode(t *testing.T) {
	c := coassemble.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
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

	// Check must also short-circuit without network or credentials.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := coassemble.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"courses": false, "screen_types": false, "trackings": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistered confirms the connector self-registers and resolves via the
// process registry, and is read-only (no write capability).
func TestRegistered(t *testing.T) {
	_ = coassemble.New() // ensure init ran
	caps := coassemble.New().Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("coassemble"); !ok {
		t.Fatal("registry did not resolve coassemble (self-registration)")
	}
}
