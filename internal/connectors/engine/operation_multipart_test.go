package engine

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

const validMultipartRestWrite = `{
	"method": "POST",
	"path": "/attachments",
	"content_type": "multipart/form-data",
	"max_bytes": 1024,
	"body_schema": {
		"type": "object",
		"additionalProperties": false,
		"required": ["message", "media_file_path"],
		"properties": {
			"message": {"type": "string"},
			"media_file_path": {"type": "string"}
		}
	},
	"multipart": {
		"max_bytes": 2048,
		"parts": [
			{"name": "message", "type": "field", "field": "message", "required": true},
			{
				"name": "attachment",
				"type": "file",
				"field": "media_file_path",
				"required": true,
				"max_bytes": 1024,
				"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]
			}
		]
	}
}`

func multipartRestWriteBundleFS(rest, kind string) fstest.MapFS {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{
		"operations": [{
			"id": "acme.attachments.create",
			"kind": %q,
			"summary": "Attach one declared local file",
			"risk": "high",
			"approval": "plan-preview-confirm-execute",
			"output_policy": "json",
			"mutation_class": "destructive",
			"confirmation": {"kind": "destructive"},
			"rest": %s
		}]
	}`, kind, rest))}
	return fsys
}

func TestBundleLoadAcceptsTypedMultipartRestWriteContract(t *testing.T) {
	_, err := Load(multipartRestWriteBundleFS(validMultipartRestWrite, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load typed multipart rest_write: %v", err)
	}
}

func TestOperationDirectWriteMetadataRecognizesTypedMultipartRestWrite(t *testing.T) {
	bundle, err := Load(multipartRestWriteBundleFS(validMultipartRestWrite, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load typed multipart rest_write: %v", err)
	}
	metadata, err := OperationDirectWriteMetadata(bundle, "acme.attachments.create")
	if err != nil {
		t.Fatalf("OperationDirectWriteMetadata typed multipart rest_write: %v", err)
	}
	if len(metadata.PayloadFileFields) != 1 || metadata.PayloadFileFields[0] != "media_file_path" {
		t.Fatalf("multipart payload file fields = %#v, want the declared source path", metadata.PayloadFileFields)
	}
	if metadata.PayloadFileMaxBytes["media_file_path"] != 1024 {
		t.Fatalf("multipart payload file cap = %#v, want declared cap", metadata.PayloadFileMaxBytes)
	}
}

func TestBundleLoadRejectsUnsafeMultipartRestWriteContracts(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		rest    string
		wantErr string
	}{
		{
			name:    "rest read cannot declare multipart",
			kind:    "rest_read",
			rest:    validMultipartRestWrite,
			wantErr: "multipart",
		},
		{
			name:    "provider search cannot declare multipart",
			kind:    "provider_search",
			rest:    validMultipartRestWrite,
			wantErr: "multipart",
		},
		{
			name: "content type must be literal multipart form data",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"content_type": "multipart/form-data"`,
				`"content_type": "multipart/form-data; boundary=caller-controlled"`,
				1,
			),
			wantErr: "multipart/form-data",
		},
		{
			name: "endpoint must be connector relative",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"path": "/attachments"`,
				`"path": "https://untrusted.example/attachments"`,
				1,
			),
			wantErr: "connector-relative",
		},
		{
			name: "response capture must be bounded separately",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 1024,`,
				`"max_bytes": 0,`,
				1,
			),
			wantErr: "response",
		},
		{
			name: "aggregate upload cap is required",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 2048,`,
				`"max_bytes": 0,`,
				1,
			),
			wantErr: "aggregate",
		},
		{
			name: "parts cannot be empty",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"parts": [
			{"name": "message", "type": "field", "field": "message", "required": true},
			{
				"name": "attachment",
				"type": "file",
				"field": "media_file_path",
				"required": true,
				"max_bytes": 1024,
				"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]
			}
		]`,
				`"parts": []`,
				1,
			),
			wantErr: "parts",
		},
		{
			name: "body schema must be closed",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"additionalProperties": false`,
				`"additionalProperties": true`,
				1,
			),
			wantErr: "additionalProperties",
		},
		{
			name: "body schema is required",
			kind: "rest_write",
			rest: `{
				"method": "POST",
				"path": "/attachments",
				"content_type": "multipart/form-data",
				"max_bytes": 1024,
				"multipart": {
					"max_bytes": 2048,
					"parts": [{
						"name": "attachment",
						"type": "file",
						"field": "media_file_path",
						"required": true,
						"max_bytes": 1024
					}]
				}
			}`,
			wantErr: "body_schema",
		},
		{
			name: "every part names a declared body field",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"field": "message"`,
				`"field": "unknown"`,
				1,
			),
			wantErr: "declared body field",
		},
		{
			name: "file source must be a required string",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"required": ["message", "media_file_path"]`,
				`"required": ["message"]`,
				1,
			),
			wantErr: "required string",
		},
		{
			name: "inline bytes cannot substitute for a file source path",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"media_file_path": {"type": "string"}`,
				`"media_file_path": {"type": "array", "items": {"type": "integer"}}`,
				1,
			),
			wantErr: "required string",
		},
		{
			name: "file source must have a positive cap",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 1024,
				"content_type": "text/plain"`,
				`"max_bytes": 0,
				"content_type": "text/plain"`,
				1,
			),
			wantErr: "file part",
		},
		{
			name: "file source must declare media policy",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				"\"max_bytes\": 1024,\n\t\t\t\t\"content_type\": \"text/plain\",\n\t\t\t\t\"allowed_media_types\": [\"text/plain\"]",
				"\"max_bytes\": 1024",
				1,
			),
			wantErr: "media policy",
		},
		{
			name:    "legacy file upload does not become an operation executor",
			kind:    "file_upload",
			rest:    validMultipartRestWrite,
			wantErr: "file",
		},
	}

	public := map[string]struct{ field, code, reason string }{
		"rest read cannot declare multipart":                       {"/operations/0/rest/multipart", "multipart_kind_invalid", "multipart is only valid for rest_write operations"},
		"provider search cannot declare multipart":                 {"/operations/0/rest/multipart", "multipart_kind_invalid", "multipart is only valid for rest_write operations"},
		"content type must be literal multipart form data":         {"/operations/0/rest/content_type", "multipart_content_type_invalid", "multipart requires literal content_type multipart/form-data"},
		"endpoint must be connector relative":                      {"/operations/0/rest/path", "multipart_path_invalid", "multipart endpoint must be connector-relative"},
		"response capture must be bounded separately":              {"/operations/0/rest/max_bytes", "multipart_response_bound_invalid", "multipart requires positive response capture max_bytes"},
		"aggregate upload cap is required":                         {"/operations/0/rest/multipart/max_bytes", "multipart_aggregate_bound_invalid", "multipart requires positive aggregate max_bytes"},
		"parts cannot be empty":                                    {"/operations/0/rest/multipart/parts", "array_too_short", "array has too few items"},
		"body schema must be closed":                               {"/operations/0/rest/body_schema/additionalProperties", "multipart_schema_open", "multipart body object must declare additionalProperties: false"},
		"body schema is required":                                  {"/operations/0/rest/body_schema", "multipart_schema_missing", "multipart requires body_schema"},
		"every part names a declared body field":                   {"/operations/0/rest/multipart/parts/0/field", "multipart_field_invalid", "multipart part must reference a declared body field"},
		"file source must be a required string":                    {"/operations/0/rest/multipart/parts/1/field", "multipart_file_field_invalid", "multipart file part must reference a required string body field"},
		"inline bytes cannot substitute for a file source path":    {"/operations/0/rest/multipart/parts/1/field", "multipart_file_field_invalid", "multipart file part must reference a required string body field"},
		"file source must have a positive cap":                     {"/operations/0/rest/multipart/parts/1/max_bytes", "multipart_file_bound_invalid", "multipart file part requires positive max_bytes"},
		"file source must declare media policy":                    {"/operations/0/rest/multipart/parts/1", "multipart_media_policy_missing", "multipart file part requires declared media policy"},
		"legacy file upload does not become an operation executor": {"/operations/0/kind", "operation_execution_kind_mismatch", "execution block must match the operation kind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(multipartRestWriteBundleFS(tt.rest, tt.kind), "acme")
			if err == nil {
				t.Fatal("Load unsafe multipart declaration: error = nil")
			}
			want, ok := public[tt.name]
			if !ok || !publicBundleMatches167(err, "operations.json", want.field, want.code, want.reason) {
				t.Fatalf("wrong public multipart diagnostic: %v", err)
			}

			if !strings.Contains(strings.ToLower(bundleCauseText165(err)), strings.ToLower(tt.wantErr)) {
				t.Fatalf("Load unsafe multipart declaration error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
