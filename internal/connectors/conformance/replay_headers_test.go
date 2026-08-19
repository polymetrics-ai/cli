package conformance

import (
	"net/http"
	"testing"
	"testing/fstest"
)

func TestStreamReplayResponseHeadersExpandBaseURL(t *testing.T) {
	fixtures := fstest.MapFS{
		"streams/widgets/page_1.json": {Data: []byte(`{
			"request":{"method":"GET","path":"/widgets","query":{}},
			"response":{"status":200,"headers":{"Link":["<{{ base_url }}/widgets?cursor=next>; rel=\"next\""]},"body":{"data":[]}}
		}`)},
	}
	srv, err := newStreamReplayServer(fixtures, "widgets", nil)
	if err != nil {
		t.Fatalf("newStreamReplayServer: %v", err)
	}
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/widgets")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got, want := resp.Header.Get("Link"), "<"+srv.URL+"/widgets?cursor=next>; rel=\"next\""; got != want {
		t.Fatalf("Link = %q, want %q", got, want)
	}
}
