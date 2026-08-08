package dockerhub_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var (
	dockerHubRawPathParameter    = regexp.MustCompile(`\{[A-Za-z_][A-Za-z0-9_]*\}`)
	dockerHubRecordPathParameter = regexp.MustCompile(`\{\{\s*record\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
)

func TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}

	for _, action := range bundle.Writes {
		t.Run(action.Name, func(t *testing.T) {
			if strings.HasPrefix(action.Path, "/v2/") {
				t.Errorf("path = %q, want engine-relative path without the base URL's /v2 prefix", action.Path)
			}
			if raw := dockerHubRawPathParameter.FindString(action.Path); raw != "" {
				t.Errorf("path = %q retains raw OpenAPI parameter %q, want a {{ record.* }} template", action.Path, raw)
			}

			pathFields := make(map[string]bool, len(action.PathFields))
			for _, field := range action.PathFields {
				pathFields[field] = true
			}
			for _, match := range dockerHubRecordPathParameter.FindAllStringSubmatch(action.Path, -1) {
				if !pathFields[match[1]] {
					t.Errorf("path template %q is missing %q from path_fields %v", match[0], match[1], action.PathFields)
				}
			}
		})
	}

	preview, err := engine.DryRunWrite(context.Background(), bundle, connectors.WriteRequest{
		Action: "create_repository",
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"base_url": "https://hub.docker.com/v2",
		}},
	}, []connectors.Record{{
		"name":       "fixture-repository",
		"namespace":  "polymetrics",
		"is_private": true,
	}}, nil)
	if err != nil {
		t.Fatalf("dry-run create_repository: %v", err)
	}

	const wantRequest = "POST https://hub.docker.com/v2/namespaces/polymetrics/repositories"
	if warnings := strings.Join(preview.Warnings, "\n"); !strings.Contains(warnings, wantRequest) {
		t.Errorf("create_repository preview warnings = %q, want resolved request %q", warnings, wantRequest)
	}
}
