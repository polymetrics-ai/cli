package encharge_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/encharge"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Encharge
// connector: the X-Encharge-Token API-key header, offset-based pagination over
// the people[] array, and record mapping. Red until internal/connectors/encharge
// exists.
//
// Pagination contract (from the Airbyte manifest): /people/all returns
// {"people":[...]} paged with limit/offset, page_size 100. A short page (fewer
// than limit records) signals the end. The server below returns a full first
// page (100 records) so the offset paginator requests a second page, then a
// short page to stop.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	const pageSize = 100
	var sawToken string
	var requestedOffsets []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-Encharge-Token")
		if r.URL.Path != "/people/all" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		requestedOffsets = append(requestedOffsets, offset)
		if r.URL.Query().Get("limit") != strconv.Itoa(pageSize) {
			t.Errorf("limit = %q, want %d", r.URL.Query().Get("limit"), pageSize)
		}

		var b strings.Builder
		b.WriteString(`{"people":[`)
		switch offset {
		case "", "0":
			// Full page of 100 -> paginator asks for the next page.
			for i := 0; i < pageSize; i++ {
				if i > 0 {
					b.WriteString(",")
				}
				fmt.Fprintf(&b, `{"id":"p_%d","email":"u%d@example.com","name":"User %d"}`, i, i, i)
			}
		case "100":
			// Short page of 1 -> paginator stops.
			b.WriteString(`{"id":"p_last","email":"last@example.com","name":"Last User"}`)
		default:
			t.Errorf("unexpected offset=%q", offset)
		}
		b.WriteString(`]}`)
		_, _ = w.Write([]byte(b.String()))
	}))
	defer srv.Close()

	c := encharge.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "enc_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "peoples", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "enc_test_123" {
		t.Fatalf("X-Encharge-Token = %q, want enc_test_123", sawToken)
	}
	if len(got) != pageSize+1 {
		t.Fatalf("records = %d, want %d (2 pages)", len(got), pageSize+1)
	}
	if len(requestedOffsets) < 2 {
		t.Fatalf("requested %d pages, want >= 2; offsets=%v", len(requestedOffsets), requestedOffsets)
	}
	last := got[len(got)-1]
	if last["id"] != "p_last" || last["email"] != "last@example.com" {
		t.Fatalf("last record mapping wrong: %+v", last)
	}
}

// TestReadSegmentsRecordPath verifies a non-paginated stream whose records live
// under a different JSON key ("segments").
func TestReadSegmentsRecordPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/segments" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"segments":[{"id":"s_1","name":"VIPs","type":"people"},{"id":"s_2","name":"Trials","type":"people"}]}`))
	}))
	defer srv.Close()

	c := encharge.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "enc_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "s_1" || got[0]["name"] != "VIPs" {
		t.Fatalf("segment record mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so credential-free conformance can run.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := encharge.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"peoples", "segments", "fields", "account_tags", "schemas"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}

	// Check in fixture mode must not require creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata verifies the published catalog and read-only
// capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := encharge.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("encharge is read-only; Write capability must be false")
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "encharge" {
		t.Fatalf("catalog connector = %q, want encharge", cat.Connector)
	}
	want := map[string]bool{"peoples": false, "segments": false, "fields": false, "account_tags": false, "schemas": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

// TestRegistryResolution confirms the connector self-registers and resolves via
// the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = encharge.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("encharge")
	if !ok {
		t.Fatal("registry did not resolve encharge (self-registration)")
	}
	if got.Name() != "encharge" {
		t.Fatalf("resolved connector name = %q, want encharge", got.Name())
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := encharge.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "enc_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "peoples", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
