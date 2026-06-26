package mailjetsms_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	mailjetsms "polymetrics.ai/internal/connectors/mailjet-sms"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mailjet SMS
// connector: Bearer auth, offset-based pagination over the Data[] array, and
// record mapping. Mailjet SMS lists return {"Data":[...],"Count":N}; the next
// page is requested with Offset advancing by Limit until a short page returns.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/sms" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("Offset")
		limit := r.URL.Query().Get("Limit")
		if limit != "2" {
			t.Errorf("Limit = %q, want 2", limit)
		}
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{"Data":[{"ID":"1","To":"+15551112222","Status":{"Code":2,"Name":"sent"},"CreationTS":1700000000},{"ID":"2","To":"+15553334444","Status":{"Code":2,"Name":"sent"},"CreationTS":1700000100}],"Count":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"Data":[{"ID":"3","To":"+15555556666","Status":{"Code":2,"Name":"sent"},"CreationTS":1700000200}],"Count":1}`))
		default:
			t.Errorf("unexpected Offset=%q", offset)
			_, _ = w.Write([]byte(`{"Data":[],"Count":0}`))
		}
	}))
	defer srv.Close()

	c := mailjetsms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["ID"] == nil || rec["To"] == nil {
			t.Fatalf("record missing ID/To: %+v", rec)
		}
	}
	// Confirm nested Status.Code was flattened into status_code.
	if got[0]["status_code"] == nil {
		t.Fatalf("record missing flattened status_code: %+v", got[0])
	}
}

// TestReadAppliesDateFilters confirms FromTS/ToTS are sourced from config
// start_date/end_date and sent as query params on the first request.
func TestReadAppliesDateFilters(t *testing.T) {
	var sawFrom, sawTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawFrom = r.URL.Query().Get("FromTS")
		sawTo = r.URL.Query().Get("ToTS")
		_, _ = w.Write([]byte(`{"Data":[],"Count":0}`))
	}))
	defer srv.Close()

	c := mailjetsms.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "1666261656",
			"end_date":   "1666281656",
		},
		Secrets: map[string]string{"token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sms", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawFrom != "1666261656" {
		t.Fatalf("FromTS = %q, want 1666261656", sawFrom)
	}
	if sawTo != "1666281656" {
		t.Fatalf("ToTS = %q, want 1666281656", sawTo)
	}
}

// TestReadCountStream confirms the sms_count stream reads the single-object
// /sms/count response and emits it as one record.
func TestReadCountStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sms/count" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Count":42}`))
	}))
	defer srv.Close()

	c := mailjetsms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sms_count", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["Count"] == nil {
		t.Fatalf("count record missing Count: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoCreds confirms conformance can run credential-free.
func TestFixtureModeNeedsNoCreds(t *testing.T) {
	c := mailjetsms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["ID"] == nil {
		t.Fatalf("fixture record missing ID: %+v", got[0])
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := mailjetsms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sms", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should fail")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = mailjetsms.New() // ensure init ran
	c := mailjetsms.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mailjet-sms"); !ok {
		t.Fatal("registry did not resolve mailjet-sms (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the SMS streams.
func TestCatalogStreams(t *testing.T) {
	c := mailjetsms.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "mailjet-sms" {
		t.Fatalf("catalog connector = %q, want mailjet-sms", cat.Connector)
	}
	want := map[string]bool{"sms": false, "sms_count": false}
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
	// sms stream uses ID primary key.
	for _, s := range cat.Streams {
		if s.Name == "sms" {
			if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != "ID" {
				t.Fatalf("sms primary key = %v, want [ID]", s.PrimaryKey)
			}
		}
	}
}
