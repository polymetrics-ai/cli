package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceRetainWritesVerifiedMachineReadableArtifactWithoutChangingLock(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	lockRaw := writeSourceRetainFixtureLock(t, defsRoot, lock)

	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "1 artifact(s) retained") || stderr.Len() != 0 {
		t.Fatalf("source-retain output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	after, err := os.ReadFile(filepath.Join(sourcesDir, "alpha-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read source lock after retain: %v", err)
	}
	if !bytes.Equal(after, lockRaw) {
		t.Fatal("source-retain changed the source lock")
	}
	fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, "alpha", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct retained reader: %v", err)
	}
	got, err := fetcher.FetchArtifact(context.Background(), lock.Rest.sourceImportArtifact)
	if err != nil {
		t.Fatalf("read retained machine-readable artifact: %v", err)
	}
	if !bytes.Equal(got, artifact) {
		t.Fatalf("retained machine-readable bytes = %q, want %q", got, artifact)
	}
}

func TestSourceRetainRetainsRenderedReferenceAndBundleArtifacts(t *testing.T) {
	t.Parallel()
	renderedLockRaw, renderedArtifact := sourceImportV3RenderedReferenceLock(t, renderedReferenceCitationURL)
	bundleLockRaw, bundleArtifact := sourceImportV3BundleLock(t)
	for _, tc := range []struct {
		name     string
		lockRaw  []byte
		artifact []byte
	}{
		{name: "rendered reference", lockRaw: renderedLockRaw, artifact: renderedArtifact},
		{name: "zip bundle", lockRaw: bundleLockRaw, artifact: bundleArtifact},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defsRoot := t.TempDir()
			lock, err := parseSourceImportLock(tc.lockRaw, "fixture")
			if err != nil {
				t.Fatalf("parse fixture lock: %v", err)
			}
			writeSourceRetainFixtureLockRaw(t, defsRoot, lock.Connector, tc.lockRaw)
			var stdout, stderr bytes.Buffer
			code := runSourceRetainWithFetcher([]string{"source-retain", "fixture", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
				return tc.artifact, nil
			}))
			if code != 0 {
				t.Fatalf("source-retain exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			retainedLock, err := parseSourceRetainLock(tc.lockRaw, lock.Connector)
			if err != nil {
				t.Fatalf("parse retain lock: %v", err)
			}
			if len(retainedLock.Artifacts) != 1 {
				t.Fatalf("retained artifacts = %d, want 1", len(retainedLock.Artifacts))
			}
			fetcher, err := newSourceImportRetainedArtifactFetcher(filepath.Join(defsRoot, "fixture", "sources"), "fixture", defaultSourceImportLimits())
			if err != nil {
				t.Fatalf("construct retained reader: %v", err)
			}
			got, err := fetcher.FetchArtifact(context.Background(), retainedLock.Artifacts[0])
			if err != nil {
				t.Fatalf("read retained %s artifact: %v", tc.name, err)
			}
			if !bytes.Equal(got, tc.artifact) {
				t.Fatalf("retained %s bytes differ", tc.name)
			}
		})
	}
}

func TestSourceRetainRefusesMismatchedProviderBytes(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	writeSourceRetainFixtureLock(t, defsRoot, lock)

	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return append(append([]byte(nil), artifact...), '\n'), nil
	}))
	if code != 1 || !strings.Contains(stderr.String(), "source-lock refresh required") {
		t.Fatalf("mismatched source-retain exit=%d stderr=%q, want locked-byte refusal", code, stderr.String())
	}
	artifactPath := filepath.Join(defsRoot, "alpha", "sources", "artifacts", strings.ToLower(lock.Rest.SHA256)+sourceImportRetainedArtifactExtension)
	if _, err := os.Stat(artifactPath); !os.IsNotExist(err) {
		t.Fatalf("mismatched bytes created retained artifact %s: %v", artifactPath, err)
	}
}

func TestSourceRetainRejectsIncompleteProvenanceBeforeAnyFetch(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	fetches := 0
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		fetches++
		return nil, nil
	}))
	if code != 2 || fetches != 0 || !strings.Contains(stderr.String(), "--terms must be non-empty provenance text") {
		t.Fatalf("incomplete provenance exit=%d fetches=%d stderr=%q", code, fetches, stderr.String())
	}
}

func TestSourceRetainHelpAndMigrationDocumentationDescribeIdentityAndWrongSource(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if code := runSourceRetain([]string{"source-retain", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("source-retain help exit = %d, stderr=%s", code, stderr.String())
	}
	for _, want := range []string{"canonical_json", "wrong source", "operation\nor parity"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("source-retain help missing %q: %s", want, stdout.String())
		}
	}
	docs, err := os.ReadFile(filepath.Join("..", "..", "docs", "migration", "conventions.md"))
	if err != nil {
		t.Fatalf("read migration conventions: %v", err)
	}
	for _, want := range []string{"canonical_json", "wrong source", "retained-byte-sha256", "does not silently re-pin"} {
		if !strings.Contains(string(docs), want) {
			t.Fatalf("migration conventions missing %q", want)
		}
	}
}

