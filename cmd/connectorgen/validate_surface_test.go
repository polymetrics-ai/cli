package main

import (
	"encoding/json"
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

func TestCheckAPISurfaceAndCLISurface_AcceptsClosedWebSocketSessionCoverage(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: false},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "GET",
				Path:      "/live",
				CoveredBy: &engine.SurfaceCoverage{WebSocketSession: "acme.live_scribe"},
			}},
		},
		Operations: []engine.OperationSpec{{
			ID:           "acme.live_scribe",
			Kind:         "websocket_session",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			WebSocket: &engine.WebSocketSessionSpec{
				Method:         "GET",
				Path:           "/live",
				Subprotocol:    "fixture-live",
				MaxInputBytes:  1024,
				MaxOutputBytes: 1024,
				MaxFrameBytes:  128,
				SessionUpdateSchema: json.RawMessage(`{
					"type":"object",
					"additionalProperties":false,
					"required":["type"],
					"properties":{"type":{"type":"string","enum":["session.update"]}}
				}`),
			},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path:         "live scribe",
			Intent:       "websocket_session",
			Availability: "implemented",
			Operation:    "acme.live_scribe",
			APISurface:   []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/live"}},
			OutputPolicy: "json_redacted",
			Flags: []engine.CLIFlag{
				{Name: "session-update", Type: "json_object", MapsTo: "body", Required: true},
				{Name: "audio-file", Type: "string", MapsTo: "input.pcm16_file", Required: true},
			},
		}}},
	}

	if findings := checkAPISurface(b); len(findings) != 0 {
		t.Fatalf("checkAPISurface rejected closed websocket-session coverage: %+v", findings)
	}
	if findings := checkCLISurface(b); len(findings) != 0 {
		t.Fatalf("checkCLISurface rejected closed websocket-session command: %+v", findings)
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
