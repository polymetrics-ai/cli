package engine

import (
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestResolveImplementedCommandBindingCoversEveryAdmissionRuntimeKind(t *testing.T) {
	bundle := Bundle{
		Name: "acme",
		Streams: []StreamSpec{{
			Name: "widgets", Method: http.MethodGet, Path: "/accounts/{{ config.account }}/widgets",
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
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/accounts/{account}/widgets"}},
			},
			wantBinding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: "widgets"},
			wantMethod:  http.MethodGet,
			wantPath:    "/accounts/{account}/widgets",
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
