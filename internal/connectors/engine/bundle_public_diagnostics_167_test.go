package engine

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
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
			files["acme/operations.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(files, "acme")
			if !publicBundleMatches167(err, "operations.json", "/operations/0"+tc.field, tc.code, tc.reason) {
				t.Errorf("wrong public diagnostic: %v", err)
			}
		})
	}
}
