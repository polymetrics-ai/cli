package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/failures"
)

func TestDeclarationAdmissionAcceptsRunnableReadAndExplicitDeferrals(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("admission findings = %+v, want none", findings)
	}
}

func TestDeclarationAdmissionRejectsCompletenessAndBindingDefects(t *testing.T) {
	tests := []struct {
		name string
		edit func(*declarationAdmissionDocument, *engine.Bundle)
		want string
	}{
		{
			name: "omitted source row",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.Declarations = document.Declarations[1:]
			},
			want: "has no declaration",
		},
		{
			name: "duplicate declaration",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.Declarations = append(document.Declarations, document.Declarations[0])
			},
			want: "duplicate declaration",
		},
		{
			name: "citation free",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = ""
			},
			want: "source URL",
		},
		{
			name: "malformed provider URL",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "not a URL"
			},
			want: "source URL",
		},
		{
			name: "stale canonical endpoint",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.Declarations[0].Canonical.Path = "/v1/renamed-widgets"
			},
			want: "stale canonical endpoint",
		},
		{
			name: "base path mismatch",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].BasePath = "/v2"
			},
			want: "base-path mismatch",
		},
		{
			name: "lane change",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.Declarations[0].Lane = declarationAdmissionLaneBinaryDownload
			},
			want: "lane",
		},
		{
			name: "delete lacks destructive metadata",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].SourceID == "delete-widget" {
						document.Declarations[index].Destructive = nil
					}
				}
			},
			want: "destructive metadata",
		},
		{
			name: "false implementation",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.CLISurface.Commands[0].Operation = "missing.operation"
			},
			want: "runtime binding",
		},
		{
			name: "false implemented write binding",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].SourceID == "create-widget" {
						document.Declarations[index].State = declarationAdmissionStateImplemented
						document.Declarations[index].Foundation = nil
					}
				}
				bundle.CLISurface.Commands[1].Availability = declarationAdmissionStateImplemented
				bundle.CLISurface.Commands[1].Foundation = nil
				bundle.CLISurface.Commands[1].Write = "create_widget"
				bundle.Writes = []engine.WriteAction{{Name: "create_widget", Method: "POST", Path: "/v1/other-widgets"}}
			},
			want: "runtime binding",
		},
		{
			name: "missing foundation",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].State == declarationAdmissionStateDeferred {
						document.Declarations[index].Foundation = nil
						break
					}
				}
				bundle.CLISurface.Commands[1].Foundation = nil
			},
			want: "foundation gap",
		},
		{
			name: "deferred foundation has no implementation evidence",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].State == declarationAdmissionStateDeferred {
						document.Declarations[index].Foundation.Evidence = ""
						break
					}
				}
			},
			want: "specific missing implementation component",
		},
		{
			name: "stale typed write action foundation",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].SourceID != "create-widget" {
						continue
					}
					document.Declarations[index].Foundation.Component = "typed_write_action"
					document.Declarations[index].Foundation.Evidence = "writes.json has no create-widget action"
					break
				}
				bundle.Writes = append(bundle.Writes, engine.WriteAction{Name: "create_widget", Method: "POST", Path: "/v1/widgets"})
			},
			want: "typed_write_action foundation is stale",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle, document := declarationAdmissionFixture()
			testCase.edit(&document, &bundle)
			findings := declarationAdmissionFindings(bundle, document)
			if len(findings) == 0 {
				t.Fatalf("findings = none, want %q", testCase.want)
			}
			for _, finding := range findings {
				if strings.Contains(finding.Message, testCase.want) {
					return
				}
			}
			t.Fatalf("findings = %+v, want message containing %q", findings, testCase.want)
		})
	}
}

func TestDeclarationAdmissionAcceptsCompleteZeroRunnableConnector(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	for index := range document.Declarations {
		if document.Declarations[index].State == declarationAdmissionStateImplemented {
			document.Declarations[index].State = declarationAdmissionStateDeferred
			document.Declarations[index].Foundation = &declarationAdmissionFoundation{
				ID: "read_runtime_foundation", Reason: "read runtime binding is intentionally pending",
				Component: "runtime_executor", Evidence: "direct-read executor binding has not been implemented",
			}
			bundle.CLISurface.Commands[index].Availability = declarationAdmissionStateDeferred
			bundle.CLISurface.Commands[index].Foundation = &engine.CommandFoundation{ID: "read_runtime_foundation", Reason: "read runtime binding is intentionally pending"}
		}
	}
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("zero-runnable deferred connector findings = %+v, want none", findings)
	}
}

