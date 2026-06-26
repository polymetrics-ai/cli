package mailjetmail_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	mailjetmail "polymetrics.ai/internal/connectors/mailjet-mail"
)

// TestReadPaginatesAndAuthenticates is the red-first test: HTTP Basic auth
// (api_key:api_key_secret), Mailjet Limit/Offset pagination over the Data[]
// envelope across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contact" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("Offset"))
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`{"Count":2,"Total":3,"Data":[{"ID":101,"Email":"a@example.com"},{"ID":102,"Email":"b@example.com"}]}`))
		case 2:
			_, _ = w.Write([]byte(`{"Count":1,"Total":3,"Data":[{"ID":103,"Email":"c@example.com"}]}`))
		default:
			t.Errorf("unexpected Offset=%d", offset)
			_, _ = w.Write([]byte(`{"Count":0,"Total":3,"Data":[]}`))
		}
	}))
	defer srv.Close()

	c := mailjetmail.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"api_key":   "pubkey",
			"page_size": "2",
		},
		Secrets: map[string]string{"api_key_secret": "secretkey"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("pubkey:secretkey"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["ID"] == nil || rec["Email"] == nil {
			t.Fatalf("record missing ID/Email: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path so
// conformance can run without live creds or a network.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mailjetmail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
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
	c := mailjetmail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	want := map[string]bool{"contacts": false, "contactslists": false, "messages": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = mailjetmail.New() // ensure init ran
	c := mailjetmail.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mailjet-mail"); !ok {
		t.Fatal("registry did not resolve mailjet-mail (self-registration)")
	}
}
