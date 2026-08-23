package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func operationBindingTestBundle(serverURL string, op OperationSpec) Bundle {
	surface := &APISurface{Endpoints: []SurfaceEndpoint{{
		Method: op.REST.Method,
		Path:   op.REST.Path,
		CoveredBy: &SurfaceCoverage{
			DirectReads: []string{"fixture"},
		},
	}}}
	bundle := Bundle{
		Name:       "acme",
		HTTP:       HTTPBase{URL: serverURL},
		Operations: []OperationSpec{op},
		Surface:    surface,
	}
	bundle.directReadLedger = deriveOperationDirectReadEndpointLedger(bundle.Operations, surface)
	return bundle
}

func TestOperationParametersEnforceEncodedPathAndQueryByteCaps(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	op := OperationSpec{
		ID: "acme.widgets.get", Kind: "rest_read", Summary: "get widget", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodGet, Path: "/widgets/{id}", MaxBytes: 1024,
			Parameters: []OperationParameter{
				{Name: "id", In: "path", Type: "string", Required: true, MaxBytes: 6},
				{Name: "filter", In: "query", Type: "string", MaxBytes: 6},
			},
		},
	}
	bundle := operationBindingTestBundle(srv.URL, op)

	if _, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: op.ID, PathParams: map[string]string{"id": "é"}, Query: map[string]string{"filter": "é"},
	}, nil); err != nil {
		t.Fatalf("encoded values at cap: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1", hits.Load())
	}

	for _, testCase := range []struct {
		name string
		req  connectors.OperationDirectReadRequest
	}{
		{name: "path cap+1 after percent encoding", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"id": "éé"}}},
		{name: "query cap+1 after percent encoding", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"id": "ok"}, Query: map[string]string{"filter": "éé"}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := OperationDirectRead(context.Background(), bundle, testCase.req, nil); err == nil || !strings.Contains(err.Error(), "byte cap") {
				t.Fatalf("OperationDirectRead error = %v, want encoded byte-cap rejection", err)
			}
		})
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits after rejected values = %d, want 1", hits.Load())
	}
}

func TestOperationParametersReportPMExecutionCapBeforeIO(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	op := OperationSpec{
		ID: "acme.search", Kind: "rest_read", Summary: "search", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodGet, Path: "/search", MaxBytes: 1024,
			Parameters: []OperationParameter{{Name: "workspaceId", In: "query", Type: "string", Required: true, MaxBytes: 6}},
		},
	}
	bundle := operationBindingTestBundle(srv.URL, op)
	if _, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: op.ID,
		Query:     map[string]string{"workspaceId": "é"},
	}, nil); err != nil {
		t.Fatalf("exact encoded cap: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1", hits.Load())
	}
	_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: op.ID,
		Query:     map[string]string{"workspaceId": "éé"},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "PM execution limit") || !strings.Contains(err.Error(), "encoded_bytes") || !strings.Contains(err.Error(), "6") {
		t.Fatalf("over-cap error = %v, want PM execution limit, encoded_bytes, and effective cap", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("over-cap value reached provider; hits = %d, want 1", hits.Load())
	}
}

