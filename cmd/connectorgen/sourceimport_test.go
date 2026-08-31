package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors/engine"
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

func TestSourceImportAcceptsStringScalarUnionPathWireContract(t *testing.T) {
	t.Parallel()
	artifact := []byte(`{
  "openapi":"3.0.3",
  "info":{"title":"fixture","version":"1"},
  "paths":{
    "/workflows/{workflow_id}":{
      "get":{
        "operationId":"workflows/get",
        "parameters":[{
          "name":"workflow_id","in":"path","required":true,
          "schema":{"oneOf":[{"type":"integer"},{"type":"string"}]}
        }],
        "responses":{"200":{"description":"ok"}}
      }
    }
  }
}`)
	result := importInlineSourceResult(t, artifact, defaultSourceImportLimits())
	if len(result.Operations) != 1 {
		t.Fatalf("operations = %d, want 1", len(result.Operations))
	}
	operation := result.Operations[0]
	if len(operation.Runtime.Gaps) != 0 || operation.Runtime.MergeBlocked {
		t.Fatalf("string-or-integer path union must remain a closed textual wire contract, got runtime=%+v", operation.Runtime)
	}
	parameter := operation.Request.Path[0]
	schema, ok := parameter.Schema.(map[string]any)
	if !ok {
		t.Fatalf("path schema = %#v, want preserved source object", parameter.Schema)
	}
	arms, _ := schema["oneOf"].([]any)
	if len(arms) != 2 {
		t.Fatalf("path union arms = %#v, want exact two-arm source contract", schema)
	}
}

func TestSourceImportVersion3RepresentsGongWorkspaceQueryWithPMExecutionEnvelope(t *testing.T) {
	t.Parallel()
	artifact := []byte(`{
  "openapi":"3.0.3",
  "info":{"title":"Gong API","version":"V2"},
  "paths":{
    "/v2/all-permission-profiles":{
      "get":{
        "operationId":"shared",
        "parameters":[{
          "name":"workspaceId",
          "in":"query",
          "required":true,
          "schema":{"type":"string"}
        }],
        "responses":{"200":{"description":"ok"}}
      }
    }
  }
}`)
	document := sourceImportV3FixtureDocument{
		ID:       "gong-v2",
		Path:     "/v2/all-permission-profiles",
		Artifact: artifact,
	}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "gong", []sourceImportV3FixtureDocument{document}), "gong")
	if err != nil {
		t.Fatalf("parse Gong-shaped v3 source lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import Gong-shaped source: %v", err)
	}
	if len(result.Operations) != 1 {
		t.Fatalf("operations = %d, want 1", len(result.Operations))
	}
	operation := result.Operations[0]
	if operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 0 {
		t.Fatalf("ordinary unbounded Gong query must be represented, runtime = %+v", operation.Runtime)
	}
	if len(operation.Request.Query) != 1 || operation.Request.Query[0].Name != "workspaceId" || !operation.Request.Query[0].Required {
		t.Fatalf("Gong query descriptor = %#v", operation.Request.Query)
	}
	parameterRaw, err := json.Marshal(operation.Request.Query[0])
	if err != nil {
		t.Fatalf("marshal Gong query descriptor: %v", err)
	}
	var parameter map[string]any
	if err := json.Unmarshal(parameterRaw, &parameter); err != nil {
		t.Fatalf("decode Gong query descriptor: %v", err)
	}
	schema, _ := parameter["schema"].(map[string]any)
	if !reflect.DeepEqual(schema, map[string]any{"type": "string"}) {
		t.Fatalf("provider schema = %#v, want exact source schema without synthetic maxLength", schema)
	}
	execution, _ := parameter["execution_envelope"].(map[string]any)
	if execution["policy_version"] != "pm-request-contract-bounds-v1" || execution["origin"] != "pm_policy" || execution["source_location"] != `request.query["workspaceId"]` {
		t.Fatalf("execution envelope provenance = %#v", execution)
	}
	limits, _ := execution["limits"].([]any)
	if len(limits) != 1 {
		t.Fatalf("execution limits = %#v, want one encoded-byte limit", limits)
	}
	encodedBytes, _ := limits[0].(map[string]any)
	if encodedBytes["kind"] != "wire_value" || encodedBytes["unit"] != "encoded_bytes" || encodedBytes["default"] != float64(4096) || encodedBytes["hard_ceiling"] != float64(65536) || encodedBytes["effective"] != float64(4096) {
		t.Fatalf("encoded-byte execution limit = %#v", encodedBytes)
	}
	result.Operations[0].Request.Query[0].ExecutionEnvelope = nil
	if err := validateSourceProjectionExecutionEnvelopes(result); err == nil || !strings.Contains(err.Error(), "requires a PM execution envelope") {
		t.Fatalf("projection accepted missing Gong execution envelope: %v", err)
	}
}

func TestSourceRequestSchemaDispositionSeparatesPolicyBoundsFromMalformedInput(t *testing.T) {
	t.Parallel()
	form := sourceDocumentForm{Family: "openapi", Version: "3.0.3"}
	for _, test := range []struct {
		name   string
		schema map[string]any
	}{
		{name: "string", schema: map[string]any{"type": "string"}},
		{name: "number", schema: map[string]any{"type": "number"}},
		{name: "array", schema: map[string]any{"type": "array", "items": map[string]any{"type": "boolean"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			strict := defaultSourceImportLimits()
			err := validateBoundedRequestSchema(test.schema, form, strict, 0)
			if got := sourceRequestSchemaDispositionOf(err); got != sourceRequestRepresentedWithPolicyBound {
				t.Fatalf("strict disposition = %q, error %v; want %q", got, err, sourceRequestRepresentedWithPolicyBound)
			}
			policy := strict
			policy.UseExecutionEnvelopes = true
			if err := validateBoundedRequestSchema(test.schema, form, policy, 0); err != nil {
				t.Fatalf("policy-bounded schema: %v", err)
			}
		})
	}

	policy := defaultSourceImportLimits()
	policy.UseExecutionEnvelopes = true
	err := validateBoundedRequestSchema(map[string]any{"type": "string", "minLength": json.Number("2"), "maxLength": json.Number("1")}, form, policy, 0)
	if err == nil || sourceRequestSchemaDispositionOf(err) != sourceRequestMalformedSourceGap || !strings.Contains(err.Error(), "contradictory") {
		t.Fatalf("contradictory schema disposition/error = %q/%v, want malformed source", sourceRequestSchemaDispositionOf(err), err)
	}
}

func TestSourceRequestGapsAdmitsClosedUnicodeScalarP5PatternButRetainsDynamicRootGap(t *testing.T) {
	t.Parallel()
	request := sourceRequestDescriptor{
		MediaType: "application/json",
		Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
			"type": "object", "properties": map[string]any{
				"origin": map[string]any{
					"type": "string", "nullable": true,
					"pattern": `^(?:[\uD800-\uDBFF][\uDC00-\uDFFF]|[^\n\uD800-\uDFFF]){1,255}$`,
				},
			},
		}},
	}
	gaps := sourceRequestGaps(request, sourceDocumentForm{Family: "openapi", Version: "3.0.3"}, defaultSourceImportLimits(), http.MethodPost)
	hasDynamicRootGap := false
	for _, gap := range gaps {
		if gap.Foundation == "cli-request-schema-foundation-r1" && gap.Location == "request body property origin" {
			t.Fatalf("request gaps = %+v, want exact Unicode-scalar field admitted", gaps)
		}
		if gap.Foundation == "cli-request-schema-foundation-r1" && gap.Location == "request body" && strings.Contains(gap.Reason, "dynamic additionalProperties") {
			hasDynamicRootGap = true
		}
	}
	if !hasDynamicRootGap {
		t.Fatalf("request gaps = %+v, want retained dynamic-root disposition", gaps)
	}
}

func TestSourceParameterExecutionEnvelopeUsesTighterProviderDerivedByteCap(t *testing.T) {
	limits := defaultSourceImportLimits()
	limits.UseExecutionEnvelopes = true
	parameter := sourceParameterValue{
		Name: "slug",
		In:   "path",
		Schema: map[string]any{
			"type":      "string",
			"maxLength": json.Number("8"),
		},
	}
	envelope := sourceParameterExecutionEnvelopeFor(parameter, limits)
	if envelope == nil || envelope.Origin != "provider_and_pm_policy" || len(envelope.Limits) != 1 {
		t.Fatalf("provider-bounded envelope = %+v", envelope)
	}
	limit := envelope.Limits[0]
	if limit.Default != engine.DefaultOperationParameterMaxBytes || limit.Effective != 8*utf8.UTFMax || limit.HardCeiling != engine.MaxOperationParameterMaxBytes || limit.Unit != "encoded_bytes" {
		t.Fatalf("provider-bounded execution limit = %+v", limit)
	}
	parameter.Schema.(map[string]any)["maxLength"] = json.Number("2000")
	envelope = sourceParameterExecutionEnvelopeFor(parameter, limits)
	if envelope.Limits[0].Effective != engine.DefaultOperationParameterMaxBytes {
		t.Fatalf("PM default did not win over larger provider-derived cap: %+v", envelope.Limits[0])
	}
}

func TestSourceImportVersion3KeepsUnboundedHeaderAsMergeBlockingGap(t *testing.T) {
	t.Parallel()
	artifact := []byte(`{
  "openapi":"3.0.3",
  "info":{"title":"header fixture","version":"1"},
  "paths":{"/headers":{"get":{
    "operationId":"shared",
    "parameters":[{"name":"X-Request-Context","in":"header","schema":{"type":"string"}}],
    "responses":{"200":{"description":"ok"}}
  }}}
}`)
	document := sourceImportV3FixtureDocument{ID: "header", Path: "/headers", Artifact: artifact}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse v3 header lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return artifact, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 header source: %v", err)
	}
	operation := result.Operations[0]
	if !operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 1 || operation.Runtime.Gaps[0].Foundation != "cli-request-schema-foundation-r1" || !strings.Contains(operation.Runtime.Gaps[0].Reason, "compatibility-censused PM byte envelope") {
		t.Fatalf("unbounded header runtime gap = %+v", operation.Runtime)
	}
	if operation.Request.Header[0].ExecutionEnvelope != nil {
		t.Fatalf("uncensused header unexpectedly received an execution envelope: %+v", operation.Request.Header[0].ExecutionEnvelope)
	}
}

func TestSourceProjectionKeepsNumericHeaderWithoutTextualByteBoundMergeBlocked(t *testing.T) {
	limits := defaultSourceImportLimits()
	limits.UseExecutionEnvelopes = true
	schema := map[string]any{
		"type":    "integer",
		"minimum": json.Number("0"),
		"maximum": json.Number("100"),
	}
	err := sourceProjectionOperationParameterGap(schema, sourceDocumentForm{Family: "openapi", Version: "3.0.3"}, limits, "header", http.MethodGet)
	if err == nil || !strings.Contains(err.Error(), "compatibility-censused PM byte envelope") {
		t.Fatalf("numeric header without a textual byte bound gap = %v", err)
	}
}

