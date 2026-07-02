package conformance

import (
	"context"
	"io/fs"
	"net/http"
	"testing"
)

// fixtureFSFor returns the fixtures/ subtree of a testdata bundle for direct
// replay-server tests (bypassing engine.Load's full bundle validation, since
// these tests target replay.go in isolation).
func fixtureFSFor(t *testing.T, root, name string) fs.FS {
	t.Helper()
	b := loadTestBundle(t, root, name)
	if b.Fixtures == nil {
		t.Fatalf("bundle %s has no fixtures/ subtree", name)
	}
	return b.Fixtures
}

func TestNewStreamReplayServer_ServesPagesInOrderExactlyOnce(t *testing.T) {
	fixtures := fixtureFSFor(t, "testdata/good", "acme")
	tracker := newHitTracker()
	srv, err := newStreamReplayServer(fixtures, "widgets", tracker)
	if err != nil {
		t.Fatalf("newStreamReplayServer: %v", err)
	}
	defer srv.Close()

	client := srv.Client()
	// page 1: ?page=1 -> 1 record
	resp1, err := client.Get(srv.URL + "/widgets?page=1")
	if err != nil {
		t.Fatalf("GET page 1: %v", err)
	}
	_ = resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("page 1 status = %d, want 200", resp1.StatusCode)
	}

	// page 2: ?page=2 -> empty (stop)
	resp2, err := client.Get(srv.URL + "/widgets?page=2")
	if err != nil {
		t.Fatalf("GET page 2: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("page 2 status = %d, want 200", resp2.StatusCode)
	}

	if got := tracker.hitsFor("widgets"); got != 2 {
		t.Fatalf("hitsFor(widgets) = %d, want 2", got)
	}
}

func TestNewStreamReplayServer_UnmatchedRequestIs404(t *testing.T) {
	fixtures := fixtureFSFor(t, "testdata/good", "acme")
	tracker := newHitTracker()
	srv, err := newStreamReplayServer(fixtures, "widgets", tracker)
	if err != nil {
		t.Fatalf("newStreamReplayServer: %v", err)
	}
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/widgets?page=99")
	if err != nil {
		t.Fatalf("GET unmatched page: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unmatched request status = %d, want 404", resp.StatusCode)
	}
}

func TestLoadFixturePages_ParsesRequestAndResponse(t *testing.T) {
	fixtures := fixtureFSFor(t, "testdata/good", "acme")
	pages, err := loadFixturePages(fixtures, "widgets")
	if err != nil {
		t.Fatalf("loadFixturePages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("loadFixturePages(widgets) returned %d pages, want 2", len(pages))
	}
	if pages[0].Request.Method != "GET" || pages[0].Request.Path != "/widgets" {
		t.Fatalf("pages[0].Request = %+v, want GET /widgets", pages[0].Request)
	}
	if pages[0].Response.Status != 200 {
		t.Fatalf("pages[0].Response.Status = %d, want 200", pages[0].Response.Status)
	}
}

func TestLoadFixturePages_MissingStreamReturnsEmpty(t *testing.T) {
	fixtures := fixtureFSFor(t, "testdata/good", "acme")
	pages, err := loadFixturePages(fixtures, "does-not-exist")
	if err != nil {
		t.Fatalf("loadFixturePages(missing stream): %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("loadFixturePages(missing stream) = %d pages, want 0", len(pages))
	}
}

// context import guard (used indirectly via engine.Read in dynamic checks;
// kept here so replay_test.go compiles standalone if that changes).
var _ = context.Background
