package sendpulse_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/sendpulse"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := sendpulse.New()
	if c.Name() != "sendpulse" {
		t.Fatalf("Name() = %q, want sendpulse", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "addressbooks" {
		t.Fatalf("catalog streams = %+v, want addressbooks first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "addressbooks", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("sendpulse"); !ok {
		t.Fatal("registry did not resolve sendpulse")
	}
}

func TestReadAddressBooksUsesClientCredentialsBearer(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/access_token":
			_, _ = w.Write([]byte(`{"access_token":"issued-token","expires_in":3600}`))
		case "/addressbooks":
			sawAuth = r.Header.Get("Authorization") == "Bearer issued-token"
			_, _ = w.Write([]byte(`[{"id":1,"name":"Customers","all_email_qty":3}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "token_url": srv.URL + "/oauth/access_token", "max_pages": "1"}, Secrets: map[string]string{"client_id": "test-client", "client_secret": "test-secret"}}
	var got []connectors.Record
	if err := sendpulse.New().Read(context.Background(), connectors.ReadRequest{Stream: "addressbooks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("oauth bearer auth was not applied")
	}
	if len(got) != 1 || got[0]["name"] != "Customers" {
		t.Fatalf("records = %+v, want Customers address book", got)
	}
}