func TestDeclarationAdmissionAdmitsGitHubImplementedDeleteControl(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repository root: %v", err)
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join(root, "internal", "connectors", "defs")), "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	document := declarationAdmissionDocument{
		SchemaVersion: declarationAdmissionSchemaVersion,
		Connector:     "github",
		SourceOperations: []declarationAdmissionSourceOperation{{
			ID:                  "github.rest.issues.delete-label",
			SourceURL:           "https://raw.githubusercontent.com/github/rest-api-description/b26c240ded1c8b79cb0fb09dee4a21239061fa23/descriptions/api.github.com/api.github.com.json",
			Location:            `paths["/repos/{owner}/{repo}/labels/{name}"].delete`,
			ProviderOperationID: "issues/delete-label",
			Method:              "DELETE",
			Path:                "/repos/{owner}/{repo}/labels/{name}",
		}},
		Declarations: []declarationAdmissionDeclaration{{
			SourceID:  "github.rest.issues.delete-label",
			Lane:      declarationAdmissionLaneReverseETL,
			Command:   "label delete",
			State:     declarationAdmissionStateImplemented,
			Canonical: declarationAdmissionEndpoint{Method: "DELETE", Path: "/repos/{owner}/{repo}/labels/{name}"},
			Destructive: &declarationAdmissionDestructive{
				Kind: "delete", Reason: "deletes a repository label and its existing issue metadata",
			},
		}},
	}
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("GitHub implemented delete findings = %+v, want none", findings)
	}
	if err := commandrunner.Preflight(engine.New(bundle, nil), []string{"label", "delete"}); err != nil {
		t.Fatalf("GitHub label delete runtime preflight: %v", err)
	}
}

func TestDeclarationAdmissionDeferredDeleteIsDiscoverable(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("deferred declaration findings = %+v, want none", findings)
	}
	command, matches := declarationAdmissionCommand(bundle, "widgets delete")
	if matches != 1 || command.Availability != declarationAdmissionStateDeferred || command.Foundation == nil {
		t.Fatalf("deferred delete command = %+v (matches=%d), want one deferred foundation command", command, matches)
	}
	err := commandrunner.Preflight(engine.New(bundle, nil), []string{"widgets", "delete"})
	var blocked *commandrunner.BlockedCommandError
	if !errors.As(err, &blocked) || blocked.Failure == nil {
		t.Fatalf("deferred delete preflight = %v, want typed blocked missing-foundation error", err)
	}
	if blocked.Failure.Code() != "missing_foundation" || blocked.Failure.Domain() != failures.DomainSystem || command.Foundation.ID == "" {
		t.Fatalf("deferred delete foundation = %+v, want system/missing_foundation with named gap", blocked)
	}
}

func TestDeclarationAdmissionRejectsPolicyOnlyDeferredFoundation(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	for index := range document.Declarations {
		if document.Declarations[index].State != declarationAdmissionStateDeferred {
			continue
		}
		document.Declarations[index].Foundation = &declarationAdmissionFoundation{
			ID: "blocked_by_default", Reason: "operation is blocked by default", Component: "blocked_by_default", Evidence: "api_surface policy",
		}
		bundle.CLISurface.Commands[index].Foundation = &engine.CommandFoundation{ID: "blocked_by_default", Reason: "operation is blocked by default"}
		break
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal policy-only declaration: %v", err)
	}
	if err := engine.ValidateDeclarationAdmission(raw); err == nil {
		t.Fatal("policy-only foundation passed declaration-admission schema")
	}
	findings := declarationAdmissionFindings(bundle, document)
	for _, finding := range findings {
		if strings.Contains(finding.Message, "specific missing implementation component") {
			return
		}
	}
	t.Fatalf("policy-only deferred foundation findings = %+v, want specific-component refusal", findings)
}

