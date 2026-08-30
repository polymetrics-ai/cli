package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestOperationRoutesFailClosedBeforeProviderIO(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()

	base := routedOperationBundle(srv.URL)
	base.Operations[0].Route = "absent"
	_, err := OperationDirectRead(context.Background(), base, routedReadRequest(srv.URL), nil)
	var missing *MissingOperationRouteError
	if !errors.As(err, &missing) {
		t.Fatalf("OperationDirectRead missing route error = %v, want MissingOperationRouteError", err)
	}
	if !strings.Contains(err.Error(), "source=https://provider.example.test/read") || !strings.Contains(err.Error(), "is blocked: missing route foundation") {
		t.Fatalf("missing route diagnostic = %q, want source-traced blocked foundation", err)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("missing route provider hits = %d, want 0", got)
	}

	base = routedOperationBundle(srv.URL)
	base.Operations[0].REST.Path = "/v2/read"
	_, err = OperationDirectRead(context.Background(), base, routedReadRequest(srv.URL), nil)
	if !errors.As(err, &missing) || !strings.Contains(err.Error(), "does not match declared path") {
		t.Fatalf("version mismatch error = %v, want source-traced missing foundation", err)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("version mismatch provider hits = %d, want 0", got)
	}
}

func TestOperationRoutesUseOneDeclaredRouteForDirectReadAndWrite(t *testing.T) {
	var reads, writes atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/read":
			reads.Add(1)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/write":
			writes.Add(1)
		default:
			t.Fatalf("unexpected provider request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	bundle := routedOperationBundle(srv.URL)
	if _, err := OperationDirectRead(context.Background(), bundle, routedReadRequest(srv.URL), nil); err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}

	request := connectors.OperationDirectWriteRequest{
		Operation: "acme.write",
		Body:      map[string]any{"value": "fixture"},
		Config: connectors.RuntimeConfig{
			Config:              map[string]string{"base_url": srv.URL + "/v2"},
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	request.PreviewDigest = preview.Digest
	request.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if got := reads.Load(); got != 1 {
		t.Fatalf("direct read requests = %d, want 1", got)
	}
	if got := writes.Load(); got != 1 {
		t.Fatalf("direct write requests = %d, want 1", got)
	}
}

func TestOperationRoutesRejectConflictingBasesBeforeProviderIO(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()

	bundle := routedOperationBundle(srv.URL)
	bundle.HTTP.Routes = append(bundle.HTTP.Routes, OperationRouteSpec{Name: "v3", BaseURL: "https://conflict.example.test", Version: "v3"})
	_, err := OperationDirectRead(context.Background(), bundle, routedReadRequest(srv.URL), nil)
	if err == nil || !strings.Contains(err.Error(), "base route \"v3\" declares conflicting bases") {
		t.Fatalf("conflicting route error = %v, want exact base conflict", err)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("conflicting bases provider hits = %d, want 0", got)
	}
}

func TestSourceBoundOperationRejectsConfiguredOriginBeforeAuthenticationOrIO(t *testing.T) {
	var declaredHits, overrideHits atomic.Int64
	declared := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { declaredHits.Add(1) }))
	defer declared.Close()
	override := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { overrideHits.Add(1) }))
	defer override.Close()

	bundle := routedOperationBundle(declared.URL)
	specRaw, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"base_url": map[string]any{"type": "string", "default": declared.URL + "/v2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Spec, err = CompileSchema(specRaw)
	if err != nil {
		t.Fatalf("compile declared test configuration: %v", err)
	}
	bundle.Operations[0].SourceOperation = &SourceOperationBinding{ID: "acme.rest.getRead", Method: http.MethodGet, Path: "/v3/read"}
	admission := &capturingAuthenticationAdmission{}
	_, err = New(bundle, nil).OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "acme.read",
		Config: connectors.RuntimeConfig{
			Config:                  map[string]string{"base_url": override.URL + "/v2"},
			AuthenticationAdmission: admission,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "rejects configured base_url override") {
		t.Fatalf("configured source-bound origin error = %v, want closed origin refusal", err)
	}
	if admission.calls != 0 || declaredHits.Load() != 0 || overrideHits.Load() != 0 {
		t.Fatalf("origin override reached auth or I/O: auth=%d declared=%d override=%d", admission.calls, declaredHits.Load(), overrideHits.Load())
	}

	streamBundle := Bundle{
		Name:    "acme-stream",
		HTTP:    HTTPBase{URL: "{{ config.base_url }}", Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "next"}},
		Spec:    bundle.Spec,
		Streams: []StreamSpec{{Name: "items", Path: "/items", Records: RecordsSpec{Path: "data"}, SchemaRef: "schemas/items.json"}},
		Operations: []OperationSpec{{
			ID: "acme.items", Kind: "stream_etl", SourceOperation: &SourceOperationBinding{ID: "acme.rest.getItems", Method: http.MethodGet, Path: "/items"},
			Composite: &CompositeOperationSpec{Steps: []string{"stream:items"}},
		}},
	}
	admission = &capturingAuthenticationAdmission{}
	err = New(streamBundle, nil).Read(context.Background(), connectors.ReadRequest{
		Stream: "items",
		Config: connectors.RuntimeConfig{
			Config:                  map[string]string{"base_url": override.URL + "/v2"},
			AuthenticationAdmission: admission,
		},
	}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "rejects configured base_url override") {
		t.Fatalf("configured source-bound stream origin error = %v, want closed origin refusal", err)
	}
	if admission.calls != 0 || declaredHits.Load() != 0 || overrideHits.Load() != 0 {
		t.Fatalf("stream origin override reached auth or I/O: auth=%d declared=%d override=%d", admission.calls, declaredHits.Load(), overrideHits.Load())
	}
}

func routedOperationBundle(baseURL string) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "{{ config.base_url }}", Routes: []OperationRouteSpec{{Name: "v3", BaseURL: "{{ config.base_url }}", Version: "v3"}}},
		Operations: []OperationSpec{
			{ID: "acme.read", Kind: "rest_read", Summary: "Read", SourceURL: "https://provider.example.test/read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", Route: "v3", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/v3/read", MaxBytes: 1024}},
			{ID: "acme.write", Kind: "rest_write", Summary: "Write", SourceURL: "https://provider.example.test/write", Risk: "medium", Approval: "plan preview approval", OutputPolicy: "json_redacted", MutationClass: "create", Confirmation: &ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive}, Route: "v3", REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/v3/write", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"string"}}}`)}},
		},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/v3/read", Operation: &SurfaceOperation{Model: "direct_read"}}, {Method: http.MethodPost, Path: "/v3/write", Operation: &SurfaceOperation{Model: "direct_write"}}}},
	}
}

func routedReadRequest(baseURL string) connectors.OperationDirectReadRequest {
	return connectors.OperationDirectReadRequest{Operation: "acme.read", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": baseURL + "/v2"}}}
}
