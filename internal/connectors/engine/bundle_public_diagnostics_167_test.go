package engine

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/defs"
)

// Expectations are fixed at the consumer, never derived from the producer's
// reason constructor. Both a wrong readable reason and wrong identity fail.
func publicBundleMatches167(err error, file, field, code, reason string) bool {
	var d *BundleDiagnosticError
	if !errors.As(err, &d) {
		return false
	}
	raw, marshalErr := json.Marshal(d)
	var v map[string]any
	if marshalErr != nil || json.Unmarshal(raw, &v) != nil {
		return false
	}
	return d.Connector == "acme" && d.Generation == "embedded-v1" && d.File == file && d.Field == field && v["reason_code"] == code && v["reason"] == reason && strings.Contains(err.Error(), reason) && !strings.Contains(string(raw)+err.Error(), "private-sentinel-167")
}

func TestBundlePublicDiagnostic167(t *testing.T) {
	for _, tc := range []struct{ name, file, data, field, code, reason string }{
		{"syntax", "rate_limits.json", `{"state":}`, "/", "invalid_json", "malformed JSON"},
		{"missing_name", "metadata.json", `{"display_name":"Acme","integration_type":"api","capabilities":{"read":true,"write":true}}`, "/name", "required_property_missing", "required property is missing"},
		{"unknown_metadata_member", "metadata.json", strings.TrimSuffix(validMetadata("acme"), "}") + `,"private-sentinel-167/~":true}`, "/<member:5>", "unknown_property", "additional property is not allowed"},
		{"unknown_schema_member", "schemas/widgets.json", `{"type":"object","private-sentinel-167/~":true}`, "/<member:0>", "unknown_schema_keyword", "unknown schema keyword"},
		{"nested_schema_member", "schemas/widgets.json", `{"type":"object","properties":{"private-sentinel-167/~":{"type":"no-such-type"}}}`, "/properties/<member:0>/type", "schema_type_invalid", "schema type is not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/"+tc.file] = &fstest.MapFile{Data: []byte(tc.data)}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, tc.file, tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
			if tc.name == "syntax" {
				var syntax *json.SyntaxError
				if !errors.As(err, &syntax) {
					t.Errorf("lost parser cause: %v", err)
				}
			}
		})
	}
	if _, err := Load(fullValidBundleFS("acme"), "acme"); err != nil {
		t.Fatal(err)
	}
}

func TestBundlePublicRateRules167(t *testing.T) {
	for _, tc := range []struct {
		name, field, code, reason string
		change                    func(map[string]any, map[string]any)
	}{
		{"duplicate", "/policies/1/id", "policy_id_duplicate", "policy id is duplicated", func(d, p map[string]any) { p["id"] = "private-sentinel-167"; d["policies"] = []any{p, p} }},
		{"source_missing", "/policies/0/source", "required_property_missing", "required property is missing", func(d, p map[string]any) { delete(p, "source") }},
		{"scope_missing", "/policies/0/scope/subject_config", "required_property_missing", "required property is missing", func(d, p map[string]any) { delete(p["scope"].(map[string]any), "subject_config") }},
		{"scope_absent", "/policies/0/scope/subject_config", "scope_property_absent", "subject_config must name a spec.json property", func(d, p map[string]any) { p["scope"].(map[string]any)["subject_config"] = "private-sentinel-167" }},
		{"scope_secret", "/policies/0/scope/subject_config", "scope_property_secret", "subject_config must name a non-secret spec.json property", func(d, p map[string]any) { p["scope"].(map[string]any)["subject_config"] = "token" }},
		{"source_url", "/policies/0/source/url", "source_url_invalid", "source URL must be absolute HTTPS without userinfo or query parameters", func(d, p map[string]any) {
			p["source"].(map[string]any)["url"] = "https://example.test/?private-sentinel-167"
		}},
		{"source_date", "/policies/0/source/retrieved_at", "source_date_invalid", "retrieved_at must be an ISO date", func(d, p map[string]any) { p["source"].(map[string]any)["retrieved_at"] = "2026-99-99" }},
		{"selector_conflict", "/policies/0/selector", "selector_conflict", "all cannot be combined with endpoint, tier, or auth selectors", func(d, p map[string]any) { p["selector"].(map[string]any)["all"] = true }},
		{"endpoint_path", "/policies/0/selector/endpoints/0/path", "selector_path_invalid", "endpoint path must be rooted and connector-relative", func(d, p map[string]any) {
			p["selector"].(map[string]any)["endpoints"].([]any)[0].(map[string]any)["path"] = "/private-sentinel-167?bad"
		}},
		{"budget_fields", "/policies/0/budgets/0", "budget_fields_conflict", "window budget must not declare capacity or restore_per_second", func(d, p map[string]any) { p["budgets"].([]any)[0].(map[string]any)["capacity"] = 2 }},
		{"cost_header", "/policies/0/budgets/0/cost/response_header", "cost_header_invalid", "cost response_header must be an HTTP field name", func(d, p map[string]any) {
			p["budgets"].([]any)[0].(map[string]any)["cost"].(map[string]any)["response_header"] = "private-sentinel-167 bad"
		}},
		{"cost_capacity", "/policies/0/budgets/0/cost/default_cost", "cost_exceeds_capacity", "default_cost must not exceed the declared budget capacity", func(d, p map[string]any) {
			p["budgets"].([]any)[0].(map[string]any)["cost"].(map[string]any)["default_cost"] = 999999
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var declaration map[string]any
			if err := json.Unmarshal([]byte(validProviderCitedRateLimits), &declaration); err != nil {
				t.Fatal(err)
			}
			policy := declaration["policies"].([]any)[0].(map[string]any)
			tc.change(declaration, policy)
			raw, err := json.Marshal(declaration)
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/rate_limits.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "rate_limits.json", tc.field, tc.code, tc.reason) {
				t.Errorf("public rate invariant mismatch: %v", err)
			}
		})
	}
}

func TestPublicBundleOracleFalsification167(t *testing.T) {
	good := &BundleDiagnosticError{Connector: "acme", Generation: "embedded-v1", File: "metadata.json", Field: "/name", ReasonCode: "required_property_missing", Reason: "required property is missing", Cause: errors.New("readable original cause")}
	accepts := func(err error) bool {
		return publicBundleMatches167(err, "metadata.json", "/name", "required_property_missing", "required property is missing")
	}
	if !accepts(good) {
		t.Fatal("correct public control rejected")
	}
	for _, change := range []func(*BundleDiagnosticError){func(d *BundleDiagnosticError) { d.Reason = "wrong but readable rejection" }, func(d *BundleDiagnosticError) { d.Reason = "invalid bundle declaration" }, func(d *BundleDiagnosticError) { d.Connector = "other" }, func(d *BundleDiagnosticError) { d.Generation = "wrong" }, func(d *BundleDiagnosticError) { d.Field = "/" }, func(d *BundleDiagnosticError) { d.ReasonCode = "other" }} {
		bad := *good
		change(&bad)
		if accepts(&bad) {
			t.Fatal("oracle accepted wrong public contract with readable cause")
		}
	}
}

func TestBundlePublicStreamHeaders167(t *testing.T) {
	for _, tc := range []struct{ name, headers, field, code, reason string }{
		{"unknown", `{"private-sentinel-167":"anything"}`, "/streams/0/headers/<member:0>", "stream_header_unsupported", "only fixed Accept headers are supported"},
		{"template", `{"Accept":"{{ private-sentinel-167 }}"}`, "/streams/0/headers/Accept", "stream_header_dynamic", "Accept header must be static"},
		{"media", `{"Accept":"application/private-sentinel-167"}`, "/streams/0/headers/Accept", "stream_header_media_invalid", "Accept header must be one fixed vendor JSON media type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/streams.json"] = &fstest.MapFile{Data: []byte(strings.Replace(validStreams, `"schema": "schemas/widgets.json"`, `"headers": `+tc.headers+`, "schema": "schemas/widgets.json"`, 1))}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public header invariant: %v", err)
			}
		})
	}
}

func TestBundlePublicMultipart167(t *testing.T) {
	for _, tc := range []struct{ name, from, to, field, code, reason string }{
		{"content_type", `"content_type": "multipart/form-data"`, `"content_type": "private-sentinel-167"`, "/operations/0/rest/content_type", "multipart_content_type_invalid", "multipart requires literal content_type multipart/form-data"},
		{"response_bound", `"max_bytes": 1024`, `"max_bytes": 0`, "/operations/0/rest/max_bytes", "multipart_response_bound_invalid", "multipart requires positive response capture max_bytes"},
		{"aggregate_bound", `"max_bytes": 2048`, `"max_bytes": 0`, "/operations/0/rest/multipart/max_bytes", "multipart_aggregate_bound_invalid", "multipart requires positive aggregate max_bytes"},
		{"closed_schema", `"additionalProperties": false`, `"additionalProperties": true`, "/operations/0/rest/body_schema/additionalProperties", "multipart_schema_open", "multipart body object must declare additionalProperties: false"},
		{"part_field", `"field": "message"`, `"field": "private-sentinel-167"`, "/operations/0/rest/multipart/parts/0/field", "multipart_field_invalid", "multipart part must reference a declared body field"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := multipartRestWriteBundleFS(strings.Replace(validMultipartRestWrite, tc.from, tc.to, 1), "rest_write")
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public multipart invariant: %v", err)
			}
		})
	}
}

