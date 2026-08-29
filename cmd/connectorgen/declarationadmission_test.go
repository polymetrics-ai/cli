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
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/failures"
)

func TestDeclarationAdmissionAcceptsRunnableReadAndExplicitDeferrals(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("admission findings = %+v, want none", findings)
	}
}

func TestImplementedCommandEndpointEquivalenceCoversExactFleet(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	loadFailed := err != nil
	if err != nil {
		t.Errorf("load production bundles: %v", err)
	}
	aliases := 0
	graphql := 0
	for _, bundle := range bundles {
		surface := engine.New(bundle, nil).CommandSurface()
		if surface == nil {
			continue
		}
		for _, command := range surface.Commands {
			if command.Availability != declarationAdmissionStateImplemented ||
				(command.Stream == "" && command.Write == "" && command.Operation == "") || len(command.APISurface) != 1 {
				continue
			}
			resolved, resolveErr := engine.ResolveImplementedCommandBinding(bundle, command)
			if resolveErr != nil {
				t.Errorf("%s %s: %v", bundle.Name, command.Path, resolveErr)
				continue
			}
			if resolved.Method == "GRAPHQL" {
				graphql++
				if resolved.Equivalence != engine.CommandEndpointGraphQLTransport {
					t.Errorf("%s %s: GraphQL equivalence = %q", bundle.Name, command.Path, resolved.Equivalence)
				}
				continue
			}
			if !strings.EqualFold(resolved.Method, resolved.TransportMethod) || resolved.Path != resolved.TransportPath {
				aliases++
				if resolved.Equivalence == "" || resolved.Equivalence == engine.CommandEndpointExact {
					t.Errorf("%s %s: aliased endpoint has proof %q (%s %s -> %s %s)", bundle.Name, command.Path, resolved.Equivalence, resolved.TransportMethod, resolved.TransportPath, resolved.Method, resolved.Path)
				}
			}
		}
	}
	if !loadFailed && (aliases != 246 || graphql != 4) {
		t.Fatalf("proved endpoint aliases = %d non-GraphQL and %d GraphQL, want 246 and 4", aliases, graphql)
	}
	t.Logf("proved endpoint aliases = %d non-GraphQL and %d GraphQL", aliases, graphql)
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
			name: "duplicate provider operation with a different binding",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				duplicate := document.SourceOperations[0]
				duplicate.ID = "list-widgets-alias"
				duplicate.Binding = declarationAdmissionBinding{Kind: "stream", ID: "widgets_alias"}
				document.SourceOperations = append(document.SourceOperations, duplicate)
			},
			want: "duplicate exact provider operation identity",
		},
		{
			name: "duplicate canonical binding for different provider operations",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				duplicate := document.SourceOperations[0]
				duplicate.ID = "get-widget"
				duplicate.Location = "Widgets > get-widget"
				duplicate.ProviderOperationID = "provider.get-widget"
				duplicate.Path = "/widgets/{id}"
				document.SourceOperations = append(document.SourceOperations, duplicate)
			},
			want: "claim one canonical binding",
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
			name: "source-declared destructive post lacks destructive metadata",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				for index := range document.SourceOperations {
					if document.SourceOperations[index].ID == "create-widget" {
						document.SourceOperations[index].DestructiveKind = "destructive"
					}
				}
				for index := range bundle.Surface.Endpoints {
					if bundle.Surface.Endpoints[index].Method == "POST" && bundle.Surface.Endpoints[index].Path == "/v1/widgets" {
						bundle.Surface.Endpoints[index].Operation.Model = "destructive_action"
						return
					}
				}
			},
			want: "destructive operation lacks destructive metadata",
		},
		{
			name: "false implementation",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.CLISurface.Commands[0].Stream = "missing_stream"
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
				bundle.Writes = []engine.WriteAction{{Name: "different_create_widget", Method: "POST", Path: "/v1/widgets"}}
			},
			want: "runtime binding",
		},
		{
			name: "false implemented delete semantics",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].SourceID == "delete-widget" {
						document.Declarations[index].State = declarationAdmissionStateImplemented
						document.Declarations[index].Foundation = nil
					}
				}
				bundle.CLISurface.Commands[2].Availability = declarationAdmissionStateImplemented
				bundle.CLISurface.Commands[2].Foundation = nil
				bundle.CLISurface.Commands[2].Write = "delete_widget"
				bundle.Writes = []engine.WriteAction{{
					Name: "delete_widget", Kind: "update", Method: "DELETE", Path: "/v1/widgets/{{ record.id }}",
					RecordSchema: json.RawMessage(`{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}}`),
				}}
			},
			want: "destructive runtime metadata",
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
					document.Declarations[index].Foundation.Evidence = "write_action_absent"
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

func TestDeclarationAdmissionRejectsNoncanonicalProviderCitations(t *testing.T) {
	tests := []struct {
		name      string
		sourceURL string
		want      string
	}{
		{name: "uppercase DNS host", sourceURL: "https://PROVIDER.EXAMPLE.TEST/v1/reference", want: "canonical provider source URL"},
		{name: "explicit default HTTPS port", sourceURL: "https://provider.example.test:443/v1/reference", want: "canonical provider source URL"},
		{name: "unstable query order", sourceURL: "https://provider.example.test/v1/reference?b=2&a=1", want: "canonical provider source URL"},
		{name: "non-normalized escaped path", sourceURL: "https://provider.example.test/v1/%72eference", want: "canonical provider source URL"},
		{name: "trailing-dot DNS host", sourceURL: "https://provider.example.test./v1/reference", want: "valid provider source URL"},
		{name: "ambiguous empty DNS label", sourceURL: "https://provider..example.test/v1/reference", want: "valid provider source URL"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle, document := declarationAdmissionFixture()
			document.SourceOperations[0].SourceURL = testCase.sourceURL
			if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, testCase.want) {
				t.Fatalf("findings = %+v, want %q refusal", findings, testCase.want)
			}
		})
	}
}

func TestDeclarationAdmissionCanonicalCitationIdentityRejectsDuplicateProviderOperationAcrossBindings(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	document.SourceOperations[0].SourceURL = "https://provider.example.test/v1/reference?a=1&b=2"
	duplicate := document.SourceOperations[0]
	duplicate.ID = "list-widgets-alias"
	duplicate.SourceURL = "https://PROVIDER.EXAMPLE.TEST:443/v1/reference?b=2&a=1"
	duplicate.Binding = declarationAdmissionBinding{Kind: "stream", ID: "widgets_alias"}
	document.SourceOperations = append(document.SourceOperations, duplicate)

	findings := declarationAdmissionFindings(bundle, document)
	if !declarationAdmissionFindingContains(findings, "canonical provider source URL") {
		t.Fatalf("findings = %+v, want noncanonical citation refusal", findings)
	}
	if !declarationAdmissionFindingContains(findings, "duplicate exact provider operation identity") {
		t.Fatalf("findings = %+v, want canonical citation duplicate identity", findings)
	}
}

