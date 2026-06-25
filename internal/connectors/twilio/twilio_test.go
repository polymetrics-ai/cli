package twilio_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/twilio"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Twilio
// connector: HTTP Basic auth (AccountSID:AuthToken), Twilio next_page_uri
// pagination over the "messages" array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var page0Path, page1Path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Query().Get("Page") {
		case "", "0":
			page0Path = r.URL.Path
			// next_page_uri is host-relative, as Twilio returns it.
			_, _ = w.Write([]byte(`{"messages":[` +
				`{"sid":"SM1","date_sent":"Mon, 01 Jan 2024 00:00:00 +0000","from":"+1000","to":"+2000","status":"delivered"},` +
				`{"sid":"SM2","date_sent":"Mon, 01 Jan 2024 00:01:00 +0000","from":"+1000","to":"+2001","status":"sent"}` +
				`],"next_page_uri":"/2010-04-01/Accounts/AC_test/Messages.json?Page=1&PageSize=2","page":0,"page_size":2}`))
		case "1":
			page1Path = r.URL.Path
			_, _ = w.Write([]byte(`{"messages":[` +
				`{"sid":"SM3","date_sent":"Mon, 01 Jan 2024 00:02:00 +0000","from":"+1000","to":"+2002","status":"queued"}` +
				`],"next_page_uri":null,"page":1,"page_size":2}`))
		default:
			t.Errorf("unexpected Page=%q", r.URL.Query().Get("Page"))
			_, _ = w.Write([]byte(`{"messages":[],"next_page_uri":null}`))
		}
	}))
	defer srv.Close()

	c := twilio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"account_sid": "AC_test", "auth_token": "tok_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("AC_test:tok_secret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["sid"] == nil {
			t.Fatalf("record missing sid: %+v", rec)
		}
	}
	// First page must hit the account-scoped Messages endpoint.
	if !strings.Contains(page0Path, "/Accounts/AC_test/Messages.json") {
		t.Fatalf("page0 path = %q, want account-scoped Messages.json", page0Path)
	}
	// next_page_uri must have been followed onto the same Messages endpoint.
	if !strings.Contains(page1Path, "/Accounts/AC_test/Messages.json") {
		t.Fatalf("page1 path = %q, want next_page_uri followed", page1Path)
	}
}

// TestFixtureModeReadsWithoutNetwork verifies credential-free conformance: in
// fixture mode the connector emits deterministic records with no HTTP calls.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := twilio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["sid"] == nil {
		t.Fatalf("fixture record missing sid: %+v", got[0])
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := twilio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckRequiresSecrets(t *testing.T) {
	c := twilio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without secrets should fail")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := twilio.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := twilio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"account_sid": "AC", "auth_token": "tok"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check with non-http base_url should fail (SSRF guard)")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = twilio.New() // ensure init ran
	c := twilio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("twilio is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("twilio"); !ok {
		t.Fatal("registry did not resolve twilio (self-registration)")
	}
}

func TestUnknownStreamErrors(t *testing.T) {
	c := twilio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("unknown stream should error")
	}
}
