package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// --- schema dialect: the bound must be expressible before it can be enforced ---

func TestCompileSchemaAcceptsMaxItems(t *testing.T) {
	// Before maxItems joined the dialect this failed with
	// `compile schema: unknown keyword "maxItems"`, so a bounded ids[] list
	// could not be declared at all.
	if _, err := CompileSchema([]byte(`{"type":"object","properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":100}}}`)); err != nil {
		t.Fatalf("CompileSchema error = %v, want maxItems accepted", err)
	}
}

// TestCompileSchemaEnforcesOneOf keeps closed declaration unions structural:
// a source binding has exactly one mapping shape, so neither an unbound
// record nor a record carrying both ordinary and tombstone mappings can pass
// schema admission and drift to later runtime validation.
func TestCompileSchemaEnforcesOneOf(t *testing.T) {
	schema, err := CompileSchema([]byte(`{
  "type":"object",
  "properties": {
    "record_mapping": {"type":"object"},
    "tombstone_mapping": {"type":"object"}
  },
  "oneOf": [
    {"required":["record_mapping"]},
    {"required":["tombstone_mapping"]}
  ]
}`))
	if err != nil {
		t.Fatalf("compile oneOf schema: %v", err)
	}
	for name, value := range map[string]map[string]any{
		"ordinary":  {"record_mapping": map[string]any{}},
		"tombstone": {"tombstone_mapping": map[string]any{}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := schema.Validate(value); err != nil {
				t.Fatalf("valid %s mapping rejected: %v", name, err)
			}
		})
	}
	for name, value := range map[string]map[string]any{
		"neither": {},
		"both":    {"record_mapping": map[string]any{}, "tombstone_mapping": map[string]any{}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := schema.Validate(value); err == nil {
				t.Fatalf("invalid %s mapping passed oneOf", name)
			}
		})
	}
}

