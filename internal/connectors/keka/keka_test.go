package keka_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/keka"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Keka connector:
// it exercises the OAuth2 client-credentials token exchange (api_key + grant_type
// + scope), Bearer auth on the data request, pageNumber/pageSize pagination over
// the `data` array, and record mapping. Red until internal/connectors/keka exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawTokenForm string
	var tokenCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			tokenCalls++
			_ = r.ParseForm()
			sawTokenForm = r.Form.Encode()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":3600}`))
		case "/hris/employees":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("pageNumber") {
			case "", "1":
				_, _ = w.Write([]byte(`{"succeeded":true,"totalPages":2,"pageNumber":1,"data":[{"id":"e1","displayName":"Ada"},{"id":"e2","displayName":"Grace"}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"succeeded":true,"totalPages":2,"pageNumber":2,"data":[{"id":"e3","displayName":"Kay"}]}`))
			default:
				t.Errorf("unexpected pageNumber=%q", r.URL.Query().Get("pageNumber"))
				_, _ = w.Write([]byte(`{"succeeded":true,"totalPages":2,"pageNumber":3,"data":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := keka.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"token_url":  srv.URL + "/connect/token",
			"client_id":  "cid",
			"api_key":    "apikey-xyz",
			"grant_type": "kekaapi",
			"scope":      "kekaapi",
			"page_size":  "2",
		},
		Secrets: map[string]string{"client_secret": "shh"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if tokenCalls == 0 {
		t.Fatal("token endpoint was never called")
	}
	for _, want := range []string{"api_key=apikey-xyz", "grant_type=kekaapi", "scope=kekaapi", "client_id=cid", "client_secret=shh"} {
		if !strings.Contains(sawTokenForm, want) {
			t.Fatalf("token form %q missing %q", sawTokenForm, want)
		}
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

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic records
// without any credentials or network access (conformance-without-creds).
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := keka.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"employees", "leave_requests", "projects"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
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
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := keka.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_secret": "shh"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url validation error for file:// scheme")
	}
}

// TestUnknownStream errors on streams not in the routing table.
func TestUnknownStream(t *testing.T) {
	c := keka.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for unknown stream")
	}
}

// TestCatalogStreams ensures the published catalog covers the core streams.
func TestCatalogStreams(t *testing.T) {
	c := keka.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"employees": false, "attendance": false, "leave_types": false, "leave_requests": false, "clients": false, "projects": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %s has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = keka.New() // ensure init ran
	caps := keka.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("keka is read-only, but Write capability is set")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("keka"); !ok {
		t.Fatal("registry did not resolve keka (self-registration)")
	}
}

// TestErrorsAreNotNil is a tiny guard so errors.Is sanity holds (mirrors stripe).
func TestErrorsAreNotNil(t *testing.T) {
	c := keka.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"base_url": srvScheme()}})
	if err == nil || errors.Is(err, nil) {
		t.Fatal("Check without secret should error")
	}
}

func srvScheme() string { return "https://acme.keka.com/api/v1" }
