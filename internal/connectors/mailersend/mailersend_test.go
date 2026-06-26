package mailersend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailersend"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the MailerSend
// connector: Bearer auth, page/limit pagination over data[], short-page stop,
// and record mapping. Red until internal/connectors/mailersend exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/domains" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		switch page {
		case "", "1":
			// limit=2 -> full page, signals there may be more.
			_, _ = w.Write([]byte(`{"data":[{"id":"dom_1","name":"a.com","is_verified":true},{"id":"dom_2","name":"b.com","is_verified":false}],"meta":{"current_page":1}}`))
		case "2":
			// short page -> last page.
			_, _ = w.Write([]byte(`{"data":[{"id":"dom_3","name":"c.com","is_verified":true}],"meta":{"current_page":2}}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := mailersend.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_token": "mlsn_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "domains", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer mlsn_test_123" {
		t.Fatalf("Authorization = %q, want Bearer mlsn_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages (%v)", len(got), sawPages)
	}
	if len(sawPages) != 2 {
		t.Fatalf("requested pages = %v, want exactly 2", sawPages)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadActivityRequiresDomainAndDateRange verifies the activity stream targets
// activity/{domain_id} and forwards the required date_from/date_to window.
func TestReadActivityRequiresDomainAndDateRange(t *testing.T) {
	var sawPath, sawFrom, sawTo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawFrom = r.URL.Query().Get("date_from")
		sawTo = r.URL.Query().Get("date_to")
		_, _ = w.Write([]byte(`{"data":[{"id":"act_1","type":"delivered","created_at":"2026-01-01T00:00:00.000000Z"}],"meta":{}}`))
	}))
	defer srv.Close()

	c := mailersend.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"domain_id": "dom_xyz",
			"date_from": "1700000000",
			"date_to":   "1700086400",
		},
		Secrets: map[string]string{"api_token": "mlsn_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "activity", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read activity: %v", err)
	}
	if sawPath != "/activity/dom_xyz" {
		t.Fatalf("activity path = %q, want /activity/dom_xyz", sawPath)
	}
	if sawFrom != "1700000000" || sawTo != "1700086400" {
		t.Fatalf("date window = (%q,%q), want (1700000000,1700086400)", sawFrom, sawTo)
	}
	if len(got) != 1 || got[0]["id"] != "act_1" {
		t.Fatalf("activity records = %+v, want one act_1", got)
	}
}

// TestActivityMissingDomainErrors ensures the activity stream fails clearly when
// domain_id is absent (it is required by the MailerSend endpoint).
func TestActivityMissingDomainErrors(t *testing.T) {
	c := mailersend.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.mailersend.com/v1"},
		Secrets: map[string]string{"api_token": "mlsn_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "activity", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("activity Read without domain_id should error")
	}
}

// TestFixtureModeNoNetwork verifies credential-free fixture mode emits
// deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mailersend.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"activity", "domains", "messages", "recipients"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := mailersend.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := mailersend.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"activity": false, "domains": false, "messages": false, "recipients": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = mailersend.New() // ensure init ran
	caps := mailersend.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mailersend"); !ok {
		t.Fatal("registry did not resolve mailersend (self-registration)")
	}
}
