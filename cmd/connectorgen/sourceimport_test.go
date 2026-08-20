package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func TestSourceImportRejectsDuplicateJSONAndYAMLMembersWithPointers(t *testing.T) {
	t.Parallel()
	lock := []byte(`{"schema_version":1,"connector":"alpha","connector":"beta","rest":{"source_url":"https://fixtures.polymetrics.invalid/openapi.json","sha256":"0000000000000000000000000000000000000000000000000000000000000000","bytes":1}}`)
	if _, err := parseSourceImportLock(lock, "alpha"); err == nil || !strings.Contains(err.Error(), "/connector") {
		t.Fatalf("duplicate lock error = %v", err)
	}
	jsonArtifact := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}},"get":{"responses":{"204":{"description":"again"}}}}}}`)
	if _, _, err := parseSourceImportDocument(jsonArtifact); err == nil || !strings.Contains(err.Error(), "/paths/~1items/get") {
		t.Fatalf("duplicate artifact error = %v", err)
	}
	yamlArtifact := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  /items:\n    get: {responses: {'200': {description: ok}}}\n    get: {responses: {'204': {description: again}}}\n")
	if _, _, err := parseSourceImportDocument(yamlArtifact); err == nil || !strings.Contains(err.Error(), "/paths/~1items/get") {
		t.Fatalf("duplicate YAML artifact error = %v", err)
	}
}

func TestSourceImportPreservesLiteralReferenceFieldsAndReferenceSiblings(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"ok":{"description":"original","content":{"application/json":{"schema":{"type":"object","additionalProperties":false,"properties":{"$ref":{"type":"string","maxLength":8},"example":{"type":"string","maxLength":8}}}}}}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/ok","description":"override"}}}}}}`)
	result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
	if len(result.Operations) != 1 {
		t.Fatalf("operation count = %d", len(result.Operations))
	}
	response := descriptorResponse(t, result.Operations[0], "200")
	declaration, ok := response.Declaration.(map[string]any)
	if !ok || declaration["description"] != "override" {
		t.Fatalf("resolved reference sibling = %#v", response.Declaration)
	}
	content := declaration["content"].(map[string]any)
	schema := content["application/json"].(map[string]any)["schema"].(map[string]any)
	properties := schema["properties"].(map[string]any)
	if _, exists := properties["$ref"]; !exists {
		t.Fatalf("literal property named $ref was lost: %#v", properties)
	}
}

func TestSourceImportPreservesInboundEventsAndExtensionsAsMergeBlocked(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{"invoice.created":{"post":{"responses":{"200":{"description":"ok"}}}}},"paths":{"x-provider-metadata":{"tier":"test"},"/deliver":{"x-route-metadata":{"owner":"provider"},"post":{"operationId":"deliver","callbacks":{"notify":{"{$request.body#/callback_url}":{"post":{"responses":{"202":{"description":"accepted"}}}}}},"responses":{"202":{"description":"accepted"}}}}}}`)
	result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
	if len(result.Operations) != 1 || len(result.InboundEvents) != 2 || len(result.Extensions) != 2 {
		t.Fatalf("result = %#v", result)
	}
	for _, event := range result.InboundEvents {
		if !event.Runtime.MergeBlocked || len(event.Runtime.Gaps) != 1 || event.Runtime.Gaps[0].Foundation != "cli-webhook-event-surface-foundation-r1" {
			t.Fatalf("event runtime = %#v", event.Runtime)
		}
	}
	encoded, err := marshalSourceImportResult(result)
	if err != nil || !strings.Contains(string(encoded), `"merge_blocked": true`) || !strings.Contains(string(encoded), "cli-webhook-event-surface-foundation-r1") {
		t.Fatalf("event descriptor output = %s, %v", encoded, err)
	}
	callbackOnly := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{"only":{"post":{"responses":{"200":{"description":"ok"}}}}}}`)
	callbackResult := importInlineSourceResult(t, callbackOnly, defaultSourceImportLimits())
	if len(callbackResult.Operations) != 0 || len(callbackResult.InboundEvents) != 1 {
		t.Fatalf("callback-only result = %#v", callbackResult)
	}
	externalInbound := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{"only":{"post":{"responses":{"200":{"content":{"application/json":{"schema":{"$ref":"https://example.invalid/schema.json"}}}}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inbound-external.json", externalInbound)
	_, err = importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return externalInbound, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "external reference") {
		t.Fatalf("inbound external reference error = %v", err)
	}
}

func TestSourceImportPreservesServerOverridesAndParameterWireContracts(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"servers":[{"url":"https://{region}.example.invalid/v1","variables":{"region":{"default":"us","enum":["us","eu"]}}}],"paths":{"/items":{"servers":[{"url":"https://path.example.invalid"}],"get":{"operationId":"items","servers":[{"url":"https://operation.example.invalid/{version}","variables":{"version":{"default":"v2"}}}],"parameters":[{"name":"filter","in":"query","style":"deepObject","explode":true,"allowReserved":true,"schema":{"type":"object","additionalProperties":false,"properties":{"tag":{"type":"string","pattern":"^[a-z]+$","maxLength":20}}}}],"responses":{"200":{"description":"ok"}}}}}}`)
	result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
	descriptor := result.Operations[0]
	if !descriptor.Servers.Root.Declared || !descriptor.Servers.PathItem.Declared || !descriptor.Servers.Operation.Declared || !descriptor.Runtime.MergeBlocked {
		t.Fatalf("server routing = %#v", descriptor)
	}
	if len(descriptor.Servers.Gaps) != 1 || descriptor.Servers.Gaps[0].Foundation != "cli-operation-route-override-foundation-r1" {
		t.Fatalf("server gap = %#v", descriptor.Servers.Gaps)
	}
	parameter := descriptor.Request.Query[0]
	if parameter.Wire.Style != "deepObject" || parameter.Wire.Explode == nil || !*parameter.Wire.Explode || parameter.Wire.AllowReserved == nil || !*parameter.Wire.AllowReserved || len(parameter.Wire.Gaps) != 1 {
		t.Fatalf("wire contract = %#v", parameter.Wire)
	}
	schema := parameter.Schema.(map[string]any)
	if schema["properties"].(map[string]any)["tag"].(map[string]any)["pattern"] != "^[a-z]+$" {
		t.Fatalf("parameter schema constraints = %#v", schema)
	}
	swagger := []byte(`{"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"tag","in":"query","type":"string","pattern":"^[a-z]+$","maxLength":20,"collectionFormat":"pipes"}],"responses":{"200":{"description":"ok"}}}}}}`)
	swaggerResult := importInlineSourceResult(t, swagger, defaultSourceImportLimits())
	swaggerParameter := swaggerResult.Operations[0].Request.Query[0]
	if swaggerParameter.Wire.CollectionFormat != "pipes" || len(swaggerParameter.Wire.Gaps) != 1 || swaggerParameter.Schema.(map[string]any)["pattern"] != "^[a-z]+$" {
		t.Fatalf("Swagger wire contract = %#v", swaggerParameter)
	}
}

func TestSourceImportRejectsDynamicAndInvalidBoundedRequestContracts(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		schema  string
		wantErr string
	}{
		{name: "dynamic reference", schema: `{"$dynamicRef":"#node","type":"string","maxLength":8}`, wantErr: "cli-openapi-dynamic-ref-foundation-r1"},
		{name: "dynamic anchor", schema: `{"$dynamicAnchor":"node","type":"string","maxLength":8}`, wantErr: "cli-openapi-dynamic-ref-foundation-r1"},
		{name: "null numeric bounds", schema: `{"type":"number","minimum":null,"maximum":null}`, wantErr: "finite number"},
		{name: "contradictory numeric bounds", schema: `{"type":"number","minimum":4,"maximum":3}`, wantErr: "contradictory request schema numeric bounds"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + tc.schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/bounds.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("import error = %v, want %q", err, tc.wantErr)
			}
		})
	}
	enum := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"string","enum":["asc","desc"]}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	result := importInlineSourceResult(t, enum, defaultSourceImportLimits())
	if result.Operations[0].Request.Body == nil {
		t.Fatal("finite enum request body was not imported")
	}
}