func TestBundlePublicPolling167(t *testing.T) {
	for _, tc := range []struct {
		name, field, code, reason string
		api                       bool
	}{
		{"lossy_cursor", "/source/cursor/codec", "polling_cursor_lossy", "polling watermark cursor codec must preserve values losslessly", false},
		{"wrong_integration", "/", "polling_integration_invalid", "polling_watermark requires metadata integration_type database", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := pollingWatermarkBundleFS("acme")
			data := strings.Replace(validPollingWatermarkDefinitionJSON, `"codec": "rfc3339_nano"`, `"codec": "float64"`, 1)
			if tc.api {
				files = fullValidBundleFS("acme")
				data = validPollingWatermarkDefinitionJSON
			}
			files["acme/polling_watermark.json"] = &fstest.MapFile{Data: []byte(data)}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "polling_watermark.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong polling public invariant: %v", err)
			}
		})
	}
}

func TestBundlePublicSchemaCompile168(t *testing.T) {
	for _, tc := range []struct{ name, data, field, code, reason string }{
		{"required", `{"required":true}`, "/required", "schema_keyword_shape_invalid", "required must be an array of property names"},
		{"properties", `{"properties":true}`, "/properties", "schema_keyword_shape_invalid", "properties must be an object of schemas"},
		{"items", `{"items":true}`, "/items", "schema_keyword_shape_invalid", "items must be a schema object"},
		{"oneOf", `{"oneOf":[]}`, "/oneOf", "schema_alternatives_empty", "oneOf must contain at least one schema object"},
		{"pattern", `{"pattern":"[private-sentinel-167"}`, "/pattern", "schema_pattern_invalid", "schema pattern cannot be compiled"},
		{"format", `{"format":true}`, "/format", "schema_keyword_shape_invalid", "format must be a string"},
		{"negative", `{"minItems":-1}`, "/minItems", "schema_bound_invalid", "schema bound must be a non-negative integer"},
		{"reversed", `{"minItems":2,"maxItems":1}`, "/maxItems", "schema_bounds_conflict", "maxItems must not be below minItems"},
		{"additional", `{"additionalProperties":{}}`, "/additionalProperties", "schema_keyword_shape_invalid", "additionalProperties must be a boolean"},
		{"secret", `{"x-secret":"private-sentinel-167"}`, "/x-secret", "schema_keyword_shape_invalid", "x-secret must be a boolean"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(tc.data)}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "schemas/widgets.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong schema compile public invariant: %v", err)
			}
		})
	}
}

func TestBundlePublicRouteOwnership168(t *testing.T) {
	for _, tc := range []struct{ name, file, data, field, code, reason string }{
		{"metadata_identity", "metadata.json", validMetadata("other"), "/name", "metadata_identity_mismatch", "metadata name must match the selected connector directory"},
		{"stream_selection", "streams.json", strings.Replace(validStreams, `"name": "widgets",`, `"name": "widgets", "route":"private-sentinel-167",`, 1), "/streams/0/route", "route_selection_invalid", "route must select a declared route with a matching version"},
		{"duplicate_base", "streams.json", strings.Replace(validStreams, `"base": {`, `"base": {"routes":[{"name":"private-sentinel-167","base_url":"https://example.test"},{"name":"private-sentinel-167","base_url":"https://example.test"}],`, 1), "/base/routes/1/name", "route_name_duplicate", "route name is duplicated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/"+tc.file] = &fstest.MapFile{Data: []byte(tc.data)}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, tc.file, tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
			if tc.name == "stream_selection" {
				var route *MissingOperationRouteError
				if !errors.As(err, &route) || route.Route != "private-sentinel-167" {
					t.Error("lost original route cause")
				}
			}
		})
	}
}