func TestDeclarationAdmissionAuditRepairRejectsWeakIdentityCitationAndDeferredTarget(t *testing.T) {
	tests := []struct {
		name string
		edit func(*declarationAdmissionDocument, *engine.Bundle)
		want string
	}{
		{
			name: "duplicate exact provider operation identity",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				duplicate := document.SourceOperations[0]
				duplicate.ID = "list-widgets-duplicate-id"
				document.SourceOperations = append(document.SourceOperations, duplicate)
				declaration := document.Declarations[0]
				declaration.SourceID = duplicate.ID
				document.Declarations = append(document.Declarations, declaration)
			},
			want: "duplicate exact provider operation identity",
		},
		{
			name: "insecure citation",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "http://provider.example.test/v1/reference"
			},
			want: "valid provider source URL",
		},
		{
			name: "credential-shaped citation query",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "https://provider.example.test/v1/reference?api_key=not-a-secret"
			},
			want: "valid provider source URL",
		},
		{
			name: "private literal citation",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "https://127.0.0.1/v1/reference"
			},
			want: "valid provider source URL",
		},
		{
			name: "userinfo citation",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "https://user@provider.example.test/v1/reference"
			},
			want: "valid provider source URL",
		},
		{
			name: "fragment citation",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				document.SourceOperations[0].SourceURL = "https://provider.example.test/v1/reference#operation"
			},
			want: "valid provider source URL",
		},
		{
			name: "noncanonical command path",
			edit: func(document *declarationAdmissionDocument, bundle *engine.Bundle) {
				document.Declarations[0].Command = "widgets  list"
				bundle.CLISurface.Commands[0].Path = "widgets  list"
			},
			want: "canonical round-trippable command path",
		},
		{
			name: "implemented target is excluded",
			edit: func(_ *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.Surface.Endpoints[0].Excluded = &engine.SurfaceExclusion{Category: "not_runnable", Reason: "runtime binding must not override an excluded source row"}
			},
			want: "does not map to the canonical API surface endpoint",
		},
		{
			name: "implemented target is policy only",
			edit: func(_ *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.Surface.Endpoints[0].Operation = &engine.SurfaceOperation{Model: "disallowed", Status: "blocked", BlockedByDefault: true, Reason: "policy-only source row"}
			},
			want: "does not map to the canonical API surface endpoint",
		},
		{
			name: "implemented target is duplicated",
			edit: func(_ *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.Surface.Endpoints = append(bundle.Surface.Endpoints, bundle.Surface.Endpoints[0])
			},
			want: "does not map to the canonical API surface endpoint",
		},
		{
			name: "command target reference is duplicated",
			edit: func(_ *declarationAdmissionDocument, bundle *engine.Bundle) {
				bundle.CLISurface.Commands[0].APISurface = append(bundle.CLISurface.Commands[0].APISurface, bundle.CLISurface.Commands[0].APISurface[0])
			},
			want: "does not map to the canonical API surface endpoint",
		},
		{
			name: "delete metadata on post",
			edit: func(document *declarationAdmissionDocument, _ *engine.Bundle) {
				for index := range document.Declarations {
					if document.Declarations[index].SourceID == "create-widget" {
						document.Declarations[index].Destructive = &declarationAdmissionDestructive{Kind: "delete", Reason: "incorrectly calls a create operation a delete"}
						return
					}
				}
			},
			want: "delete semantics do not match the independent source semantic",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle, document := declarationAdmissionFixture()
			testCase.edit(&document, &bundle)
			findings := declarationAdmissionFindings(bundle, document)
			for _, finding := range findings {
				if strings.Contains(finding.Message, testCase.want) {
					return
				}
			}
			t.Fatalf("findings = %+v, want message containing %q", findings, testCase.want)
		})
	}
}

func TestDeclarationAdmissionAuditRepairRejectsDeferredPolicyTargetBeforeMissingFoundation(t *testing.T) {
	tests := []struct {
		name string
		edit func(*engine.Bundle)
	}{
		{
			name: "excluded endpoint",
			edit: func(bundle *engine.Bundle) {
				for index := range bundle.Surface.Endpoints {
					if bundle.Surface.Endpoints[index].Method == "DELETE" {
						bundle.Surface.Endpoints[index].Excluded = &engine.SurfaceExclusion{Category: "destructive_admin", Reason: "policy-only exclusion"}
						return
					}
				}
			},
		},
		{
			name: "disallowed policy operation",
			edit: func(bundle *engine.Bundle) {
				for index := range bundle.Surface.Endpoints {
					if bundle.Surface.Endpoints[index].Method == "DELETE" {
						bundle.Surface.Endpoints[index].Operation.Model = "disallowed"
						return
					}
				}
			},
		},
		{
			name: "duplicate exact endpoint",
			edit: func(bundle *engine.Bundle) {
				for _, endpoint := range bundle.Surface.Endpoints {
					if endpoint.Method == "DELETE" {
						bundle.Surface.Endpoints = append(bundle.Surface.Endpoints, endpoint)
						return
					}
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle, document := declarationAdmissionFixture()
			testCase.edit(&bundle)
			if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, "deferred target") {
				t.Fatalf("findings = %+v, want deferred target refusal", findings)
			}

			err := commandrunner.Preflight(engine.New(bundle, nil), []string{"widgets", "delete"})
			var blocked *commandrunner.BlockedCommandError
			if !errors.As(err, &blocked) {
				t.Fatalf("deferred policy target preflight = %v, want blocked error", err)
			}
			if blocked.Failure != nil && blocked.Failure.Code() == "missing_foundation" {
				t.Fatalf("deferred policy target reached missing_foundation: %+v", blocked)
			}
		})
	}
}

func declarationAdmissionFindingContains(findings []Finding, want string) bool {
	for _, finding := range findings {
		if strings.Contains(finding.Message, want) {
			return true
		}
	}
	return false
}

func TestDeclarationAdmissionAcceptsCompleteZeroRunnableConnector(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	for index := range document.Declarations {
		if document.Declarations[index].State == declarationAdmissionStateImplemented {
			document.Declarations[index].State = declarationAdmissionStateDeferred
			document.Declarations[index].Foundation = &declarationAdmissionFoundation{
				ID: "read_runtime_foundation", Reason: "read runtime binding is intentionally pending",
				Component: "runtime_executor", Evidence: "runtime_executor_absent",
				Target: declarationAdmissionEndpoint{Method: "GET", Path: "/v1/widgets"},
			}
			bundle.CLISurface.Commands[index].Availability = declarationAdmissionStateDeferred
			bundle.CLISurface.Commands[index].Foundation = &engine.CommandFoundation{
				ID: "read_runtime_foundation", Reason: "read runtime binding is intentionally pending",
				Component: "runtime_executor", Evidence: "runtime_executor_absent",
				Target: engine.CommandFoundationTarget{
					SourceID: "list-widgets", ProviderOperationID: "provider.list-widgets",
					Binding:         engine.CommandBindingIdentity{Kind: "stream", ID: "widgets"},
					DestructiveKind: "none", Method: "GET", Path: "/v1/widgets",
				},
			}
			bundle.Surface.Endpoints[0].Operation = &engine.SurfaceOperation{Model: "direct_read", Status: "blocked", Risk: "low", BlockedByDefault: true, Reason: "read runtime binding is intentionally pending"}
			bundle.CLISurface.Commands[index].Stream = ""
			bundle.Streams = nil
		}
	}
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("zero-runnable deferred connector findings = %+v, want none", findings)
	}
}

