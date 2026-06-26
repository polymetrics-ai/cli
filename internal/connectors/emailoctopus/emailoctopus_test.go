package emailoctopus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/emailoctopus"
)

// TestReadPaginatesAndAuthenticates is the red-first test: api_key query auth,
// EmailOctopus page-based pagination across two pages following paging.next, and
// record mapping for the lists stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var nextHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("api_key")
		if r.URL.Path != "/lists" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			// paging.next points at page 2 on this same server.
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"list_1","name":"Foo","created_at":"2026-06-24T07:01:25+00:00","counts":{"subscribed":24,"unsubscribed":1,"pending":0}},` +
				`{"id":"list_2","name":"Bar","created_at":"2026-06-24T08:01:25+00:00","counts":{"subscribed":5,"unsubscribed":0,"pending":2}}` +
				`],"paging":{"next":"` + nextHost + `/lists?api_key=secret&page=2&limit=100","previous":null}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"list_3","name":"Baz","created_at":"2026-06-24T09:01:25+00:00","counts":{"subscribed":3,"unsubscribed":0,"pending":0}}` +
				`],"paging":{"next":null,"previous":null}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[],"paging":{"next":null}}`))
		}
	}))
	defer srv.Close()
	nextHost = srv.URL

	c := emailoctopus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "secret" {
		t.Fatalf("api_key query = %q, want secret", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	// Verify nested counts were flattened by the mapper.
	if got[0]["subscribed_count"] == nil {
		t.Fatalf("expected subscribed_count flattened from counts: %+v", got[0])
	}
}

// TestReadCampaignsMapsRecords exercises the campaigns stream mapper, including
// the nested from object flattening.
func TestReadCampaignsMapsRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[` +
			`{"id":"camp_1","status":"SENT","name":"Foo","subject":"Bar","from":{"name":"John Doe","email_address":"john@example.com"},"created_at":"2026-06-22T07:01:27+00:00","sent_at":"2026-06-23T07:01:27+00:00"}` +
			`],"paging":{"next":null}}`))
	}))
	defer srv.Close()

	c := emailoctopus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["from_email_address"] != "john@example.com" {
		t.Fatalf("from_email_address = %v, want john@example.com", got[0]["from_email_address"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := emailoctopus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"lists", "campaigns", "list_contacts"} {
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
	// Check should succeed in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestBaseURLValidation rejects non-http(s) and hostless overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := emailoctopus.New()
	for _, bad := range []string{"file:///etc/passwd", "ftp://host/x", "://nohost", "https://"} {
		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": bad},
			Secrets: map[string]string{"api_key": "secret"},
		}
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(connectors.Record) error { return nil })
		if err == nil {
			t.Fatalf("base_url=%q should be rejected", bad)
		}
		if !strings.Contains(err.Error(), "base_url") {
			t.Fatalf("base_url=%q error = %v, want base_url validation error", bad, err)
		}
	}
}

// TestRegistryResolvesEmailOctopus confirms self-registration and capabilities.
func TestRegistryResolvesEmailOctopus(t *testing.T) {
	_ = emailoctopus.New() // ensure init ran
	caps := emailoctopus.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("emailoctopus"); !ok {
		t.Fatal("registry did not resolve emailoctopus (self-registration)")
	}
}
