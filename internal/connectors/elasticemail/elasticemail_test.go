package elasticemail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/elasticemail"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// X-ElasticEmail-ApiKey header is sent, offset/limit pagination walks two pages
// of a top-level JSON array, and records are mapped. Red until the package
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-ElasticEmail-ApiKey")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		// page size is 2 in this test; first page returns a full page (2),
		// second page returns a short page (1) to terminate pagination.
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`[{"Email":"a@example.com","Status":"Active","DateAdded":"2026-01-01T00:00:00Z"},{"Email":"b@example.com","Status":"Active","DateAdded":"2026-01-02T00:00:00Z"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"Email":"c@example.com","Status":"Bounced","DateAdded":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := elasticemail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "ee_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "ee_test_key" {
		t.Fatalf("X-ElasticEmail-ApiKey = %q, want ee_test_key", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	emails := map[string]bool{}
	for _, rec := range got {
		if rec["Email"] == nil {
			t.Fatalf("record missing Email: %+v", rec)
		}
		emails[rec["Email"].(string)] = true
	}
	for _, want := range []string{"a@example.com", "b@example.com", "c@example.com"} {
		if !emails[want] {
			t.Fatalf("missing mapped email %q in %+v", want, got)
		}
	}
}

// TestPageSizeTerminatesOnExactBoundary makes sure a final full page followed by
// an empty page terminates cleanly (offset pagination stop condition).
func TestPageSizeTerminatesOnExactBoundary(t *testing.T) {
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lists" {
			http.NotFound(w, r)
			return
		}
		pages++
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`[{"ListName":"L1"},{"ListName":"L2"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := elasticemail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "ee_test_key"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if pages != 2 {
		t.Fatalf("pages fetched = %d, want 2 (full page then empty page)", pages)
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := elasticemail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"contacts", "campaigns", "lists", "segments", "templates"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogListsCoreStreams(t *testing.T) {
	c := elasticemail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "elasticemail" {
		t.Fatalf("catalog connector = %q, want elasticemail", cat.Connector)
	}
	byName := map[string]connectors.Stream{}
	for _, s := range cat.Streams {
		byName[s.Name] = s
	}
	if contacts, ok := byName["contacts"]; !ok {
		t.Fatal("catalog missing contacts stream")
	} else if len(contacts.PrimaryKey) != 1 || contacts.PrimaryKey[0] != "Email" {
		t.Fatalf("contacts primary key = %v, want [Email]", contacts.PrimaryKey)
	}
	for _, want := range []string{"campaigns", "lists", "segments", "templates"} {
		if _, ok := byName[want]; !ok {
			t.Fatalf("catalog missing %q stream", want)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := elasticemail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "ee_test_key"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}

func TestRegistryResolvesElasticEmail(t *testing.T) {
	_ = elasticemail.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("elasticemail")
	if !ok {
		t.Fatal("registry did not resolve elasticemail (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
