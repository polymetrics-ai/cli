package microsoftlistsparity_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	microsoftlistshooks "polymetrics.ai/internal/connectors/hooks/microsoft-lists"
	microsoftlists "polymetrics.ai/internal/connectors/microsoft-lists"
)

// loadBundle resolves the "microsoft-lists" bundle from defs.FS via
// engine.Load (single-bundle discovery), NOT LoadAll — see
// microsoft-entra-id's identical parity suite doc comment for the full
// "other in-progress migration agents' directories" reasoning.
func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "microsoft-lists")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, microsoft-lists): %v", err)
	}
	return b
}

func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func runtimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{
		"base_url":  baseURL,
		"token_url": baseURL + "/oauth2/v2.0/token",
		"site_id":   "site1",
	}
	for k, v := range extra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{
		Config: cfg,
		Secrets: map[string]string{
			"client_id":     "client-fixture",
			"client_secret": "secret-fixture",
			"tenant_id":     "tenant-fixture",
		},
	}
}

func readAllRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return out
}

func newEngineConnector(bundle engine.Bundle, baseURL string) connectors.Connector {
	return engine.New(withBaseURL(bundle, baseURL), microsoftlistshooks.New())
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"fixture-access-token","token_type":"Bearer","expires_in":3600}`))
}

// listsTwoPageServer serves /sites/site1/lists with a 2-page
// @odata.nextLink sequence (3 records total) and fails the test on any
// request beyond the second page.
func listsTwoPageServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth2/v2.0/token":
			tokenHandler(w, r)
			return
		case r.URL.Path == "/sites/site1/lists":
			paths = append(paths, r.URL.String())
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Query().Get("$skiptoken") {
			case "":
				_, _ = w.Write([]byte(fmt.Sprintf(`{"value":[{"id":"1","name":"L1","displayName":"List One","list":{"template":"genericList"}},{"id":"2","name":"L2","displayName":"List Two","list":{"template":"genericList"}}],"@odata.nextLink":"%s/sites/site1/lists?$skiptoken=page2"}`, srv.URL)))
			case "page2":
				_, _ = w.Write([]byte(`{"value":[{"id":"3","name":"L3","displayName":"List Three","list":{"template":"genericList"}}]}`))
			default:
				t.Errorf("unexpected 3rd request: %s", r.URL.String())
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &paths
}

// TestParityMicrosoftLists_ListsNextLinkPagination is the authoritative
// substitute named by every stream's conformance.skip_dynamic marker and by
// metadata.json's bundle-level marker: proves both legacy's hand-rolled
// harvest loop and the engine's StreamHook follow @odata.nextLink
// identically across 2 pages/3 records, terminating on nextLink absence,
// with byte-identical emitted records (including the nested list.template
// -> list_template flatten).
func TestParityMicrosoftLists_ListsNextLinkPagination(t *testing.T) {
	legacySrv, legacyPaths := listsTwoPageServer(t)
	t.Cleanup(legacySrv.Close)
	legacy := microsoftlists.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "lists", Config: runtimeConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy lists records = %d, want 3 (2 pages, test fixture bug)", len(legacyRecs))
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requests = %d, want exactly 2; paths=%v", len(*legacyPaths), *legacyPaths)
	}

	bundle := loadBundle(t)
	engSrv, engPaths := listsTwoPageServer(t)
	t.Cleanup(engSrv.Close)
	eng := newEngineConnector(bundle, engSrv.URL)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "lists", Config: runtimeConfig(engSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine lists records = %d, want 3 (2 pages)", len(engRecs))
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requests = %d, want exactly 2 (nextLink absence must stop pagination without a trailing request); paths=%v", len(*engPaths), *engPaths)
	}

	for i := range legacyRecs {
		if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
			t.Fatalf("lists record %d mismatch:\nengine: %+v\nlegacy: %+v", i, engRecs[i], legacyRecs[i])
		}
	}
}

// TestParityMicrosoftLists_ListScopedStreamsRecordShape proves the
// remaining 3 streams' record mapping/schema-projection shape (including
// list_items' $expand=fields query param and its nested contentType.id
// flatten) matches legacy exactly on a single page.
func TestParityMicrosoftLists_ListScopedStreamsRecordShape(t *testing.T) {
	bundle := loadBundle(t)

	cases := []struct {
		stream string
		path   string
		body   string
	}{
		{
			stream: "list_items",
			path:   "/sites/site1/lists/list1/items",
			body:   `{"value":[{"id":"i1","contentType":{"id":"0x0100FIXTURE"},"webUrl":"https://contoso.sharepoint.com/lists/1/items/1","createdDateTime":"2026-01-01T00:00:00Z","lastModifiedDateTime":"2026-01-01T00:00:00Z","eTag":"etag-1","fields":{"Title":"Item 1"}}]}`,
		},
		{
			stream: "columns",
			path:   "/sites/site1/lists/list1/columns",
			body:   `{"value":[{"id":"c1","name":"Title","displayName":"Title","description":"desc","columnGroup":"Custom Columns","required":true,"readOnly":false,"hidden":false,"indexed":false}]}`,
		},
		{
			stream: "content_types",
			path:   "/sites/site1/lists/list1/contentTypes",
			body:   `{"value":[{"id":"0x0100FIXTURE","name":"Item","description":"desc","group":"Custom Content Types","hidden":false,"readOnly":false,"sealed":false}]}`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.stream, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/v2.0/token":
					tokenHandler(w, r)
				case tc.path:
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(tc.body))
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(srv.Close)

			cfg := runtimeConfig(srv.URL, map[string]string{"list_id": "list1"})

			legacy := microsoftlists.New()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: tc.stream, Config: cfg})
			if len(legacyRecs) == 0 || legacyRecs[0]["id"] == nil {
				t.Fatalf("legacy %s emitted no usable records (test fixture bug): %+v", tc.stream, legacyRecs)
			}

			eng := newEngineConnector(bundle, srv.URL)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: tc.stream, Config: cfg})

			if len(engRecs) != len(legacyRecs) {
				t.Fatalf("%s record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", tc.stream, len(engRecs), len(legacyRecs), engRecs, legacyRecs)
			}
			for i := range legacyRecs {
				if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
					t.Fatalf("%s record %d mismatch:\nengine: %+v\nlegacy: %+v", tc.stream, i, engRecs[i], legacyRecs[i])
				}
			}
		})
	}
}

// TestParityMicrosoftLists_MissingListIDErrorsForScopedStream proves the
// hook's list_id validation matches legacy's own needsListID error path for
// a list-scoped stream with no list_id configured.
func TestParityMicrosoftLists_MissingListIDErrorsForScopedStream(t *testing.T) {
	bundle := loadBundle(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/v2.0/token" {
			tokenHandler(w, r)
			return
		}
		t.Fatal("no declarative/legacy request should be sent when list_id is missing for a list-scoped stream")
	}))
	t.Cleanup(srv.Close)

	cfg := runtimeConfig(srv.URL, nil) // no list_id

	legacy := microsoftlists.New()
	if err := legacy.Read(context.Background(), connectors.ReadRequest{Stream: "columns", Config: cfg}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("legacy Read did not error on missing list_id for columns stream")
	}

	eng := newEngineConnector(bundle, srv.URL)
	if err := eng.Read(context.Background(), connectors.ReadRequest{Stream: "columns", Config: cfg}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("engine Read did not error on missing list_id for columns stream")
	}
}
