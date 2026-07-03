package microsoftdataverse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newTestRuntime(srv *httptest.Server, cfg connectors.RuntimeConfig) *engine.Runtime {
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    &engine.Bundle{Name: "microsoft-dataverse"},
		Config:    cfg,
	}
}

func TestReadStream_NextLinkPaginationFollowsAbsoluteURL(t *testing.T) {
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			_, _ = fmt.Fprintf(w, `{"value":[{"accountid":"a1","name":"Ada Co","emailaddress1":"ada@example.com"},{"accountid":"a2","name":"Grace Co","emailaddress1":"grace@example.com"}],"@odata.nextLink":"%s/accounts?$skiptoken=page2"}`, srv.URL)
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"accountid":"a3","name":"Cleo Co","emailaddress1":"cleo@example.com"}]}`))
		default:
			t.Errorf("unexpected 3rd request: %s", r.URL.String())
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "accounts"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "accounts", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
	if len(paths) != 2 {
		t.Fatalf("requests = %d, want exactly 2; paths=%v", len(paths), paths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3; got=%+v", len(got), got)
	}
	if got[0]["id"] != "a1" || got[0]["name"] != "Ada Co" || got[2]["name"] != "Cleo Co" {
		t.Fatalf("record mapping wrong: %+v", got)
	}
}

func TestReadStream_NoNextLinkStopsAfterOnePage(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"systemuserid":"u1","fullname":"Backend User","internalemailaddress":"u1@example.com"}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "systemusers"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "systemusers", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want exactly 1 (no @odata.nextLink -> single page)", requests)
	}
	if len(got) != 1 || got[0]["name"] != "Backend User" || got[0]["email"] != "u1@example.com" {
		t.Fatalf("records = %+v, want one Backend systemuser", got)
	}
}

func TestReadStream_LeadNameFallbackChain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"leadid":"l1","subject":"Fallback Subject"}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "leads"}

	var got []connectors.Record
	h := New()
	if _, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "leads", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "Fallback Subject" {
		t.Fatalf("lead name fallback failed: %+v", got)
	}
}

func TestReadStream_UnknownStreamNotHandled(t *testing.T) {
	h := New()
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "widgets"}, connectors.ReadRequest{Stream: "widgets"}, &engine.Runtime{}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream handled = true for unknown stream, want false")
	}
}
