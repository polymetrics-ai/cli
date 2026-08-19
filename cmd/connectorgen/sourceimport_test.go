package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceImportProducesClosedCanonicalDescriptors(t *testing.T) {
	t.Parallel()
	locks := []sourceImportLock{
		loadSourceImportFixtureLock(t, "beta"),
		loadSourceImportFixtureLock(t, "alpha"),
	}
	fetch := fixtureSourceImportFetcher(t)

	got, err := importSourceLocks(context.Background(), locks, fetch, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import fixture locks: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("descriptor count = %d, want 5", len(got))
	}

	encoded, err := marshalSourceImportDescriptors(got)
	if err != nil {
		t.Fatalf("marshal descriptors: %v", err)
	}
	reversed, err := importSourceLocks(context.Background(), []sourceImportLock{locks[1], locks[0]}, fetch, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("reimport reversed locks: %v", err)
	}
	reversedEncoded, err := marshalSourceImportDescriptors(reversed)
	if err != nil {
		t.Fatalf("marshal reversed descriptors: %v", err)
	}
	if !bytes.Equal(encoded, reversedEncoded) {
		t.Fatalf("descriptor output changes with input order\nfirst: %s\nsecond: %s", encoded, reversedEncoded)
	}

	byID := map[string]sourceOperationDescriptor{}
	for _, descriptor := range got {
		byID[descriptor.SourceID] = descriptor
	}
	read := byID["getWidget"]
	if read.Connector != "alpha" || read.Method != "get" || read.Path != "/widgets/{widget_id}" {
		t.Fatalf("read identity = %#v", read)
	}
	if read.Source.URL != "https://fixtures.polymetrics.invalid/alpha-openapi.yaml" ||
		read.Source.SHA256 != "c4e02cbb72d377f29cb4e4a839224a515e33f3f921d8c647f3a3a95c20093bd7" ||
		read.Source.Bytes != 2472 ||
		read.Source.Location != `paths["/widgets/{widget_id}"].get` {
		t.Fatalf("provider provenance = %#v", read.Source)
	}
	if len(read.Request.Path) != 1 || len(read.Request.Query) != 1 || len(read.Request.Header) != 1 || read.Request.Body != nil {
		t.Fatalf("request schema separation = %#v", read.Request)
	}
	if read.Request.Path[0].Name != "widget_id" || read.Request.Query[0].Name != "include_archived" || read.Request.Header[0].Name != "x-request-id" {
		t.Fatalf("separated parameter names = %#v", read.Request)
	}
	if read.Pagination == nil || read.ByteLimits.Response != 4096 || !read.AuthScopes.Declared || len(read.AuthScopes.AnyOf) != 1 || len(read.AuthScopes.AnyOf[0].AllOf) != 1 || read.AuthScopes.AnyOf[0].AllOf[0].Scheme != "token" || len(read.AuthScopes.AnyOf[0].AllOf[0].Scopes) != 1 || read.AuthScopes.AnyOf[0].AllOf[0].Scopes[0] != "widgets.read" {
		t.Fatalf("read metadata = %#v", read)
	}
	response200 := descriptorResponse(t, read, "200")
	response403 := descriptorResponse(t, read, "403")
	responseJSON, err := json.Marshal([]sourceResponseDescriptor{response200, response403})
	if err != nil {
		t.Fatalf("marshal response declaration: %v", err)
	}
	for _, field := range []string{"rare_paid_result", "access_token", "Plan tier denial"} {
		if !strings.Contains(string(responseJSON), field) {
			t.Fatalf("response declaration silently lost %q: %s", field, responseJSON)
		}
	}

	write := byID["alpha.rest.post_/widgets"]
	if write.ProviderOperationID != "" || write.Request.Body == nil || write.Request.MediaType != "application/json" {
		t.Fatalf("derived-ID write descriptor = %#v", write)
	}
	if write.Output.Class != sourceOutputJSON {
		t.Fatalf("write output class = %q, want json", write.Output.Class)
	}
	if byID["healthStatus"].Output.Class != sourceOutputStatus || byID["exportPlain"].Output.Class != sourceOutputText || byID["downloadArchive"].Output.Class != sourceOutputBinary {
		t.Fatalf("output classes = health=%q plain=%q download=%q", byID["healthStatus"].Output.Class, byID["exportPlain"].Output.Class, byID["downloadArchive"].Output.Class)
	}
}