func TestDeclarationAdmissionRejectsMissingCommandProjection(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	document.Declarations[0].Command = ""
	document.Declarations[0].State = declarationAdmissionStateDeferred
	document.Declarations[0].Foundation = &declarationAdmissionFoundation{
		ID: "command_projection_foundation", Reason: "the command path encoder is not available",
		Component: "runtime_executor", Evidence: "runtime_executor_absent",
		Target: declarationAdmissionEndpoint{Method: "GET", Path: "/v1/widgets"},
	}
	bundle.CLISurface.Commands = bundle.CLISurface.Commands[1:]

	if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, "discoverable command mapping") {
		t.Fatalf("missing command projection findings = %+v, want discoverable-command refusal", findings)
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
			Protocol:            "rest",
			SourceURL:           "https://raw.githubusercontent.com/github/rest-api-description/b26c240ded1c8b79cb0fb09dee4a21239061fa23/descriptions/api.github.com/api.github.com.json",
			Location:            `paths["/repos/{owner}/{repo}/labels/{name}"].delete`,
			ProviderOperationID: "issues/delete-label",
			Method:              "DELETE",
			Path:                "/repos/{owner}/{repo}/labels/{name}",
			Binding:             declarationAdmissionBinding{Kind: "write", ID: "delete_label"},
			DestructiveKind:     "delete",
		}},
		Declarations: []declarationAdmissionDeclaration{{
			SourceID:  "github.rest.issues.delete-label",
			Lane:      declarationAdmissionLaneReverseETL,
			Command:   "label delete",
			State:     declarationAdmissionStateImplemented,
			Canonical: declarationAdmissionEndpoint{Method: "DELETE", Path: "/repos/{owner}/{repo}/labels/{name}"},
			Binding:   declarationAdmissionBinding{Kind: "write", ID: "delete_label"},
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

// TestDeclarationAdmissionOutreachRealBundleResolverCompatibility loads the
// real stream/write shapes but synthesizes Outreach's absent discovery surface
// in memory. It proves only generic admission and resolver compatibility; it is
// not shipped CLI, credential-boundary, source-evidence, or zero-transport proof.
func TestDeclarationAdmissionOutreachRealBundleResolverCompatibility(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repository root: %v", err)
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join(root, "internal", "connectors", "defs")), "outreach")
	if err != nil {
		t.Fatalf("load Outreach bundle: %v", err)
	}
	bundle.CLISurface = &engine.CLISurface{Commands: []engine.CLICommand{
		{
			Path: "prospects list", Intent: declarationAdmissionLaneETL, Availability: declarationAdmissionStateImplemented,
			Stream: "prospects", APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/api/v2/prospects"}},
		},
		{
			Path: "accounts delete", Intent: declarationAdmissionLaneReverseETL, Availability: declarationAdmissionStateImplemented,
			Write: "delete_account", APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/api/v2/accounts/{id}"}},
		},
	}}
	document := declarationAdmissionDocument{
		SchemaVersion: declarationAdmissionSchemaVersion,
		Connector:     "outreach",
		SourceOperations: []declarationAdmissionSourceOperation{
			{
				ID: "outreach.prospects.list", Protocol: "rest", SourceURL: "https://developers.outreach.io/api",
				Location: "Prospect resource > list prospects", ProviderOperationID: "prospects/list",
				Method: "GET", BasePath: "/api/v2", Path: "/prospects",
				Binding: declarationAdmissionBinding{Kind: "stream", ID: "prospects"}, DestructiveKind: "none",
			},
			{
				ID: "outreach.accounts.delete", Protocol: "rest", SourceURL: "https://developers.outreach.io/api",
				Location: "Account resource > delete account", ProviderOperationID: "accounts/delete",
				Method: "DELETE", BasePath: "/api/v2", Path: "/accounts/{id}",
				Binding: declarationAdmissionBinding{Kind: "write", ID: "delete_account"}, DestructiveKind: "delete",
			},
		},
		Declarations: []declarationAdmissionDeclaration{
			{
				SourceID: "outreach.prospects.list", Lane: declarationAdmissionLaneETL, Command: "prospects list",
				State: declarationAdmissionStateImplemented, Canonical: declarationAdmissionEndpoint{Method: "GET", Path: "/api/v2/prospects"},
				Binding: declarationAdmissionBinding{Kind: "stream", ID: "prospects"},
			},
			{
				SourceID: "outreach.accounts.delete", Lane: declarationAdmissionLaneReverseETL, Command: "accounts delete",
				State: declarationAdmissionStateImplemented, Canonical: declarationAdmissionEndpoint{Method: "DELETE", Path: "/api/v2/accounts/{id}"},
				Binding:     declarationAdmissionBinding{Kind: "write", ID: "delete_account"},
				Destructive: &declarationAdmissionDestructive{Kind: "delete", Reason: "deletes an Outreach account resource"},
			},
		},
	}
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("Outreach real-bundle compatibility findings = %+v", findings)
	}
	connector := engine.New(bundle, nil)
	for _, path := range [][]string{{"prospects", "list"}, {"accounts", "delete"}} {
		if err := commandrunner.Preflight(connector, path); err != nil {
			t.Fatalf("Outreach synthetic-discovery %v preflight: %v", path, err)
		}
	}
}

func TestDeclarationAdmissionUsesRuntimeResolverForImplementedCommandKinds(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repository root: %v", err)
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join(root, "internal", "connectors", "defs")), "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}

	tests := []struct {
		name, command, lane, protocol, method, path, bindingKind, bindingID string
	}{
		{name: "templated ETL stream", command: "issue list", lane: declarationAdmissionLaneETL, protocol: "rest", method: "GET", path: "/repos/{owner}/{repo}/issues", bindingKind: "stream", bindingID: "issues"},
		{name: "operation-free direct read", command: "users get-authenticated", lane: declarationAdmissionLaneDirectRead, protocol: "rest", method: "GET", path: "/user", bindingKind: "command", bindingID: "users get-authenticated"},
		{name: "GraphQL ETL read", command: "discussion list", lane: declarationAdmissionLaneETL, protocol: "graphql", method: "GRAPHQL", path: "ListDiscussions", bindingKind: "stream", bindingID: "discussions"},
		{name: "GraphQL direct write", command: "discussion create", lane: declarationAdmissionLaneDirectWrite, protocol: "graphql", method: "POST", path: "/graphql", bindingKind: "operation", bindingID: "github.graphql.mutation.create-discussion"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			document := declarationAdmissionDocument{
				SchemaVersion: declarationAdmissionSchemaVersion,
				Connector:     "github",
				SourceOperations: []declarationAdmissionSourceOperation{{
					ID: "github.audit." + strings.ReplaceAll(testCase.command, " ", "-"), SourceURL: "https://docs.github.com/rest",
					Location: "audit > " + testCase.command, ProviderOperationID: "audit/" + strings.ReplaceAll(testCase.command, " ", "-"),
					Protocol: testCase.protocol, Method: testCase.method, Path: testCase.path,
					Binding: declarationAdmissionBinding{Kind: testCase.bindingKind, ID: testCase.bindingID}, DestructiveKind: "none",
				}},
				Declarations: []declarationAdmissionDeclaration{{
					SourceID: "github.audit." + strings.ReplaceAll(testCase.command, " ", "-"), Lane: testCase.lane,
					Command: testCase.command, State: declarationAdmissionStateImplemented,
					Canonical: declarationAdmissionEndpoint{Method: testCase.method, Path: testCase.path},
					Binding:   declarationAdmissionBinding{Kind: testCase.bindingKind, ID: testCase.bindingID},
				}},
			}
			if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
				t.Fatalf("implemented runtime-resolved command findings = %+v", findings)
			}
		})
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
		bundle.CLISurface.Commands[index].Foundation = &engine.CommandFoundation{
			ID: "blocked_by_default", Reason: "operation is blocked by default", Component: "blocked_by_default", Evidence: "api_surface policy",
			Target: engine.CommandFoundationTarget{Method: bundle.CLISurface.Commands[index].APISurface[0].Method, Path: bundle.CLISurface.Commands[index].APISurface[0].Path},
		}
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