func TestOperationParametersPreserveExactFiniteNumericLexemes(t *testing.T) {
	var hits atomic.Int32
	var gotQuery atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		hits.Add(1)
		gotQuery.Store(request.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	integerMinimum := connectors.ExactNumber("9999999999999999999999999999999999999998")
	numberMaximum := connectors.ExactNumber("0.123456789012345678901234567890123456790")
	op := OperationSpec{
		ID: "acme.numeric", Kind: "rest_read", Summary: "numeric", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodGet, Path: "/numeric", MaxBytes: 1024,
			Parameters: []OperationParameter{
				{Name: "integer", In: "query", Type: "integer", Required: true, Minimum: &integerMinimum, MaxBytes: 128},
				{Name: "number", In: "query", Type: "number", Required: true, Maximum: &numberMaximum, MaxBytes: 128},
			},
		},
	}
	integer := "9999999999999999999999999999999999999999"
	number := "0.123456789012345678901234567890123456789"
	_, err := OperationDirectRead(context.Background(), operationBindingTestBundle(srv.URL, op), connectors.OperationDirectReadRequest{
		Operation: op.ID,
		Query:     map[string]string{"integer": integer, "number": number},
	}, nil)
	if err != nil {
		t.Fatalf("exact finite numerics: %v", err)
	}
	query, _ := gotQuery.Load().(url.Values)
	if query.Get("integer") != integer || query.Get("number") != number {
		t.Fatalf("wire numerics = integer %q number %q, want exact source lexemes", query.Get("integer"), query.Get("number"))
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1", hits.Load())
	}
	for _, test := range []struct {
		name  string
		query map[string]string
		want  string
	}{
		{name: "below exact minimum", query: map[string]string{"integer": "9999999999999999999999999999999999999997", "number": number}, want: integerMinimum.String()},
		{name: "above exact maximum", query: map[string]string{"integer": integer, "number": "0.123456789012345678901234567890123456791"}, want: numberMaximum.String()},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := OperationDirectRead(context.Background(), operationBindingTestBundle(srv.URL, op), connectors.OperationDirectReadRequest{Operation: op.ID, Query: test.query}, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("numeric bound error = %v, want exact bound %s", err, test.want)
			}
		})
	}
	oversizedExponent := "1e" + strings.Repeat("9", 256)
	_, err = OperationDirectRead(context.Background(), operationBindingTestBundle(srv.URL, op), connectors.OperationDirectReadRequest{
		Operation: op.ID,
		Query:     map[string]string{"integer": integer, "number": oversizedExponent},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "PM execution limit") || !strings.Contains(err.Error(), "encoded_bytes") {
		t.Fatalf("oversized numeric lexeme error = %v, want pre-parse PM byte limit", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("out-of-bound numerics reached provider; hits = %d, want 1", hits.Load())
	}
}

func TestOperationDirectReadValidatesEffectiveConfigPathParamsBeforeIO(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	for _, test := range []struct {
		name       string
		parameter  OperationParameter
		configID   string
		pathParams map[string]string
		ok         bool
	}{
		{name: "configured integer", parameter: OperationParameter{Name: "id", In: "path", Type: "integer", Required: true}, configID: "not-an-integer"},
		{name: "configured enum", parameter: OperationParameter{Name: "id", In: "path", Type: "enum", Required: true, Values: []string{"alpha"}}, configID: "beta"},
		{name: "configured encoded byte cap", parameter: OperationParameter{Name: "id", In: "path", Type: "string", Required: true, MaxBytes: 6}, configID: "éé"},
		{name: "valid config fallback", parameter: OperationParameter{Name: "id", In: "path", Type: "enum", Required: true, Values: []string{"alpha"}}, configID: "alpha", ok: true},
		{name: "caller overrides invalid config", parameter: OperationParameter{Name: "id", In: "path", Type: "integer", Required: true}, configID: "not-an-integer", pathParams: map[string]string{"id": "42"}, ok: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			op := OperationSpec{
				ID: "acme.widgets.get", Kind: "rest_read", Summary: "get widget", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
				REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/widgets/{id}", MaxBytes: 1024, Parameters: []OperationParameter{test.parameter}},
			}
			before := hits.Load()
			_, err := OperationDirectRead(context.Background(), operationBindingTestBundle(srv.URL, op), connectors.OperationDirectReadRequest{
				Operation: op.ID,
				Config: connectors.RuntimeConfig{Config: map[string]string{
					"id": test.configID,
				}},
				PathParams: test.pathParams,
			}, nil)
			if test.ok {
				if err != nil {
					t.Fatalf("OperationDirectRead: %v", err)
				}
				if hits.Load() != before+1 {
					t.Fatalf("provider hits = %d, want %d", hits.Load(), before+1)
				}
				return
			}
			if err == nil {
				t.Fatal("invalid configured path parameter was accepted")
			}
			if hits.Load() != before {
				t.Fatalf("invalid configured path parameter reached provider %d times", hits.Load()-before)
			}
		})
	}
}

