package zendesktalk_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	zendesktalk "polymetrics.ai/internal/connectors/zendesk-talk"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts API-token
// Basic auth, Zendesk next_page body pagination over the phone_numbers[] array,
// and record mapping. Red until internal/connectors/zendesk-talk exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v2/channels/voice/phone_numbers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"phone_numbers":[` +
				`{"id":1,"number":"+15550001","display_number":"+1 (555) 000-1","nickname":"a","created_at":"2026-01-01T00:00:00Z"},` +
				`{"id":2,"number":"+15550002","display_number":"+1 (555) 000-2","nickname":"b","created_at":"2026-01-02T00:00:00Z"}` +
				`],"next_page":"` + srvURL + `/api/v2/channels/voice/phone_numbers?page=2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"phone_numbers":[` +
				`{"id":3,"number":"+15550003","display_number":"+1 (555) 000-3","nickname":"c","created_at":"2026-01-03T00:00:00Z"}` +
				`],"next_page":null}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"phone_numbers":[],"next_page":null}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := zendesktalk.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{
			"credentials.email":     "agent@example.com",
			"credentials.api_token": "tok_123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "phone_numbers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantUser := "agent@example.com/token:tok_123"
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(wantUser))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["number"] == nil {
			t.Fatalf("record missing id/number: %+v", rec)
		}
	}
}

// TestReadBearerOAuth asserts that an OAuth access_token uses Bearer auth.
func TestReadBearerOAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"greetings":[{"id":10,"name":"hello"}],"next_page":null}`))
	}))
	defer srv.Close()

	c := zendesktalk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"credentials.access_token": "oauth_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "greetings", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer oauth_abc" {
		t.Fatalf("Authorization = %q, want Bearer oauth_abc", sawAuth)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance can run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := zendesktalk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "phone_numbers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := zendesktalk.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"phone_numbers": false, "greetings": false, "greeting_categories": false, "ivrs": false, "agents_activity": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %q missing primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base_url overrides.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := zendesktalk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "subdomain": "acme"},
		Secrets: map[string]string{"credentials.access_token": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "greetings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry().
func TestRegistryResolution(t *testing.T) {
	_ = zendesktalk.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("zendesk-talk"); !ok {
		t.Fatal("registry did not resolve zendesk-talk (self-registration)")
	}
	caps := zendesktalk.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
