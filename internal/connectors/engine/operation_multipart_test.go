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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(multipartRestWriteBundleFS(tt.rest, tt.kind), "acme")
			if err == nil {
				t.Fatal("Load unsafe multipart declaration: error = nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("Load unsafe multipart declaration error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
