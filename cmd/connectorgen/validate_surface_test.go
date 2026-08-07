package main

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestCheckAPISurface_POSTDirectReadDoesNotRequireWriteCapability(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: false},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "POST",
				Path:      "/freeBusy",
				CoveredBy: &engine.SurfaceCoverage{DirectRead: "freebusy query"},
			}},
		},
		CLISurface: &engine.CLISurface{
			Commands: []engine.CLICommand{{
				Path:         "freebusy query",
				Intent:       "direct_read",
				Availability: "implemented",
			}},
		},
	}

	if findings := checkAPISurface(b); len(findings) != 0 {
		t.Fatalf("checkAPISurface rejected POST direct read with write disabled: %+v", findings)
	}
}

// One endpoint can back more than one write action, and covered_by.write is a
// single string. github ships three actions on PATCH /repos/{owner}/{repo}/issues/
// {issue_number} -- update_issue, close_issue and reopen_issue -- because the
// close and reopen bodies are distinct contracts assembled by a Go hook that
// switches on the action NAME (internal/connectors/hooks/github/hooks.go), and
// certify's create/cleanup pairing binds create_issue to close_issue.
//
// Before covered_by.writes existed, the only way to reference all three was to
// invent a second path -- "PATCH .../issues/{issue_number} (close)" -- encoding
// a behaviour variant into the path. No such path is documented by any provider,
// and it corrupts every documented-operation count taken from api_surface.json.
// The plural array mirrors covered_by.direct_reads, which already solved exactly
// this shape for reads.
func TestCheckAPISurface_EndpointMayBackMultipleWriteActions(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: true},
		},
		Writes: []engine.WriteAction{
			{Name: "update_widget", Method: "PATCH", Path: "/widgets/{{ record.id }}"},
			{Name: "close_widget", Method: "PATCH", Path: "/widgets/{{ record.id }}"},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "PATCH",
				Path:      "/widgets/{id}",
				CoveredBy: &engine.SurfaceCoverage{Writes: []string{"update_widget", "close_widget"}},
			}},
		},
	}

	if findings := checkAPISurface(b); len(findings) != 0 {
		t.Fatalf("checkAPISurface rejected two write actions on one endpoint: %+v", findings)
	}
}

// A plural entry naming a write action the bundle does not declare is still a
// finding: widening the shape must not widen what goes unchecked.
func TestCheckAPISurface_PluralWriteCoverageStillRejectsUnknownTargets(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: true},
		},
		Writes: []engine.WriteAction{
			{Name: "update_widget", Method: "PATCH", Path: "/widgets/{{ record.id }}"},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "PATCH",
				Path:      "/widgets/{id}",
				CoveredBy: &engine.SurfaceCoverage{Writes: []string{"update_widget", "no_such_action"}},
			}},
		},
	}

	findings := checkAPISurface(b)
	if len(findings) != 1 {
		t.Fatalf("want exactly one finding for the undeclared write action, got %+v", findings)
	}
	if !strings.Contains(findings[0].Message, "no_such_action") {
		t.Fatalf("finding does not name the undeclared action: %q", findings[0].Message)
	}
}
