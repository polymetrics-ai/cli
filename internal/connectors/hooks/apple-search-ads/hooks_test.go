package applesearchads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	applesearchadshooks "polymetrics.ai/internal/connectors/hooks/apple-search-ads"
)

// TestReadStream_CampaignsPaginatesOverQuery is the red-first test for the
// GET-with-query access pattern: campaigns paginate with offset/limit query
// params over the {data, pagination} envelope, stopping on a short page.
func TestReadStream_CampaignsPaginatesOverQuery(t *testing.T) {
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		pages++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":111,"name":"Camp A","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"},{"id":222,"name":"Camp B","status":"PAUSED","modificationTime":"2026-01-02T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":333,"name":"Camp C","status":"ENABLED","modificationTime":"2026-01-03T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":2}}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	h := applesearchadshooks.Hooks{}
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	req := connectors.ReadRequest{Stream: "campaigns", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "2"}}}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream(campaigns) handled = false, want true")
	}
	if pages != 2 {
		t.Fatalf("campaign requests = %d, want 2 (pagination across 2 pages)", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil || got[0]["modification_time"] == nil {
		t.Fatalf("record missing mapped id/name/modification_time: %+v", got[0])
	}
}

// TestReadStream_AdgroupsFindPaginatesOverBody verifies the POST .../find
// access pattern: a pagination selector body (not query params), paginating
// across two pages using pagination.totalResults.
func TestReadStream_AdgroupsFindPaginatesOverBody(t *testing.T) {
	var (
		findRequests int
		sawMethod    string
		sawOffsets   []json.Number
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/adgroups/find" {
			http.NotFound(w, r)
			return
		}
		findRequests++
		sawMethod = r.Method
		var body struct {
			Pagination struct {
				Offset json.Number `json:"offset"`
				Limit  json.Number `json:"limit"`
			} `json:"pagination"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawOffsets = append(sawOffsets, body.Pagination.Offset)
		w.Header().Set("Content-Type", "application/json")
		switch body.Pagination.Offset.String() {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":11,"campaignId":111,"name":"AG A","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"},{"id":12,"campaignId":111,"name":"AG B","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":13,"campaignId":111,"name":"AG C","status":"PAUSED","modificationTime":"2026-01-02T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":2}}`))
		default:
			t.Errorf("unexpected offset=%q", body.Pagination.Offset.String())
		}
	}))
	defer srv.Close()

	h := applesearchadshooks.Hooks{}
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	req := connectors.ReadRequest{Stream: "adgroups", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "2"}}}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "adgroups"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream(adgroups) handled = false, want true")
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("find method = %q, want POST", sawMethod)
	}
	if findRequests != 2 {
		t.Fatalf("find requests = %d, want 2 (pagination across 2 pages)", findRequests)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[2]["campaign_id"] == nil || got[2]["name"] == nil {
		t.Fatalf("adgroup record missing mapped campaign_id/name: %+v", got[2])
	}
}

// TestReadStream_UnknownStreamNotHandled verifies ReadStream returns
// handled=false for a stream name this hook set does not recognize, letting
// the declarative fallback (or a "stream not found" error upstream) take
// over instead of silently swallowing it.
func TestReadStream_UnknownStreamNotHandled(t *testing.T) {
	h := applesearchadshooks.Hooks{}
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: "https://example.com"}}
	req := connectors.ReadRequest{Stream: "not-a-real-stream", Config: connectors.RuntimeConfig{}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not-a-real-stream"}, req, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream(unknown) handled = true, want false")
	}
}

// TestReadStream_InvalidPageSizeErrors verifies a malformed page_size config
// value is rejected before any request is issued.
func TestReadStream_InvalidPageSizeErrors(t *testing.T) {
	h := applesearchadshooks.Hooks{}
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: "https://example.com"}}
	req := connectors.ReadRequest{Stream: "campaigns", Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": "not-a-number"}}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, req, rt, func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("ReadStream(campaigns) with invalid page_size handled = false, want true (handled, but erroring)")
	}
	if err == nil {
		t.Fatal("ReadStream(campaigns) with invalid page_size should error")
	}
}

func TestConnectorName(t *testing.T) {
	h := applesearchadshooks.Hooks{}
	if h.ConnectorName() != "apple-search-ads" {
		t.Fatalf("ConnectorName() = %q, want apple-search-ads", h.ConnectorName())
	}
}