func TestBundlePublicRouteSiblings168(t *testing.T) {
	for _, tc := range []struct{ name, routes, field, code, reason string }{
		{"origin", `[{"name":"primary","base_url":"https://private-sentinel-167.invalid/path"}]`, "/base/routes/0/base_url", "route_origin_invalid", "route base_url must be a fixed HTTP origin or the declared base_url template"},
		{"version", `[{"name":"primary","base_url":"https://example.test","version":"private-sentinel-167/path"}]`, "/base/routes/0/version", "route_version_invalid", "route version must be one path segment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/streams.json"] = &fstest.MapFile{Data: []byte(strings.Replace(validStreams, `"base": {`, `"base": {"routes":`+tc.routes+`,`, 1))}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicOperationStructure168(t *testing.T) {
	const read = `{"id":"acme.widgets.get","kind":"rest_read","summary":"Read widgets","risk":"low","approval":"none","output_policy":"json","rest":{"method":"GET","path":"/widgets"}}`
	for _, tc := range []struct{ name, ops, field, code, reason string }{
		{"duplicate", read + "," + read, "/operations/1/id", "operation_id_duplicate", "operation id is duplicated"},
		{"missing_block", strings.Replace(read, `,"rest":{"method":"GET","path":"/widgets"}`, "", 1), "/operations/0", "operation_execution_count_invalid", "operation must declare exactly one execution block"},
		{"wrong_block", strings.Replace(read, `"kind":"rest_read"`, `"kind":"graphql_query"`, 1), "/operations/0/kind", "operation_execution_kind_mismatch", "execution block must match the operation kind"},
		{"secret_policy", strings.Replace(secretWriteOp, "%s", "", 1), "/operations/0/sensitive_policy", "sensitive_policy_missing", "secret-sensitive operation must declare sensitive_policy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{"operations":[` + tc.ops + `]}`)}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
	files := fullValidBundleFS("acme")
	files["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{"operations":[` + read + `]}`)}
	if _, err := Load(files, "acme"); err != nil {
		t.Fatal(err)
	}
}

func TestBundlePublicGraphQL168(t *testing.T) {
	for _, tc := range []struct{ name, document, variables, suffix, code, reason string }{
		{"template_document", "query {{ private-sentinel-167 }}", `{}`, "/document", "graphql_document_dynamic", "GraphQL document must be fixed bundle metadata"},
		{"variable_type", "", `{"input":{"template":"x","type":"private-sentinel-167"}}`, "/variables/<member:0>/type", "graphql_variable_type_invalid", "GraphQL variable type is not supported"},
		{"default_type", "", `{"input":{"template":"x","default":3}}`, "/variables/<member:0>/default", "graphql_default_type_invalid", "GraphQL variable default must be a string"},
		{"default_value", "", `{"input":{"template":"x","type":"integer","default":"private-sentinel-167"}}`, "/variables/<member:0>/default", "graphql_default_invalid", "GraphQL variable default must match its declared type"},
		{"omit_type", "", `{"input":{"template":"x","omit_when_empty":"private-sentinel-167"}}`, "/variables/<member:0>/omit_when_empty", "graphql_omit_type_invalid", "GraphQL omit_when_empty must be a boolean"},
		{"unknown_member", "", `{"input":{"template":"x","private-sentinel-167":true}}`, "/variables/<member:0>/<member:0>", "graphql_template_property_invalid", "GraphQL template object has an unsupported property"},
		{"variable_name", "", `{"private-sentinel-167":true}`, "/variables/<member:0>", "graphql_variable_name_invalid", "GraphQL variable name is invalid"},
		{"nested_variable", "", `{"input":{"nested":{"template":"x","type":"private-sentinel-167"}}}`, "/variables/<member:0>/<member:0>/type", "graphql_variable_type_invalid", "GraphQL variable type is not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var streams map[string]any
			if err := json.Unmarshal([]byte(validStreams), &streams); err != nil {
				t.Fatal(err)
			}
			var variables map[string]any
			if err := json.Unmarshal([]byte(tc.variables), &variables); err != nil {
				t.Fatal(err)
			}
			doc := tc.document
			if doc == "" {
				doc = "query ListWidgets { widgets { id } }"
			}
			stream := streams["streams"].([]any)[0].(map[string]any)
			stream["method"] = "POST"
			stream["graphql"] = map[string]any{"document": doc, "operation_name": "ListWidgets", "variables": variables}
			raw, err := json.Marshal(streams)
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/streams.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", "/streams/0/graphql"+tc.suffix, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicGraphQLDirectOwnership168(t *testing.T) {
	for _, kind := range []string{"query", "mutation"} {
		t.Run(kind, func(t *testing.T) {
			op := map[string]any{"id": "acme.graphql", "kind": "graphql_" + kind, "summary": "GraphQL fixture", "risk": "low", "approval": "none", "output_policy": "json", "graphql": map[string]any{"document": "query {{ private-sentinel-167 }}", "operation_name": "Widgets"}}
			raw, err := json.Marshal(map[string]any{"operations": []any{op}})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/operations.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", "/operations/0/graphql/document", "graphql_document_dynamic", "GraphQL document must be fixed bundle metadata") {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundleUnknownRateWithReasonRejectsPolicy168(t *testing.T) {
	files := fullValidBundleFS("acme")
	files["acme/rate_limits.json"] = &fstest.MapFile{Data: []byte(strings.Replace(validProviderCitedRateLimits, `"state": "declared"`, `"state": "unknown", "reason": "provider policy unavailable"`, 1))}
	_, err := Load(files, "acme")
	if !publicBundleMatches167(err, "rate_limits.json", "/policies", "policies_forbidden", "undeclared state must not declare policies") {
		t.Fatalf("did not reach policy-prohibition predicate: %v", err)
	}
}

func TestBundlePublicStreamSemantics168(t *testing.T) {
	for _, tc := range []struct{ name, key, data, field, code, reason string }{
		{"header_name", "response_header_projection", `[{"headers_path":"headers","values_path":"values","header_name":" ","allowed_headers":["id"]}]`, "/streams/0/response_header_projection/0/header_name", "header_projection_name_invalid", "projected header name must not be blank"},
		{"header_value", "response_header_projection", `[{"headers_path":"headers","values_path":"values","value_field":" ","allowed_headers":["id"]}]`, "/streams/0/response_header_projection/0/value_field", "header_projection_value_invalid", "projected value field must not be blank"},
		{"header_blank", "response_header_projection", `[{"headers_path":"headers","values_path":"values","allowed_headers":[" "]}]`, "/streams/0/response_header_projection/0/allowed_headers/0", "header_projection_header_invalid", "allowed header must not be blank"},
		{"zip_field", "array_zip_projection", `{"array_fields":[{"field":" ","path":"a"}]}`, "/streams/0/array_zip_projection/array_fields/0/field", "array_zip_field_invalid", "array projection field must be nonblank and trimmed"},
		{"zip_duplicate", "array_zip_projection", `{"static_fields":[{"field":"private-sentinel-167","path":"a"}],"array_fields":[{"field":"private-sentinel-167","path":"b"}]}`, "/streams/0/array_zip_projection/array_fields/0/field", "array_zip_field_duplicate", "array projection field must be unique"},
		{"response_missing", "response_error", `{"path":" "}`, "/streams/0/response_error", "response_error_path_required", "response error requires path or success_path"},
		{"response_success", "response_error", `{"success_path":"a..b"}`, "/streams/0/response_error/success_path", "response_error_path_invalid", "response error path must be trimmed and contain no empty segment"},
		{"response_message", "response_error", `{"path":"error","message_field":" "}`, "/streams/0/response_error/message_field", "response_error_message_invalid", "response error message field must not be blank"},
		{"offset_position", "pagination", `{"type":"offset_count","limit_param":"limit","page_size":10,"start_page":1}`, "/streams/0/pagination", "offset_count_conflict", "offset_count cannot combine with page or start-index selection"},
		{"offset_deterministic", "pagination", `{"type":"offset_count","limit_param":"limit","page_size":10,"cursor_param":"private-sentinel-167","body_cursor_field":"private-sentinel-167"}`, "/streams/0/pagination/body_cursor_field", "offset_count_conflict", "offset_count cannot combine with another navigation control"},
		{"header_path", "response_header_projection", `[{"headers_path":" ","values_path":"values","allowed_headers":["id"]}]`, "/streams/0/response_header_projection/0", "header_projection_paths_required", "header projection requires nonblank headers_path and values_path"},
		{"header_duplicate", "response_header_projection", `[{"headers_path":"headers","values_path":"values","allowed_headers":["private-sentinel-167","private-sentinel-167"]}]`, "/streams/0/response_header_projection/0/allowed_headers/1", "header_projection_duplicate", "projected header must be unique"},
		{"zip_path", "array_zip_projection", `{"array_fields":[{"field":"private-sentinel-167","path":"a..b"}]}`, "/streams/0/array_zip_projection/array_fields/0/path", "array_zip_path_invalid", "array projection path must be nonblank, trimmed and contain no empty segment"},
		{"response_path", "response_error", `{"path":"private-sentinel-167..x"}`, "/streams/0/response_error/path", "response_error_path_invalid", "response error path must be trimmed and contain no empty segment"},
		{"offset_limit", "pagination", `{"type":"offset_count","limit_param":" ","page_size":10}`, "/streams/0/pagination", "offset_count_limit_required", "offset_count requires limit_param and positive page_size"},
		{"offset_conflict", "pagination", `{"type":"offset_count","limit_param":"limit","page_size":10,"cursor_param":"private-sentinel-167"}`, "/streams/0/pagination/cursor_param", "offset_count_conflict", "offset_count cannot combine with another navigation control"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			var doc map[string]any
			if err := json.Unmarshal([]byte(validStreams), &doc); err != nil {
				t.Fatal(err)
			}
			var value any
			if err := json.Unmarshal([]byte(tc.data), &value); err != nil {
				t.Fatal(err)
			}
			doc["streams"].([]any)[0].(map[string]any)[tc.key] = value
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			files["acme/streams.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicDynamicFields168(t *testing.T) {
	for _, tc := range []struct{ name, patch, field, code, reason string }{
		{"body_type", `{"body_type":"form"}`, "/actions/0/body_type", "dynamic_fields_body_type", "dynamic_fields requires body_type json"},
		{"field_blank", `{"dynamic_fields":{"field":" ","key_pattern":".*"}}`, "/actions/0/dynamic_fields/field", "dynamic_fields_field_required", "dynamic_fields requires a nonblank field"},
		{"pattern_blank", `{"dynamic_fields":{"field":"custom","key_pattern":" "}}`, "/actions/0/dynamic_fields/key_pattern", "dynamic_fields_pattern_required", "dynamic_fields requires a nonblank key_pattern"},
		{"pattern_invalid", `{"dynamic_fields":{"field":"custom","key_pattern":"[private-sentinel-167"}}`, "/actions/0/dynamic_fields/key_pattern", "dynamic_fields_pattern_invalid", "dynamic_fields key_pattern must be a valid regular expression"},
		{"path_collision", `{"path_fields":["custom"]}`, "/actions/0/dynamic_fields/field", "dynamic_fields_binding_conflict", "dynamic field cannot also be a path or body binding"},
		{"body_collision", `{"body_fields":["custom"]}`, "/actions/0/dynamic_fields/field", "dynamic_fields_binding_conflict", "dynamic field cannot also be a path or body binding"},
		{"single_body_collision", `{"body_field":"custom"}`, "/actions/0/dynamic_fields/field", "dynamic_fields_binding_conflict", "dynamic field cannot also be a path or body binding"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var action, patch map[string]any
			if err := json.Unmarshal([]byte(`{"name":"sync_member","kind":"upsert","method":"POST","path":"/members","risk":"low","record_schema":{"type":"object"},"dynamic_fields":{"field":"custom","key_pattern":".*"}}`), &action); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				action[k] = v
			}
			raw, err := json.Marshal(map[string]any{"actions": []any{action}})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/writes.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "writes.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicWriteSemantics168(t *testing.T) {
	for _, tc := range []struct{ name, patch, field, code, reason string }{
		{"success_duplicate", `{"success_statuses":[200,200]}`, "/actions/0/success_statuses/1", "write_status_duplicate", "write success status must be unique"},
		{"origin_duplicate", `{"base_url":"https://example.com","allowed_base_url_origins":["https://example.com","https://EXAMPLE.com"]}`, "/actions/0/allowed_base_url_origins/1", "write_allowed_origin_duplicate", "allowed origin must be unique"},
		{"multipart_missing", `{"body_type":"multipart"}`, "/actions/0/multipart/parts", "write_multipart_parts_required", "multipart body requires nonempty multipart.parts"},
		{"multipart_blank", `{"body_type":"multipart","multipart":{"parts":[{"name":" ","type":"field","field":"field"}]}}`, "/actions/0/multipart/parts/0", "write_multipart_binding_required", "multipart part requires nonblank name and field"},
		{"base64_source_blank", `{"body_type":"base64_upload","base64_upload":{"source_field":" ","content_field":"content","max_decoded_bytes":10}}`, "/actions/0/base64_upload/source_field", "base64_source_field_required", "base64_upload requires a nonblank source_field"},
		{"base64_content_blank", `{"body_type":"base64_upload","base64_upload":{"source_field":"file","content_field":" ","max_decoded_bytes":10}}`, "/actions/0/base64_upload/content_field", "base64_content_field_required", "base64_upload requires a nonblank content_field"},
		{"base64_ceiling", `{"body_type":"base64_upload","base64_upload":{"source_field":"file","content_field":"content","max_decoded_bytes":16777217}}`, "/actions/0/base64_upload/max_decoded_bytes", "base64_decoded_ceiling", "base64_upload max_decoded_bytes must not exceed 16777216"},
		{"hook_missing", `{"hook_fields":["custom"]}`, "/actions/0/hook", "write_hook_required", "hook_fields requires hook"},
		{"hook_blank", `{"hook":"test-hook","hook_fields":[" "]}`, "/actions/0/hook_fields/0", "write_hook_field_invalid", "hook field must be nonblank and trimmed"},
		{"hook_duplicate", `{"hook":"test-hook","hook_fields":["custom","custom"],"record_schema":{"type":"object","properties":{"custom":{"type":"string"}}}}`, "/actions/0/hook_fields/1", "write_hook_field_duplicate", "hook field must be unique"},
		{"hook_undeclared", `{"hook":"test-hook","hook_fields":["private-sentinel-167"]}`, "/actions/0/hook_fields/0", "write_hook_field_undeclared", "hook field must be declared in record_schema"},
		{"hook_overlap", `{"hook":"test-hook","hook_fields":["custom"],"body_fields":["custom"],"record_schema":{"type":"object","properties":{"custom":{"type":"string"}}}}`, "/actions/0/hook_fields/0", "write_hook_field_overlap", "hook field must not overlap the primary request contract"},
		{"binary_media", `{"body_type":"binary_upload","binary_upload":{"source_field":"file","max_bytes":10,"allowed_media_types":["private-sentinel-167;="]}}`, "/actions/0/binary_upload/allowed_media_types/0", "upload_media_type_invalid", "allowed upload media type must be valid"},
		{"multipart_media", `{"body_type":"multipart","multipart":{"parts":[{"name":"file","type":"file","field":"file","allowed_media_types":["private-sentinel-167;="]}]}}`, "/actions/0/multipart/parts/0/allowed_media_types/0", "multipart_media_type_invalid", "allowed media type is invalid"},
		{"body_required", `{"body_type":"form","body_required":true}`, "/actions/0/body_required", "write_body_required_type", "body_required requires body_type json"},
		{"idempotency_header", `{"idempotency_key_header":"private-sentinel-167 bad"}`, "/actions/0/idempotency_key_header", "pattern_mismatch", "value does not match the required pattern"},
		{"graphql_conflict", `{"graphql":{"document":"mutation Update { ok }","operation_name":"Update","variables":{}}}`, "/actions/0/body_type", "write_graphql_type_conflict", "graphql requires matching body_type"},
		{"graphql_fields", `{"body_type":"graphql","graphql":{"document":"mutation Update { ok }","operation_name":"Update","variables":{}},"body_fields":["custom"]}`, "/actions/0/body_fields", "write_graphql_body_fields", "GraphQL body cannot declare body_fields"},
		{"graphql_method", `{"body_type":"graphql","graphql":{"document":"mutation Update { ok }","operation_name":"Update","variables":{}},"method":"PUT"}`, "/actions/0/method", "write_graphql_method", "GraphQL action method must be POST"},
		{"array_field", `{"body_type":"json_array"}`, "/actions/0/body_field", "write_array_field_required", "json_array requires a nonblank body_field"},
		{"array_schema", `{"body_type":"json_array","body_field":"items"}`, "/actions/0/body_schema", "write_array_schema_required", "json_array requires body_schema"},
		{"origin_missing", `{"allowed_base_url_origins":["https://example.com"]}`, "/actions/0/base_url", "write_origin_required", "allowed_base_url_origins requires base_url"},
		{"origin_invalid", `{"base_url":"https://example.com/private-sentinel-167"}`, "/actions/0/base_url", "pattern_mismatch", "value does not match the required pattern"},
		{"allowed_origin_invalid", `{"base_url":"https://example.com","allowed_base_url_origins":["https://example.com/private-sentinel-167"]}`, "/actions/0/allowed_base_url_origins/0", "pattern_mismatch", "value does not match the required pattern"},
		{"binary_missing", `{"body_type":"binary_upload"}`, "/actions/0/binary_upload", "binary_upload_spec_required", "body_type binary_upload requires binary_upload"},
		{"binary_bound", `{"body_type":"binary_upload","binary_upload":{"source_field":"file","max_bytes":67108865}}`, "/actions/0/binary_upload/max_bytes", "binary_upload_bound_invalid", "binary_upload max_bytes must be between 1 and 67108864"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var action, patch map[string]any
			if err := json.Unmarshal([]byte(`{"name":"sync_member","kind":"upsert","method":"POST","path":"/members","risk":"low","record_schema":{"type":"object"}}`), &action); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				action[k] = v
			}
			raw, err := json.Marshal(map[string]any{"actions": []any{action}})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/writes.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "writes.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicDeclaredBatch168(t *testing.T) {
	load := func(actions []WriteAction) error {
		for i := range actions {
			actions[i].Risk = "low"
		}
		raw, err := json.Marshal(map[string]any{"actions": actions})
		if err != nil {
			t.Fatal(err)
		}
		files := fullValidBundleFS("acme")
		files["acme/writes.json"] = &fstest.MapFile{Data: raw}
		_, err = Load(files, "acme")
		return err
	}
	if err := load(declaredBatchTestBundle("https://example.com").Writes); err != nil {
		t.Fatalf("healthy batch fixture: %v", err)
	}
	for _, tc := range []struct {
		name, field, code, reason string
		mutate                    func([]WriteAction)
	}{
		{"missing", "/actions/2/declared_batch", "batch_spec_required", "body_type declared_batch requires declared_batch", func(a []WriteAction) { a[2].DeclaredBatch = nil }},
		{"method", "/actions/2/method", "batch_method_invalid", "declared_batch method must be POST", func(a []WriteAction) { a[2].Method = "PUT" }},
		{"field_collision", "/actions/2/declared_batch", "batch_fields_conflict", "batch method, path and data fields must be distinct", func(a []WriteAction) { a[2].DeclaredBatch.ProviderMethodField = "data" }},
		{"duplicate_method", "/actions/2/declared_batch/allowed_methods/2", "batch_method_duplicate", "allowed batch method must be unique", func(a []WriteAction) {
			a[2].DeclaredBatch.AllowedMethods = append(a[2].DeclaredBatch.AllowedMethods, "POST")
		}},
		{"self", "/actions/2/declared_batch/allowed_actions/2", "batch_self_reference", "declared batch cannot select itself", func(a []WriteAction) {
			a[2].DeclaredBatch.AllowedActions = append(a[2].DeclaredBatch.AllowedActions, "submit_batch")
		}},
		{"duplicate_action", "/actions/2/declared_batch/allowed_actions/2", "batch_action_duplicate", "allowed batch action must be unique", func(a []WriteAction) {
			a[2].DeclaredBatch.AllowedActions = append(a[2].DeclaredBatch.AllowedActions, "create_item")
		}},
		{"unknown", "/actions/2/declared_batch/allowed_actions/0", "batch_action_unknown", "batch action must name a declared write action", func(a []WriteAction) { a[2].DeclaredBatch.AllowedActions[0] = "private_sentinel_167" }},
		{"method_outside", "/actions/2/declared_batch/allowed_actions/1", "batch_action_method", "batch action method must be in allowed_methods", func(a []WriteAction) { a[2].DeclaredBatch.AllowedMethods = []string{"POST"} }},
		{"body_type", "/actions/2/declared_batch/allowed_actions/0", "batch_action_body", "batch action body_type must be json or none", func(a []WriteAction) { a[0].BodyType = "form" }},
		{"alternate", "/actions/2/declared_batch/allowed_actions/0", "batch_action_execution", "batch action must use the primary execution contract", func(a []WriteAction) { a[0].BaseURL = "https://example.com" }},
		{"confirmation", "/actions/2/confirm", "batch_confirmation_required", "batch containing destructive actions requires destructive confirmation", func(a []WriteAction) { a[2].Confirm = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := declaredBatchTestBundle("https://example.com").Writes
			tc.mutate(a)
			err := load(a)
			if !publicBundleMatches167(err, "writes.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicProviderSearch168(t *testing.T) {
	load := func(op OperationSpec) error {
		raw, err := json.Marshal(map[string]any{"operations": []OperationSpec{op}})
		if err != nil {
			t.Fatal(err)
		}
		files := fullValidBundleFS("acme")
		files["acme/operations.json"] = &fstest.MapFile{Data: raw}
		_, err = Load(files, "acme")
		return err
	}
	if err := load(providerSearchOp(nil)); err != nil {
		t.Fatalf("healthy provider search fixture: %v", err)
	}
	for _, tc := range []struct {
		name, field, code, reason string
		mutate                    func(*OperationSpec)
	}{
		{"method", "/operations/0/rest/method", "search_method_invalid", "provider_search method must be POST", func(o *OperationSpec) { o.REST.Method = "GET" }},
		{"content_type", "/operations/0/rest/content_type", "search_content_type_invalid", "provider_search requires application/json content_type", func(o *OperationSpec) { o.REST.ContentType = "text/plain" }},
		{"bound", "/operations/0/rest/max_bytes", "search_bound_required", "provider_search must declare positive max_bytes", func(o *OperationSpec) { o.REST.MaxBytes = 0 }},
		{"mutation", "/operations/0/mutation_class", "search_mutation_forbidden", "provider_search must not declare a mutating mutation_class", func(o *OperationSpec) { o.MutationClass = "create" }},
		{"missing_schema", "/operations/0/rest/body_schema", "search_schema_required", "provider_search must declare body_schema", func(o *OperationSpec) { o.REST.BodySchema = nil }},
		{"open_root", "/operations/0/rest/body_schema/additionalProperties", "search_object_open", "provider_search objects must declare additionalProperties false", func(o *OperationSpec) { o.REST.BodySchema = json.RawMessage(`{"type":"object"}`) }},
		{"unbounded", "/operations/0/rest/body_schema/properties/<member:0>/maxItems", "search_array_unbounded", "provider_search arrays must declare maxItems", func(o *OperationSpec) {
			o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"private-sentinel-167":{"type":"array","items":{"type":"string"}}}}`)
		}},
		{"nested_open", "/operations/0/rest/body_schema/properties/<member:0>/items/additionalProperties", "search_object_open", "provider_search objects must declare additionalProperties false", func(o *OperationSpec) {
			o.REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"private-sentinel-167":{"type":"array","maxItems":2,"items":{"type":"object"}}}}`)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := load(providerSearchOp(tc.mutate))
			if !publicBundleMatches167(err, "operations.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicOperationRules168(t *testing.T) {
	for _, tc := range []struct{ name, patch, field, code, reason string }{
		{"pager_source_missing", `{"rest":{"method":"GET","path":"/widgets","pagination":{"type":"offset_limit","limit_param":"limit","offset_param":"offset","page_size":10}}}`, "/operations/0/rest/pagination_parameters", "pagination_source_required", "query pagination requires source pagination_parameters"},
		{"pager_source_mismatch", `{"rest":{"method":"GET","path":"/widgets","pagination":{"type":"offset_limit","limit_param":"limit","offset_param":"offset","page_size":10},"pagination_parameters":[{"name":"limit","in":"query"}]}}`, "/operations/0/rest/pagination_parameters", "pagination_source_mismatch", "pagination controls must exactly match source pagination_parameters"},
		{"pager_source_unused", `{"rest":{"method":"GET","path":"/widgets","pagination":{"type":"offset_limit","limit_param":"limit","offset_param":"offset","page_size":10},"pagination_parameters":[{"name":"limit","in":"query"},{"name":"offset","in":"query"},{"name":"private_sentinel_167","in":"query"}]}}`, "/operations/0/rest/pagination_parameters", "pagination_source_mismatch", "pagination controls must exactly match source pagination_parameters"},
		{"binary_content_type", `{"kind":"binary_download","rest":null,"binary":{"method":"GET","path":"/widgets","max_bytes":10,"content_types":["private-sentinel-167"]}}`, "/operations/0/binary/content_types/0", "binary_content_type_invalid", "declared binary content type must be a valid media range"},
		{"binary_charset", `{"kind":"binary_download","rest":null,"binary":{"method":"GET","path":"/widgets","max_bytes":10,"charset":"private-sentinel-167;="}}`, "/operations/0/binary/charset", "binary_charset_invalid", "declared binary charset must be valid"},
		{"text_plain_body", `{"rest":{"method":"POST","path":"/widgets","content_type":"text/plain","body_schema":{"type":"string"},"body":{"private-sentinel-167":true}}}`, "/operations/0/rest/body", "text_plain_body_forbidden", "text/plain direct read must not declare rest.body"},
		{"text_plain_schema", `{"rest":{"method":"POST","path":"/widgets","content_type":"text/plain","body_schema":{"type":"object"}}}`, "/operations/0/rest/body_schema/type", "text_plain_schema_type", "text/plain direct read requires a root string body_schema"},
		{"secret_response_pair", `{"sensitive_policy":{"response_secret_field":"private-sentinel-167"}}`, "/operations/0/sensitive_policy", "sensitive_response_pair", "response_secret_field and response_secret_store_key must be declared together"},
		{"read_method", `{"rest":{"method":"PUT","path":"/widgets"}}`, "/operations/0/rest/method", "rest_read_method_invalid", "rest_read method must be GET or POST"},
		{"read_body", `{"rest":{"method":"POST","path":"/widgets"}}`, "/operations/0/rest/body_schema", "rest_read_schema_required", "rest_read POST must declare body_schema"},
		{"read_mutation", `{"mutation_class":"create"}`, "/operations/0/mutation_class", "rest_read_mutation_forbidden", "rest_read must not declare a mutating mutation_class"},
		{"write_mutation", `{"kind":"rest_write","rest":{"method":"POST","path":"/widgets"},"approval":"required"}`, "/operations/0/mutation_class", "operation_mutation_required", "mutating operation must declare mutation_class"},
		{"write_approval", `{"kind":"rest_write","rest":{"method":"POST","path":"/widgets"},"mutation_class":"create"}`, "/operations/0/approval", "operation_approval_required", "mutating operation must declare approval requirements"},
		{"status_bound", `{"kind":"rest_status","rest":{"method":"HEAD","path":"/widgets","max_bytes":1025},"output_policy":"status"}`, "/operations/0/rest/max_bytes", "status_bound_invalid", "rest_status max_bytes must be between 1 and 1024"},
		{"binary_method", `{"kind":"binary_download","rest":null,"binary":{"method":"POST","path":"/widgets","max_bytes":10}}`, "/operations/0/binary/method", "binary_method_invalid", "binary_download method must be GET"},
		{"binary_accept", `{"kind":"binary_download","rest":null,"binary":{"method":"GET","path":"/widgets","max_bytes":10,"accept":"private-sentinel-167;="}}`, "/operations/0/binary/accept", "binary_accept_invalid", "binary_download accept must be a valid media type"},
		{"text_method", `{"kind":"text_export","rest":null,"binary":{"method":"POST","path":"/widgets","max_bytes":10,"accept":"text/csv"},"output_policy":"file_manifest"}`, "/operations/0/binary/method", "text_export_method_invalid", "text_export method must be GET"},
		{"text_accept", `{"kind":"text_export","rest":null,"binary":{"method":"GET","path":"/widgets","max_bytes":10,"accept":"text/plain"},"output_policy":"file_manifest"}`, "/operations/0/binary/accept", "text_export_accept_invalid", "text_export accept must be text/csv"},
		{"text_output", `{"kind":"text_export","rest":null,"binary":{"method":"GET","path":"/widgets","max_bytes":10,"accept":"text/csv"}}`, "/operations/0/output_policy", "text_export_output_invalid", "text_export output_policy must be file_manifest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var op, patch map[string]any
			if err := json.Unmarshal([]byte(`{"id":"acme.widgets.read","kind":"rest_read","summary":"read widgets","risk":"low","approval":"none","output_policy":"json","rest":{"method":"GET","path":"/widgets"}}`), &op); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				if v == nil {
					delete(op, k)
				} else {
					op[k] = v
				}
			}
			raw, err := json.Marshal(map[string]any{"operations": []any{op}})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/operations.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicOperationHeaders168(t *testing.T) {
	for _, tc := range []struct{ name, key, data, field, code, reason string }{
		{"runtime_owned", "parameters", `[{"name":"X-Runtime-Token","in":"header","type":"string","max_bytes":64,"schema":{"type":"string"}}]`, "/rest/parameters/0/name", "parameter_header_protected", "request header is protected and runtime-owned"},
		{"protected", "parameters", `[{"name":"Authorization","in":"header","type":"string","max_bytes":64,"schema":{"type":"string"}}]`, "/rest/parameters/0/name", "parameter_header_protected", "request header is protected and runtime-owned"},
		{"duplicate", "parameters", `[{"name":"X-Test","in":"header","type":"string","max_bytes":64,"schema":{"type":"string"}},{"name":"x-test","in":"header","type":"string","max_bytes":64,"schema":{"type":"string"}}]`, "/rest/parameters/1/name", "parameter_header_duplicate", "request header must be unique ignoring case"},
		{"missing_schema", "parameters", `[{"name":"X-Test","in":"header","type":"string","max_bytes":64}]`, "/rest/parameters/0/schema", "parameter_header_schema_required", "request header requires a bounded string schema"},
		{"schema_type", "parameters", `[{"name":"X-Test","in":"header","type":"string","max_bytes":64,"schema":{"type":"integer"}}]`, "/rest/parameters/0/schema/type", "parameter_header_schema_type", "request header schema type must be string"},
		{"bound", "parameters", `[{"name":"X-Test","in":"header","type":"string","max_bytes":0,"schema":{"type":"string"}}]`, "/rest/parameters/0/max_bytes", "parameter_header_bound", "request header max_bytes must be between 1 and 16384"},
		{"numeric_type", "parameters", `[{"name":"limit","in":"query","type":"string","minimum":1}]`, "/rest/parameters/0", "parameter_numeric_type", "numeric bounds require integer or number type"},
		{"numeric_order", "parameters", `[{"name":"limit","in":"query","type":"integer","minimum":5,"maximum":1}]`, "/rest/parameters/0", "parameter_numeric_bounds", "minimum must not exceed maximum"},
		{"response_duplicate", "response", `{"headers":[{"name":"X-Test","max_bytes":64},{"name":"x-test","max_bytes":64}]}`, "/rest/response/headers/1/name", "response_header_duplicate", "response header must be unique ignoring case"},
		{"response_bound", "response", `{"headers":[{"name":"X-Test","max_bytes":0}]}`, "/rest/response/headers/0/max_bytes", "response_header_bound", "response header max_bytes must be between 1 and 16384"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value any
			if err := json.Unmarshal([]byte(tc.data), &value); err != nil {
				t.Fatal(err)
			}
			rest := map[string]any{"method": "GET", "path": "/widgets", tc.key: value}
			op := map[string]any{"id": "acme.widgets.read", "kind": "rest_read", "summary": "read widgets", "risk": "low", "approval": "none", "output_policy": "json", "rest": rest}
			raw, err := json.Marshal(map[string]any{"operations": []any{op}})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			if tc.name == "runtime_owned" {
				var streams map[string]any
				if err := json.Unmarshal([]byte(validStreams), &streams); err != nil {
					t.Fatal(err)
				}
				streams["base"].(map[string]any)["headers"] = map[string]any{"X-Runtime-Token": "{{ secrets.token }}"}
				data, err := json.Marshal(streams)
				if err != nil {
					t.Fatal(err)
				}
				files["acme/streams.json"] = &fstest.MapFile{Data: data}
			}
			files["acme/operations.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", "/operations/0"+tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicChangefeed168(t *testing.T) {
	const baseline = `{"status":"implemented","mechanism":"logical_replication","source":{"artifact_url":"https://example.com/source","artifact_version":"v1","retrieved_at":"2026-09-01"},"executor":{"kind":"native","id":"acme-logical"},"checkpoint":{"kind":"lsn","keys":["lsn"],"commit_after":"warehouse_durable","on_invalid":"fail"},"delivery":{"ordering":"per_stream","duplicates":"at_least_once","deletes":"tombstone","dedupe_key":["id"]},"streams":["widgets"]}`
	load := func(patch string) error {
		var doc, changes map[string]any
		if err := json.Unmarshal([]byte(baseline), &doc); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(patch), &changes); err != nil {
			t.Fatal(err)
		}
		for k, v := range changes {
			if v == nil {
				delete(doc, k)
			} else {
				doc[k] = v
			}
		}
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		files := fullValidBundleFS("acme")
		files["acme/changefeed.json"] = &fstest.MapFile{Data: raw}
		_, err = Load(files, "acme")
		return err
	}
	if err := load(`{}`); err != nil {
		t.Fatalf("healthy changefeed fixture: %v", err)
	}
	for _, tc := range []struct{ name, patch, field, code, reason string }{
		{"source_blank", `{"source":{"artifact_url":"https://example.com/source","artifact_version":" ","retrieved_at":"2026-09-01"}}`, "/source", "changefeed_source_required", "changefeed source requires nonblank artifact_url, artifact_version and retrieved_at"},
		{"source_url", `{"source":{"artifact_url":"private-sentinel-167","artifact_version":"v1","retrieved_at":"2026-09-01"}}`, "/source/artifact_url", "format_mismatch", "value does not match the required format"},
		{"source_scheme", `{"source":{"artifact_url":"ftp://example.com/source","artifact_version":"v1","retrieved_at":"2026-09-01"}}`, "/source/artifact_url", "changefeed_source_url", "changefeed source artifact_url must be an absolute HTTP or HTTPS URL"},
		{"source_date", `{"source":{"artifact_url":"https://example.com/source","artifact_version":"v1","retrieved_at":"private-sentinel-167"}}`, "/source/retrieved_at", "changefeed_source_date", "changefeed source retrieved_at must be an ISO-8601 date"},
		{"executor", `{"executor":null}`, "/executor", "changefeed_executor_required", "implemented changefeed requires a named executor"},
		{"checkpoint", `{"checkpoint":null}`, "/checkpoint", "changefeed_checkpoint_required", "implemented changefeed requires checkpoint kind, keys, commit_after and on_invalid"},
		{"checkpoint_blank", `{"checkpoint":{"kind":"lsn","keys":[" "],"commit_after":"warehouse_durable","on_invalid":"fail"}}`, "/checkpoint/keys/0", "changefeed_key_blank", "changefeed key must not be blank"},
		{"checkpoint_duplicate", `{"checkpoint":{"kind":"lsn","keys":["private-sentinel-167","private-sentinel-167"],"commit_after":"warehouse_durable","on_invalid":"fail"}}`, "/checkpoint/keys/1", "changefeed_key_duplicate", "changefeed key must be unique"},
		{"delivery", `{"delivery":null}`, "/delivery", "changefeed_delivery_required", "implemented changefeed requires ordering, duplicates and deletes guarantees"},
		{"dedupe_duplicate", `{"delivery":{"ordering":"per_stream","duplicates":"at_least_once","deletes":"tombstone","dedupe_key":["private-sentinel-167","private-sentinel-167"]}}`, "/delivery/dedupe_key/1", "changefeed_key_duplicate", "changefeed key must be unique"},
		{"streams_duplicate", `{"streams":["private-sentinel-167","private-sentinel-167"]}`, "/streams/1", "changefeed_key_duplicate", "changefeed key must be unique"},
		{"reason", `{"status":"unsupported","reason":" "}`, "/reason", "changefeed_reason_required", "unsupported changefeed requires a nonblank reason"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := load(tc.patch)
			if !publicBundleMatches167(err, "changefeed.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicTransportBasics168(t *testing.T) {
	const source = `{"executor":{"family":"native_api","id":"acme_snapshot_source"},"eligible_streams":["widgets"],"modes":["full_append"],"delivery":{"idempotency":"at_least_once","ordering":"source_ordered","deletes":"not_available"}}`
	const destination = `{"executor":{"family":"native_database","id":"acme_stage_destination"},"eligible_actions":["stage_append"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append"}]}`
	for _, tc := range []struct{ name, role, patch, field, code, reason string }{
		{"empty", "", `{}`, "/", "transport_role_required", "source or destination transport must be declared"},
		{"strategy_unknown_mode", "destination_transport", `{"apply_strategies":[{"mode":"private-sentinel-167","strategy":"append","action":"stage_append"}]}`, "/destination_transport/apply_strategies/0/mode", "transport_mode_invalid", "transport sync mode is not supported"},
		{"strategy_unknown_kind", "destination_transport", `{"apply_strategies":[{"mode":"full_append","strategy":"private-sentinel-167","action":"stage_append"}]}`, "/destination_transport/apply_strategies/0/strategy", "transport_strategy_invalid", "apply strategy is not supported"},
		{"destination_change_capture", "destination_transport", `{"modes":["change_capture"]}`, "/destination_transport/modes/0", "destination_change_capture_forbidden", "change_capture is source-only into the connection warehouse"},
		{"source_executor", "source_transport", `{"executor":{"family":"native_api","id":" "}}`, "/source_transport/executor/id", "transport_executor_id", "transport executor requires a concrete ID"},
		{"source_wildcard", "source_transport", `{"eligible_streams":["*","widgets"]}`, "/source_transport/eligible_streams/0", "transport_stream_wildcard", "source stream wildcard must be the only entry"},
		{"source_duplicate", "source_transport", `{"eligible_streams":["private-sentinel-167","private-sentinel-167"]}`, "/source_transport/eligible_streams/1", "transport_name_duplicate", "transport name must be unique"},
		{"source_modes", "source_transport", `{"modes":["full_append","full_append"]}`, "/source_transport/modes/1", "transport_mode_duplicate", "transport sync mode must be unique"},
		{"source_delivery", "source_transport", `{"delivery":{"idempotency":"single_attempt","ordering":"source_ordered","deletes":"not_available"}}`, "/source_transport/delivery/idempotency", "source_single_attempt_forbidden", "source transport cannot declare single_attempt delivery"},
		{"destination_duplicate", "destination_transport", `{"eligible_actions":["stage_append","stage_append"]}`, "/destination_transport/eligible_actions/1", "transport_name_duplicate", "transport name must be unique"},
		{"destination_modes", "destination_transport", `{"modes":["full_append","full_append"]}`, "/destination_transport/modes/1", "transport_mode_duplicate", "transport sync mode must be unique"},
		{"strategy_mode", "destination_transport", `{"apply_strategies":[{"mode":"full_overwrite","strategy":"replace","action":"stage_append"}]}`, "/destination_transport/apply_strategies/0/mode", "transport_strategy_mode", "apply strategy mode must be a declared destination mode"},
		{"strategy_action", "destination_transport", `{"apply_strategies":[{"mode":"full_append","strategy":"append","action":"private-sentinel-167"}]}`, "/destination_transport/apply_strategies/0/action", "transport_strategy_action", "apply strategy action must be an eligible action"},
		{"strategy_tombstone", "destination_transport", `{"apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append","tombstone_action":"private-sentinel-167"}]}`, "/destination_transport/apply_strategies/0/tombstone_action", "transport_tombstone_action", "tombstone action must be an eligible action"},
		{"strategy_same_tombstone", "destination_transport", `{"apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append","tombstone_action":"stage_append"}]}`, "/destination_transport/apply_strategies/0/tombstone_action", "transport_tombstone_conflict", "tombstone action must differ from its ordinary apply action"},
		{"strategy_duplicate", "destination_transport", `{"apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append"},{"mode":"full_append","strategy":"append","action":"stage_append"}]}`, "/destination_transport/apply_strategies/1/action", "transport_strategy_duplicate", "apply strategy action must be unique within its sync mode"},
		{"strategy_missing_mode", "destination_transport", `{"modes":["full_append","incremental_append"]}`, "/destination_transport/modes/1", "transport_strategy_missing", "destination sync mode requires a declared apply strategy"},
		{"strategy_missing_action", "destination_transport", `{"eligible_actions":["stage_append","private-sentinel-167"]}`, "/destination_transport/eligible_actions/1", "transport_action_strategy_missing", "eligible destination action requires a declared apply strategy"},
		{"copy_database", "destination_transport", `{"copy_worker_maximum":1}`, "/destination_transport/copy_worker_maximum", "copy_database_required", "copy_worker_maximum requires a database resource declaration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := source
			if tc.role == "destination_transport" {
				base = destination
			}
			var role, patch map[string]any
			if err := json.Unmarshal([]byte(base), &role); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				role[k] = v
			}
			document := map[string]any{"schema_version": 1}
			if tc.role != "" {
				document[tc.role] = role
			}
			raw, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/sync_transport.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "sync_transport.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicSourceBindings168(t *testing.T) {
	const binding = `{"executor":{"family":"native_api","id":"acme_source"},"eligible_streams":["widgets"],"record_mapping":{"kind":"input_fields","inputs":[{"input":"widget_id","field":"id"}]}}`
	for _, tc := range []struct {
		name, patch, field, code, reason string
		duplicate                        bool
	}{
		{"executor", `{"executor":{"family":"native_api","id":" "}}`, "/executor/id", "transport_executor_id", "transport executor requires a concrete ID", false},
		{"streams", `{"eligible_streams":["widgets","widgets"]}`, "/eligible_streams/1", "transport_name_duplicate", "transport name must be unique", false},
		{"config_inputs", `{"record_mapping":{"kind":"config_match","config_key":"workspace","record_field":"workspace","inputs":[{"input":"id","field":"id"}]}}`, "/record_mapping/inputs", "mapping_config_inputs_forbidden", "config_match mapping must not declare input fields", false},
		{"input_config", `{"record_mapping":{"kind":"input_fields","config_key":"workspace","inputs":[{"input":"id","field":"id"}]}}`, "/record_mapping", "mapping_input_config_forbidden", "input_fields mapping must not declare config_match fields", false},
		{"input_duplicate", `{"record_mapping":{"kind":"input_fields","inputs":[{"input":"private-sentinel-167","field":"id"},{"input":"private-sentinel-167","field":"name"}]}}`, "/record_mapping/inputs/1/input", "mapping_input_duplicate", "source mapping input must be unique", false},
		{"field_duplicate", `{"record_mapping":{"kind":"input_fields","inputs":[{"input":"id","field":"private-sentinel-167"},{"input":"name","field":"private-sentinel-167"}]}}`, "/record_mapping/inputs/1/field", "mapping_field_duplicate", "source mapping field must be unique", false},
		{"tombstone_blank", `{"record_mapping":null,"tombstone_mapping":{"image":"key","inputs":[{"input":"","field":"id"}]}}`, "/tombstone_mapping/inputs/0", "tombstone_fields_required", "tombstone mapping requires nonempty input and field names", false},
		{"tombstone_input", `{"record_mapping":null,"tombstone_mapping":{"image":"key","inputs":[{"input":"private-sentinel-167","field":"id"},{"input":"private-sentinel-167","field":"name"}]}}`, "/tombstone_mapping/inputs/1/input", "tombstone_input_duplicate", "tombstone mapping input must be unique", false},
		{"tombstone_field", `{"record_mapping":null,"tombstone_mapping":{"image":"key","inputs":[{"input":"id","field":"private-sentinel-167"},{"input":"name","field":"private-sentinel-167"}]}}`, "/tombstone_mapping/inputs/1/field", "tombstone_field_duplicate", "tombstone mapping field must be unique", false},
		{"overlap", `{}`, "", "source_binding_duplicate", "source bindings must not overlap for the same executor and action", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b, patch map[string]any
			if err := json.Unmarshal([]byte(binding), &b); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				if v == nil {
					delete(b, k)
				} else {
					b[k] = v
				}
			}
			bindings := []any{b}
			index := "0"
			if tc.duplicate {
				bindings = append(bindings, b)
				index = "1"
			}
			var destination map[string]any
			if err := json.Unmarshal([]byte(`{"executor":{"family":"native_database","id":"acme_stage_destination"},"eligible_actions":["stage_append"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append"}]}`), &destination); err != nil {
				t.Fatal(err)
			}
			destination["source_bindings"] = bindings
			raw, err := json.Marshal(map[string]any{"schema_version": 1, "destination_transport": destination})
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/sync_transport.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "sync_transport.json", "/destination_transport/source_bindings/"+index+tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicReadBack168(t *testing.T) {
	for _, owner := range []string{"read_back", "tombstone_read_back", "strategy_read_back", "strategy_tombstone_read_back"} {
		t.Run(owner, func(t *testing.T) {
			load := func(patch string) error {
				var policy, changes, destination map[string]any
				if err := json.Unmarshal([]byte(`{"operation":"read_widget","identity":[{"provider_field":"id","expected_field":"id"}],"expected":[{"provider_field":"name","expected_field":"name"}],"max_records":10,"max_attempts":2,"timeout_milliseconds":1000,"receipt_locator":{"response_index":0,"body_field":"id","query_parameter":"id","max_value_bytes":128,"max_pages":2}}`), &policy); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(owner, "tombstone") {
					delete(policy, "expected")
				}
				if err := json.Unmarshal([]byte(patch), &changes); err != nil {
					t.Fatal(err)
				}
				for k, v := range changes {
					policy[k] = v
				}
				if err := json.Unmarshal([]byte(`{"executor":{"family":"native_database","id":"acme_stage_destination"},"eligible_actions":["stage_append"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":"append","action":"stage_append"}]}`), &destination); err != nil {
					t.Fatal(err)
				}
				key := strings.TrimPrefix(owner, "strategy_")
				if strings.HasPrefix(owner, "strategy_") {
					destination["apply_strategies"].([]any)[0].(map[string]any)[key] = policy
				} else {
					destination[key] = policy
				}
				raw, err := json.Marshal(map[string]any{"schema_version": 1, "destination_transport": destination})
				if err != nil {
					t.Fatal(err)
				}
				files := fullValidBundleFS("acme")
				files["acme/sync_transport.json"] = &fstest.MapFile{Data: raw}
				_, err = Load(files, "acme")
				return err
			}
			if err := load(`{}`); err != nil {
				t.Fatalf("healthy read-back fixture: %v", err)
			}
			prefix := "/destination_transport/" + owner
			if strings.HasPrefix(owner, "strategy_") {
				prefix = "/destination_transport/apply_strategies/0/" + strings.TrimPrefix(owner, "strategy_")
			}
			for _, tc := range []struct{ name, patch, field, code, reason string }{
				{"operation", `{"operation":" "}`, "/operation", "readback_operation_invalid", "read-back requires a concrete operation"},
				{"records", `{"max_records":0}`, "/max_records", "readback_records_invalid", "read-back max_records must be between 1 and 10000"},
				{"attempts", `{"max_attempts":0}`, "/max_attempts", "readback_attempts_invalid", "read-back max_attempts must be between 1 and 10"},
				{"timeout", `{"timeout_milliseconds":0}`, "/timeout_milliseconds", "readback_timeout_invalid", "read-back timeout_milliseconds must be between 1 and 60000"},
				{"delay", `{"retry_delay_milliseconds":-1}`, "/retry_delay_milliseconds", "readback_delay_invalid", "read-back retry_delay_milliseconds must be between 0 and 10000"},
				{"provider_duplicate", `{"identity":[{"provider_field":"private-sentinel-167","expected_field":"id"},{"provider_field":"private-sentinel-167","expected_field":"name"}]}`, "/identity/1/provider_field", "readback_provider_duplicate", "read-back provider field must be unique"},
				{"expected_duplicate", `{"identity":[{"provider_field":"id","expected_field":"private-sentinel-167"},{"provider_field":"name","expected_field":"private-sentinel-167"}]}`, "/identity/1/expected_field", "readback_expected_duplicate", "read-back expected field must be unique"},
				{"locator_index", `{"receipt_locator":{"response_index":-1,"body_field":"id","query_parameter":"id","max_value_bytes":128,"max_pages":2}}`, "/receipt_locator/response_index", "receipt_locator_index_invalid", "receipt locator response_index must be between 0 and 1023"},
				{"locator_fields", `{"receipt_locator":{"response_index":0,"body_field":" ","query_parameter":"id","max_value_bytes":128,"max_pages":2}}`, "/receipt_locator", "receipt_locator_fields_invalid", "receipt locator requires concrete body_field and query_parameter"},
				{"locator_bytes", `{"receipt_locator":{"response_index":0,"body_field":"id","query_parameter":"id","max_value_bytes":0,"max_pages":2}}`, "/receipt_locator/max_value_bytes", "receipt_locator_bytes_invalid", "receipt locator max_value_bytes must be between 1 and 4096"},
				{"locator_pages", `{"receipt_locator":{"response_index":0,"body_field":"id","query_parameter":"id","max_value_bytes":128,"max_pages":0}}`, "/receipt_locator/max_pages", "receipt_locator_pages_invalid", "receipt locator max_pages must be between 1 and 10"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					err := load(tc.patch)
					if !publicBundleMatches167(err, "sync_transport.json", prefix+tc.field, tc.code, tc.reason) {
						t.Errorf("wrong public diagnostic: %v", err)
					}
				})
			}
		})
	}
}

func TestBundlePublicStreamGraphQLSiblings168(t *testing.T) {
	for _, tc := range []struct{ name, field, code, reason string }{
		{"body", "/streams/0", "stream_graphql_body_conflict", "stream must not declare both body and graphql"},
		{"method", "/streams/0/method", "stream_graphql_method", "GraphQL stream method must be POST"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc map[string]any
			if err := json.Unmarshal([]byte(validStreams), &doc); err != nil {
				t.Fatal(err)
			}
			stream := doc["streams"].([]any)[0].(map[string]any)
			stream["method"] = "POST"
			stream["graphql"] = map[string]any{"document": "query Widgets { widgets { id } }", "operation_name": "Widgets", "variables": map[string]any{}}
			if tc.name == "body" {
				stream["body"] = map[string]any{"private-sentinel-167": "private-sentinel-167"}
			} else {
				stream["method"] = "GET"
			}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/streams.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicPollingChangefeed168(t *testing.T) {
	const baseline = `{"status":"implemented","mechanism":"polling_watermark","source":{"artifact_url":"https://example.com/source","artifact_version":"v1","retrieved_at":"2026-09-01"},"executor":{"kind":"engine","id":"polling_watermark"},"checkpoint":{"kind":"watermark","keys":["updated_at","id"],"commit_after":"warehouse_durable","on_invalid":"fail"},"delivery":{"ordering":"per_stream","duplicates":"at_least_once","deletes":"not_available","dedupe_key":["id"]},"streams":["widgets"],"polling_watermark":{"watermark":{"kind":"timestamp","path":"updated_at"},"tie_breaker":{"path":"id"},"boundary":"inclusive","safety_lag_seconds":0,"page_size":10,"max_pages":2,"request_budget":2}}`
	load := func(scope, patch string) error {
		var doc, changes map[string]any
		if err := json.Unmarshal([]byte(baseline), &doc); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(patch), &changes); err != nil {
			t.Fatal(err)
		}
		target := doc
		if scope == "polling" {
			target = doc["polling_watermark"].(map[string]any)
		}
		for k, v := range changes {
			if v == nil {
				delete(target, k)
			} else {
				target[k] = v
			}
		}
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		files := fullValidBundleFS("acme")
		files["acme/changefeed.json"] = &fstest.MapFile{Data: raw}
		_, err = Load(files, "acme")
		return err
	}
	if err := load("root", `{}`); err != nil {
		t.Fatalf("healthy polling changefeed fixture: %v", err)
	}
	for _, tc := range []struct{ name, scope, patch, field, code, reason string }{
		{"missing", "root", `{"polling_watermark":null}`, "/polling_watermark", "changefeed_polling_required", "implemented polling changefeed requires polling_watermark declaration"},
		{"executor", "root", `{"executor":{"kind":"native","id":"polling_watermark"}}`, "/executor", "changefeed_polling_executor", "polling changefeed requires executor engine/polling_watermark"},
		{"watermark_path", "polling", `{"watermark":{"kind":"timestamp","path":"private-sentinel-167..x"}}`, "/polling_watermark/watermark/path", "changefeed_polling_path", "polling field path must contain safe nonempty identifier segments"},
		{"tie_path", "polling", `{"tie_breaker":{"path":" "}}`, "/polling_watermark/tie_breaker/path", "changefeed_polling_path", "polling field path must contain safe nonempty identifier segments"},
		{"negative_lag", "polling", `{"safety_lag_seconds":-1}`, "/polling_watermark/safety_lag_seconds", "changefeed_polling_lag_negative", "polling safety_lag_seconds must not be negative"},
		{"overflow_lag", "polling", `{"safety_lag_seconds":9223372037}`, "/polling_watermark/safety_lag_seconds", "changefeed_polling_lag_overflow", "polling safety_lag_seconds must fit a time duration"},
		{"lag_kind", "polling", `{"watermark":{"kind":"monotonic_sequence","path":"updated_at"},"safety_lag_seconds":1}`, "/polling_watermark/safety_lag_seconds", "changefeed_polling_lag_kind", "safety_lag_seconds is only valid for timestamp watermarks"},
		{"page_bound", "polling", `{"page_size":0}`, "/polling_watermark", "changefeed_polling_bounds", "polling requires positive page_size, max_pages and request_budget"},
		{"delete_budget", "polling", `{"request_budget":1,"deletion_endpoint":{"path":"/deleted","records_path":"data"}}`, "/polling_watermark/request_budget", "changefeed_polling_delete_budget", "deletion endpoint requires request_budget of at least 2"},
		{"checkpoint", "root", `{"checkpoint":{"kind":"watermark","keys":["id","updated_at"],"commit_after":"warehouse_durable","on_invalid":"fail"}}`, "/checkpoint/keys", "changefeed_polling_checkpoint", "checkpoint keys must be watermark path then tie_breaker path"},
		{"duplicates", "root", `{"delivery":{"ordering":"per_stream","duplicates":"none","deletes":"not_available","dedupe_key":["id"]}}`, "/delivery/duplicates", "changefeed_polling_duplicates", "polling delivery duplicates must be at_least_once"},
		{"delete_conflict", "polling", `{"soft_delete":{"path":"deleted"},"deletion_endpoint":{"path":"/deleted","records_path":"data"}}`, "/polling_watermark", "changefeed_polling_delete_conflict", "polling may declare soft_delete or deletion_endpoint, but not both"},
		{"soft_path", "polling", `{"soft_delete":{"path":"a..b"}}`, "/polling_watermark/soft_delete/path", "changefeed_polling_path", "polling field path must contain safe nonempty identifier segments"},
		{"endpoint_path", "polling", `{"deletion_endpoint":{"path":"/../private-sentinel-167","records_path":"data"}}`, "/polling_watermark/deletion_endpoint/path", "changefeed_polling_endpoint", "deletion endpoint must be a safe connector-relative path"},
		{"endpoint_records", "polling", `{"deletion_endpoint":{"path":"/deleted","records_path":"a..b"}}`, "/polling_watermark/deletion_endpoint/records_path", "changefeed_polling_path", "polling field path must contain safe nonempty identifier segments"},
		{"observable_deletes", "polling", `{"soft_delete":{"path":"deleted"}}`, "/delivery/deletes", "changefeed_polling_tombstones", "observable polling deletes require tombstone delivery"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := load(tc.scope, tc.patch)
			if !publicBundleMatches167(err, "changefeed.json", tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}

func TestBundlePublicDatabase168(t *testing.T) {
	baseline, err := fs.ReadFile(defs.FS, "postgres/database.json")
	if err != nil {
		t.Fatal(err)
	}
	load := func(raw []byte) error {
		files := fullValidBundleFS("acme")
		files["acme/database.json"] = &fstest.MapFile{Data: raw}
		_, err := Load(files, "acme")
		return err
	}
	if err := load(baseline); err != nil {
		t.Fatalf("healthy database fixture: %v", err)
	}
	for _, tc := range []struct{ name, old, replacement, path, reason string }{
		{"null", `"driver": {`, `"driver": null, "ignored": {`, "$.driver", "null is not permitted"},
		{"unknown", `"id": "postgres"`, `"private-sentinel-167": "secret"`, "$.driver", "unknown member is not permitted"},
		{"case_alias", `"id": "postgres"`, `"ID": "postgres"`, "$.driver.id", "case-aliased member is not permitted"},
		{"duplicate", `"id": "postgres"`, `"id": "postgres", "id":"private-sentinel-167"`, "$.driver.id", "duplicate member is not permitted"},
		{"required", `"id": "postgres",`, ``, "$.driver.id", "required member is missing"},
		{"integer_type", `"api_version": 1`, `"api_version": "private-sentinel-167"`, "$.driver.api_version", "JSON must match the closed schema"},
		{"minimum", `"max_bytes": 63`, `"max_bytes": 0`, "$.identifiers.max_bytes", "value violates the declared minimum"},
		{"maximum", `"max_bytes": 63`, `"max_bytes": 999999999`, "$.identifiers.max_bytes", "value violates the declared maximum"},
		{"enum", `"schema_version": 1`, `"schema_version": 2`, "$.schema_version", "value violates the declared enum"},
		{"driver", `"id": "postgres"`, `"id": "private-sentinel-167 bad"`, "$.driver", "database driver declaration is invalid"},
		{"catalog", `["schema", "relation"]`, `["relation", "schema"]`, "$.catalog", "database catalog qualification policy is invalid"},
		{"identifiers", `"quote_style": "double_quote"`, `"quote_style": "private-sentinel-167"`, "$.identifiers", "database identifier policy is invalid"},
		{"resources", `"default": 2, "maximum": 8`, `"default": 9, "maximum": 8`, "$.resources", "database resource policy is invalid"},
		{"native", `"name": "int2"`, `"name": "private-sentinel-167 bad"`, "$.type_mappings[0]", "database type mapping is invalid"},
		{"logical", `"kind": "signed_integer", "bits": 16`, `"kind": "private-sentinel-167"`, "$.type_mappings[0].logical", "database logical type declaration is invalid"},
		{"mapping_duplicate", `"name": "int4"`, `"name": "int2"`, "$.type_mappings[1]", "database definition contains duplicate native type mappings"},
		{"mode", `"full_overwrite"`, `"private-sentinel-167"`, "$.admitted_modes[0]", "database definition declares an unsupported sync mode"},
		{"mode_duplicate", `"full_append"`, `"full_overwrite"`, "$.admitted_modes[1]", "database definition declares a duplicate sync mode"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(string(baseline), tc.old) {
				t.Fatal("fixture replacement did not match")
			}
			raw := strings.Replace(string(baseline), tc.old, tc.replacement, 1)
			expectedPath := tc.path
			if tc.name == "unknown" {
				expectedPath += "@byte:" + strconv.Itoa(strings.Index(raw, `"private-sentinel-167"`)+len(`"private-sentinel-167"`))
			}
			err := load([]byte(raw))
			if !publicBundleMatches167(err, "database.json", expectedPath, "database_definition_invalid", tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
			if !errors.Is(err, database.ErrInvalidDefinition) {
				t.Errorf("lost database sentinel: %v", err)
			}
		})
	}
}

type databaseReadFailure168 struct {
	fs.FS
	failure error
	reads   *int
}

func (f databaseReadFailure168) ReadFile(name string) ([]byte, error) {
	if name == "acme/database.json" {
		*f.reads++
		if *f.reads > 1 {
			return nil, f.failure
		}
	}
	return fs.ReadFile(f.FS, name)
}
func TestBundleDatabaseReadCause168(t *testing.T) {
	files := fullValidBundleFS("acme")
	files["acme/database.json"] = &fstest.MapFile{Data: []byte(`{}`)}
	sentinel := errors.New("private-sentinel-167 read failure")
	reads := 0
	_, err := Load(databaseReadFailure168{FS: files, failure: sentinel, reads: &reads}, "acme")
	if reads != 2 {
		t.Fatalf("database read observer saw %d reads, want identity then loader", reads)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("lost original read failure: %v", err)
	}
	if !publicBundleMatches167(err, "database.json", "$", "database_definition_invalid", "database.json is unavailable") {
		t.Errorf("wrong public diagnostic: %v", err)
	}
}

func TestBundleMixedLoadAllAndSourceIndependence169(t *testing.T) {
	files := fullValidBundleFS("acme")
	for _, name := range []string{"beta", "broken-rate", "broken-stream"} {
		for path, file := range fullValidBundleFS(name) {
			files[path] = file
		}
	}
	files["broken-rate/rate_limits.json"] = &fstest.MapFile{Data: []byte(`{"state":}`)}
	files["broken-stream/streams.json"] = &fstest.MapFile{Data: []byte(`{"base":}`)}
	before, err := Load(files, "acme")
	if err != nil {
		t.Fatalf("healthy selected control: %v", err)
	}
	// These deliberately malformed authoring inputs must not participate in the
	// execution identity or decoder. No source evidence can promote execution.
	files["acme/source.lock.json"] = &fstest.MapFile{Data: []byte(`private-sentinel-167 invalid source`)}
	files["acme/source.proof.json"] = &fstest.MapFile{Data: []byte(`private-sentinel-167 invalid proof`)}
	after, err := Load(files, "acme")
	if err != nil {
		t.Fatalf("source-only files changed healthy selection: %v", err)
	}
	if before.Identity != after.Identity || before.Name != after.Name || len(after.Streams) != len(before.Streams) {
		t.Fatalf("source-only files changed execution projection: before=%+v after=%+v", before.Identity, after.Identity)
	}
	bundles, err := LoadAll(files)
	var all *LoadAllError
	if !errors.As(err, &all) {
		t.Fatalf("missing aggregate failure: %v", err)
	}
	if len(bundles) != 2 || bundles[0].Name != "acme" || bundles[1].Name != "beta" {
		t.Fatalf("wrong retained healthy identities: %v", bundles)
	}
	failures := all.GetFailures()
	if len(failures) != 2 {
		t.Fatalf("wrong failure count: %v", failures)
	}
	for i, want := range []struct{ name, file string }{{"broken-rate", "rate_limits.json"}, {"broken-stream", "streams.json"}} {
		failure := failures[i]
		var d *BundleDiagnosticError
		var syntax *json.SyntaxError
		if failure.Name != want.name || !errors.As(failure.Err, &d) || !errors.As(failure.Err, &syntax) {
			t.Fatalf("wrong per-failure graph: %v", failure)
		}
		if d.Connector != want.name || d.Generation != "embedded-v1" || d.File != want.file || d.Field != "/" || d.ReasonCode != "invalid_json" || d.Reason != "malformed JSON" || syntax.Offset == 0 {
			t.Fatalf("wrong per-failure safe metadata: %+v", d)
		}
		raw, e := json.Marshal(d)
		if e != nil || strings.Contains(string(raw)+err.Error(), "private-sentinel-167") {
			t.Fatalf("unsafe aggregate or selected projection: %s %v", raw, err)
		}
	}
}
