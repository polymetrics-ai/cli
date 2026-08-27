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

	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
	}{
		{name: "projects stream", command: circleCICompositeCommand("projects list", "etl", "projects", "", http.MethodGet, "/project/{project-slug}")},
		{name: "pipelines stream", command: circleCICompositeCommand("pipelines list", "etl", "pipelines", "", http.MethodGet, "/project/{project-slug}/pipeline")},
		{name: "schedules stream", command: circleCICompositeCommand("schedules list", "etl", "schedules", "", http.MethodGet, "/project/{project-slug}/schedule")},
		{name: "checkout keys stream", command: circleCICompositeCommand("checkout keys list", "etl", "checkout_keys", "", http.MethodGet, "/project/{project-slug}/checkout-key")},
		{name: "environment variables stream", command: circleCICompositeCommand("environment variables list", "etl", "environment_variables", "", http.MethodGet, "/project/{project-slug}/envvar")},
		{name: "insights workflow summary stream", command: circleCICompositeCommand("insights workflow summary list", "etl", "insights_workflow_summary", "", http.MethodGet, "/insights/{project-slug}/workflows")},
		{name: "create schedule", command: circleCICompositeCommand("create schedule apply", "reverse_etl", "", "create_schedule", http.MethodPost, "/project/{project-slug}/schedule")},
		{name: "create environment variable", command: circleCICompositeCommand("create environment variable apply", "reverse_etl", "", "create_environment_variable", http.MethodPost, "/project/{project-slug}/envvar")},
		{name: "create checkout key", command: circleCICompositeCommand("create checkout key apply", "reverse_etl", "", "create_checkout_key", http.MethodPost, "/project/{project-slug}/checkout-key")},
		{name: "delete environment variable", command: circleCICompositeCommand("delete environment variable apply", "reverse_etl", "", "delete_environment_variable", http.MethodDelete, "/project/{project-slug}/envvar/{name}")},
		{name: "delete checkout key", command: circleCICompositeCommand("delete checkout key apply", "reverse_etl", "", "delete_checkout_key", http.MethodDelete, "/project/{project-slug}/checkout-key/{fingerprint}")},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resolved, err := ResolveImplementedCommandBinding(bundle, testCase.command)
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
	baseCommand := circleCICompositeCommand("projects list", "etl", "projects", "", http.MethodGet, "/project/{project-slug}")
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
			name: "unretained source digest",
			bundle: func() Bundle {
				bundle := circleCICompositeProjectSlugBundle()
				bundle.CompositeProviderPathIdentity.SourceSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
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
			command: circleCICompositeCommand("schedule upload", "binary_upload", "", "create_schedule", http.MethodPost, "/project/{project-slug}/schedule"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if resolved, err := ResolveImplementedCommandBinding(testCase.bundle(), testCase.command); err == nil {
				t.Fatalf("ResolveImplementedCommandBinding = %+v, nil error; want closed composite-proof refusal", resolved)
			}
		})
	}
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
	if len(identity.Bindings) != 11 || identity.SourceSHA256 != circleCICompositeProviderPathSourceSHA {
		t.Fatalf("CircleCI composite identity = %+v, want retained digest and eleven bindings", identity)
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
	if _, err := Load(fSys, "acme"); err == nil || !strings.Contains(err.Error(), "only defined for circleci") {
		t.Fatalf("Load cross-connector composite identity error = %v, want closed connector rejection", err)
	}
}

func circleCICompositeCommand(path, intent, stream, write, method, endpoint string) connectors.CommandSurfaceCommand {
	return connectors.CommandSurfaceCommand{
		Path: path, Intent: intent, Availability: "implemented", Stream: stream, Write: write,
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: method, Path: endpoint}},
	}
}

