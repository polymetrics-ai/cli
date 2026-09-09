package engine

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func statusCheckBundle(baseURL, method string) Bundle {
	return Bundle{
		Name: "acme", HTTP: HTTPBase{URL: baseURL},
		Operations: []OperationSpec{{
			ID: "acme.tags.status", Kind: "rest_status", Summary: "Check tags", Risk: "low", Approval: "none", OutputPolicy: "status",
			REST: &RESTOperationSpec{Method: method, Path: "/v2/tags", MaxBytes: 1, Response: &OperationResponseSpec{SuccessStatuses: []string{"200-299"}}},
		}},
	}
}

func TestOperationStatusCheckUsesDeclaredHEADWithoutJSONBody(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		w.Header().Set("ETag", "abc")
		w.Header().Set("X-Provider-Status", "ordinary-metadata")
		w.Header().Add("Set-Cookie", "transport-secret")
		w.Header().Add("X-Token", "credential-secret")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response = &OperationResponseSpec{SuccessStatuses: []string{"204"}, Headers: []OperationResponseHeaderSpec{
		{Name: "ETag", MaxBytes: 16},
		{Name: "X-Provider-Status", MaxBytes: 32},
		{Name: "Set-Cookie", MaxBytes: 32},
		{Name: "X-Token", MaxBytes: 32},
	}}
	connector := New(bundle, nil)
	if err := connector.PreflightOperationStatusCheck("acme.tags.status", http.MethodHead, "/v2/tags", "status"); err != nil {
		t.Fatalf("PreflightOperationStatusCheck: %v", err)
	}
	result, err := connector.OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err != nil {
		t.Fatalf("OperationStatusCheck: %v", err)
	}
	if result.Status != http.StatusNoContent || result.BodyBytes != 0 || requests != 1 {
		t.Fatalf("result = %+v, requests = %d", result, requests)
	}
	if got := result.Headers["ETag"].Values; len(got) != 1 || got[0] != "abc" {
		t.Fatalf("ETag = %#v, want declared ordinary metadata", got)
	}
	if got := result.Headers["X-Provider-Status"].Values; len(got) != 1 || got[0] != "ordinary-metadata" {
		t.Fatalf("X-Provider-Status = %#v, want declared ordinary metadata", got)
	}
	// Engine results are immutable provider evidence. The commandrunner public
	// boundary preserves provider values equal to configured credentials and
	// never classifies a value from header spelling alone.
	if cookie, ok := result.Headers["Set-Cookie"]; !ok || len(cookie.Values) != 1 || cookie.Values[0] != "transport-secret" {
		t.Fatalf("Set-Cookie = %#v, want exact internal provider metadata", cookie)
	}
	if token, ok := result.Headers["X-Token"]; !ok || len(token.Values) != 1 || token.Values[0] != "credential-secret" {
		t.Fatalf("X-Token = %#v, want exact internal provider metadata", token)
	}
}