func TestSourceImportBoundsAggregateResolvedDescriptorsAndKeepsMixedResponseMedia(t *testing.T) {
	t.Parallel()
	largeDescription := strings.Repeat("x", 6000)
	amplified := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"large":{"description":"` + largeDescription + `"}}},"paths":{"/one":{"get":{"responses":{"200":{"$ref":"#/components/responses/large"}}}},"/two":{"get":{"responses":{"200":{"$ref":"#/components/responses/large"}}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxResolvedDescriptorBytes = 10000
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/amplified.json", amplified)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return amplified, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit") {
		t.Fatalf("amplification error = %v", err)
	}
	mixed := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/archive":{"get":{"responses":{"200":{"description":"archive","content":{"application/zip":{"schema":{"type":"string"}}}},"403":{"description":"denied","content":{"application/json":{"schema":{"type":"object","properties":{"error":{"type":"string"}}}}}}}}}}}`)
	result := importInlineSourceResult(t, mixed, defaultSourceImportLimits())
	operation := result.Operations[0]
	if len(operation.Output.Success) != 1 || operation.Output.Success[0].Status != "200" || operation.Output.Success[0].Class != sourceOutputBinary {
		t.Fatalf("success output = %#v", operation.Output)
	}
	response403 := descriptorResponse(t, operation, "403")
	if len(response403.Media) != 1 || response403.Media[0].MediaType != "application/json" || response403.Media[0].Class != sourceOutputJSON {
		t.Fatalf("error response media = %#v", response403)
	}
}

func TestSourceImportResolvesGrammarScopedLinkAndExampleReferences(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"links":{"next":{"operationId":"listNext"}},"examples":{"provider":{"summary":"provider example","value":{"$ref":"literal provider field"}}}},"paths":{"/items":{"get":{"operationId":"listNext","responses":{"200":{"description":"ok","links":{"next":{"$ref":"#/components/links/next"}},"content":{"application/json":{"examples":{"provider":{"$ref":"#/components/examples/provider"}}}}}}}}}}`)
	result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
	response := descriptorResponse(t, result.Operations[0], "200")
	declaration := response.Declaration.(map[string]any)
	link := declaration["links"].(map[string]any)["next"].(map[string]any)
	if link["operationId"] != "listNext" {
		t.Fatalf("resolved link = %#v", link)
	}
	example := declaration["content"].(map[string]any)["application/json"].(map[string]any)["examples"].(map[string]any)["provider"].(map[string]any)
	if example["value"].(map[string]any)["$ref"] != "literal provider field" {
		t.Fatalf("resolved example lost literal field = %#v", example)
	}

	for _, tc := range []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "external link reference",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","links":{"next":{"$ref":"https://example.invalid/link.json"}}}}}}}}`),
			want: "external reference",
		},
		{
			name: "wrong response target kind",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"notResponse":{"type":"string"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/schemas/notResponse"}}}}}}`),
			want: "expected kind",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/reference-kind.json", tc.raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return tc.raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("import error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSourceImportValidatesParameterRepresentationsAndContentSchemas(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "schema and content",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","schema":{"type":"string","maxLength":8},"content":{"application/json":{"schema":{"type":"string","maxLength":8}}}}],"responses":{"200":{"description":"ok"}}}}}}`),
			want: "exactly one of schema or content",
		},
		{
			name: "unbounded content schema",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","content":{"application/json":{"schema":{"type":"object","properties":{"tag":{"type":"string","maxLength":8}}}}}}],"responses":{"200":{"description":"ok"}}}}}}`),
			want: "dynamic additionalProperties",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/parameter-content.json", tc.raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return tc.raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("import error = %v, want %q", err, tc.want)
			}
		})
	}
	valid := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","content":{"application/json":{"schema":{"type":"string","maxLength":8}}}}],"responses":{"200":{"description":"ok"}}}}}}`)
	result := importInlineSourceResult(t, valid, defaultSourceImportLimits())
	if len(result.Operations[0].Request.Query) != 1 || result.Operations[0].Request.Query[0].Content == nil || result.Operations[0].Request.Query[0].Schema != nil {
		t.Fatalf("content parameter descriptor = %#v", result.Operations[0].Request.Query)
	}
}

func TestSourceImportValidatesPrefixItemsAndUnsupportedApplicators(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		schema string
		want   string
	}{
		{
			name:   "unbounded prefix item",
			schema: `{"type":"array","maxItems":1,"items":{"type":"string","maxLength":8},"prefixItems":[{"type":"object","additionalProperties":true,"properties":{}}]}`,
			want:   "dynamic additionalProperties",
		},
		{
			name:   "unsupported contains",
			schema: `{"type":"array","maxItems":1,"items":{"type":"string","maxLength":8},"contains":{"type":"string","maxLength":8}}`,
			want:   "unsupported request schema keyword contains",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + tc.schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/prefix-items.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("import error = %v, want %q", err, tc.want)
			}
		})
	}
	valid := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"array","maxItems":1,"prefixItems":[{"type":"string","maxLength":8}]}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	if result := importInlineSourceResult(t, valid, defaultSourceImportLimits()); result.Operations[0].Request.Body == nil {
		t.Fatal("bounded prefixItems request body was not imported")
	}
}

