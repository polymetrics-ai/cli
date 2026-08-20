package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func binaryBundle(srv *httptest.Server, spec *BinaryOperationSpec) Bundle {
	if spec.Method == "" {
		spec.Method = http.MethodGet
	}
	if spec.Path == "" {
		spec.Path = "/files/{id}"
	}
	if spec.MaxBytes == 0 {
		spec.MaxBytes = 1 << 20
	}
	if len(spec.ContentTypes) == 0 {
		spec.ContentTypes = []string{"application/*"}
	}
	if spec.Response == nil {
		spec.Response = &OperationResponseSpec{SuccessStatuses: []string{"200-299"}}
	} else if len(spec.Response.SuccessStatuses) == 0 {
		spec.Response.SuccessStatuses = []string{"200-299"}
	}
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:           "acme.download_file",
			Kind:         "binary_download",
			Summary:      "download a file",
			Risk:         "low",
			Approval:     "none",
			OutputPolicy: "binary_file_bounded",
			Binary:       spec,
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodGet,
			Path:      spec.Path,
			Operation: &SurfaceOperation{},
		}}},
	}
}

func binaryServer(t *testing.T, body []byte, headers map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if headers == nil || headers["Content-Type"] == "" {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func downloadReq(dest string) BinaryDownloadRequest {
	return BinaryDownloadRequest{
		Operation:  "acme.download_file",
		PathParams: map[string]string{"id": "f1"},
		DestRoot:   dest,
	}
}

// TestBinaryDownloadWritesBoundedFile: happy path — bytes land on disk with
// the right size and digest, and the record points at them.
func TestBinaryDownloadWritesBoundedFile(t *testing.T) {
	payload := []byte("%PDF-1.4 hello world")
	srv := binaryServer(t, payload, map[string]string{"Content-Type": "application/pdf"})
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})

	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("OperationBinaryDownload: %v", err)
	}
	rec := res.Record
	path, _ := rec["file_path"].(string)
	if path == "" {
		t.Fatalf("record has no file_path: %v", rec)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("file contents = %q", got)
	}
	if rec["file_size_bytes"] != int64(len(payload)) {
		t.Fatalf("file_size_bytes = %v, want %d", rec["file_size_bytes"], len(payload))
	}
	sum := sha256.Sum256(payload)
	if rec["file_sha256"] != hex.EncodeToString(sum[:]) {
		t.Fatalf("file_sha256 = %v", rec["file_sha256"])
	}
	if rec["truncated"] != false {
		t.Fatalf("truncated = %v, want false", rec["truncated"])
	}
	if !strings.HasPrefix(path, dest) {
		t.Fatalf("file escaped the destination root: %s", path)
	}
}

// TestBinaryDownloadRejectsOverflow: a body one byte past the limit must be
// REJECTED, not silently truncated. A truncated PDF looks like a valid write.
func TestBinaryDownloadRejectsOverflow(t *testing.T) {
	payload := []byte(strings.Repeat("A", 101))
	srv := binaryServer(t, payload, nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{MaxBytes: 100})

	_, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err == nil {
		t.Fatal("want overflow rejection")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("error should name the size limit, got: %v", err)
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 0 {
		t.Fatalf("rejected download must leave no file behind, found %d", len(entries))
	}
}

func TestBinaryDownloadRequiresDeclaredMediaAndSuccessfulStatusBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	b := binaryBundle(srv, &BinaryOperationSpec{})
	b.Operations[0].Binary.ContentTypes = nil
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil || !strings.Contains(err.Error(), "content_types") {
		t.Fatalf("missing media contract error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("missing media contract reached provider %d times", requests)
	}
	b = binaryBundle(srv, &BinaryOperationSpec{})
	b.Operations[0].Binary.Response.SuccessStatuses = nil
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil || !strings.Contains(err.Error(), "success_statuses") {
		t.Fatalf("missing status contract error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("missing status contract reached provider %d times", requests)
	}
}

func TestBinaryDownloadRejectsUndeclaredSuccessfulStatusBeforeArtifact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("partial"))
	}))
	t.Cleanup(srv.Close)
	b := binaryBundle(srv, &BinaryOperationSpec{ContentTypes: []string{"application/octet-stream"}, Response: &OperationResponseSpec{SuccessStatuses: []string{"200"}}})
	dest := t.TempDir()
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil); err == nil || !strings.Contains(err.Error(), "not declared") {
		t.Fatalf("undeclared status error = %v", err)
	}
	entries, err := os.ReadDir(dest)
	if err != nil || len(entries) != 0 {
		t.Fatalf("undeclared status left artifacts: %v / %v", entries, err)
	}
}

