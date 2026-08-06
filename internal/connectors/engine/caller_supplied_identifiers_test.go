package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

func TestCallerSuppliedIdentifierSetsWireEncodings(t *testing.T) {
	bundle, err := Load(os.DirFS("testdata"), "caller-supplied-identifiers")
	if err != nil {
		t.Fatalf("Load caller-supplied identifier test bundle: %v", err)
	}

	const addressA = "eth:0x1111111111111111111111111111111111111111"
	const addressB = "base:0x2222222222222222222222222222222222222222"
	tests := []struct {
		name      string
		operation string
		sets      map[string][]string
		assert    func(*testing.T, *http.Request)
	}{
		{
			name:      "comma separated query",
			operation: "caller.identifiers.comma",
			sets:      map[string][]string{"coins": {"fixture-coin-a", "fixture-coin-b"}},
			assert: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.URL.Query().Get("coins"); got != "fixture-coin-a,fixture-coin-b" {
					t.Fatalf("coins query = %q", got)
				}
			},
		},
		{
			name:      "repeated structured query",
			operation: "caller.identifiers.repeated",
			sets:      map[string][]string{"markets": {addressA, addressB}},
			assert: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.URL.Query()["markets"]; len(got) != 2 || got[0] != addressA || got[1] != addressB {
					t.Fatalf("markets query = %#v", got)
				}
			},
		},
		{
			name:      "JSON body array",
			operation: "caller.identifiers.body",
			sets:      map[string][]string{"ids": {"fixture-id-a", "fixture-id-b"}},
			assert: func(t *testing.T, r *http.Request) {
				t.Helper()
				var body struct {
					IDs []string `json:"ids"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got := strings.Join(body.IDs, ","); got != "fixture-id-a,fixture-id-b" {
					t.Fatalf("body ids = %#v", body.IDs)
				}
			},
		},
		{
			name:      "one element path segment",
			operation: "caller.identifiers.path",
			sets:      map[string][]string{"id": {"fixture-id"}},
			assert: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.URL.Path; got != "/lookups/path/fixture-id" {
					t.Fatalf("path = %q", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.assert(t, r)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer srv.Close()

			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:      tt.operation,
				Config:         connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}},
				IdentifierSets: tt.sets,
			}, nil)
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
		})
	}
}

func TestCallerSuppliedIdentifierSetsRejectBeforeNetworkWithoutValueDisclosure(t *testing.T) {
	bundle, err := Load(os.DirFS("testdata"), "caller-supplied-identifiers")
	if err != nil {
		t.Fatalf("Load caller-supplied identifier test bundle: %v", err)
	}

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	tests := []struct {
		name      string
		operation string
		sets      map[string][]string
		want      []string
		secret    string
		reason    CallerSuppliedIdentifierSetErrorReason
	}{
		{
			name:      "maximum",
			operation: "caller.identifiers.comma",
			sets:      map[string][]string{"coins": {"fixture-coin-a", "fixture-coin-b", "private-coin-never-render"}},
			want:      []string{"coins", "maximum of 2"},
			secret:    "private-coin-never-render",
			reason:    CallerSuppliedIdentifierSetAboveMax,
		},
		{
			name:      "malformed composite",
			operation: "caller.identifiers.repeated",
			sets:      map[string][]string{"markets": {"wallet-value-never-render"}},
			want:      []string{"markets", "element 1", "chain_address"},
			secret:    "wallet-value-never-render",
			reason:    CallerSuppliedIdentifierSetMalformed,
		},
		{
			name:      "comma wire delimiter collision",
			operation: "caller.identifiers.comma",
			sets:      map[string][]string{"coins": {"private-coin-never-render,other"}},
			want:      []string{"coins", "element 1", "opaque_string"},
			secret:    "private-coin-never-render,other",
			reason:    CallerSuppliedIdentifierSetMalformed,
		},
		{
			name:      "absent set",
			operation: "caller.identifiers.body",
			sets:      map[string][]string{},
			want:      []string{"ids", "required"},
			reason:    CallerSuppliedIdentifierSetMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits = 0
			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:      tt.operation,
				Config:         connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}},
				IdentifierSets: tt.sets,
			}, nil)
			if err == nil {
				t.Fatal("OperationDirectRead error = nil, want rejected input")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("OperationDirectRead error = %v, want %q", err, want)
				}
			}
			var rejection *CallerSuppliedIdentifierSetError
			if !errors.As(err, &rejection) || rejection.Reason != tt.reason {
				t.Fatalf("OperationDirectRead error = %#v, want caller-supplied identifier rejection %q", err, tt.reason)
			}
			if tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
				t.Fatalf("OperationDirectRead error leaked supplied identifier: %q", err)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0", hits)
			}
		})
	}
}

func TestCallerSuppliedIdentifierSetsPreserveExplicitEmptyBodyArray(t *testing.T) {
	bundle, err := Load(os.DirFS("testdata"), "caller-supplied-identifiers")
	if err != nil {
		t.Fatalf("Load caller-supplied identifier test bundle: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(body["ids"]); got != "[]" {
			t.Fatalf("body ids = %s, want literal empty JSON array", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	_, err = OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation:      "caller.identifiers.body",
		Config:         connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}},
		IdentifierSets: map[string][]string{"ids": {}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
}

func TestCallerSuppliedIdentifierSetsPreserveExplicitEmptyRepeatedQuery(t *testing.T) {
	bundle, err := Load(os.DirFS("testdata"), "caller-supplied-identifiers")
	if err != nil {
		t.Fatalf("Load caller-supplied identifier test bundle: %v", err)
	}

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		markets, present := r.URL.Query()["markets"]
		if !present || len(markets) != 1 || markets[0] != "" {
			t.Fatalf("markets query = %#v, want one explicit empty value", markets)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	_, err = OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation:      "caller.identifiers.repeated",
		Config:         connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}},
		IdentifierSets: map[string][]string{"markets": {}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if hits != 1 {
		t.Fatalf("server hits = %d, want 1", hits)
	}
}

func TestCallerSuppliedIdentifierSetsRejectSensitiveRepositoryPathBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.repository.lookup", Kind: "rest_read", Summary: "Look up a repository file", Risk: "low", Approval: "none", OutputPolicy: "repository_contents_file_metadata",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/repos/{owner}/contents/{path}", MaxBytes: 1024,
				CallerSuppliedIdentifierSets: []CallerSuppliedIdentifierSetSpec{{Name: "path", ElementShape: "opaque_string", Wire: "path_segment", MinItems: 1, MaxItems: 1}},
			},
		}},
	}

	_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation:      "acme.repository.lookup",
		PathParams:     map[string]string{"owner": "acme"},
		IdentifierSets: map[string][]string{"path": {".env"}},
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want sensitive path rejection")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("OperationDirectRead error = %q, want blocked", err.Error())
	}
	if strings.Contains(err.Error(), ".env") {
		t.Fatalf("OperationDirectRead error leaked supplied identifier: %q", err)
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestCallerSuppliedIdentifierSetsRejectTraversalShapedPathBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.path.lookup", Kind: "rest_read", Summary: "Look up a path identifier", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/lookups/path/{id}?fixed=true", MaxBytes: 1024,
				CallerSuppliedIdentifierSets: []CallerSuppliedIdentifierSetSpec{{Name: "id", ElementShape: "opaque_string", Wire: "path_segment", MinItems: 1, MaxItems: 1}},
			},
		}},
	}

	for _, tt := range []struct {
		name  string
		value string
	}{
		{name: "dot dot segment", value: ".."},
		{name: "encoded slash traversal", value: "private-segment/../target"},
		{name: "backslash traversal", value: "private-segment\\..\\target"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hits = 0
			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:      "acme.path.lookup",
				IdentifierSets: map[string][]string{"id": {tt.value}},
			}, nil)
			if err == nil {
				t.Fatal("OperationDirectRead error = nil, want path traversal rejection")
			}
			if !strings.Contains(err.Error(), "path traversal") {
				t.Fatalf("OperationDirectRead error = %q, want path traversal rejection", err.Error())
			}
			if strings.Contains(err.Error(), tt.value) {
				t.Fatalf("OperationDirectRead error leaked supplied identifier: %q", err)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0", hits)
			}
		})
	}
}

func TestCallerSuppliedIdentifierSetsRejectEmptyRequiredQueryBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.markets.lookup", Kind: "rest_read", Summary: "Look up explicit markets", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/lookups/markets", MaxBytes: 1024,
				RequiredQuery:                []RequiredQueryGroup{{AnyOf: []string{"markets"}}},
				CallerSuppliedIdentifierSets: []CallerSuppliedIdentifierSetSpec{{Name: "markets", ElementShape: "opaque_string", Wire: "query_repeated", MinItems: 0, MaxItems: 2}},
			},
		}},
	}
	tests := []struct {
		name   string
		query  map[string]string
		reason CallerSuppliedIdentifierSetErrorReason
		secret string
	}{
		{name: "generic query collision", query: map[string]string{"markets": "private-value-never-render"}, reason: CallerSuppliedIdentifierSetQueryConflict, secret: "private-value-never-render"},
		{name: "empty required set", reason: CallerSuppliedIdentifierSetEmptyRequiredQuery},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits = 0
			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:      "acme.markets.lookup",
				Query:          tt.query,
				IdentifierSets: map[string][]string{"markets": {}},
			}, nil)
			if err == nil {
				t.Fatal("OperationDirectRead error = nil, want rejected input")
			}
			var rejection *CallerSuppliedIdentifierSetError
			if !errors.As(err, &rejection) || rejection.Reason != tt.reason || rejection.Parameter != "markets" {
				t.Fatalf("OperationDirectRead error = %#v, want caller-supplied identifier rejection %q", err, tt.reason)
			}
			if tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
				t.Fatalf("OperationDirectRead error leaked supplied identifier: %q", err)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0", hits)
			}
		})
	}
}

func TestCallerSuppliedIdentifierSetDeclarationsRejectUnsafeContracts(t *testing.T) {
	tests := []struct {
		name string
		rest string
		want string
	}{
		{
			name: "maximum is required by the closed schema",
			rest: `{"method":"GET","path":"/lookups","caller_supplied_identifier_sets":[{"name":"ids","element_shape":"opaque_string","wire":"query_comma_separated","min_items":0}]}`,
			want: `max_items`,
		},
		{
			name: "maximum must be positive",
			rest: `{"method":"GET","path":"/lookups","caller_supplied_identifier_sets":[{"name":"ids","element_shape":"opaque_string","wire":"query_comma_separated","min_items":0,"max_items":0}]}`,
			want: `max_items must be positive`,
		},
		{
			name: "path set is exactly one item",
			rest: `{"method":"GET","path":"/lookups/{id}","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":0,"max_items":1}]}`,
			want: `path_segment requires min_items and max_items of 1`,
		},
		{
			name: "path set placeholder must be in the pathname",
			rest: `{"method":"GET","path":"/lookups?ids={id}","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":1,"max_items":1}]}`,
			want: `path_segment requires exactly one well-formed {id} path variable in the pathname`,
		},
		{
			name: "path set placeholder must be well formed",
			rest: `{"method":"GET","path":"/lookups/{{id}}","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":1,"max_items":1}]}`,
			want: `path_segment requires exactly one well-formed {id} path variable in the pathname`,
		},
		{
			name: "path set placeholder must not repeat in query",
			rest: `{"method":"GET","path":"/lookups/{id}?mirror={id}","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":1,"max_items":1}]}`,
			want: `path_segment requires exactly one well-formed {id} path variable in the pathname`,
		},
		{
			name: "path set placeholder must not repeat in fragment",
			rest: `{"method":"GET","path":"/lookups/{id}#mirror={id}","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":1,"max_items":1}]}`,
			want: `path_segment requires exactly one well-formed {id} path variable in the pathname`,
		},
		{
			name: "body schema repeats the exact bounds",
			rest: `{"method":"POST","path":"/lookups","content_type":"application/json","body_schema":{"type":"object","required":["ids"],"properties":{"ids":{"type":"array","minItems":0,"maxItems":3,"items":{"type":"string"}}}},"caller_supplied_identifier_sets":[{"name":"ids","element_shape":"opaque_string","wire":"body_json_array","min_items":0,"max_items":2}]}`,
			want: `matching minItems and maxItems`,
		},
		{
			name: "body schema root must allow objects",
			rest: `{"method":"POST","path":"/lookups","content_type":"application/json","body_schema":{"type":"array","properties":{"ids":{"type":"array","minItems":0,"maxItems":2,"items":{"type":"string"}}}},"caller_supplied_identifier_sets":[{"name":"ids","element_shape":"opaque_string","wire":"body_json_array","min_items":0,"max_items":2}]}`,
			want: `body_schema must be an object`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fullValidBundleFS("acme")
			fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"operations":[{"id":"acme.lookup","kind":"rest_read","summary":"Lookup explicit identifiers","risk":"low","approval":"none","output_policy":"json_redacted","rest":%s}]}`, tt.rest))}
			_, err := Load(fsys, "acme")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestCallerSuppliedIdentifierSetPathSegmentAllowsStaticQuery(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{"operations":[{"id":"acme.lookup","kind":"rest_read","summary":"Lookup explicit identifiers","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/lookups/{id}?fixed=true","caller_supplied_identifier_sets":[{"name":"id","element_shape":"opaque_string","wire":"path_segment","min_items":1,"max_items":1}]}}]}`)}
	if _, err := Load(fsys, "acme"); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestCallerSuppliedIdentifierSetStaticQueryIsNotPathTraversal(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if got := r.URL.Query().Get("fixed"); got != ".." {
			t.Fatalf("fixed query = %q, want ..", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.path.lookup", Kind: "rest_read", Summary: "Look up a path identifier", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/lookups/path/{id}?fixed=..", MaxBytes: 1024,
				CallerSuppliedIdentifierSets: []CallerSuppliedIdentifierSetSpec{{Name: "id", ElementShape: "opaque_string", Wire: "path_segment", MinItems: 1, MaxItems: 1}},
			},
		}},
	}

	_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation:      "acme.path.lookup",
		IdentifierSets: map[string][]string{"id": {"fixture-id"}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if hits != 1 {
		t.Fatalf("server hits = %d, want 1", hits)
	}
}
