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
			name: "body schema repeats the exact bounds",
			rest: `{"method":"POST","path":"/lookups","content_type":"application/json","body_schema":{"type":"object","required":["ids"],"properties":{"ids":{"type":"array","minItems":0,"maxItems":3,"items":{"type":"string"}}}},"caller_supplied_identifier_sets":[{"name":"ids","element_shape":"opaque_string","wire":"body_json_array","min_items":0,"max_items":2}]}`,
			want: `matching minItems and maxItems`,
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
