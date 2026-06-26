package sendinblue_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/sendinblue"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := sendinblue.New()
	if c.Name() != "sendinblue" {
		t.Fatalf("Name() = %q, want sendinblue", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "contacts" {
		t.Fatalf("catalog streams = %+v, want contacts first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("sendinblue"); !ok {
		t.Fatal("registry did not resolve sendinblue")
	}
}

func TestReadContactsUsesAPIKeyAndContactsKey(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("api-key") == "test-token"
		_, _ = w.Write([]byte(`{"contacts":[{"id":1,"email":"a@example.com","modifiedAt":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "max_pages": "1"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	if err := sendinblue.New().Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("api-key header was not applied")
	}
	if len(got) != 1 || got[0]["email"] != "a@example.com" {
		t.Fatalf("records = %+v, want contact email", got)
	}
}