// TestBinaryDownloadExactLimitSucceeds: exactly at the limit is fine — the
// read-one-past technique must not create an off-by-one rejection.
func TestBinaryDownloadExactLimitSucceeds(t *testing.T) {
	payload := []byte(strings.Repeat("A", 100))
	srv := binaryServer(t, payload, nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{MaxBytes: 100})

	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil); err != nil {
		t.Fatalf("exactly-at-limit must succeed: %v", err)
	}
}

// TestBinaryDownloadClampsMaxBytes: request clamps below spec, and spec always
// wins over a larger request.
func TestBinaryDownloadClampsMaxBytes(t *testing.T) {
	payload := []byte(strings.Repeat("A", 500))
	srv := binaryServer(t, payload, nil)
	b := binaryBundle(srv, &BinaryOperationSpec{MaxBytes: 1000})

	t.Run("request clamps below spec", func(t *testing.T) {
		req := downloadReq(t.TempDir())
		req.MaxBytes = 100
		if _, err := OperationBinaryDownload(context.Background(), b, req, nil); err == nil {
			t.Fatal("smaller request limit must apply")
		}
	})
	t.Run("request cannot exceed spec", func(t *testing.T) {
		small := binaryBundle(srv, &BinaryOperationSpec{MaxBytes: 100})
		req := downloadReq(t.TempDir())
		req.MaxBytes = 1 << 30
		if _, err := OperationBinaryDownload(context.Background(), small, req, nil); err == nil {
			t.Fatal("request must not be able to raise the spec limit")
		}
	})
}

// TestBinaryDownloadRejectsExtractArchives: zip-slip and decompression bombs
// are a separate capability, never a flag. Enforced at EXECUTION time because
// two github operations already declare it true and this phase modifies no
// connector bundle.
func TestBinaryDownloadRejectsExtractArchives(t *testing.T) {
	srv := binaryServer(t, []byte("PK\x03\x04"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{ExtractArchives: true})
	_, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil)
	if err == nil {
		t.Fatal("extract_archives must be refused")
	}
	if !strings.Contains(err.Error(), "extract_archives") {
		t.Fatalf("error should name extract_archives, got: %v", err)
	}
}

