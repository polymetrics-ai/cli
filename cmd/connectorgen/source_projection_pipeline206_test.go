package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceProjection206SourceOnlyPublishedRead(t *testing.T) {
	sourceProjectionPublishedRead206(t, "", false)
}

func TestSourceProjection206DefaultCommandName(t *testing.T) {
	sourceProjectionPublishedRead206(t, "direct_camel_identity", false)
}

func TestSourceProjection206ArchivedOperationIdentity(t *testing.T) {
	t.Run("envelope_identity", func(t *testing.T) { sourceProjectionPublishedRead206(t, "envelope_identity", false) })
	t.Run("contradictory_identity", func(t *testing.T) { sourceProjectionPublishedRead206(t, "contradictory_identity", true) })
}

func TestSourceProjection206RetainedRawDocument(t *testing.T) {
	sourceProjectionPublishedRead206(t, "raw_schema3", false)
}

func TestSourceProjection206DirectInputs(t *testing.T) {
	for _, variant := range []string{"inherited_path", "optional_query_absent", "optional_query_present"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+variant, false) })
	}
}

func TestSourceProjection206SourceInputs(t *testing.T) {
	for _, variant := range []string{"inherited_path", "operation_override", "optional_query_absent", "optional_query_present", "missing_path"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, variant, false) })
	}
}

func TestSourceProjection206UnsupportedSourceSemantics(t *testing.T) {
	for _, variant := range []string{"protocol", "request_body", "response_media", "response_status"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, variant, true) })
	}
}

func TestSourceProjection206ScalarInputs(t *testing.T) {
	for _, variant := range []string{"query_integer", "query_bigint", "query_decimal", "query_boolean"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+variant, false) })
	}
}

func TestSourceProjectionTypedQuery255(t *testing.T) {
	for _, variant := range []string{"typed_query_map255", "typed_query_repeated255", "typed_query_bracket255", "typed_query_default255"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+variant, false) })
	}
}

func TestSourceProjectionTypedQueryShapes255(t *testing.T) {
	for _, variant := range []string{"typed_query_indexed255", "typed_query_dynamic255", "typed_query_empty255"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+variant, false) })
	}
}

func TestSourceProjectionTypedQueryAuthority255(t *testing.T) {
	for _, variant := range []string{"missing_evidence", "wrong_span", "unknown_parameter", "explicit_style", "unknown_mode"} {
		t.Run(variant, func(t *testing.T) {
			sourceProjectionPublishedRead206(t, "direct_typed_query_bracket255|"+variant, true)
		})
	}
}

