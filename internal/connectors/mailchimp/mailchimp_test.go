package mailchimp_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailchimp"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mailchimp
// connector: HTTP Basic auth (anystring:apikey), Mailchimp count/offset
// pagination over the named array (lists[]) using total_items as the stop
// signal, and record mapping. Red until internal/connectors/mailchimp exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/lists" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`{"lists":[{"id":"a1","name":"List A","date_created":"2026-01-01T00:00:00+00:00"},{"id":"a2","name":"List B","date_created":"2026-01-02T00:00:00+00:00"}],"total_items":3}`))
		case 2:
			_, _ = w.Write([]byte(`{"lists":[{"id":"a3","name":"List C","date_created":"2026-01-03T00:00:00+00:00"}],"total_items":3}`))
		default:
			t.Errorf("unexpected offset=%d", offset)
			_, _ = w.Write([]byte(`{"lists":[],"total_items":3}`))
		}
	}))
	defer srv.Close()

	c := mailchimp.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"credentials.apikey": "key-us6"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("anystring:key-us6"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadBearerToken confirms OAuth access_token auth uses a Bearer header.
func TestReadBearerToken(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"campaigns":[{"id":"c1","type":"regular","status":"sent"}],"total_items":1}`))
	}))
	defer srv.Close()

	c := mailchimp.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "tok123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok123" {
		t.Fatalf("Authorization = %q, want Bearer tok123", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "c1" {
		t.Fatalf("records = %+v, want one campaign c1", got)
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records with no credentials and no network access.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := mailchimp.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := mailchimp.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "mailchimp" {
		t.Fatalf("catalog connector = %q, want mailchimp", cat.Connector)
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

func TestRegisteredInRegistry(t *testing.T) {
	_ = mailchimp.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("mailchimp")
	if !ok {
		t.Fatal("registry did not resolve mailchimp (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