// TestBinaryDownloadRejectsWrongKind: a rest_read operation cannot be run
// through the binary executor.
func TestBinaryDownloadRejectsWrongKind(t *testing.T) {
	srv := binaryServer(t, []byte("x"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{})
	b.Operations[0].Kind = "rest_read"
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil {
		t.Fatal("non-binary_download kind must be refused")
	}
}

// TestBinaryDownloadRejectsNonGET mirrors the declared GET-only semantics.
func TestBinaryDownloadRejectsNonGET(t *testing.T) {
	srv := binaryServer(t, []byte("x"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{Method: http.MethodPost})
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil {
		t.Fatal("non-GET must be refused")
	}
}

// TestBinaryDownloadRejectsAbsoluteEndpoint keeps the connector-relative
// invariant the JSON direct-read path also enforces.
func TestBinaryDownloadRejectsAbsoluteEndpoint(t *testing.T) {
	srv := binaryServer(t, []byte("x"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{Path: "https://evil.example/blob"})
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil {
		t.Fatal("absolute endpoint must be refused")
	}
}

// TestBinaryDownloadRequiresDeclaredEndpoint: an endpoint absent from
// api_surface is not downloadable.
func TestBinaryDownloadRequiresDeclaredEndpoint(t *testing.T) {
	srv := binaryServer(t, []byte("x"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{})
	b.Surface.Endpoints = nil
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err == nil {
		t.Fatal("undeclared endpoint must be refused")
	}
}

// TestBinaryDownloadFilenameSanitized: provider-supplied names are never
// trusted. RFC 6266 counts BOTH / and \ as separators, and filepath.Base on
// Linux happily returns `..\..\etc\passwd` unchanged.
func TestBinaryDownloadFilenameSanitized(t *testing.T) {
	for _, disposition := range []string{
		`attachment; filename="../../etc/passwd"`,
		`attachment; filename="..\\..\\etc\\passwd"`,
		`attachment; filename="/etc/passwd"`,
		`attachment; filename=".."`,
		`attachment; filename=""`,
	} {
		t.Run(disposition, func(t *testing.T) {
			srv := binaryServer(t, []byte("data"), map[string]string{"Content-Disposition": disposition})
			dest := t.TempDir()
			b := binaryBundle(srv, &BinaryOperationSpec{})
			res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
			if err != nil {
				t.Fatalf("sanitization must fall back, not fail: %v", err)
			}
			path, _ := res.Record["file_path"].(string)
			resolvedDest, _ := filepath.EvalSymlinks(dest)
			resolvedPath, _ := filepath.EvalSymlinks(path)
			if !strings.HasPrefix(resolvedPath, resolvedDest) {
				t.Fatalf("file escaped root: %s not under %s", resolvedPath, resolvedDest)
			}
			name, _ := res.Record["file_name"].(string)
			if strings.Contains(name, "/") || strings.Contains(name, `\`) || name == ".." || name == "" {
				t.Fatalf("unsanitized file_name %q", name)
			}
		})
	}
}

// TestBinaryDownloadRFC5987Filename: mime.ParseMediaType decodes filename*
// into the UNSTARRED key, so reading params["filename*"] would silently miss.
func TestBinaryDownloadRFC5987Filename(t *testing.T) {
	srv := binaryServer(t, []byte("data"), map[string]string{
		"Content-Disposition": `attachment; filename*=UTF-8''r%C3%A9sum%C3%A9.pdf`,
	})
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if name, _ := res.Record["file_name"].(string); name != "résumé.pdf" {
		t.Fatalf("file_name = %q, want the decoded RFC 5987 value", name)
	}
}

// TestBinaryDownloadCallerFileNameContained: a caller-supplied name cannot
// traverse out of the root either.
func TestBinaryDownloadCallerFileNameContained(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	for _, name := range []string{"../escape.txt", "/abs.txt", "a/../../escape.txt"} {
		dest := t.TempDir()
		b := binaryBundle(srv, &BinaryOperationSpec{})
		req := downloadReq(dest)
		req.FileName = name
		if _, err := OperationBinaryDownload(context.Background(), b, req, nil); err == nil {
			t.Fatalf("caller file name %q must be refused", name)
		}
	}
}

// TestBinaryDownloadRejectsSymlinkEscape is the os.Root assertion: a SINGLE
// -SEGMENT name inside the root that happens to be a symlink pointing outside
// must not be followed. Name validation alone cannot catch this — the name is
// perfectly local — which is exactly why containment uses os.Root rather than
// safety.ValidateLocalWritePath's purely lexical check.
func TestBinaryDownloadRejectsSymlinkEscape(t *testing.T) {
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.bin")
	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed victim: %v", err)
	}
	dest := t.TempDir()
	// "f.bin" is a valid single-segment name, but inside the root it is a
	// symlink escaping to another directory.
	if err := os.Symlink(victim, filepath.Join(dest, "f.bin")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	srv := binaryServer(t, []byte("attacker"), nil)
	// allow_overwrite so the refusal can only come from symlink containment,
	// not from the exclusive-create collision check.
	b := binaryBundle(srv, &BinaryOperationSpec{AllowOverwrite: true})
	req := downloadReq(dest)
	req.FileName = "f.bin"

	_, err := OperationBinaryDownload(context.Background(), b, req, nil)
	got, readErr := os.ReadFile(victim)
	if readErr != nil {
		t.Fatalf("read victim: %v", readErr)
	}
	if string(got) != "original" {
		t.Fatalf("symlink escape wrote through the link to %s: %q", victim, got)
	}
	if err == nil {
		// Writing a real file at dest/f.bin (replacing the link) is an
		// acceptable outcome; writing THROUGH the link is not, and is already
		// caught above.
		info, lerr := os.Lstat(filepath.Join(dest, "f.bin"))
		if lerr != nil {
			t.Fatalf("lstat: %v", lerr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatal("download reported success but destination is still the escaping symlink")
		}
	}
}

// TestBinaryDownloadCallerFileNameNotRewritten: a traversing caller name must
// be REFUSED, not silently basename-ed into something safe, or a traversal
// attempt is hidden rather than reported.
func TestBinaryDownloadCallerFileNameNotRewritten(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	req := downloadReq(dest)
	req.FileName = "../escape.txt"
	if _, err := OperationBinaryDownload(context.Background(), b, req, nil); err == nil {
		t.Fatal("traversing caller name must be refused")
	}
	if _, err := os.Stat(filepath.Join(dest, "escape.txt")); err == nil {
		t.Fatal("traversing caller name was silently rewritten to a safe name")
	}
}

// TestBinaryDownloadHonoursAllowOverwrite.
func TestBinaryDownloadHonoursAllowOverwrite(t *testing.T) {
	srv := binaryServer(t, []byte("second"), map[string]string{
		"Content-Disposition": `attachment; filename="f.bin"`,
	})
	t.Run("refuses by default", func(t *testing.T) {
		dest := t.TempDir()
		if err := os.WriteFile(filepath.Join(dest, "f.bin"), []byte("first"), 0o600); err != nil {
			t.Fatalf("seed: %v", err)
		}
		b := binaryBundle(srv, &BinaryOperationSpec{})
		if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil); err == nil {
			t.Fatal("existing file must not be overwritten by default")
		}
		got, _ := os.ReadFile(filepath.Join(dest, "f.bin"))
		if string(got) != "first" {
			t.Fatalf("original file was clobbered: %q", got)
		}
	})
	t.Run("allows when declared", func(t *testing.T) {
		dest := t.TempDir()
		if err := os.WriteFile(filepath.Join(dest, "f.bin"), []byte("first"), 0o600); err != nil {
			t.Fatalf("seed: %v", err)
		}
		b := binaryBundle(srv, &BinaryOperationSpec{AllowOverwrite: true})
		if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil); err != nil {
			t.Fatalf("allow_overwrite must permit replacement: %v", err)
		}
		got, _ := os.ReadFile(filepath.Join(dest, "f.bin"))
		if string(got) != "second" {
			t.Fatalf("file not replaced: %q", got)
		}
	})
}

// TestBinaryDownloadFilePermissions: downloaded content is often invoices and
// identity documents.
func TestBinaryDownloadFilePermissions(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	info, err := os.Stat(res.Record["file_path"].(string))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file mode = %o, want 600", perm)
	}
}

// TestBinaryDownloadSniffsContentType: never trust Content-Type and never
// infer from the path. Marketo serves CSV bytes from a path ending .json.
// Record both, surface the mismatch, do not reject on it.
func TestBinaryDownloadSniffsContentType(t *testing.T) {
	srv := binaryServer(t, []byte("id,name\n1,alice\n"), map[string]string{
		"Content-Type": "application/json",
	})
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("a content-type mismatch must NOT fail the download: %v", err)
	}
	if got := res.Record["content_type"]; got != "application/json" {
		t.Fatalf("content_type = %v, want the provider's claim verbatim", got)
	}
	sniffed, _ := res.Record["content_type_sniffed"].(string)
	if sniffed == "" || strings.HasPrefix(sniffed, "application/json") {
		t.Fatalf("content_type_sniffed = %q, want an independently sniffed value", sniffed)
	}
}

// TestBinaryDownloadRecordIsFlatAndSurvivesRedaction: records are flat
// map[string]any and pass through schema projection, so a nested object would
// not survive. And shouldRedactJSONField silently redacts anything containing
// both "download" and "url" — a field named download_url would become
// download_url_redacted:true, so the reference field must be source_ref.
func TestBinaryDownloadRecordIsFlatAndSurvivesRedaction(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	for k, v := range res.Record {
		switch v.(type) {
		case map[string]any, []any, connectors.Record:
			t.Fatalf("record field %q is not a flat scalar: %T", k, v)
		}
		if shouldRedactJSONField(k) {
			t.Fatalf("record field %q would be silently redacted by shouldRedactJSONField", k)
		}
	}
	for _, want := range []string{
		"file_path", "file_name", "file_size_bytes", "file_sha256",
		"content_type", "content_type_sniffed", "source_operation",
		"source_ref", "downloaded_at", "truncated",
	} {
		if _, ok := res.Record[want]; !ok {
			t.Fatalf("record missing %q: %v", want, res.Record)
		}
	}
	if _, bad := res.Record["download_url"]; bad {
		t.Fatal("record must not use download_url; it is auto-redacted")
	}
}

// TestBinaryDownloadRequiresDestRoot: there is no implicit destination.
func TestBinaryDownloadRequiresDestRoot(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{})
	req := downloadReq("")
	if _, err := OperationBinaryDownload(context.Background(), b, req, nil); err == nil {
		t.Fatal("missing dest root must be refused")
	}
}

// TestBinaryDownloadPreservesHTTPErrorTextAndLeavesNoFile.
func TestBinaryDownloadPreservesHTTPErrorTextAndLeavesNoFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("trace"), "binary-download-fixture"; got != want {
			t.Fatalf("trace = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"diagnostic":"binary-download-fixture-body"}`))
	}))
	t.Cleanup(srv.Close)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	req := downloadReq(dest)
	req.Query = map[string]string{"trace": "binary-download-fixture"}
	_, err := OperationBinaryDownload(context.Background(), b, req, nil)
	if err == nil {
		t.Fatal("want error for 403")
	}
	for _, want := range []string{"trace=binary-download-fixture", `{"diagnostic":"binary-download-fixture-body"}`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("OperationBinaryDownload error = %q, want complete diagnostic %q", err.Error(), want)
		}
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 0 {
		t.Fatalf("failed download must leave no file, found %d", len(entries))
	}
}

// TestOperationsSchemaAcceptsBinaryPolicyFields COMPILES and RUNS the real
// operations meta-schema against a document using the new binary fields. The
// binary block is additionalProperties:false, so a field that is not properly
// declared makes this fail — unlike a substring check, which would pass on a
// field name appearing anywhere in the file.
func TestOperationsSchemaAcceptsBinaryPolicyFields(t *testing.T) {
	sch, err := CompileSchema([]byte(operationsSchemaJSON))
	if err != nil {
		t.Fatalf("compile operations schema: %v", err)
	}
	doc := map[string]any{"operations": []any{map[string]any{
		"id": "acme.download_file", "kind": "binary_download",
		"summary": "download", "risk": "low", "approval": "none",
		"output_policy": "binary_file_bounded",
		"binary": map[string]any{
			"method": "GET", "path": "/files/{id}", "max_bytes": 1024,
			"accept":           "application/pdf",
			"allow_cross_host": true, "allowed_hosts": []any{"cdn.example"},
			"content_types": []any{"application/pdf"}, "stall_timeout_seconds": 30,
		},
	}}}
	if err := sch.Validate(doc); err != nil {
		t.Fatalf("operations.schema.json must accept the binary policy fields: %v", err)
	}
	// And the closed block must still reject an undeclared field.
	bad := map[string]any{"operations": []any{map[string]any{
		"id": "acme.download_file", "kind": "binary_download",
		"summary": "download", "risk": "low", "approval": "none",
		"output_policy": "binary_file_bounded",
		"binary": map[string]any{
			"method": "GET", "path": "/files/{id}", "unknown_field": true,
		},
	}}}
	if err := sch.Validate(bad); err == nil {
		t.Fatal("binary block must stay additionalProperties:false")
	}
}

// redirectingBinaryServers wires an origin that redirects to a separate CDN
// host, recording what the CDN receives.
func redirectingBinaryServers(t *testing.T) (origin *httptest.Server, cdnHost *string, cdnSawAuth *string) {
	t.Helper()
	sawAuth := ""
	host := ""
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-API-Key")
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("cdn-bytes"))
	}))
	t.Cleanup(cdn.Close)
	host = strings.TrimPrefix(cdn.URL, "http://")
	origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/blob", http.StatusFound)
	}))
	t.Cleanup(origin.Close)
	return origin, &host, &sawAuth
}