func sourceProjectionPublishedRead206(t *testing.T, variant string, wantRefusal bool) {
	t.Helper()
	direct := strings.HasPrefix(variant, "direct_")
	variant = strings.TrimPrefix(variant, "direct_")
	variant, selectedBad, selectBad := strings.Cut(variant, "|")
	documentaryQuery := variant == "typed_query_bracket255" || variant == "typed_query_indexed255" || variant == "typed_query_dynamic255" || variant == "typed_query_empty255"
	queryMode, queryEmpty := "bracket_repeated", "omit"
	if variant == "typed_query_indexed255" {
		queryMode = "indexed"
	}
	if variant == "typed_query_dynamic255" {
		queryMode = "brackets"
	}
	if variant == "typed_query_empty255" {
		queryEmpty = "empty"
	}

	var requests []string
	var requestMu sync.Mutex
	requestSnapshot := func() []string {
		requestMu.Lock()
		defer requestMu.Unlock()
		return append([]string(nil), requests...)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMu.Lock()
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		requestMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"alpha"},{"id":"bravo"},{"id":"charlie"}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	connector := filepath.Join(root, "acme")
	if err := os.MkdirAll(filepath.Join(connector, "sources"), 0o700); err != nil {
		t.Fatal(err)
	}
	source := []byte(`{"schema_version":2,"connector":"acme","rest":{"operations":[{"id":"fixture.widgets","method":"GET","path":"/widgets","protocol":"rest","source_operation":{"operationId":"widgets","responses":{"200":{"description":"Page","content":{"application/json":{"schema":{"type":"object","required":["data"],"properties":{"data":{"type":"array","items":{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}}}}}}}}}}}]},"counts":{"total":1}}`)
	switch variant {
	case "camel_identity":
		source = bytes.Replace(source, []byte(`"operationId":"widgets"`), []byte(`"operationId":"listWidgets"`), 1)
	case "envelope_identity":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"rest","operation_id":"widgets"`), 1)
		source = bytes.Replace(source, []byte(`"operationId":"widgets",`), nil, 1)
	case "contradictory_identity":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"rest","operation_id":"different"`), 1)
	case "protocol":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"graphql"`), 1)
	case "request_body":
		source = bytes.Replace(source, []byte(`"operationId":"widgets"`), []byte(`"operationId":"widgets","requestBody":{"required":true,"content":{"application/json":{"schema":{"type":"object","required":["filter"],"properties":{"filter":{"type":"string"}}}}}}`), 1)
	case "response_media":
		source = bytes.Replace(source, []byte(`"application/json"`), []byte(`"application/xml"`), 1)
	case "response_status":
		source = bytes.Replace(source, []byte(`"200"`), []byte(`"201"`), 1)
	}
	inputVariant := strings.HasPrefix(variant, "typed_query_") || strings.HasPrefix(variant, "query_") || variant == "inherited_path" || variant == "operation_override" || variant == "optional_query_absent" || variant == "optional_query_present" || variant == "missing_path"
	expectedPath, expectedWire := "/widgets", "GET /widgets"
	runtimeConfig := map[string]string{}
	expectedName := "widgets"
	expectedCommand := "widgets"
	if variant == "camel_identity" {
		expectedName = "list_widgets"
		expectedCommand = "list-widgets"
	}
	if inputVariant {
		var envelope map[string]any
		if err := json.Unmarshal(source, &envelope); err != nil {
			t.Fatal(err)
		}
		row := envelope["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
		operation := row["source_operation"].(map[string]any)
		pathParameter := func(name string) map[string]any {
			return map[string]any{"in": "path", "name": name, "required": true, "schema": map[string]any{"type": "string"}}
		}
		if variant == "inherited_path" || variant == "operation_override" || variant == "missing_path" {
			row["path"] = "/namespaces/{namespace}/repositories/{repository}/widgets"
			operation["path_parameters"] = []any{pathParameter("namespace"), pathParameter("repository")}
			expectedPath = "/namespaces/{{ config.namespace }}/repositories/{{ config.repository }}/widgets"
			expectedWire = "GET /namespaces/acme/repositories/warehouse/widgets"
			runtimeConfig["namespace"], runtimeConfig["repository"] = "acme", "warehouse"
			if variant == "missing_path" {
				delete(runtimeConfig, "repository")
			}
			if variant == "operation_override" {
				override := pathParameter("namespace")
				override["schema"] = map[string]any{"type": "string", "enum": []any{"acme"}}
				operation["parameters"] = []any{override}
			}
		} else if strings.HasPrefix(variant, "typed_query_") {
			parameter := map[string]any{"in": "query", "name": "filter", "required": true, "style": "deepObject", "explode": true,
				"schema": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"owner", "tag"}, "properties": map[string]any{"owner": map[string]any{"type": "string"}, "tag": map[string]any{"type": "string"}}}}
			runtimeConfig["filter"] = `{"tag":"x&y", "owner":"a+b c"}`
			expectedWire = "GET /widgets?filter%5Bowner%5D=a%2Bb+c&filter%5Btag%5D=x%26y"
			if variant == "typed_query_repeated255" || variant == "typed_query_bracket255" {
				parameter["style"] = "form"
				parameter["schema"] = map[string]any{"type": "array", "minItems": 1, "maxItems": 3, "items": map[string]any{"type": "string"}}
				runtimeConfig["filter"] = `["a+b c","x&y","a+b c"]`
				expectedWire = "GET /widgets?filter=a%2Bb+c&filter=x%26y&filter=a%2Bb+c"
			}
			if variant == "typed_query_bracket255" {
				delete(parameter, "style")
				delete(parameter, "explode")
				if selectedBad == "explicit_style" {
					parameter["style"] = "form"
					parameter["explode"] = true
				}
				expectedWire = "GET /widgets?filter%5B%5D=a%2Bb+c&filter%5B%5D=x%26y&filter%5B%5D=a%2Bb+c"
			}
			if variant == "typed_query_indexed255" {
				parameter["schema"] = map[string]any{"type": "array", "minItems": 1, "maxItems": 2, "items": map[string]any{"type": "object", "required": []string{"key", "value"}, "additionalProperties": false, "properties": map[string]any{"key": map[string]any{"type": "string"}, "value": map[string]any{"type": "integer"}}}}
				runtimeConfig["filter"] = `[{"key":"alpha","value":9007199254740993},{"key":"beta","value":2}]`
				expectedWire = "GET /widgets?filter%5B0%5D%5Bkey%5D=alpha&filter%5B0%5D%5Bvalue%5D=9007199254740993&filter%5B1%5D%5Bkey%5D=beta&filter%5B1%5D%5Bvalue%5D=2"
			}
			if variant == "typed_query_dynamic255" {
				parameter["schema"] = map[string]any{"type": "object", "additionalProperties": false, "patternProperties": map[string]any{"^[a-z]+$": map[string]any{"type": "integer"}}}
				runtimeConfig["filter"] = `{"alpha":9007199254740993,"beta":2}`
				expectedWire = "GET /widgets?filter%5Balpha%5D=9007199254740993&filter%5Bbeta%5D=2"
			}
			if variant == "typed_query_empty255" {
				parameter["schema"] = map[string]any{"type": "array", "maxItems": 3, "items": map[string]any{"type": "string"}}
				runtimeConfig["filter"] = `[]`
				expectedWire = "GET /widgets?filter%5B%5D="
			}
			if documentaryQuery {
				delete(parameter, "style")
				delete(parameter, "explode")
				if selectedBad == "explicit_style" {
					parameter["style"] = "form"
					parameter["explode"] = true
				}
			}
			if variant == "typed_query_default255" {
				parameter["required"] = false
				parameter["schema"].(map[string]any)["default"] = map[string]any{"tag": "x&y", "owner": "a+b c"}
				delete(runtimeConfig, "filter")
			}
			operation["parameters"] = []any{parameter}
		} else if strings.HasPrefix(variant, "query_") {
			schema := map[string]any{"type": "integer", "minimum": json.Number("1"), "maximum": json.Number("5")}
			runtimeConfig["page_size"] = "3"
			expectedWire = "GET /widgets?page_size=3"
			if variant == "query_bigint" {
				schema["minimum"] = json.Number("9007199254740993")
				schema["maximum"] = json.Number("9007199254740995")
				runtimeConfig["page_size"] = "9007199254740994"
				expectedWire = "GET /widgets?page_size=9007199254740994"
			}
			if variant == "query_decimal" {
				schema["type"] = "number"
				schema["minimum"] = json.Number("0.10000000000000001")
				schema["maximum"] = json.Number("0.10000000000000003")
				runtimeConfig["page_size"] = "0.10000000000000002"
				expectedWire = "GET /widgets?page_size=0.10000000000000002"
			}
			if variant == "query_boolean" {
				schema = map[string]any{"type": "boolean"}
				runtimeConfig["page_size"] = "false"
				expectedWire = "GET /widgets?page_size=false"
			}
			operation["parameters"] = []any{map[string]any{"in": "query", "name": "page_size", "required": false, "schema": schema}}
		} else {
			operation["parameters"] = []any{map[string]any{"in": "query", "name": "owner-id", "required": false, "schema": map[string]any{"type": "string"}}}
			if variant == "optional_query_present" {
				runtimeConfig["owner-id"] = "a+b c"
				expectedWire = "GET /widgets?owner-id=a%2Bb+c"
			}
		}
		var err error
		source, err = json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !json.Valid(source) {
		t.Fatal("invalid retained test fixture")
	}
	if variant == "raw_schema3" {
		var envelope map[string]any
		if err := json.Unmarshal(source, &envelope); err != nil {
			t.Fatal(err)
		}
		row := envelope["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
		operation := row["source_operation"]
		delete(row, "source_operation")
		artifact, err := json.Marshal(map[string]any{"openapi": "3.0.0", "info": map[string]string{"title": "Independent fixture", "version": "1"}, "paths": map[string]any{"/widgets": map[string]any{"get": operation}}})
		if err != nil {
			t.Fatal(err)
		}
		artifactDir := filepath.Join(connector, "sources", "artifacts")
		if err := os.Mkdir(artifactDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifactDir, sourceBytesHash(artifact)+".artifact"), artifact, 0o600); err != nil {
			t.Fatal(err)
		}
		envelope["schema_version"] = 3
		envelope["rest"] = map[string]any{"source_documents": []any{map[string]any{"id": "raw", "content_type": "application/json", "artifact": map[string]any{"sha256": sourceBytesHash(artifact), "bytes": len(artifact), "openapi": "3.0.0"}, "operations": []any{row}}}}
		source, err = json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(connector, "sources", "operations.json"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	lock := minimalVNextLockForTest()
	lock.Operations, lock.Schemas = nil, nil
	if direct {
		lock.Lanes["direct_read"] = "implemented"
	}
	lock.HTTP = json.RawMessage(strings.Replace(string(lock.HTTP), "https://api.acme.example", server.URL, 1))
	pointer := "/data"
	lock.SourceProjection = &vNextSourceProjection{
		Version:     1,
		Inventories: []vNextSourceProjectionInventory{{ID: "primary", Path: "sources/operations.json", SHA256: sourceBytesHash(source), Bytes: int64(len(source))}},
		Semantics: []vNextSourceProjectionSemantic{{Source: vNextSourceProjectionKey{Inventory: "primary", ID: "fixture.widgets"}, Collection: &vNextSourceProjectionCollection{
			Records: vNextSourceCoordinate{Response: &vNextSourceResponseCoordinate{Status: "200", Media: "application/json", Pointer: &pointer}}, PrimaryKey: []string{"/id"},
		}}},
	}
	if strings.HasPrefix(variant, "document_") {
		documentary := []byte("Doc: widgets returns three records. An example is not membership.")
		if err := os.WriteFile(filepath.Join(connector, "sources", "guide.txt"), documentary, 0600); err != nil {
			t.Fatal(err)
		}
		lock.SourceProjection.Documents = []vNextSourceProjectionDocument{{ID: "guide", Path: "sources/guide.txt", SHA256: sourceBytesHash(documentary), Bytes: int64(len(documentary)), Format: "text", SourceURL: "https://docs.example.test/widgets", Revision: "fixture1", RetrievedAt: "2026-09-09T00:00:00Z"}}
		span := &sourceFactSpan{Offset: 5, Length: 7, SHA256: sourceBytesHash([]byte("widgets"))}
		lock.SourceProjection.Semantics[0].Evidence = []sourceFactRef{{DocumentID: "guide", Span: span}}
		switch variant {
		case "document_bad_pin":
			lock.SourceProjection.Documents[0].SHA256 = strings.Repeat("a", 64)
		case "document_bad_span":
			span.SHA256 = strings.Repeat("a", 64)
		case "document_outside_span":
			span.Offset = int64(len(documentary))
		case "document_conflicting_selector":
			lock.SourceProjection.Semantics[0].Evidence[0].Pointer = "/fake"
		}
	}
	if variant == "response_media" {
		lock.SourceProjection.Semantics[0].Collection.Records.Response.Media = "application/xml"
	}
	if variant == "response_status" {
		lock.SourceProjection.Semantics[0].Collection.Records.Response.Status = "201"
	}
	if documentaryQuery {
		evidence := []byte("Fixture query dialect is " + queryMode + "; empty array rule is " + queryEmpty + "; null is rejected.")
		if err := os.WriteFile(filepath.Join(connector, "sources", "query-policy.txt"), evidence, 0600); err != nil {
			t.Fatal(err)
		}
		lock.SourceProjection.Documents = []vNextSourceProjectionDocument{{ID: "query-policy", Path: "sources/query-policy.txt", SHA256: sourceBytesHash(evidence), Bytes: int64(len(evidence)), Format: "text", SourceURL: "https://docs.example.test/query", Revision: "fixture1", RetrievedAt: "2026-09-09T00:00:00Z"}}
		lock.SourceProjection.Semantics[0].Evidence = []sourceFactRef{{DocumentID: "query-policy", Span: &sourceFactSpan{Offset: 0, Length: int64(len(evidence)), SHA256: sourceBytesHash(evidence)}}}
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if documentaryQuery {
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		projection := document["source_projection"].(map[string]any)
		semantic := projection["semantics"].([]any)[0].(map[string]any)
		semantic["query_encoding"] = map[string]any{"filter": map[string]any{"mode": queryMode, "empty_array": queryEmpty, "null": "reject"}}
		switch selectedBad {
		case "missing_evidence":
			delete(semantic, "evidence")
		case "wrong_span":
			semantic["evidence"].([]any)[0].(map[string]any)["span"].(map[string]any)["sha256"] = strings.Repeat("a", 64)
		case "unknown_parameter":
			semantic["query_encoding"] = map[string]any{"other": map[string]any{"mode": "bracket_repeated", "empty_array": "omit", "null": "reject"}}
		case "unknown_mode":
			semantic["query_encoding"].(map[string]any)["filter"].(map[string]any)["mode"] = "automatic"
		}

		raw, err = json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(connector, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &stdout, &stderr)
	if wantRefusal {
		if code == 0 {
			t.Fatal("unsupported retained source semantics published as implemented execution")
		}
		if len(requestSnapshot()) != 0 {
			t.Fatal("source refusal sent provider request")
		}
		if _, err := os.Lstat(filepath.Join(connector, vNextPublicationCurrentFile)); !os.IsNotExist(err) {
			t.Fatalf("source refusal changed CURRENT: %v", err)
		}
		return
	}
	if code != 0 {
		t.Fatalf("source-only lock-render refused before admitted execution: code=%d error=%s", code, stderr.String())
	}
	if len(requestSnapshot()) != 0 {
		t.Fatal("authoring contacted provider")
	}
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()
	execution := map[string][]byte{}
	for _, name := range handle.Files() {
		switch name {
		case "manifest.json", "provenance.json", "atlas.json", "index.json", "proof.json":
			continue
		}
		execution[name], err = handle.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := engine.Load(newVNextExecutionFS("acme", execution), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if variant == "typed_query_map255" {
		for _, field := range []string{"format", "values", "minimum", "maximum", "min_items", "max_items"} {
			t.Run("contradictory_flag_"+field, func(t *testing.T) {
				var surface map[string]any
				if err := json.Unmarshal(execution["cli_surface.json"], &surface); err != nil {
					t.Fatal(err)
				}
				for _, command := range surface["commands"].([]any) {
					for _, rawFlag := range command.(map[string]any)["flags"].([]any) {
						flag := rawFlag.(map[string]any)
						if flag["name"] != "filter" {
							continue
						}
						switch field {
						case "format":
							flag[field] = "date-time"
						case "values":
							flag[field] = []string{"contradiction"}
						default:
							flag[field] = 1
						}
					}
				}
				raw, err := json.Marshal(surface)
				if err != nil {
					t.Fatal(err)
				}
				candidate := map[string][]byte{}
				for name, value := range execution {
					candidate[name] = value
				}
				candidate["cli_surface.json"] = raw
				if _, err := engine.Load(newVNextExecutionFS("acme", candidate), "acme"); err == nil {
					t.Fatalf("contradictory structured flag %s admitted", field)
				}
				if len(requestSnapshot()) != 0 {
					t.Fatal("admission sent HTTP")
				}
			})
		}
	}
	if len(bundle.Streams) != 1 || bundle.Streams[0].Name != expectedName || bundle.Streams[0].Path != expectedPath {
		t.Fatalf("source route/identity not preserved: %+v", bundle.Streams)
	}
	if variant == "operation_override" {
		var schema map[string]any
		contract := bundle.Streams[0].RequestInputs
		if contract == nil {
			t.Fatal("missing selected request input contract")
		}
		if err := json.Unmarshal(execution[contract.Schema], &schema); err != nil {
			t.Fatal(err)
		}
		pathSchema := schema["properties"].(map[string]any)["path"].(map[string]any)
		if !reflect.DeepEqual(pathSchema["properties"].(map[string]any)["namespace"].(map[string]any)["enum"], []any{"acme"}) {
			t.Fatal("operation override constraint lost")
		}
		badConfig := map[string]string{}
		for key, value := range runtimeConfig {
			badConfig[key] = value
		}
		badConfig["namespace"] = "wrong"
		badRows := 0
		badErr := engine.Read(t.Context(), bundle, connectors.ReadRequest{Stream: expectedName, Config: connectors.RuntimeConfig{Config: badConfig}}, nil, func(connectors.Record) error { badRows++; return nil })
		if badErr == nil || badRows != 0 || len(requestSnapshot()) != 0 {
			t.Fatalf("overridden enum did not refuse before I/O: %v rows=%d requests=%v", badErr, badRows, requestSnapshot())
		}
	}
	if bundle.Streams[0].RequestInputs != nil {
		rawProvenance, err := handle.ReadFile("provenance.json")
		if err != nil {
			t.Fatal(err)
		}
		var provenance []vNextSourceExecutionProvenance
		if err := json.Unmarshal(rawProvenance, &provenance); err != nil {
			t.Fatal(err)
		}
		foundInput := false
		for _, row := range provenance {
			if row.TargetKind == "schema" && row.TargetID == bundle.Streams[0].RequestInputs.Schema && strings.HasSuffix(row.FieldPath, "/schema_refs/input") {
				foundInput = true
			}
		}
		if !foundInput {
			t.Fatal("published request input schema lost source-to-execution provenance")
		}
	}
	var ids []string
	readErr := engine.Read(context.Background(), bundle, connectors.ReadRequest{Stream: expectedName, Config: connectors.RuntimeConfig{Config: runtimeConfig}}, nil, func(record connectors.Record) error {
		id, ok := record["id"].(string)
		if !ok {
			t.Error("returned record lost string identity")
		}
		ids = append(ids, id)
		return nil
	})
	if variant == "missing_path" {
		if readErr == nil || len(ids) != 0 || len(requestSnapshot()) != 0 {
			t.Fatalf("missing path input sent or emitted: error=%v ids=%v requests=%v", readErr, ids, requestSnapshot())
		}
		return
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !reflect.DeepEqual(ids, []string{"alpha", "bravo", "charlie"}) || !reflect.DeepEqual(requestSnapshot(), []string{expectedWire}) {
		t.Fatalf("actual returned identities=%v requests=%v", ids, requestSnapshot())
	}
	if direct {
		consumer := engine.New(bundle, nil)
		if err := commandrunner.Preflight(consumer, []string{"api", expectedCommand}); err != nil {
			t.Fatalf("generated direct preflight: %v", err)
		}
		flags := map[string][]string{}
		for key, value := range runtimeConfig {
			flags[key] = []string{value}
		}
		result, err := commandrunner.Run(t.Context(), consumer, commandrunner.Request{Path: []string{"api", expectedCommand}, Flags: flags}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(variant, "typed_query_") {
			invalid := []string{`null`, `[]`, `{}`, `{"owner":"x","tag":7}`, `{"owner":"x","tag":"y","extra":"z"}`, `{"owner":"first","owner":"second","tag":"y"}`, `{"owner":"x","tag":"y"} {}`}
			if variant == "typed_query_repeated255" || variant == "typed_query_bracket255" {
				invalid = []string{`null`, `{}`, `[]`, `["a","b","c","d"]`, `["a",7]`, `["a"] []`}
			}
			switch variant {
			case "typed_query_indexed255":
				invalid = []string{`null`, `{}`, `[]`, `[{"key":"alpha","value":"2"}]`, `[{"key":"alpha","value":2,"other":true}]`}
			case "typed_query_dynamic255":
				invalid = []string{`null`, `[]`, `{"alpha":"2"}`, `{"bad[key]":2}`, `{"alpha":1,"alpha":2}`}
			case "typed_query_empty255":
				invalid = []string{`null`, `{}`, `["a",2]`, `["a","b","c","d"]`}
			}
			for _, bad := range invalid {
				_, badErr := commandrunner.Run(t.Context(), consumer, commandrunner.Request{Path: []string{"api", expectedCommand}, Flags: map[string][]string{"filter": {bad}}}, nil)
				if badErr == nil || len(requestSnapshot()) != 2 {
					t.Fatalf("invalid typed query %q: error=%v requests=%v", bad, badErr, requestSnapshot())
				}
				emitted := 0
				savedErr := consumer.Read(t.Context(), connectors.ReadRequest{Stream: expectedName, Config: connectors.RuntimeConfig{Config: map[string]string{"filter": bad}}}, func(connectors.Record) error { emitted++; return nil })
				if savedErr == nil || emitted != 0 || len(requestSnapshot()) != 2 {
					t.Fatalf("invalid saved typed query %q: error=%v rows=%d requests=%v", bad, savedErr, emitted, requestSnapshot())
				}
			}
		}
		if variant == "typed_query_map255" {
			for i := range bundle.CLISurface.Commands {
				for j := range bundle.CLISurface.Commands[i].Flags {
					flag := &bundle.CLISurface.Commands[i].Flags[j]
					if flag.Name == "filter" {
						flag.MaxBytes = 8
					}
				}
			}
			bounded := engine.New(bundle, nil)
			_, capErr := commandrunner.Run(t.Context(), bounded, commandrunner.Request{Path: []string{"api", expectedCommand}, Flags: flags}, nil)
			if capErr == nil || len(requestSnapshot()) != 2 {
				t.Fatalf("selected flag byte cap bypassed: %v requests=%v", capErr, requestSnapshot())
			}
		}
		if strings.HasPrefix(variant, "query_") {
			invalid := []string{"0", "6", "not-an-integer"}
			if variant == "query_bigint" {
				invalid = []string{"9007199254740992", "9007199254740996", "9007199254740994.5"}
			}
			if variant == "query_decimal" {
				invalid = []string{"0.10000000000000000", "0.10000000000000004", "1/10"}
			}
			if variant == "query_boolean" {
				invalid = []string{"not-a-boolean"}
			}
			if selectBad {
				invalid = []string{selectedBad}
			}
			for _, bad := range invalid {
				_, badErr := commandrunner.Run(t.Context(), consumer, commandrunner.Request{Path: []string{"api", expectedCommand}, Flags: map[string][]string{"page_size": {bad}}}, nil)
				if badErr == nil || len(requestSnapshot()) != 2 {
					t.Fatalf("invalid scalar %q error=%v requests=%v", bad, badErr, requestSnapshot())
				}
				emitted := 0
				savedErr := consumer.Read(t.Context(), connectors.ReadRequest{Stream: expectedName, Config: connectors.RuntimeConfig{Config: map[string]string{"page_size": bad}}}, func(connectors.Record) error { emitted++; return nil })
				if savedErr == nil || emitted != 0 || len(requestSnapshot()) != 2 {
					t.Fatalf("invalid saved scalar %q error=%v emitted=%d requests=%v", bad, savedErr, emitted, requestSnapshot())
				}

			}
		}
		if result.DirectRead == nil {
			t.Fatal("generated direct command returned no direct result")
		}
		body, err := json.Marshal(result.DirectRead.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"data":[{"id":"alpha"},{"id":"bravo"},{"id":"charlie"}]}` || !reflect.DeepEqual(requestSnapshot(), []string{expectedWire, expectedWire}) {
			t.Fatalf("direct body=%s requests=%v", body, requestSnapshot())
		}
	}
}

func TestSourceProjectionDocumentClosure230(t *testing.T) {
	for _, variant := range []string{"document_healthy", "document_bad_pin", "document_bad_span", "document_outside_span", "document_conflicting_selector"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, variant, variant != "document_healthy") })
	}
}

func TestSourceProjectionAdditionalScalarInvalid234(t *testing.T) {
	for _, tc := range []struct {
		kind   string
		values []string
	}{
		{"query_integer", []string{"6", "not-an-integer"}},
		{"query_bigint", []string{"9007199254740996", "9007199254740994.5"}},
		{"query_decimal", []string{"0.10000000000000004", "1/10"}},
		{"query_boolean", []string{"1", "TRUE", ""}},
	} {
		for _, bad := range tc.values {
			t.Run(tc.kind+"/"+bad, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+tc.kind+"|"+bad, false) })
		}
	}
}