func TestOperationStatusCheckPreservesConfiguredEqualResponseHeaderValues(t *testing.T) {
	const credential = "configured-header-material"
	const occurrenceID = "occurrence-9007199254740993"
	const unconfiguredToken = "ghp_unconfigured_provider_token"
	encoded := base64.StdEncoding.EncodeToString([]byte(credential))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		w.Header().Add("X-Provider-Metadata", credential)
		w.Header().Add("X-Provider-Metadata", occurrenceID)
		w.Header().Add("X-Provider-Metadata", encoded)
		w.Header().Add("X-Provider-Metadata", unconfiguredToken)
		w.Header().Add("X-Provider-Metadata", credential)
		w.Header().Set(credential, occurrenceID)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response = &OperationResponseSpec{SuccessStatuses: []string{"204"}, Headers: []OperationResponseHeaderSpec{
		{Name: "X-Provider-Metadata", MaxBytes: 512},
		{Name: credential, MaxBytes: 64},
	}}
	result, err := OperationStatusCheck(context.Background(), bundle, connectors.OperationStatusCheckRequest{
		Operation: "acme.tags.status",
		Config:    connectors.RuntimeConfig{Secrets: map[string]string{"credential": credential}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationStatusCheck: %v", err)
	}
	want := []string{credential, occurrenceID, encoded, unconfiguredToken, credential}
	if !reflect.DeepEqual(result.Headers["X-Provider-Metadata"].Values, want) {
		t.Fatalf("header values = %#v, want %#v", result.Headers["X-Provider-Metadata"].Values, want)
	}
	if result.Headers[credential].Values[0] != occurrenceID {
		t.Fatalf("configured header name was changed: %#v", result.Headers)
	}
}

func TestOperationStatusCheckPreservesTerminalNon2xxMetadataAndDeclaredHeaders(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		w.Header().Set("X-Provider-Status", "not-found")
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response = &OperationResponseSpec{
		SuccessStatuses: []string{"200-299"},
		Headers: []OperationResponseHeaderSpec{{
			Name:     "X-Provider-Status",
			MaxBytes: 32,
		}},
	}

	result, err := New(bundle, nil).OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err != nil {
		t.Fatalf("OperationStatusCheck: %v", err)
	}
	if result.Status != http.StatusNotFound || result.BodyBytes != 0 || requests != 1 {
		t.Fatalf("result = %+v, requests = %d; want final 404 metadata", result, requests)
	}
	if got := result.Headers["X-Provider-Status"].Values; len(got) != 1 || got[0] != "not-found" {
		t.Fatalf("X-Provider-Status = %#v, want final declared 404 header", got)
	}
}

func TestOperationStatusCheckPreservesFinalNon2xxStatus(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	result, err := New(statusCheckBundle(srv.URL, http.MethodHead), nil).OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err != nil {
		t.Fatalf("OperationStatusCheck: %v", err)
	}
	if result.Status != http.StatusNotFound || result.BodyBytes != 0 || requests != 1 {
		t.Fatalf("result = %+v, requests = %d", result, requests)
	}
}

func TestOperationStatusCheckPreservesPostResponseFailureResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		w.Header().Set("X-Trace", "too-large-for-declaration")
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response = &OperationResponseSpec{
		SuccessStatuses: []string{"200-299"},
		Headers:         []OperationResponseHeaderSpec{{Name: "X-Trace", MaxBytes: 1}},
	}
	result, err := OperationStatusCheck(context.Background(), bundle, connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"}, nil)
	if err == nil || !strings.Contains(err.Error(), "X-Trace") {
		t.Fatalf("OperationStatusCheck error = %v, want declared post-response header failure", err)
	}
	if result.Operation != "acme.tags.status" || result.Method != http.MethodHead || result.Path != "/v2/tags" || result.Status != http.StatusBadGateway {
		t.Fatalf("post-response status result = %#v, want complete operation and provider status", result)
	}
	if result.Receipt == nil || !result.Receipt.ResponseReceived || result.Receipt.Status != http.StatusBadGateway || result.Receipt.Body != nil {
		t.Fatalf("post-response receipt = %#v, want bounded raw HEAD metadata without body decode", result.Receipt)
	}
	if got := result.Receipt.Headers["X-Trace"].Values; len(got) != 1 || got[0] != "too-large-for-declaration" {
		t.Fatalf("post-response receipt headers = %#v, want complete provider metadata", result.Receipt.Headers)
	}
}

func TestOperationStatusCheckRejectsNonHEADBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	_, err := New(statusCheckBundle(srv.URL, http.MethodGet), nil).OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err == nil || !strings.Contains(err.Error(), "HEAD") {
		t.Fatalf("OperationStatusCheck error = %v, want HEAD refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
	}
}

func TestOperationStatusCheckRejectsOversizedMetadataCapBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.MaxBytes = defaultOperationStatusMaxBytes + 1
	_, err := New(bundle, nil).OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err == nil || !strings.Contains(err.Error(), "cap") {
		t.Fatalf("OperationStatusCheck error = %v, want bounded-cap refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
	}
}

func TestOperationStatusCheckRejectsMissingMetadataCapBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.MaxBytes = 0
	_, err := New(bundle, nil).OperationStatusCheck(context.Background(), connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"})
	if err == nil || !strings.Contains(err.Error(), "cap") {
		t.Fatalf("OperationStatusCheck error = %v, want bounded-cap refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
	}
}

func TestOperationStatusCheckRejectsMissingOrUndeclaredStatusBeforeBodyHandling(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	bundle := statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response.SuccessStatuses = nil
	if _, err := OperationStatusCheck(context.Background(), bundle, connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"}, nil); err == nil || !strings.Contains(err.Error(), "success_statuses") {
		t.Fatalf("missing status policy error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("missing status policy reached provider %d times", requests)
	}
	bundle = statusCheckBundle(srv.URL, http.MethodHead)
	bundle.Operations[0].REST.Response.SuccessStatuses = []string{"200"}
	if _, err := OperationStatusCheck(context.Background(), bundle, connectors.OperationStatusCheckRequest{Operation: "acme.tags.status"}, nil); err == nil || !strings.Contains(err.Error(), "not declared") {
		t.Fatalf("undeclared status error = %v", err)
	}
	if requests != 1 {
		t.Fatalf("undeclared status requests = %d, want 1", requests)
	}
}

func TestOperationStatusCheckStripsDeclaredHeaderAcrossAllowedRedirect(t *testing.T) {
	seen := make(chan string, 1)
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("X-Mode")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(cdn.Close)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/v2/tags", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	bundle := statusCheckBundle(origin.URL, http.MethodHead)
	bundle.Operations[0].REST.Parameters = []OperationParameter{{
		Name: "X-Mode", In: "header", Type: "string", Schema: []byte(`{"type":"string"}`), MaxBytes: 16,
	}}
	bundle.Operations[0].REST.Redirect = &OperationRedirectSpec{MaxHops: 1, AllowedHosts: []string{strings.TrimPrefix(cdn.URL, "http://")}}
	bundle.Operations[0].REST.Response = &OperationResponseSpec{SuccessStatuses: []string{"204"}}

	result, err := OperationStatusCheck(context.Background(), bundle, connectors.OperationStatusCheckRequest{
		Operation: "acme.tags.status", HeaderValues: map[string][]string{"X-Mode": {"safe"}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationStatusCheck: %v", err)
	}
	if result.Status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusNoContent)
	}
	select {
	case got := <-seen:
		if got != "" {
			t.Fatalf("redirect target received declared header %q", got)
		}
	default:
		t.Fatal("allowed redirect did not reach target")
	}
}

func TestOperationStatusCheckRejectsUndeclaredOrUnsafeParametersBeforeProviderIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("view"); got != "full" {
			t.Errorf("view = %q, want full", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	b := statusCheckBundle(srv.URL, http.MethodHead)
	b.Operations[0].REST.Path = "/v1/tags/{id}"
	b.Operations[0].REST.Parameters = []OperationParameter{
		{Name: "id", In: "path", Type: "string", Required: true, MaxBytes: 8},
		{Name: "view", In: "query", Type: "string", Required: true, MaxBytes: 8},
	}
	for _, tc := range []struct {
		name string
		req  connectors.OperationStatusCheckRequest
		ok   bool
	}{
		{name: "declared parameters", ok: true, req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}, Query: map[string]string{"view": "full"}}},
		{name: "declared path parameter from config", ok: true, req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", Config: connectors.RuntimeConfig{Config: map[string]string{"id": "report"}}, Query: map[string]string{"view": "full"}}},
		{name: "undeclared path parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report", "admin": "true"}, Query: map[string]string{"view": "full"}}},
		{name: "undeclared query parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}, Query: map[string]string{"view": "full", "admin": "true"}}},
		{name: "over cap query parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}, Query: map[string]string{"view": "too-long-value"}}},
		{name: "over cap configured path parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", Config: connectors.RuntimeConfig{Config: map[string]string{"id": "too-long-id"}}, Query: map[string]string{"view": "full"}}},
		{name: "missing required query parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}}},
		{name: "empty required query parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}, Query: map[string]string{"view": ""}}},
		{name: "whitespace required query parameter", req: connectors.OperationStatusCheckRequest{Operation: "acme.tags.status", PathParams: map[string]string{"id": "report"}, Query: map[string]string{"view": "   "}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := requests
			_, err := OperationStatusCheck(context.Background(), b, tc.req, nil)
			if tc.ok {
				if err != nil {
					t.Fatalf("OperationStatusCheck: %v", err)
				}
				if requests != before+1 {
					t.Fatalf("requests = %d, want %d", requests, before+1)
				}
				return
			}
			if err == nil {
				t.Fatal("unsafe binding was accepted")
			}
			if requests != before {
				t.Fatalf("unsafe binding reached provider %d times", requests-before)
			}
		})
	}
}
