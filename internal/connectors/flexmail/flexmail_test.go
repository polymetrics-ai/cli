package flexmail_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/flexmail"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Flexmail uses HTTP
// Basic auth (account_id:personal_access_token), records live under
// _embedded.item, and contacts paginate via offset/limit (OffsetIncrement).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// First page: a full page (limit) of records signals more to come.
			_, _ = w.Write([]byte(`{"_embedded":{"item":[{"id":1,"email":"a@example.com","name":"A"},{"id":2,"email":"b@example.com","name":"B"}]}}`))
		case "2":
			// Second page: short page ends pagination.
			_, _ = w.Write([]byte(`{"_embedded":{"item":[{"id":3,"email":"c@example.com","name":"C"}]}}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"_embedded":{"item":[]}}`))
		}
	}))
	defer srv.Close()

	c := flexmail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "acct_42", "page_size": "2"},
		Secrets: map[string]string{"personal_access_token": "pat_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("acct_42:pat_secret"))
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

// TestReadNonPaginatedStream covers a stream like custom_fields that has no
// paginator: a single request, records under _embedded.item.
func TestReadNonPaginatedStream(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/custom-fields" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"_embedded":{"item":[{"id":"cf_1","name":"Loyalty","type":"text"},{"id":"cf_2","name":"Tier","type":"text"}]}}`))
	}))
	defer srv.Close()

	c := flexmail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "acct_42"},
		Secrets: map[string]string{"personal_access_token": "pat_secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "custom_fields", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (no pagination)", calls)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "cf_1" || got[0]["name"] != "Loyalty" {
		t.Fatalf("unexpected record mapping: %+v", got[0])
	}
}

// TestFixtureMode confirms credential-free fixture reads for conformance.
func TestFixtureMode(t *testing.T) {
	c := flexmail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}

	for _, stream := range []string{"contacts", "custom_fields", "interests", "segments", "sources"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s, fixture) = %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := flexmail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "flexmail" {
		t.Fatalf("Catalog.Connector = %q, want flexmail", cat.Connector)
	}
	want := map[string]bool{"contacts": true, "custom_fields": true, "interests": true, "segments": true, "sources": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := flexmail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"personal_access_token": "pat_secret"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

func TestReadOnlyNoWrite(t *testing.T) {
	c := flexmail.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("flexmail is read-only, Write should be false: %+v", caps)
	}
	_, err := c.Write(context.Background(), connectors.WriteRequest{Action: "anything"}, nil)
	if err == nil {
		t.Fatal("Write should return ErrUnsupportedOperation for read-only connector")
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = flexmail.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("flexmail"); !ok {
		t.Fatal("registry did not resolve flexmail (self-registration)")
	}
}
