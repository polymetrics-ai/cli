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
