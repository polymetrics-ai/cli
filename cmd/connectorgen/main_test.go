// Command connectorgen is the wave0 migration-tooling CLI: it validates
// declarative connector definition bundles (defs/), regenerates the two
// deterministic wiring files hookset_gen.go/nativeset_gen.go, and scaffolds
// new bundles.
package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/boundary"
	"polymetrics.ai/internal/connectors/engine"
)

// --- boundary: scans connector definition boundary --------------------------

func TestBoundaryCommand_JSONCleanExitZero(t *testing.T) {
	root := newBoundaryCommandFixture(t, nil)
	var stdout, stderr bytes.Buffer
	exit := run([]string{"boundary", root, "--json"}, &stdout, &stderr)
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	var report boundary.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode boundary report: %v\n%s", err, stdout.String())
	}
	if report.Outcome != boundary.OutcomeClean || len(report.Findings) != 0 {
		t.Fatalf("unexpected clean report: %+v", report)
	}
}

func TestBoundaryCommand_PolicyViolationExitOne(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		"internal/connectors/engine/branch.go": `package engine

func branch(connector string) bool { return connector == "gong" }
`,
	})
	var stdout, stderr bytes.Buffer
	exit := run([]string{"boundary", root, "--json"}, &stdout, &stderr)
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	var report boundary.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode boundary report: %v\n%s", err, stdout.String())
	}
	if report.Outcome != boundary.OutcomePolicyViolations || len(report.Findings) != 1 {
		t.Fatalf("unexpected violation report: %+v", report)
	}
	if got := report.Findings[0]; got.Rule != boundary.RuleConnectorSwitch || got.Connector != "gong" || got.Match != "gong" {
		t.Fatalf("unexpected finding: %+v", got)
	}
}

func TestBoundaryCommand_InvalidInvocationExitTwo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := run([]string{"boundary", "--base="}, &stdout, &stderr)
	if exit != 2 {
		t.Fatalf("exit = %d, want 2", exit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("invalid invocation wrote stdout: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--base requires a value") {
		t.Fatalf("stderr missing invalid invocation reason: %s", stderr.String())
	}
}

func TestBoundaryCommand_InvalidConfigurationExitTwo(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		boundary.DefaultExceptionsPath: `{"exceptions":[{"id":"missing-fields"}]}`,
	})
	var stdout, stderr bytes.Buffer
	exit := run([]string{"boundary", root, "--json"}, &stdout, &stderr)
	if exit != 2 {
		t.Fatalf("exit = %d, want 2\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("invalid config wrote stdout: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "rule is required") {
		t.Fatalf("stderr missing invalid config reason: %s", stderr.String())
	}
}

func TestBoundaryCommand_HumanOutput(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		"internal/connectors/commandrunner/helper.go": `package commandrunner
const policy = "github_date_range"
`,
	})
	var stdout, stderr bytes.Buffer
	exit := run([]string{"boundary", root}, &stdout, &stderr)
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	for _, want := range []string{"connectorgen boundary:", "internal/connectors/commandrunner/helper.go", "provider_policy", "github_date_range"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("human output missing %q:\n%s", want, stdout.String())
		}
	}
}

// --- validate: accepts the golden control bundle -----------------------------

func TestValidate_AcceptsGoodBundle(t *testing.T) {
	report, err := validateDir(os.DirFS("testdata/valid"))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for the good bundle, got %+v", report.Findings)
	}
	if report.ConnectorsChecked != 1 {
		t.Fatalf("ConnectorsChecked = %d, want 1", report.ConnectorsChecked)
	}
}

func TestValidatePath_AcceptsSingleBundleDirectory(t *testing.T) {
	report, err := validatePath(filepath.Join("testdata", "valid", "goodconn"))
	if err != nil {
		t.Fatalf("validatePath: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for the good bundle, got %+v", report.Findings)
	}
	if report.ConnectorsChecked != 1 {
		t.Fatalf("ConnectorsChecked = %d, want 1", report.ConnectorsChecked)
	}
}

func TestValidatePath_AcceptsCurrentWorkingBundleDirectory(t *testing.T) {
	t.Chdir(filepath.Join("testdata", "valid", "goodconn"))
	report, err := validatePath(".")
	if err != nil {
		t.Fatalf("validatePath(.): %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for the good bundle, got %+v", report.Findings)
	}
	if report.ConnectorsChecked != 1 {
		t.Fatalf("ConnectorsChecked = %d, want 1", report.ConnectorsChecked)
	}
}

// TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses is the
// S3 engine mini-wave item 2 regression case (wave1-pilot SUMMARY.md carried
// queue / REVIEW-A.md re-review R1/R3): a `when` clause using the `==`/`in`
// grammar against a REAL, spec-declared key (`auth_type`) must pass
// connectorgen validate cleanly. Before ResolveCheckWhen existed,
// ResolveCheck's bare-namespace.key-only parsing treated the entire
// "auth_type == 'token'" expression as one dotted reference and always
// hard-failed with an "unknown spec key" finding — even though `auth_type`
// IS declared — because no `==`/`in`-shaped reference could ever look like a
// valid two-segment "namespace.key" split. This fixture lives in its own
// parent dir (not testdata/valid, which TestValidate_AcceptsGoodBundle
// asserts contains exactly one connector) so it doesn't disturb that count.
func TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "when-clause-equality-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a spec-known ==/in when clause, got %+v", report.Findings)
	}
}

// TestValidate_FanOutBundlePassesCleanly proves a well-formed fan_out block
// (S4 engine mini-wave item 2) — including a "{{ fanout.id }}" reference in
// stream.Path — passes connectorgen validate with zero findings.
func TestValidate_FanOutBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "fanout-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed fan_out block, got %+v", report.Findings)
	}
}

// TestValidate_KeyedObjectBundlePassesCleanly proves a well-formed
// records.keyed_object/key_field block (S4 engine mini-wave item 3) passes
// connectorgen validate with zero findings.
func TestValidate_KeyedObjectBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "keyed-object-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed records.keyed_object block, got %+v", report.Findings)
	}
}

// TestValidate_OAuth2ExtraParamsBundlePassesCleanly proves a well-formed
// oauth2_client_credentials auth.extra_params block (S4 engine mini-wave item
// 4) passes connectorgen validate with zero findings.
func TestValidate_OAuth2ExtraParamsBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "oauth2-extra-params-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed oauth2_client_credentials extra_params block, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceValidReferencesPassCleanly(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(validCLISurfaceJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid cli_surface.json, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceValidationDeclarationsPassCleanly(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(validCLISurfaceValidationJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid CLI validation declarations, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceRejectsMalformedValidationDeclarations(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "format requires string flag",
			json: strings.Replace(validCLISurfaceValidationJSON(), `"type": "string", "summary": "Start bound.", "maps_to": "query.started_after", "format": "date-time"`, `"type": "integer", "summary": "Start bound.", "maps_to": "query.started_after", "format": "date-time"`, 1),
		},
		{
			name: "constraint requires value type",
			json: strings.Replace(validCLISurfaceValidationJSON(), `, "value_type": "date-time"`, ``, 1),
		},
		{
			name: "fallback must be config",
			json: strings.Replace(validCLISurfaceValidationJSON(), `"left_fallback": "config.default_start"`, `"left_fallback": "query.default_start"`, 1),
		},
		{
			name: "constraint target must be mapped",
			json: strings.Replace(validCLISurfaceValidationJSON(), `"right": "query.started_before"`, `"right": "query.unmapped_before"`, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := validateDir(cliSurfaceBundleFS(tt.json))
			if err != nil {
				t.Fatalf("validateDir: %v", err)
			}
			assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
		})
	}
}

func TestValidate_CLISurfaceUnknownStreamIsHardFinding(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `"stream": "widgets"`, `"stream": "missing_widgets"`)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceUnknownTarget)
}

func TestValidate_CLISurfaceImplementedETLRequiresStream(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `"stream": "widgets",`, "")))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceMissingMapping)
}

func TestValidate_CLISurfaceSecretLookingExampleIsHardFinding(t *testing.T) {
	token := "gh" + "p_" + "1234567890abcdef1234567890abcdef1234"
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `pm cli-surface widget list --json`, `pm cli-surface auth --token `+token)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSecretLiteral)
}

func TestValidate_CLISurfaceAPIRefCannotUseExcludedEndpoint(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `{ "method": "GET", "path": "/widgets" }`, `{ "method": "GET", "path": "/widgets/export" }`, 1)
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/widgets/export", "excluded": { "category": "out_of_scope", "reason": "not exposed" } }
		]
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceAPIRefMustMatchStreamOrWrite(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `{ "method": "GET", "path": "/widgets" }`, `{ "method": "GET", "path": "/widget-writes" }`, 1)
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "GET", "path": "/widget-writes", "covered_by": { "write": "create_widget" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } }
		]
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceAPIRefFailsWhenSurfaceHasZeroEndpoints(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": []
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceUnknownTarget)
}

