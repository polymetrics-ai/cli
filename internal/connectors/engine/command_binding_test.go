package engine

import (
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestResolveImplementedCommandBindingCoversEveryAdmissionRuntimeKind(t *testing.T) {
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://api.example.test/v2/accounts/{{ config.account_id }}"},
		Streams: []StreamSpec{{
			Name: "widgets", Method: http.MethodGet, Path: "/widgets",
		}},
		Writes: []WriteAction{
			{Name: "create_widget", Kind: "create", Method: http.MethodPost, Path: "/widgets", RecordSchema: []byte(`{"type":"object"}`)},
			{Name: "upload_widget", Kind: "create", Method: http.MethodPut, Path: "/widgets/{{ record.id }}/content", BinaryUpload: &BinaryUploadSpec{SourceField: "file", MaxBytes: 1024}},
		},
		Operations: []OperationSpec{
			{ID: "acme.widget.get", Kind: "rest_read", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/widgets/{id}"}},
			{ID: "acme.widget.patch", Kind: "rest_write", REST: &RESTOperationSpec{Method: http.MethodPatch, Path: "/widgets/{id}"}},
			{ID: "acme.graphql.viewer", Kind: "graphql_query", GraphQL: &GraphQLOperationSpec{Path: "/graphql"}},
			{ID: "acme.graphql.create-widget", Kind: "graphql_mutation", GraphQL: &GraphQLOperationSpec{Path: "/graphql"}},
			{ID: "acme.widget.archive", Kind: "binary_download", Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/widgets/{id}/archive"}},
		},
	}

	tests := []struct {
		name        string
		command     connectors.CommandSurfaceCommand
		wantBinding connectors.CommandBindingIdentity
		wantMethod  string
		wantPath    string
	}{
		{
			name: "templated ETL stream",
			command: connectors.CommandSurfaceCommand{
				Path: "widget list", Intent: "etl", Availability: "implemented", Stream: "widgets",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/v2/accounts/{account_id}/widgets"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: "widgets"},
			wantMethod:  http.MethodGet,
			wantPath:    "/v2/accounts/{account_id}/widgets",
		},
		{
			name: "operation-free direct read",
			command: connectors.CommandSurfaceCommand{
				Path: "profile get", Intent: "direct_read", Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/profile"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingCommand, ID: "profile get"},
			wantMethod:  http.MethodGet,
			wantPath:    "/profile",
		},
		{
			name: "REST direct read",
			command: connectors.CommandSurfaceCommand{
				Path: "widget get", Intent: "direct_read", Availability: "implemented", Operation: "acme.widget.get",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/widgets/{id}"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.widget.get"},
			wantMethod:  http.MethodGet,
			wantPath:    "/widgets/{id}",
		},
		{
			name: "REST direct write",
			command: connectors.CommandSurfaceCommand{
				Path: "widget patch", Intent: "direct_write", Availability: "implemented", Operation: "acme.widget.patch",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPatch, Path: "/widgets/{id}"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.widget.patch"},
			wantMethod:  http.MethodPatch,
			wantPath:    "/widgets/{id}",
		},
		{
			name: "GraphQL direct read",
			command: connectors.CommandSurfaceCommand{
				Path: "viewer get", Intent: "direct_read", Availability: "implemented", Operation: "acme.graphql.viewer",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/graphql"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.graphql.viewer"},
			wantMethod:  http.MethodPost,
			wantPath:    "/graphql",
		},
		{
			name: "GraphQL direct write",
			command: connectors.CommandSurfaceCommand{
				Path: "widget graphql-create", Intent: "direct_write", Availability: "implemented", Operation: "acme.graphql.create-widget",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/graphql"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.graphql.create-widget"},
			wantMethod:  http.MethodPost,
			wantPath:    "/graphql",
		},
		{
			name: "reverse ETL write",
			command: connectors.CommandSurfaceCommand{
				Path: "widget create", Intent: "reverse_etl", Availability: "implemented", Write: "create_widget",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/widgets"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingWrite, ID: "create_widget"},
			wantMethod:  http.MethodPost,
			wantPath:    "/widgets",
		},
		{
			name: "binary download",
			command: connectors.CommandSurfaceCommand{
				Path: "widget download", Intent: "binary_download", Availability: "implemented", Operation: "acme.widget.archive",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/widgets/{id}/archive"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "acme.widget.archive"},
			wantMethod:  http.MethodGet,
			wantPath:    "/widgets/{id}/archive",
		},
		{
			name: "binary upload",
			command: connectors.CommandSurfaceCommand{
				Path: "widget upload", Intent: "binary_upload", Availability: "implemented", Write: "upload_widget",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPut, Path: "/widgets/{id}/content"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingWrite, ID: "upload_widget"},
			wantMethod:  http.MethodPut,
			wantPath:    "/widgets/{id}/content",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resolved, err := ResolveImplementedCommandBinding(bundle, testCase.command)
			if err != nil {
				t.Fatalf("ResolveImplementedCommandBinding: %v", err)
			}
			if resolved.Binding != testCase.wantBinding || resolved.Method != testCase.wantMethod || resolved.Path != testCase.wantPath {
				t.Fatalf("resolved = %+v, want binding=%+v method=%s path=%s", resolved, testCase.wantBinding, testCase.wantMethod, testCase.wantPath)
			}
		})
	}
}

func TestResolveImplementedCommandBindingRejectsUnprovenEndpointAliases(t *testing.T) {
	tests := []struct {
		name    string
		bundle  Bundle
		command connectors.CommandSurfaceCommand
	}{
		{
			name:   "stream path",
			bundle: Bundle{Name: "acme", Streams: []StreamSpec{{Name: "widgets", Method: http.MethodGet, Path: "/widgets"}}},
			command: connectors.CommandSurfaceCommand{Path: "widget list", Intent: "etl", Availability: "implemented", Stream: "widgets",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/unrelated"}}},
		},
		{
			name:   "write method",
			bundle: Bundle{Name: "acme", Writes: []WriteAction{{Name: "create_widget", Method: http.MethodPost, Path: "/widgets"}}},
			command: connectors.CommandSurfaceCommand{Path: "widget create", Intent: "reverse_etl", Availability: "implemented", Write: "create_widget",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodDelete, Path: "/widgets"}}},
		},
		{
			name: "REST operation path",
			bundle: Bundle{Name: "acme", Operations: []OperationSpec{{ID: "acme.widget.get", Kind: "rest_read",
				REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/widgets/{id}"}}}},
			command: connectors.CommandSurfaceCommand{Path: "widget get", Intent: "direct_read", Availability: "implemented", Operation: "acme.widget.get",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/accounts/{id}"}}},
		},
		{
			name: "binary operation method",
			bundle: Bundle{Name: "acme", Operations: []OperationSpec{{ID: "acme.widget.archive", Kind: "binary_download",
				Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/widgets/{id}/archive"}}}},
			command: connectors.CommandSurfaceCommand{Path: "widget download", Intent: "binary_download", Availability: "implemented", Operation: "acme.widget.archive",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/widgets/{id}/archive"}}},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if resolved, err := ResolveImplementedCommandBinding(testCase.bundle, testCase.command); err == nil {
				t.Fatalf("ResolveImplementedCommandBinding = %+v, nil error; want unproved alias refusal", resolved)
			}
		})
	}
}

func TestResolveImplementedCommandBindingNamesGraphQLStreamOperation(t *testing.T) {
	bundle := Bundle{Name: "acme", Streams: []StreamSpec{{
		Name: "widgets", Method: http.MethodPost, Path: "/graphql",
		GraphQL: &GraphQLRequestSpec{OperationName: "ListWidgets", Document: "query ListWidgets { widgets { id } }"},
	}}}
	command := connectors.CommandSurfaceCommand{
		Path: "widget list", Intent: "etl", Availability: "implemented", Stream: "widgets",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: "GRAPHQL", Path: "ListWidgets"}},
	}

	resolved, err := ResolveImplementedCommandBinding(bundle, command)
	if err != nil {
		t.Fatalf("ResolveImplementedCommandBinding: %v", err)
	}
	if resolved.Method != "GRAPHQL" || resolved.Path != "ListWidgets" {
		t.Fatalf("resolved = %+v, want canonical GRAPHQL ListWidgets", resolved)
	}
}

type testCommandBindingTransportHook struct{}

func (testCommandBindingTransportHook) ConnectorName() string { return "acme" }

func (testCommandBindingTransportHook) CommandBindingTransport(binding connectors.CommandBindingIdentity) (string, string, bool) {
	if binding == (connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: "widgets"}) {
		return http.MethodPost, "/search", true
	}
	return "", "", false
}

func TestResolveImplementedCommandBindingRequiresRegisteredHookTransportProof(t *testing.T) {
	bundle := Bundle{
		Name: "acme", HTTP: HTTPBase{URL: "https://api.example.test/v1"},
		Streams: []StreamSpec{{Name: "widgets", Path: "/search"}},
	}
	command := connectors.CommandSurfaceCommand{
		Path: "widget list", Intent: "etl", Availability: "implemented", Stream: "widgets",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/v1/search (object=widget)"}},
	}
	if _, err := resolveImplementedCommandBinding(bundle, command, nil); err == nil {
		t.Fatal("hook-routed command passed through its declarative fallback endpoint")
	}
	resolved, err := resolveImplementedCommandBinding(bundle, command, testCommandBindingTransportHook{})
	if err != nil {
		t.Fatalf("resolve registered hook transport: %v", err)
	}
	if resolved.Equivalence != CommandEndpointHookTransport || resolved.TransportMethod != http.MethodPost || resolved.TransportPath != "/search" {
		t.Fatalf("registered hook resolution = %+v", resolved)
	}
}

func TestResolveImplementedCommandBindingMapsGraphQLOperationToTransport(t *testing.T) {
	bundle := Bundle{Name: "acme", Operations: []OperationSpec{{
		ID: "acme.graphql.create-widget", Kind: "graphql_mutation",
		GraphQL: &GraphQLOperationSpec{OperationName: "CreateWidget", Path: "/graphql"},
	}}}
	command := connectors.CommandSurfaceCommand{
		Path: "widget create", Intent: "direct_write", Availability: "implemented", Operation: "acme.graphql.create-widget",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: "GRAPHQL", Path: "CreateWidget"}},
	}
	resolved, err := ResolveImplementedCommandBinding(bundle, command)
	if err != nil {
		t.Fatalf("resolve GraphQL operation transport: %v", err)
	}
	if resolved.Equivalence != CommandEndpointGraphQLTransport || resolved.TransportMethod != http.MethodPost || resolved.TransportPath != "/graphql" {
		t.Fatalf("GraphQL operation resolution = %+v", resolved)
	}
}
