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
			REST: &RESTOperationSpec{Method: method, Path: "/v2/tags", MaxBytes: 1},
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
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	connector := New(statusCheckBundle(srv.URL, http.MethodHead), nil)
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