func TestDeclarationAdmissionCommandRejectsMissingIndependentCohort(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"declaration-admission", "testdata/valid", "--json"}, &stdout, &stderr); code == 0 {
		t.Fatalf("missing independent cohort passed; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestDeclarationAdmissionAllowsEmptyRawProviderOperationID(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	document.SourceOperations[0].ProviderOperationID = ""
	if findings := declarationAdmissionFindings(bundle, document); len(findings) != 0 {
		t.Fatalf("empty provider operation ID findings = %+v, want deterministic source ID to remain authoritative", findings)
	}
}

func TestDeclarationAdmissionRejectsDeferredDeleteSemanticsOnDestructivePost(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	for index := range bundle.Surface.Endpoints {
		if bundle.Surface.Endpoints[index].Method == "POST" && bundle.Surface.Endpoints[index].Path == "/v1/widgets" {
			bundle.Surface.Endpoints[index].Operation.Model = "destructive_action"
		}
	}
	for index := range document.Declarations {
		if document.Declarations[index].SourceID == "create-widget" {
			document.Declarations[index].Destructive = &declarationAdmissionDestructive{Kind: "delete", Reason: "self-labelled delete"}
		}
	}
	if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, "delete semantics") {
		t.Fatalf("destructive POST false-delete findings = %+v, want exact source semantic rejection", findings)
	}
}

func TestDeclarationAdmissionRejectsSameEndpointBindingSwap(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	bundle.Streams = append(bundle.Streams, engine.StreamSpec{Name: "mirror_widgets", Method: "GET", Path: "/v1/widgets"})
	document.SourceOperations[0].Binding.ID = "mirror_widgets"
	if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, "canonical binding does not match") {
		t.Fatalf("same-endpoint binding swap findings = %+v, want independent binding mismatch", findings)
	}
}

func TestDeclarationAdmissionRejectsFalseImplementedSameEndpointBinding(t *testing.T) {
	bundle, document := declarationAdmissionFixture()
	bundle.Streams = append(bundle.Streams, engine.StreamSpec{Name: "mirror_widgets", Method: "GET", Path: "/v1/widgets"})
	document.SourceOperations[0].Binding.ID = "mirror_widgets"
	document.Declarations[0].Binding.ID = "mirror_widgets"
	if findings := declarationAdmissionFindings(bundle, document); !declarationAdmissionFindingContains(findings, "implemented declaration has no valid runtime binding") {
		t.Fatalf("false implemented same-endpoint binding findings = %+v, want runtime identity rejection", findings)
	}
}

func TestDeclarationAdmissionCommandRequiresInventoryAndBothCatalogs(t *testing.T) {
	tests := []struct {
		name string
		edit func(*testing.T, string)
	}{
		{
			name: "missing independent inventory",
			edit: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.Remove(filepath.Join(dir, "declaration_admission_inventory.json")); err != nil {
					t.Fatalf("remove declaration inventory: %v", err)
				}
			},
		},
		{
			name: "missing source cohort",
			edit: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.Remove(filepath.Join(dir, "declaration_admission_sources.json")); err != nil {
					t.Fatalf("remove source cohort: %v", err)
				}
			},
		},
		{
			name: "missing declaration catalog",
			edit: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.Remove(filepath.Join(dir, "declaration_admissions.json")); err != nil {
					t.Fatalf("remove declaration catalog: %v", err)
				}
			},
		},
		{
			name: "legacy source count escape hatch",
			edit: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "declaration_admission_sources.json")
				var catalog map[string]any
				readDeclarationAdmissionJSON(t, path, &catalog)
				catalog["expected_source_operations"] = 1
				writeDeclarationAdmissionJSON(t, path, catalog)
			},
		},
		{
			name: "legacy declaration count escape hatch",
			edit: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "declaration_admissions.json")
				var catalog map[string]any
				readDeclarationAdmissionJSON(t, path, &catalog)
				catalog["expected_declarations"] = 1
				writeDeclarationAdmissionJSON(t, path, catalog)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defsRoot := declarationAdmissionCatalogFixtureDir(t)
			testCase.edit(t, defsRoot)
			var stdout, stderr bytes.Buffer
			if code := run([]string{"declaration-admission", defsRoot, "--json"}, &stdout, &stderr); code == 0 {
				t.Fatalf("catalog defect passed: stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
		})
	}
}

func TestDeclarationAdmissionCommandLoadsReviewedLockInventoryWithoutReadingRetainedArtifact(t *testing.T) {
	defsRoot := declarationAdmissionCatalogFixtureDir(t)

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

func TestDeclarationAdmissionMappingEvidenceDoesNotRequireRetentionMetadata(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "absent bytes and hash",
			mutate: func(rest map[string]any) {
				delete(rest, "bytes")
				delete(rest, "sha256")
			},
		},
		{
			name: "malformed byte count",
			mutate: func(rest map[string]any) {
				rest["bytes"] = -7
			},
		},
		{
			name: "malformed hash",
			mutate: func(rest map[string]any) {
				rest["sha256"] = "not-a-sha256"
			},
		},
		{
			name: "wrong retention wire types",
			mutate: func(rest map[string]any) {
				rest["bytes"] = "not-a-byte-count"
				rest["sha256"] = map[string]any{"not": "a hash"}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defsRoot := declarationAdmissionCatalogFixtureDir(t)
			lockPath := filepath.Join(defsRoot, "cli-surface", "sources", "cli-surface-operation-source-lock.json")
			var lock map[string]any
			readDeclarationAdmissionJSON(t, lockPath, &lock)
			testCase.mutate(lock["rest"].(map[string]any))
			writeDeclarationAdmissionJSON(t, lockPath, lock)

			raw, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("read mutated source lock: %v", err)
			}
			if _, err := parseSourceImportLock(raw, "cli-surface"); err == nil {
				t.Fatal("full source-import parser accepted invalid retention metadata")
			}

			report, err := declarationAdmissionPathCheck(defsRoot)
			if err != nil {
				t.Fatalf("mapping-only declaration admission rejected retention metadata: %v", err)
			}
			if len(report.Findings) != 0 {
				t.Fatalf("mapping-only declaration admission findings = %+v, want none", report.Findings)
			}
		})
	}
}

