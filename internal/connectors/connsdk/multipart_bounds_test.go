package connsdk

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// uploadEcho starts a server that captures the bytes of the "mediaFile" part.
//
// A refused upload aborts the request mid-stream, so the handler legitimately
// sees a half-written body and a closed connection. That is the behaviour under
// test, not a server error: the assertions below are on what did or did not
// reach *got, so an incomplete request must leave it empty and return quietly
// rather than failing the test. Treating the abort as an error made this a flake
// — whether the handler got as far as ParseMultipartForm depended on timing.
func uploadEcho(t *testing.T, got *string, completed ...chan<- struct{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if len(completed) == 1 && completed[0] != nil {
				close(completed[0])
			}
		}()
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return
		}
		parts := r.MultipartForm.File["mediaFile"]
		if len(parts) == 0 {
			return
		}
		f, err := parts[0].Open()
		if err != nil {
			return
		}
		defer f.Close()
		raw, err := io.ReadAll(f)
		if err != nil {
			return
		}
		*got = string(raw)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

// TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation pins the
// TOCTOU fix. Before os.Root confinement this uploaded the outside file: the
// path was validated once, then re-opened by path at send time, so swapping it
// for an escaping symlink in between put content from outside the root on the
// wire. Auth.Apply is the injection point because it runs per attempt, after
// validateMultipartForm and before the body is streamed — the same seam
// TestRequesterDoMultipartRejectsGrowthAfterPreflightValidation uses.
func TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation(t *testing.T) {
	rootDir := t.TempDir()
	outsideDir := t.TempDir()

	secret := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(secret, []byte("OUTSIDE-ROOT-SECRET"), 0o600); err != nil {
		t.Fatal(err)
	}
	innocent := filepath.Join(rootDir, "payload.txt")
	if err := os.WriteFile(innocent, []byte("innocent"), 0o600); err != nil {
		t.Fatal(err)
	}

	var sawFile string
	handlerDone := make(chan struct{})
	srv := uploadEcho(t, &sawFile, handlerDone)
	defer srv.Close()

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	r := &Requester{
		BaseURL:        srv.URL,
		Sleep:          noSleep,
		DisableRetries: true,
		Auth: AuthFunc(func(context.Context, *http.Request) error {
			if err := os.Remove(innocent); err != nil {
				return err
			}
			return os.Symlink(secret, innocent)
		}),
	}
	_, err = r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName: "mediaFile",
			Root:      root,
			RelPath:   "payload.txt",
			// Path is deliberately the absolute path the pre-fix engine handed
			// down. If any access site falls back to opening it, the swapped
			// symlink is followed and this test fails — which is what makes it a
			// real pin rather than one that passes because a relative path
			// happens not to resolve.
			Path:     innocent,
			MaxBytes: 1024,
		}},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want refusal of the escaping symlink")
	}
	select {
	case <-handlerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("upload handler did not complete before assertion")
	}
	if sawFile == "OUTSIDE-ROOT-SECRET" {
		t.Fatalf("content from outside the root reached the wire: %q", sawFile)
	}
}

// TestRequesterDoMultipartRefusesRootRelativeTraversal covers the plain
// traversal case at every access site rather than only at a pre-check.
func TestRequesterDoMultipartRefusesRootRelativeTraversal(t *testing.T) {
	rootDir := t.TempDir()
	outside := filepath.Join(filepath.Dir(rootDir), "escaped.txt")
	if err := os.WriteFile(outside, []byte("ESCAPED"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)

	// Counting requests pins the pre-read stat site independently: containment
	// there means the traversal is refused before anything is sent at all, not
	// merely before the body is written.
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err = r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName: "mediaFile",
			Root:      root,
			RelPath:   filepath.Join("..", filepath.Base(outside)),
			// Absolute, for the same reason as above: an unconfined fallback
			// would succeed here, so the test fails if containment regresses.
			Path:     outside,
			MaxBytes: 1024,
		}},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want traversal refusal")
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("HTTP calls = %d, want the traversal refused before any request", got)
	}
}

func TestRequesterDoMultipartRejectsScalarAndFramingOverAggregateCapBeforeRequest(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls.Add(1)
	}))
	t.Cleanup(srv.Close)
	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Fields:   map[string]string{"message": strings.Repeat("x", 128)},
		MaxBytes: 32,
	})
	if err == nil || !strings.Contains(err.Error(), "multipart payload too large") {
		t.Fatalf("DoMultipart error = %v, want aggregate-cap refusal", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("requests = %d, want scalar/framing cap refusal before I/O", calls.Load())
	}
}

// pngBytes is a minimal PNG header; http.DetectContentType sniffs on the magic.
var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 32)...)

// zipBytes is a minimal ZIP local-file-header magic.
var zipBytes = append([]byte("PK\x03\x04"), bytes.Repeat([]byte{0}, 32)...)

func TestRequesterDoMultipartRejectsDisallowedMediaType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-an-image.png")
	if err := os.WriteFile(path, zipBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	var sawFile string
	srv := uploadEcho(t, &sawFile)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:         "mediaFile",
			Path:              path,
			ContentType:       "image/png",
			AllowedMediaTypes: []string{"image/png", "image/jpeg"},
			MaxBytes:          1 << 20,
		}},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want media-type refusal")
	}
	// The message must name what was sniffed so a connector author can widen the
	// declaration deliberately rather than by guesswork.
	if !bytes.Contains([]byte(err.Error()), []byte("application/zip")) {
		t.Fatalf("error = %q, want it to name the sniffed type", err.Error())
	}
	if sawFile != "" {
		t.Fatalf("bytes reached the wire before the type was checked: %q", sawFile)
	}
}

