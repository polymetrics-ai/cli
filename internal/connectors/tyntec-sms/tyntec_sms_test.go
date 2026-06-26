package tyntecsms_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	tyntecsms "polymetrics.ai/internal/connectors/tyntec-sms"
)

func TestReadMessagesAuthenticatesAndMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sms/v1/messages" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("apikey") != "fixture-token" {
			t.Fatal("apikey header was not applied")
		}
		_, _ = w.Write([]byte(`{"messages":[{"id":"msg_1","from":"15550000001","to":"15550000002","status":"DELIVERED","createdAt":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := tyntecsms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key": "fixture-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "msg_1" || got[0]["status"] != "DELIVERED" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := tyntecsms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "tyntec-sms" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["fixture"] != true || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "x"}}); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, ok := connectors.NewRegistry().Get("tyntec-sms"); !ok {
		t.Fatal("tyntec-sms was not self-registered")
	}
}
