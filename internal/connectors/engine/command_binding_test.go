package engine

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
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

// TestResolveImplementedCommandBindingProvesEveryCircleCICompositeProjectSlugBinding
// reproduces the eleven source-backed Batch-1 bindings that cannot pass the
// ordinary one-slot placeholder proof: CircleCI documents one project-slug
// identity while the existing transport requires its vcs/org/repo components.
func TestResolveImplementedCommandBindingProvesEveryCircleCICompositeProjectSlugBinding(t *testing.T) {
	bundle := circleCICompositeProjectSlugBundle()
	identity := bundle.CompositeProviderPathIdentity
	if identity == nil || len(identity.Bindings) != 11 {
		t.Fatalf("composite identity = %+v, want eleven declared source bindings", identity)
	}

	for _, binding := range identity.Bindings {
		binding := binding
		t.Run(binding.SourceID, func(t *testing.T) {
			resolved, err := ResolveImplementedCommandBinding(bundle, compositeProviderPathCommand(binding))
			if err != nil {
				t.Fatalf("ResolveImplementedCommandBinding: %v", err)
			}
			if resolved.Equivalence != CommandEndpointCompositeProviderPathIdentity {
				t.Fatalf("equivalence = %q, want %q", resolved.Equivalence, CommandEndpointCompositeProviderPathIdentity)
			}
		})
	}
}