func TestRequesterDoMultipartAcceptsAllowedMediaType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, pngBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	var sawFile string
	srv := uploadEcho(t, &sawFile)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:         "mediaFile",
			Path:              path,
			ContentType:       "image/png",
			AllowedMediaTypes: []string{"image/png"},
			MaxBytes:          1 << 20,
		}},
	})
	if err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if sawFile != string(pngBytes) {
		t.Fatalf("uploaded bytes = %d, want the png", len(sawFile))
	}
}

// TestRequesterDoMultipartRejectsUnclassifiableContent pins the distinct message
// for content http.DetectContentType cannot classify, so "unknown" is not
// silently reported as an ordinary type mismatch.
func TestRequesterDoMultipartRejectsUnclassifiableContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02, 0xff, 0xfe}, 0o600); err != nil {
		t.Fatal(err)
	}
	srv := uploadEcho(t, new(string))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:         "mediaFile",
			Path:              path,
			AllowedMediaTypes: []string{"image/png"},
			MaxBytes:          1 << 20,
		}},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want refusal")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("could not be classified")) {
		t.Fatalf("error = %q, want the distinct unclassifiable message", err.Error())
	}
}

// TestRequesterDoMultipartUnconstrainedMediaTypeStillUploads keeps "no declared
// bound" a valid, explicit choice rather than an accident of the new field.
func TestRequesterDoMultipartUnconstrainedMediaTypeStillUploads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "anything.bin")
	if err := os.WriteFile(path, zipBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	var sawFile string
	srv := uploadEcho(t, &sawFile)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{FieldName: "mediaFile", Path: path, MaxBytes: 1 << 20}},
	}); err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if sawFile != string(zipBytes) {
		t.Fatalf("uploaded %d bytes, want the file unchanged", len(sawFile))
	}
}

// uploadHeaderEcho captures the Content-Type header of the "mediaFile" part,
// which is the claim we actually make to the provider. It is deliberately
// separate from uploadEcho: these tests assert on the header rather than the
// bytes, and a refused upload must leave *got empty for the same mid-stream
// reason documented there.
func uploadHeaderEcho(t *testing.T, got *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return
		}
		parts := r.MultipartForm.File["mediaFile"]
		if len(parts) == 0 {
			return
		}
		*got = parts[0].Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

// jpegBytes is a minimal JFIF header, enough for http.DetectContentType.
var jpegBytes = append([]byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00"), bytes.Repeat([]byte{0}, 32)...)

// TestRequesterDoMultipartSendsSniffedContentTypeWhenBounded pins the decision
// that the part header describes the bytes we send rather than the type the
// bundle hoped for. The file here is a JPEG while content_type declares PNG, and
// both are allowed: before this, the request asserted "image/png" over JPEG
// bytes — a claim we had just proven false. The allowlist stays the restriction;
// declaring a single-entry list is how a bundle demands exactly one type, and
// TestRequesterDoMultipartRejectsDisallowedMediaType covers that path.
func TestRequesterDoMultipartSendsSniffedContentTypeWhenBounded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.png")
	if err := os.WriteFile(path, jpegBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	var sawType string
	srv := uploadHeaderEcho(t, &sawType)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:         "mediaFile",
			Path:              path,
			ContentType:       "image/png",
			AllowedMediaTypes: []string{"image/png", "image/jpeg"},
			MaxBytes:          1 << 20,
		}},
	}); err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if sawType != "image/jpeg" {
		t.Fatalf("part Content-Type = %q, want the sniffed image/jpeg rather than the declared image/png", sawType)
	}
}

// TestRequesterDoMultipartKeepsDeclaredContentTypeWhenUnbounded is the other
// half of the decision: with no allowlist the sniff was never checked against a
// declaration, and http.DetectContentType is coarse (every CSV sniffs as
// text/plain), so a deliberate content_type must survive untouched.
func TestRequesterDoMultipartKeepsDeclaredContentTypeWhenUnbounded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")
	if err := os.WriteFile(path, []byte("id,name\n1,ada\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var sawType string
	srv := uploadHeaderEcho(t, &sawType)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:   "mediaFile",
			Path:        path,
			ContentType: "text/csv",
			MaxBytes:    1 << 20,
		}},
	}); err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if sawType != "text/csv" {
		t.Fatalf("part Content-Type = %q, want the declared text/csv preserved", sawType)
	}
}

// TestRequesterDoMultipartRootConfinedHappyPathUploads proves confinement does
// not break the ordinary case.
func TestRequesterDoMultipartRootConfinedHappyPathUploads(t *testing.T) {
	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "nested", "payload.txt"), []byte("hello upload"), 0o600); err != nil {
		t.Fatal(err)
	}
	var sawFile string
	srv := uploadEcho(t, &sawFile)
	defer srv.Close()

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName: "mediaFile",
			Root:      root,
			RelPath:   filepath.Join("nested", "payload.txt"),
			Path:      filepath.Join("nested", "payload.txt"),
			MaxBytes:  1024,
		}},
	}); err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if sawFile != "hello upload" {
		t.Fatalf("uploaded = %q", sawFile)
	}
}
