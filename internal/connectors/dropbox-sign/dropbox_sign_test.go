package dropboxsign_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	dropboxsign "polymetrics.ai/internal/connectors/dropbox-sign"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Dropbox Sign
// connector: HTTP Basic auth (API key as username, empty password), page-number
// pagination over list_info.num_pages, and record mapping for the
// signature_requests stream. Red until the dropboxsign package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/signature_request/list" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{
				"list_info": {"page": 1, "num_pages": 2, "num_results": 3, "page_size": 2},
				"signature_requests": [
					{"signature_request_id": "sr_1", "title": "NDA", "is_complete": false, "created_at": 1700000000},
					{"signature_request_id": "sr_2", "title": "MSA", "is_complete": true, "created_at": 1700000100}
				]
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"list_info": {"page": 2, "num_pages": 2, "num_results": 3, "page_size": 2},
				"signature_requests": [
					{"signature_request_id": "sr_3", "title": "SOW", "is_complete": false, "created_at": 1700000200}
				]
			}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"list_info":{"page":3,"num_pages":2,"num_results":3,"page_size":2},"signature_requests":[]}`))
		}
	}))
	defer srv.Close()

	c := dropboxsign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "signature_requests", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_abc123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if !strings.Contains(sawAccept, "application/json") {
		t.Fatalf("Accept = %q, want application/json", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["signature_request_id"] == nil {
			t.Fatalf("record missing signature_request_id: %+v", rec)
		}
	}
	if got[0]["title"] != "NDA" {
		t.Fatalf("first record title = %v, want NDA", got[0]["title"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := dropboxsign.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"signature_requests", "templates", "team_members", "account"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %q emitted no records", stream)
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// their primary keys.
func TestCatalogStreams(t *testing.T) {
	c := dropboxsign.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]string{
		"signature_requests": "signature_request_id",
		"templates":          "template_id",
		"team_members":       "account_id",
		"account":            "account_id",
	}
	got := map[string]string{}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) != 1 {
			t.Fatalf("stream %q primary key = %v, want single field", s.Name, s.PrimaryKey)
		}
		got[s.Name] = s.PrimaryKey[0]
	}
	for name, pk := range want {
		if got[name] != pk {
			t.Fatalf("stream %q primary key = %q, want %q", name, got[name], pk)
		}
	}
}

// TestBaseURLValidationRejectsBadScheme guards the SSRF check on base_url.
func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := dropboxsign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "key_abc123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "signature_requests", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http(s) base_url should fail")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = dropboxsign.New() // ensure init ran
	caps := dropboxsign.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("dropbox-sign"); !ok {
		t.Fatal("registry did not resolve dropbox-sign (self-registration)")
	}
}
