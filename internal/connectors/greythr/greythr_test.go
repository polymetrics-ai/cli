package greythr_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/greythr"
)

// TestReadAuthenticatesAndPaginates is the red-first test for the greytHR
// connector. It exercises the full session-token auth handshake (Basic-auth
// login -> access_token -> ACCESS-TOKEN header on data requests), the
// x-greythr-domain header, PageIncrement pagination over two pages, and record
// mapping from the data[] envelope.
func TestReadAuthenticatesAndPaginates(t *testing.T) {
	var (
		sawLoginAuth string
		sawToken     string
		sawDomain    string
		loginHits    int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/uas/v1/oauth2/client-token":
			loginHits++
			sawLoginAuth = r.Header.Get("Authorization")
			if r.Method != http.MethodPost {
				t.Errorf("login method = %s, want POST", r.Method)
			}
			_, _ = w.Write([]byte(`{"access_token":"tok_abc123","token_type":"bearer"}`))
		case r.URL.Path == "/employee/v2/employees":
			sawToken = r.Header.Get("ACCESS-TOKEN")
			sawDomain = r.Header.Get("x-greythr-domain")
			switch r.URL.Query().Get("page") {
			case "0":
				_, _ = w.Write([]byte(`{"data":[{"employeeId":1,"name":"Ada","email":"ada@x.com"},{"employeeId":2,"name":"Grace","email":"grace@x.com"}]}`))
			case "1":
				_, _ = w.Write([]byte(`{"data":[{"employeeId":3,"name":"Kay","email":"kay@x.com"}]}`))
			default:
				t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
				_, _ = w.Write([]byte(`{"data":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// domain points at the same test server. A full http(s) URL is accepted so
	// the local httptest server (which is http) can be targeted; in production a
	// bare host defaults to https. The login URL is built as
	// <domain>/uas/v1/oauth2/client-token.
	domain := srv.URL

	c := greythr.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"domain":    domain,
			"username":  "user@acme",
			"page_size": "2",
		},
		Secrets: map[string]string{"password": "s3cr3t"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantLogin := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@acme:s3cr3t"))
	if sawLoginAuth != wantLogin {
		t.Fatalf("login Authorization = %q, want %q", sawLoginAuth, wantLogin)
	}
	if sawToken != "tok_abc123" {
		t.Fatalf("ACCESS-TOKEN = %q, want tok_abc123", sawToken)
	}
	if sawDomain != domain {
		t.Fatalf("x-greythr-domain = %q, want %q", sawDomain, domain)
	}
	if loginHits != 1 {
		t.Fatalf("login requested %d times, want exactly 1 (token must be cached)", loginHits)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["employeeId"] == nil {
			t.Fatalf("record missing employeeId: %+v", rec)
		}
	}
}

// TestReadUsersListRootArray verifies the Users List stream, whose records live
// at the response root (no data envelope) and start paging from page 1.
func TestReadUsersListRootArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/uas/v1/oauth2/client-token":
			_, _ = w.Write([]byte(`{"access_token":"tok"}`))
		case "/user/v2/users":
			if r.URL.Query().Get("page") == "1" {
				_, _ = w.Write([]byte(`[{"id":10,"userName":"alice"},{"id":11,"userName":"bob"}]`))
			} else {
				_, _ = w.Write([]byte(`[]`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := greythr.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"domain":    srv.URL,
			"username":  "u",
			"page_size": "2",
		},
		Secrets: map[string]string{"password": "p"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("users = %d, want 2", len(got))
	}
	if got[0]["userName"] != "alice" {
		t.Fatalf("first user = %+v, want userName alice", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so credential-free conformance can run.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := greythr.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := greythr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "greythr" {
		t.Fatalf("catalog connector = %q, want greythr", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
}

func TestMetadataReadOnly(t *testing.T) {
	c := greythr.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("greythr is read-only; Write should be false")
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = greythr.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("greythr"); !ok {
		t.Fatal("registry did not resolve greythr (self-registration)")
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := greythr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil", "domain": "x", "username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}