func TestValidate_CLISurfaceReverseETLRequiresRiskAndApproval(t *testing.T) {
	cliSurface := strings.ReplaceAll(validCLISurfaceJSON(), `
				"risk": "creates a widget",
				"approval": "reverse ETL writes require plan, preview, approval, execute",
`, "")
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceReverseETLRequiresRequiredRecordFlagMappings(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `"maps_to": "record.name"`, `"maps_to": "query.name"`, 1)
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceMissingMapping)
}

func TestValidate_CLISurfaceReverseETLAcceptsSchemaBoundNestedArrayMappings(t *testing.T) {
	fsys := cliSurfaceBundleFS(`{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "patient create",
				"summary": "Create patient",
				"intent": "reverse_etl",
				"availability": "implemented",
				"write": "create_patient",
				"api_surface": [{ "method": "POST", "path": "/widgets" }],
				"flags": [
					{ "name": "identifier", "type": "string", "maps_to": "record.identifiers.0.identifier" },
					{ "name": "identifier-type", "type": "string", "maps_to": "record.identifiers.0.identifierType" },
					{ "name": "identifier-location", "type": "string", "maps_to": "record.identifiers.0.location" },
					{ "name": "identifier-preferred", "type": "boolean", "maps_to": "record.identifiers.0.preferred" },
					{ "name": "given-name", "type": "string", "maps_to": "record.person.names.0.givenName" },
					{ "name": "family-name", "type": "string", "maps_to": "record.person.names.0.familyName" },
					{ "name": "gender", "type": "enum", "values": ["M", "F", "O"], "maps_to": "record.person.gender" },
					{ "name": "birthdate", "type": "string", "maps_to": "record.person.birthdate" }
				],
				"risk": "creates a patient",
				"approval": "reverse ETL writes require plan, preview, approval, execute"
			}
		]
	}`)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_patient" } }
		]
	}`)}
	fsys["cli-surface/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [{
			"name": "create_patient",
			"kind": "create",
			"method": "POST",
			"path": "/widgets",
			"record_schema": {
				"type": "object",
				"required": ["identifiers", "person"],
				"properties": {
					"identifiers": { "type": "array", "items": { "type": "object", "required": ["identifier", "identifierType", "location", "preferred"], "properties": { "identifier": { "type": "string" }, "identifierType": { "type": "string" }, "location": { "type": "string" }, "preferred": { "type": "boolean" } }, "additionalProperties": false } },
					"person": { "type": "object", "required": ["names", "gender", "birthdate"], "properties": { "names": { "type": "array", "items": { "type": "object", "required": ["givenName", "familyName"], "properties": { "givenName": { "type": "string" }, "familyName": { "type": "string" } }, "additionalProperties": false } }, "gender": { "type": "string" }, "birthdate": { "type": "string" } }, "additionalProperties": false }
				},
				"additionalProperties": false
			},
			"risk": "creates a patient"
		}]
	}`)}
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for schema-bound nested mappings, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceReverseETLRejectsStructuredStringAndMissingNestedMapping(t *testing.T) {
	fsys := cliSurfaceBundleFS(`{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "patient create",
				"summary": "Create patient",
				"intent": "reverse_etl",
				"availability": "implemented",
				"write": "create_patient",
				"api_surface": [{ "method": "POST", "path": "/widgets" }],
				"flags": [
					{ "name": "identifiers", "type": "string", "maps_to": "record.identifiers" },
					{ "name": "given-name", "type": "string", "maps_to": "record.person.names.0.givenName" }
				],
				"risk": "creates a patient",
				"approval": "reverse ETL writes require plan, preview, approval, execute"
			}
		]
	}`)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_patient" } }
		]
	}`)}
	fsys["cli-surface/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [{
			"name": "create_patient",
			"kind": "create",
			"method": "POST",
			"path": "/widgets",
			"record_schema": {
				"type": "object",
				"required": ["identifiers", "person"],
				"properties": {
					"identifiers": { "type": "array", "items": { "type": "object", "required": ["identifier"], "properties": { "identifier": { "type": "string" } }, "additionalProperties": false } },
					"person": { "type": "object", "required": ["names"], "properties": { "names": { "type": "array", "items": { "type": "object", "required": ["givenName", "familyName"], "properties": { "givenName": { "type": "string" }, "familyName": { "type": "string" } }, "additionalProperties": false } } }, "additionalProperties": false }
				},
				"additionalProperties": false
			},
			"risk": "creates a patient"
		}]
	}`)}
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceMissingMapping)
}

func TestValidate_CLISurfaceReverseETLStructuredJSONRequiresDeclaredTopLevelContainer(t *testing.T) {
	const originalFlag = `{ "name": "name", "type": "string", "summary": "Widget name.", "maps_to": "record.name" }`
	newBundle := func(target, payloadSchema string) fstest.MapFS {
		cliSurface := strings.Replace(validCLISurfaceJSON(), originalFlag, `{"name": "payload", "type": "json", "summary": "Structured widget payload.", "maps_to": "`+target+`", "required": true}`, 1)
		fsys := cliSurfaceBundleFS(cliSurface)
		fsys["cli-surface/writes.json"] = &fstest.MapFile{Data: []byte(`{
			"actions": [{
				"name": "create_widget",
				"kind": "create",
				"method": "POST",
				"path": "/widgets",
				"record_schema": {
					"type": "object",
					"required": ["payload"],
					"properties": {"payload": ` + payloadSchema + `},
					"additionalProperties": false
				},
				"risk": "creates a widget"
			}]
		}`)}
		return fsys
	}

	for _, tc := range []struct {
		name, target, payloadSchema string
		wantFinding                 bool
	}{
		{
			name:          "closed object field is accepted",
			target:        "record.payload",
			payloadSchema: `{"type":"object","required":["kind"],"properties":{"kind":{"type":"string"}},"additionalProperties":false}`,
		},
		{
			name:          "scalar field is rejected",
			target:        "record.payload",
			payloadSchema: `{"type":"string"}`,
			wantFinding:   true,
		},
		{
			name:          "nested field is rejected",
			target:        "record.payload.kind",
			payloadSchema: `{"type":"object","properties":{"kind":{"type":"string"}},"additionalProperties":false}`,
			wantFinding:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, err := validateDir(newBundle(tc.target, tc.payloadSchema))
			if err != nil {
				t.Fatalf("validateDir: %v", err)
			}
			if tc.wantFinding {
				assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
				return
			}
			if len(report.Findings) != 0 {
				t.Fatalf("valid declared structured JSON flag has findings: %+v", report.Findings)
			}
		})
	}

	findings := checkCLISurfaceStructuredJSONFlags(engine.Bundle{Name: "cli-surface"}, 0, engine.CLICommand{
		Path: "widget raw", Intent: "direct_read", Availability: "implemented", Operation: "widgets.raw",
		Flags: []engine.CLIFlag{{Name: "payload", Type: "json", MapsTo: "body.payload"}},
	})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "allowed only on a declared reverse-ETL record field") {
		t.Fatalf("non-reverse structured JSON findings = %+v, want closed placement rejection", findings)
	}
}

func TestValidate_CLISurfaceFixedGraphQLCommandRequiresDeclaredTopLevelJSONVariable(t *testing.T) {
	bundle := engine.Bundle{
		Name: "cli-surface",
		Operations: []engine.OperationSpec{{
			ID:           "cli-surface.widgets.query",
			Kind:         "graphql_query",
			Summary:      "Read one widget",
			Risk:         "low",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			GraphQL: &engine.GraphQLOperationSpec{
				OperationName:   "QueryWidget",
				Document:        "query QueryWidget($input: WidgetInput!) { widget(input: $input) { __typename } }",
				Path:            "/graphql",
				MaxBytes:        1024,
				VariablesSchema: json.RawMessage("{\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"input\"],\"properties\":{\"input\":{\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"id\"],\"properties\":{\"id\":{\"type\":\"string\"}}}}}"),
			},
		}},
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method: "POST",
			Path:   "/graphql",
			CoveredBy: &engine.SurfaceCoverage{
				Operations: []string{"cli-surface.widgets.query"},
			},
		}}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path:         "graphql query widget",
			Summary:      "Read one widget",
			Intent:       "direct_read",
			Availability: "implemented",
			Operation:    "cli-surface.widgets.query",
			OutputPolicy: "json_redacted",
			Flags: []engine.CLIFlag{{
				Name: "input", Type: "json", Required: true, MapsTo: "body.input",
			}},
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}},
		}}},
	}

	if findings := checkCLISurface(bundle); len(findings) != 0 {
		t.Fatalf("fixed GraphQL command findings = %+v, want none", findings)
	}

	bundle.CLISurface.Commands[0].Flags[0].MapsTo = "body.input.id"
	findings := checkCLISurface(bundle)
	if len(findings) == 0 || !strings.Contains(findings[0].Message, "top-level") {
		t.Fatalf("nested fixed GraphQL JSON flag findings = %+v, want top-level rejection", findings)
	}
}

func TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract(t *testing.T) {
	op := engine.OperationSpec{
		ID:            "cli-surface.organization.migration",
		Kind:          "graphql_mutation",
		MutationClass: "secret",
		SensitivePolicy: &engine.SensitivePolicySpec{
			InputMode:    "env",
			RedactFields: []string{"body.input"},
			ApprovalMode: "typed_confirmation",
		},
	}
	command := engine.CLICommand{
		Path:         "graphql mutation start-organization-migration",
		Intent:       "direct_write",
		Availability: "implemented",
		Operation:    op.ID,
		Flags: []engine.CLIFlag{{
			Name: "input", Type: "json", Required: true, MapsTo: "body.input", EnvOnly: true,
		}},
	}
	operations := map[string]engine.OperationSpec{op.ID: op}
	if findings := checkCLISurfaceEnvOnlyFlags(engine.Bundle{Name: "cli-surface"}, 0, command, operations); len(findings) != 0 {
		t.Fatalf("declared secret GraphQL env_only contract findings = %+v, want none", findings)
	}

	command.Flags[0].EnvOnly = false
	if findings := checkCLISurfaceEnvOnlyFlags(engine.Bundle{Name: "cli-surface"}, 0, command, operations); len(findings) != 0 {
		t.Fatalf("ordinary typed input should not require env_only findings = %+v", findings)
	}

	command.Flags[0].EnvOnly = true
	op.SensitivePolicy.InputMode = "inline"
	operations[op.ID] = op
	findings := checkCLISurfaceEnvOnlyFlags(engine.Bundle{Name: "cli-surface"}, 0, command, operations)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "secret GraphQL mutation") {
		t.Fatalf("env_only without environment redaction contract findings = %+v, want one closed-surface rejection", findings)
	}
}

func TestValidate_CLISurfaceImplementedRawAPIIsBlocked(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `"intent": "etl"`, `"intent": "raw_api"`, 1)
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedDirectWritePasses(t *testing.T) {
	report, err := validateDir(directWriteCLISurfaceBundleFS())
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid direct_write cli surface, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceImplementedMultipartDirectWritePasses(t *testing.T) {
	fsys := directWriteCLISurfaceBundleFS()
	cli := string(fsys["cli-surface/cli_surface.json"].Data)
	originalCLI := cli
	cli = strings.Replace(cli, `
					{ "name": "id", "type": "string", "maps_to": "path.id" }
				`, `
					{ "name": "id", "type": "string", "maps_to": "path.id" },
					{ "name": "media-file-path", "type": "string", "maps_to": "body.media_file_path", "required": true }
				`, 1)
	if cli == originalCLI {
		t.Fatal("add multipart body flag to cli surface")
	}
	fsys["cli-surface/cli_surface.json"] = &fstest.MapFile{Data: []byte(cli)}

	operations := string(fsys["cli-surface/operations.json"].Data)
	originalOperations := operations
	operations = strings.Replace(operations, `"max_bytes": 1024`, `"content_type": "multipart/form-data",
					"max_bytes": 1024,
					"body_schema": {
						"type": "object",
						"additionalProperties": false,
						"required": ["media_file_path"],
						"properties": {"media_file_path": {"type": "string"}}
					},
					"multipart": {
						"max_bytes": 1024,
						"parts": [{
							"name": "media",
							"type": "file",
							"field": "media_file_path",
							"required": true,
							"max_bytes": 1024,
							"content_type": "text/plain",
							"allowed_media_types": ["text/plain"]
						}]
					}`, 1)
	if operations == originalOperations {
		t.Fatal("add typed multipart contract to operation")
	}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for typed multipart direct_write cli surface, got %+v", report.Findings)
	}
}

