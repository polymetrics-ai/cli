package engine

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestDeferredCommandRejectsStructurallyInvalidExactTargets(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "whitespace method", method: " POST", path: "/widgets"},
		{name: "control", method: "POST", path: "/widgets/\narchive"},
		{name: "absolute URL", method: "POST", path: "https://provider.example/widgets"},
		{name: "query", method: "POST", path: "/widgets?confirm=true"},
		{name: "fragment", method: "POST", path: "/widgets#archive"},
		{name: "dot segment", method: "POST", path: "/widgets/../archive"},
		{name: "encoded traversal", method: "POST", path: "/widgets/%2e%2e/archive"},
		{name: "backslash", method: "POST", path: `/widgets\archive`},
		{name: "repeated separator", method: "POST", path: "/widgets//archive"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := deferredAuditBundle(testCase.method, testCase.path)
			command := deferredAuditCommand(testCase.method, testCase.path)
			if err := PreflightDeferredCommand(bundle, command); err == nil {
				t.Fatal("invalid exact target reached deferred success")
			}
		})
	}
}

func TestDeferredCommandGraphQLOperationCannotHideAsRuntimeExecutorGap(t *testing.T) {
	bundle := deferredAuditBundle("POST", "/graphql")
	bundle.Operations = []OperationSpec{{
		ID: "acme.graphql.query.viewer", Kind: "graphql_query",
		GraphQL: &GraphQLOperationSpec{Path: "/graphql"},
	}}
	command := deferredAuditCommand("POST", "/graphql")
	command.Intent = "direct_read"
	command.Foundation.Component = connectors.FoundationComponentRuntimeExecutor
	command.Foundation.Evidence = "runtime_executor_absent"
	bundle.Surface.Endpoints[0].Operation.Model = "direct_read"

	err := PreflightDeferredCommand(bundle, command)
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("GraphQL runtime foundation = %v, want stale rejection", err)
	}
}

func TestDeferredCommandGraphQLSharedTransportCannotSwapOperationIdentity(t *testing.T) {
	bundle := deferredAuditBundle("POST", "/graphql")
	bundle.Operations = []OperationSpec{{
		ID: "acme.graphql.query.viewer", Kind: "graphql_query",
		GraphQL: &GraphQLOperationSpec{Path: "/graphql"},
	}}
	command := deferredAuditCommand("POST", "/graphql")
	command.Intent = "direct_read"
	command.Foundation.Component = connectors.FoundationComponentRuntimeExecutor
	command.Foundation.Evidence = "runtime_executor_absent"
	command.Foundation.Target = connectors.CommandFoundationTarget{
		SourceID: "acme.graphql.query.organization", ProviderOperationID: "organization",
		Binding:         connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.graphql.query.organization"},
		DestructiveKind: "none", Method: "POST", Path: "/graphql",
	}
	bundle.Surface.Endpoints[0].Operation.Model = "direct_read"

	err := PreflightDeferredCommand(bundle, command)
	if err == nil || !strings.Contains(err.Error(), "different runtime binding") {
		t.Fatalf("GraphQL shared-transport swap = %v, want binding-identity rejection", err)
	}
}

func deferredAuditBundle(method, path string) Bundle {
	return Bundle{
		Name: "acme",
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: method,
			Path:   path,
			Operation: &SurfaceOperation{
				Model: "sensitive_reverse_etl", Status: "blocked", Risk: "medium",
				BlockedByDefault: true, Reason: "foundation pending",
			},
		}}},
	}
}

func deferredAuditCommand(method, path string) connectors.CommandSurfaceCommand {
	return connectors.CommandSurfaceCommand{
		Path: "widgets archive", Intent: "reverse_etl", Availability: "deferred",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: method, Path: path}},
		Foundation: &connectors.CommandFoundation{
			ID: "runtime_executor", Reason: "runtime executor is pending",
			Component: connectors.FoundationComponentRuntimeExecutor, Evidence: "runtime_executor_absent",
			Target: connectors.CommandFoundationTarget{Method: method, Path: path},
		},
	}
}
