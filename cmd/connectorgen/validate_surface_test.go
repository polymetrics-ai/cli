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

func TestCheckCLISurfaceEndpointCoverageAllowsSourceBoundPartialRead(t *testing.T) {
	bundle := engine.Bundle{
		Name: "acme",
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method: "GET", Path: "/widgets",
			Operation: &engine.SurfaceOperation{
				Model: "direct_read", Status: "blocked", Risk: "low", BlockedByDefault: true,
				Reason: "Locked source operation acme.widgets.list has no field-complete declaration-owned executable route.",
				Notes:  "source_operation=acme.widgets.list",
			},
		}}},
	}
	command := engine.CLICommand{
		Path: "widgets list", Intent: "direct_read", Availability: "partial",
		Notes:      "Blocked: locked source operation acme.widgets.list has no declaration-owned executable stream, direct-read, binary, or status route.",
		APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets"}},
	}
	if findings := checkCLISurfaceEndpointCoverage(bundle, 0, command, cliSurfaceEndpointStates(bundle.Surface)); len(findings) != 0 {
		t.Fatalf("source-bound partial read must be allowed to reference its blocked endpoint: %+v", findings)
	}
}

func TestCheckCLISurfaceEndpointCoverageAllowsDeclarationBoundDeferredCommand(t *testing.T) {
	bundle := engine.Bundle{
		Name: "acme",
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method: "DELETE", Path: "/widgets/{id}", Operation: &engine.SurfaceOperation{Model: "destructive_action", Status: "blocked", BlockedByDefault: true},
		}}},
	}
	command := engine.CLICommand{
		Path: "widgets delete", Intent: "reverse_etl", Availability: declarationAdmissionStateDeferred,
		APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/widgets/{id}"}},
		Foundation: &engine.CommandFoundation{
			ID: "typed_write_action", Reason: "the endpoint has no typed write action",
			Component: "typed_write_action", Evidence: "write_action_absent",
			Target: engine.CommandFoundationTarget{Method: "DELETE", Path: "/widgets/{id}"},
		},
	}
	if findings := checkCLISurfaceEndpointCoverage(bundle, 0, command, cliSurfaceEndpointStates(bundle.Surface)); len(findings) != 0 {
		t.Fatalf("deferred declaration command must retain its cited blocked endpoint: %+v", findings)
	}
}

// A fixed GraphQL query is a read even though its shared transport is POST.
// The executable source root must therefore not let a connector claim
// capabilities.read=false merely because the REST-specific check only looked
// for GET endpoints.
func TestCheckAPISurface_FixedGraphQLQueryRequiresReadCapability(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: false, Write: false},
		},
		Operations: []engine.OperationSpec{{
			ID:      "acme.graphql.query.viewer",
			Kind:    "graphql_query",
			GraphQL: &engine.GraphQLOperationSpec{Path: "/graphql"},
		}},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "POST",
				Path:      "/graphql",
				CoveredBy: &engine.SurfaceCoverage{Operations: []string{"acme.graphql.query.viewer"}},
			}},
		},
	}

	findings := checkAPISurface(b)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "capabilities.read") {
		t.Fatalf("checkAPISurface GraphQL read-capability findings = %+v, want one read-capability finding", findings)
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