func TestSupportedDirectWriteContentTypeRequiresTypedMultipart(t *testing.T) {
	tests := []struct {
		name string
		rest *engine.RESTOperationSpec
		want bool
	}{
		{
			name: "typed literal multipart",
			rest: &engine.RESTOperationSpec{ContentType: "multipart/form-data", Multipart: &engine.MultipartSpec{}},
			want: true,
		},
		{
			name: "multipart annotation without contract",
			rest: &engine.RESTOperationSpec{ContentType: "multipart/form-data"},
		},
		{
			name: "multipart boundary is rejected",
			rest: &engine.RESTOperationSpec{ContentType: "multipart/form-data; boundary=caller-controlled", Multipart: &engine.MultipartSpec{}},
		},
		{
			name: "multipart whitespace is rejected",
			rest: &engine.RESTOperationSpec{ContentType: " multipart/form-data ", Multipart: &engine.MultipartSpec{}},
		},
		{
			name: "existing JSON content type remains supported",
			rest: &engine.RESTOperationSpec{ContentType: "application/json"},
			want: true,
		},
		{
			name: "closed SCIM JSON content type is supported",
			rest: &engine.RESTOperationSpec{ContentType: "application/scim+json"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := supportedDirectWriteContentType(tt.rest); got != tt.want {
				t.Fatalf("supportedDirectWriteContentType() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestValidate_CLISurfaceImplementedDirectReadWithOutputPolicyPasses(t *testing.T) {
	report, err := validateDir(directReadCLISurfaceBundleFS(validDirectReadCLISurfaceJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid direct_read cli surface, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceRepositoryOutputPolicyRequiresPathVariable(t *testing.T) {
	cliSurface := strings.ReplaceAll(validDirectReadCLISurfaceJSON(), "/widgets/{path}", "/widgets/{file_path}")
	cliSurface = strings.Replace(cliSurface, `"maps_to": "path.path"`, `"maps_to": "path.file_path"`, 1)
	apiSurface := strings.ReplaceAll(validDirectReadAPISurface(), "/widgets/{path}", "/widgets/{file_path}")
	report, err := validateDir(directReadCLISurfaceBundleFSWithAPI(cliSurface, apiSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedDirectReadRejectsBlockedOperationLedgerEndpoint(t *testing.T) {
	report, err := validateDir(directReadCLISurfaceBundleFSWithAPI(validDirectReadCLISurfaceJSON(), validOperationLedgerAPISurface()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedDirectReadRequiresOutputPolicy(t *testing.T) {
	cliSurface := strings.Replace(validDirectReadCLISurfaceJSON(), `
				"output_policy": "repository_contents_file_metadata",
`, "", 1)
	report, err := validateDir(directReadCLISurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedDirectReadRequiresOneEndpoint(t *testing.T) {
	cliSurface := strings.Replace(
		validDirectReadCLISurfaceJSON(),
		`{ "method": "GET", "path": "/widgets/{path}" }`,
		`{ "method": "GET", "path": "/widgets/{path}" }, { "method": "GET", "path": "/widgets" }`,
		1,
	)
	report, err := validateDir(directReadCLISurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceMissingMapping)
}

func TestValidate_CLISurfaceImplementedDirectReadRequiresGETRelativeEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		apiSurface string
		ref        string
	}{
		{
			name:       "post",
			apiSurface: strings.Replace(validDirectReadAPISurface(), `"method": "GET",`+"\n"+`				"path": "/widgets/{path}"`, `"method": "POST",`+"\n"+`				"path": "/widgets/{path}"`, 1),
			ref:        `{ "method": "POST", "path": "/widgets/{path}" }`,
		},
		{
			name:       "absolute",
			apiSurface: strings.Replace(validDirectReadAPISurface(), `"path": "/widgets/{path}"`, `"path": "https://evil.example.test/widgets/{path}"`, 1),
			ref:        `{ "method": "GET", "path": "https://evil.example.test/widgets/{path}" }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliSurface := strings.Replace(validDirectReadCLISurfaceJSON(), `{ "method": "GET", "path": "/widgets/{path}" }`, tt.ref, 1)
			report, err := validateDir(directReadCLISurfaceBundleFSWithAPI(cliSurface, tt.apiSurface))
			if err != nil {
				t.Fatalf("validateDir: %v", err)
			}
			assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
		})
	}
}

func TestValidate_CLISurfaceRejectsCommandWithStreamAndWrite(t *testing.T) {
	cliSurface := strings.Replace(
		validCLISurfaceJSON(),
		`"stream": "widgets",`,
		`"stream": "widgets",
				"write": "create_widget",`,
		1,
	)
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceOperationReferencePasses(t *testing.T) {
	report, err := validateDir(operationCLISurfaceBundleFS(validOperationCLISurfaceJSON(), validOperationsJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid operation-backed cli surface, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceOperationDirectReadRequiresBodyMappings(t *testing.T) {
	cliSurface := `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget preview",
				"summary": "Preview widget metadata",
				"intent": "direct_read",
				"availability": "implemented",
				"operation": "cli-surface.widgets.preview",
				"api_surface": [
					{ "method": "POST", "path": "/widgets:preview" }
				],
				"output_policy": "json_redacted",
				"examples": ["pm cli-surface widget preview --json"]
			}
		]
	}`
	operations := `{
		"operations": [
			{
				"id": "cli-surface.widgets.preview",
				"kind": "rest_read",
				"summary": "Preview widget metadata",
				"risk": "low",
				"approval": "none",
				"output_policy": "json_redacted",
				"rest": {
					"method": "POST",
					"path": "/widgets:preview",
					"content_type": "application/json",
					"max_bytes": 1024,
					"body_schema": {
						"type": "object",
						"additionalProperties": false,
						"required": ["payload"],
						"properties": {
							"payload": { "type": "string" }
						}
					},
					"body": {}
				}
			}
		]
	}`
	apiSurface := `{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "POST", "path": "/widgets:preview", "covered_by": { "direct_read": "widget preview" } }
		]
	}`
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceOperationDirectReadRequiresRequiredBodyFlags(t *testing.T) {
	cliSurface := `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget preview",
				"summary": "Preview widget metadata",
				"intent": "direct_read",
				"availability": "implemented",
				"operation": "cli-surface.widgets.preview",
				"api_surface": [
					{ "method": "POST", "path": "/widgets:preview" }
				],
				"output_policy": "json_redacted",
				"flags": [
					{ "name": "payload", "type": "string", "maps_to": "body.payload" }
				],
				"examples": ["pm cli-surface widget preview --payload fixture --json"]
			}
		]
	}`
	operations := `{
		"operations": [
			{
				"id": "cli-surface.widgets.preview",
				"kind": "rest_read",
				"summary": "Preview widget metadata",
				"risk": "low",
				"approval": "none",
				"output_policy": "json_redacted",
				"rest": {
					"method": "POST",
					"path": "/widgets:preview",
					"content_type": "application/json",
					"max_bytes": 1024,
					"body_schema": {
						"type": "object",
						"additionalProperties": false,
						"required": ["payload"],
						"properties": {
							"payload": { "type": "string" }
						}
					},
					"body": {}
				}
			}
		]
	}`
	apiSurface := `{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "POST", "path": "/widgets:preview", "covered_by": { "direct_read": "widget preview" } }
		]
	}`
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)

	fixed := strings.Replace(cliSurface, `"maps_to": "body.payload"`, `"maps_to": "body.payload", "required": true`, 1)
	fsys = cliSurfaceBundleFS(fixed)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err = validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir fixed: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for required body flag, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceOperationDirectReadPlainTextBodyRequiresOneRequiredStringFlag(t *testing.T) {
	cliSurface := `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "markdown raw",
				"summary": "Render raw Markdown",
				"intent": "direct_read",
				"availability": "implemented",
				"operation": "cli-surface.markdown.raw",
				"api_surface": [
					{ "method": "POST", "path": "/markdown/raw" }
				],
				"output_policy": "text",
				"flags": [
					{ "name": "text", "type": "string", "required": true, "maps_to": "body" }
				],
				"examples": ["pm cli-surface markdown raw --text '# heading' --json"]
			}
		]
	}`
	operations := `{
		"operations": [
			{
				"id": "cli-surface.markdown.raw",
				"kind": "rest_read",
				"summary": "Render raw Markdown",
				"risk": "low",
				"approval": "none",
				"output_policy": "text",
				"rest": {
					"method": "POST",
					"path": "/markdown/raw",
					"content_type": "text/plain",
					"max_bytes": 1024,
					"body_schema": { "type": "string" }
				}
			}
		]
	}`
	apiSurface := `{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "POST", "path": "/markdown/raw", "covered_by": { "direct_read": "markdown raw" } }
		]
	}`

	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for declared text/plain raw body, got %+v", report.Findings)
	}

	invalid := strings.Replace(cliSurface, `"required": true, "maps_to": "body"`, `"maps_to": "body"`, 1)
	fsys = cliSurfaceBundleFS(invalid)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err = validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir invalid: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)

	mixed := strings.Replace(cliSurface, `"flags": [`, `"flags": [
					{ "name": "context", "type": "string", "maps_to": "body.context" },`, 1)
	fsys = cliSurfaceBundleFS(mixed)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	report, err = validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir mixed: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceOperationRepositoryOutputPolicyRequiresPathVariable(t *testing.T) {
	cliSurface := strings.Replace(validOperationCLISurfaceJSON(), `"output_policy": "json_redacted"`, `"output_policy": "repository_contents_file_metadata"`, 1)
	operations := strings.Replace(validOperationsJSON(), `"output_policy": "json_redacted"`, `"output_policy": "repository_contents_file_metadata"`, 1)
	report, err := validateDir(operationCLISurfaceBundleFS(cliSurface, operations))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceUnknownOperationIsHardFinding(t *testing.T) {
	cliSurface := strings.ReplaceAll(validOperationCLISurfaceJSON(), `"operation": "cli-surface.widgets.get"`, `"operation": "cli-surface.widgets.missing"`)
	report, err := validateDir(operationCLISurfaceBundleFS(cliSurface, validOperationsJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceUnknownTarget)
}

func TestValidate_CLISurfaceRejectsCommandWithStreamAndOperation(t *testing.T) {
	cliSurface := strings.ReplaceAll(
		validOperationCLISurfaceJSON(),
		`"operation": "cli-surface.widgets.get"`,
		`"stream": "widgets", "operation": "cli-surface.widgets.get"`,
	)
	report, err := validateDir(operationCLISurfaceBundleFS(cliSurface, validOperationsJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_InvalidOperationsJSONFindingNamesOperationsFile(t *testing.T) {
	report, err := validateDir(operationCLISurfaceBundleFS(validOperationCLISurfaceJSON(), strings.Replace(
		validOperationsJSON(),
		`"risk": "low"`,
		`"risk": "unbounded"`,
		1,
	)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.Connector == "cli-surface" && finding.Rule == ruleMetaSchema {
			if finding.File != "operations.json" {
				t.Fatalf("meta_schema finding file = %q, want operations.json; finding=%+v", finding.File, finding)
			}
			return
		}
	}
	t.Fatalf("no meta_schema finding for invalid operations.json; findings=%+v", report.Findings)
}

func TestValidate_OperationsSecretLookingLiteralIsHardFinding(t *testing.T) {
	token := "gh" + "p_" + "1234567890abcdef1234567890abcdef1234"
	operations := strings.Replace(validOperationsJSON(), "Read widget metadata", "Read "+token, 1)
	report, err := validateDir(operationCLISurfaceBundleFS(validOperationCLISurfaceJSON(), operations))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.Connector == "cli-surface" && finding.Rule == ruleSecretLiteral {
			if finding.File != "operations.json" {
				t.Fatalf("secret finding file = %q, want operations.json; finding=%+v", finding.File, finding)
			}
			if strings.Contains(finding.Message, token) {
				t.Fatalf("secret finding message leaked token: %q", finding.Message)
			}
			return
		}
	}
	t.Fatalf("no secret literal finding for operations.json; findings=%+v", report.Findings)
}

func TestValidate_APISurfaceOperationLedgerValidRowsPassCleanly(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(validOperationLedgerAPISurface()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid operation ledger, got %+v", report.Findings)
	}
}

func TestValidate_APISurfaceV2ProvenanceUsesSharedValidation(t *testing.T) {
	tests := []struct {
		name          string
		apiSurface    string
		wantRule      string
		wantMessage   string
		wantCleanPass bool
	}{
		{
			name:          "complete_v2",
			apiSurface:    validV2ProvenanceAPISurface(),
			wantCleanPass: true,
		},
		{
			name:        "missing_endpoint_citation",
			apiSurface:  strings.Replace(validV2ProvenanceAPISurface(), `"source_url": "https://docs.acme.test/api/widgets#sensitive"`, `"source_url": ""`, 1),
			wantRule:    ruleSurfaceProvenance,
			wantMessage: "provenance.source_url is required",
		},
		{
			name:        "unknown_artifact",
			apiSurface:  strings.Replace(validV2ProvenanceAPISurface(), `"artifact": "acme-openapi-2026-08-06"`, `"artifact": "unknown-artifact"`, 1),
			wantRule:    ruleSurfaceProvenance,
			wantMessage: `resolves to 0 artifacts`,
		},
		{
			name: "duplicate_artifact_id",
			apiSurface: strings.Replace(validV2ProvenanceAPISurface(), `"artifacts": [`, `"artifacts": [{
				"id": "acme-openapi-2026-08-06",
				"url": "https://docs.acme.test/openapi-copy.yaml",
				"retrieved_at": "2026-08-06"
			},`, 1),
			wantRule:    ruleSurfaceProvenance,
			wantMessage: `resolves to 2 artifacts`,
		},
		{
			name:        "non_https_endpoint_citation",
			apiSurface:  strings.Replace(validV2ProvenanceAPISurface(), `"source_url": "https://docs.acme.test/api/widgets#sensitive"`, `"source_url": "http://docs.acme.test/api/widgets#sensitive"`, 1),
			wantRule:    ruleSurfaceProvenance,
			wantMessage: "provenance.source_url must be an absolute HTTPS URL",
		},
		{
			name:        "invalid_artifact_date",
			apiSurface:  strings.Replace(validV2ProvenanceAPISurface(), `"retrieved_at": "2026-08-06"`, `"retrieved_at": "2026-08-06T12:00:00Z"`, 1),
			wantRule:    ruleSurfaceProvenance,
			wantMessage: "retrieved_at must be an ISO-8601 full-date",
		},
		{
			name:          "v1_is_legacy_compatible",
			apiSurface:    validOperationLedgerAPISurface(),
			wantCleanPass: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report, err := validateDir(operationLedgerBundleFS(tc.apiSurface))
			if err != nil {
				t.Fatalf("validateDir: %v", err)
			}
			if tc.wantCleanPass {
				if len(report.Findings) != 0 {
					t.Fatalf("findings = %+v, want none", report.Findings)
				}
				return
			}
			for _, finding := range report.Findings {
				if finding.Connector == "cli-surface" && finding.File == "api_surface.json" && finding.Rule == tc.wantRule && strings.Contains(finding.Message, tc.wantMessage) {
					return
				}
			}
			t.Fatalf("findings = %+v, want %s containing %q", report.Findings, tc.wantRule, tc.wantMessage)
		})
	}
}

func TestValidate_APISurfaceOperationLedgerRejectsLegacyExclusion(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/widgets/export", "excluded": { "category": "out_of_scope", "reason": "legacy exclusion" } }
		]
	}`))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSurfaceOperation)
}

func TestValidate_APISurfaceOperationLedgerRejectsDualClassification(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{
				"method": "GET",
				"path": "/widgets/{id}",
				"covered_by": { "stream": "widgets" },
				"operation": {
					"model": "direct_read",
					"status": "blocked",
					"risk": "low",
					"blocked_by_default": true,
					"reason": "dual classified row"
				}
			}
		]
	}`))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSurfaceCoverage)
}

func TestValidate_APISurfaceOperationLedgerRejectsUnblockedOperationInSchema(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(strings.Replace(
		validOperationLedgerAPISurface(),
		`"blocked_by_default": true`,
		`"blocked_by_default": false`,
		1,
	)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleMetaSchema)
}

func TestValidate_APISurfaceOperationLedgerRequiresReason(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(strings.Replace(
		validOperationLedgerAPISurface(),
		`"reason": "point lookup candidate, not yet modeled as a stream"`,
		`"reason": ""`,
		1,
	)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSurfaceOperation)
}

func TestValidate_APISurfaceOperationLedgerRequiresDuplicateTarget(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(strings.Replace(
		validOperationLedgerAPISurface(),
		`"model": "direct_read"`,
		`"model": "duplicate"`,
		1,
	)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSurfaceOperation)
}

func TestValidate_APISurfaceOperationLedgerRequiresSourceOrNotesForSensitiveRows(t *testing.T) {
	report, err := validateDir(operationLedgerBundleFS(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{
				"method": "POST",
				"path": "/org/widgets",
				"operation": {
					"model": "admin_reverse_etl",
					"status": "blocked",
					"risk": "high",
					"blocked_by_default": true,
					"reason": "requires organization administration scope"
				}
			}
		]
	}`))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSurfaceOperation)
}

