package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMultipartPolicyAdmission217(t *testing.T) {
	for _, owner := range []string{"writes", "operations"} {
		for _, scenario := range []string{"healthy", "identity", "unknown_encoding", "field_encoding", "zero_metadata", "negative_metadata", "wrong_total", "overflow", "duplicate"} {
			t.Run(owner+"/"+scenario, func(t *testing.T) {
				var rest map[string]any
				if err := json.Unmarshal([]byte(validMultipartRestWrite), &rest); err != nil {
					t.Fatal(err)
				}
				spec := rest["multipart"].(map[string]any)
				parts := spec["parts"].([]any)
				file := parts[1].(map[string]any)
				spec["max_metadata_bytes"] = 1024
				file["filename_encoding"] = "url_percent_utf8"
				switch scenario {
				case "identity":
					file["filename_encoding"] = "identity"
				case "unknown_encoding":
					file["filename_encoding"] = "base64"
				case "field_encoding":
					parts[0].(map[string]any)["filename_encoding"] = "identity"
				case "zero_metadata":
					spec["max_metadata_bytes"] = 0
				case "negative_metadata":
					spec["max_metadata_bytes"] = -1
				case "wrong_total":
					spec["max_bytes"] = 2047
				case "overflow":
					file["max_bytes"] = json.Number("9223372036854775807")
				case "duplicate":
					file["name"] = "message"
				}
				raw, err := json.Marshal(rest)
				if err != nil {
					t.Fatal(err)
				}
				files := multipartRestWriteBundleFS(string(raw), "rest_write")
				if owner == "writes" {
					files = fullValidBundleFS("acme")
					action := map[string]any{"name": "attach", "kind": "create", "risk": "low", "method": "POST", "path": "/attachments", "body_type": "multipart", "multipart": spec, "record_schema": rest["body_schema"]}
					raw, err = json.Marshal(map[string]any{"actions": []any{action}})
					if err != nil {
						t.Fatal(err)
					}
					files["acme/writes.json"] = &fstest.MapFile{Data: raw}
				}
				_, err = Load(files, "acme")
				if scenario == "healthy" || scenario == "identity" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if err == nil {
					t.Fatalf("admitted invalid multipart contract %s", scenario)
				}
			})
		}
	}
}

func TestMultipartPolicyWire217(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(io.LimitReader(r.Body, 2049))
		if err != nil {
			t.Error(err)
		}
		body = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"server_value":"complete"}`))
	}))
	defer server.Close()
	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "résumé +%.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	budget := int64(1024)
	bundle.Operations[0].REST.Multipart.MaxMetadataBytes = &budget
	bundle.Operations[0].REST.Multipart.Parts[1].FilenameEncoding = "url_percent_utf8"
	req := multipartOperationRequest(dir, path, payload)
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		t.Fatal("preview sent a request")
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	bundle.Operations[0].REST.Multipart.Parts[1].FilenameEncoding = "identity"
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "no longer matches") {
		t.Fatalf("changed filename policy retained approval: %v", err)
	}
	if body != "" {
		t.Fatal("stale filename policy sent")
	}
	bundle.Operations[0].REST.Multipart.Parts[1].FilenameEncoding = "url_percent_utf8"
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `filename="r%C3%A9sum%C3%A9%20%2B%25.txt"`) || !strings.Contains(body, string(payload)) || strings.Contains(body, "filename*=") {
		t.Fatalf("unexpected multipart wire %q", body)
	}
	if got, ok := result.Body.(map[string]any); !ok || got["server_value"] != "complete" {
		t.Fatalf("returned body %#v", result.Body)
	}
}

func TestMultipartPolicyPreviewName217(t *testing.T) {
	var sends int
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { sends++ }))
	defer server.Close()
	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "invalid\nname.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	bundle.Operations[0].REST.Multipart.Parts[1].FilenameEncoding = "url_percent_utf8"
	req := multipartOperationRequest(dir, path, payload)
	if _, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Error("preview admitted an invalid wire filename")
	}
	if sends != 0 {
		t.Fatalf("preview sent %d requests", sends)
	}
}