func binaryBundleWithAuth(srv *httptest.Server, spec *BinaryOperationSpec) Bundle {
	b := binaryBundle(srv, spec)
	b.HTTP.Auth = []AuthSpec{{Mode: "api_key_header", Header: "X-API-Key", Value: "supersecret"}}
	return b
}

// TestBinaryDownloadRefusesCrossHostRedirectByDefault EXECUTES the capability
// to prove the declaration is enforced, not merely accepted by the schema.
// This is the 71-connector custom-auth-header hazard: without the policy, Go
// forwards X-API-Key to the CDN.
func TestBinaryDownloadRefusesCrossHostRedirectByDefault(t *testing.T) {
	origin, _, cdnSawAuth := redirectingBinaryServers(t)
	b := binaryBundleWithAuth(origin, &BinaryOperationSpec{})
	_, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil)
	if err == nil {
		t.Fatal("cross-host redirect must be refused when the operation does not declare it")
	}
	if *cdnSawAuth != "" {
		t.Fatalf("credential reached the CDN: %q", *cdnSawAuth)
	}
}

// TestBinaryDownloadAllowCrossHostIsEnforced: declaring allow_cross_host
// actually changes behaviour, and the credential still does not travel.
func TestBinaryDownloadAllowCrossHostIsEnforced(t *testing.T) {
	origin, cdnHost, cdnSawAuth := redirectingBinaryServers(t)
	dest := t.TempDir()
	b := binaryBundleWithAuth(origin, &BinaryOperationSpec{Redirect: &OperationRedirectSpec{MaxHops: 1, AllowedHosts: []string{*cdnHost}}})
	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil)
	if err != nil {
		t.Fatalf("declared allow_cross_host must permit the hop: %v", err)
	}
	got, err := os.ReadFile(res.Record["file_path"].(string))
	if err != nil || string(got) != "cdn-bytes" {
		t.Fatalf("cdn bytes not written: %q %v", got, err)
	}
	if *cdnSawAuth != "" {
		t.Fatalf("credential leaked to the CDN even on a permitted hop: %q", *cdnSawAuth)
	}
}