// TestValidate_EmptyTreeIsFine mirrors the loader contract: an empty defs/
// tree (no bundle directories) passes with a zero connector count, so wave0's
// bundle-less internal/connectors/defs/ tree does not fail CI.
func TestValidate_EmptyTreeIsFine(t *testing.T) {
	dir := t.TempDir()
	report, err := validateDir(os.DirFS(dir))
	if err != nil {
		t.Fatalf("validateDir on empty tree: %v", err)
	}
	if report.ConnectorsChecked != 0 {
		t.Fatalf("ConnectorsChecked = %d, want 0", report.ConnectorsChecked)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings on an empty tree, got %+v", report.Findings)
	}
}

func TestValidate_RejectsMalformedCertificationMetadata(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/certification.json"] = &fstest.MapFile{Data: []byte(`{"schema_version":1,"surprise":true}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	var found *Finding
	for i := range report.Findings {
		if report.Findings[i].Connector == "cli-surface" && report.Findings[i].File == "certification.json" {
			found = &report.Findings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected certification.json finding, got %+v", report.Findings)
	}
	if found.Rule != ruleMetaSchema || !strings.Contains(found.Message, "surprise") {
		t.Fatalf("finding = %+v, want meta_schema surprise", *found)
	}
}

func TestValidate_RejectsIncrementalCursorSchemaMismatch(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/widgets" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"records": { "path": "data" },
				"incremental": { "cursor_field": "updated_at" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}
	fsys["cli-surface/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "created_at",
		"properties": {
			"id": { "type": "integer" },
			"created_at": { "type": "string" },
			"updated_at": { "type": "string" }
		}
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleIncrementalCursorMismatch)
}

func TestValidate_AcceptsIncrementalCursorWithoutSchemaCursor(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/widgets" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"records": { "path": "data" },
				"incremental": { "cursor_field": "updated_at" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}
	fsys["cli-surface/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"x-primary-key": ["id"],
		"properties": {
			"id": { "type": "integer" },
			"updated_at": { "type": "string" }
		}
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("validateDir findings = %+v, want none", report.Findings)
	}
}

// --- validate: seeded-invalid corpus (>=10 seeded, >=8 distinct classes) ----

func TestValidate_RejectsSeededInvalidBundles(t *testing.T) {
	cases := []struct {
		dir      string // testdata/invalid/<dir>
		wantRule string
	}{
		{"missing-metadata-file", ruleMissingFile},
		{"bad-spec-schema", ruleMetaSchema},
		{"unresolvable-interpolation", ruleInterpolationUnresolved},
		{"missing-schema-ref", ruleSchemaRefMissing},
		{"pk-not-in-schema", rulePrimaryKeyMissing},
		{"cursor-not-in-schema", ruleCursorFieldMissing},
		{"write-path-fields-not-in-schema", ruleWritePathFields},
		{"surface-both-covered-and-excluded", ruleSurfaceCoverage},
		{"surface-missing-stream", ruleSurfaceIncomplete},
		{"Source-GitHub", ruleNameRegex},
		{"secret-literal-in-fixture", ruleSecretLiteral},
		{"docs-missing-heading", ruleDocsHeading},
		{"surface-unknown-category", ruleMetaSchema},
		{"write-false-with-mutation-endpoint", ruleSurfaceFailFirstRun},
		{"auth-field-unknown-spec-key", ruleInterpolationUnresolved},
		{"unknown-filter-in-template", ruleInterpolationUnresolved},
		{"when-clause-equality-unknown-spec-key", ruleInterpolationUnresolved},
		{"skip-marker-missing-reason", ruleConformanceSkipReason},
		{"skip-marker-missing-reason-bundle", ruleConformanceSkipReason},
		{"default-type-mismatch", ruleDefaultTypeMismatch},
		{"unknown-base-key", ruleMetaSchema},
		// checkquery-ledger.md: base.check.query is now a real, engine-level
		// field (RequestSpec.Query) rather than an unknown key, so its
		// templates must be statically validated exactly like stream.Query's
		// — a check.query entry templating an undeclared spec key is a
		// ruleInterpolationUnresolved finding, the same rule
		// auth-field-unknown-spec-key already exercises for base.auth.
		{"check-query-unknown-spec-key", ruleInterpolationUnresolved},
		// S4 engine mini-wave item 2: fan_out.ids_from.request.path gets the
		// same static ResolveCheck treatment as an ordinary stream.Path.
		{"fanout-request-path-unknown-spec-key", ruleInterpolationUnresolved},
		// ResolveCheck scans {{ }} only, so a bare {projectId} in a stream
		// path used to validate clean and then reach the wire verbatim (12
		// help-scout streams shipped that way). A declarative read path binds
		// values only through interpolation, so an unbound single-brace
		// placeholder is a finding. writes.json is deliberately exempt:
		// WriteAction.path_fields binds {owner}/{repo}-style placeholders.
		{"stream-path-literal-placeholder", ruleInterpolationUnresolved},
		// S4 engine mini-wave item 4: oauth2_client_credentials auth.extra_params
		// values get the same static ResolveCheck treatment as token_url/
		// client_id/client_secret/scopes.
		{"oauth2-extra-params-unknown-spec-key", ruleInterpolationUnresolved},
	}

	seenRules := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			// validateDir mirrors engine.LoadAll's contract: its fsys root is
			// the PARENT of bundle directories, not a bundle directory
			// itself (a bundle's own subdirectories like schemas/ and
			// fixtures/ must never be mistaken for sibling bundles). So each
			// seeded case is validated in isolation by rooting fsys one
			// level up and filtering findings down to that one connector.
			fsys := singleBundleFS(t, "testdata/invalid", tc.dir)
			report, err := validateDir(fsys)
			if err != nil {
				// A hard structural error (e.g. missing metadata.json) surfaces
				// as a returned error from the loader rather than a Finding;
				// validate must still translate it into a named finding via
				// the caller. Exercise that path explicitly here too.
				t.Fatalf("validateDir(%s) returned a bare error instead of findings: %v", tc.dir, err)
			}
			var relevant []Finding
			for _, f := range report.Findings {
				if f.Connector == tc.dir {
					relevant = append(relevant, f)
				}
			}
			if len(relevant) == 0 {
				t.Fatalf("validateDir(%s): expected at least one finding for connector %q, got none (all findings: %+v)", tc.dir, tc.dir, report.Findings)
			}
			var found *Finding
			for i := range relevant {
				if relevant[i].Rule == tc.wantRule {
					found = &relevant[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("validateDir(%s): no finding with rule %q; got %+v", tc.dir, tc.wantRule, relevant)
			}
			if found.Connector == "" {
				t.Fatalf("finding %+v missing connector name", found)
			}
			if found.File == "" {
				t.Fatalf("finding %+v missing file name", found)
			}
			if found.Message == "" {
				t.Fatalf("finding %+v missing message", found)
			}
		})
		seenRules[tc.wantRule] = true
	}

	if len(cases) < 10 {
		t.Fatalf("seeded corpus has %d cases, want >= 10", len(cases))
	}
	if len(seenRules) < 8 {
		t.Fatalf("seeded corpus covers %d distinct rules, want >= 8: %v", len(seenRules), seenRules)
	}
}