func TestSourceImportVersion3RepresentsCommonBodyBoundsWithSeparateEnvelope(t *testing.T) {
	t.Parallel()
	artifact := []byte(`{
  "openapi":"3.0.3",
  "info":{"title":"body fixture","version":"1"},
  "paths":{"/records":{"post":{
    "operationId":"createRecord",
    "requestBody":{"required":true,"content":{"application/json":{"schema":{
      "type":"object",
      "additionalProperties":false,
      "properties":{
        "name":{"type":"string"},
        "weight":{"type":"number"},
        "enabled":{"type":"boolean"},
        "tags":{"type":"array","items":{"type":"string"}}
      },
      "required":["name","weight","tags"]
    }}}},
    "responses":{"200":{"description":"ok"}}
  }}}
}`)
	document := sourceImportV3FixtureDocument{ID: "body", Path: "/records", Method: "POST", OperationID: "createRecord", Artifact: artifact}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse v3 body lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return artifact, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 body source: %v", err)
	}
	operation := result.Operations[0]
	if operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 0 || operation.Request.Body == nil {
		t.Fatalf("common body bounds were not represented: request/runtime = %+v/%+v", operation.Request, operation.Runtime)
	}
	body, _ := operation.Request.Body.Schema.(map[string]any)
	properties, _ := body["properties"].(map[string]any)
	name, _ := properties["name"].(map[string]any)
	weight, _ := properties["weight"].(map[string]any)
	tags, _ := properties["tags"].(map[string]any)
	if _, exists := name["maxLength"]; exists {
		t.Fatalf("source name schema received synthetic maxLength: %#v", name)
	}
	if _, exists := weight["minimum"]; exists {
		t.Fatalf("source number schema received synthetic minimum: %#v", weight)
	}
	if _, exists := weight["maximum"]; exists {
		t.Fatalf("source number schema received synthetic maximum: %#v", weight)
	}
	if _, exists := tags["maxItems"]; exists {
		t.Fatalf("source array schema received synthetic maxItems: %#v", tags)
	}
	envelope := operation.Request.Body.ExecutionEnvelope
	if envelope == nil || envelope.PolicyVersion != engine.OperationParameterExecutionPolicyVersion || len(envelope.Limits) != 5 {
		t.Fatalf("body execution envelope = %+v", envelope)
	}
	envelope.Limits[0].Effective--
	policy := defaultSourceImportLimits()
	policy.UseExecutionEnvelopes = true
	if err := validateSourceRequestBodyExecutionEnvelope(envelope, operation.Request.MediaType, policy); err == nil || !strings.Contains(err.Error(), "valid PM execution envelope") {
		t.Fatalf("projection accepted altered body execution envelope: %v", err)
	}
}

func TestSourceImportRetainsRecursiveSchemaReferencesAsSourceBoundGaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		components  string
		responseRef string
		wantPointer string
		wantSchema  string
		wantGap     bool
	}{
		{
			name:        "direct self reference",
			components:  `"Folder":{"type":"object","additionalProperties":false,"properties":{"children":{"type":"array","items":{"$ref":"#/components/schemas/Folder"}}}}`,
			responseRef: "#/components/schemas/Folder",
			wantPointer: "#/components/schemas/Folder",
			wantSchema:  "Folder",
			wantGap:     true,
		},
		{
			name:        "mutually recursive schemas",
			components:  `"Folder":{"type":"object","additionalProperties":false,"properties":{"parent":{"$ref":"#/components/schemas/Parent"}}},"Parent":{"type":"object","additionalProperties":false,"properties":{"folder":{"$ref":"#/components/schemas/Folder"}}}`,
			responseRef: "#/components/schemas/Folder",
			wantPointer: "#/components/schemas/Folder",
			wantSchema:  "Folder",
			wantGap:     true,
		},
		{
			name:        "deeply nested cycle",
			components:  `"Folder":{"type":"object","additionalProperties":false,"properties":{"metadata":{"type":"object","additionalProperties":false,"properties":{"tree":{"type":"object","additionalProperties":false,"properties":{"child":{"$ref":"#/components/schemas/Folder"}}}}}}}`,
			responseRef: "#/components/schemas/Folder",
			wantPointer: "#/components/schemas/Folder",
			wantSchema:  "Folder",
			wantGap:     true,
		},
		{
			name:        "non cyclic schema",
			components:  `"Folder":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string","maxLength":64}}}`,
			responseRef: "#/components/schemas/Folder",
			wantGap:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.1.0","info":{"title":"recursive fixture","version":"1"},"components":{"schemas":{` + tt.components + `}},"paths":{"/folders":{"get":{"operationId":"folders/get","responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"$ref":"` + tt.responseRef + `"}}}}}}}}}`)
			result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
			if len(result.Operations) != 1 {
				t.Fatalf("operations = %d, want retained operation", len(result.Operations))
			}
			operation := result.Operations[0]
			if operation.SourceID != "folders/get" || operation.ProviderOperationID != "folders/get" || operation.Source.Location != `paths["/folders"].get` || operation.Source.URL != "https://fixtures.polymetrics.invalid/inline-openapi.json" {
				t.Fatalf("operation source trace = %#v", operation.Source)
			}

			var cycleGap *sourceContractGap
			for index := range operation.Runtime.Gaps {
				gap := &operation.Runtime.Gaps[index]
				if gap.Foundation == "cli-recursive-schema-foundation-r1" {
					cycleGap = gap
					break
				}
			}
			if !tt.wantGap {
				if cycleGap != nil || operation.Runtime.MergeBlocked {
					t.Fatalf("non-cyclic runtime gaps = %#v", operation.Runtime)
				}
				return
			}
			if cycleGap == nil || !operation.Runtime.MergeBlocked {
				t.Fatalf("recursive schema runtime gap = %#v", operation.Runtime)
			}
			if !strings.Contains(cycleGap.Location, `response 200`) || !strings.Contains(cycleGap.Reason, tt.wantPointer) || !strings.Contains(cycleGap.Reason, tt.wantSchema) {
				t.Fatalf("cycle gap = %#v, want response location and schema pointer %q", cycleGap, tt.wantPointer)
			}

			response := descriptorResponse(t, operation, "200")
			encoded, err := json.Marshal(response.Declaration)
			if err != nil {
				t.Fatalf("marshal retained response declaration: %v", err)
			}
			if !strings.Contains(string(encoded), `"$ref":"`+tt.wantPointer+`"`) {
				t.Fatalf("recursive schema was flattened or truncated: %s", encoded)
			}
		})
	}
}

func TestSourceImport_CheckedInGitHubLockCoversRESTAndGraphQL(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read checked-in GitHub source lock: %v", err)
	}
	lock, err := parseSourceImportLock(raw, "github")
	if err != nil {
		t.Fatalf("parse checked-in GitHub source lock: %v", err)
	}
	if got := len(lock.Rest.Operations) + len(lock.GraphQL.QueryFields) + len(lock.GraphQL.MutationFields); got != 1525 {
		t.Fatalf("checked-in source identities = %d, want 1525", got)
	}
	if len(lock.Rest.Operations) != 1220 || len(lock.GraphQL.QueryFields) != 31 || len(lock.GraphQL.MutationFields) != 274 {
		t.Fatalf("checked-in source identity split = REST %d query %d mutation %d", len(lock.Rest.Operations), len(lock.GraphQL.QueryFields), len(lock.GraphQL.MutationFields))
	}
	if lock.Rest.Operations[0].SourceLocation != `paths["/"].get` {
		t.Fatalf("first REST source location = %q", lock.Rest.Operations[0].SourceLocation)
	}
	if lock.GraphQL.QueryFields[0].Line <= 0 || lock.GraphQL.QueryFields[0].Root != "Query" {
		t.Fatalf("first GraphQL source location = %#v", lock.GraphQL.QueryFields[0])
	}
	if lock.Rest.Bytes != 12920264 || lock.GraphQL.Bytes != 1551372 || len(lock.Rest.SHA256) != 64 || len(lock.GraphQL.SHA256) != 64 {
		t.Fatalf("checked-in source byte/digest pins = REST %#v GraphQL %#v", lock.Rest, lock.GraphQL.sourceImportArtifact)
	}
}

func TestSourceImport_CheckedInGitHubProjectionDigestIsCanonical(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lock sourceImportLock
	if err := decodeSourceStrictJSON(raw, &lock); err != nil {
		t.Fatal(err)
	}
	projection, err := canonicalSourceGraphQLProjection(lock.GraphQL)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(projection)
	wantDigest := hex.EncodeToString(digest[:])
	if lock.GraphQL.ProjectionSHA256 != wantDigest || lock.GraphQL.ProjectionBytes != int64(len(projection)) {
		t.Fatalf("checked-in GraphQL projection pin = %s/%d, want %s/%d", lock.GraphQL.ProjectionSHA256, lock.GraphQL.ProjectionBytes, wantDigest, len(projection))
	}
}

func TestSourceImportVersion2UsesEmbeddedLockedGraphQLProjection(t *testing.T) {
	t.Parallel()
	rest := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":"items/list","responses":{"200":{"description":"ok"}}}}}}`)
	graphql := []byte("type Query { viewer: User }\ntype User { id: ID! }\n")
	graphqlDigest := sha256.Sum256(graphql)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/rest.json", rest)
	lock.SchemaVersion = 2
	lock.GraphQL = sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{SourceURL: "https://fixtures.polymetrics.invalid/unversioned.graphql", SHA256: hex.EncodeToString(graphqlDigest[:]), Bytes: int64(len(graphql))},
		QueryFields:          []sourceGraphQLField{{Root: "Query", Name: "viewer", Line: 7, Signature: "viewer: User", ReturnType: sourceGraphQLTypeRef{Kind: "OBJECT", Name: "User"}}},
		TypeSystem:           sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{}, Unions: []sourceGraphQLNamedType{}},
	}
	lock = sourceImportTestWithProjectionDigest(t, lock)
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		switch sourceURL {
		case lock.Rest.SourceURL:
			return rest, nil
		case lock.GraphQL.SourceURL:
			return graphql, nil
		default:
			t.Fatalf("version 2 import fetched unexpected URL %q", sourceURL)
		}
		return nil, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("version 2 import: %v", err)
	}
	if len(result.Operations) != 2 || result.Operations[1].GraphQL == nil || result.Operations[1].Source.SHA256 != lock.GraphQL.ProjectionSHA256 || result.Operations[1].Source.Bytes != lock.GraphQL.ProjectionBytes {
		t.Fatalf("embedded GraphQL projection = %#v", result.Operations)
	}
}

// A version-two lock embeds a generated GraphQL projection, but it still pins
// the provider schema's raw bytes. The retained-artifact contract must verify
// that pin as well: otherwise a checked-in schema mirror can silently be
// missing while source-import claims the whole lock was verified.
func TestSourceImportVersion2RequiresRawGraphQLArtifactAlongsideProjection(t *testing.T) {
	t.Parallel()
	rest := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":"items/list","responses":{"200":{"description":"ok"}}}}}}`)
	graphql := []byte("type Query { viewer: User }\ntype User { id: ID! }\n")
	graphqlDigest := sha256.Sum256(graphql)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/rest.json", rest)
	lock.SchemaVersion = 2
	lock.GraphQL = sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{SourceURL: "https://fixtures.polymetrics.invalid/schema.graphql", SHA256: hex.EncodeToString(graphqlDigest[:]), Bytes: int64(len(graphql))},
		QueryFields:          []sourceGraphQLField{{Root: "Query", Name: "viewer", Line: 1, Signature: "viewer: User", ReturnType: sourceGraphQLTypeRef{Kind: "OBJECT", Name: "User"}}},
		TypeSystem:           sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{}, Unions: []sourceGraphQLNamedType{}},
	}
	lock = sourceImportTestWithProjectionDigest(t, lock)
	_, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		if sourceURL == lock.Rest.SourceURL {
			return rest, nil
		}
		if sourceURL == lock.GraphQL.SourceURL {
			return nil, fmt.Errorf("retained GraphQL artifact is absent")
		}
		t.Fatalf("unexpected source URL %q", sourceURL)
		return nil, nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "fetch locked GraphQL source artifact") {
		t.Fatalf("missing raw GraphQL artifact error = %v, want retained-artifact refusal", err)
	}
}

func TestSourceImportVersion2RejectsEmbeddedGraphQLProjectionDigestDrift(t *testing.T) {
	rest := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"paths":{"/items":{"get":{"operationId":"items/list","responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/rest.json", rest)
	lock.SchemaVersion = 2
	lock.GraphQL = sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{SourceURL: "https://fixtures.polymetrics.invalid/unversioned.graphql", SHA256: strings.Repeat("a", 64), Bytes: 128},
		QueryFields:          []sourceGraphQLField{{Root: "Query", Name: "viewer", Line: 7, Signature: "viewer: User", ReturnType: sourceGraphQLTypeRef{Kind: "OBJECT", Name: "User"}}},
		TypeSystem:           sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{}, Unions: []sourceGraphQLNamedType{}},
	}
	lock = sourceImportTestWithProjectionDigest(t, lock)
	lock.GraphQL.QueryFields[0].Signature = "viewer: Organization"
	_, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return rest, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "GraphQL projection") {
		t.Fatalf("embedded projection drift error = %v, want authenticated projection refusal", err)
	}
}

