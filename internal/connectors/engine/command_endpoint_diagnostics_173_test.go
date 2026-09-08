package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCLITargetDiagnostics173(t *testing.T) {
	for _, owner := range []string{"foundation_gap", "unsupported_disposition"} {
		t.Run(owner, func(t *testing.T) {
			for _, invalid := range []bool{false, true} {
				name := "healthy"
				path := "/widgets"
				if invalid {
					name = "invalid_path"
					path = "/../private-sentinel-171"
				}
				t.Run(name, func(t *testing.T) {
					target := map[string]any{"source_id": "source-171", "operation_id": "operation-171", "method": "GET", "path": path}
					disposition := map[string]any{"reason": "fixture unsupported operation", "target": target}
					availability := "unsupported_with_provider_evidence"
					if owner == "foundation_gap" {
						availability = "deferred"
						disposition["id"] = "fixture_gap"
						disposition["component"] = "runtime_executor"
						disposition["evidence"] = "runtime_executor_absent"
					}
					raw, e := json.Marshal(map[string]any{"tagline": "Fixture", "usage": "acme", "commands": []any{map[string]any{"path": "widgets list", "summary": "Fixture", "intent": "direct_read", "availability": availability, owner: disposition}}})
					if e != nil {
						t.Fatal(e)
					}
					files := fullValidBundleFS("acme")
					files["acme/cli_surface.json"] = &fstest.MapFile{Data: raw}
					_, e = Load(files, "acme")
					if !invalid {
						if e != nil {
							t.Fatalf("healthy Load setup: %v", e)
						}
						return
					}
					var d *BundleDiagnosticError
					if !errors.As(e, &d) || d.Cause == nil || !strings.Contains(d.Cause.Error(), "noncanonical path segment") {
						t.Fatalf("intended semantic frontier not reached: %v", e)
					}
					expected := fmt.Sprintf("/commands/0/%s/target/path", owner)
					if d.File != "cli_surface.json" || d.Field != expected || d.ReasonCode == "bundle_invalid" || d.Reason == "invalid bundle declaration" {
						t.Fatalf("semantic target diagnostic lacks useful exact property: file=%q field=%q code=%q reason=%q", d.File, d.Field, d.ReasonCode, d.Reason)
					}
					if strings.Contains(e.Error(), "private-sentinel-171") {
						t.Fatal("public output exposed authored target")
					}
				})
			}
		})
	}
}

func TestCLITargetSemanticSiblings173(t *testing.T) {
	cases := []struct{ name, method, path, property, code, reason, cause string }{
		{"shape", "GET", "relative", "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"},
		{"whitespace", "GET", " /widgets", "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"},
		{"query", "GET", "/widgets?private=sentinel", "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"},
		{"fragment", "GET", "/widgets#private", "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"},
		{"slash", "GET", "/widgets\\private", "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"},
		{"encoding", "GET", "/%private", "path", "command_target_path_encoding_invalid", "command target path contains invalid percent encoding", "invalid percent encoding"},
		{"dot", "GET", "/%2e%2e/private", "path", "command_target_path_segment_invalid", "command target path contains a noncanonical segment", "noncanonical path segment"},
		{"encoded_slash", "GET", "/%2fprivate", "path", "command_target_path_segment_invalid", "command target path contains a noncanonical segment", "noncanonical path segment"},
		{"control", "GET", "/private\u200b", "path", "command_target_path_characters_invalid", "command target path contains forbidden characters", "invalid unicode characters"},
		{"graphql_shape", "GRAPHQL", " operation", "path", "command_target_graphql_shape_invalid", "GRAPHQL command target requires a bounded canonical operation identity", "bounded canonical operation identity"},
		{"graphql_identifier", "GRAPHQL", "operation/private", "path", "command_target_graphql_identifier_invalid", "GRAPHQL command target must be a safe operation identifier", "invalid character"},
		{"method", "OPTIONS", "/widgets", "method", "command_target_method_invalid", "command target method is not supported", "supported canonical method"},
	}
	cases = append(cases, struct{ name, method, path, property, code, reason, cause string }{"bound", "GET", "/" + strings.Repeat("x", 8192), "path", "command_target_path_invalid", "command target path must be bounded, canonical, and connector-relative", "canonical connector-relative"})
	for _, owner := range []string{"foundation_gap", "unsupported_disposition"} {
		for _, tc := range cases {
			if (tc.name == "method" || tc.name == "bound") && owner == "unsupported_disposition" {
				continue
			} // schema owns this refusal
			t.Run(owner+"/"+tc.name, func(t *testing.T) {
				target := map[string]any{"source_id": "source-173", "operation_id": "operation-173", "method": tc.method, "path": tc.path}
				disposition := map[string]any{"reason": "fixture unsupported operation", "target": target}
				availability := "unsupported_with_provider_evidence"
				if owner == "foundation_gap" {
					availability = "deferred"
					disposition["id"] = "fixture_gap"
					disposition["component"] = "runtime_executor"
					disposition["evidence"] = "runtime_executor_absent"
				}
				raw, err := json.Marshal(map[string]any{"tagline": "Fixture", "usage": "acme", "commands": []any{map[string]any{"path": "widgets list", "summary": "Fixture", "intent": "direct_read", "availability": availability, owner: disposition}}})
				if err != nil {
					t.Fatal(err)
				}
				files := fullValidBundleFS("acme")
				files["acme/cli_surface.json"] = &fstest.MapFile{Data: raw}
				_, err = Load(files, "acme")
				var d *BundleDiagnosticError
				if !errors.As(err, &d) || d.Cause == nil || !strings.Contains(d.Cause.Error(), tc.cause) {
					t.Fatalf("semantic frontier: %v", err)
				}
				if d.File != "cli_surface.json" || d.Field != "/commands/0/"+owner+"/target/"+tc.property || d.ReasonCode != tc.code || d.Reason != tc.reason {
					t.Fatalf("diagnostic tuple: %+v", d)
				}
			})
		}
	}
}