// --- gap-loop cycle-1 item 6 (REVIEW-A.md C3): validate-time hard FINDING
// for a spec.json "default" that does not type-check against its own
// declared "type" -------------------------------------------------------
//
// C3's materialization increment (engine/read.go's materializeConfigDefaults)
// fills an absent config key from spec.json's "default" verbatim; a default
// whose JSON type mismatches the property's declared type would silently
// materialize a wrong-shaped config value (e.g. default: 100 landing in a
// string-typed RuntimeConfig.Config, or a non-boolean string landing where a
// boolean was declared) — a hard validate FINDING (not a warning: this is a
// structural defect a bundle author can and must fix, unlike N2's
// plausibility heuristic below).

func TestValidate_DefaultTypeMismatchIsHardFinding(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "default-type-mismatch")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	var found *Finding
	for i := range report.Findings {
		if report.Findings[i].Connector == "default-type-mismatch" && report.Findings[i].Rule == ruleDefaultTypeMismatch {
			found = &report.Findings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("validateDir(default-type-mismatch): expected a %q finding, got %+v", ruleDefaultTypeMismatch, report.Findings)
	}
	if !strings.Contains(found.Message, "max_pages") {
		t.Fatalf("finding message %q does not name the offending property max_pages", found.Message)
	}
}

// TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule proves a
// well-typed default (base_url's string default in the same seeded bundle)
// never itself triggers the rule — only the genuinely mismatched property
// (max_pages) does.
func TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "default-type-mismatch")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, f := range report.Findings {
		if f.Rule == ruleDefaultTypeMismatch && strings.Contains(f.Message, "base_url") {
			t.Fatalf("well-typed base_url default incorrectly flagged: %+v", f)
		}
	}
}

// --- N2 (wave0 REVIEW.md carried flag): validate-time WARNING for a
// digit-shaped non-unix start_config_key value ------------------------------
//
// N2's narrow, honest scope (SPEC.md §4 "noted, not blocking... promote to a
// validate-time guard"): formatParam's digits-passthrough (B1) is CORRECT
// for param_format unix_seconds (an all-digits config value there really
// does mean Unix seconds) but is a silent-misinterpretation risk for
// timestamp-parsing param formats such as date or rfc3339_utc, where a
// free-form (no declared date-ish format) start_config_key spec property could hold a value like
// "20260101" (yyyymmdd) that would be silently treated as a 1970s-era
// Unix-seconds lower bound instead of erroring. A property that DOES
// declare format:date-time (or format:date) is not flagged: an operator
// filling in a date-time-typed config field is exceedingly unlikely to type
// a bare yyyymmdd digit string, and the risk is specifically about
// UNDECLARED free-form string config. This is a WARNING (Report.Warnings,
// not Report.Findings) — never blocks validate's exit code or the "0
// findings" self-verify contract — because it is a plausibility heuristic,
// not a structural defect: a legitimately-Unix-seconds start_config_key
// used with a timestamp-parsing output format (unusual but not
// inexpressible) would otherwise be a false positive if this were a hard error.

func TestValidate_StartDateFreeFormStringWarns(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "start-date-free-form-string")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero hard Findings (this is warning-only), got %+v", report.Findings)
	}
	var found *Finding
	for i := range report.Warnings {
		if report.Warnings[i].Connector == "start-date-free-form-string" && report.Warnings[i].Rule == ruleStartDateFreeFormString {
			found = &report.Warnings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected a %s warning for start-date-free-form-string, got %+v", ruleStartDateFreeFormString, report.Warnings)
	}
	if found.File == "" || found.Message == "" {
		t.Fatalf("warning %+v missing file/message", found)
	}
}

// TestValidate_StartDateWithDateTimeFormatNoWarning is the no-false-positive
// companion: an identical incremental shape whose start_config_key spec
// property DOES declare format:date-time must not warn.
func TestValidate_StartDateWithDateTimeFormatNoWarning(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "start-date-rfc3339-format-no-warning")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, w := range report.Warnings {
		if w.Connector == "start-date-rfc3339-format-no-warning" && w.Rule == ruleStartDateFreeFormString {
			t.Fatalf("unexpected %s warning for a format:date-time start_config_key: %+v", ruleStartDateFreeFormString, w)
		}
	}
}

// TestValidate_UnixSecondsStartDateNeverWarns locks in that param_format
// unix_seconds is never flagged by this rule at all (digits ARE the correct
// shape there) — reusing the real stripe golden, whose start_date has no
// declared format annotation either, proves this directly against
// production defs, not just a synthetic fixture.
func TestValidate_UnixSecondsStartDateNeverWarns(t *testing.T) {
	report, err := validateDir(os.DirFS(filepath.Join("..", "..", "internal", "connectors", "defs")))
	if err != nil {
		t.Fatalf("validateDir(defs): %v", err)
	}
	for _, w := range report.Warnings {
		if w.Rule == ruleStartDateFreeFormString {
			t.Fatalf("unexpected %s warning against the real defs corpus: %+v", ruleStartDateFreeFormString, w)
		}
	}
}

func TestValidate_IncrementalParamFormatRejectsWhitespace(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/widgets" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"records": { "path": "data" },
				"incremental": { "cursor_field": "updated_at", "request_param": "since", "param_format": "rfc3339_utc " },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}
	fsys["cli-surface/schemas/widgets.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "updated_at",
		"properties": {
			"id": { "type": "integer" },
			"name": { "type": "string" },
			"updated_at": { "type": "string", "format": "date-time" }
		}
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleIncrementalPolicy)
}

// TestValidate_ExitCodeReflectsFindings exercises the run() entry point (the
// one main() calls) end to end: a directory with findings must exit 1; a
// clean directory must exit 0.
func TestValidate_ExitCodeReflectsFindings(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/invalid/bad-spec-schema"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run(validate invalid) exit = %d, want 1; stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"validate", "testdata/valid"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(validate valid) exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

// --- validate --json shape ---------------------------------------------------

func TestValidate_JSONOutputShape(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/invalid/bad-spec-schema", "--json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run --json exit = %d, want 1", code)
	}

	var generic map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &generic); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\noutput: %s", err, stdout.String())
	}
	if _, ok := generic["connectors_checked"]; !ok {
		t.Fatalf("--json output missing connectors_checked: %s", stdout.String())
	}
	findingsRaw, ok := generic["findings"].([]any)
	if !ok || len(findingsRaw) == 0 {
		t.Fatalf("--json output missing non-empty findings array: %s", stdout.String())
	}
	entry, ok := findingsRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("findings[0] is not an object: %v", findingsRaw[0])
	}
	for _, key := range []string{"connector", "file", "rule", "message"} {
		if _, ok := entry[key]; !ok {
			t.Fatalf("findings[0] missing key %q: %s", key, stdout.String())
		}
	}
}

// TestValidate_JSONOutputCleanRun asserts the --json shape on a passing run
// too (empty findings array, not a missing key / null).
func TestValidate_JSONOutputCleanRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/valid", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run --json (clean) exit = %d, want 0; stderr=%s", code, stderr.String())
	}
	var generic map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &generic); err != nil {
		t.Fatalf("--json output is not valid JSON: %v", err)
	}
	findingsRaw, ok := generic["findings"].([]any)
	if !ok {
		t.Fatalf("--json output findings is not an array: %s", stdout.String())
	}
	if len(findingsRaw) != 0 {
		t.Fatalf("--json output findings should be empty for a clean run, got %v", findingsRaw)
	}
}

// --- gen: deterministic byte-stable regeneration -----------------------------

func TestGen_HooksetWritesEmptyImportList(t *testing.T) {
	hooksRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(hooksRoot, "hookset"), 0o755); err != nil {
		t.Fatalf("mkdir hookset: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read hookset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), "package hookset") {
		t.Fatalf("hookset_gen.go missing package clause: %s", raw)
	}
	if !strings.Contains(string(raw), "Code generated") {
		t.Fatalf("hookset_gen.go missing generated-by header: %s", raw)
	}
	if strings.Contains(string(raw), "_ \"") {
		t.Fatalf("hookset_gen.go should have an empty import list (no hooks/<name> packages exist yet): %s", raw)
	}
}