func TestSchemaEnforcesMaxItems(t *testing.T) {
	sch, err := CompileSchema([]byte(`{"type":"object","properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":2}}}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := sch.Validate(map[string]any{"ids": []any{"a", "b"}}); err != nil {
		t.Fatalf("Validate at the bound = %v, want accepted", err)
	}
	err = sch.Validate(map[string]any{"ids": []any{"a", "b", "c"}})
	if err == nil {
		t.Fatal("Validate over the bound = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "maxItems 2 exceeded") {
		t.Fatalf("Validate error = %q", err.Error())
	}
}

func TestSchemaEnforcesMinItems(t *testing.T) {
	sch, err := CompileSchema([]byte(`{"type":"object","properties":{"ids":{"type":"array","items":{"type":"string"},"minItems":1,"maxItems":5}}}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := sch.Validate(map[string]any{"ids": []any{}}); err == nil {
		t.Fatal("Validate under the minimum = nil, want rejection")
	}
}

func TestCompileSchemaRejectsContradictoryItemBounds(t *testing.T) {
	if _, err := CompileSchema([]byte(`{"type":"array","minItems":5,"maxItems":2}`)); err == nil {
		t.Fatal("CompileSchema error = nil, want minItems>maxItems rejected")
	}
}

// --- provider_search load-time contract ---

func providerSearchOp(mutate func(*OperationSpec)) OperationSpec {
	op := OperationSpec{
		ID:           "search_users",
		Kind:         "provider_search",
		Summary:      "bounded user subset search",
		Risk:         "low",
		Approval:     "none",
		OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method:      "POST",
			Path:        "/users/fetch",
			ContentType: "application/json",
			MaxBytes:    1 << 20,
			BodySchema: json.RawMessage(`{
				"type":"object",
				"additionalProperties":false,
				"required":["ids"],
				"properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":100}}
			}`),
		},
	}
	if mutate != nil {
		mutate(&op)
	}
	return op
}

func TestProviderSearchLoadContract(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*OperationSpec)
		wantErr string
	}{
		{name: "valid", mutate: nil},
		{
			name: "unbounded array is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ids":{"type":"array","items":{"type":"string"}}}}`)
			},
			wantErr: "array without maxItems",
		},
		{
			name: "open body root is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":10}}}`)
			},
			wantErr: "additionalProperties: false",
		},
		{
			name:    "non-POST is refused",
			mutate:  func(o *OperationSpec) { o.REST.Method = "GET" },
			wantErr: "method must be POST",
		},
		{
			name:    "absolute URL is refused",
			mutate:  func(o *OperationSpec) { o.REST.Path = "https://example.com/users/fetch" },
			wantErr: "connector-relative",
		},
		{
			name:    "missing body_schema is refused",
			mutate:  func(o *OperationSpec) { o.REST.BodySchema = nil },
			wantErr: "must declare body_schema",
		},
		{
			name:    "mutating mutation_class is refused",
			mutate:  func(o *OperationSpec) { o.MutationClass = "update" },
			wantErr: "must not declare mutating mutation_class",
		},
		{
			name:    "non-json content type is refused",
			mutate:  func(o *OperationSpec) { o.REST.ContentType = "application/x-www-form-urlencoded" },
			wantErr: "application/json content_type",
		},
		{
			name:    "missing max_bytes is refused",
			mutate:  func(o *OperationSpec) { o.REST.MaxBytes = 0 },
			wantErr: "positive max_bytes",
		},
		{
			name: "nested unbounded array is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"filter":{"type":"object","additionalProperties":false,"properties":{"tags":{"type":"array","items":{"type":"string"}}}}}}`)
			},
			wantErr: "array without maxItems",
		},
		{
			name: "multi-form type list array is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ids":{"type":["array","null"],"items":{"type":"string"}}}}`)
			},
			wantErr: "array without maxItems",
		},
		{
			name: "items-bearing node without a string type is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ids":{"items":{"type":"string"}}}}`)
			},
			wantErr: "array without maxItems",
		},
		{
			name: "nested multi-form array is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"filter":{"type":"object","additionalProperties":false,"properties":{"tags":{"type":["array","null"],"items":{"type":"string"}}}}}}`)
			},
			wantErr: "array without maxItems",
		},
		{
			name: "open nested object is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"filter":{"type":"object","properties":{"tags":{"type":"array","items":{"type":"string"},"maxItems":10}}}}}`)
			},
			wantErr: "must declare additionalProperties: false",
		},
		{
			name: "deeply open nested object is refused",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"filter":{"type":"object","additionalProperties":false,"properties":{"inner":{"type":"object","properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":10}}}}}}}`)
			},
			wantErr: "must declare additionalProperties: false",
		},
		{
			name: "valid with closed nested object",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":100},"filter":{"type":"object","additionalProperties":false,"properties":{"tags":{"type":"array","items":{"type":"string"},"maxItems":20}}}}}`)
			},
			wantErr: "",
		},
		{
			name: "multiple unbounded lists keep a deterministic error",
			mutate: func(o *OperationSpec) {
				o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"zeta":{"type":"array","items":{"type":"string"}},"alpha":{"type":"array","items":{"type":"string"}}}}`)
			},
			wantErr: "body_schema/alpha declares an array without maxItems",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOperationSemantics(0, providerSearchOp(tt.mutate))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateOperationSemantics error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateOperationSemantics error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateOperationSemantics error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// --- provider_search execution, against a real server ---

func TestOperationDirectReadExecutesProviderSearch(t *testing.T) {
	var gotBody map[string]any
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"users":[{"id":"u1"},{"id":"u2"}]}`))
	}))
	defer srv.Close()

	b := providerSearchBundle(srv.URL)
	res, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "search_users",
		Body:      map[string]any{"ids": []any{"u1", "u2"}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead error = %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/users/fetch" {
		t.Fatalf("request = %s %s, want POST /users/fetch", gotMethod, gotPath)
	}
	ids, _ := gotBody["ids"].([]any)
	if len(ids) != 2 {
		t.Fatalf("request body ids = %v, want the two supplied ids", gotBody["ids"])
	}
	if res.Status != http.StatusOK {
		t.Fatalf("status = %d", res.Status)
	}
}

