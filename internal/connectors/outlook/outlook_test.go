package outlook_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/outlook"
)

func TestReadRefreshesTokenPaginatesAndMaps(t *testing.T) {
	var sawTokenRequest bool
	var sawAuth string
	var graphCalls int
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		sawTokenRequest = true
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"graph_tok","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/me/messages", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		graphCalls++
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			_, _ = w.Write([]byte(`{"value":[{"id":"msg_1","subject":"Hello","receivedDateTime":"2026-01-01T00:00:00Z"}],"@odata.nextLink":"` + srv.URL + `/me/messages?$skiptoken=page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"id":"msg_2","subject":"World","receivedDateTime":"2026-01-02T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected skiptoken %q", r.URL.Query().Get("$skiptoken"))
		}
	})

	c := outlook.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "token_url": srv.URL + "/token", "client_id": "client"},
		Secrets: map[string]string{"client_secret": "secret", "refresh_token": "refresh"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawTokenRequest || sawAuth != "Bearer graph_tok" || graphCalls != 2 {
		t.Fatalf("token/auth/pagination wrong: token=%v auth=%q calls=%d", sawTokenRequest, sawAuth, graphCalls)
	}
	if len(got) != 2 || got[0]["subject"] != "Hello" || got[1]["received_date_time"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("mapped records wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := outlook.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "mail_folders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("outlook"); !ok {
		t.Fatal("registry did not resolve outlook")
	}
}
