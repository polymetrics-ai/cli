package customerly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/customerly"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, page
// increment pagination over data.users[] across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users/list" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		if r.URL.Query().Get("per_page") == "" {
			t.Errorf("expected per_page query param to be set")
		}
		switch page {
		case "0":
			// Full page of 2 records (page_size in test is 2 via config).
			_, _ = w.Write([]byte(`{"data":{"users":[` +
				`{"user_id":11,"email":"a@example.com","name":"A","last_update":"2026-01-01 00:00:00"},` +
				`{"user_id":12,"email":"b@example.com","name":"B","last_update":"2026-01-02 00:00:00"}` +
				`]}}`))
		case "1":
			// Short page (1 record) -> pagination stops after this page.
			_, _ = w.Write([]byte(`{"data":{"users":[` +
				`{"user_id":13,"email":"c@example.com","name":"C","last_update":"2026-01-03 00:00:00"}` +
				`]}}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":{"users":[]}}`))
		}
	}))
	defer srv.Close()

	c := customerly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer key_test_123" {
		t.Fatalf("Authorization = %q, want Bearer key_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPages) != 2 || sawPages[0] != "0" || sawPages[1] != "1" {
		t.Fatalf("paged requests = %v, want [0 1]", sawPages)
	}
	for _, rec := range got {
		if rec["user_id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing user_id/email: %+v", rec)
		}
	}
	// Verify mapping of the first record's id is preserved as a string-ish value.
	if _, err := strconv.Atoi(toString(got[0]["user_id"])); err != nil {
		t.Fatalf("user_id should be numeric, got %v", got[0]["user_id"])
	}
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		// json.Number stringifies cleanly.
		return toStringFallback(v)
	}
}

func toStringFallback(v any) string {
	type stringer interface{ String() string }
	if s, ok := v.(stringer); ok {
		return s.String()
	}
	return ""
}

func TestReadLeadsStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/leads/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"leads":[{"crmhero_user_id":"lead_1","email":"l@example.com","last_update":"2026-01-01 00:00:00"}]}}`))
	}))
	defer srv.Close()

	c := customerly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read leads: %v", err)
	}
	if len(got) != 1 || got[0]["crmhero_user_id"] != "lead_1" {
		t.Fatalf("leads = %+v, want one lead_1 record", got)
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := customerly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["user_id"] == nil {
			t.Fatalf("fixture record missing user_id: %+v", rec)
		}
	}
	// Check in fixture mode must succeed without creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := customerly.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "leads": false}
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

func TestRegistryResolvesCustomerly(t *testing.T) {
	_ = customerly.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("customerly")
	if !ok {
		t.Fatal("registry did not resolve customerly (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatal("customerly is read-only; Write capability should be false")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := customerly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