func TestSourceImportPreservesExactFormAndSwaggerRouteBinding(t *testing.T) {
	t.Parallel()
	for _, raw := range [][]byte{
		[]byte(`{"openapi":"3.1.0","swagger":"2.0","info":{"title":"x","version":"1"},"paths":{}}`),
		[]byte(`{"openapi":"3.2.0","info":{"title":"x","version":"1"},"paths":{}}`),
		[]byte(`{"openapi":"3.1.0.1","info":{"title":"x","version":"1"},"paths":{}}`),
	} {
		if _, _, err := parseSourceImportDocument(raw); err == nil {
			t.Fatalf("unsupported document form was accepted: %s", raw)
		}
	}
	openAPI30Sibling := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"responses":{"ok":{"description":"ok"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/ok","description":"override"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/openapi30-sibling.json", openAPI30Sibling)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return openAPI30Sibling, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "ambiguous response reference") {
		t.Fatalf("OpenAPI 3.0 reference sibling error = %v", err)
	}
	swagger := []byte(`{"swagger":"2.0","info":{"title":"x","version":"1"},"host":"api.example.invalid","basePath":"/v1","schemes":["https","http"],"paths":{"/items":{"get":{"schemes":["http"],"responses":{"200":{"description":"ok"}}}}}}`)
	result := importInlineSourceResult(t, swagger, defaultSourceImportLimits())
	descriptor := result.Operations[0]
	if descriptor.Source.Form != "swagger" || descriptor.Source.Version != "2.0" || descriptor.Path != "/items" || descriptor.Servers.Swagger == nil || !descriptor.Servers.Swagger.Declared || descriptor.Servers.Swagger.Host != "api.example.invalid" || descriptor.Servers.Swagger.BasePath != "/v1" || descriptor.Servers.Swagger.EffectivePath != "/v1/items" || !descriptor.Runtime.MergeBlocked {
		t.Fatalf("Swagger route binding = %#v", descriptor)
	}
	if got := descriptor.Servers.Swagger.RootSchemes; len(got) != 2 || got[0] != "https" || got[1] != "http" {
		t.Fatalf("Swagger root schemes = %#v", got)
	}
	if got := descriptor.Servers.Swagger.OperationSchemes; len(got) != 1 || got[0] != "http" {
		t.Fatalf("Swagger operation schemes = %#v", got)
	}
	if got := descriptor.Servers.Swagger.Schemes; len(got) != 1 || got[0] != "http" {
		t.Fatalf("Swagger effective schemes = %#v", got)
	}
}

func TestSourceImportRejectsCrossKindIdentitiesAndNonStringYAMLKeys(t *testing.T) {
	t.Parallel()
	collision := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{"invoice.created":{"post":{"responses":{"200":{"description":"ok"}}}}},"paths":{"/items":{"get":{"operationId":"alpha.inbound.webhook.invoice.created","responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/collision.json", collision)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return collision, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "duplicate source identity") {
		t.Fatalf("cross-kind identity error = %v", err)
	}
	yaml := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  !!int \"01\": {}\n")
	if _, _, err := parseSourceImportDocument(yaml); err == nil || !strings.Contains(err.Error(), "must be a string") {
		t.Fatalf("non-string YAML key error = %v", err)
	}
}

func TestSourceImportReservesSchemaExpansionBeforeRetainingReferences(t *testing.T) {
	t.Parallel()
	var properties strings.Builder
	for index, name := range []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"} {
		if index > 0 {
			properties.WriteByte(',')
		}
		properties.WriteString(`"` + name + `":{"$ref":"#/components/schemas/Large"}`)
	}
	largeDescription := strings.Repeat("x", 2400)
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Large":{"type":"object","additionalProperties":false,"description":"` + largeDescription + `","properties":{"value":{"type":"string","maxLength":8}}}}},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"object","additionalProperties":false,"properties":{` + properties.String() + `}}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxResolvedDescriptorBytes = 9000
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/schema-amplification.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining request media") {
		t.Fatalf("schema expansion error = %v", err)
	}
}

func TestSourceImportTreatsFiniteCompositeEnumsAsBounded(t *testing.T) {
	t.Parallel()
	for _, schema := range []string{
		`{"type":"object","enum":[{"mode":"fixed"}],"additionalProperties":true,"properties":{"mode":{"type":"string"}}}`,
		`{"type":"array","enum":[["fixed"]],"items":{"type":"string"}}`,
	} {
		raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
		result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
		if result.Operations[0].Request.Body == nil {
			t.Fatalf("finite enum request body was not imported: %#v", result.Operations[0].Request)
		}
	}
}

func TestSourceImportScopesReferenceResolutionToOpenAPIGrammar(t *testing.T) {
	t.Parallel()
	swagger := []byte(`{"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","links":{"next":{"$ref":"https://example.invalid/literal"}}}}}}}}`)
	result := importInlineSourceResult(t, swagger, defaultSourceImportLimits())
	response := descriptorResponse(t, result.Operations[0], "200")
	links := response.Declaration.(map[string]any)["links"].(map[string]any)
	if links["next"].(map[string]any)["$ref"] != "https://example.invalid/literal" {
		t.Fatalf("literal Swagger response field changed: %#v", links)
	}

	openAPI := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"multipart/form-data":{"schema":{"type":"object","additionalProperties":false,"properties":{}},"encoding":{"payload":{"headers":{"X-Trace":{"$ref":"https://example.invalid/header"}}}}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/encoding-header.json", openAPI)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return openAPI, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "external reference") {
		t.Fatalf("encoding header reference error = %v", err)
	}
}