func TestSourceImportRejectsUnsafeOrUnboundedSourceForms(t *testing.T) {
	t.Parallel()
	baseLimits := defaultSourceImportLimits()
	cases := []struct {
		name     string
		artifact string
		want     string
		limits   sourceImportLimits
	}{
		{name: "external reference", artifact: "external-ref.json", want: "external reference", limits: baseLimits},
		{name: "unresolved reference", artifact: "unresolved-ref.json", want: "unresolved reference", limits: baseLimits},
		{name: "cyclic reference", artifact: "cyclic-ref.json", want: "reference cycle", limits: baseLimits},
		{name: "ambiguous request", artifact: "ambiguous-request.json", want: "ambiguous request schema", limits: baseLimits},
		{name: "duplicate identity", artifact: "duplicate-id.json", want: "duplicate source identity", limits: baseLimits},
		{name: "callback route", artifact: "callback-route.json", want: "callback-only route", limits: baseLimits},
		{name: "unbounded request", artifact: "unbounded-request.json", want: "unbounded request schema", limits: baseLimits},
		{name: "missing additional properties", artifact: "missing-additional-properties.json", want: "dynamic additionalProperties", limits: baseLimits},
		{name: "unsupported encoding", artifact: "unsupported-encoding.json", want: "unsupported request encoding", limits: baseLimits},
		{name: "invalid relative path", artifact: "invalid-relative-path.json", want: "connector-relative", limits: baseLimits},
		{name: "whitespace path", artifact: "whitespace-path.json", want: "connector-relative", limits: baseLimits},
		{name: "missing path parameter", artifact: "missing-path-parameter.json", want: "path placeholder", limits: baseLimits},
		{name: "multiple YAML documents", artifact: "multiple-documents.yaml", want: "multiple YAML documents", limits: baseLimits},
		{name: "reference depth", artifact: "deep-reference.json", want: "reference depth limit", limits: sourceImportLimits{MaxArtifactBytes: baseLimits.MaxArtifactBytes, MaxSchemaBytes: baseLimits.MaxSchemaBytes, MaxOperations: baseLimits.MaxOperations, MaxReferences: baseLimits.MaxReferences, MaxReferenceDepth: 1}},
		{name: "reference count", artifact: "many-references.json", want: "reference count limit", limits: sourceImportLimits{MaxArtifactBytes: baseLimits.MaxArtifactBytes, MaxSchemaBytes: baseLimits.MaxSchemaBytes, MaxOperations: baseLimits.MaxOperations, MaxReferences: 1, MaxReferenceDepth: baseLimits.MaxReferenceDepth}},
		{name: "operation count", artifact: "many-operations.json", want: "operation count limit", limits: sourceImportLimits{MaxArtifactBytes: baseLimits.MaxArtifactBytes, MaxSchemaBytes: baseLimits.MaxSchemaBytes, MaxOperations: 1, MaxReferences: baseLimits.MaxReferences, MaxReferenceDepth: baseLimits.MaxReferenceDepth}},
		{name: "schema size", artifact: "large-schema.json", want: "schema byte limit", limits: sourceImportLimits{MaxArtifactBytes: baseLimits.MaxArtifactBytes, MaxSchemaBytes: 16, MaxOperations: baseLimits.MaxOperations, MaxReferences: baseLimits.MaxReferences, MaxReferenceDepth: baseLimits.MaxReferenceDepth}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := loadSourceImportFixture(t, filepath.Join("invalid", tc.artifact))
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/invalid/"+tc.artifact, raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), tc.limits)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("import error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSourceImportRejectsUnsafeArtifactDestinations(t *testing.T) {
	t.Parallel()
	for _, sourceURL := range []string{
		"https://127.0.0.1/openapi.json",
		"https://artifact.example/openapi.json?x=1",
		"https://user@artifact.example/openapi.json",
	} {
		artifact := sourceImportArtifact{SourceURL: sourceURL, SHA256: strings.Repeat("0", sha256.Size*2), Bytes: 1}
		if err := validateSourceImportArtifact(artifact); err == nil {
			t.Fatalf("validate source artifact %q: accepted unsafe URL", sourceURL)
		}
	}

	privateLookup := batchArtifactLookupIPAddr(func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("169.254.169.254")}}, nil
	})
	fetcher := httpSourceImportFetcher{limits: defaultSourceImportLimits(), lookup: privateLookup}
	if _, err := fetcher.Fetch(context.Background(), "https://artifact.example/openapi.json"); err == nil || !strings.Contains(err.Error(), "not public") {
		t.Fatalf("private resolved source artifact error = %v", err)
	}

	publicLookup := batchArtifactLookupIPAddr(func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	redirectURL, err := url.Parse("https://artifact.example/redirected.json")
	if err != nil {
		t.Fatalf("parse redirect URL: %v", err)
	}
	if err := newSourceImportHTTPClient(publicLookup).CheckRedirect(&http.Request{URL: redirectURL}, nil); err == nil {
		t.Fatal("source importer accepted a redirect")
	}
}

func TestSourceImportProjectsSwaggerBodiesPathOverridesAndAuthGroups(t *testing.T) {
	t.Parallel()
	raw := loadSourceImportFixture(t, filepath.Join("supported", "swagger2-body.json"))
	lock := sourceImportFixtureLock("beta", "https://fixtures.polymetrics.invalid/swagger2-body.json", raw)
	descriptors, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return raw, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import Swagger body fixture: %v", err)
	}
	if len(descriptors) != 1 {
		t.Fatalf("Swagger descriptor count = %d, want 1", len(descriptors))
	}
	descriptor := descriptors[0]
	if descriptor.Request.Body == nil || !descriptor.Request.Body.Required || descriptor.Request.MediaType != "application/json" {
		t.Fatalf("Swagger body descriptor = %#v", descriptor.Request)
	}
	if len(descriptor.Request.Path) != 1 {
		t.Fatalf("effective Swagger path parameters = %#v", descriptor.Request.Path)
	}
	pathSchema, ok := descriptor.Request.Path[0].Schema.(map[string]any)
	if !ok || sourcePositiveInteger(pathSchema["maxLength"]) != 24 {
		t.Fatalf("effective Swagger path schema = %#v", descriptor.Request.Path[0].Schema)
	}
	if len(descriptor.Request.Query) != 1 || descriptor.Request.Query[0].Name != "include_archived" {
		t.Fatalf("inherited Swagger query parameters = %#v", descriptor.Request.Query)
	}
	auth := descriptor.AuthScopes
	if !auth.Declared || len(auth.AnyOf) != 2 || len(auth.AnyOf[0].AllOf) != 2 || auth.AnyOf[0].AllOf[0].Scheme != "apiKey" || len(auth.AnyOf[0].AllOf[0].Scopes) != 0 || auth.AnyOf[0].AllOf[1].Scheme != "oauth2" || len(auth.AnyOf[0].AllOf[1].Scopes) != 1 || auth.AnyOf[0].AllOf[1].Scopes[0] != "widgets.write" || len(auth.AnyOf[1].AllOf) != 1 || auth.AnyOf[1].AllOf[0].Scheme != "bearerAuth" || len(auth.AnyOf[1].AllOf[0].Scopes) != 0 {
		t.Fatalf("Swagger auth requirements = %#v", auth)
	}
}

