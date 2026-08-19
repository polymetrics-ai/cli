package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func textExportBundle(baseURL string, maxBytes int) Bundle {
	return Bundle{Name: "acme", HTTP: HTTPBase{URL: baseURL}, Operations: []OperationSpec{{
		ID: "acme.audit.export", Kind: "text_export", Summary: "Export audit CSV", Risk: "medium", Approval: "explicit destination", OutputPolicy: "file_manifest",
		Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/v2/audit.csv", MaxBytes: maxBytes, Accept: "text/csv"},
	}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/v2/audit.csv", Operation: &SurfaceOperation{Model: "text_export"}}}}}
}

func TestOperationTextExportWritesBoundedCSV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("id,name\n1,ada\n"))
	}))
	t.Cleanup(srv.Close)
	result, err := OperationBinaryDownload(context.Background(), textExportBundle(srv.URL, 1024), BinaryDownloadRequest{Operation: "acme.audit.export", DestRoot: t.TempDir(), FileName: "audit.csv"}, nil)
	if err != nil || result.Record["file_size_bytes"] == nil {
		t.Fatalf("OperationBinaryDownload result/error = %#v / %v", result, err)
	}
}

func TestOperationTextExportRejectsUnboundedBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	_, err := OperationBinaryDownload(context.Background(), textExportBundle(srv.URL, 0), BinaryDownloadRequest{Operation: "acme.audit.export", DestRoot: t.TempDir()}, nil)
	if err == nil || !strings.Contains(err.Error(), "positive max_bytes") {
		t.Fatalf("OperationBinaryDownload error = %v, want bounded-export refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
	}
}

func TestOperationTextExportOverCapLeavesNoArtifact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("id,name\n1,ada\n2,grace\n"))
	}))
	t.Cleanup(srv.Close)
	dest := t.TempDir()
	_, err := OperationBinaryDownload(context.Background(), textExportBundle(srv.URL, 8), BinaryDownloadRequest{Operation: "acme.audit.export", DestRoot: dest, FileName: "audit.csv"}, nil)
	if err == nil {
		t.Fatal("OperationBinaryDownload error = nil, want cap rejection")
	}
	entries, readErr := os.ReadDir(dest)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("destination entries/error = %v / %v, want no partial export", entries, readErr)
	}
}
