package boldsign_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/boldsign"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the BoldSign
// connector: X-API-KEY auth, Page-number pagination over the result[] array,
// and record mapping. The server returns a full page (PageSize records) on
// page 1 to force a second page, then a short page to stop.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	const pageSize = 2
	var sawKey string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-KEY")
		if r.URL.Path != "/v1/document/list" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("Page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "1":
			// Full page (== PageSize) forces another fetch.
			_, _ = w.Write([]byte(`{"result":[{"documentId":"doc_1","status":"Completed"},{"documentId":"doc_2","status":"InProgress"}],"pageDetails":{"page":1,"totalPages":2}}`))
		case "2":
			// Short page stops pagination.
			_, _ = w.Write([]byte(`{"result":[{"documentId":"doc_3","status":"Declined"}],"pageDetails":{"page":2,"totalPages":2}}`))
		default:
			t.Errorf("unexpected Page=%q", page)
			_, _ = w.Write([]byte(`{"result":[]}`))
		}
	}))
	defer srv.Close()

	c := boldsign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": strconv.Itoa(pageSize)},
		Secrets: map[string]string{"api_key": "bs_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "bs_test_key" {
		t.Fatalf("X-API-KEY = %q, want bs_test_key", sawKey)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("pages fetched = %d (%v), want 2", len(pagesSeen), pagesSeen)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["document_id"] != "doc_1" {
		t.Fatalf("record[0].document_id = %v, want doc_1", got[0]["document_id"])
	}
	if got[2]["status"] != "Declined" {
		t.Fatalf("record[2].status = %v, want Declined", got[2]["status"])
	}
}

// TestReadTeamsUsesResultsPath verifies the teams stream reads from the
// "results" envelope (not "result"), which is the documented BoldSign quirk.
func TestReadTeamsUsesResultsPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"teamId":"team_1","teamName":"Legal"}],"pageDetails":{"page":1}}`))
	}))
	defer srv.Close()

	c := boldsign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "bs_test_key"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "teams", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read teams: %v", err)
	}
	if len(got) != 1 || got[0]["team_id"] != "team_1" {
		t.Fatalf("teams records = %+v, want one with team_id=team_1", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call (conformance runs credential-free).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := boldsign.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["document_id"] == nil {
			t.Fatalf("fixture record missing document_id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := boldsign.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"documents": true, "templates": true, "teams": true, "contacts": true, "brands": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Errorf("stream %q has no fields", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

func TestReadOnlyCapabilities(t *testing.T) {
	c := boldsign.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}

func TestRegistryResolvesBoldsign(t *testing.T) {
	_ = boldsign.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("boldsign"); !ok {
		t.Fatal("registry did not resolve boldsign (self-registration)")
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := boldsign.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "bs_test_key"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected for SSRF safety")
	}
}