func TestSourceOutputClassifiesOnlyJSONAsJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		mediaTypes []string
		want       sourceOutputClass
		wantErr    bool
	}{
		{name: "JSON", mediaTypes: []string{"application/vnd.api+json"}, want: sourceOutputJSON},
		{name: "text", mediaTypes: []string{"text/csv"}, want: sourceOutputText},
		{name: "image", mediaTypes: []string{"image/png"}, want: sourceOutputBinary},
		{name: "gzip", mediaTypes: []string{"application/gzip"}, want: sourceOutputBinary},
		{name: "mixed", mediaTypes: []string{"application/json", "image/png"}, wantErr: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := sourceOutputClassFor("get", tc.mediaTypes)
			if tc.wantErr {
				if err == nil {
					t.Fatal("mixed response media types were accepted")
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("output class = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestSourceImportRejectsArtifactDriftAndSizeBeforeParsing(t *testing.T) {
	t.Parallel()
	raw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", raw)
	lock.Rest.SHA256 = strings.Repeat("0", 64)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "source-lock refresh required") {
		t.Fatalf("digest drift error = %v", err)
	}

	limits := defaultSourceImportLimits()
	limits.MaxArtifactBytes = 1
	_, err = importSourceLock(context.Background(), sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", raw), sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "artifact byte limit") {
		t.Fatalf("oversized artifact error = %v", err)
	}
}

func TestSourceImportCommandContractAndMigrationDocumentation(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if exit := run([]string{"source-import", "--help"}, &stdout, &stderr); exit != 0 {
		t.Fatalf("source-import help exit = %d, stderr = %s", exit, stderr.String())
	}
	for _, forbidden := range []string{"--url", "--method", "--path", "--header", "--body", "--credential"} {
		if strings.Contains(stdout.String(), forbidden) {
			t.Fatalf("source-import help exposes %s: %s", forbidden, stdout.String())
		}
	}
	if !strings.Contains(stdout.String(), "source-import <connector>") || !strings.Contains(stdout.String(), "source-lock") {
		t.Fatalf("source-import help is incomplete: %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if exit := run([]string{"source-import", "alpha", "--help"}, &stdout, &stderr); exit != 0 || !strings.Contains(stdout.String(), "source-import <connector>") {
		t.Fatalf("connector-qualified source-import help exit = %d, stdout = %s, stderr = %s", exit, stdout.String(), stderr.String())
	}
	docs, err := os.ReadFile(filepath.Join("..", "..", "docs", "migration", "conventions.md"))
	if err != nil {
		t.Fatalf("read migration conventions: %v", err)
	}
	if !strings.Contains(string(docs), "connectorgen source-import") || !strings.Contains(string(docs), "source-lock refresh") {
		t.Fatalf("migration conventions lack source-import adoption contract")
	}
}

func TestSourceImportCommandUsesOnlyConnectorOwnedLockAndCheckMode(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	lockPath := filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-source-lock.json")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("create fixture lock directory: %v", err)
	}
	if err := os.WriteFile(lockPath, loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json")), 0o644); err != nil {
		t.Fatalf("write fixture lock: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "alpha-descriptors.json")
	var stdout, stderr bytes.Buffer
	args := []string{"source-import", "alpha", "--defs", defsRoot, "--out", outPath}
	if exit := runSourceImportWithFetcher(args, &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 0 {
		t.Fatalf("source-import exit = %d, stderr = %s", exit, stderr.String())
	}
	if raw, err := os.ReadFile(outPath); err != nil || !strings.Contains(string(raw), "getWidget") {
		t.Fatalf("descriptor output = %q, read error = %v", raw, err)
	}
	stdout.Reset()
	stderr.Reset()
	if exit := runSourceImportWithFetcher(append(args, "--check"), &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 0 {
		t.Fatalf("source-import --check exit = %d, stderr = %s", exit, stderr.String())
	}

	badLock := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json"))
	badLock = bytes.Replace(badLock, []byte("c4e02cbb72d377f29cb4e4a839224a515e33f3f921d8c647f3a3a95c20093bd7"), []byte(strings.Repeat("0", 64)), 1)
	if err := os.WriteFile(lockPath, badLock, 0o644); err != nil {
		t.Fatalf("write drifted lock: %v", err)
	}
	driftOutput := filepath.Join(t.TempDir(), "must-not-exist.json")
	stdout.Reset()
	stderr.Reset()
	if exit := runSourceImportWithFetcher([]string{"source-import", "alpha", "--defs", defsRoot, "--out", driftOutput}, &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 1 {
		t.Fatalf("drift source-import exit = %d, stderr = %s", exit, stderr.String())
	}
	if _, err := os.Stat(driftOutput); !os.IsNotExist(err) {
		t.Fatalf("drifted source lock created descriptor output: %v", err)
	}

	if _, err := loadConnectorSourceImportLock(defsRoot, "beta"); err == nil {
		t.Fatal("missing beta connector-owned lock was accepted")
	}
}

func TestSourceImportRejectsSourcesDirectoryEscapingConnectorBundle(t *testing.T) {
	t.Parallel()
	defsRoot := filepath.Join(t.TempDir(), "defs")
	bundleDir := filepath.Join(defsRoot, "alpha")
	externalSources := filepath.Join(t.TempDir(), "external-sources")
	if err := os.MkdirAll(externalSources, 0o755); err != nil {
		t.Fatalf("create external source directory: %v", err)
	}
	lockPath := filepath.Join(externalSources, "alpha-operation-source-lock.json")
	if err := os.WriteFile(lockPath, loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json")), 0o644); err != nil {
		t.Fatalf("write external source lock: %v", err)
	}
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("create bundle directory: %v", err)
	}
	if err := os.Symlink(externalSources, filepath.Join(bundleDir, "sources")); err != nil {
		t.Fatalf("create escaping source symlink: %v", err)
	}
	if _, err := loadConnectorSourceImportLock(defsRoot, "alpha"); err == nil || !strings.Contains(err.Error(), "outside connector-owned bundle") {
		t.Fatalf("escaping source directory error = %v", err)
	}
}

func loadSourceImportFixtureLock(t *testing.T, connector string) sourceImportLock {
	t.Helper()
	raw := loadSourceImportFixture(t, filepath.Join(connector, connector+"-operation-source-lock.json"))
	lock, err := parseSourceImportLock(raw, connector)
	if err != nil {
		t.Fatalf("parse %s fixture lock: %v", connector, err)
	}
	return lock
}

func fixtureSourceImportFetcher(t *testing.T) sourceImportFetchFunc {
	t.Helper()
	return func(_ context.Context, rawURL string) ([]byte, error) {
		switch rawURL {
		case "https://fixtures.polymetrics.invalid/alpha-openapi.yaml":
			return loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml")), nil
		case "https://fixtures.polymetrics.invalid/beta-openapi.json":
			return loadSourceImportFixture(t, filepath.Join("beta", "beta-openapi.json")), nil
		default:
			t.Fatalf("unexpected fixture source URL %q", rawURL)
			return nil, nil
		}
	}
}

func loadSourceImportFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "sourceimport", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return raw
}

func sourceImportFixtureLock(connector, sourceURL string, raw []byte) sourceImportLock {
	digest := sha256.Sum256(raw)
	return sourceImportLock{Connector: connector, Rest: sourceImportArtifact{SourceURL: sourceURL, SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(raw))}}
}

func descriptorResponse(t *testing.T, descriptor sourceOperationDescriptor, status string) sourceResponseDescriptor {
	t.Helper()
	for _, response := range descriptor.Responses {
		if response.Status == status {
			return response
		}
	}
	t.Fatalf("descriptor %q is missing response status %q", descriptor.SourceID, status)
	return sourceResponseDescriptor{}
}
