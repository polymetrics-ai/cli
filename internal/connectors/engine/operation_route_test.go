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

func routedOperationBundle(baseURL string) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "{{ config.base_url }}", Routes: []OperationRouteSpec{{Name: "v3", BaseURL: "{{ config.base_url }}", Version: "v3"}}},
		Operations: []OperationSpec{
			{ID: "acme.read", Kind: "rest_read", Summary: "Read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", Route: "v3", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/v3/read", MaxBytes: 1024}},
			{ID: "acme.write", Kind: "rest_write", Summary: "Write", Risk: "medium", Approval: "plan preview approval", OutputPolicy: "json_redacted", MutationClass: "create", Confirmation: &ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive}, Route: "v3", REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/v3/write", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"string"}}}`)}},
		},
	}
}

func routedReadRequest(baseURL string) connectors.OperationDirectReadRequest {
	return connectors.OperationDirectReadRequest{Operation: "acme.read", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": baseURL + "/v2"}}}
}