func TestSourceRetainRetainsV3ArtifactWithoutImportTimeFormInventory(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	digest := sha256.Sum256(artifact)
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 3,
		"connector":      "alpha",
		"rest": map[string]any{
			"retrieval": "fixture retain-only capture",
			"source_documents": []any{map[string]any{
				"id": "alpha",
				"artifact": map[string]any{
					"source_url": "https://fixtures.polymetrics.invalid/alpha-openapi.yaml",
					"sha256":     hex.EncodeToString(digest[:]),
					"bytes":      len(artifact),
				},
				"published_source": map[string]any{
					"source_url":  "https://docs.polymetrics.invalid/alpha",
					"capture_url": "https://fixtures.polymetrics.invalid/alpha-openapi.yaml",
					"sha256":      hex.EncodeToString(digest[:]),
					"bytes":       len(artifact),
					"adapter":     "fixture",
				},
				"operations": []any{map[string]any{
					"id":              "alpha.widgets.list",
					"protocol":        "rest",
					"method":          "GET",
					"path":            "/widgets",
					"source_location": "paths./widgets.get",
				}},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	})
	if err != nil {
		t.Fatalf("marshal retain-only fixture lock: %v", err)
	}
	writeSourceRetainFixtureLockRaw(t, defsRoot, "alpha", lockRaw)

	if _, err := parseSourceImportLock(lockRaw, "alpha"); err == nil {
		t.Fatal("fixture unexpectedly passes source-import validation without an OpenAPI form inventory")
	}
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want 0 for retain-only lock; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if got, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-source-lock.json")); err != nil || !bytes.Equal(got, lockRaw) {
		t.Fatalf("retain-only source lock changed or unreadable: %v", err)
	}
}

func TestSourceRetainHTTPFetchDoesNotRequireImportFormValidation(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	lock.Rest.OpenAPI = "3.2.0" // Source import rejects this unknown form; retention must not.
	writeSourceRetainFixtureLock(t, defsRoot, lock)
	fetcher := httpSourceImportFetcher{
		limits: defaultSourceImportLimits(),
		lookup: batchArtifactLookupIPAddr(func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
		}),
		client: &http.Client{Transport: sourceImportRoundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(artifact)), Request: request}, nil
		})},
	}
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, fetcher)
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want retain despite unknown import form; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestSourceRetainSupportsParitySourceLock(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	digest := sha256.Sum256(artifact)
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"connector":      "alpha",
		"rest": map[string]any{
			"source_url":       "https://fixtures.polymetrics.invalid/alpha-openapi.yaml",
			"sha256":           hex.EncodeToString(digest[:]),
			"bytes":            len(artifact),
			"operation_counts": map[string]any{"get": 1},
			"operations":       []any{},
		},
	})
	if err != nil {
		t.Fatalf("marshal parity source lock: %v", err)
	}
	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create parity source directory: %v", err)
	}
	lockPath := filepath.Join(sourcesDir, "alpha-parity-source-lock.json")
	if err := os.WriteFile(lockPath, lockRaw, 0o644); err != nil {
		t.Fatalf("write parity source lock: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want parity-lock success; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if got, err := os.ReadFile(lockPath); err != nil || !bytes.Equal(got, lockRaw) {
		t.Fatalf("parity source lock changed or unreadable: %v", err)
	}
}