func TestDeclarationAdmissionMappingDoesNotRequireCertificationOverlay(t *testing.T) {
	defsRoot := declarationAdmissionCatalogFixtureDir(t)
	certificationPath := filepath.Join(defsRoot, "cli-surface", "certification.json")
	if err := os.WriteFile(certificationPath, []byte(`{"schema_version":"invalid"}`), 0o600); err != nil {
		t.Fatalf("write malformed certification overlay: %v", err)
	}

	report, err := declarationAdmissionPathCheck(defsRoot)
	if err != nil {
		t.Fatalf("declaration admission read certification-only overlay: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("declaration admission findings = %+v, want mapping independent of certification", report.Findings)
	}
}

func TestDeclarationAdmissionV3MappingEvidenceDoesNotRequireRetentionMetadata(t *testing.T) {
	defsRoot := declarationAdmissionCatalogFixtureDir(t)
	lockPath := filepath.Join(defsRoot, "cli-surface", "sources", "cli-surface-operation-source-lock.json")
	operation := map[string]any{
		"id": "provider.rest.widgets.list", "protocol": "rest", "method": "GET", "path": "/widgets",
		"operation_id": "provider.list-widgets", "deprecated": false, "source_location": "Widgets > List",
	}
	lock := map[string]any{
		"schema_version": 3,
		"connector":      "cli-surface",
		"captured_at":    "retention metadata is irrelevant to mapping admission",
		"rest": map[string]any{
			"retrieval": "retained artifact capture",
			"openapi":   []any{"3.0.3"},
			"source_documents": []any{map[string]any{
				"id": "primary",
				"artifact": map[string]any{
					"source_url": "https://retained.example.test/openapi.json", "sha256": "not-a-sha256", "bytes": -7, "openapi": "3.0.3",
				},
				"published_source": map[string]any{
					"source_url": "https://provider.example.test/openapi.json", "capture_url": 123,
					"sha256": map[string]any{"not": "a hash"}, "bytes": "not-a-count", "adapter": false,
				},
				"operations": []any{operation},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
	writeDeclarationAdmissionJSON(t, lockPath, lock)
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read v3 source lock: %v", err)
	}
	if _, err := parseSourceImportLock(raw, "cli-surface"); err == nil {
		t.Fatal("full source-import parser accepted malformed v3 retention metadata")
	}
	report, err := declarationAdmissionPathCheck(defsRoot)
	if err != nil {
		t.Fatalf("v3 mapping-only declaration admission rejected retention metadata: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("v3 mapping-only declaration admission findings = %+v, want none", report.Findings)
	}
}

// TestDeclarationAdmissionMappingV3EventSchemaInventoryIsClosedAndIgnored
// proves the mapping reader can admit the closed selector envelope without
// projecting event schema details into declaration admission. Source import
// remains responsible for proving selector/document resolution against the
// retained provider artifact.
func TestDeclarationAdmissionMappingV3EventSchemaInventoryIsClosedAndIgnored(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "valid source selector is ignored after structural admission",
		},
		{
			name: "unknown inventory field",
			mutate: func(inventory map[string]any) {
				inventory["provider_schema"] = map[string]any{"type": "object"}
			},
			want: `unknown field "provider_schema"`,
		},
		{
			name: "unknown nested selector field",
			mutate: func(inventory map[string]any) {
				inventory["schemas"].([]any)[0].(map[string]any)["provider_schema"] = true
			},
			want: `unknown field "provider_schema"`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			raw := declarationAdmissionV3EventSchemaInventoryFixture(t)
			if testCase.mutate != nil {
				var lock map[string]any
				if err := json.Unmarshal(raw, &lock); err != nil {
					t.Fatalf("decode v3 event-schema inventory fixture: %v", err)
				}
				rest := lock["rest"].(map[string]any)
				testCase.mutate(rest["event_schema_inventory"].(map[string]any))
				var err error
				raw, err = json.Marshal(lock)
				if err != nil {
					t.Fatalf("encode mutated v3 event-schema inventory fixture: %v", err)
				}
			}

			reviewed, err := parseDeclarationAdmissionSourceLock(raw, "fixture")
			if testCase.want != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.want) {
					t.Fatalf("mapping reader error = %v, want %q", err, testCase.want)
				}
				return
			}
			if err != nil {
				t.Fatalf("mapping reader rejected closed event-schema inventory: %v", err)
			}
			if got, want := len(reviewed.Operations), 1; got != want {
				t.Fatalf("reviewed operation count = %d, want %d", got, want)
			}
		})
	}
}

func TestDeclarationAdmissionMappingReaderAcceptsRetainedAsanaV3EventSchemaInventory(t *testing.T) {
	lockPath := filepath.Join("..", "..", "internal", "connectors", "defs", "asana", "sources", "asana-operation-source-lock.json")
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read retained Asana v3 source lock: %v", err)
	}
	reviewed, err := parseDeclarationAdmissionSourceLock(raw, "asana")
	if err != nil {
		t.Fatalf("mapping reader rejected retained Asana v3 event-schema inventory: %v", err)
	}
	if got, want := len(reviewed.Operations), 249; got != want {
		t.Fatalf("reviewed Asana operation count = %d, want %d", got, want)
	}
}

func declarationAdmissionV3EventSchemaInventoryFixture(t *testing.T) []byte {
	t.Helper()
	raw := sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{{
		ID:       "primary",
		Path:     "/widgets",
		Artifact: []byte(`{"openapi":"3.0.3"}`),
	}})
	return sourceImportAddEventSchemaInventory(t, raw, sourceImportEventInventory("primary", "EventResponse"))
}

func TestDeclarationAdmissionMappingEvidenceReaderPreservesIdentityValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "connector mismatch",
			mutate: func(lock map[string]any) {
				lock["connector"] = "other"
			},
			want: "does not match requested connector",
		},
		{
			name: "noncanonical provider URL",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["source_url"] = "http://provider.example.test/openapi.json"
			},
			want: "invalid REST source URL",
		},
		{
			name: "missing protocol",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["protocol"] = ""
			},
			want: "incomplete REST operation identity",
		},
		{
			name: "noncanonical provider operation ID",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["operation_id"] = " provider.list-widgets"
			},
			want: "incomplete REST operation identity",
		},
		{
			name: "missing method",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["method"] = ""
			},
			want: "incomplete REST operation identity",
		},
		{
			name: "invalid path",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["path"] = "widgets"
			},
			want: "invalid path",
		},
		{
			name: "missing location",
			mutate: func(lock map[string]any) {
				lock["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_location"] = ""
			},
			want: "incomplete REST operation identity",
		},
		{
			name: "duplicate operation",
			mutate: func(lock map[string]any) {
				rest := lock["rest"].(map[string]any)
				operations := rest["operations"].([]any)
				rest["operations"] = append(operations, operations[0])
				counts := lock["counts"].(map[string]any)
				counts["rest"] = 2
				counts["total"] = 2
			},
			want: "duplicates operation identity",
		},
		{
			name: "count mismatch",
			mutate: func(lock map[string]any) {
				lock["counts"].(map[string]any)["total"] = 2
			},
			want: "counts do not match",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defsRoot := declarationAdmissionCatalogFixtureDir(t)
			lockPath := filepath.Join(defsRoot, "cli-surface", "sources", "cli-surface-operation-source-lock.json")
			var lock map[string]any
			readDeclarationAdmissionJSON(t, lockPath, &lock)
			testCase.mutate(lock)
			writeDeclarationAdmissionJSON(t, lockPath, lock)
			raw, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("read mutated source lock: %v", err)
			}
			if _, err := parseDeclarationAdmissionSourceLock(raw, "cli-surface"); err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("mapping reader error = %v, want %q", err, testCase.want)
			}
		})
	}
}