func sourceImportTestWithProjectionDigest(t *testing.T, lock sourceImportLock) sourceImportLock {
	t.Helper()
	projection, err := json.Marshal(map[string]any{
		"query_fields":    lock.GraphQL.QueryFields,
		"mutation_fields": lock.GraphQL.MutationFields,
		"type_system":     lock.GraphQL.TypeSystem,
	})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(projection)
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	graphql := document["graphql"].(map[string]any)
	graphql["projection_sha256"] = hex.EncodeToString(digest[:])
	graphql["projection_bytes"] = len(projection)
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatal(err)
	}
	return lock
}

func TestSourceImport_RejectsUnknownSectionAndIndependentIndexOverflow(t *testing.T) {
	t.Parallel()
	fixture := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json"))
	var document map[string]any
	if err := json.Unmarshal(fixture, &document); err != nil {
		t.Fatalf("decode source lock fixture: %v", err)
	}
	document["unexpected"] = true
	unknown, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("encode unknown source lock: %v", err)
	}
	if _, err := parseSourceImportLock(unknown, "alpha"); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown source-lock member error = %v", err)
	}

	artifact := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", artifact)
	limits := defaultSourceImportLimits()
	limits.MaxIndexBytes = 1
	_, err = importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return artifact, nil }), limits)
	if err == nil || !strings.Contains(err.Error(), "source grammar position byte limit") {
		t.Fatalf("independent index overflow error = %v", err)
	}

	limits = defaultSourceImportLimits()
	limits.MaxResolvedDescriptorBytes = 1 << 20
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return artifact, nil }), limits); err != nil {
		t.Fatalf("descriptor ceiling incorrectly constrained grammar index: %v", err)
	}
}

func TestSourceImportImportsLockedRESTAndGraphQLIdentities(t *testing.T) {
	t.Parallel()
	restRaw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	graphqlRaw := []byte("type Query { viewer: String }\ntype Mutation { updateWidget(input: String!): Boolean }\n")
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.yaml", restRaw)
	lock.SchemaVersion = 2
	lock.Rest.Operations = []sourceImportRESTOperation{
		{ID: "alpha.rest.getWidget", Protocol: "rest", Method: "GET", Path: "/widgets/{widget_id}", OperationID: "getWidget", SourceLocation: `paths["/widgets/{widget_id}"].get`},
		{ID: "alpha.rest.post_widgets", Protocol: "rest", Method: "POST", Path: "/widgets", SourceLocation: `paths["/widgets"].post`},
	}
	graphqlDigest := sha256.Sum256(graphqlRaw)
	lock.GraphQL = sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{SourceURL: "https://fixtures.polymetrics.invalid/schema.graphql", SHA256: hex.EncodeToString(graphqlDigest[:]), Bytes: int64(len(graphqlRaw))},
		QueryFields:          []sourceGraphQLField{{Root: "Query", Name: "viewer", Line: 1, Signature: "viewer: String", Arguments: []sourceGraphQLArgument{}, ReturnType: sourceGraphQLTypeRef{Kind: "named", Name: "String"}}},
		MutationFields:       []sourceGraphQLField{{Root: "Mutation", Name: "updateWidget", Line: 2, Signature: "updateWidget(input: String!): Boolean", Arguments: []sourceGraphQLArgument{{Name: "input", Type: sourceGraphQLTypeRef{Kind: "named", Name: "String", NonNull: true}}}, ReturnType: sourceGraphQLTypeRef{Kind: "named", Name: "Boolean"}}},
		TypeSystem:           sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{"Boolean", "String"}, Unions: []sourceGraphQLNamedType{}},
	}
	lock = sourceImportTestWithProjectionDigest(t, lock)
	lock.Counts = sourceImportCounts{REST: 2, GraphQLQuery: 1, GraphQLMutation: 1, Total: 4}
	fetcher := sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		switch sourceURL {
		case lock.Rest.SourceURL:
			return restRaw, nil
		case lock.GraphQL.SourceURL:
			return graphqlRaw, nil
		default:
			return nil, fmt.Errorf("unexpected source URL %q", sourceURL)
		}
	})
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import combined source lock: %v", err)
	}
	if len(result.Operations) != 4 || len(result.GraphQLSchemas) != 1 {
		t.Fatalf("combined descriptor counts = operations %d schemas %d", len(result.Operations), len(result.GraphQLSchemas))
	}
	byID := make(map[string]sourceOperationDescriptor, len(result.Operations))
	for _, operation := range result.Operations {
		byID[operation.SourceID] = operation
	}
	query := byID["alpha.graphql.query.viewer"]
	mutation := byID["alpha.graphql.mutation.updateWidget"]
	if query.Protocol != "graphql" || query.Source.Location != `graphql.query_fields["viewer"]@line:1` || query.GraphQL == nil || query.GraphQL.ReturnType.Name != "String" {
		t.Fatalf("GraphQL query descriptor = %#v", query)
	}
	if mutation.GraphQL == nil || len(mutation.GraphQL.Arguments) != 1 || !mutation.GraphQL.Arguments[0].Type.NonNull {
		t.Fatalf("GraphQL mutation descriptor = %#v", mutation)
	}
}

func TestSourceImportRejectsUnsafeOrRetainsMalformedSourceForms(t *testing.T) {
	t.Parallel()
	baseLimits := defaultSourceImportLimits()
	cases := []struct {
		name     string
		artifact string
		want     string
		wantGap  string
		limits   sourceImportLimits
	}{
		{name: "external reference", artifact: "external-ref.json", want: "external reference", limits: baseLimits},
		{name: "unresolved response schema reference", artifact: "unresolved-ref.json", want: "unresolved reference", wantGap: sourceMalformedReferenceFoundation, limits: baseLimits},
		{name: "ambiguous request", artifact: "ambiguous-request.json", want: "ambiguous request schema", limits: baseLimits},
		{name: "duplicate identity", artifact: "duplicate-id.json", want: "duplicate source identity", limits: baseLimits},
		{name: "unbounded request", artifact: "unbounded-request.json", want: "unbounded request schema", limits: baseLimits},
		{name: "missing additional properties", artifact: "missing-additional-properties.json", want: "dynamic additionalProperties", limits: baseLimits},
		{name: "unsupported encoding", artifact: "unsupported-encoding.json", want: "unsupported request encoding", limits: baseLimits},
		{name: "invalid relative path", artifact: "invalid-relative-path.json", want: "connector-relative", limits: baseLimits},
		{name: "whitespace path", artifact: "whitespace-path.json", want: "connector-relative", limits: baseLimits},
		{name: "missing path parameter", artifact: "missing-path-parameter.json", want: "path placeholder", wantGap: sourceMalformedPathParameterFoundation, limits: baseLimits},
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
			result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), tc.limits)
			if tc.wantGap != "" {
				if err != nil {
					t.Fatalf("retained malformed source import: %v", err)
				}
				if len(result.Operations) != 1 || !result.Operations[0].Runtime.MergeBlocked {
					t.Fatalf("retained malformed operation = %#v", result.Operations)
				}
				for _, gap := range result.Operations[0].Runtime.Gaps {
					if gap.Foundation == tc.wantGap && strings.Contains(gap.Location, result.Operations[0].Source.Location) && strings.Contains(gap.Reason, tc.want) {
						return
					}
				}
				t.Fatalf("retained malformed source gaps = %#v", result.Operations[0].Runtime.Gaps)
			}
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
	if got := newSourceImportHTTPClient(publicLookup).Timeout; got != defaultSourceImportFetchTimeout {
		t.Fatalf("source importer cold fetch timeout = %s, want %s", got, defaultSourceImportFetchTimeout)
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

func TestSourceInferredNextURLPaginationRetainsClosedLimitOffsetControls(t *testing.T) {
	responses := []sourceResponseDescriptor{{
		Status: "200",
		Declaration: map[string]any{"content": map[string]any{
			"application/json": map[string]any{"schema": map[string]any{"properties": map[string]any{
				"next_page": map[string]any{"type": "object", "properties": map[string]any{
					"uri": map[string]any{"type": "string", "format": "uri"},
				}},
			}}},
		}},
	}}
	query := []sourceParameterDescriptor{
		{Name: "limit", Schema: map[string]any{"type": "integer", "minimum": json.Number("1"), "maximum": json.Number("100")}},
		{Name: "offset", Schema: map[string]any{"type": "string"}},
	}
	got := sourceInferredNextURLPagination(responses, query)
	want := map[string]any{
		"type": "next_url", "next_url_path": "next_page.uri",
		"size_param": "limit", "limit_param": "limit", "offset_param": "offset", "page_size": 100,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("inferred next-url pagination = %#v, want %#v", got, want)
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

func TestSourceImportRetainedArtifactImportsMachineReadableSpecWithoutProvider(t *testing.T) {
	t.Parallel()
	raw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://provider.example.invalid/alpha-openapi.yaml", raw)
	sourcesDir := filepath.Join(t.TempDir(), "sources")
	writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)

	fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct retained artifact fetcher: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import retained source with unreachable provider URL: %v", err)
	}
	if len(result.Operations) != 2 || result.Operations[0].Source.URL != lock.Rest.SourceURL {
		t.Fatalf("retained source import result = %#v", result.Operations)
	}
}

func TestSourceImportRetainedArtifactRejectsMissingAndMismatchedCopies(t *testing.T) {
	t.Parallel()
	raw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://provider.example.invalid/alpha-openapi.yaml", raw)

	t.Run("missing retained copy never falls back to provider", func(t *testing.T) {
		sourcesDir := filepath.Join(t.TempDir(), "sources")
		writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, nil)
		fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("construct retained artifact fetcher: %v", err)
		}
		_, err = importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "retained source artifact is missing") || strings.Contains(err.Error(), "provider.example.invalid") {
			t.Fatalf("missing retained artifact error = %v", err)
		}
	})

	t.Run("retained bytes must match preexisting lock", func(t *testing.T) {
		sourcesDir := filepath.Join(t.TempDir(), "sources")
		writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, []byte("not the locked bytes"))
		fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("construct retained artifact fetcher: %v", err)
		}
		_, err = importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "retained source artifact does not match locked bytes and SHA-256") {
			t.Fatalf("mismatched retained artifact error = %v", err)
		}
	})

	t.Run("manifest provenance cannot redirect a lock", func(t *testing.T) {
		sourcesDir := filepath.Join(t.TempDir(), "sources")
		writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)
		manifestPath := filepath.Join(sourcesDir, lock.Connector+"-retained-artifacts.json")
		manifest, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("read fixture manifest: %v", err)
		}
		manifest = bytes.Replace(manifest, []byte("provider.example.invalid"), []byte("redirect.example.invalid"), 1)
		if err := os.WriteFile(manifestPath, manifest, 0o644); err != nil {
			t.Fatalf("tamper fixture manifest: %v", err)
		}
		fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("construct retained artifact fetcher: %v", err)
		}
		_, err = importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "provenance does not match source lock") {
			t.Fatalf("redirected provenance error = %v", err)
		}
	})

	t.Run("manifest identity-query declaration cannot change a lock", func(t *testing.T) {
		sourcesDir := filepath.Join(t.TempDir(), "sources")
		writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)
		manifestPath := filepath.Join(sourcesDir, lock.Connector+"-retained-artifacts.json")
		manifest, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("read fixture manifest: %v", err)
		}
		manifest = bytes.Replace(manifest, []byte(`"identity_query":false`), []byte(`"identity_query":true`), 1)
		if err := os.WriteFile(manifestPath, manifest, 0o644); err != nil {
			t.Fatalf("tamper identity-query fixture manifest: %v", err)
		}
		_, err = newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "identity artifact query is missing") {
			t.Fatalf("redirected identity-query error = %v", err)
		}
	})

	t.Run("symlinked retained copy is rejected", func(t *testing.T) {
		sourcesDir := filepath.Join(t.TempDir(), "sources")
		writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)
		artifactPath := filepath.Join(sourcesDir, "artifacts", strings.ToLower(lock.Rest.SHA256)+".artifact")
		external := filepath.Join(t.TempDir(), "external.artifact")
		if err := os.WriteFile(external, raw, 0o644); err != nil {
			t.Fatalf("write external retained fixture: %v", err)
		}
		if err := os.Remove(artifactPath); err != nil {
			t.Fatalf("remove fixture artifact before symlink: %v", err)
		}
		if err := os.Symlink(external, artifactPath); err != nil {
			t.Fatalf("create retained artifact symlink: %v", err)
		}
		fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("construct retained artifact fetcher: %v", err)
		}
		_, err = importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
		if err == nil || !strings.Contains(err.Error(), "regular non-symlink") {
			t.Fatalf("symlinked retained artifact error = %v", err)
		}
	})
}

