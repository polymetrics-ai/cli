package bitly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bitly"
)

// TestReadGroupsAuthenticates is the red-first test: Bearer auth on the
// Authorization header and record mapping for the groups stream (records live at
// body["groups"], no pagination).
func TestReadGroupsAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/groups" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"groups":[{"guid":"g1","name":"Acme","organization_guid":"o1","is_active":true},{"guid":"g2","name":"Beta","organization_guid":"o1","is_active":false}]}`))
	}))
	defer srv.Close()

	c := bitly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "groups", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["guid"] != "g1" || got[0]["name"] != "Acme" {
		t.Fatalf("record[0] = %+v, want guid=g1 name=Acme", got[0])
	}
}

// TestReadBitlinksPaginates exercises Bitly's body cursor pagination: list
// responses carry a pagination.next absolute URL until exhausted. The bitlinks
// endpoint is nested under a group, and records live at body["links"].
func TestReadBitlinksPaginates(t *testing.T) {
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.URL.Path == "/groups/g1/bitlinks" && r.URL.Query().Get("search_after") == "":
			_, _ = w.Write([]byte(`{"links":[{"id":"bit.ly/a","link":"https://bit.ly/a","long_url":"https://example.com/a"}],"pagination":{"next":"` + srv.URL + `/groups/g1/bitlinks?search_after=tok2","search_after":"tok2","size":1}}`))
		case r.URL.Path == "/groups/g1/bitlinks" && r.URL.Query().Get("search_after") == "tok2":
			_, _ = w.Write([]byte(`{"links":[{"id":"bit.ly/b","link":"https://bit.ly/b","long_url":"https://example.com/b"}],"pagination":{"next":"","size":1}}`))
		default:
			t.Errorf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := bitly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "group_guid": "g1"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bitlinks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (across 2 pages); paths=%v", len(got), paths)
	}
	if got[0]["id"] != "bit.ly/a" || got[1]["id"] != "bit.ly/b" {
		t.Fatalf("ids = %v / %v, want bit.ly/a then bit.ly/b", got[0]["id"], got[1]["id"])
	}
	if len(paths) != 2 {
		t.Fatalf("requested %d pages, want 2: %v", len(paths), paths)
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any HTTP call, so conformance can run.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := bitly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"groups", "organizations", "campaigns", "bitlinks"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits without network in fixture
// mode, and Catalog lists the core streams.
func TestCheckAndCatalog(t *testing.T) {
	c := bitly.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "bitly" {
		t.Fatalf("catalog connector = %q, want bitly", cat.Connector)
	}
	want := map[string]bool{"groups": false, "organizations": false, "campaigns": false, "bitlinks": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := bitly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "groups", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url = %v, want base_url scheme error", err)
	}
}

// TestRegistryResolvesBitly confirms self-registration via init().
func TestRegistryResolvesBitly(t *testing.T) {
	_ = bitly.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("bitly"); !ok {
		t.Fatal("registry did not resolve bitly (self-registration)")
	}
	caps := bitly.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
}