func TestDeclarationAdmissionInventoryRequiresSchemaV2(t *testing.T) {
	v2Root := declarationAdmissionCatalogFixtureDir(t)
	v2Path := filepath.Join(v2Root, "declaration_admission_inventory.json")
	var v2 map[string]any
	readDeclarationAdmissionJSON(t, v2Path, &v2)
	v2["schema_version"] = 2
	writeDeclarationAdmissionJSON(t, v2Path, v2)
	if report, err := declarationAdmissionPathCheck(v2Root); err != nil || len(report.Findings) != 0 {
		t.Fatalf("schema-v2 inventory admission report=%+v err=%v, want accepted", report, err)
	}

	legacyRoot := declarationAdmissionCatalogFixtureDir(t)
	legacyPath := filepath.Join(legacyRoot, "declaration_admission_inventory.json")
	var legacy map[string]any
	readDeclarationAdmissionJSON(t, legacyPath, &legacy)
	legacy["schema_version"] = 1
	writeDeclarationAdmissionJSON(t, legacyPath, legacy)
	if report, err := declarationAdmissionPathCheck(legacyRoot); err == nil && len(report.Findings) == 0 {
		t.Fatalf("legacy schema-v1 inventory passed: report=%+v", report)
	}
}

func TestDeclarationAdmissionBindsSourceRowsToReviewedConnectorOwnedLock(t *testing.T) {
	tests := []struct {
		name string
		edit func(*testing.T, string)
	}{
		{
			name: "unrelated provider host",
			edit: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "declaration_admission_sources.json")
				var catalog declarationAdmissionSourceCatalog
				readDeclarationAdmissionJSON(t, path, &catalog)
				catalog.SourceOperations[0].SourceURL = "https://unrelated.example.test/openapi.json"
				writeDeclarationAdmissionJSON(t, path, catalog)
			},
		},
		{
			name: "nonexistent locked operation",
			edit: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "declaration_admission_inventory.json")
				var inventory map[string]any
				readDeclarationAdmissionJSON(t, path, &inventory)
				inventory["operations"].([]any)[0].(map[string]any)["source_operation_id"] = "provider.rest.widgets.missing"
				writeDeclarationAdmissionJSON(t, path, inventory)
			},
		},
		{
			name: "semantic location alias",
			edit: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "declaration_admission_sources.json")
				var catalog declarationAdmissionSourceCatalog
				readDeclarationAdmissionJSON(t, path, &catalog)
				catalog.SourceOperations[0].Location = "Widgets > List (semantic alias)"
				writeDeclarationAdmissionJSON(t, path, catalog)
			},
		},
		{
			name: "lock outside connector ownership",
			edit: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "declaration_admission_inventory.json")
				var inventory map[string]any
				readDeclarationAdmissionJSON(t, path, &inventory)
				inventory["operations"].([]any)[0].(map[string]any)["source_lock"] = "other/sources/cli-surface-operation-source-lock.json"
				writeDeclarationAdmissionJSON(t, path, inventory)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defsRoot := declarationAdmissionCatalogFixtureDir(t)
			testCase.edit(t, defsRoot)
			report, err := declarationAdmissionPathCheck(defsRoot)
			if err == nil && len(report.Findings) == 0 {
				t.Fatalf("source-lock defect passed: report=%+v", report)
			}
		})
	}
}

func TestDeclarationAdmissionIndependentInventoryCannotShrinkWithCatalogRows(t *testing.T) {
	defsRoot := declarationAdmissionCatalogFixtureDir(t)
	sourcePath := filepath.Join(defsRoot, "declaration_admission_sources.json")
	declarationPath := filepath.Join(defsRoot, "declaration_admissions.json")

	var sources map[string]any
	readDeclarationAdmissionJSON(t, sourcePath, &sources)
	var declarations map[string]any
	readDeclarationAdmissionJSON(t, declarationPath, &declarations)
	sources["source_operations"] = []any{}
	declarations["declarations"] = []any{}
	writeDeclarationAdmissionJSON(t, sourcePath, sources)
	writeDeclarationAdmissionJSON(t, declarationPath, declarations)

	report, err := declarationAdmissionPathCheck(defsRoot)
	if err == nil && len(report.Findings) == 0 {
		t.Fatalf("catalog row deletion passed while independent inventory retained its operation: report=%+v", report)
	}

	// Restoring the old adjacent-count escape hatch cannot change the source
	// inventory denominator: schema v2 rejects these mutable count fields.
	sources["expected_source_operations"] = 0
	sources["expected_connectors"] = 0
	declarations["expected_declarations"] = 0
	writeDeclarationAdmissionJSON(t, sourcePath, sources)
	writeDeclarationAdmissionJSON(t, declarationPath, declarations)
	if _, err := declarationAdmissionPathCheck(defsRoot); err == nil {
		t.Fatal("deleted rows plus decremented adjacent counts passed declaration admission")
	}
}

func TestDeclarationAdmissionAcceptsProviderEvidencedUnsupportedAcrossAllSixLanes(t *testing.T) {
	lanes := []string{
		declarationAdmissionLaneETL,
		declarationAdmissionLaneReverseETL,
		declarationAdmissionLaneDirectRead,
		declarationAdmissionLaneDirectWrite,
		declarationAdmissionLaneBinaryDownload,
		declarationAdmissionLaneBinaryUpload,
	}
	for _, lane := range lanes {
		t.Run(lane, func(t *testing.T) {
			defsRoot := declarationAdmissionUnsupportedFixtureDir(t, lane)
			report, err := declarationAdmissionPathCheck(defsRoot)
			if err != nil {
				t.Fatalf("unsupported %s admission: %v", lane, err)
			}
			if len(report.Findings) != 0 {
				t.Fatalf("unsupported %s findings = %+v, want none", lane, report.Findings)
			}
		})
	}
}

