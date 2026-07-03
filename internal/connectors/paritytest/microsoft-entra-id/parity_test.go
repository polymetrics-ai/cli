package microsoftentraidparity_test

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
	microsoftentraidhooks "polymetrics.ai/internal/connectors/hooks/microsoft-entra-id"
	microsoftentraid "polymetrics.ai/internal/connectors/microsoft-entra-id"
)

// loadBundle resolves the "microsoft-entra-id" bundle from defs.FS via
// engine.Load (single-bundle discovery), NOT LoadAll: defs.FS legitimately
// contains other in-progress migration agents' structurally-incomplete
// directories at any given moment, and LoadAll fails hard on the first
// malformed bundle anywhere in the tree — an unrelated sibling's in-flight
// write should never fail this connector's own parity suite (sentry
// precedent).
func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(defs.FS, "microsoft-entra-id")
	if err != nil {
		t.Fatalf("engine.Load(defs.FS, microsoft-entra-id): %v", err)
	}
	return b
}

func withBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func runtimeConfig(baseURL string, extra map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL, "token_url": baseURL + "/oauth2/v2.0/token"}
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
	return engine.New(withBaseURL(bundle, baseURL), microsoftentraidhooks.New())
}

// tokenHandler answers any /oauth2/v2.0/token POST with a fixture bearer
// token, satisfying both legacy's connsdk.OAuth2ClientCredentials and the
// engine's declarative oauth2_client_credentials auth mode identically.
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"fixture-access-token","token_type":"Bearer","expires_in":3600}`))
}

// usersTwoPageServer serves /users with a 2-page @odata.nextLink sequence
// (3 records total) and fails the test on any request beyond the second
// page — proving termination is driven by nextLink absence, not a fixed
// page count.
func usersTwoPageServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth2/v2.0/token":
			tokenHandler(w, r)
			return
		case r.URL.Path == "/users":
			paths = append(paths, r.URL.String())
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Query().Get("$skiptoken") {
			case "":
				_, _ = w.Write([]byte(fmt.Sprintf(`{"value":[{"id":"1","displayName":"Alice","userPrincipalName":"alice@example.com"},{"id":"2","displayName":"Bob","userPrincipalName":"bob@example.com"}],"@odata.nextLink":"%s/users?$skiptoken=page2"}`, srv.URL)))
			case "page2":
				_, _ = w.Write([]byte(`{"value":[{"id":"3","displayName":"Carol","userPrincipalName":"carol@example.com"}]}`))
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

// TestParityMicrosoftEntraID_UsersNextLinkPagination is the authoritative
// substitute named by every stream's conformance.skip_dynamic marker and by
// metadata.json's bundle-level marker: proves both legacy's hand-rolled
// harvest loop and the engine's StreamHook follow @odata.nextLink
// identically across 2 pages/3 records, terminating on nextLink absence
// (never a fixed page count), with byte-identical emitted records.
func TestParityMicrosoftEntraID_UsersNextLinkPagination(t *testing.T) {
	legacySrv, legacyPaths := usersTwoPageServer(t)
	t.Cleanup(legacySrv.Close)
	legacy := microsoftentraid.New()
	legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: "users", Config: runtimeConfig(legacySrv.URL, nil)})
	if len(legacyRecs) != 3 {
		t.Fatalf("legacy users records = %d, want 3 (2 pages, test fixture bug)", len(legacyRecs))
	}
	if len(*legacyPaths) != 2 {
		t.Fatalf("legacy requests = %d, want exactly 2; paths=%v", len(*legacyPaths), *legacyPaths)
	}

	bundle := loadBundle(t)
	engSrv, engPaths := usersTwoPageServer(t)
	t.Cleanup(engSrv.Close)
	eng := newEngineConnector(bundle, engSrv.URL)
	engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: "users", Config: runtimeConfig(engSrv.URL, nil)})
	if len(engRecs) != 3 {
		t.Fatalf("engine users records = %d, want 3 (2 pages)", len(engRecs))
	}
	if len(*engPaths) != 2 {
		t.Fatalf("engine requests = %d, want exactly 2 (nextLink absence must stop pagination without a trailing request); paths=%v", len(*engPaths), *engPaths)
	}

	for i := range legacyRecs {
		if !reflect.DeepEqual(engRecs[i], legacyRecs[i]) {
			t.Fatalf("users record %d mismatch:\nengine: %+v\nlegacy: %+v", i, engRecs[i], legacyRecs[i])
		}
	}
}

// TestParityMicrosoftEntraID_OtherStreamsRecordShape proves the remaining 4
// streams' record mapping/schema-projection shape matches legacy exactly on
// a single page (pagination termination itself is fully covered by the
// users test above; every stream shares the identical harvest/nextLink
// mechanism).
func TestParityMicrosoftEntraID_OtherStreamsRecordShape(t *testing.T) {
	bundle := loadBundle(t)

	cases := []struct {
		stream string
		path   string
		body   string
	}{
		{
			stream: "groups",
			path:   "/groups",
			body:   `{"value":[{"id":"g1","displayName":"Engineering","description":"Eng team","mail":"eng@example.com","mailNickname":"eng","mailEnabled":true,"securityEnabled":true,"visibility":"Private","createdDateTime":"2026-01-01T00:00:00Z"}]}`,
		},
		{
			stream: "applications",
			path:   "/applications",
			body:   `{"value":[{"id":"a1","appId":"app-1","displayName":"Internal Tool","description":"desc","signInAudience":"AzureADMyOrg","publisherDomain":"example.com","createdDateTime":"2026-01-01T00:00:00Z"}]}`,
		},
		{
			stream: "serviceprincipals",
			path:   "/servicePrincipals",
			body:   `{"value":[{"id":"sp1","appId":"app-1","displayName":"SP One","servicePrincipalType":"Application","accountEnabled":true,"appOwnerOrganizationId":"org-1","signInAudience":"AzureADMyOrg"}]}`,
		},
		{
			stream: "directoryroles",
			path:   "/directoryRoles",
			body:   `{"value":[{"id":"r1","displayName":"Global Administrator","description":"desc","roleTemplateId":"role-tpl-1"}]}`,
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

			legacy := microsoftentraid.New()
			legacyRecs := readAllRecords(t, legacy, connectors.ReadRequest{Stream: tc.stream, Config: runtimeConfig(srv.URL, nil)})
			if len(legacyRecs) == 0 || legacyRecs[0]["id"] == nil {
				t.Fatalf("legacy %s emitted no usable records (test fixture bug): %+v", tc.stream, legacyRecs)
			}

			eng := newEngineConnector(bundle, srv.URL)
			engRecs := readAllRecords(t, eng, connectors.ReadRequest{Stream: tc.stream, Config: runtimeConfig(srv.URL, nil)})

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
