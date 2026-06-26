package insightly_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/insightly"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Insightly
// connector: HTTP Basic auth (API token as username, blank password), skip/top
// offset pagination over a top-level JSON array, and record mapping. Insightly
// returns a short final page to signal the end.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	const token = "ac9a2292-f25a-4483-9d54-000000000000"
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(token+":"))

	var sawAuth string
	var sawSkips []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/Contacts" {
			http.NotFound(w, r)
			return
		}
		skip := r.URL.Query().Get("skip")
		sawSkips = append(sawSkips, skip)
		// top is the requested page size; the test server uses a page size of 2.
		switch skip {
		case "", "0":
			_, _ = w.Write([]byte(`[{"CONTACT_ID":1,"FIRST_NAME":"Ada","DATE_UPDATED_UTC":"2026-01-01 00:00:00"},{"CONTACT_ID":2,"FIRST_NAME":"Grace","DATE_UPDATED_UTC":"2026-01-02 00:00:00"}]`))
		case "2":
			// Short page (one record < page size) -> last page.
			_, _ = w.Write([]byte(`[{"CONTACT_ID":3,"FIRST_NAME":"Katherine","DATE_UPDATED_UTC":"2026-01-03 00:00:00"}]`))
		default:
			t.Errorf("unexpected skip=%q", skip)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := insightly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"token": token},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawSkips) != 2 {
		t.Fatalf("requested %d pages (skips=%v), want 2", len(sawSkips), sawSkips)
	}
	// Second page must have advanced the offset by the page size.
	if sawSkips[1] != strconv.Itoa(2) {
		t.Fatalf("second page skip = %q, want 2", sawSkips[1])
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing mapped id: %+v", rec)
		}
		if rec["contact_id"] == nil {
			t.Fatalf("record missing contact_id: %+v", rec)
		}
	}
}

// TestCheckAuthenticates verifies Check performs a bounded authenticated read.
func TestCheckAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := insightly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"token": "tok123"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if sawAuth == "" {
		t.Fatal("Check did not send an Authorization header")
	}
}

// TestFixtureModeNeedsNoNetwork verifies the credential-free fixture path.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := insightly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}

	for _, stream := range []string{"contacts", "organisations", "opportunities", "leads", "projects", "tasks"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s, fixture): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCatalogStreams verifies the published catalog covers the core streams.
func TestCatalogStreams(t *testing.T) {
	c := insightly.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{
		"contacts": false, "organisations": false, "opportunities": false,
		"leads": false, "projects": false, "tasks": false,
	}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := insightly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"token": "tok"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = insightly.New() // ensure init ran
	caps := insightly.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("insightly"); !ok {
		t.Fatal("registry did not resolve insightly (self-registration)")
	}
}