func declarationAdmissionCatalogFixtureDir(t *testing.T) string {
	t.Helper()
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
	sources := declarationAdmissionSourceCatalog{
		SchemaVersion: 2,
		Cohort:        "test-cohort",
		SourceOperations: []declarationAdmissionSourceOperation{{
			Connector: "cli-surface", ID: "list-widgets", Protocol: "rest", SourceURL: "https://provider.example.test/openapi.json", Location: "Widgets > List", ProviderOperationID: "provider.list-widgets",
			Method: "GET", Path: "/widgets", Binding: declarationAdmissionBinding{Kind: "stream", ID: "widgets"}, DestructiveKind: "none",
		}},
	}
	declarations := declarationAdmissionCatalog{
		SchemaVersion: 2,
		Cohort:        "test-cohort",
		Declarations: []declarationAdmissionDeclaration{{
			Connector: "cli-surface", SourceID: "list-widgets", Lane: declarationAdmissionLaneETL, Command: "widget list", State: declarationAdmissionStateImplemented,
			Canonical: declarationAdmissionEndpoint{Method: "GET", Path: "/widgets"}, Binding: declarationAdmissionBinding{Kind: "stream", ID: "widgets"},
		}},
	}
	for name, catalog := range map[string]any{
		"declaration_admission_sources.json": sources,
		"declaration_admissions.json":        declarations,
		"declaration_admission_inventory.json": map[string]any{
			"schema_version": 2,
			"cohort":         "test-cohort",
			"operations": []any{map[string]any{
				"connector": "cli-surface", "source_id": "list-widgets",
				"source_lock":         "cli-surface/sources/cli-surface-operation-source-lock.json",
				"source_operation_id": "provider.rest.widgets.list",
			}},
		},
	} {
		raw, err := json.Marshal(catalog)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(defsRoot, name), raw, 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	writeDeclarationAdmissionSourceLock(t, defsRoot, "cli-surface", []map[string]any{{
		"id": "provider.rest.widgets.list", "protocol": "rest", "method": "GET", "path": "/widgets",
		"operation_id": "provider.list-widgets", "deprecated": false, "source_location": "Widgets > List",
	}})
	return defsRoot
}

func declarationAdmissionUnsupportedFixtureDir(t *testing.T, lane string) string {
	t.Helper()
	defsRoot := declarationAdmissionCatalogFixtureDir(t)
	method := "GET"
	if lane == declarationAdmissionLaneReverseETL || lane == declarationAdmissionLaneDirectWrite || lane == declarationAdmissionLaneBinaryUpload {
		method = "POST"
	}
	command := "widget unsupported"
	target := map[string]any{
		"source_id": "list-widgets", "operation_id": "provider.list-widgets",
		"method": method, "path": "/widgets",
	}
	disposition := map[string]any{
		"reason": "the provider documents this operation but the CLI cannot support its semantics",
		"target": target,
	}
	cliSurface := map[string]any{
		"tagline": "Unsupported operation fixture", "usage": "pm cli-surface <command>",
		"commands": []any{map[string]any{
			"path": command, "summary": "Retain an unsupported provider operation", "intent": lane,
			"availability":            "unsupported_with_provider_evidence",
			"api_surface":             []any{map[string]any{"method": method, "path": "/widgets"}},
			"unsupported_disposition": disposition,
		}},
	}
	writeDeclarationAdmissionJSON(t, filepath.Join(defsRoot, "cli-surface", "cli_surface.json"), cliSurface)
	apiSurface := map[string]any{
		"api":       "test API v1",
		"endpoints": []any{map[string]any{"method": method, "path": "/widgets"}},
	}
	writeDeclarationAdmissionJSON(t, filepath.Join(defsRoot, "cli-surface", "api_surface.json"), apiSurface)
	sources := map[string]any{
		"schema_version": 2, "cohort": "test-cohort",
		"source_operations": []any{map[string]any{
			"connector": "cli-surface", "id": "list-widgets", "protocol": "rest",
			"source_url": "https://provider.example.test/openapi.json", "location": "Widgets > List",
			"operation_id": "provider.list-widgets", "method": method, "path": "/widgets",
			"binding": map[string]any{"kind": "command", "id": command}, "destructive_kind": "none",
		}},
	}
	declarations := map[string]any{
		"schema_version": 2, "cohort": "test-cohort",
		"declarations": []any{map[string]any{
			"connector": "cli-surface", "source_id": "list-widgets", "lane": lane, "command": command,
			"state":                   "unsupported_with_provider_evidence",
			"canonical":               map[string]any{"method": method, "path": "/widgets"},
			"binding":                 map[string]any{"kind": "command", "id": command},
			"unsupported_disposition": disposition,
		}},
	}
	writeDeclarationAdmissionJSON(t, filepath.Join(defsRoot, "declaration_admission_sources.json"), sources)
	writeDeclarationAdmissionJSON(t, filepath.Join(defsRoot, "declaration_admissions.json"), declarations)
	writeDeclarationAdmissionSourceLock(t, defsRoot, "cli-surface", []map[string]any{{
		"id": "provider.rest.widgets.list", "protocol": "rest", "method": method, "path": "/widgets",
		"operation_id": "provider.list-widgets", "deprecated": false, "source_location": "Widgets > List",
	}})
	return defsRoot
}

func writeDeclarationAdmissionSourceLock(t *testing.T, defsRoot, connector string, operations []map[string]any) {
	t.Helper()
	path := filepath.Join(defsRoot, connector, "sources", connector+"-operation-source-lock.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create source-lock fixture directory: %v", err)
	}
	lock := map[string]any{
		"schema_version": 2,
		"connector":      connector,
		"rest": map[string]any{
			"source_url": "https://provider.example.test/openapi.json",
			"sha256":     strings.Repeat("0", 64), "bytes": 1, "openapi": "3.0.3",
			"operations": operations,
		},
		"counts": map[string]any{
			"rest": len(operations), "graphql_query": 0, "graphql_mutation": 0, "total": len(operations),
		},
	}
	writeDeclarationAdmissionJSON(t, path, lock)
}

func readDeclarationAdmissionJSON(t *testing.T, path string, target any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func writeDeclarationAdmissionJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
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
	bundle.CLISurface.Commands[1].Foundation = &engine.CommandFoundation{
		ID: "write_plan_foundation", Reason: "write plan compiler is not available", Component: "runtime_executor", Evidence: "runtime_executor_absent",
		Target: engine.CommandFoundationTarget{Method: "POST", Path: "/v1/widgets"},
	}
	findings = checkCLISurfaceFoundation(bundle, 1, bundle.CLISurface.Commands[1])
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "requires its matching availability") {
		t.Fatalf("implemented foundation findings = %+v", findings)
	}
}

func TestDeclarationAdmissionDeferredCLISurfaceRequiresAdmissibleExactTarget(t *testing.T) {
	tests := []struct {
		name string
		edit func(*engine.Bundle)
	}{
		{
			name: "excluded",
			edit: func(bundle *engine.Bundle) {
				bundle.Surface.Endpoints[2].Excluded = &engine.SurfaceExclusion{Category: "destructive_admin", Reason: "policy exclusion"}
			},
		},
		{
			name: "disallowed",
			edit: func(bundle *engine.Bundle) {
				bundle.Surface.Endpoints[2].Operation.Model = "disallowed"
			},
		},
		{
			name: "duplicate",
			edit: func(bundle *engine.Bundle) {
				bundle.Surface.Endpoints = append(bundle.Surface.Endpoints, bundle.Surface.Endpoints[2])
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle, _ := declarationAdmissionFixture()
			testCase.edit(&bundle)
			findings := checkCLISurfaceFoundation(bundle, 2, bundle.CLISurface.Commands[2])
			if len(findings) != 1 || !strings.Contains(findings[0].Message, "deferred target") {
				t.Fatalf("deferred target findings = %+v, want exact-target refusal", findings)
			}
		})
	}
}

func TestDeclarationAdmissionNormalizesWriteActionTemplatePath(t *testing.T) {
	if got := declarationAdmissionActionPath("/v1/widgets/{{ record.widget_id }}"); got != "/v1/widgets/{widget_id}" {
		t.Fatalf("normalized action path = %q", got)
	}
}