func TestSourceImportClosesSchemaGrammarAndOperationIDBoundaries(t *testing.T) {
	t.Parallel()
	nonStringOperationID := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":7,"responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/operation-id.json", nonStringOperationID)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return nonStringOperationID, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), `paths["/items"].get.operationId must be a string`) {
		t.Fatalf("operation ID error = %v", err)
	}

	resolved := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Bounded":{"type":"string","maxLength":8}},"parameters":{"filter":{"name":"filter","in":"query","schema":{"type":"string","maxLength":8}}}},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","schema":{"$ref":"#/components/parameters/filter/schema"}}],"responses":{"200":{"description":"ok","schema":{"$ref":"https://provider.invalid/literal-response-field"},"content":{"application/json":{"schema":{"type":"array","contentSchema":{"$ref":"#/components/schemas/Bounded"},"unevaluatedItems":{"$ref":"#/components/schemas/Bounded"},"prefixItems":[{"type":"array","unevaluatedItems":false}]}}}}}}}}}`)
	result := importInlineSourceResult(t, resolved, defaultSourceImportLimits())
	query := result.Operations[0].Request.Query[0].Schema.(map[string]any)
	if query["type"] != "string" {
		t.Fatalf("nested parameter schema reference = %#v", query)
	}
	response := descriptorResponse(t, result.Operations[0], "200")
	declaration := response.Declaration.(map[string]any)
	if declaration["schema"].(map[string]any)["$ref"] != "https://provider.invalid/literal-response-field" {
		t.Fatalf("OpenAPI response literal schema field = %#v", declaration["schema"])
	}
	schema := declaration["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if schema["contentSchema"].(map[string]any)["type"] != "string" || schema["unevaluatedItems"].(map[string]any)["type"] != "string" || schema["prefixItems"].([]any)[0].(map[string]any)["unevaluatedItems"] != false {
		t.Fatalf("resolved OpenAPI 3.1 schema positions = %#v", schema)
	}

	externalContentSchema := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"string","maxLength":8,"contentSchema":{"$ref":"https://provider.invalid/content-schema"}}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	lock = sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/content-schema.json", externalContentSchema)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return externalContentSchema, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "external reference") {
		t.Fatalf("external contentSchema error = %v", err)
	}
}

func TestSourceImportRecognizesFiniteOpenAPI31SchemaForms(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		schema string
	}{
		{name: "const", schema: `{"const":"fixed"}`},
		{name: "nullable string union", schema: `{"type":["string","null"],"maxLength":8}`},
		{name: "closed tuple", schema: `{"type":"array","prefixItems":[{"type":"string","maxLength":8}],"items":false}`},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + tc.schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
			result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
			if result.Operations[0].Request.Body == nil {
				t.Fatalf("bounded %s request was not imported", tc.name)
			}
		})
	}

	openAPI30Union := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":["string","null"],"maxLength":8}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/openapi30-union.json", openAPI30Union)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return openAPI30Union, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "type union requires OpenAPI 3.1") {
		t.Fatalf("OpenAPI 3.0 union error = %v", err)
	}

	contradictory := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"number","minimum":9007199254740993,"maximum":9007199254740992}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	lock = sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/exact-bounds.json", contradictory)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return contradictory, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "contradictory request schema numeric bounds") {
		t.Fatalf("exact numeric bounds error = %v", err)
	}
}

func TestSourceImportBoundsInlineSchemaResolution(t *testing.T) {
	t.Parallel()
	schema := `{"type":"string","maxLength":1}`
	for range 6 {
		schema = `{"type":"array","maxItems":1,"items":` + schema + `}`
	}
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
	for _, tc := range []struct {
		name   string
		limits sourceImportLimits
		want   string
	}{
		{name: "depth", limits: func() sourceImportLimits {
			limits := defaultSourceImportLimits()
			limits.MaxReferenceDepth = 3
			return limits
		}(), want: "schema depth limit exceeded"},
		{name: "nodes", limits: func() sourceImportLimits {
			limits := defaultSourceImportLimits()
			limits.MaxSchemaNodes = 10
			return limits
		}(), want: "node limit exceeded"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inline-schema.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), tc.limits)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s inline schema error = %v, want %q", tc.name, err, tc.want)
			}
		})
	}
}

func TestSourceImportIndexesSchemaReferenceTargetsByGrammarPosition(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","schema":{"$ref":"#/paths/~1items/get/responses/200"}}],"responses":{"200":{"description":"ok","type":"string","maxLength":8}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/wrong-schema-target.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "expected kind") {
		t.Fatalf("wrong schema target error = %v", err)
	}
}

func TestSourceImportPreflightsUnusedGrammarObjects(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "unused schema external reference",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Unused":{"$ref":"https://provider.invalid/unused.json"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
			want: "external reference",
		},
		{
			name: "unused schema cycle",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Unused":{"$ref":"#/components/schemas/Unused"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
			want: "reference cycle",
		},
		{
			name: "unused dynamic schema",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Unused":{"$dynamicRef":"#unused"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
			want: "cli-openapi-dynamic-ref-foundation-r1",
		},
		{
			name: "unused example external reference",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"examples":{"Unused":{"$ref":"https://provider.invalid/unused-example.json"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
			want: "external reference",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/unused-grammar.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unused grammar error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSourceImportPreservesExactYAMLNumericBounds(t *testing.T) {
	t.Parallel()
	raw := []byte(`openapi: 3.1.0
info:
  title: x
  version: "1"
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: number
              minimum: 100000000000000000001
              maximum: 100000000000000000000
      responses:
        "200":
          description: ok