// TestBinaryDownloadAllowedHostsIsEnforced: the per-operation allowlist admits
// exactly the named host and nothing else.
func TestBinaryDownloadAllowedHostsIsEnforced(t *testing.T) {
	origin, cdnHost, cdnSawAuth := redirectingBinaryServers(t)

	b := binaryBundleWithAuth(origin, &BinaryOperationSpec{Redirect: &OperationRedirectSpec{MaxHops: 1, AllowedHosts: []string{*cdnHost}}})
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil); err != nil {
		t.Fatalf("allowlisted host must be permitted: %v", err)
	}
	if *cdnSawAuth != "" {
		t.Fatalf("credential leaked to an allowlisted host: %q", *cdnSawAuth)
	}

	other := binaryBundleWithAuth(origin, &BinaryOperationSpec{Redirect: &OperationRedirectSpec{MaxHops: 1, AllowedHosts: []string{"somewhere.invalid:443"}}})
	if _, err := OperationBinaryDownload(context.Background(), other, downloadReq(t.TempDir()), nil); err == nil {
		t.Fatal("a host outside allowed_hosts must stay refused")
	}
}

// TestBinaryDownloadStallTimeoutIsEnforced: stall_timeout_seconds must be
// plumbed through and actually abort a stalled transfer, rather than being an
// inert schema field.
func TestBinaryDownloadStallTimeoutIsEnforced(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		_, _ = w.Write([]byte("partial"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		<-release // stall forever until the test releases us
	}))
	t.Cleanup(func() { close(release); srv.Close() })

	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{StallTimeoutSeconds: 1})
	start := time.Now()
	if _, err := OperationBinaryDownload(context.Background(), b, downloadReq(dest), nil); err == nil {
		t.Fatal("a stalled transfer must be aborted")
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Fatalf("stall timeout did not fire promptly: %v", elapsed)
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 0 {
		t.Fatalf("aborted download must leave no file, found %d", len(entries))
	}
}