func TestDeclarationAdmissionCommandReportsMachineReadableCleanResult(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"declaration-admission", "testdata/valid", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var report declarationAdmissionReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, stdout.String())
	}
	if report.ConnectorsChecked != 0 || len(report.Findings) != 0 {
		t.Fatalf("clean report = %+v, want no admission sidecars/findings", report)
	}
}

func TestDeclarationAdmissionCommandLoadsCitedSidecarWithoutRetainedArtifact(t *testing.T) {
	defsRoot := t.TempDir()
	for name, file := range cliSurfaceBundleFS(validCLISurfaceJSON()) {
		path := filepath.Join(defsRoot, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create fixture directory: %v", err)
		}
		if err := os.WriteFile(path, file.Data, 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	document := declarationAdmissionDocument{
		SchemaVersion: declarationAdmissionSchemaVersion,
		Connector:     "cli-surface",
		SourceOperations: []declarationAdmissionSourceOperation{{
			ID: "list-widgets", SourceURL: "https://provider.example.test/reference", Location: "Widgets > List", ProviderOperationID: "listWidgets",
			Method: "GET", Path: "/widgets",
		}},
		Declarations: []declarationAdmissionDeclaration{{
			SourceID: "list-widgets", Lane: declarationAdmissionLaneETL, Command: "widget list", State: declarationAdmissionStateImplemented,
			Canonical: declarationAdmissionEndpoint{Method: "GET", Path: "/widgets"},
		}},
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal admission declaration: %v", err)
	}
	sidecar := filepath.Join(defsRoot, "cli-surface", "sources", "cli-surface-declaration-admission.json")
	if err := os.MkdirAll(filepath.Dir(sidecar), 0o755); err != nil {
		t.Fatalf("create admission directory: %v", err)
	}
	if err := os.WriteFile(sidecar, raw, 0o600); err != nil {
		t.Fatalf("write admission declaration: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"declaration-admission", defsRoot, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var report declarationAdmissionReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, stdout.String())
	}
	if report.ConnectorsChecked != 1 || report.SourceOperations != 1 || len(report.Findings) != 0 {
		t.Fatalf("report = %+v, want one clean source-cited operation", report)
	}
}

func TestDeclarationAdmissionDeferredCLISurfaceNeedsFoundationGap(t *testing.T) {
	bundle, _ := declarationAdmissionFixture()
	bundle.CLISurface.Commands[1].Foundation = nil
	findings := checkCLISurfaceFoundation(bundle, 1, bundle.CLISurface.Commands[1])
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "requires foundation_gap") {
		t.Fatalf("deferred foundation findings = %+v", findings)
	}
	bundle.CLISurface.Commands[1].Availability = declarationAdmissionStateImplemented
	bundle.CLISurface.Commands[1].Foundation = &engine.CommandFoundation{ID: "write_plan_foundation", Reason: "write plan compiler is not available"}
	findings = checkCLISurfaceFoundation(bundle, 1, bundle.CLISurface.Commands[1])
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "requires deferred availability") {
		t.Fatalf("implemented foundation findings = %+v", findings)
	}
}

func TestDeclarationAdmissionNormalizesWriteActionTemplatePath(t *testing.T) {
	if got := declarationAdmissionActionPath("/v1/widgets/{{ record.widget_id }}"); got != "/v1/widgets/{widget_id}" {
		t.Fatalf("normalized action path = %q", got)
	}
}

func declarationAdmissionFixture() (engine.Bundle, declarationAdmissionDocument) {
	commands := []engine.CLICommand{
		{
			Path: "widgets list", Intent: "direct_read", Availability: declarationAdmissionStateImplemented,
			Operation: "acme.widgets.list", APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/v1/widgets"}},
		},
		{
			Path: "widgets create", Intent: "reverse_etl", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/v1/widgets"}},
			Foundation: &engine.CommandFoundation{ID: "write_plan_foundation", Reason: "write plan compiler is not available"},
		},
		{
			Path: "widgets delete", Intent: "reverse_etl", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/v1/widgets/{id}"}},
			Foundation: &engine.CommandFoundation{ID: "delete_plan_foundation", Reason: "delete plan compiler is not available"},
		},
		{
			Path: "widgets download", Intent: "binary_download", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/v1/widgets/{id}/archive"}},
			Foundation: &engine.CommandFoundation{ID: "binary_download_foundation", Reason: "binary response binding is not available"},
		},
		{
			Path: "widgets upload", Intent: "binary_upload", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/v1/widgets/{id}/archive"}},
			Foundation: &engine.CommandFoundation{ID: "binary_upload_foundation", Reason: "binary request binding is not available"},
		},
		{
			Path: "widgets descriptor", Intent: "direct_write", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "PATCH", Path: "/v1/widgets/{id}"}},
			Foundation: &engine.CommandFoundation{ID: "source_descriptor_importer", Reason: "provider descriptor importer cannot yet represent this request"},
		},
	}
	operations := []engine.OperationSpec{{
		ID: "acme.widgets.list", Kind: "rest_read", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/v1/widgets"},
	}}
	bundle := engine.Bundle{
		Name:       "acme",
		Operations: operations,
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{
			{Method: "GET", Path: "/v1/widgets"}, {Method: "POST", Path: "/v1/widgets"}, {Method: "DELETE", Path: "/v1/widgets/{id}"},
			{Method: "GET", Path: "/v1/widgets/{id}/archive"}, {Method: "POST", Path: "/v1/widgets/{id}/archive"}, {Method: "PATCH", Path: "/v1/widgets/{id}"},
		}},
		CLISurface: &engine.CLISurface{Commands: commands},
	}
	rows := []struct {
		id, method, sourcePath, command, lane, state, foundationID, foundationReason, foundationComponent, foundationEvidence string
		destructive                                                                                                           *declarationAdmissionDestructive
	}{
		{"list-widgets", "GET", "/widgets", "widgets list", declarationAdmissionLaneDirectRead, declarationAdmissionStateImplemented, "", "", "", "", nil},
		{"create-widget", "POST", "/widgets", "widgets create", declarationAdmissionLaneReverseETL, declarationAdmissionStateDeferred, "write_plan_foundation", "write plan compiler is not available", "runtime_executor", "reverse-ETL plan compiler has no widget binding", nil},
		{"delete-widget", "DELETE", "/widgets/{id}", "widgets delete", declarationAdmissionLaneReverseETL, declarationAdmissionStateDeferred, "delete_plan_foundation", "delete plan compiler is not available", "runtime_executor", "reverse-ETL plan compiler has no widget delete binding", &declarationAdmissionDestructive{Kind: "delete", Reason: "provider operation deletes a widget"}},
		{"download-widget", "GET", "/widgets/{id}/archive", "widgets download", declarationAdmissionLaneBinaryDownload, declarationAdmissionStateDeferred, "binary_download_foundation", "binary response binding is not available", "binary_transfer_binding", "binary response binding has not been implemented", nil},
		{"upload-widget", "POST", "/widgets/{id}/archive", "widgets upload", declarationAdmissionLaneBinaryUpload, declarationAdmissionStateDeferred, "binary_upload_foundation", "binary request binding is not available", "binary_transfer_binding", "binary request binding has not been implemented", nil},
		{"patch-widget", "PATCH", "/widgets/{id}", "widgets descriptor", declarationAdmissionLaneDirectWrite, declarationAdmissionStateDeferred, "source_descriptor_importer", "provider descriptor importer cannot yet represent this request", "source_importer", "provider descriptor importer cannot represent the operation", nil},
	}
	document := declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: "acme"}
	for _, row := range rows {
		document.SourceOperations = append(document.SourceOperations, declarationAdmissionSourceOperation{
			ID: row.id, SourceURL: "https://provider.example.test/v1/reference", Location: "Widgets > " + row.id, ProviderOperationID: "provider." + row.id,
			Method: row.method, Path: row.sourcePath, BasePath: "/v1",
		})
		declaration := declarationAdmissionDeclaration{
			SourceID: row.id, Lane: row.lane, Command: row.command, State: row.state,
			Canonical: declarationAdmissionEndpoint{Method: row.method, Path: "/v1" + row.sourcePath}, Destructive: row.destructive,
		}
		if row.foundationID != "" {
			declaration.Foundation = &declarationAdmissionFoundation{
				ID: row.foundationID, Reason: row.foundationReason, Component: row.foundationComponent, Evidence: row.foundationEvidence,
			}
		}
		document.Declarations = append(document.Declarations, declaration)
	}
	return bundle, document
}