func TestSourceImportRetainedArtifactSupportsMachineRenderedAndBundleShapes(t *testing.T) {
	t.Parallel()
	spec := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	bundle := sourceImportRetainedZIP(t, "openapi.yaml", spec)
	tests := []struct {
		name string
		url  string
		raw  []byte
	}{
		{name: "machine readable", url: "https://provider.example.invalid/openapi.yaml", raw: spec},
		{name: "rendered citation", url: "https://provider.example.invalid/reference.html", raw: []byte("<html><body>provider reference</body></html>")},
		{name: "zip bundle", url: "https://provider.example.invalid/openapi.zip", raw: bundle},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			artifact := sourceImportFixtureLock("alpha", tc.url, tc.raw).Rest.sourceImportArtifact
			sourcesDir := filepath.Join(t.TempDir(), "sources")
			writeSourceImportRetainedFixture(t, sourcesDir, "alpha", artifact, tc.raw)
			fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, "alpha", defaultSourceImportLimits())
			if err != nil {
				t.Fatalf("construct retained artifact fetcher: %v", err)
			}
			got, err := fetcher.FetchArtifact(context.Background(), artifact)
			if err != nil || !bytes.Equal(got, tc.raw) {
				t.Fatalf("retained %s bytes = %d/%v, want exact %d bytes", tc.name, len(got), err, len(tc.raw))
			}
		})
	}
}

func TestSourceImportRetainedArtifactEncodesElasticsearchAndZoomRecoveryRegressions(t *testing.T) {
	t.Parallel()
	raw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "elasticsearch source bytes changed upstream",
			url:  "https://raw.githubusercontent.com/elastic/elasticsearch-specification/main/output/openapi/elasticsearch-openapi.json",
		},
		{
			name: "zoom accounts source is upstream 404",
			url:  "https://developers.zoom.us/_next/data/2026-08-17T14-01-06-06-00/docs/api/accounts.json?slug=accounts",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lock := sourceImportFixtureLock("alpha", tc.url, raw)
			lock.Rest.IdentityQuery = strings.Contains(tc.url, "?")
			sourcesDir := filepath.Join(t.TempDir(), "sources")
			writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)
			fetcher, err := newSourceImportRetainedArtifactFetcher(sourcesDir, lock.Connector, defaultSourceImportLimits())
			if err != nil {
				t.Fatalf("construct retained artifact fetcher: %v", err)
			}
			result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
			if err != nil || len(result.Operations) != 2 {
				t.Fatalf("retained historic source result = %d/%v", len(result.Operations), err)
			}
		})
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
	for _, required := range []string{
		"source-import <connector>",
		"byte-backed",
		"declaration-only source reference",
		"source_contract_unavailable",
		"retained content-addressed document artifacts",
	} {
		if !strings.Contains(stdout.String(), required) {
			t.Fatalf("source-import help is missing %q: %s", required, stdout.String())
		}
	}
	if strings.Contains(stdout.String(), "--cache-dir") {
		t.Fatalf("source-import help exposes obsolete cache flag: %s", stdout.String())
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
	if !strings.Contains(string(docs), "connectorgen source-import") || !strings.Contains(string(docs), "retained-artifacts.json") || !strings.Contains(string(docs), "missing copy") || !strings.Contains(string(docs), "identity_query") || !strings.Contains(string(docs), "pm-request-contract-bounds-v1") || !strings.Contains(string(docs), "request_execution_limits") {
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

func TestSourceImportCommandDerivesWriteDisabledMutationArtifacts(t *testing.T) {
	defsRoot := t.TempDir()
	bundleDir := filepath.Join(defsRoot, "alpha")
	sourcesDir := filepath.Join(bundleDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create source fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "alpha-operation-source-lock.json"), loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json")), 0o644); err != nil {
		t.Fatalf("write source lock: %v", err)
	}
	const metadata = `{
	  "name":"alpha",
	  "display_name":"Alpha",
	  "description":"fixture",
	  "integration_type":"api",
	  "release_stage":"ga",
	  "capabilities":{"check":true,"read":true,"write":false,"query":false,"cdc":false,"dynamic_schema":false}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "metadata.json"), []byte(metadata), 0o644); err != nil {
		t.Fatalf("write write-disabled metadata: %v", err)
	}

	outPath := filepath.Join(sourcesDir, "alpha-operation-descriptor.json")
	args := []string{"source-import", "alpha", "--defs", defsRoot, "--out", outPath}
	var stdout, stderr bytes.Buffer
	if exit := runSourceImportWithFetcher(args, &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 0 {
		t.Fatalf("write-disabled source-import exit = %d, stderr = %s", exit, stderr.String())
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read imported descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(raw, &descriptor); err != nil {
		t.Fatalf("decode imported descriptor: %v", err)
	}
	var mutation sourceOperationDescriptor
	for _, operation := range descriptor.Operations {
		if operation.SourceID == "alpha.rest.post_/widgets" {
			mutation = operation
			break
		}
	}
	if !sourceProjectionHasNonExecutableMutationDisposition(mutation) {
		t.Fatalf("imported mutation = %#v, want source-cited non-executable artifact", mutation)
	}
	if mutation.Runtime.NonExecutableMutation.Reason != sourceWriteDisabledMutationArtifactReason ||
		mutation.Runtime.NonExecutableMutation.Source.SourceID != mutation.SourceID ||
		!strings.EqualFold(mutation.Runtime.NonExecutableMutation.Source.Method, mutation.Method) ||
		mutation.Runtime.NonExecutableMutation.Source.Path != mutation.Path {
		t.Fatalf("imported mutation artifact = %#v, want exact automatic source citation", mutation.Runtime.NonExecutableMutation)
	}
	if !sourceOperationHasFoundationGap(mutation, sourceNonExecutableMutationDispositionFoundation) {
		t.Fatalf("imported mutation gaps = %#v, want named non-executable foundation", mutation.Runtime.Gaps)
	}

	stdout.Reset()
	stderr.Reset()
	if exit := runSourceImportWithFetcher(append(args, "--check"), &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 0 {
		t.Fatalf("write-disabled source-import --check exit = %d, stderr = %s", exit, stderr.String())
	}
}

func TestSourceImportCommandRejectsOmittedWriteCapabilityBeforeArtifactAdmission(t *testing.T) {
	defsRoot := t.TempDir()
	bundleDir := filepath.Join(defsRoot, "alpha")
	sourcesDir := filepath.Join(bundleDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create source fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "alpha-operation-source-lock.json"), loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json")), 0o644); err != nil {
		t.Fatalf("write source lock: %v", err)
	}
	const metadataWithoutWrite = `{
	  "name":"alpha",
	  "display_name":"Alpha",
	  "description":"fixture",
	  "integration_type":"api",
	  "release_stage":"ga",
	  "capabilities":{"check":true,"read":true,"query":false,"cdc":false,"dynamic_schema":false}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "metadata.json"), []byte(metadataWithoutWrite), 0o644); err != nil {
		t.Fatalf("write metadata without write capability: %v", err)
	}

	outPath := filepath.Join(sourcesDir, "alpha-operation-descriptor.json")
	args := []string{"source-import", "alpha", "--defs", defsRoot, "--out", outPath}
	var stdout, stderr bytes.Buffer
	if exit := runSourceImportWithFetcher(args, &stdout, &stderr, fixtureSourceImportFetcher(t)); exit != 1 || !strings.Contains(stderr.String(), "capabilities.write must be explicitly declared") {
		t.Fatalf("source-import missing capabilities.write exit = %d stderr = %s, want explicit write declaration refusal", exit, stderr.String())
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatalf("source-import with omitted capabilities.write wrote descriptor: %v", err)
	}
}

func TestSourceImportRejectsSymlinkedSourcesDirectoryEvenInsideConnectorBundle(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	bundleDir := filepath.Join(defsRoot, "alpha")
	realSourcesDir := filepath.Join(bundleDir, "real-sources")
	if err := os.MkdirAll(realSourcesDir, 0o755); err != nil {
		t.Fatalf("create real sources directory: %v", err)
	}
	lockRaw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-operation-source-lock.json"))
	if err := os.WriteFile(filepath.Join(realSourcesDir, "alpha-operation-source-lock.json"), lockRaw, 0o644); err != nil {
		t.Fatalf("write source lock: %v", err)
	}
	if err := os.Symlink(realSourcesDir, filepath.Join(bundleDir, "sources")); err != nil {
		t.Fatalf("create in-bundle sources symlink: %v", err)
	}
	if _, err := loadConnectorSourceImportLock(defsRoot, "alpha"); err == nil || !strings.Contains(err.Error(), "source directory must not be a symlink") {
		t.Fatalf("in-bundle sources symlink error = %v, want terminal refusal", err)
	}
}

func TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified(t *testing.T) {
	t.Parallel()
	defsRoot := filepath.Join("..", "..", "internal", "connectors", "defs")
	lock, err := loadConnectorSourceImportLock(defsRoot, "github")
	if err != nil {
		t.Fatalf("load GitHub source lock: %v", err)
	}
	fetcher, err := newConnectorSourceImportRetainedArtifactFetcher(defsRoot, "github", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct GitHub retained-artifact reader: %v", err)
	}
	for _, artifact := range []sourceImportArtifact{lock.Rest.sourceImportArtifact, lock.GraphQL.sourceImportArtifact} {
		raw, fetchErr := fetcher.FetchArtifact(context.Background(), artifact)
		if fetchErr != nil {
			t.Fatalf("read GitHub retained %s: %v", artifact.SourceURL, fetchErr)
		}
		if validateErr := validateSourceImportArtifactBytes(raw, artifact); validateErr != nil {
			t.Fatalf("verify GitHub retained %s: %v", artifact.SourceURL, validateErr)
		}
	}
}

func TestSourceImportRetainedAsanaPreservesLockedRESTOperationIDs(t *testing.T) {
	defsRoot := filepath.Join("..", "..", "internal", "connectors", "defs")
	lock, err := loadConnectorSourceImportLock(defsRoot, "asana")
	if err != nil {
		t.Fatalf("load Asana source lock: %v", err)
	}
	fetcher, err := newConnectorSourceImportRetainedArtifactFetcher(defsRoot, "asana", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct Asana retained-artifact reader: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import retained Asana source lock: %v", err)
	}
	byRoute := make(map[string]sourceOperationDescriptor, len(result.Operations))
	for _, operation := range result.Operations {
		byRoute[strings.ToUpper(operation.Method)+"\x00"+operation.Path] = operation
	}
	for _, want := range []struct {
		name        string
		sourceID    string
		operationID string
		method      string
		path        string
	}{
		{
			name:        "fixed 100 ETL custom fields",
			sourceID:    "asana.rest.getCustomFieldsForWorkspace",
			operationID: "getCustomFieldsForWorkspace",
			method:      "GET",
			path:        "/workspaces/{workspace_gid}/custom_fields",
		},
		{
			name:        "partial mutation custom field",
			sourceID:    "asana.rest.createCustomField",
			operationID: "createCustomField",
			method:      "POST",
			path:        "/custom_fields",
		},
		{
			name:        "fan-out sections stream",
			sourceID:    "asana.rest.getSectionsForProject",
			operationID: "getSectionsForProject",
			method:      "GET",
			path:        "/projects/{project_gid}/sections",
		},
	} {
		want := want
		t.Run(want.name, func(t *testing.T) {
			operation, found := byRoute[want.method+"\x00"+want.path]
			if !found {
				t.Fatalf("imported Asana operation %s %s is absent", want.method, want.path)
			}
			if operation.SourceID != want.sourceID || operation.ProviderOperationID != want.operationID {
				t.Fatalf("imported Asana operation %s %s identity = source_id=%q operation_id=%q, want source_id=%q operation_id=%q", want.method, want.path, operation.SourceID, operation.ProviderOperationID, want.sourceID, want.operationID)
			}
		})
	}

	t.Run("rejects mismatched locked provider identity before ID propagation", func(t *testing.T) {
		mismatched := lock
		mismatched.Rest.SourceDocuments = append([]sourceImportRESTDocument(nil), lock.Rest.SourceDocuments...)
		mismatched.Rest.SourceDocuments[0].Operations = append([]sourceImportRESTOperation(nil), lock.Rest.SourceDocuments[0].Operations...)
		for index := range mismatched.Rest.SourceDocuments[0].Operations {
			if mismatched.Rest.SourceDocuments[0].Operations[index].ID == "asana.rest.getSectionsForProject" {
				mismatched.Rest.SourceDocuments[0].Operations[index].OperationID = "wrongSectionsOperation"
				break
			}
		}
		if _, err := importSourceLockResult(context.Background(), mismatched, fetcher, defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "disagrees with source document \"asana-openapi\" inventory") {
			t.Fatalf("mismatched locked provider identity error = %v, want locked/provider refusal", err)
		}
	})
}

