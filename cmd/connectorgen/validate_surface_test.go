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

func TestCheckAPISurface_DirectWriteCoverageRequiresWriteCapability(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: true},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "POST",
				Path:      "/votes",
				CoveredBy: &engine.SurfaceCoverage{DirectWrite: "vote create"},
			}},
		},
		CLISurface: &engine.CLISurface{
			Commands: []engine.CLICommand{{
				Path:         "vote create",
				Intent:       "direct_write",
				Availability: "implemented",
			}},
		},
	}

	if findings := checkAPISurface(b); len(findings) != 0 {
		t.Fatalf("checkAPISurface rejected implemented direct_write coverage: %+v", findings)
	}

	b.Metadata.Capabilities.Write = false
	if findings := checkAPISurface(b); len(findings) == 0 {
		t.Fatal("checkAPISurface accepted direct_write coverage while write capability is false")
	}
}