func TestGen_HooksetImportsEveryHookPackageExceptHookset(t *testing.T) {
	hooksRoot := t.TempDir()
	for _, name := range []string{"hookset", "acme"} {
		if err := os.MkdirAll(filepath.Join(hooksRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "acme", "hooks.go"), []byte("package acme\n"), 0o644); err != nil {
		t.Fatalf("write acme/hooks.go: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read hookset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), `_ "polymetrics.ai/internal/connectors/hooks/acme"`) {
		t.Fatalf("hookset_gen.go missing blank import for acme: %s", raw)
	}
	if strings.Contains(string(raw), "/hooks/hookset\"") {
		t.Fatalf("hookset_gen.go must not import itself: %s", raw)
	}
}

func TestGen_NativesetWritesEmptyImportListWhenNoNativePackages(t *testing.T) {
	nativeRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(nativeRoot, "nativeset"), 0o755); err != nil {
		t.Fatalf("mkdir nativeset: %v", err)
	}

	if err := genNativeset(nativeRoot); err != nil {
		t.Fatalf("genNativeset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(nativeRoot, "nativeset", "nativeset_gen.go"))
	if err != nil {
		t.Fatalf("read nativeset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), "package nativeset") {
		t.Fatalf("nativeset_gen.go missing package clause: %s", raw)
	}
	if strings.Contains(string(raw), "_ \"") {
		t.Fatalf("nativeset_gen.go should have an empty import list, got: %s", raw)
	}
}

func TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries(t *testing.T) {
	nativeRoot := t.TempDir()
	for _, name := range []string{"nativeset", "postgres", "dbtest", "sqltls"} {
		if err := os.MkdirAll(filepath.Join(nativeRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(nativeRoot, "postgres", "connector.go"), []byte("package postgres\n"), 0o644); err != nil {
		t.Fatalf("write postgres/connector.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nativeRoot, "dbtest", "harness.go"), []byte("package dbtest\n"), 0o644); err != nil {
		t.Fatalf("write dbtest/harness.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nativeRoot, "sqltls", "sqltls.go"), []byte("package sqltls\n"), 0o644); err != nil {
		t.Fatalf("write sqltls/sqltls.go: %v", err)
	}

	if err := genNativeset(nativeRoot); err != nil {
		t.Fatalf("genNativeset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(nativeRoot, "nativeset", "nativeset_gen.go"))
	if err != nil {
		t.Fatalf("read nativeset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), `_ "polymetrics.ai/internal/connectors/native/postgres"`) {
		t.Fatalf("nativeset_gen.go missing blank import for postgres: %s", raw)
	}
	for _, supportPackage := range []string{"dbtest", "sqltls"} {
		if strings.Contains(string(raw), "/native/"+supportPackage+`"`) {
			t.Fatalf("nativeset_gen.go must not import support package %q: %s", supportPackage, raw)
		}
	}
}

// TestGen_ByteStableOnRerun is the core determinism guarantee: running gen
// twice against the same input tree must produce byte-identical output.
func TestGen_ByteStableOnRerun(t *testing.T) {
	hooksRoot := t.TempDir()
	for _, name := range []string{"hookset", "acme", "beta"} {
		if err := os.MkdirAll(filepath.Join(hooksRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "acme", "hooks.go"), []byte("package acme\n"), 0o644); err != nil {
		t.Fatalf("write acme/hooks.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "beta", "hooks.go"), []byte("package beta\n"), 0o644); err != nil {
		t.Fatalf("write beta/hooks.go: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset (1st): %v", err)
	}
	first, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read 1st: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset (2nd): %v", err)
	}
	second, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read 2nd: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("genHookset not byte-stable across reruns:\n1st: %s\n2nd: %s", first, second)
	}
}

// TestGen_RunCommandRegeneratesBothFiles exercises `connectorgen gen` through
// run() against a scratch tree shaped like the real repo layout.
func TestGen_RunCommandRegeneratesBothFiles(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, "internal/connectors/hooks")
	nativeDir := filepath.Join(root, "internal/connectors/native")
	if err := os.MkdirAll(filepath.Join(hooksDir, "hookset"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nativeDir, "nativeset"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runGenAt([]string{"gen"}, &stdout, &stderr, hooksDir, nativeDir)
	if code != 0 {
		t.Fatalf("runGenAt(gen) exit = %d, stderr=%s", code, stderr.String())
	}

	if _, err := os.Stat(filepath.Join(hooksDir, "hookset", "hookset_gen.go")); err != nil {
		t.Fatalf("hookset_gen.go not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nativeDir, "nativeset", "nativeset_gen.go")); err != nil {
		t.Fatalf("nativeset_gen.go not written: %v", err)
	}
}

// --- new: scaffolding ---------------------------------------------------------

func TestNew_ScaffoldsBundleThatPassesValidate(t *testing.T) {
	root := t.TempDir()

	if err := scaffoldNew(root, "acme-widgets"); err != nil {
		t.Fatalf("scaffoldNew: %v", err)
	}

	for _, f := range []string{"metadata.json", "spec.json", "streams.json", "api_surface.json", "docs.md"} {
		if _, err := os.Stat(filepath.Join(root, "acme-widgets", f)); err != nil {
			t.Fatalf("scaffold missing %s: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "acme-widgets", "schemas")); err != nil {
		t.Fatalf("scaffold missing schemas/ dir: %v", err)
	}

	report, err := validateDir(os.DirFS(root))
	if err != nil {
		t.Fatalf("validateDir(scaffold): %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("scaffolded bundle failed validate: %+v", report.Findings)
	}
}

func TestNew_RejectsInvalidName(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Acme", "-acme", "acme_widgets", "", "acme widgets"} {
		if err := scaffoldNew(root, name); err == nil {
			t.Fatalf("scaffoldNew(%q) succeeded, want name-regex rejection", name)
		}
	}
}

func TestNew_RejectsExistingDir(t *testing.T) {
	root := t.TempDir()
	if err := scaffoldNew(root, "acme-widgets"); err != nil {
		t.Fatalf("scaffoldNew (first): %v", err)
	}
	if err := scaffoldNew(root, "acme-widgets"); err == nil {
		t.Fatalf("scaffoldNew (second, same name) succeeded, want existing-dir rejection")
	}
}

// TestNew_RunCommandScaffolds exercises `connectorgen new <name>` through
// run() end to end.
func TestNew_RunCommandScaffolds(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runNewAt([]string{"new", "acme-widgets"}, &stdout, &stderr, root)
	if code != 0 {
		t.Fatalf("runNewAt(new) exit = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "acme-widgets", "metadata.json")); err != nil {
		t.Fatalf("new did not scaffold metadata.json: %v", err)
	}
}

func TestNew_RunCommandMissingArgIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"new"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(new) with no name should fail, exit = 0")
	}
}

// --- main() usage / unknown subcommand ----------------------------------------

func TestRun_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"bogus"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(bogus) should fail, exit = 0")
	}
}

func TestRun_NoArgsIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(nil) should fail, exit = 0")
	}
}

// --- test helpers --------------------------------------------------------------

// singleBundleFS returns an fs.FS rooted at parent that exposes exactly one
// top-level directory (name), so validateDir (which walks its root looking
// for candidate bundle dirs, exactly like engine.LoadAll) sees only that one
// bundle and not any of parent's other sibling seeded-invalid fixtures.
func singleBundleFS(t *testing.T, parent, name string) fs.FS {
	t.Helper()
	return onlyDirFS{FS: os.DirFS(parent), name: name}
}

func assertFindingRule(t *testing.T, report Report, connector, rule string) {
	t.Helper()
	for _, f := range report.Findings {
		if f.Connector == connector && f.Rule == rule {
			return
		}
	}
	t.Fatalf("no finding for connector %q with rule %q; findings=%+v", connector, rule, report.Findings)
}

func cliSurfaceBundleFS(cliSurface string) fstest.MapFS {
	return fstest.MapFS{
		"cli-surface/metadata.json": &fstest.MapFile{Data: []byte(`{
			"name": "cli-surface",
			"display_name": "CLI Surface",
			"description": "test connector",
			"integration_type": "api",
			"release_stage": "ga",
			"capabilities": { "check": true, "read": true, "write": true, "query": false, "cdc": false, "dynamic_schema": false }
		}`)},
		"cli-surface/spec.json": &fstest.MapFile{Data: []byte(`{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"required": ["base_url"],
			"properties": {
				"base_url": { "type": "string" }
			}
		}`)},
		"cli-surface/streams.json": &fstest.MapFile{Data: []byte(`{
			"base": {
				"url": "{{ config.base_url }}",
				"check": { "method": "GET", "path": "/widgets" }
			},
			"streams": [
				{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
			]
		}`)},
		"cli-surface/writes.json": &fstest.MapFile{Data: []byte(`{
			"actions": [
				{
					"name": "create_widget",
					"kind": "create",
					"method": "POST",
					"path": "/widgets",
					"record_schema": { "type": "object", "required": ["name"], "properties": { "name": { "type": "string" } } },
					"risk": "creates a widget"
				}
			]
		}`)},
		"cli-surface/api_surface.json": &fstest.MapFile{Data: []byte(`{
			"api": "test API v1",
			"endpoints": [
				{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
				{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } }
			]
		}`)},
		"cli-surface/schemas/widgets.json": &fstest.MapFile{Data: []byte(`{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"x-primary-key": ["id"],
			"properties": {
				"id": { "type": "integer" },
				"name": { "type": "string" }
			}
		}`)},
		"cli-surface/docs.md": &fstest.MapFile{Data: []byte(`# Overview

test

## Auth setup

none

## Streams notes

none

## Write actions & risks

none

## Known limits

none
`)},
		"cli-surface/cli_surface.json": &fstest.MapFile{Data: []byte(cliSurface)},
	}
}

func operationCLISurfaceBundleFS(cliSurface, operations string) fstest.MapFS {
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/widgets/{id}", "operation": { "model": "direct_read", "status": "blocked", "risk": "low", "blocked_by_default": true, "reason": "typed operation metadata" } }
		]
	}`)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(operations)}
	return fsys
}

func validOperationsJSON() string {
	return `{
		"operations": [
			{
				"id": "cli-surface.widgets.get",
				"kind": "rest_read",
				"summary": "Read widget metadata",
				"risk": "low",
				"approval": "none",
				"output_policy": "json_redacted",
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}",
					"max_bytes": 1024
				}
			}
		]
	}`
}

func validOperationCLISurfaceJSON() string {
	return `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget view",
				"summary": "View widget metadata",
				"intent": "direct_read",
				"availability": "implemented",
				"operation": "cli-surface.widgets.get",
				"api_surface": [
					{ "method": "GET", "path": "/widgets/{id}" }
				],
				"output_policy": "json_redacted",
				"source_cli_path": "clis widget view",
				"examples": ["pm cli-surface widget view --id w_1 --json"]
			}
		]
	}`
}

func validCLISurfaceValidationJSON() string {
	return strings.Replace(validCLISurfaceJSON(), `"api_surface": [
					{ "method": "GET", "path": "/widgets" }
				],`, `"api_surface": [
					{ "method": "GET", "path": "/widgets" }
				],
				"flags": [
					{ "name": "start", "type": "string", "summary": "Start bound.", "maps_to": "query.started_after", "format": "date-time", "allow_empty": false },
					{ "name": "end", "type": "string", "summary": "End bound.", "maps_to": "query.started_before", "format": "date-time", "allow_empty": false }
				],
				"constraints": [
					{ "kind": "order", "left": "query.started_after", "left_fallback": "config.default_start", "op": "lt", "right": "query.started_before", "value_type": "date-time", "message": "start must be before end" }
				],`, 1)
}

func validCLISurfaceJSON() string {
	return `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget list",
				"summary": "List widgets",
				"intent": "etl",
				"availability": "implemented",
				"stream": "widgets",
				"source_cli_path": "clis widget list",
				"api_surface": [
					{ "method": "GET", "path": "/widgets" }
				],
				"examples": ["pm cli-surface widget list --json"]
			},
			{
				"path": "widget create",
				"summary": "Create a widget",
				"intent": "reverse_etl",
				"availability": "implemented",
				"write": "create_widget",
				"source_cli_path": "clis widget create",
				"api_surface": [
					{ "method": "POST", "path": "/widgets" }
				],
				"flags": [
					{ "name": "name", "type": "string", "summary": "Widget name.", "maps_to": "record.name" }
				],
				"risk": "creates a widget",
				"approval": "reverse ETL writes require plan, preview, approval, execute",
				"examples": ["pm cli-surface widget create --json"]
			}
		]
	}`
}

func validDirectReadCLISurfaceJSON() string {
	return `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget read",
				"summary": "Read widget metadata",
				"intent": "direct_read",
				"availability": "implemented",
				"source_cli_path": "clis widget read",
				"api_surface": [
					{ "method": "GET", "path": "/widgets/{path}" }
				],
				"output_policy": "repository_contents_file_metadata",
				"flags": [
					{ "name": "path", "type": "string", "maps_to": "path.path" }
				],
				"examples": ["pm cli-surface widget read --path README.md --json"]
			}
		]
	}`
}

func directReadCLISurfaceBundleFS(cliSurface string) fstest.MapFS {
	return directReadCLISurfaceBundleFSWithAPI(cliSurface, validDirectReadAPISurface())
}

func directReadCLISurfaceBundleFSWithAPI(cliSurface, apiSurface string) fstest.MapFS {
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	return fsys
}

func directWriteCLISurfaceBundleFS() fstest.MapFS {
	fsys := cliSurfaceBundleFS(`{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget archive",
				"summary": "Archive a widget",
				"intent": "direct_write",
				"availability": "implemented",
				"operation": "cli-surface.widgets.archive",
				"source_cli_path": "clis widget archive",
				"api_surface": [
					{ "method": "POST", "path": "/widgets/{id}/archive" }
				],
				"output_policy": "json_redacted",
				"flags": [
					{ "name": "id", "type": "string", "maps_to": "path.id" }
				],
				"risk": "archives a widget",
				"approval": "requires plan, preview, typed confirmation, and execute",
				"examples": ["pm cli-surface widget archive --id w_1 --json"]
			}
		]
	}`)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{
				"method": "POST",
				"path": "/widgets/{id}/archive",
				"operation": {
					"model": "destructive_action",
					"status": "blocked",
					"risk": "high",
					"blocked_by_default": true,
					"reason": "Typed rest_write operation",
					"notes": "Declared only through the direct-write executor."
				}
			}
		]
	}`)}
	fsys["cli-surface/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "cli-surface.widgets.archive",
				"kind": "rest_write",
				"summary": "Archive a widget",
				"risk": "high",
				"approval": "requires plan, preview, typed confirmation, and execute",
				"output_policy": "json_redacted",
				"mutation_class": "destructive",
				"destructive": true,
				"confirmation": { "kind": "destructive" },
				"batchable": false,
				"rest": {
					"method": "POST",
					"path": "/widgets/{id}/archive",
					"max_bytes": 1024
				}
			}
		]
	}`)}
	return fsys
}

func operationLedgerBundleFS(apiSurface string) fstest.MapFS {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(apiSurface)}
	return fsys
}

func validOperationLedgerAPISurface() string {
	return `{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{
				"method": "GET",
				"path": "/widgets/{path}",
				"operation": {
					"model": "direct_read",
					"status": "blocked",
					"risk": "low",
					"blocked_by_default": true,
					"reason": "point lookup candidate, not yet modeled as a stream",
					"source_url": "https://example.invalid/rest/widgets"
				}
			}
		]
	}`
}

func validV2ProvenanceAPISurface() string {
	return `{
		"api": "test API v2",
		"operation_ledger_version": 2,
		"artifacts": [
			{
				"id": "acme-openapi-2026-08-06",
				"url": "https://docs.acme.test/openapi.yaml",
				"retrieved_at": "2026-08-06",
				"sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}
		],
		"endpoints": [
			{
				"method": "GET",
				"path": "/widgets",
				"provenance": {
					"artifact": "acme-openapi-2026-08-06",
					"source_url": "https://docs.acme.test/api/widgets"
				},
				"covered_by": { "stream": "widgets" }
			},
			{
				"method": "POST",
				"path": "/widgets",
				"provenance": {
					"artifact": "acme-openapi-2026-08-06",
					"source_url": "https://docs.acme.test/api/widgets#create"
				},
				"covered_by": { "write": "create_widget" }
			},
			{
				"method": "POST",
				"path": "/widgets/sensitive",
				"provenance": {
					"artifact": "acme-openapi-2026-08-06",
					"source_url": "https://docs.acme.test/api/widgets#sensitive"
				},
				"operation": {
					"model": "sensitive_reverse_etl",
					"status": "blocked",
					"risk": "high",
					"blocked_by_default": true,
					"reason": "requires sensitive-data safeguards"
				}
			}
		]
	}`
}

func validDirectReadAPISurface() string {
	return `{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{
				"method": "GET",
				"path": "/widgets/{path}",
				"covered_by": {
					"direct_read": "widget read"
				}
			}
		]
	}`
}

func newBoundaryCommandFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	writeBoundaryCommandFile(t, root, "go.mod", "module polymetrics.ai\n\ngo 1.25\n")
	writeBoundaryCommandFile(t, root, "internal/connectors/defs/github/metadata.json", `{"name":"github","display_name":"GitHub"}`)
	writeBoundaryCommandFile(t, root, "internal/connectors/defs/gong/metadata.json", `{"name":"gong","display_name":"Gong"}`)
	for rel, content := range files {
		writeBoundaryCommandFile(t, root, rel, content)
	}
	return root
}

func writeBoundaryCommandFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// onlyDirFS wraps an fs.FS and restricts ReadDir(".") to a single named
// entry, while passing every other operation straight through (so reads
// inside name/... still resolve normally).
type onlyDirFS struct {
	fs.FS
	name string
}

func (o onlyDirFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(o.FS, dir)
	if err != nil {
		return nil, err
	}
	if dir != "." {
		return entries, nil
	}
	var out []fs.DirEntry
	for _, e := range entries {
		if e.Name() == o.name {
			out = append(out, e)
		}
	}
	return out, nil
}
