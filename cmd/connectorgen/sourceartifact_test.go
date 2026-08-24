package main

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestSourceRetainWritesVerifiedLegacyParityArtifactWithoutChangingLock(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lockRaw := writeSourceRetainLegacyParityFixtureLock(t, defsRoot, "alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)

	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T12:05:00Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "1 artifact(s) retained") || stderr.Len() != 0 {
		t.Fatalf("source-retain output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	after, err := os.ReadFile(filepath.Join(sourcesDir, "alpha-parity-source-lock.json"))
	if err != nil {
		t.Fatalf("read legacy parity lock after retain: %v", err)
	}
	if !bytes.Equal(after, lockRaw) {
		t.Fatal("source-retain changed the legacy parity lock")
	}
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, "alpha", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct retained reader: %v", err)
	}
	got, err := fetcher.FetchArtifact(context.Background(), lock.Rest.sourceImportArtifact)
	if err != nil {
		t.Fatalf("read retained legacy parity artifact: %v", err)
	}
	if !bytes.Equal(got, artifact) {
		t.Fatalf("retained legacy parity bytes = %q, want %q", got, artifact)
	}
}

func TestSourceRetainLegacyParityArtifactPreservesFixedIdentityQuery(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	const sourceURL = "https://fixtures.polymetrics.invalid/alpha-openapi.yaml?version=v1"
	writeSourceRetainLegacyParityFixtureLock(t, defsRoot, "alpha", sourceURL, artifact)

	var stdout, stderr bytes.Buffer
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T12:05:00Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(_ context.Context, gotURL string) ([]byte, error) {
		if gotURL != sourceURL {
			t.Fatalf("fetched URL = %q, want %q", gotURL, sourceURL)
		}
		return artifact, nil
	}))
	if code != 0 {
		t.Fatalf("source-retain exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	manifestRaw, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-retained-artifacts.json"))
	if err != nil {
		t.Fatalf("read retained artifact manifest: %v", err)
	}
	var manifest sourceImportRetainedArtifactManifestDocument
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		t.Fatalf("decode retained artifact manifest: %v", err)
	}
	if len(manifest.Artifacts) != 1 || !manifest.Artifacts[0].IdentityQuery {
		t.Fatalf("legacy parity query provenance = %#v, want one identity-query artifact", manifest.Artifacts)
	}
}

func TestSourceRetainDoesNotFallBackToParityWhenOperationLockExists(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	writeSourceRetainLegacyParityFixtureLock(t, defsRoot, "alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	operationLockPath := filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-source-lock.json")
	if err := os.WriteFile(operationLockPath, []byte(`{"schema_version":99,"connector":"alpha"}`), 0o644); err != nil {
		t.Fatalf("write malformed operation lock: %v", err)
	}

	var stdout, stderr bytes.Buffer
	fetches := 0
	code := runSourceRetainWithFetcher([]string{"source-retain", "alpha", "--defs", defsRoot, "--retrieved-at", "2026-08-24T12:05:00Z", "--license", "fixture-license", "--terms", "fixture-terms"}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		fetches++
		return artifact, nil
	}))
	if code != 1 || fetches != 0 || !strings.Contains(stderr.String(), "unsupported schema version") {
		t.Fatalf("operation-lock authority exit=%d fetches=%d stderr=%q", code, fetches, stderr.String())
	}
}

func TestSourceRetainHelpExplainsLegacyParityFallback(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := runSourceRetain([]string{"source-retain", "--help"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("source-retain help exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy parity source lock") || !strings.Contains(stdout.String(), "source-import remain offline") {
		t.Fatalf("source-retain help = %q, want legacy parity/offline contract", stdout.String())
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
			artifacts := sourceRetainLockArtifacts(lock)
			if len(artifacts) != 1 {
				t.Fatalf("retained artifacts = %d, want 1", len(artifacts))
			}
			fetcher, err := newSourceImportRetainedArtifactFetcher(filepath.Join(defsRoot, "fixture", "sources"), "fixture", defaultSourceImportLimits())
			if err != nil {
				t.Fatalf("construct retained reader: %v", err)
			}
			got, err := fetcher.FetchArtifact(context.Background(), artifacts[0])
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

func writeSourceRetainLegacyParityFixtureLock(t *testing.T, defsRoot, connector, sourceURL string, artifact []byte) []byte {
	t.Helper()
	locked := sourceImportFixtureLock(connector, sourceURL, artifact).Rest.sourceImportArtifact
	raw, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"connector":      connector,
		"source_retrieval": map[string]any{
			"state": "fetched",
			"artifacts": []map[string]any{{
				"source_url": sourceURL,
				"sha256":     locked.SHA256,
				"bytes":      locked.Bytes,
			}},
		},
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy parity lock: %v", err)
	}
	sourcesDir := filepath.Join(defsRoot, connector, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create legacy parity sources directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, connector+"-parity-source-lock.json"), append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write legacy parity source lock: %v", err)
	}
	return append(raw, '\n')
}
