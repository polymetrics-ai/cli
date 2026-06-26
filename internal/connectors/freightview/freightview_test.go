package freightview_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/freightview"
)

// TestReadShipmentsPaginatesAndAuthenticates is the red-first test: it exercises
// the session-token login (client_id/client_secret -> access_token), the Bearer
// auth on data requests, continuationToken pagination across two pages of
// /shipments, and record mapping. Red until internal/connectors/freightview is
// implemented.
func TestReadShipmentsPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawDataAuth  string
		loginBody    string
		loginCalls   int
		shipmentHits int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/auth/token":
			loginCalls++
			buf := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(buf)
			loginBody = string(buf)
			if r.Method != http.MethodPost {
				t.Errorf("login method = %s, want POST", r.Method)
			}
			_, _ = w.Write([]byte(`{"access_token":"sess_tok_abc","expires_in":86400}`))
		case r.URL.Path == "/shipments":
			shipmentHits++
			sawDataAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("continuationToken") {
			case "":
				_, _ = w.Write([]byte(`{"shipments":[{"shipmentId":"s_1","status":"booked"},{"shipmentId":"s_2","status":"delivered"}],"continuationToken":"page2"}`))
			case "page2":
				_, _ = w.Write([]byte(`{"shipments":[{"shipmentId":"s_3","status":"quoted"}],"continuationToken":null}`))
			default:
				t.Errorf("unexpected continuationToken=%q", r.URL.Query().Get("continuationToken"))
				_, _ = w.Write([]byte(`{"shipments":[],"continuationToken":null}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := freightview.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_id": "cid_1", "client_secret": "csecret_1"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "shipments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if loginCalls == 0 {
		t.Fatal("expected at least one login (token) request")
	}
	if !strings.Contains(loginBody, "csecret_1") || !strings.Contains(loginBody, "client_credentials") {
		t.Fatalf("login body = %q, want client_secret + grant_type", loginBody)
	}
	if sawDataAuth != "Bearer sess_tok_abc" {
		t.Fatalf("data Authorization = %q, want Bearer sess_tok_abc", sawDataAuth)
	}
	if shipmentHits != 2 {
		t.Fatalf("shipment requests = %d, want 2 (pagination)", shipmentHits)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["shipmentId"] == nil {
			t.Fatalf("record missing shipmentId: %+v", rec)
		}
	}
}

// TestReadQuotesSubstream verifies the quotes substream fans out over the parent
// shipments and reads /shipments/{id}/quotes with the quotes record selector.
func TestReadQuotesSubstream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/auth/token":
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":86400}`))
		case r.URL.Path == "/shipments":
			_, _ = w.Write([]byte(`{"shipments":[{"shipmentId":"s_1"}],"continuationToken":null}`))
		case r.URL.Path == "/shipments/s_1/quotes":
			_, _ = w.Write([]byte(`{"quotes":[{"quoteId":"q_1","amount":100},{"quoteId":"q_2","amount":200}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := freightview.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "quotes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read quotes: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("quote records = %d, want 2", len(got))
	}
	if got[0]["quoteId"] == nil {
		t.Fatalf("quote record missing quoteId: %+v", got[0])
	}
}

// TestFixtureMode confirms fixture mode emits deterministic records with no
// network access for each core stream.
func TestFixtureMode(t *testing.T) {
	c := freightview.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"shipments", "quotes", "tracking"} {
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
}

// TestCheckFixtureMode verifies Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := freightview.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := freightview.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"shipments": false, "quotes": false, "tracking": false}
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

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := freightview.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "shipments", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = freightview.New()
	caps := freightview.New().Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("freightview"); !ok {
		t.Fatal("registry did not resolve freightview (self-registration)")
	}
}