func declarationAdmissionFixture() (engine.Bundle, declarationAdmissionDocument) {
	target := func(sourceID, bindingKind, bindingID, destructiveKind, method, path string) engine.CommandFoundationTarget {
		return engine.CommandFoundationTarget{
			SourceID: sourceID, ProviderOperationID: "provider." + sourceID,
			Binding:         engine.CommandBindingIdentity{Kind: bindingKind, ID: bindingID},
			DestructiveKind: destructiveKind, Method: method, Path: path,
		}
	}
	commands := []engine.CLICommand{
		{
			Path: "widgets list", Intent: "etl", Availability: declarationAdmissionStateImplemented,
			Stream: "widgets", APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/v1/widgets"}},
		},
		{
			Path: "widgets create", Intent: "reverse_etl", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/v1/widgets"}},
			Foundation: &engine.CommandFoundation{ID: "write_plan_foundation", Reason: "write plan compiler is not available", Component: "runtime_executor", Evidence: "runtime_executor_absent", Target: target("create-widget", "write", "create_widget", "none", "POST", "/v1/widgets")},
		},
		{
			Path: "widgets delete", Intent: "reverse_etl", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/v1/widgets/{id}"}},
			Foundation: &engine.CommandFoundation{ID: "delete_plan_foundation", Reason: "delete plan compiler is not available", Component: "runtime_executor", Evidence: "runtime_executor_absent", Target: target("delete-widget", "write", "delete_widget", "delete", "DELETE", "/v1/widgets/{id}")},
		},
		{
			Path: "widgets download", Intent: "binary_download", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/v1/widgets/{id}/archive"}},
			Foundation: &engine.CommandFoundation{ID: "binary_download_foundation", Reason: "binary response binding is not available", Component: "binary_transfer_binding", Evidence: "binary_transfer_binding_absent", Target: target("download-widget", "operation", "download_widget", "none", "GET", "/v1/widgets/{id}/archive")},
		},
		{
			Path: "widgets upload", Intent: "binary_upload", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/v1/widgets/{id}/archive"}},
			Foundation: &engine.CommandFoundation{ID: "binary_upload_foundation", Reason: "binary request binding is not available", Component: "binary_transfer_binding", Evidence: "binary_transfer_binding_absent", Target: target("upload-widget", "write", "upload_widget", "none", "POST", "/v1/widgets/{id}/archive")},
		},
		{
			Path: "widgets descriptor", Intent: "direct_write", Availability: declarationAdmissionStateDeferred,
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "PATCH", Path: "/v1/widgets/{id}"}},
			Foundation: &engine.CommandFoundation{ID: "source_descriptor_importer", Reason: "provider descriptor importer cannot yet represent this request", Component: "source_importer", Evidence: "source_importer_absent", Target: target("patch-widget", "operation", "patch_widget", "none", "PATCH", "/v1/widgets/{id}")},
		},
	}
	bundle := engine.Bundle{
		Name:    "acme",
		Streams: []engine.StreamSpec{{Name: "widgets", Method: "GET", Path: "/v1/widgets"}},
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{
			{Method: "GET", Path: "/v1/widgets"},
			{Method: "POST", Path: "/v1/widgets", Operation: &engine.SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", Risk: "medium", BlockedByDefault: true, Reason: "write runtime binding is pending"}},
			{Method: "DELETE", Path: "/v1/widgets/{id}", Operation: &engine.SurfaceOperation{Model: "destructive_action", Status: "blocked", Risk: "high", BlockedByDefault: true, Reason: "delete runtime binding is pending"}},
			{Method: "GET", Path: "/v1/widgets/{id}/archive", Operation: &engine.SurfaceOperation{Model: "binary_read", Status: "blocked", Risk: "medium", BlockedByDefault: true, Reason: "binary response binding is pending"}},
			{Method: "POST", Path: "/v1/widgets/{id}/archive", Operation: &engine.SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", Risk: "medium", BlockedByDefault: true, Reason: "binary upload binding is pending"}},
			{Method: "PATCH", Path: "/v1/widgets/{id}", Operation: &engine.SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", Risk: "medium", BlockedByDefault: true, Reason: "source importer is pending"}},
		}},
		CLISurface: &engine.CLISurface{Commands: commands},
	}
	rows := []struct {
		id, method, sourcePath, command, lane, state, foundationID, foundationReason, foundationComponent, foundationEvidence string
		bindingKind, bindingID, destructiveKind                                                                               string
		destructive                                                                                                           *declarationAdmissionDestructive
	}{
		{"list-widgets", "GET", "/widgets", "widgets list", declarationAdmissionLaneETL, declarationAdmissionStateImplemented, "", "", "", "", "stream", "widgets", "none", nil},
		{"create-widget", "POST", "/widgets", "widgets create", declarationAdmissionLaneReverseETL, declarationAdmissionStateDeferred, "write_plan_foundation", "write plan compiler is not available", "runtime_executor", "runtime_executor_absent", "write", "create_widget", "none", nil},
		{"delete-widget", "DELETE", "/widgets/{id}", "widgets delete", declarationAdmissionLaneReverseETL, declarationAdmissionStateDeferred, "delete_plan_foundation", "delete plan compiler is not available", "runtime_executor", "runtime_executor_absent", "write", "delete_widget", "delete", &declarationAdmissionDestructive{Kind: "delete", Reason: "provider operation deletes a widget"}},
		{"download-widget", "GET", "/widgets/{id}/archive", "widgets download", declarationAdmissionLaneBinaryDownload, declarationAdmissionStateDeferred, "binary_download_foundation", "binary response binding is not available", "binary_transfer_binding", "binary_transfer_binding_absent", "operation", "download_widget", "none", nil},
		{"upload-widget", "POST", "/widgets/{id}/archive", "widgets upload", declarationAdmissionLaneBinaryUpload, declarationAdmissionStateDeferred, "binary_upload_foundation", "binary request binding is not available", "binary_transfer_binding", "binary_transfer_binding_absent", "write", "upload_widget", "none", nil},
		{"patch-widget", "PATCH", "/widgets/{id}", "widgets descriptor", declarationAdmissionLaneDirectWrite, declarationAdmissionStateDeferred, "source_descriptor_importer", "provider descriptor importer cannot yet represent this request", "source_importer", "source_importer_absent", "operation", "patch_widget", "none", nil},
	}
	document := declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: "acme"}
	for _, row := range rows {
		document.SourceOperations = append(document.SourceOperations, declarationAdmissionSourceOperation{
			ID: row.id, SourceURL: "https://provider.example.test/v1/reference", Location: "Widgets > " + row.id, ProviderOperationID: "provider." + row.id,
			Protocol: "rest", Method: row.method, Path: row.sourcePath, BasePath: "/v1",
			Binding: declarationAdmissionBinding{Kind: row.bindingKind, ID: row.bindingID}, DestructiveKind: row.destructiveKind,
		})
		declaration := declarationAdmissionDeclaration{
			SourceID: row.id, Lane: row.lane, Command: row.command, State: row.state,
			Canonical: declarationAdmissionEndpoint{Method: row.method, Path: "/v1" + row.sourcePath},
			Binding:   declarationAdmissionBinding{Kind: row.bindingKind, ID: row.bindingID}, Destructive: row.destructive,
		}
		if row.foundationID != "" {
			declaration.Foundation = &declarationAdmissionFoundation{
				ID: row.foundationID, Reason: row.foundationReason, Component: row.foundationComponent, Evidence: row.foundationEvidence,
				Target: declarationAdmissionEndpoint{Method: row.method, Path: "/v1" + row.sourcePath},
			}
		}
		document.Declarations = append(document.Declarations, declaration)
	}
	return bundle, document
}
