package navan_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/navan"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Navan
// connector: it exchanges client credentials for a bearer token at
// /ta-auth/oauth/token, applies that token to the data requests, paginates the
// page-increment /v1/bookings endpoint across two pages, and maps records.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawAuth      string
		sawGrant     string
		sawClientID  string
		tokenCalls   int
		seenPages    []string
		bookingTypes []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/ta-auth/oauth/token":
			tokenCalls++
			_ = r.ParseForm()
			sawGrant = r.PostFormValue("grant_type")
			sawClientID = r.PostFormValue("client_id")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":3600}`))
		case r.URL.Path == "/v1/bookings":
			sawAuth = r.Header.Get("Authorization")
			page := r.URL.Query().Get("page")
			seenPages = append(seenPages, page)
			bookingTypes = append(bookingTypes, r.URL.Query().Get("bookingType"))
			switch page {
			case "0":
				_, _ = w.Write([]byte(`{"data":[{"uuid":"bk_1","bookingType":"FLIGHT","lastModified":"2026-01-01T00:00:00.000Z","grandTotal":120.5,"currency":"USD"},{"uuid":"bk_2","bookingType":"FLIGHT","lastModified":"2026-01-02T00:00:00.000Z","grandTotal":80.0,"currency":"USD"}]}`))
			case "1":
				_, _ = w.Write([]byte(`{"data":[{"uuid":"bk_3","bookingType":"FLIGHT","lastModified":"2026-01-03T00:00:00.000Z","grandTotal":42.0,"currency":"USD"}]}`))
			default:
				_, _ = w.Write([]byte(`{"data":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := navan.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"page_size": "2",
		},
		Secrets: map[string]string{
			"client_id":     "cid_123",
			"client_secret": "csecret_456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalls == 0 {
		t.Fatal("expected an OAuth token exchange, got none")
	}
	if sawGrant != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", sawGrant)
	}
	if sawClientID != "cid_123" {
		t.Fatalf("client_id = %q, want cid_123", sawClientID)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; pages=%v", len(got), seenPages)
	}
	if len(seenPages) < 2 || seenPages[0] != "0" || seenPages[1] != "1" {
		t.Fatalf("pages requested = %v, want at least [0 1]", seenPages)
	}
	for _, rec := range got {
		if rec["uuid"] == nil || rec["last_modified"] == nil {
			t.Fatalf("record missing uuid/last_modified: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records without any HTTP call or live credentials.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := navan.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["uuid"] == nil {
		t.Fatalf("fixture record missing uuid: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreamsAndMetadata locks in the published streams and read-only
// capabilities.
func TestCatalogStreamsAndMetadata(t *testing.T) {
	c := navan.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("navan is read-only; Write should be false, got %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	found := false
	for _, s := range cat.Streams {
		if s.Name == "bookings" {
			found = true
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "uuid" {
				t.Fatalf("bookings primary key = %v, want [uuid]", s.PrimaryKey)
			}
		}
	}
	if !found {
		t.Fatal("bookings stream not in catalog")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := navan.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "x", "client_secret": "y"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

// TestRegistryResolution verifies self-registration via init().
func TestRegistryResolution(t *testing.T) {
	_ = navan.New()
	r := connectors.NewRegistry()
	if _, ok := r.Get("navan"); !ok {
		t.Fatal("registry did not resolve navan (self-registration)")
	}
}

// TestRecordMapperFromJSON sanity-checks the bookings mapper against a raw
// payload decoded the same way connsdk decodes responses.
func TestRecordMapperFromJSON(t *testing.T) {
	raw := `{"uuid":"bk_9","bookingType":"HOTEL","bookingStatus":"CONFIRMED","lastModified":"2026-02-01T00:00:00.000Z","grandTotal":250.0,"currency":"EUR"}`
	var item map[string]any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&item); err != nil {
		t.Fatalf("decode: %v", err)
	}
	rec := navan.MapBookingForTest(item)
	if rec["uuid"] != "bk_9" {
		t.Fatalf("uuid = %v, want bk_9", rec["uuid"])
	}
	if rec["booking_type"] != "HOTEL" {
		t.Fatalf("booking_type = %v, want HOTEL", rec["booking_type"])
	}
	if rec["last_modified"] != "2026-02-01T00:00:00.000Z" {
		t.Fatalf("last_modified = %v", rec["last_modified"])
	}
}