`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/exact-yaml-bounds.yaml", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "contradictory request schema numeric bounds") {
		t.Fatalf("exact YAML numeric bounds error = %v", err)
	}
}

func TestSourceImportRejectsDeepSchemaBeforeRetention(t *testing.T) {
	t.Parallel()
	schema := `{"type":"string","maxLength":1}`
	for range 64 {
		schema = `{"type":"array","maxItems":1,"items":` + schema + `}`
	}
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":` + schema + `}}},"responses":{"200":{"description":"ok"}}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxReferenceDepth = 8
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/deep-retention.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "schema depth limit exceeded") {
		t.Fatalf("deep schema error = %v", err)
	}
}

func TestSourceImportGatesSchemaKeywordShapesByDocumentForm(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "Swagger const",
			raw:  []byte(`{"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"parameters":[{"name":"body","in":"body","schema":{"const":"fixed"}}],"consumes":["application/json"],"responses":{"200":{"description":"ok"}}}}}}`),
			want: "keyword const",
		},
		{
			name: "OpenAPI 3.0 numeric exclusive minimum",
			raw:  []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"number","minimum":0,"exclusiveMinimum":1,"maximum":2}}}},"responses":{"200":{"description":"ok"}}}}}}`),
			want: "must be a boolean",
		},
		{
			name: "OpenAPI 3.1 boolean exclusive minimum",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"number","minimum":0,"exclusiveMinimum":true,"maximum":2}}}},"responses":{"200":{"description":"ok"}}}}}}`),
			want: "must be a finite number",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/version-schema.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("version schema error = %v, want %q", err, tc.want)
			}
		})
	}

	valid := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"post":{"requestBody":{"content":{"application/json":{"schema":{"type":"string","nullable":true,"maxLength":8}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	if result := importInlineSourceResult(t, valid, defaultSourceImportLimits()); result.Operations[0].Request.Body == nil {
		t.Fatal("valid OpenAPI 3.0 nullable schema was not imported")
	}
}

func TestSourceImportReservesResponseExpansionBeforeAppend(t *testing.T) {
	t.Parallel()
	const responseCount = 40
	largeDescription := strings.Repeat("x", 64<<10)
	var responses strings.Builder
	for index := 0; index < responseCount; index++ {
		if index > 0 {
			responses.WriteByte(',')
		}
		fmt.Fprintf(&responses, `"%d":{"$ref":"#/components/responses/Large"}`, 200+index)
	}
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"Large":{"description":"` + largeDescription + `"}}},"paths":{"/items":{"get":{"responses":{` + responses.String() + `}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxResolvedDescriptorBytes = 1 << 20
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/response-amplification.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining response") {
		t.Fatalf("response expansion error = %v", err)
	}
}

func TestSourceImportIndexesXPrefixedComponentsWithoutTreatingThemAsExtensions(t *testing.T) {
	t.Parallel()
	valid := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"headers":{"x-trace":{"description":"trace","schema":{"type":"string","maxLength":8}}}},"paths":{"x-provider-metadata":{"$ref":"https://provider.invalid/literal"},"/items":{"get":{"responses":{"200":{"description":"ok","headers":{"x-trace":{"$ref":"#/components/headers/x-trace"}}}}}}}}`)
	result := importInlineSourceResult(t, valid, defaultSourceImportLimits())
	if len(result.Extensions) != 1 || result.Extensions[0].Location != `paths["x-provider-metadata"]` {
		t.Fatalf("path extension = %#v", result.Extensions)
	}
	response := descriptorResponse(t, result.Operations[0], "200")
	declaration, ok := response.Declaration.(map[string]any)
	if !ok {
		t.Fatalf("response declaration = %T", response.Declaration)
	}
	headers, ok := declaration["headers"].(map[string]any)
	if !ok || headers["x-trace"] == nil {
		t.Fatalf("x-prefixed header component was not resolved: %#v", declaration)
	}

	for _, component := range []string{"schemas", "responses", "parameters", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks", "pathItems"} {
		component := component
		t.Run(component, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"` + component + `":{"x-unused":{"$ref":"https://provider.invalid/external"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/x-prefixed-component.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), "external reference") {
				t.Fatalf("%s x-prefixed component error = %v", component, err)
			}
		})
	}
}

func TestSourceImportRejectsNonRequestContentEncodings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  []byte
	}{
		{
			name: "response content",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","content":{"application/json":{"encoding":{"part":{"headers":{"X-Trace":{"$ref":"https://provider.invalid/header"}}}}}}}}}}}}`),
		},
		{
			name: "parameter content",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","content":{"application/json":{"encoding":{"part":{"headers":{"X-Trace":{"$ref":"https://provider.invalid/header"}}}}}}}],"responses":{"200":{"description":"ok"}}}}}}`),
		},
		{
			name: "header content",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","headers":{"X-Trace":{"content":{"application/json":{"encoding":{"part":{"headers":{"X-Trace":{"$ref":"https://provider.invalid/header"}}}}}}}}}}}}}}`),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/non-request-encoding.json", tc.raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return tc.raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), "unsupported encoding on non-request content") {
				t.Fatalf("%s encoding error = %v", tc.name, err)
			}
		})
	}
}

func TestSourceImportReservesInboundResponseExpansionBeforeResolution(t *testing.T) {
	t.Parallel()
	const responseCount = 40
	largeDescription := strings.Repeat("x", 64<<10)
	var responses strings.Builder
	for index := 0; index < responseCount; index++ {
		if index > 0 {
			responses.WriteByte(',')
		}
		fmt.Fprintf(&responses, `"%d":{"$ref":"#/components/responses/Large"}`, 200+index)
	}
	cases := []struct {
		name string
		raw  []byte
	}{
		{
			name: "webhook",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"Large":{"description":"` + largeDescription + `"}}},"webhooks":{"invoice":{"post":{"responses":{` + responses.String() + `}}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
		},
		{
			name: "callback",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"Large":{"description":"` + largeDescription + `"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}},"callbacks":{"deliver":{"{$request.body#/url}":{"post":{"responses":{` + responses.String() + `}}}}}}}}}`),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			limits := defaultSourceImportLimits()
			limits.MaxResolvedDescriptorBytes = 1 << 20
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inbound-response-amplification.json", tc.raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return tc.raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining inbound event") {
				t.Fatalf("%s inbound response expansion error = %v", tc.name, err)
			}
		})
	}
}

func TestSourceImportBoundsGrammarIndexBeforeSortingComponents(t *testing.T) {
	t.Parallel()
	var schemas strings.Builder
	for index := 0; index < 8; index++ {
		if index > 0 {
			schemas.WriteByte(',')
		}
		fmt.Fprintf(&schemas, `"S%d":{"type":"string","maxLength":1}`, index)
	}
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{` + schemas.String() + `}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxSchemaNodes = 3
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/grammar-index-limit.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "source grammar position limit exceeded") {
		t.Fatalf("grammar index limit error = %v", err)
	}
}

func TestSourceImportReservesResolvedResponseChildrenBeforeCloning(t *testing.T) {
	t.Parallel()
	const responseCount = 24
	largeExample := strings.Repeat("x", 64<<10)
	var responses strings.Builder
	for index := 0; index < responseCount; index++ {
		if index > 0 {
			responses.WriteByte(',')
		}
		fmt.Fprintf(&responses, `"%d":{"description":"ok","content":{"application/json":{"examples":{"provider":{"$ref":"#/components/examples/Large"}}}}}`, 200+index)
	}
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"examples":{"Large":{"value":{"payload":"` + largeExample + `"}},"UnusedExternal":{"$ref":"https://provider.invalid/example"}}},"paths":{"/items":{"get":{"responses":{` + responses.String() + `}}}}}`)
	limits := defaultSourceImportLimits()
	limits.MaxResolvedDescriptorBytes = 1 << 20
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/referenced-response-children.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining response") {
		t.Fatalf("resolved response child budget error = %v", err)
	}
}