func circleCICompositeProjectSlugBundle() Bundle {
	return Bundle{
		Name: "circleci",
		CompositeProviderPathIdentity: &CompositeProviderPathIdentity{
			SourceURL:    "https://circleci.com/api/v2/openapi.json",
			SourceSHA256: "61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07",
			Placeholder:  "project-slug",
			ConfigKeys:   []string{"vcs_type", "org", "repo"},
			Bindings: []CompositeProviderPathBinding{
				{SourceID: "circleci.rest.getProjectBySlug", ProviderOperationID: "getProjectBySlug", SourceLocation: `paths["/project/{project-slug}"].get`, Intent: "etl", BindingKind: "stream", BindingID: "projects", Method: http.MethodGet, Path: "/project/{project-slug}"},
				{SourceID: "circleci.rest.listPipelinesForProject", ProviderOperationID: "listPipelinesForProject", SourceLocation: `paths["/project/{project-slug}/pipeline"].get`, Intent: "etl", BindingKind: "stream", BindingID: "pipelines", Method: http.MethodGet, Path: "/project/{project-slug}/pipeline"},
				{SourceID: "circleci.rest.listSchedulesForProject", ProviderOperationID: "listSchedulesForProject", SourceLocation: `paths["/project/{project-slug}/schedule"].get`, Intent: "etl", BindingKind: "stream", BindingID: "schedules", Method: http.MethodGet, Path: "/project/{project-slug}/schedule"},
				{SourceID: "circleci.rest.listCheckoutKeys", ProviderOperationID: "listCheckoutKeys", SourceLocation: `paths["/project/{project-slug}/checkout-key"].get`, Intent: "etl", BindingKind: "stream", BindingID: "checkout_keys", Method: http.MethodGet, Path: "/project/{project-slug}/checkout-key"},
				{SourceID: "circleci.rest.listEnvVars", ProviderOperationID: "listEnvVars", SourceLocation: `paths["/project/{project-slug}/envvar"].get`, Intent: "etl", BindingKind: "stream", BindingID: "environment_variables", Method: http.MethodGet, Path: "/project/{project-slug}/envvar"},
				{SourceID: "circleci.rest.getProjectWorkflowMetrics", ProviderOperationID: "getProjectWorkflowMetrics", SourceLocation: `paths["/insights/{project-slug}/workflows"].get`, Intent: "etl", BindingKind: "stream", BindingID: "insights_workflow_summary", Method: http.MethodGet, Path: "/insights/{project-slug}/workflows"},
				{SourceID: "circleci.rest.createSchedule", ProviderOperationID: "createSchedule", SourceLocation: `paths["/project/{project-slug}/schedule"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_schedule", Method: http.MethodPost, Path: "/project/{project-slug}/schedule"},
				{SourceID: "circleci.rest.createEnvVar", ProviderOperationID: "createEnvVar", SourceLocation: `paths["/project/{project-slug}/envvar"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_environment_variable", Method: http.MethodPost, Path: "/project/{project-slug}/envvar"},
				{SourceID: "circleci.rest.createCheckoutKey", ProviderOperationID: "createCheckoutKey", SourceLocation: `paths["/project/{project-slug}/checkout-key"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_checkout_key", Method: http.MethodPost, Path: "/project/{project-slug}/checkout-key"},
				{SourceID: "circleci.rest.deleteEnvVar", ProviderOperationID: "deleteEnvVar", SourceLocation: `paths["/project/{project-slug}/envvar/{name}"].delete`, Intent: "reverse_etl", BindingKind: "write", BindingID: "delete_environment_variable", Method: http.MethodDelete, Path: "/project/{project-slug}/envvar/{name}"},
				{SourceID: "circleci.rest.deleteCheckoutKey", ProviderOperationID: "deleteCheckoutKey", SourceLocation: `paths["/project/{project-slug}/checkout-key/{fingerprint}"].delete`, Intent: "reverse_etl", BindingKind: "write", BindingID: "delete_checkout_key", Method: http.MethodDelete, Path: "/project/{project-slug}/checkout-key/{fingerprint}"},
			},
		},
		Streams: []StreamSpec{
			{Name: "projects", Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}"},
			{Name: "pipelines", Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/pipeline"},
			{Name: "schedules", Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule"},
			{Name: "checkout_keys", Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key"},
			{Name: "environment_variables", Method: http.MethodGet, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar"},
			{Name: "insights_workflow_summary", Method: http.MethodGet, Path: "/insights/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/workflows"},
		},
		Writes: []WriteAction{
			{Name: "create_schedule", Method: http.MethodPost, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule"},
			{Name: "create_environment_variable", Method: http.MethodPost, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar"},
			{Name: "create_checkout_key", Method: http.MethodPost, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key"},
			{Name: "delete_environment_variable", Method: http.MethodDelete, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar/{{ record.name }}"},
			{Name: "delete_checkout_key", Method: http.MethodDelete, Path: "/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key/{{ record.fingerprint }}"},
		},
	}
}
