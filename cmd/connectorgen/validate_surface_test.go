package main

import (
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