func TestSourceImportReservesOperationAndInboundCountsAtDiscovery(t *testing.T) {
	const count = defaultSourceImportOperations + 1
	cases := []struct {
		name string
		raw  func() []byte
		want string
	}{
		{
			name: "operations",
			raw: func() []byte {
				var paths strings.Builder
				for index := 0; index < count; index++ {
					if index > 0 {
						paths.WriteByte(',')
					}
					fmt.Fprintf(&paths, `"/items/%d":{"get":{"responses":{"200":{"description":"ok"}}}}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{` + paths.String() + `}}`)
			},
			want: "operation count limit exceeded",
		},
		{
			name: "webhooks",
			raw: func() []byte {
				var webhooks strings.Builder
				for index := 0; index < count; index++ {
					if index > 0 {
						webhooks.WriteByte(',')
					}
					fmt.Fprintf(&webhooks, `"event-%d":{"post":{"responses":{"200":{"description":"ok"}}}}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{` + webhooks.String() + `},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
			},
			want: "inbound event count limit exceeded",
		},
		{
			name: "callbacks",
			raw: func() []byte {
				var callbacks strings.Builder
				for index := 0; index < count; index++ {
					if index > 0 {
						callbacks.WriteByte(',')
					}
					fmt.Fprintf(&callbacks, `"callback-%d":{"{$request.body#/url}":{"post":{"responses":{"200":{"description":"ok"}}}}}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}},"callbacks":{` + callbacks.String() + `}}}}}`)
			},
			want: "inbound event count limit exceeded",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw()
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/count-limit-"+tc.name+".json", raw)
			limits := defaultSourceImportLimits()
			limits.MaxArtifactBytes = 128 << 20
			limits.MaxResolvedDescriptorBytes = 128 << 20
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s count error = %v, want %q", tc.name, err, tc.want)
			}
		})
	}
}

func TestSourceImportValidatesLinkOperationTargets(t *testing.T) {
	t.Parallel()
	valid := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":"listItems","responses":{"200":{"description":"ok","links":{"next":{"operationRef":"#/paths/~1next/get"}}}}}},"/next":{"get":{"operationId":"listNext","responses":{"200":{"description":"ok"}}}}}}`)
	if result := importInlineSourceResult(t, valid, defaultSourceImportLimits()); len(result.Operations) != 2 {
		t.Fatalf("valid link operations = %#v", result.Operations)
	}
	for _, tc := range []struct {
		name string
		link string
		want string
	}{
		{name: "missing target", link: `{}`, want: "exactly one of operationId or operationRef"},
		{name: "conflicting targets", link: `{"operationId":"listItems","operationRef":"#/paths/~1items/get"}`, want: "exactly one of operationId or operationRef"},
		{name: "external operation ref", link: `{"operationRef":"https://provider.invalid/operation"}`, want: "external reference"},
		{name: "unknown operation id", link: `{"operationId":"unknown"}`, want: "does not identify an in-artifact operation"},
		{name: "non-operation pointer", link: `{"operationRef":"#/paths/~1items/get/responses/200"}`, want: "does not resolve to an operation"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":"listItems","responses":{"200":{"description":"ok","links":{"next":` + tc.link + `}}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/link-target.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s link error = %v, want %q", tc.name, err, tc.want)
			}
		})
	}
	duplicate := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/one":{"get":{"operationId":"same","responses":{"200":{"description":"ok"}}}},"/two":{"get":{"operationId":"same","responses":{"200":{"description":"ok"}}}},"/links":{"get":{"responses":{"200":{"description":"ok","links":{"next":{"operationId":"same"}}}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/ambiguous-link-target.json", duplicate)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return duplicate, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "link operationId \"same\" is ambiguous") {
		t.Fatalf("ambiguous link error = %v", err)
	}
}

func TestSourceImportPreservesRequestMediaEncodingAsMergeBlocked(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"headers":{"Trace":{"description":"trace","required":false,"style":"simple","explode":false,"schema":{"type":"string","maxLength":8}}}},"paths":{"/upload":{"post":{"operationId":"upload","requestBody":{"required":false,"content":{"multipart/form-data":{"schema":{"type":"object","additionalProperties":false,"properties":{"payload":{"type":"string","maxLength":16}}},"encoding":{"payload":{"contentType":"image/png","style":"form","explode":false,"allowReserved":false,"headers":{"X-Trace":{"$ref":"#/components/headers/Trace"}}}}}}},"responses":{"201":{"description":"created"}}}}}}`)
	result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
	operation := result.Operations[0]
	if operation.Request.MediaType != "multipart/form-data" || operation.Request.Body == nil || operation.Request.Body.Encoding == nil {
		t.Fatalf("request encoding descriptor = %#v", operation.Request)
	}
	encoding := operation.Request.Body.Encoding.(map[string]any)["payload"].(map[string]any)
	if encoding["contentType"] != "image/png" || encoding["style"] != "form" || encoding["explode"] != false || encoding["allowReserved"] != false {
		t.Fatalf("encoding fields = %#v", encoding)
	}
	header := encoding["headers"].(map[string]any)["X-Trace"].(map[string]any)
	if header["required"] != false || header["explode"] != false || header["style"] != "simple" {
		t.Fatalf("encoding header = %#v", header)
	}
	if !operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 1 || operation.Runtime.Gaps[0].Foundation != "cli-request-encoding-foundation-r1" || operation.Source.Location != `paths["/upload"].post` {
		t.Fatalf("request encoding runtime = %#v", operation)
	}
	encoded, err := marshalSourceImportResult(result)
	if err != nil || !strings.Contains(string(encoded), `"allowReserved": false`) || !strings.Contains(string(encoded), `"merge_blocked": true`) {
		t.Fatalf("request encoding output = %s, %v", encoded, err)
	}
}

func TestSourceImportNormalizesReachableLinkOperationReferences(t *testing.T) {
	t.Parallel()
	encoded := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/users/{id}":{"get":{"operationId":"getUser","parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string","maxLength":8}}],"responses":{"200":{"description":"ok"}}}},"/links":{"get":{"responses":{"200":{"description":"ok","links":{"user":{"operationRef":"#/paths/~1users~1%7Bid%7D/get"}}}}}}}}`)
	if result := importInlineSourceResult(t, encoded, defaultSourceImportLimits()); len(result.Operations) != 2 {
		t.Fatalf("percent-encoded link operations = %#v", result.Operations)
	}
	for _, tc := range []struct {
		name string
		ref  string
		want string
	}{
		{name: "malformed percent escape", ref: "#/paths/%ZZ/get", want: "invalid percent escape"},
		{name: "double encoded fragment", ref: "#%252Fpaths~1items/get", want: "must be a local artifact JSON Pointer"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}},"/links":{"get":{"responses":{"200":{"description":"ok","links":{"next":{"operationRef":"` + tc.ref + `"}}}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/link-normalization.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s link error = %v, want %q", tc.name, err, tc.want)
			}
		})
	}
	shared := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"pathItems":{"shared":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}},"paths":{"/a":{"$ref":"#/components/pathItems/shared"},"/links":{"get":{"responses":{"200":{"description":"ok","links":{"shared":{"operationRef":"#/components/pathItems/shared/get"}}}}}}}}`)
	if result := importInlineSourceResult(t, shared, defaultSourceImportLimits()); len(result.Operations) != 2 {
		t.Fatalf("single reachable component operation = %#v", result.Operations)
	}
	ambiguous := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"pathItems":{"shared":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}},"paths":{"/a":{"$ref":"#/components/pathItems/shared"},"/b":{"$ref":"#/components/pathItems/shared"},"/links":{"get":{"responses":{"200":{"description":"ok","links":{"shared":{"operationRef":"#/components/pathItems/shared/get"}}}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/ambiguous-component-operation-ref.json", ambiguous)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return ambiguous, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "ambiguous across reachable operations") {
		t.Fatalf("ambiguous reachable component operation error = %v", err)
	}
}

func TestSourceImportNormalizesDirectReferenceFragments(t *testing.T) {
	t.Parallel()
	encoded := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"pathItems":{"shared":{"get":{"responses":{"200":{"$ref":"#/paths/~1users~1%7Bid%7D/get/responses/200"}}}}}},"paths":{"/users/{id}":{"get":{"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string","maxLength":8}}],"responses":{"200":{"description":"ok"}}}},"/alias":{"$ref":"#/components/pathItems/%73hared"}}}`)
	if result := importInlineSourceResult(t, encoded, defaultSourceImportLimits()); len(result.Operations) != 2 {
		t.Fatalf("percent-encoded direct references = %#v", result.Operations)
	}
	for _, tc := range []struct {
		name string
		ref  string
		want string
	}{
		{name: "malformed percent escape", ref: "#/paths/%ZZ/get/responses/200", want: "invalid percent escape"},
		{name: "second decode attempt", ref: "#/paths/~1users~1%257Bid%257D/get/responses/200", want: "unresolved reference"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/users/{id}":{"get":{"responses":{"200":{"description":"ok"}}}},"/items":{"get":{"responses":{"200":{"$ref":"` + tc.ref + `"}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/direct-reference-normalization.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s direct reference error = %v, want %q", tc.name, err, tc.want)
			}
		})
	}
	canonical, err := sourceNormalizeLocalReference("#/components/schemas/Root/properties/percent%252F")
	if err != nil || canonical != "#/components/schemas/Root/properties/percent%2F" {
		t.Fatalf("literal percent canonical reference = %q, %v", canonical, err)
	}
	literalPercent := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Root":{"type":"object","additionalProperties":false,"properties":{"percent%2F":{"type":"string","maxLength":8}}}}},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","schema":{"$ref":"#/components/schemas/Root/properties/percent%252F"}}],"responses":{"200":{"description":"ok"}}}}}}`)
	result := importInlineSourceResult(t, literalPercent, defaultSourceImportLimits())
	if len(result.Operations) != 1 || len(result.Operations[0].Request.Query) != 1 {
		t.Fatalf("literal percent source operation = %#v", result.Operations)
	}
	schema, ok := result.Operations[0].Request.Query[0].Schema.(map[string]any)
	if !ok || schema["type"] != "string" || fmt.Sprint(schema["maxLength"]) != "8" {
		t.Fatalf("literal percent resolved schema = %#v", result.Operations[0].Request.Query[0].Schema)
	}
	cycle := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"responses":{"A":{"$ref":"#/components/responses/B"},"B":{"$ref":"#/components/responses/%41"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/A"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/canonical-reference-cycle.json", cycle)
	_, err = importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return cycle, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "reference cycle") {
		t.Fatalf("canonical reference cycle error = %v", err)
	}
	literalPercentCycle := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Root":{"type":"object","additionalProperties":false,"properties":{"percent%2F":{"$ref":"#/components/schemas/Root/properties/percent%252F"}}}}},"paths":{"/items":{"get":{"parameters":[{"name":"filter","in":"query","schema":{"$ref":"#/components/schemas/Root"}}],"responses":{"200":{"description":"ok"}}}}}}`)
	lock = sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/literal-percent-reference-cycle.json", literalPercentCycle)
	_, err = importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return literalPercentCycle, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "reference cycle") {
		t.Fatalf("literal percent reference cycle error = %v", err)
	}
}

