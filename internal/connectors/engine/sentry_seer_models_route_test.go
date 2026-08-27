package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

const (
	sentrySeerModelsCommand         = "seer list-models"
	sentrySeerModelsOperation       = "sentry.seer_models_list"
	sentrySeerModelsSourceOperation = "sentry.rest.listSeerModels"
	sentrySeerModelsPath            = "/api/0/seer/models/"
	sentrySeerModelsRoute           = "sentry_api_v0"
	sentrySeerModelsSourceURL       = "https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json"
)

func TestSentrySeerModelsSourceBoundRoute(t *testing.T) {
	bundle := loadSentrySeerModelsBundle(t)
	surface := synthesizeCommandSurface(bundle)
	if surface == nil {
		t.Fatal("sentry command surface is missing")
	}

	var commands []connectors.CommandSurfaceCommand
	for _, command := range surface.Commands {
		if command.Path == sentrySeerModelsCommand {
			commands = append(commands, command)
		}
	}
	if len(commands) != 1 {
		t.Fatalf("%q command count = %d, want exactly one", sentrySeerModelsCommand, len(commands))
	}
	command := commands[0]
	if command.Intent != "direct_read" || command.Availability != "implemented" || command.Operation != sentrySeerModelsOperation || command.SourceOperation != sentrySeerModelsSourceOperation {
		t.Fatalf("command = %+v, want the one implemented source-bound Seer Models direct read", command)
	}
	if len(command.Flags) != 0 {
		t.Fatalf("command flags = %+v, want no caller-supplied provider route, base, method, or path override", command.Flags)
	}
	if len(command.APISurface) != 1 || command.APISurface[0].Method != http.MethodGet || command.APISurface[0].Path != sentrySeerModelsPath {
		t.Fatalf("command api_surface = %+v, want GET %s", command.APISurface, sentrySeerModelsPath)
	}

	routeCount := 0
	for _, route := range bundle.HTTP.Routes {
		if route.Name != sentrySeerModelsRoute {
			continue
		}
		routeCount++
		if route.BaseURL != "{{ config.base_url }}" || route.Version != "api" {
			t.Fatalf("route = %+v, want the declared Sentry base identity with api version", route)
		}
	}
	if routeCount != 1 {
		t.Fatalf("%q route count = %d, want one declared Sentry route identity", sentrySeerModelsRoute, routeCount)
	}

	operation := sentrySeerModelsOperationSpec(t, bundle)
	if operation.Route != sentrySeerModelsRoute || operation.SourceURL != sentrySeerModelsSourceURL {
		t.Fatalf("operation route/source = %q/%q, want %q/%q", operation.Route, operation.SourceURL, sentrySeerModelsRoute, sentrySeerModelsSourceURL)
	}
	if operation.SourceOperation == nil || operation.SourceOperation.ID != sentrySeerModelsSourceOperation || operation.SourceOperation.Method != http.MethodGet || operation.SourceOperation.Path != sentrySeerModelsPath {
		t.Fatalf("operation source binding = %+v, want exact Seer Models source identity", operation.SourceOperation)
	}
	if operation.REST == nil || operation.REST.Method != http.MethodGet || operation.REST.Path != sentrySeerModelsPath {
		t.Fatalf("operation REST binding = %+v, want GET %s", operation.REST, sentrySeerModelsPath)
	}

	resolved, err := ResolveImplementedCommandPath(bundle, sentrySeerModelsCommand)
	if err != nil {
		t.Fatalf("ResolveImplementedCommandPath(%q): %v", sentrySeerModelsCommand, err)
	}
	if resolved.Binding.Kind != connectors.CommandBindingOperation || resolved.Binding.ID != sentrySeerModelsOperation || resolved.Method != http.MethodGet || resolved.Path != sentrySeerModelsPath || resolved.TransportMethod != http.MethodGet || resolved.TransportPath != sentrySeerModelsPath {
		t.Fatalf("resolved command binding = %+v, want exact Seer Models operation endpoint", resolved)
	}

	entries := OperationDirectReadEndpointLedgerEntries(bundle)
	if len(entries) != 1 || entries[0] != (OperationEndpointLedgerEntry{Method: http.MethodGet, Path: sentrySeerModelsPath, Kind: "rest_read", MaxBytes: 1048576}) {
		t.Fatalf("Sentry endpoint ledger = %+v, want one exact Seer Models entry", entries)
	}
}

