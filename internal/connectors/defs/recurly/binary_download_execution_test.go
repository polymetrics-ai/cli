// Package recurly holds connector-local execution evidence for the Recurly
// declarative bundle.
package recurly

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

// TestBinaryDownloadCommandsExecute proves each promoted binary command through
// the same path an operator invokes: commandrunner resolves the command and
// flags, then the declarative engine performs a bounded GET and writes only
// below the caller-selected destination. The fixture server is deliberately
// local; no Recurly credential or live provider request is used.
func TestBinaryDownloadCommandsExecute(t *testing.T) {
	t.Parallel()

	type response struct {
		contentType string
		body        []byte
	}
	responses := map[string]response{
		"/invoices/invoice-fixture.pdf": {
			contentType: "application/pdf",
			body:        []byte("%PDF-1.7\\nfixture invoice\\n"),
		},
		"/export_dates": {
			contentType: "application/json",
			body:        []byte("[\"2026-01-02\"]\\n"),
		},
		"/export_dates/2026-01-02/export_files": {
			contentType: "application/json",
			body:        []byte("[{\"href\":\"/exports/fixture.csv\"}]\\n"),
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		fixture, ok := responses[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", fixture.contentType)
		_, _ = w.Write(fixture.body)
	}))
	defer server.Close()

	bundle, err := engine.Load(os.DirFS(".."), "recurly")
	if err != nil {
		t.Fatalf("load Recurly bundle: %v", err)
	}
	connector := engine.New(bundle, nil)
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL},
		Secrets: map[string]string{"api_key": "synthetic-test-api-key"},
	}

	tests := []struct {
		name      string
		path      []string
		flags     map[string][]string
		operation string
		fileName  string
		body      []byte
	}{
		{
			name:      "invoice PDF",
			path:      []string{"invoice", "pdf", "get"},
			flags:     map[string][]string{"invoice-id": {"invoice-fixture"}},
			operation: "get_invoice_pdf",
			fileName:  "invoice.pdf",
			body:      responses["/invoices/invoice-fixture.pdf"].body,
		},
		{
			name:      "export dates",
			path:      []string{"export", "dates", "get"},
			operation: "get_export_dates",
			fileName:  "export-dates.json",
			body:      responses["/export_dates"].body,
		},
		{
			name:      "export files",
			path:      []string{"export", "files", "get"},
			flags:     map[string][]string{"export-date": {"2026-01-02"}},
			operation: "get_export_files",
			fileName:  "export-files.json",
			body:      responses["/export_dates/2026-01-02/export_files"].body,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := commandrunner.Preflight(connector, test.path); err != nil {
				t.Fatalf("Preflight(%q): %v", strings.Join(test.path, " "), err)
			}

			destination := t.TempDir()
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:     test.path,
				Flags:    test.flags,
				Config:   config,
				DestRoot: destination,
				FileName: test.fileName,
			}, func(connectors.Record) error {
				return io.ErrUnexpectedEOF
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(test.path, " "), err)
			}
			if result.BinaryDownload == nil {
				t.Fatalf("Run(%q) did not return a binary download", strings.Join(test.path, " "))
			}
			if result.BinaryDownload.Operation != test.operation {
				t.Fatalf("operation = %q, want %q", result.BinaryDownload.Operation, test.operation)
			}
			got, err := os.ReadFile(filepath.Join(destination, test.fileName))
			if err != nil {
				t.Fatalf("read downloaded fixture: %v", err)
			}
			if !bytes.Equal(got, test.body) {
				t.Fatalf("downloaded bytes = %q, want %q", got, test.body)
			}
		})
	}
}
