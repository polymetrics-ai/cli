package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMultipartFieldDiagnosticCoordinates171(t *testing.T) {
	for _, owner := range []string{"writes", "operations"} {
		for _, property := range []string{"healthy", "allowed_media_types", "media_policy"} {
			t.Run(owner+"/"+property, func(t *testing.T) {
				var rest map[string]any
				if err := json.Unmarshal([]byte(validMultipartRestWrite), &rest); err != nil {
					t.Fatal(err)
				}
				part := rest["multipart"].(map[string]any)["parts"].([]any)[0].(map[string]any)
				if property == "allowed_media_types" {
					part[property] = []string{"text/plain"}
				}
				if property == "media_policy" {
					part[property] = "provider_unrestricted"
				}
				raw, err := json.Marshal(rest)
				if err != nil {
					t.Fatal(err)
				}
				files := multipartRestWriteBundleFS(string(raw), "rest_write")
				file, prefix := "operations.json", "/operations/0/rest/multipart/parts/0/"
				if owner == "writes" {
					files = fullValidBundleFS("acme")
					action := map[string]any{"name": "attach", "kind": "create", "risk": "low", "method": "POST", "path": "/attachments", "body_type": "multipart", "multipart": rest["multipart"], "record_schema": rest["body_schema"]}
					raw, err = json.Marshal(map[string]any{"actions": []any{action}})
					if err != nil {
						t.Fatal(err)
					}
					files["acme/writes.json"] = &fstest.MapFile{Data: raw}
					file, prefix = "writes.json", "/actions/0/multipart/parts/0/"
				}
				_, err = Load(files, "acme")
				if property == "healthy" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				code, reason := "multipart_media_types_type_invalid", "allowed_media_types is only meaningful on a file part"
				if property == "media_policy" {
					code, reason = "multipart_media_policy_type_invalid", "media policy is only meaningful on a file part"
				}
				if !publicBundleMatches167(err, file, prefix+property, code, reason) {
					t.Errorf("wrong independently expected property/reason: %v", err)
				}
				var diagnostic *BundleDiagnosticError
				if !errors.As(err, &diagnostic) || diagnostic.Cause == nil || !strings.Contains(diagnostic.Cause.Error(), property) {
					t.Errorf("original property cause lost: %v", err)
				}
			})
		}
	}
}

func TestChangeApplyDiagnosticCoordinates171(t *testing.T) {
	for _, strategy := range []string{"append", "change_apply"} {
		t.Run(strategy, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/sync_transport.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"schema_version":1,"destination_transport":{"executor":{"family":"native_database","id":"acme_stage_destination"},"eligible_actions":["stage_append"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"not_available"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":%q,"action":"stage_append"}]}}`, strategy))}
			_, err := Load(files, "acme")
			if strategy == "append" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !publicBundleMatches167(err, "sync_transport.json", "/destination_transport/apply_strategies/0/strategy", "transport_strategy_mode_conflict", "change_apply strategy is only valid for change_capture mode") {
				t.Errorf("wrong strategy property/reason: %v", err)
			}
			var d *BundleDiagnosticError
			if !errors.As(err, &d) || d.Cause == nil || !strings.Contains(d.Cause.Error(), "destination change_apply strategy is only valid for change_capture mode") {
				t.Errorf("original strategy cause lost: %v", err)
			}
		})
	}
}

func TestDiagnosticCoordinateOraclesRejectPerturbations171(t *testing.T) {
	for _, tc := range []struct{ file, field, code, reason string }{
		{"writes.json", "/actions/0/multipart/parts/0/allowed_media_types", "multipart_media_types_type_invalid", "allowed_media_types is only meaningful on a file part"},
		{"operations.json", "/operations/0/rest/multipart/parts/0/allowed_media_types", "multipart_media_types_type_invalid", "allowed_media_types is only meaningful on a file part"},
		{"sync_transport.json", "/destination_transport/apply_strategies/0/strategy", "transport_strategy_mode_conflict", "change_apply strategy is only valid for change_capture mode"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/diagnostic_coordinates_171.json")
			if err != nil {
				t.Fatal(err)
			}
			var fixtures map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fixtures); err != nil {
				t.Fatal(err)
			}
			files := fullValidBundleFS("acme")
			files["acme/"+tc.file] = &fstest.MapFile{Data: fixtures[strings.TrimSuffix(tc.file, ".json")]}
			_, err = Load(files, "acme")
			var d *BundleDiagnosticError
			if !errors.As(err, &d) || !publicBundleMatches167(err, tc.file, tc.field, tc.code, tc.reason) {
				t.Fatalf("actual baseline tuple: %v", err)
			}
			for _, property := range []string{"field", "reason", "code"} {
				perturbed := *d
				switch property {
				case "field":
					perturbed.Field = "/"
				case "reason":
					perturbed.Reason = "invalid bundle declaration"
				case "code":
					perturbed.ReasonCode = "bundle_invalid"
				}
				if publicBundleMatches167(&perturbed, tc.file, tc.field, tc.code, tc.reason) {
					t.Fatalf("oracle accepted isolated %s corruption", property)
				}
			}
		})
	}
}