func TestSentrySeerModelsRouteRejectsIdentityDriftBeforeProviderIO(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*Bundle, *connectors.CommandSurfaceCommand, *OperationSpec)
		check  func(Bundle, connectors.CommandSurfaceCommand, OperationSpec) error
	}{
		{
			name: "source id",
			mutate: func(_ *Bundle, command *connectors.CommandSurfaceCommand, _ *OperationSpec) {
				command.SourceOperation = "sentry.rest.notSeerModels"
			},
			check: func(bundle Bundle, command connectors.CommandSurfaceCommand, _ OperationSpec) error {
				return PreflightSourceBoundRead(bundle, command.Operation, command.SourceOperation, command.APISurface[0].Method, command.APISurface[0].Path)
			},
		},
		{
			name: "route",
			mutate: func(_ *Bundle, _ *connectors.CommandSurfaceCommand, operation *OperationSpec) {
				operation.Route = "untrusted_route"
			},
			check: func(bundle Bundle, _ connectors.CommandSurfaceCommand, operation OperationSpec) error {
				return preflightSourceBoundOperationOrigin(bundle, connectors.RuntimeConfig{}, operation)
			},
		},
		{
			name:   "caller base",
			mutate: func(_ *Bundle, _ *connectors.CommandSurfaceCommand, _ *OperationSpec) {},
			check: func(bundle Bundle, _ connectors.CommandSurfaceCommand, operation OperationSpec) error {
				return preflightSourceBoundOperationOrigin(bundle, connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://untrusted.example"}}, operation)
			},
		},
		{
			name: "method",
			mutate: func(_ *Bundle, command *connectors.CommandSurfaceCommand, _ *OperationSpec) {
				command.APISurface[0].Method = http.MethodPost
			},
			check: func(bundle Bundle, command connectors.CommandSurfaceCommand, _ OperationSpec) error {
				return PreflightSourceBoundRead(bundle, command.Operation, command.SourceOperation, command.APISurface[0].Method, command.APISurface[0].Path)
			},
		},
		{
			name: "path",
			mutate: func(_ *Bundle, command *connectors.CommandSurfaceCommand, _ *OperationSpec) {
				command.APISurface[0].Path = "/api/0/seer/other-models/"
			},
			check: func(bundle Bundle, command connectors.CommandSurfaceCommand, _ OperationSpec) error {
				return PreflightSourceBoundRead(bundle, command.Operation, command.SourceOperation, command.APISurface[0].Method, command.APISurface[0].Path)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := loadSentrySeerModelsBundle(t)
			surface := synthesizeCommandSurface(bundle)
			command := sentrySeerModelsCommandSpec(t, surface)
			operation := sentrySeerModelsOperationSpec(t, bundle)
			testCase.mutate(&bundle, &command, &operation)
			if err := testCase.check(bundle, command, operation); err == nil {
				t.Fatal("closed Sentry source/route/base binding accepted drift")
			}
		})
	}
}

func TestSentrySeerModelsRoutePreservesPathAcrossBaseSlashForms(t *testing.T) {
	for _, trailingSlash := range []bool{false, true} {
		t.Run(map[bool]string{false: "base without trailing slash", true: "base with trailing slash"}[trailingSlash], func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				if request.Method != http.MethodGet || request.URL.Path != sentrySeerModelsPath {
					t.Fatalf("provider request = %s %s, want GET %s", request.Method, request.URL.Path, sentrySeerModelsPath)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
			}))
			defer server.Close()

			bundle := loadSentrySeerModelsBundle(t)
			for index := range bundle.HTTP.Routes {
				if bundle.HTTP.Routes[index].Name == sentrySeerModelsRoute {
					bundle.HTTP.Routes[index].BaseURL = server.URL
					if trailingSlash {
						bundle.HTTP.Routes[index].BaseURL += "/"
					}
				}
			}
			if _, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation: sentrySeerModelsOperation,
				Config:    connectors.RuntimeConfig{Secrets: map[string]string{"auth_token": "test-token"}},
			}, nil); err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			if got := requests.Load(); got != 1 {
				t.Fatalf("provider requests = %d, want one", got)
			}
		})
	}
}

func loadSentrySeerModelsBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := Load(defs.FS, "sentry")
	if err != nil {
		t.Fatalf("Load(sentry): %v", err)
	}
	return bundle
}

func sentrySeerModelsCommandSpec(t *testing.T, surface *connectors.CommandSurface) connectors.CommandSurfaceCommand {
	t.Helper()
	if surface == nil {
		t.Fatal("sentry command surface is missing")
	}
	for _, command := range surface.Commands {
		if command.Path == sentrySeerModelsCommand {
			return command
		}
	}
	t.Fatalf("Sentry command surface does not declare %q", sentrySeerModelsCommand)
	return connectors.CommandSurfaceCommand{}
}

func sentrySeerModelsOperationSpec(t *testing.T, bundle Bundle) OperationSpec {
	t.Helper()
	for _, operation := range bundle.Operations {
		if operation.ID == sentrySeerModelsOperation {
			return operation
		}
	}
	t.Fatalf("Sentry bundle does not declare %q", sentrySeerModelsOperation)
	return OperationSpec{}
}

func TestSentrySeerModelsRouteTestConstantsAreProviderRelative(t *testing.T) {
	if !strings.HasPrefix(sentrySeerModelsPath, "/") || strings.Contains(sentrySeerModelsPath, "://") {
		t.Fatalf("source path = %q, want connector-relative provider path", sentrySeerModelsPath)
	}
}
