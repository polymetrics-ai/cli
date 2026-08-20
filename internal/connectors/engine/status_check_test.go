package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: method, Path: "/v2/tags", Operation: &SurfaceOperation{Model: "status_check"}}}},
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
	if cookie, ok := result.Headers["Set-Cookie"]; !ok || !cookie.Redacted {
		t.Fatalf("Set-Cookie = %#v, want explicit redaction marker", cookie)
	}
	if token, ok := result.Headers["X-Token"]; !ok || !token.Redacted {
		t.Fatalf("X-Token = %#v, want explicit redaction marker", token)
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
