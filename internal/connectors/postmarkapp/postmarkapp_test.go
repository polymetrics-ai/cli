package postmarkapp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/postmarkapp"
)

func TestReadOutboundMessagesAuthenticatesPaginatesAndMaps(t *testing.T) {
	var sawToken string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages/outbound" {
			http.NotFound(w, r)
			return
		}
		sawToken = r.Header.Get("X-Postmark-Server-Token")
		pages = append(pages, r.URL.Query().Get("offset"))
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"Messages":[{"MessageID":"m1","To":[{"Email":"a@example.com"}],"Subject":"First","ReceivedAt":"2026-01-01T00:00:00Z"},{"MessageID":"m2","To":[{"Email":"b@example.com"}],"Subject":"Second","ReceivedAt":"2026-01-02T00:00:00Z"}],"TotalCount":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"Messages":[{"MessageID":"m3","To":[{"Email":"c@example.com"}],"Subject":"Third","ReceivedAt":"2026-01-03T00:00:00Z"}],"TotalCount":3}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := postmarkapp.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{
			"X-Postmark-Server-Token":  "server-token",
			"X-Postmark-Account-Token": "account-token",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "outbound_messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "server-token" {
		t.Fatalf("server token header = %q", sawToken)
	}
	if len(pages) != 2 || pages[0] != "0" || pages[1] != "2" {
		t.Fatalf("offsets = %v, want [0 2]", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "m1" || got[0]["subject"] != "First" || got[2]["id"] != "m3" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestReadServersUsesAccountToken(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers" {
			http.NotFound(w, r)
			return
		}
		sawToken = r.Header.Get("X-Postmark-Account-Token")
		_, _ = w.Write([]byte(`{"Servers":[{"ID":10,"Name":"Primary","Color":"blue"}],"TotalCount":1}`))
	}))
	defer srv.Close()

	c := postmarkapp.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{
			"X-Postmark-Server-Token":  "server-token",
			"X-Postmark-Account-Token": "account-token",
		},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "servers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "account-token" {
		t.Fatalf("account token header = %q", sawToken)
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["name"] != "Primary" {
		t.Fatalf("unexpected server records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := postmarkapp.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "outbound_messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "postmarkapp" || len(cat.Streams) < 2 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("postmarkapp"); !ok {
		t.Fatal("registry did not resolve postmarkapp")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