func TestSourceImportCommandReadsConnectorOwnedRetainedArtifact(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	raw := loadSourceImportFixture(t, filepath.Join("alpha", "alpha-openapi.yaml"))
	lock := sourceImportFixtureLock("alpha", "https://provider.example.invalid/alpha-openapi.yaml", raw)
	lockRaw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal fixture lock: %v", err)
	}
	writeSourceImportRetainedFixture(t, sourcesDir, lock.Connector, lock.Rest.sourceImportArtifact, raw)
	if err := os.WriteFile(filepath.Join(sourcesDir, "alpha-operation-source-lock.json"), lockRaw, 0o644); err != nil {
		t.Fatalf("write retained fixture lock: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "alpha-descriptors.json")
	var stdout, stderr bytes.Buffer
	if exit := runSourceImportWithFetcher([]string{"source-import", "alpha", "--defs", defsRoot, "--out", outPath}, &stdout, &stderr, nil); exit != 0 {
		t.Fatalf("retained source-import exit = %d, stderr = %s", exit, stderr.String())
	}
	if descriptor, err := os.ReadFile(outPath); err != nil || !strings.Contains(string(descriptor), "getWidget") {
		t.Fatalf("retained source-import descriptor = %q, error = %v", descriptor, err)
	}
}

func TestSourceImportReportsUnavailableSourceBeforeRequiringRetainedManifest(t *testing.T) {
	t.Parallel()
	defsRoot := t.TempDir()
	sourcesDir := filepath.Join(defsRoot, "zoom", "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("create Zoom sources directory: %v", err)
	}
	lock := map[string]any{
		"schema_version": 3,
		"connector":      "zoom",
		"rest": map[string]any{
			"retrieval": "Zoom accounts historical source unavailable",
			"openapi":   []any{},
			"coverage_confidence": map[string]any{
				"level": "unavailable-public-source",
				"basis": "accounts source returned HTTP 404; no verified historic copy exists",
			},
			"source_documents": []any{map[string]any{
				"id":                 "accounts",
				"kind":               "unavailable",
				"unavailable_reason": "accounts source returned HTTP 404; no verified historic copy exists",
				"operations":         []any{},
			}},
		},
		"counts": map[string]any{"rest": 0, "graphql_query": 0, "graphql_mutation": 0, "total": 0},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal unavailable Zoom lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "zoom-operation-source-lock.json"), raw, 0o644); err != nil {
		t.Fatalf("write unavailable Zoom lock: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if exit := runSourceImportWithFetcher([]string{"source-import", "zoom", "--defs", defsRoot}, &stdout, &stderr, nil); exit != 1 || !strings.Contains(stderr.String(), "source document \"accounts\" is unavailable") || strings.Contains(stderr.String(), "retained artifact manifest") {
		t.Fatalf("unavailable source-import exit/stderr = %d/%q", exit, stderr.String())
	}
}

func TestSourceImportCommandRejectsCacheDirWithoutCallingFetcher(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	calls := 0
	exit := runSourceImportWithFetcher([]string{"source-import", "alpha", "--cache-dir", t.TempDir()}, &stdout, &stderr, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		calls++
		return nil, fmt.Errorf("cache fallback must not be invoked")
	}))
	if exit != 2 || !strings.Contains(stderr.String(), "no longer supported") || calls != 0 {
		t.Fatalf("cache-dir source-import exit/stderr/calls = %d/%q/%d", exit, stderr.String(), calls)
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
	if _, err := loadConnectorSourceImportLock(defsRoot, "alpha"); err == nil || !strings.Contains(err.Error(), "source directory must not be a symlink") {
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

func TestSourceImportNormalizesScalarYAMLMappingKeys(t *testing.T) {
	t.Parallel()
	numeric := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  /items:\n    get:\n      responses:\n        200: {description: ok}\n")
	quoted := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  /items:\n    get:\n      responses:\n        '200': {description: ok}\n")
	importArtifact := func(t *testing.T, sourceURL string, raw []byte) sourceImportResult {
		t.Helper()
		lock := sourceImportFixtureLock("alpha", sourceURL, raw)
		result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
			return raw, nil
		}), defaultSourceImportLimits())
		if err != nil {
			t.Fatalf("import %s: %v", sourceURL, err)
		}
		return result
	}

	numericResult := importArtifact(t, "https://fixtures.polymetrics.invalid/numeric.yaml", numeric)
	quotedResult := importArtifact(t, "https://fixtures.polymetrics.invalid/quoted.yaml", quoted)
	if len(numericResult.Operations) != 1 || len(quotedResult.Operations) != 1 {
		t.Fatalf("operation counts numeric=%d quoted=%d", len(numericResult.Operations), len(quotedResult.Operations))
	}
	if response := descriptorResponse(t, numericResult.Operations[0], "200"); response.Declaration.(map[string]any)["description"] != "ok" {
		t.Fatalf("numeric response = %#v", response)
	}
	numericDescriptor := numericResult.Operations[0]
	quotedDescriptor := quotedResult.Operations[0]
	// The source digest and URL intentionally differ because the two pinned
	// artifacts have distinct bytes; their imported operation contract must not.
	numericDescriptor.Source = sourceImportSource{}
	quotedDescriptor.Source = sourceImportSource{}
	if !reflect.DeepEqual(numericDescriptor, quotedDescriptor) {
		t.Fatalf("numeric and quoted YAML imports differ\nnumeric: %#v\nquoted: %#v", numericDescriptor, quotedDescriptor)
	}

	duplicate := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  /items:\n    get:\n      responses:\n        200: {description: ok}\n        '200': {description: duplicate}\n")
	duplicateLock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/duplicate-normalized.yaml", duplicate)
	if _, err := importSourceLockResult(context.Background(), duplicateLock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return duplicate, nil
	}), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "duplicate YAML mapping key") {
		t.Fatalf("duplicate normalized YAML key error = %v", err)
	}

	nonScalar := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  ? [/items]\n  : get: {responses: {'200': {description: ok}}}\n")
	nonScalarLock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/non-scalar-key.yaml", nonScalar)
	if _, err := importSourceLockResult(context.Background(), nonScalarLock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return nonScalar, nil
	}), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "must be a scalar") {
		t.Fatalf("non-scalar YAML key error = %v", err)
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

func TestSourceImportRetainsBoundedOpenAPI30ReferenceSiblings(t *testing.T) {
	t.Parallel()
	for _, sibling := range []struct {
		name  string
		field string
		value string
	}{
		{name: "description", field: "description", value: "provider description"},
		{name: "summary", field: "summary", value: "provider summary"},
	} {
		sibling := sibling
		t.Run(sibling.name, func(t *testing.T) {
			raw := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"responses":{"ok":{"description":"original"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/ok","` + sibling.field + `":"` + sibling.value + `"}}}}}}`)
			result := importInlineSourceResult(t, raw, defaultSourceImportLimits())
			response := descriptorResponse(t, result.Operations[0], "200")
			declaration, ok := response.Declaration.(map[string]any)
			if !ok || declaration[sibling.field] != sibling.value {
				t.Fatalf("OpenAPI 3.0 %s sibling declaration = %#v", sibling.field, response.Declaration)
			}
		})
	}

	extension := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"responses":{"ok":{"description":"original"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/ok","x-provider-note":"preserved"}}}}}}`)
	if result := importInlineSourceResult(t, extension, defaultSourceImportLimits()); len(result.Operations) != 1 {
		t.Fatalf("OpenAPI 3.0 extension sibling result = %#v", result.Operations)
	}

	readOnly := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"schemas":{"identifier":{"type":"string","maxLength":8}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"$ref":"#/components/schemas/identifier","readOnly":true}}}}}}}}}`)
	result := importInlineSourceResult(t, readOnly, defaultSourceImportLimits())
	response := descriptorResponse(t, result.Operations[0], "200")
	media := response.Declaration.(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)
	schema := media["schema"].(map[string]any)
	if schema["readOnly"] != true {
		t.Fatalf("OpenAPI 3.0 readOnly schema sibling = %#v", schema)
	}

	matchingType := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"schemas":{"identifier":{"type":"string","maxLength":8}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"$ref":"#/components/schemas/identifier","type":"string"}}}}}}}}}`)
	result = importInlineSourceResult(t, matchingType, defaultSourceImportLimits())
	response = descriptorResponse(t, result.Operations[0], "200")
	media = response.Declaration.(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)
	schema = media["schema"].(map[string]any)
	if schema["type"] != "string" {
		t.Fatalf("OpenAPI 3.0 matching schema type sibling = %#v", schema)
	}

	conflictingType := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"schemas":{"identifier":{"type":"string","maxLength":8}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"$ref":"#/components/schemas/identifier","type":"integer"}}}}}}}}}`)
	result = importInlineSourceResult(t, conflictingType, defaultSourceImportLimits())
	response = descriptorResponse(t, result.Operations[0], "200")
	media = response.Declaration.(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)
	schema = media["schema"].(map[string]any)
	if schema["type"] != "integer" {
		t.Fatalf("OpenAPI 3.0 retained type sibling = %#v", schema)
	}
	var typeGap *sourceContractGap
	for index := range result.Operations[0].Runtime.Gaps {
		gap := &result.Operations[0].Runtime.Gaps[index]
		if gap.Foundation == "cli-openapi30-reference-sibling-foundation-r1" {
			typeGap = gap
			break
		}
	}
	if typeGap == nil || !result.Operations[0].Runtime.MergeBlocked || result.Operations[0].Source.Location != `paths["/items"].get` || typeGap.Location != "schema reference #/components/schemas/identifier" || !strings.Contains(typeGap.Reason, `sibling field "type"`) {
		t.Fatalf("OpenAPI 3.0 type sibling source-bound gap = %#v", result.Operations[0].Runtime)
	}

	semantic := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"responses":{"ok":{"description":"original"}}},"paths":{"/items":{"get":{"responses":{"200":{"$ref":"#/components/responses/ok","content":{"application/json":{"schema":{"type":"string"}}}}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/openapi30-semantic-sibling.json", semantic)
	_, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return semantic, nil }), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), `ambiguous response reference with sibling field "content"`) {
		t.Fatalf("OpenAPI 3.0 semantic reference sibling error = %v", err)
	}
}

func TestSourceImportQualifiesReferenceSiblingGapByRequestOrResponsePhase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		raw          []byte
		wantPhase    string
		wantBlocking bool
	}{
		{
			name: "response-only schema sibling remains evidence without blocking a bounded read",
			raw: []byte(`{
  "openapi":"3.0.3","info":{"title":"x","version":"1"},
  "components":{"schemas":{"Identifier":{"type":"string","maxLength":32}}},
  "paths":{"/items":{"get":{"operationId":"getItems","responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Identifier","type":"integer"}}}}}}}}
}`),
			wantPhase:    "response",
			wantBlocking: false,
		},
		{
			name: "request schema sibling still blocks an incomplete request contract",
			raw: []byte(`{
  "openapi":"3.0.3","info":{"title":"x","version":"1"},
  "components":{"schemas":{"Identifier":{"type":"string","maxLength":32}}},
  "paths":{"/items":{"get":{"operationId":"getItems","parameters":[{"name":"id","in":"query","required":true,"schema":{"$ref":"#/components/schemas/Identifier","type":"boolean"}}],"responses":{"200":{"description":"ok"}}}}}
}`),
			wantPhase:    "request",
			wantBlocking: true,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := importInlineSourceResult(t, testCase.raw, defaultSourceImportLimits())
			if len(result.Operations) != 1 {
				t.Fatalf("operation count = %d, want 1", len(result.Operations))
			}
			operation := result.Operations[0]
			var siblingGap *sourceContractGap
			for index := range operation.Runtime.Gaps {
				if operation.Runtime.Gaps[index].Foundation == sourceOpenAPI30ReferenceSiblingFoundation {
					siblingGap = &operation.Runtime.Gaps[index]
					break
				}
			}
			if siblingGap == nil {
				t.Fatalf("runtime gaps = %+v, want retained OpenAPI 3.0 sibling evidence", operation.Runtime.Gaps)
			}
			encoded, err := json.Marshal(siblingGap)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(encoded), `"phase":"`+testCase.wantPhase+`"`) {
				t.Fatalf("qualified sibling gap = %s, want phase %q", encoded, testCase.wantPhase)
			}
			if got := sourceProjectionReadHasBlockingGap(operation); got != testCase.wantBlocking {
				t.Fatalf("read blocking = %t, want %t for %s-only sibling gap", got, testCase.wantBlocking, testCase.wantPhase)
			}
		})
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

func TestSourceImportRejectsCrossKindIdentitiesAndUnsupportedYAMLKeyTags(t *testing.T) {
	t.Parallel()
	collision := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"webhooks":{"invoice.created":{"post":{"responses":{"200":{"description":"ok"}}}}},"paths":{"/items":{"get":{"operationId":"alpha.inbound.webhook.invoice.created","responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/collision.json", collision)
	if _, err := importSourceLock(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return collision, nil }), defaultSourceImportLimits()); err == nil || !strings.Contains(err.Error(), "duplicate source identity") {
		t.Fatalf("cross-kind identity error = %v", err)
	}
	yaml := []byte("openapi: 3.1.0\ninfo: {title: x, version: '1'}\npaths:\n  !!timestamp \"2026-08-23\": {}\n")
	if _, _, err := parseSourceImportDocument(yaml); err == nil || !strings.Contains(err.Error(), "must use a JSON scalar type") {
		t.Fatalf("unsupported YAML key tag error = %v", err)
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

func TestSourceImportScopesSchemaExpansionToEachRoot(t *testing.T) {
	t.Parallel()
	limits := defaultSourceImportLimits()
	limits.MaxSchemaNodes = 3
	resolver := sourceReferenceResolver{limits: limits, form: sourceDocumentForm{Family: "openapi", Version: "3.0.3"}}
	schema := map[string]any{"type": "string", "maxLength": json.Number("8")}
	for index := 0; index < 4; index++ {
		if _, err := resolver.resolveSchema(schema, nil, 0); err != nil {
			t.Fatalf("root schema %d shared an expansion budget with a prior root: %v", index, err)
		}
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

	unusedCycle := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"components":{"schemas":{"Unused":{"$ref":"#/components/schemas/Unused"}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/unused-grammar-cycle.json", unusedCycle)
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return unusedCycle, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("unused recursive schema import: %v", err)
	}
	if len(result.Gaps) != 1 {
		t.Fatalf("unused recursive schema gaps = %#v", result.Gaps)
	}
	gap := result.Gaps[0]
	if gap.Foundation != sourceRecursiveSchemaFoundation || !strings.Contains(gap.Location, "#/components/schemas/Unused") || !strings.Contains(gap.Reason, "#/components/schemas/Unused") {
		t.Fatalf("unused recursive schema gap = %#v", gap)
	}
	encoded, err := marshalSourceImportResult(result)
	if err != nil {
		t.Fatalf("marshal unused recursive schema descriptor: %v", err)
	}
	if !strings.Contains(string(encoded), `"merge_blocked": true`) || !strings.Contains(string(encoded), `"foundation": "cli-recursive-schema-foundation-r1"`) {
		t.Fatalf("unused recursive schema evidence disappeared from descriptor: %s", encoded)
	}

	unusedTypeSibling := []byte(`{"openapi":"3.0.3","info":{"title":"x","version":"1"},"components":{"schemas":{"Base":{"anyOf":[{"type":"object","additionalProperties":false,"properties":{}}]},"Unused":{"allOf":[{"$ref":"#/components/schemas/Base","type":"object"}]}}},"paths":{"/items":{"get":{"responses":{"200":{"description":"ok"}}}}}}`)
	lock = sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/unused-grammar-type-sibling.json", unusedTypeSibling)
	result, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return unusedTypeSibling, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("unused type sibling import: %v", err)
	}
	if len(result.Gaps) != 1 {
		t.Fatalf("unused type sibling gaps = %#v", result.Gaps)
	}
	gap = result.Gaps[0]
	if gap.Foundation != sourceOpenAPI30ReferenceSiblingFoundation || gap.Location != "schema reference #/components/schemas/Base" || !strings.Contains(gap.Reason, `sibling field "type"`) {
		t.Fatalf("unused type sibling gap = %#v", gap)
	}
	encoded, err = marshalSourceImportResult(result)
	if err != nil {
		t.Fatalf("marshal unused type sibling descriptor: %v", err)
	}
	if !strings.Contains(string(encoded), `"merge_blocked": true`) || !strings.Contains(string(encoded), `"foundation": "cli-openapi30-reference-sibling-foundation-r1"`) {
		t.Fatalf("unused type sibling evidence disappeared from descriptor: %s", encoded)
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
	cycleResult, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return literalPercentCycle, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("literal percent recursive schema import: %v", err)
	}
	if len(cycleResult.Operations) != 1 || !cycleResult.Operations[0].Runtime.MergeBlocked {
		t.Fatalf("literal percent recursive schema operation = %#v", cycleResult.Operations)
	}
	var cycleGap *sourceContractGap
	for index := range cycleResult.Operations[0].Runtime.Gaps {
		gap := &cycleResult.Operations[0].Runtime.Gaps[index]
		if gap.Foundation == sourceRecursiveSchemaFoundation {
			cycleGap = gap
			break
		}
	}
	if cycleGap == nil || !strings.Contains(cycleGap.Reason, "#/components/schemas/Root/properties/percent%2F") {
		t.Fatalf("literal percent recursive schema gap = %#v", cycleResult.Operations[0].Runtime.Gaps)
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

// TestSourceImportProviderDialectContracts keeps the seven provider failures
// reported by the Batch-1 repair run behavioral. Each fixture is reduced to
// the affected operation and its provider-declared response/request fragment;
// its path, operationId, pointer, and dialect keyword are copied from the
// pinned upstream artifact recorded in issue #4325's RUN-STATE.md. The test
// exercises the importer, not a declaration count or a schema-shape helper.
func TestSourceImportProviderDialectContracts(t *testing.T) {
	tests := []struct {
		name            string
		raw             []byte
		wantOperation   string
		wantResponseKey string
		wantGap         string
		wantTrace       string
	}{
		{
			name:            "Bitbucket pull request comment response stays within the raised finite schema bound",
			raw:             sourceProviderNestedResponseDocument("3.0.0", "/repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/comments/{comment_id}", "getPullRequestComment", 16),
			wantOperation:   "getPullRequestComment",
			wantResponseKey: "level_00",
		},
		{
			name:            "Notion meeting notes response stays within the raised finite schema bound",
			raw:             sourceProviderNestedResponseDocument("3.1.0", "/v1/blocks/meeting_notes/query", "query-meeting-notes", 16),
			wantOperation:   "query-meeting-notes",
			wantResponseKey: "level_00",
		},
		{
			name:            "Stripe GET account follows a finite provider reference chain",
			raw:             sourceProviderReferenceChainDocument(30),
			wantOperation:   "GetAccount",
			wantResponseKey: "next",
		},
		{
			name:            "Vercel API key response retains OpenAPI 3.0 pattern properties",
			raw:             sourceVercelPatternPropertiesDocument(),
			wantOperation:   "createApiKeys",
			wantResponseKey: "patternProperties",
		},
		{
			name:          "Docker Hub malformed response target is retained with a source trace",
			raw:           sourceDockerHubMalformedReferenceDocument(),
			wantOperation: "createToken",
			wantGap:       "cli-malformed-source-reference-foundation-r1",
			wantTrace:     "#/components/responses/team_repo",
		},
		{
			name:          "GitLab malformed epic issue path stays present with a source trace",
			raw:           sourceGitLabMissingPathParameterDocument(),
			wantOperation: "putApiV4GroupsIdDashEpicsEpicIidIssuesEpicIssueId",
			wantGap:       "cli-malformed-path-parameter-foundation-r1",
			wantTrace:     "epic_issue_id",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result, err := importInlineSourceResultError(tt.raw, defaultSourceImportLimits())
			if err != nil {
				t.Fatalf("provider-derived source import: %v", err)
			}
			if len(result.Operations) != 1 || result.Operations[0].SourceID != tt.wantOperation {
				t.Fatalf("provider operation was not retained: %#v", result.Operations)
			}
			operation := result.Operations[0]
			encoded, err := json.Marshal(operation.Responses)
			if err != nil {
				t.Fatalf("marshal retained provider response: %v", err)
			}
			if tt.wantResponseKey != "" && !strings.Contains(string(encoded), tt.wantResponseKey) {
				t.Fatalf("retained response is missing %q: %s", tt.wantResponseKey, encoded)
			}
			if tt.wantGap == "" {
				if operation.Runtime.MergeBlocked || len(operation.Runtime.Gaps) != 0 {
					t.Fatalf("supported provider dialect unexpectedly has gaps: %#v", operation.Runtime)
				}
				return
			}
			var gap *sourceContractGap
			for index := range operation.Runtime.Gaps {
				candidate := &operation.Runtime.Gaps[index]
				if candidate.Foundation == tt.wantGap {
					gap = candidate
					break
				}
			}
			if gap == nil || !operation.Runtime.MergeBlocked || !strings.Contains(gap.Location, operation.Source.Location) || !strings.Contains(gap.Reason, tt.wantTrace) {
				t.Fatalf("provider malformed-contract trace = %#v", operation.Runtime)
			}
		})
	}
}

func TestSourceImportKeepsDepthBoundFiniteAfterProviderIncrease(t *testing.T) {
	raw := sourceProviderNestedResponseDocument("3.1.0", "/pathological", "pathological", 65)
	_, err := importInlineSourceResultError(raw, defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "schema depth limit exceeded") {
		t.Fatalf("pathological provider schema error = %v, want finite depth refusal", err)
	}
}

func sourceProviderNestedResponseDocument(openAPI, path, operationID string, depth int) []byte {
	schema := sourceProviderNestedSchema(depth)
	return sourceProviderOperationDocument(openAPI, path, "get", operationID, map[string]any{
		"200": map[string]any{
			"description": "provider response",
			"content":     map[string]any{"application/json": map[string]any{"schema": schema}},
		},
	}, nil, sourceProviderRequiredPathParameterNames(path)...)
}

func sourceProviderRequiredPathParameterNames(path string) []string {
	parameters, err := sourcePathTemplateParameters(path)
	if err != nil {
		panic(err)
	}
	return parameters
}

func sourceProviderNestedSchema(depth int) map[string]any {
	schema := map[string]any{"type": "string"}
	for index := depth - 1; index >= 0; index-- {
		schema = map[string]any{
			"type":       "object",
			"properties": map[string]any{fmt.Sprintf("level_%02d", index): schema},
		}
	}
	return schema
}

func sourceProviderReferenceChainDocument(depth int) []byte {
	schemas := map[string]any{}
	for index := 0; index < depth; index++ {
		name := fmt.Sprintf("account_level_%02d", index)
		child := map[string]any{"type": "string"}
		if index+1 < depth {
			child = map[string]any{"$ref": "#/components/schemas/" + fmt.Sprintf("account_level_%02d", index+1)}
		}
		schemas[name] = map[string]any{"type": "object", "properties": map[string]any{"next": child}}
	}
	return sourceProviderOperationDocument("3.0.0", "/v1/account", "get", "GetAccount", map[string]any{
		"200": map[string]any{
			"description": "account response",
			"content":     map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/account_level_00"}}},
		},
	}, map[string]any{"schemas": schemas})
}

func sourceVercelPatternPropertiesDocument() []byte {
	return sourceProviderOperationDocument("3.0.0", "/api-keys", "post", "createApiKeys", map[string]any{
		"200": map[string]any{
			"description": "Information about the newly created API key.",
			"content":     map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/APIKey"}}},
		},
	}, map[string]any{"schemas": map[string]any{
		"APIKey": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"metadata": map[string]any{"type": "object", "patternProperties": map[string]any{"^(.*)$": map[string]any{}}},
			},
		},
	}})
}

func sourceDockerHubMalformedReferenceDocument() []byte {
	return sourceProviderOperationDocument("3.0.3", "/v2/auth/token", "post", "createToken", map[string]any{
		"401": map[string]any{
			"description": "provider response",
			"content":     map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/responses/team_repo"}}},
		},
	}, map[string]any{"responses": map[string]any{
		"team_repo": map[string]any{"description": "provider response"},
	}})
}

func sourceGitLabMissingPathParameterDocument() []byte {
	path := "/api/v4/groups/{id}/(-/)epics/{epic_iid}/issues/{epic_issue_id}"
	return sourceProviderOperationDocument("3.0.0", path, "put", "putApiV4GroupsIdDashEpicsEpicIidIssuesEpicIssueId", map[string]any{
		"200": map[string]any{"description": "updated"},
	}, nil, "id", "epic_iid")
}

func sourceProviderOperationDocument(openAPI, path, method, operationID string, responses map[string]any, components map[string]any, suppliedPathParameters ...string) []byte {
	parameters := make([]any, 0, len(suppliedPathParameters))
	for _, name := range suppliedPathParameters {
		parameters = append(parameters, map[string]any{"name": name, "in": "path", "required": true, "schema": map[string]any{"type": "string", "maxLength": 256}})
	}
	operation := map[string]any{"operationId": operationID, "responses": responses}
	if len(parameters) != 0 {
		operation["parameters"] = parameters
	}
	document := map[string]any{
		"openapi": openAPI,
		"info":    map[string]any{"title": "provider-derived", "version": "1"},
		"paths":   map[string]any{path: map[string]any{method: operation}},
	}
	if components != nil {
		document["components"] = components
	}
	raw, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	return raw
}

func importInlineSourceResult(t *testing.T, raw []byte, limits sourceImportLimits) sourceImportResult {
	t.Helper()
	result, err := importInlineSourceResultError(raw, limits)
	if err != nil {
		t.Fatalf("import inline source: %v", err)
	}
	return result
}

func importInlineSourceResultError(raw []byte, limits sourceImportLimits) (sourceImportResult, error) {
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/inline-openapi.json", raw)
	return importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return raw, nil }), limits)
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
	return sourceImportLock{SchemaVersion: 1, Connector: connector, Rest: sourceImportREST{sourceImportArtifact: sourceImportArtifact{SourceURL: sourceURL, SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(raw))}}}
}

func writeSourceImportRetainedFixture(t *testing.T, sourcesDir, connector string, artifact sourceImportArtifact, raw []byte) {
	t.Helper()
	manifest := map[string]any{
		"schema_version": 1,
		"connector":      connector,
		"artifacts": []any{map[string]any{
			"sha256":         artifact.SHA256,
			"bytes":          artifact.Bytes,
			"source_url":     artifact.SourceURL,
			"identity_query": artifact.IdentityQuery,
			"retrieved_at":   "2026-08-24T00:00:00Z",
			"license":        "undetermined",
			"terms":          "undetermined",
		}},
	}
	manifestRaw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal retained artifact fixture manifest: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sourcesDir, "artifacts"), 0o755); err != nil {
		t.Fatalf("create retained artifact fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, connector+"-retained-artifacts.json"), manifestRaw, 0o644); err != nil {
		t.Fatalf("write retained artifact fixture manifest: %v", err)
	}
	if raw == nil {
		return
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "artifacts", strings.ToLower(artifact.SHA256)+".artifact"), raw, 0o644); err != nil {
		t.Fatalf("write retained artifact fixture: %v", err)
	}
}

func sourceImportRetainedZIP(t *testing.T, name string, raw []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entry, err := archive.Create(name)
	if err != nil {
		t.Fatalf("create retained zip fixture entry: %v", err)
	}
	if _, err := entry.Write(raw); err != nil {
		t.Fatalf("write retained zip fixture entry: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close retained zip fixture: %v", err)
	}
	return buffer.Bytes()
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

type sourceImportV3FixtureDocument struct {
	ID            string
	Path          string
	Method        string
	OperationID   string
	Artifact      []byte
	ArtifactURL   string
	IdentityQuery *bool
	PublishedURL  string
}

func sourceImportV3FixtureLock(t *testing.T, connector string, documents []sourceImportV3FixtureDocument) []byte {
	t.Helper()
	sourceDocuments := make([]any, 0, len(documents))
	for _, document := range documents {
		artifactDigest := sha256.Sum256(document.Artifact)
		artifactURL := document.ArtifactURL
		if artifactURL == "" {
			artifactURL = "https://fixtures.polymetrics.invalid/" + document.ID + ".openapi.json"
		}
		publishedURL := document.PublishedURL
		if publishedURL == "" {
			publishedURL = "https://published.polymetrics.invalid/" + document.ID + "?slug=" + document.ID
		}
		artifact := map[string]any{
			"source_url": artifactURL,
			"sha256":     hex.EncodeToString(artifactDigest[:]),
			"bytes":      len(document.Artifact),
			"openapi":    "3.0.3",
		}
		if document.IdentityQuery != nil {
			artifact["identity_query"] = *document.IdentityQuery
		}
		method := document.Method
		if method == "" {
			method = "GET"
		}
		operationID := document.OperationID
		if operationID == "" {
			operationID = "shared"
		}
		sourceDocuments = append(sourceDocuments, map[string]any{
			"id":       document.ID,
			"artifact": artifact,
			"published_source": map[string]any{
				"source_url":  publishedURL,
				"capture_url": "https://fixtures.polymetrics.invalid/" + document.ID + ".capture.json",
				"sha256":      hex.EncodeToString(artifactDigest[:]),
				"bytes":       len(document.Artifact),
				"adapter":     "fixture-openapi-capture-v1",
			},
			"info_version": "1",
			"operations": []any{map[string]any{
				"id":              connector + ".rest." + document.ID + ".shared",
				"protocol":        "rest",
				"method":          method,
				"path":            document.Path,
				"operation_id":    operationID,
				"source_location": `paths["` + document.Path + `"].` + strings.ToLower(method),
			}},
		})
	}
	lock := map[string]any{
		"schema_version": 3,
		"connector":      connector,
		"rest": map[string]any{
			"retrieval":        "hermetic multi-document fixture capture",
			"openapi":          []any{"3.0.3"},
			"source_documents": sourceDocuments,
		},
		"counts": map[string]any{"rest": len(documents), "graphql_query": 0, "graphql_mutation": 0, "total": len(documents)},
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("encode v3 fixture lock: %v", err)
	}
	return raw
}

func TestSourceImportVersion3FetchesDeclaredIdentityQuery(t *testing.T) {
	t.Parallel()
	identityQuery := true
	document := sourceImportV3FixtureDocument{
		ID:            "identity",
		Path:          "/identity",
		Artifact:      []byte(`{"openapi":"3.0.3","info":{"title":"identity","version":"1"},"paths":{"/identity":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`),
		ArtifactURL:   "https://fixtures.polymetrics.invalid/identity.openapi.json?version=2026-08-01",
		IdentityQuery: &identityQuery,
	}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse v3 identity-query lock: %v", err)
	}
	lookup := batchArtifactLookupIPAddr(func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	requests := 0
	var fetched *url.URL
	fetcher := httpSourceImportFetcher{
		limits: defaultSourceImportLimits(),
		lookup: lookup,
		client: &http.Client{Transport: sourceImportRoundTripFunc(func(request *http.Request) (*http.Response, error) {
			requests++
			fetched = request.URL
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(document.Artifact)),
				Request:    request,
			}, nil
		})},
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 identity-query lock: %v", err)
	}
	if len(result.Operations) != 1 || result.Operations[0].SourceID != "fixture.rest.identity.shared" {
		t.Fatalf("identity-query import result = %#v", result.Operations)
	}
	gotQuery := ""
	if fetched != nil {
		gotQuery = fetched.Query().Get("version")
	}
	if requests != 1 || fetched == nil || fetched.String() != document.ArtifactURL || gotQuery != "2026-08-01" {
		t.Fatalf("identity artifact request count/URL/query = %d/%v/%q", requests, fetched, gotQuery)
	}
}

func TestSourceImportVersion3LeavesCaptureQueryAsProvenanceOnly(t *testing.T) {
	t.Parallel()
	document := sourceImportV3FixtureDocument{
		ID:           "capture",
		Path:         "/capture",
		Artifact:     []byte(`{"openapi":"3.0.3","info":{"title":"capture","version":"1"},"paths":{"/capture":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`),
		PublishedURL: "https://published.polymetrics.invalid/capture?slug=rotating-capture",
	}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse v3 capture-query lock: %v", err)
	}
	var fetched []string
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		fetched = append(fetched, sourceURL)
		return document.Artifact, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 capture-query lock: %v", err)
	}
	if len(result.Operations) != 1 || !reflect.DeepEqual(fetched, []string{"https://fixtures.polymetrics.invalid/capture.openapi.json"}) {
		t.Fatalf("capture provenance import operations/requests = %#v/%#v", result.Operations, fetched)
	}
}

func TestSourceImportVersion3AbsentOrFalseIdentityQueryProjectsIdentically(t *testing.T) {
	t.Parallel()
	identityQuery := false
	document := sourceImportV3FixtureDocument{
		ID:       "legacy",
		Path:     "/legacy",
		Artifact: []byte(`{"openapi":"3.0.3","info":{"title":"legacy","version":"1"},"paths":{"/legacy":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`),
	}
	absentLock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse absent identity declaration: %v", err)
	}
	document.IdentityQuery = &identityQuery
	falseLock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
	if err != nil {
		t.Fatalf("parse false identity declaration: %v", err)
	}
	fetch := sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return document.Artifact, nil })
	absentResult, err := importSourceLockResult(context.Background(), absentLock, fetch, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import absent identity declaration: %v", err)
	}
	falseResult, err := importSourceLockResult(context.Background(), falseLock, fetch, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import false identity declaration: %v", err)
	}
	absentProjection, err := marshalSourceImportResult(absentResult)
	if err != nil {
		t.Fatalf("marshal absent identity declaration: %v", err)
	}
	falseProjection, err := marshalSourceImportResult(falseResult)
	if err != nil {
		t.Fatalf("marshal false identity declaration: %v", err)
	}
	if !bytes.Equal(absentProjection, falseProjection) {
		t.Fatalf("absent and false identity-query projections differ\nabsent: %s\nfalse: %s", absentProjection, falseProjection)
	}
}

func TestSourceImportVersion3RejectsUnsafeIdentityArtifactQueries(t *testing.T) {
	t.Parallel()
	identityQuery := true
	excessKeys := make([]string, maxSourceImportPublishedQueryKeys+1)
	for index := range excessKeys {
		excessKeys[index] = fmt.Sprintf("part%d=value", index)
	}
	cases := []struct {
		name string
		url  string
	}{
		{name: "credential-shaped key", url: "https://fixtures.polymetrics.invalid/identity.openapi.json?api_key=value"},
		{name: "oversized query", url: "https://fixtures.polymetrics.invalid/identity.openapi.json?version=" + strings.Repeat("a", maxSourceImportPublishedQueryBytes)},
		{name: "too many keys", url: "https://fixtures.polymetrics.invalid/identity.openapi.json?" + strings.Join(excessKeys, "&")},
		{name: "missing query", url: "https://fixtures.polymetrics.invalid/identity.openapi.json"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			document := sourceImportV3FixtureDocument{
				ID:            "identity",
				Path:          "/identity",
				Artifact:      []byte(`{"openapi":"3.0.3","info":{"title":"identity","version":"1"},"paths":{"/identity":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`),
				ArtifactURL:   tc.url,
				IdentityQuery: &identityQuery,
			}
			_, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
			if err == nil || !strings.Contains(err.Error(), "identity artifact query") {
				t.Fatalf("unsafe identity artifact URL %q error = %v", tc.url, err)
			}
		})
	}
}

func TestSourceImportIdentityQueryRequiresV3RESTDocument(t *testing.T) {
	t.Parallel()
	artifact := []byte(`{"openapi":"3.0.3","info":{"title":"identity","version":"1"},"paths":{"/identity":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)
	digest := sha256.Sum256(artifact)
	identityArtifact := sourceImportArtifact{
		SourceURL:     "https://fixtures.polymetrics.invalid/identity.openapi.json?version=1",
		SHA256:        hex.EncodeToString(digest[:]),
		Bytes:         int64(len(artifact)),
		OpenAPI:       "3.0.3",
		IdentityQuery: true,
	}
	legacyRaw, err := json.Marshal(sourceImportLock{
		SchemaVersion: 2,
		Connector:     "fixture",
		Rest:          sourceImportREST{sourceImportArtifact: identityArtifact},
	})
	if err != nil {
		t.Fatalf("marshal legacy identity-query lock: %v", err)
	}
	if _, err := parseSourceImportLock(legacyRaw, "fixture"); err == nil || !strings.Contains(err.Error(), "v3 REST source document") {
		t.Fatalf("legacy identity-query lock error = %v", err)
	}

	identityQuery := true
	document := sourceImportV3FixtureDocument{ID: "identity", Path: "/identity", Artifact: artifact, ArtifactURL: identityArtifact.SourceURL, IdentityQuery: &identityQuery}
	var v3Lock map[string]any
	if err := json.Unmarshal(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), &v3Lock); err != nil {
		t.Fatalf("decode v3 identity-query lock: %v", err)
	}
	v3Lock["graphql"] = map[string]any{
		"source_url":     identityArtifact.SourceURL,
		"sha256":         identityArtifact.SHA256,
		"bytes":          identityArtifact.Bytes,
		"identity_query": true,
	}
	v3Raw, err := json.Marshal(v3Lock)
	if err != nil {
		t.Fatalf("marshal v3 GraphQL identity-query lock: %v", err)
	}
	if _, err := parseSourceImportLock(v3Raw, "fixture"); err == nil || !strings.Contains(err.Error(), "v3 REST source document") {
		t.Fatalf("v3 GraphQL identity-query lock error = %v", err)
	}
}

func TestSourceImportIdentityQueryRetainsArtifactURLGuards(t *testing.T) {
	t.Parallel()
	identityQuery := true
	artifact := []byte(`{"openapi":"3.0.3","info":{"title":"identity","version":"1"},"paths":{"/identity":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)
	cases := []struct {
		name string
		url  string
	}{
		{name: "non HTTPS", url: "http://fixtures.polymetrics.invalid/identity.openapi.json?version=1"},
		{name: "userinfo", url: "https://user@fixtures.polymetrics.invalid/identity.openapi.json?version=1"},
		{name: "fragment", url: "https://fixtures.polymetrics.invalid/identity.openapi.json?version=1#fragment"},
		{name: "non-public host", url: "https://localhost/identity.openapi.json?version=1"},
		{name: "private literal", url: "https://127.0.0.1/identity.openapi.json?version=1"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			document := sourceImportV3FixtureDocument{ID: "identity", Path: "/identity", Artifact: artifact, ArtifactURL: tc.url, IdentityQuery: &identityQuery}
			_, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture")
			if err == nil || !strings.Contains(err.Error(), "invalid artifact") {
				t.Fatalf("identity artifact URL %q error = %v", tc.url, err)
			}
		})
	}

	digest := sha256.Sum256(artifact)
	lockedArtifact := sourceImportArtifact{
		SourceURL:     "https://fixtures.polymetrics.invalid/identity.openapi.json?version=1",
		SHA256:        hex.EncodeToString(digest[:]),
		Bytes:         int64(len(artifact)),
		OpenAPI:       "3.0.3",
		IdentityQuery: true,
	}
	called := false
	fetcher := httpSourceImportFetcher{
		limits: defaultSourceImportLimits(),
		lookup: batchArtifactLookupIPAddr(func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("169.254.169.254")}}, nil
		}),
		client: &http.Client{Transport: sourceImportRoundTripFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return nil, fmt.Errorf("unexpected request")
		})},
	}
	if _, err := fetcher.FetchArtifact(context.Background(), lockedArtifact); err == nil || !strings.Contains(err.Error(), "not public") {
		t.Fatalf("private resolved identity artifact error = %v", err)
	}
	if called {
		t.Fatal("private resolved identity artifact reached the HTTP transport")
	}
}

type sourceImportRoundTripFunc func(*http.Request) (*http.Response, error)

func (f sourceImportRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSourceImportVersion3ImportsDocumentOwnedProvenanceAndDuplicateProviderIDs(t *testing.T) {
	t.Parallel()
	documents := []sourceImportV3FixtureDocument{
		{ID: "alpha", Path: "/alpha", Artifact: []byte(`{"openapi":"3.0.3","info":{"title":"alpha","version":"1"},"paths":{"/alpha":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)},
		{ID: "bravo", Path: "/bravo", Artifact: []byte(`{"openapi":"3.0.3","info":{"title":"bravo","version":"1"},"paths":{"/bravo":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)},
	}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", documents), "fixture")
	if err != nil {
		t.Fatalf("parse v3 fixture lock: %v", err)
	}
	fetched := map[string]int{}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		fetched[sourceURL]++
		for _, document := range documents {
			if sourceURL == "https://fixtures.polymetrics.invalid/"+document.ID+".openapi.json" {
				return document.Artifact, nil
			}
		}
		return nil, fmt.Errorf("unexpected fixture source URL %q", sourceURL)
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 fixture lock: %v", err)
	}
	if len(result.Operations) != 2 {
		t.Fatalf("v3 imported operations = %d, want 2", len(result.Operations))
	}
	if fetched["https://published.polymetrics.invalid/alpha?slug=alpha"] != 0 || fetched["https://published.polymetrics.invalid/bravo?slug=bravo"] != 0 {
		t.Fatalf("importer fetched a published citation: %#v", fetched)
	}
	byID := map[string]sourceOperationDescriptor{}
	for _, operation := range result.Operations {
		byID[operation.SourceID] = operation
	}
	for _, document := range documents {
		identity := "fixture.rest." + document.ID + ".shared"
		operation, ok := byID[identity]
		if !ok {
			t.Fatalf("v3 operation %q is absent: %#v", identity, result.Operations)
		}
		if operation.ProviderOperationID != "shared" || operation.Path != document.Path {
			t.Fatalf("v3 operation identity = %#v", operation)
		}
		encoded, marshalErr := json.Marshal(operation)
		if marshalErr != nil {
			t.Fatalf("marshal v3 operation: %v", marshalErr)
		}
		for _, want := range []string{`"document_id":"` + document.ID + `"`, `"published_url":"https://published.polymetrics.invalid/` + document.ID + `?slug=` + document.ID + `"`, `"published_capture_url":"https://fixtures.polymetrics.invalid/` + document.ID + `.capture.json"`} {
			if !strings.Contains(string(encoded), want) {
				t.Fatalf("v3 provenance omitted %s from %s", want, encoded)
			}
		}
	}
	marshaled, err := marshalSourceImportResult(result)
	if err != nil {
		t.Fatalf("marshal v3 descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(marshaled, &descriptor); err != nil {
		t.Fatalf("decode v3 descriptor: %v", err)
	}
	if descriptor.SchemaVersion != 3 || !strings.Contains(string(marshaled), `"document_id": "alpha"`) || !strings.Contains(string(marshaled), `"published_url": "https://published.polymetrics.invalid/bravo?slug=bravo"`) {
		t.Fatalf("v3 descriptor wire provenance = %s", marshaled)
	}
}

func TestSourceImportVersion3RejectsMissingOrDriftedDocument(t *testing.T) {
	t.Parallel()
	documents := []sourceImportV3FixtureDocument{
		{ID: "alpha", Path: "/alpha", Artifact: []byte(`{"openapi":"3.0.3","info":{"title":"alpha","version":"1"},"paths":{"/alpha":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)},
		{ID: "bravo", Path: "/bravo", Artifact: []byte(`{"openapi":"3.0.3","info":{"title":"bravo","version":"1"},"paths":{"/bravo":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)},
	}
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", documents), "fixture")
	if err != nil {
		t.Fatalf("parse v3 fixture lock: %v", err)
	}
	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		if strings.Contains(sourceURL, "bravo") {
			return nil, fmt.Errorf("fixture source is missing")
		}
		return documents[0].Artifact, nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "fixture source is missing") {
		t.Fatalf("missing v3 document error = %v", err)
	}

	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return []byte(`{"openapi":"3.0.3","info":{"title":"drift","version":"1"},"paths":{}}`), nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "source-lock refresh required") {
		t.Fatalf("drifted v3 document error = %v", err)
	}
}

func TestSourceImportVersion3RejectsUnsafePublishedCitationQuery(t *testing.T) {
	t.Parallel()
	document := sourceImportV3FixtureDocument{
		ID:           "alpha",
		Path:         "/alpha",
		Artifact:     []byte(`{"openapi":"3.0.3","info":{"title":"alpha","version":"1"},"paths":{"/alpha":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`),
		PublishedURL: "https://published.polymetrics.invalid/alpha?access_token=value",
	}
	if _, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{document}), "fixture"); err == nil || !strings.Contains(err.Error(), "published source URL") {
		t.Fatalf("unsafe published citation query error = %v", err)
	}
}

func TestSourceImportVersion3SynchronizesDuplicateArtifactDigests(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"openapi":"3.0.3","info":{"title":"fixture","version":"1"},"paths":{}}`)
	digest := sha256.Sum256(raw)
	artifact := sourceImportArtifact{SourceURL: "https://fixtures.polymetrics.invalid/shared.openapi.json", SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(raw)), OpenAPI: "3.0.3"}
	documents := []sourceImportRESTDocument{{ID: "alpha", Artifact: artifact}, {ID: "bravo", Artifact: artifact}}
	var mu sync.Mutex
	calls := 0
	got, err := fetchSourceImportV3Documents(context.Background(), documents, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return raw, nil
	}))
	if err != nil {
		t.Fatalf("fetch duplicate digest documents: %v", err)
	}
	if calls != 1 || !bytes.Equal(got["alpha"], raw) || !bytes.Equal(got["bravo"], raw) {
		t.Fatalf("duplicate digest synchronization calls=%d documents=%#v", calls, got)
	}
}

func TestSourceImportPreservesFrozenGitHubArtifacts(t *testing.T) {
	t.Parallel()
	checks := []struct {
		path   string
		bytes  int
		sha256 string
	}{
		{path: filepath.Join("..", "..", "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"), bytes: 3420025, sha256: "79f6eaf203394aabe7d2558d0f87a8100a7a084b2e39bd264f7773f235acf2c8"},
		{path: filepath.Join("..", "..", "internal", "connectors", "defs", "github", "sources", "github-operation-descriptor.json"), bytes: 43355704, sha256: "69b23e5146480eb67f10c3ba65b45fc6fac466cfb8ae244953b19b8373f10062"},
		{path: filepath.Join("..", "..", ".planning", "phases", "github-parity-extract-r1", "GITHUB-COMBINED-OPERATION-LEDGER.json"), bytes: 2553169, sha256: "b2bc566e8c844fcf307b37000c8bf3d482a9da932aff5c8d375f54b6a6ed3391"},
	}
	for _, check := range checks {
		check := check
		t.Run(filepath.Base(check.path), func(t *testing.T) {
			raw, err := os.ReadFile(check.path)
			if err != nil {
				t.Fatalf("read frozen artifact: %v", err)
			}
			digest := sha256.Sum256(raw)
			if len(raw) != check.bytes || hex.EncodeToString(digest[:]) != check.sha256 {
				t.Fatalf("frozen artifact %s = %d/%s, want %d/%s", check.path, len(raw), hex.EncodeToString(digest[:]), check.bytes, check.sha256)
			}
		})
	}
}