// The composite proof must stay closed. These cases cover the places a future
// generic placeholder feature might accidentally widen it: identity metadata,
// connector, lane, binding, method, and every relevant transport-path shape.
func TestResolveImplementedCommandBindingRejectsUnprovenCircleCICompositeProjectSlugSubstitutions(t *testing.T) {
	base := circleCICompositeProjectSlugBundle()
	if base.CompositeProviderPathIdentity == nil {
		t.Fatal("composite identity is required for the rejection matrix")
	}
	baseCommand := compositeProviderPathCommand(base.CompositeProviderPathIdentity.Bindings[0])
	binaryUploadCommand := compositeProviderPathCommand(base.CompositeProviderPathIdentity.Bindings[6])
	binaryUploadCommand.Intent = "binary_upload"
	tests := []struct {
		name    string
		bundle  func() Bundle
		command connectors.CommandSurfaceCommand
	}{
		{
			name: "missing identity configuration",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity = nil
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "cross connector declaration",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Name = "acme"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "invalid source digest",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.SourceSHA256 = "not-a-sha256"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "non HTTPS source URL",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.SourceURL = "file:///tmp/openapi.json"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "reordered config keys",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.ConfigKeys = []string{"org", "vcs_type", "repo"}
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "partial config keys",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.ConfigKeys = []string{"vcs_type", "org"}
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "extra config key",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.ConfigKeys = []string{"vcs_type", "org", "repo", "project"}
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "repeated config key",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.ConfigKeys = []string{"vcs_type", "org", "org"}
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "reordered source binding",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bindings := bundle.CompositeProviderPathIdentity.Bindings
				bindings[0], bindings[1] = bindings[1], bindings[0]
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "partial runtime path",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path = "/project/{{ config.vcs_type }}/{{ config.org }}"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "reordered runtime components",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path = "/project/{{ config.org }}/{{ config.vcs_type }}/{{ config.repo }}"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "extra runtime component",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path = "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/extra"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "changed runtime literal",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path = "/projects/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "absolute runtime path",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path = "https://circleci.com/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "query runtime path",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Path += "?page-token={token}"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "route transport",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Route = "alternate"
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "wrong method",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Streams[0].Method = http.MethodPost
				return bundle
			},
			command: baseCommand,
		},
		{
			name: "direct read operation",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Operations = []OperationSpec{{ID: "circleci.project.get", Kind: "rest_read", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"}}}
				return bundle
			},
			command: connectors.CommandSurfaceCommand{Path: "project get", Intent: "direct_read", Availability: "implemented", Operation: "circleci.project.get", APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/project/{project-slug}"}}},
		},
		{
			name: "direct write operation",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Operations = []OperationSpec{{ID: "circleci.project.put", Kind: "rest_write", REST: &RESTOperationSpec{Method: http.MethodPut, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"}}}
				return bundle
			},
			command: connectors.CommandSurfaceCommand{Path: "project update", Intent: "direct_write", Availability: "implemented", Operation: "circleci.project.put", APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPut, Path: "/project/{project-slug}"}}},
		},
		{
			name: "binary download operation",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.Operations = []OperationSpec{{ID: "circleci.project.download", Kind: "binary_download", Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"}}}
				return bundle
			},
			command: connectors.CommandSurfaceCommand{Path: "project download", Intent: "binary_download", Availability: "implemented", Operation: "circleci.project.download", APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/project/{project-slug}"}}},
		},
		{
			name:    "binary upload write",
			bundle:  circleCICompositeProjectSlugBundle,
			command: binaryUploadCommand,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if resolved, err := ResolveImplementedCommandBinding(testCase.bundle(), testCase.command); err == nil {
				t.Fatalf("ResolveImplementedCommandBinding = %+v, nil error; want closed composite-proof refusal", resolved)
			}
		})
	}
	hook := compositeProviderPathHook{
		binding: connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: base.CompositeProviderPathIdentity.Bindings[0].BindingID},
		method:  base.CompositeProviderPathIdentity.Bindings[0].Method,
		path: strings.Replace(
			base.CompositeProviderPathIdentity.Bindings[0].Path,
			"{"+base.CompositeProviderPathIdentity.Placeholder+"}",
			compositeProviderPathRuntimeSegments(base.CompositeProviderPathIdentity.ConfigKeys),
			1,
		),
	}
	if _, err := resolveImplementedCommandBinding(circleCICompositeProjectSlugBundle(), baseCommand, hook); err == nil {
		t.Fatal("hook-routed CircleCI transport passed the closed composite identity proof")
	}
}

type compositeProviderPathHook struct {
	binding connectors.CommandBindingIdentity
	method  string
	path    string
}

func (compositeProviderPathHook) ConnectorName() string { return "test" }

func (hook compositeProviderPathHook) CommandBindingTransport(binding connectors.CommandBindingIdentity) (string, string, bool) {
	if binding == hook.binding {
		return hook.method, hook.path, true
	}
	return "", "", false
}

func TestCircleCICompositeProviderPathIdentityLoadsFromTheSourceCitedSurface(t *testing.T) {
	bundle, err := Load(defs.FS, "circleci")
	if err != nil {
		t.Fatalf("Load(defs.FS, circleci): %v", err)
	}
	identity := bundle.CompositeProviderPathIdentity
	if identity == nil {
		t.Fatal("CircleCI composite_provider_path_identity.json omitted the closed source identity")
	}
	if len(identity.Bindings) != 11 || identity.Connector != bundle.Name {
		t.Fatalf("CircleCI composite identity = %+v, want its declaration-owned connector and eleven bindings", identity)
	}
}

func TestCompositeProviderPathIdentityOwnsItsConnectorIdentity(t *testing.T) {
	bundle, err := Load(defs.FS, "circleci")
	if err != nil {
		t.Fatalf("Load(defs.FS, circleci): %v", err)
	}
	identity := bundle.CompositeProviderPathIdentity
	if identity == nil {
		t.Fatal("composite_provider_path_identity.json omitted the closed source identity")
	}
	if identity.Connector != bundle.Name {
		t.Fatalf("identity connector = %q, want bundle connector %q", identity.Connector, bundle.Name)
	}
}

func TestBundleLoadRejectsCompositeProviderPathIdentityForAnotherConnector(t *testing.T) {
	fSys := fullValidBundleFS("acme")
	identity := circleCICompositeProjectSlugBundle().CompositeProviderPathIdentity
	raw, err := json.Marshal(identity)
	if err != nil {
		t.Fatalf("json.Marshal(identity): %v", err)
	}
	fSys["acme/"+compositeProviderPathIdentityFile] = &fstest.MapFile{Data: raw}
	if _, err := Load(fSys, "acme"); err == nil || !strings.Contains(err.Error(), "does not match bundle") {
		t.Fatalf("Load cross-connector composite identity error = %v, want closed connector rejection", err)
	}
}

func circleCICompositeProjectSlugBundle() Bundle {
	bundle, err := Load(defs.FS, "circleci")
	if err != nil {
		panic(err)
	}
	return bundle
}

func compositeProviderPathCommand(binding CompositeProviderPathBinding) connectors.CommandSurfaceCommand {
	command := connectors.CommandSurfaceCommand{
		Path:         binding.SourceID,
		Intent:       binding.Intent,
		Availability: "implemented",
		APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: binding.Method, Path: binding.Path}},
	}
	if binding.BindingKind == "stream" {
		command.Stream = binding.BindingID
	} else {
		command.Write = binding.BindingID
	}
	return command
}