func TestSourceRetainAllowsLockedParityQuery(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := []byte(`{"openapi":"3.0.3","info":{"title":"query fixture","version":"1"},"paths":{}}`)
	digest := sha256.Sum256(artifact)
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"connector":      "alpha",
		"rest": map[string]any{
			"source_url":       "https://fixtures.polymetrics.invalid/alpha.json?download",
			"sha256":           hex.EncodeToString(digest[:]),
			"bytes":            len(artifact),
			"operation_counts": map[string]any{"get": 1},
		},
	})
	if err != nil {
		t.Fatalf("marshal parity query lock: %v", err)
	}
	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create parity query source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "alpha-parity-source-lock.json"), lockRaw, 0o644); err != nil {
		t.Fatalf("write parity query source lock: %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		if sourceURL != "https://fixtures.polymetrics.invalid/alpha.json?download" {
			t.Fatalf("source-retain fetched %q, want exact locked query URL", sourceURL)
		}
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want locked-query parity success; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestSourceRetainDoesNotMisclassifyExpectedHTMLWithLoginText(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := []byte("<!doctype html><html><body>documentation navigation sign in help</body></html>")
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/reference", artifact)
	writeSourceRetainFixtureLock(t, defsRoot, lock)
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 || strings.Contains(stderr.String(), "wrong source") {
		t.Fatalf("source-retain exit/stderr = %d/%q, want expected HTML source retained", code, stderr.String())
	}
}

func TestSourceRetainEnrichesExistingManifestWithIdentityAndDetectedForm(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	writeSourceRetainFixtureLock(t, defsRoot, lock)
	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	writeSourceImportRetainedFixture(t, sourcesDir, "alpha", lock.Rest.sourceImportArtifact, artifact)
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want legacy-manifest enrichment; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	manifestRaw, err := os.ReadFile(filepath.Join(sourcesDir, "alpha-retained-artifacts.json"))
	if err != nil {
		t.Fatalf("read enriched retained manifest: %v", err)
	}
	for _, want := range []string{`"identity": "byte"`, `"retained_sha256"`, `"form": "openapi"`, `"version": "3.0.3"`} {
		if !bytes.Contains(manifestRaw, []byte(want)) {
			t.Fatalf("enriched retained manifest missing %s: %s", want, manifestRaw)
		}
	}
}

func TestSourceRetainVerifiesReorderedJSONByCanonicalIdentity(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	first := []byte(`{"a":{"nested":true},"b":[2,1]}`)
	second := []byte(`{"b":[2,1],"a":{"nested":true}}`)
	firstDigest := sha256.Sum256(first)
	canonicalDigest := sourceRetainTestCanonicalJSONDigest(t, first)
	if canonicalDigest != sourceRetainTestCanonicalJSONDigest(t, second) || bytes.Equal(first, second) {
		t.Fatal("fixture must prove equal canonical JSON with unequal bytes")
	}
	lockRaw, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"connector":      "alpha",
		"rest": map[string]any{
			"source_url":       "https://fixtures.polymetrics.invalid/discovery.json",
			"sha256":           hex.EncodeToString(firstDigest[:]),
			"bytes":            len(first),
			"identity":         "canonical_json",
			"canonical_sha256": canonicalDigest,
			"operations":       []any{},
		},
		"counts": map[string]any{"rest": 0, "graphql_query": 0, "graphql_mutation": 0, "total": 0},
	})
	if err != nil {
		t.Fatalf("marshal canonical identity lock: %v", err)
	}
	writeSourceRetainFixtureLockRaw(t, defsRoot, "alpha", lockRaw)
	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return second, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want canonical-identity success; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	manifestRaw, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-retained-artifacts.json"))
	if err != nil {
		t.Fatalf("read canonical retained artifact manifest: %v", err)
	}
	if !bytes.Contains(manifestRaw, []byte(`"identity": "canonical_json"`)) || !bytes.Contains(manifestRaw, []byte(canonicalDigest)) {
		t.Fatalf("canonical retained manifest = %s", manifestRaw)
	}
	lock, err := parseSourceImportLock(lockRaw, "alpha")
	if err != nil {
		t.Fatalf("parse canonical source lock: %v", err)
	}
	retainedFetcher, err := newSourceImportRetainedArtifactFetcher(filepath.Join(defsRoot, "alpha", "sources"), "alpha", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct canonical retained fetcher: %v", err)
	}
	got, err := retainedFetcher.FetchArtifact(context.Background(), lock.Rest.sourceImportArtifact)
	if err != nil || !bytes.Equal(got, second) {
		t.Fatalf("canonical retained artifact = %q/%v, want second serialisation", got, err)
	}
}

func TestSourceRetainReportsWrongSourceBeforeDrift(t *testing.T) {
	t.Parallel()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	for _, tc := range []struct {
		name    string
		fetcher sourceImportFetchFunc
	}{
		{
			name: "http forbidden",
			fetcher: func(context.Context, string) ([]byte, error) {
				return nil, fmt.Errorf("source-lock artifact returned HTTP 403 Forbidden")
			},
		},
		{
			name: "landing page too small",
			fetcher: func(context.Context, string) ([]byte, error) {
				return []byte("landing"), nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defsRoot := t.TempDir()
			writeSourceRetainFixtureLock(t, defsRoot, lock)
			var stdout, stderr bytes.Buffer
			code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T07:02:03Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, tc.fetcher)
			if code != 1 || !strings.Contains(stderr.String(), "wrong source") || strings.Contains(stderr.String(), "source-lock refresh required") {
				t.Fatalf("source-retain exit/stderr = %d/%q, want wrong-source classification before drift", code, stderr.String())
			}
		})
	}
}

func sourceRetainTestCanonicalJSONDigest(t *testing.T, raw []byte) string {
	t.Helper()
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("parse canonical JSON fixture: %v", err)
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal canonical JSON fixture: %v", err)
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:])
}

func writeSourceRetainFixtureLock(t *testing.T, defsRoot string, lock sourceImportLock) []byte {
	t.Helper()
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal fixture source lock: %v", err)
	}
	writeSourceRetainFixtureLockRaw(t, defsRoot, lock.Connector, raw)
	return raw
}

func writeSourceRetainFixtureLockRaw(t *testing.T, defsRoot, connector string, raw []byte) {
	t.Helper()
	sourcesDir := filepath.Join(defsRoot, connector, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create fixture sources directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, connector+"-operation-source-lock.json"), raw, 0o644); err != nil {
		t.Fatalf("write fixture source lock: %v", err)
	}
}