func TestSourceImportPreservesFormDefaultsAndRejectsNonFormEncoding(t *testing.T) {
	t.Parallel()
	for _, mediaType := range []string{"multipart/form-data", "application/x-www-form-urlencoded", "Multipart/Related; boundary=source"} {
		mediaType := mediaType
		t.Run(mediaType, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/upload":{"post":{"operationId":"upload","requestBody":{"content":{"` + mediaType + `":{"schema":{"type":"object","additionalProperties":false,"properties":{"payload":{"type":"string","maxLength":8}}}}}},"responses":{"201":{"description":"created"}}}}}}`)
			result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
			operation := result.Operations[0]
			if operation.Request.MediaType != mediaType || operation.Request.Body == nil || operation.Request.Body.Encoding != nil {
				t.Fatalf("form default request = %#v", operation.Request)
			}
			if !operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 1 || operation.Runtime.Gaps[0].Foundation != "cli-request-encoding-foundation-r1" {
				t.Fatalf("form default runtime = %#v", operation.Runtime)
			}
		})
	}
	for _, mediaType := range []string{"application/json", "image/png", "application/octet-stream"} {
		mediaType := mediaType
		t.Run("reject encoding "+mediaType, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/upload":{"post":{"requestBody":{"content":{"` + mediaType + `":{"schema":{"type":"object","additionalProperties":false,"properties":{"payload":{"type":"string","maxLength":8}}},"encoding":{"payload":{}}}}},"responses":{"201":{"description":"created"}}}}}}`)
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/non-form-encoding.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), "unsupported encoding on request content media") {
				t.Fatalf("non-form encoding error = %v", err)
			}
		})
	}
}