// TestBinaryDownloadAppliesRedactFields: a command's declared redact_fields
// reaches the record. source_ref carries the resolved endpoint path, including
// whatever path parameters the caller supplied, so a connector that declares it
// sensitive must actually get it redacted rather than merely documented.
func TestBinaryDownloadAppliesRedactFields(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	dest := t.TempDir()
	b := binaryBundle(srv, &BinaryOperationSpec{})
	req := downloadReq(dest)
	req.RedactFields = []string{"source_ref"}

	res, err := OperationBinaryDownload(context.Background(), b, req, nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if _, present := res.Record["source_ref"]; present {
		t.Fatalf("source_ref survived redaction: %v", res.Record)
	}
	if res.Record["source_ref_redacted"] != true {
		t.Fatalf("record does not report the redaction: %v", res.Record)
	}
	// Redaction must not disturb the rest of the record, and must not nest it.
	if res.Record["file_name"] == nil {
		t.Fatalf("redaction dropped unrelated fields: %v", res.Record)
	}
	for k, v := range res.Record {
		switch v.(type) {
		case map[string]any, []any, connectors.Record:
			t.Fatalf("redaction made field %q non-flat: %T", k, v)
		}
	}
}

// TestBinaryDownloadWithoutRedactFieldsKeepsRecord: the no-redaction path must
// stay byte-identical, so an empty declaration costs nothing.
func TestBinaryDownloadWithoutRedactFieldsKeepsRecord(t *testing.T) {
	srv := binaryServer(t, []byte("data"), nil)
	b := binaryBundle(srv, &BinaryOperationSpec{})

	res, err := OperationBinaryDownload(context.Background(), b, downloadReq(t.TempDir()), nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if res.Record["source_ref"] != "/files/f1" {
		t.Fatalf("source_ref = %v, want the resolved path", res.Record["source_ref"])
	}
}