func TestOperationDirectReadRejectsBlankRequiredQueryBeforeIO(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	op := OperationSpec{
		ID: "acme.widgets.search", Kind: "rest_read", Summary: "search widgets", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodGet, Path: "/widgets", MaxBytes: 1024,
			Parameters: []OperationParameter{{Name: "query", In: "query", Type: "string", Required: true}},
		},
	}
	bundle := operationBindingTestBundle(srv.URL, op)
	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := hits.Load()
			_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation: op.ID,
				Query:     map[string]string{"query": test.value},
			}, nil)
			if err == nil {
				t.Fatal("blank required query parameter was accepted")
			}
			if hits.Load() != before {
				t.Fatalf("blank required query parameter reached provider %d times", hits.Load()-before)
			}
		})
	}
}

func TestOperationDirectReadClosedBindingsRejectUnknownsBeforeIO(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	op := OperationSpec{
		ID: "acme.widgets.search", Kind: "rest_read", Summary: "search", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{
			Method: http.MethodPost, Path: "/widgets/{tenant}", ContentType: "application/json", MaxBytes: 1024,
			Query: map[string]string{"mode": "exact"},
			Parameters: []OperationParameter{
				{Name: "tenant", In: "path", Type: "string", Required: true},
				{Name: "cursor", In: "query", Type: "string"},
			},
			BodySchema: json.RawMessage(`{
				"type":"object","additionalProperties":false,
				"properties":{"name":{"type":"string"}}
			}`),
		},
	}
	bundle := operationBindingTestBundle(srv.URL, op)

	for _, testCase := range []struct {
		name string
		req  connectors.OperationDirectReadRequest
	}{
		{name: "unknown path", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"tenant": "acme", "extra": "escape"}}},
		{name: "unknown query", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"tenant": "acme"}, Query: map[string]string{"extra": "escape"}}},
		{name: "fixed query collision", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"tenant": "acme"}, Query: map[string]string{"mode": "caller"}}},
		{name: "unknown body", req: connectors.OperationDirectReadRequest{Operation: op.ID, PathParams: map[string]string{"tenant": "acme"}, Body: map[string]any{"escape": true}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := OperationDirectRead(context.Background(), bundle, testCase.req, nil); err == nil {
				t.Fatal("OperationDirectRead accepted undeclared caller authority")
			}
		})
	}
	if hits.Load() != 0 {
		t.Fatalf("provider hits = %d, want no I/O for rejected bindings", hits.Load())
	}

	if err := PreflightOperationDirectReadBindings(bundle, op.ID,
		[]string{"tenant", "extra"}, []string{"cursor"}, []string{"name"}, false); err == nil {
		t.Fatal("binding preflight accepted an undeclared path mapping")
	}
}

func TestOperationDirectReadLegacyCommandBindingIsSealedAndRevalidated(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hits.Add(1)
		if got := req.URL.Query().Get("filter"); got != "active" {
			t.Errorf("filter = %q, want active", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	op := OperationSpec{
		ID: "acme.widgets.legacy", Kind: "rest_read", Summary: "legacy search", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/widgets", MaxBytes: 1024},
	}
	bundle := operationBindingTestBundle(srv.URL, op)
	bundle.CLISurface = &CLISurface{Commands: []CLICommand{{
		Path: "widgets legacy", Intent: "direct_read", Availability: "implemented", Operation: op.ID,
		Flags: []CLIFlag{{Name: "filter", Type: "string", MapsTo: "query.filter"}},
	}}}
	bindings := &connectors.OperationDirectReadBindings{Query: []string{"filter"}}
	if err := PreflightOperationDirectReadBindings(bundle, op.ID, nil, bindings.Query, nil, false); err != nil {
		t.Fatalf("legacy descriptor preflight: %v", err)
	}
	if _, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: op.ID, Query: map[string]string{"filter": "active"}, CommandBindings: bindings,
	}, nil); err != nil {
		t.Fatalf("sealed legacy descriptor read: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1", hits.Load())
	}

	for _, request := range []connectors.OperationDirectReadRequest{
		{Operation: op.ID, Query: map[string]string{"filter": "active"}},
		{Operation: op.ID, Query: map[string]string{"escape": "true"}, CommandBindings: &connectors.OperationDirectReadBindings{Query: []string{"escape"}}},
	} {
		if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err == nil {
			t.Fatal("operation accepted an unsealed or forged legacy query binding")
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("provider hits after rejected bindings = %d, want 1", hits.Load())
	}
}