func TestSourceImportReservesRequestMediaExpansionBeforeCloning(t *testing.T) {
	large := strings.Repeat("x", 900<<10)
	limits := defaultSourceImportLimits()
	limits.MaxArtifactBytes = 4 << 20
	limits.MaxResolvedDescriptorBytes = 8 << 20
	cases := []struct {
		name string
		raw  func() []byte
	}{
		{
			name: "parameter content examples",
			raw: func() []byte {
				var parameters strings.Builder
				for index := 0; index < 16; index++ {
					if index > 0 {
						parameters.WriteByte(',')
					}
					fmt.Fprintf(&parameters, `{"name":"p%d","in":"query","content":{"application/json":{"schema":{"type":"string","maxLength":1},"examples":{"large":{"$ref":"#/components/examples/Large"}}}}}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"examples":{"Large":{"value":"` + large + `"}}},"paths":{"/items":{"get":{"parameters":[` + parameters.String() + `],"responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
		{
			name: "request encoding headers",
			raw: func() []byte {
				var encodings strings.Builder
				for index := 0; index < 16; index++ {
					if index > 0 {
						encodings.WriteByte(',')
					}
					fmt.Fprintf(&encodings, `"part%d":{"headers":{"X-Large":{"$ref":"#/components/headers/Large"}}}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"examples":{"Large":{"value":"` + large + `"}},"headers":{"Large":{"content":{"application/json":{"schema":{"type":"string","maxLength":1},"examples":{"large":{"$ref":"#/components/examples/Large"}}}}}}},"paths":{"/upload":{"post":{"requestBody":{"content":{"multipart/form-data":{"schema":{"type":"object","additionalProperties":false,"properties":{"part":{"type":"string","maxLength":1}}},"encoding":{` + encodings.String() + `}}}},"responses":{"201":{"description":"created"}}}}}}`)
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw()
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/request-media-expansion.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining request media") {
				t.Fatalf("request media expansion error = %v", err)
			}
		})
	}
}

func TestSourceImportReservesInboundExpansionBeforeEventConstruction(t *testing.T) {
	large := strings.Repeat("x", 900<<10)
	limits := defaultSourceImportLimits()
	limits.MaxArtifactBytes = 4 << 20
	limits.MaxResolvedDescriptorBytes = 8 << 20
	cases := []struct {
		name string
		raw  func() []byte
	}{
		{
			name: "webhook path item",
			raw: func() []byte {
				var webhooks strings.Builder
				for index := 0; index < 16; index++ {
					if index > 0 {
						webhooks.WriteByte(',')
					}
					fmt.Fprintf(&webhooks, `"event-%d":{"$ref":"#/components/pathItems/Shared"}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"pathItems":{"Shared":{"x-large":"` + large + `","post":{"responses":{"200":{"description":"ok"}}}}}},"webhooks":{` + webhooks.String() + `},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
		{
			name: "callback extension",
			raw: func() []byte {
				var callbacks strings.Builder
				for index := 0; index < 16; index++ {
					if index > 0 {
						callbacks.WriteByte(',')
					}
					fmt.Fprintf(&callbacks, `"callback-%d":{"$ref":"#/components/callbacks/Shared"}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"callbacks":{"Shared":{"x-large":"` + large + `","{$request.body#/hook}":{"post":{"responses":{"200":{"description":"ok"}}}}}}},"paths":{"/items":{"get":{"callbacks":{` + callbacks.String() + `},"responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw()
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inbound-expansion.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining inbound event") {
				t.Fatalf("inbound expansion error = %v", err)
			}
		})
	}
}

func TestSourceImportReservesCallbackReferenceChainsBeforeCloning(t *testing.T) {
	large := strings.Repeat("x", 4<<20)
	raw := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"callbacks":{"AOuter":{"$ref":"#/components/callbacks/ZOne"},"ZOne":{"$ref":"#/components/callbacks/ZTwo"},"ZTwo":{"$ref":"#/components/callbacks/ZThree"},"ZThree":{"$ref":"#/components/callbacks/ZZBase"},"ZZBase":{"x-large":"` + large + `","{$request.body#/hook}":{"post":{"responses":{"200":{"description":"ok"}}}}}}},"paths":{"/items":{"get":{"callbacks":{"notify":{"$ref":"#/components/callbacks/AOuter"}},"responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/callback-reference-chain.json", raw)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining reference target") {
		t.Fatalf("callback reference chain error = %v", err)
	}
}

func TestSourceImportReservesNonResponseReferenceExpansionBeforeCloning(t *testing.T) {
	t.Parallel()
	const aliasCount = 16
	large := strings.Repeat("x", 900<<10)
	limits := defaultSourceImportLimits()
	limits.MaxArtifactBytes = 4 << 20
	limits.MaxResolvedDescriptorBytes = 8 << 20
	aliases := func(reference string) string {
		var entries strings.Builder
		for index := 0; index < aliasCount; index++ {
			if index > 0 {
				entries.WriteByte(',')
			}
			fmt.Fprintf(&entries, `"Alias%02d":{"$ref":"%s"}`, index, reference)
		}
		return entries.String()
	}
	cases := []struct {
		name string
		raw  func() []byte
	}{
		{
			name: "ordinary path item references",
			raw: func() []byte {
				var paths strings.Builder
				for index := 0; index < aliasCount; index++ {
					if index > 0 {
						paths.WriteByte(',')
					}
					fmt.Fprintf(&paths, `"/items/%d":{"$ref":"#/components/pathItems/Shared"}`, index)
				}
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"pathItems":{"Shared":{"x-large":"` + large + `","get":{"responses":{"200":{"description":"ok"}}}}}},"paths":{` + paths.String() + `}}`)
			},
		},
		{
			name: "example aliases",
			raw: func() []byte {
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"examples":{"Large":{"value":"` + large + `"},` + aliases("#/components/examples/Large") + `}},"paths":{"/items":{"get":{"operationId":"items","responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
		{
			name: "header aliases",
			raw: func() []byte {
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"headers":{"Large":{"description":"` + large + `","schema":{"type":"string","maxLength":1}},` + aliases("#/components/headers/Large") + `}},"paths":{"/items":{"get":{"operationId":"items","responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
		{
			name: "link aliases",
			raw: func() []byte {
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"links":{"Large":{"operationId":"items","description":"` + large + `"},` + aliases("#/components/links/Large") + `}},"paths":{"/items":{"get":{"operationId":"items","responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
		{
			name: "security aliases",
			raw: func() []byte {
				return []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"securitySchemes":{"Large":{"type":"http","description":"` + large + `"},` + aliases("#/components/securitySchemes/Large") + `}},"paths":{"/items":{"get":{"operationId":"items","responses":{"200":{"description":"ok"}}}}}}`)
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw()
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/non-response-reference-expansion.json", raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), "resolved descriptor byte limit exceeded while retaining reference target") {
				t.Fatalf("non-response reference expansion error = %v", err)
			}
		})
	}
}

func TestSourceImportBoundsExtensionKeysBeforeSorting(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{
			name: "root extensions",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"x-one":{},"x-two":{},"x-three":{},"x-four":{},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
		},
		{
			name: "components extensions",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"x-one":{},"x-two":{},"x-three":{},"x-four":{}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
		},
		{
			name: "paths extensions",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"x-one":{},"x-two":{},"x-three":{},"x-four":{},"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`),
		},
		{
			name: "path item extensions",
			raw:  []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/items":{"x-one":{},"x-two":{},"x-three":{},"x-four":{},"get":{"responses":{"200":{"description":"ok"}}}}}}`),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			limits := defaultSourceImportLimits()
			limits.MaxSchemaNodes = 3
			lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/extension-position-limit.json", tc.raw)
			_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return tc.raw, nil }), limits)
			if err == nil || !strings.Contains(err.Error(), "source grammar position limit exceeded") {
				t.Fatalf("extension position error = %v", err)
			}
		})
	}
}

func importInlineSourceResult(t *testing.T, raw []byte, limits sourceImportLimits) sourceImportResult {
	t.Helper()
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inline-openapi.json", raw)
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
	if err != nil {
		t.Fatalf("import inline source: %v", err)
	}
	return result
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
	return sourceImportLock{SchemaVersion: 1, Connector: connector, Rest: sourceImportArtifact{SourceURL: sourceURL, SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(raw))}}
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
