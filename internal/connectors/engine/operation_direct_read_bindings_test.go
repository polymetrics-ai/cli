package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