func TestPreflightOperationDirectReadValidatesDeclaredContract(t *testing.T) {
	base := providerSearchBundle("https://api.widget.test")
	tests := []struct {
		name     string
		mutate   func(*Bundle)
		method   string
		path     string
		maxBytes int
		policy   string
		wantErr  string
	}{
		{
			name:     "provider search",
			method:   http.MethodPost,
			path:     "/users/fetch",
			maxBytes: 16 << 20,
			policy:   "json_redacted",
		},
		{
			name: "rest read",
			mutate: func(b *Bundle) {
				b.Operations[0].Kind = "rest_read"
				b.Operations[0].REST.Method = http.MethodGet
			},
			method:   http.MethodGet,
			path:     "/users/fetch",
			maxBytes: 16 << 20,
			policy:   "json_redacted",
		},
		{
			name: "unsupported operation kind",
			mutate: func(b *Bundle) {
				// graphql_query is now a supported fixed-document direct-read
				// kind. graphql_mutation remains intentionally write-only, so it
				// preserves this test's unsupported-read boundary.
				b.Operations[0].Kind = "graphql_mutation"
			},
			method:   http.MethodPost,
			path:     "/users/fetch",
			maxBytes: 16 << 20,
			policy:   "json_redacted",
			wantErr:  "rest_read or provider_search",
		},
		{
			name:     "method mismatch",
			method:   http.MethodGet,
			path:     "/users/fetch",
			maxBytes: 16 << 20,
			policy:   "json_redacted",
			wantErr:  "does not match declared operation method",
		},
		{
			name:     "path mismatch",
			method:   http.MethodPost,
			path:     "/other",
			maxBytes: 16 << 20,
			policy:   "json_redacted",
			wantErr:  "does not match declared operation path",
		},
		{
			name:     "missing command cap",
			method:   http.MethodPost,
			path:     "/users/fetch",
			maxBytes: 0,
			policy:   "json_redacted",
			wantErr:  "requires positive max_bytes",
		},
		{
			name:     "unsupported output policy",
			method:   http.MethodPost,
			path:     "/users/fetch",
			maxBytes: 16 << 20,
			policy:   "json",
			wantErr:  "not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := base
			op := base.Operations[0]
			rest := *op.REST
			op.REST = &rest
			bundle.Operations = []OperationSpec{op}
			if tt.mutate != nil {
				tt.mutate(&bundle)
			}
			err := PreflightOperationDirectRead(bundle, "search_users", tt.method, tt.path, tt.maxBytes, tt.policy)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("PreflightOperationDirectRead error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("PreflightOperationDirectRead error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestOperationDirectReadRejectsOverlongProviderSearchList(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ids := make([]any, 101)
	for i := range ids {
		ids[i] = "u"
	}
	_, err := OperationDirectRead(context.Background(), providerSearchBundle(srv.URL), connectors.OperationDirectReadRequest{
		Operation: "search_users",
		Body:      map[string]any{"ids": ids},
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want the 101-item list rejected")
	}
	if !strings.Contains(err.Error(), "maxItems 100 exceeded") {
		t.Fatalf("error = %q", err.Error())
	}
	if calls != 0 {
		t.Fatalf("HTTP calls = %d, want the bound enforced before any request", calls)
	}
}

// TestOperationDirectReadRejectsUndeclaredProviderSearchBodyKey pins the
// no-escape-hatch invariant: a caller cannot introduce a body key the schema
// does not declare.
func TestOperationDirectReadRejectsUndeclaredProviderSearchBodyKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	_, err := OperationDirectRead(context.Background(), providerSearchBundle(srv.URL), connectors.OperationDirectReadRequest{
		Operation: "search_users",
		Body:      map[string]any{"ids": []any{"u1"}, "sql": "SELECT 1"},
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want the undeclared body key rejected")
	}
}

// TestOperationDirectReadRejectsUndeclaredNestedProviderSearchKey pins that the
// no-escape-hatch invariant holds recursively, not just at the body root: a
// caller cannot inject an undeclared key inside a nested object either.
func TestOperationDirectReadRejectsUndeclaredNestedProviderSearchKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	op := providerSearchOp(func(o *OperationSpec) {
		o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ids":{"type":"array","items":{"type":"string"},"maxItems":100},"filter":{"type":"object","additionalProperties":false,"properties":{"tags":{"type":"array","items":{"type":"string"},"maxItems":20}}}}}`)
	})
	b := providerSearchBundle(srv.URL)
	b.Operations = []OperationSpec{op}

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "search_users",
		Body:      map[string]any{"ids": []any{"u1"}, "filter": map[string]any{"tags": []any{"a"}, "evil": "x"}},
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want the undeclared nested body key rejected")
	}
}

func providerSearchBundle(baseURL string) Bundle {
	return Bundle{
		Name:       "widget-demo",
		HTTP:       HTTPBase{URL: baseURL},
		Operations: []OperationSpec{providerSearchOp(nil)},
	}
}
