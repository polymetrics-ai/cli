package commandrunner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type fakeConnector struct {
	surface       *connectors.CommandSurface
	manifest      connectors.Manifest
	readReq       connectors.ReadRequest
	directReadReq connectors.DirectReadRequest
	validateReq   connectors.WriteRequest
	dryRunReq     connectors.WriteRequest
	writeReq      connectors.WriteRequest
	writeRecords  []connectors.Record
	validateErr   error
	dryRunErr     error
	writeErr      error
	preview       connectors.WritePreview
	writeResult   connectors.WriteResult
}

func (f *fakeConnector) Name() string { return "github" }
func (f *fakeConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "github", DisplayName: "GitHub"}
}
func (f *fakeConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (f *fakeConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}
func (f *fakeConnector) Read(_ context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	f.readReq = req
	return emit(connectors.Record{"number": 101, "state": req.Query["state"]})
}
func (f *fakeConnector) DirectRead(_ context.Context, req connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	f.directReadReq = req
	return connectors.DirectReadResult{
		Connector: "github",
		Method:    req.Method,
		Path:      "/repos/octo/hello/contents/README.md",
		Status:    200,
		Body: map[string]any{
			"name": "README.md",
			"type": "file",
		},
	}, nil
}
func (f *fakeConnector) Write(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	f.writeReq = req
	f.writeRecords = append([]connectors.Record(nil), records...)
	if f.writeErr != nil {
		return connectors.WriteResult{}, f.writeErr
	}
	if f.writeResult != (connectors.WriteResult{}) {
		return f.writeResult, nil
	}
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}
func (f *fakeConnector) CommandSurface() *connectors.CommandSurface { return f.surface }
func (f *fakeConnector) Manifest() connectors.Manifest              { return f.manifest }
func (f *fakeConnector) ValidateWrite(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) error {
	f.validateReq = req
	return f.validateErr
}
func (f *fakeConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	f.dryRunReq = req
	if f.dryRunErr != nil {
		return connectors.WritePreview{}, f.dryRunErr
	}
	if f.preview.Action != "" || f.preview.RecordsStaged != 0 || len(f.preview.Warnings) > 0 {
		return f.preview, nil
	}
	return connectors.WritePreview{Action: req.Action, RecordsStaged: len(records)}, nil
}

func TestRunImplementedStreamCommand(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "issue list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "issues",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed", "all"}, MapsTo: "query.state"},
				},
			},
		},
	}}

	var records []connectors.Record
	result, err := Run(context.Background(), connector, Request{
		Path:   []string{"issue", "list"},
		Flags:  map[string][]string{"state": []string{"closed"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{"owner": "octocat", "repo": "hello-world"}},
		Limit:  1,
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Command != "issue list" || result.Stream != "issues" || result.Count != 1 {
		t.Fatalf("result = %+v, want command issue list stream issues count 1", result)
	}
	if connector.readReq.Stream != "issues" {
		t.Fatalf("read stream = %q, want issues", connector.readReq.Stream)
	}
	if got := connector.readReq.Query["state"]; got != "closed" {
		t.Fatalf("read query state = %q, want closed", got)
	}
	if len(records) != 1 || records[0]["state"] != "closed" {
		t.Fatalf("records = %+v, want one closed record", records)
	}
}

func TestRunCoreStreamMappings(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{Path: "pr list", Intent: "etl", Availability: "implemented", Stream: "pull_requests"},
			{Path: "release list", Intent: "etl", Availability: "implemented", Stream: "releases"},
			{Path: "workflow list", Intent: "etl", Availability: "implemented", Stream: "workflows"},
		},
	}}

	tests := []struct {
		name   string
		path   []string
		stream string
	}{
		{name: "pull requests", path: []string{"pr", "list"}, stream: "pull_requests"},
		{name: "releases", path: []string{"release", "list"}, stream: "releases"},
		{name: "workflows", path: []string{"workflow", "list"}, stream: "workflows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var records []connectors.Record
			result, err := Run(context.Background(), connector, Request{Path: tt.path, Limit: 1}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.Stream != tt.stream {
				t.Fatalf("stream = %q, want %q", result.Stream, tt.stream)
			}
			if connector.readReq.Stream != tt.stream {
				t.Fatalf("read stream = %q, want %q", connector.readReq.Stream, tt.stream)
			}
			if len(records) != 1 {
				t.Fatalf("records = %d, want 1", len(records))
			}
		})
	}
}

func TestRunBlocksNonStreamCommands(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo clone",
				Intent:       "local_workflow",
				Availability: "unsupported_local",
				Notes:        "local git clone workflow",
			},
		},
	}}

	tests := []struct {
		name string
		path []string
		want string
	}{
		{name: "local_workflow", path: []string{"repo", "clone"}, want: "unsupported_local"},
		{name: "unknown", path: []string{"issue", "frobnicate"}, want: "unknown command"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{Path: tt.path}, func(connectors.Record) error {
				t.Fatal("emit called for blocked command")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want blocker")
			}
			var blocked *BlockedCommandError
			if !errors.As(err, &blocked) {
				t.Fatalf("Run error type = %T, want BlockedCommandError", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestBuildWriteCommandPlansWithoutExecuting(t *testing.T) {
	connector := reverseETLFakeConnector()

	result, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"issue", "create"},
		Flags: map[string][]string{
			"title": []string{"Ship connector commands"},
			"body":  []string{"Plan first"},
		},
		Config: connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.Connector != "github" || result.Command != "issue create" || result.Write != "create_issue" {
		t.Fatalf("write command identity = %+v, want github issue create create_issue", result)
	}
	if result.MutationClass != "create" || result.TargetResource != "issue" {
		t.Fatalf("mutation target = %+v, want create issue", result)
	}
	if result.ApprovalRequired != true {
		t.Fatalf("ApprovalRequired = false, want true")
	}
	if connector.validateReq.Action != "create_issue" {
		t.Fatalf("ValidateWrite action = %q, want create_issue", connector.validateReq.Action)
	}
	if connector.dryRunReq.Action != "" {
		t.Fatalf("DryRunWrite action = %q, want not called", connector.dryRunReq.Action)
	}
	if connector.writeReq.Action != "" {
		t.Fatalf("Write action = %q, want not called", connector.writeReq.Action)
	}
	if got := result.Record["title"]; got != "Ship connector commands" {
		t.Fatalf("plan record title = %#v, want title", got)
	}
}

func TestBuildWriteCommandPreviewDryRunsAndRedactsSecretLikeFields(t *testing.T) {
	connector := reverseETLFakeConnector()
	connector.preview = connectors.WritePreview{
		Action:        "create_deploy_key",
		RecordsStaged: 1,
		Warnings:      []string{"resolved request: POST https://api.github.test/repos/octo/hello/keys"},
	}

	result, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"repo", "deploy-key", "add"},
		Flags: map[string][]string{
			"title": []string{"deploy"},
			"key":   []string{"ssh-rsa AAAA-sensitive"},
		},
		Config:  connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.Preview == nil {
		t.Fatalf("result = %+v, want plan and preview", result)
	}
	if connector.dryRunReq.Action != "create_deploy_key" {
		t.Fatalf("DryRunWrite action = %q, want create_deploy_key", connector.dryRunReq.Action)
	}
	if connector.writeReq.Action != "" {
		t.Fatalf("Write action = %q, want not called", connector.writeReq.Action)
	}
	if got := result.RedactedRecord["key"]; got != "***" {
		t.Fatalf("plan record key = %#v, want redacted", got)
	}
}

func TestRunReverseETLCommandRemainsNonExecutableInGenericRunner(t *testing.T) {
	connector := reverseETLFakeConnector()

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"issue", "close"},
		Flags: map[string][]string{"issue-number": []string{"101"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for reverse ETL command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want blocked generic runner")
	}
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Run error type = %T, want BlockedCommandError", err)
	}
	if !strings.Contains(err.Error(), "reverse_etl") {
		t.Fatalf("Run error = %q, want reverse_etl", err.Error())
	}
}

func TestRunReverseETLRejectsMissingWriteAndUnsupportedFlagMapping(t *testing.T) {
	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
		flags   map[string][]string
		want    string
	}{
		{
			name: "missing write",
			command: connectors.CommandSurfaceCommand{
				Path:         "issue create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Risk:         "creates issue",
				Approval:     "approval required",
			},
			want: "must reference write action",
		},
		{
			name: "no flag mappings",
			command: connectors.CommandSurfaceCommand{
				Path:         "repo fork",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "create_fork",
				Risk:         "creates fork",
				Approval:     "approval required",
			},
			want: "no declared flag mappings",
		},
		{
			name: "unsupported flag mapping",
			command: connectors.CommandSurfaceCommand{
				Path:         "issue create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "create_issue",
				Risk:         "creates issue",
				Approval:     "approval required",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "string", MapsTo: "query.state"},
				},
			},
			flags: map[string][]string{"state": []string{"open"}},
			want:  "unsupported target",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{
				surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{tt.command}},
				manifest: connectors.Manifest{WriteActions: []connectors.WriteActionSpec{
					{Name: "create_issue", Method: "POST", Path: "/issues"},
					{Name: "create_fork", Method: "POST", Path: "/forks"},
				}},
			}
			_, err := BuildWriteCommand(context.Background(), connector, Request{
				Path:  strings.Fields(tt.command.Path),
				Flags: tt.flags,
			})
			if err == nil {
				t.Fatal("Run error = nil, want rejection")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunImplementedOperationCommandIsFeatureGated(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "project list",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "github.projects.list",
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{Path: []string{"project", "list"}}, func(connectors.Record) error {
		t.Fatal("emit called for feature-gated operation command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want feature gate")
	}
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Run error type = %T, want BlockedCommandError", err)
	}
	if !strings.Contains(err.Error(), "operation github.projects.list") ||
		!strings.Contains(err.Error(), "executor is not implemented") {
		t.Fatalf("Run error = %q, want operation feature gate", err.Error())
	}
}

func reverseETLFakeConnector() *fakeConnector {
	return &fakeConnector{
		surface: &connectors.CommandSurface{
			Commands: []connectors.CommandSurfaceCommand{
				{
					Path:         "issue create",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_issue",
					Risk:         "creates a visible issue",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "title", Type: "string", MapsTo: "record.title"},
						{Name: "body", Type: "string", MapsTo: "record.body"},
					},
				},
				{
					Path:         "issue close",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "close_issue",
					Risk:         "closes an issue",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "issue-number", Type: "integer", MapsTo: "record.issue_number"},
					},
				},
				{
					Path:         "repo deploy-key add",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_deploy_key",
					Risk:         "adds deploy key",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "title", Type: "string", MapsTo: "record.title"},
						{Name: "key", Type: "string", MapsTo: "record.key"},
					},
				},
			},
		},
		manifest: connectors.Manifest{
			WriteActions: []connectors.WriteActionSpec{
				{Name: "create_issue", Method: "POST", Path: "/repos/{owner}/{repo}/issues", Risk: "creates issue"},
				{Name: "close_issue", Method: "PATCH", Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "closes issue"},
				{Name: "create_deploy_key", Method: "POST", Path: "/repos/{owner}/{repo}/keys", Risk: "adds deploy key"},
			},
		},
	}
}

func TestRunRejectsUnknownFlagAndInvalidEnum(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "issue list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "issues",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed", "all"}, MapsTo: "query.state"},
				},
			},
		},
	}}

	tests := []struct {
		name  string
		flags map[string][]string
		want  string
	}{
		{name: "unknown flag", flags: map[string][]string{"author": []string{"octocat"}}, want: "unknown flag"},
		{name: "invalid enum", flags: map[string][]string{"state": []string{"merged"}}, want: "invalid --state"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{
				Path:  []string{"issue", "list"},
				Flags: tt.flags,
			}, func(connectors.Record) error {
				t.Fatal("emit called for invalid flags")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunImplementedDirectReadCommand(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				OutputPolicy: "github_contents_file_metadata",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}}

	result, err := Run(context.Background(), connector, Request{
		Path:  []string{"repo", "read-file"},
		Flags: map[string][]string{"path": []string{"README.md"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
		}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for direct-read command")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.DirectRead == nil {
		t.Fatalf("DirectRead = nil, want result")
	}
	if connector.directReadReq.Method != "GET" {
		t.Fatalf("direct read method = %q, want GET", connector.directReadReq.Method)
	}
	if connector.directReadReq.PathParams["path"] != "README.md" {
		t.Fatalf("direct read path param = %q, want README.md", connector.directReadReq.PathParams["path"])
	}
	if connector.directReadReq.OutputPolicy != "github_contents_file_metadata" {
		t.Fatalf("direct read output policy = %q, want github_contents_file_metadata", connector.directReadReq.OutputPolicy)
	}
}

func TestRunDirectReadRejectsUnsafeEndpointMetadata(t *testing.T) {
	tests := []struct {
		name     string
		endpoint connectors.CommandSurfaceEndpointRef
		want     string
	}{
		{
			name:     "mutation method",
			endpoint: connectors.CommandSurfaceEndpointRef{Method: "POST", Path: "/repos/{owner}/{repo}/contents/{path}"},
			want:     "GET",
		},
		{
			name:     "absolute url",
			endpoint: connectors.CommandSurfaceEndpointRef{Method: "GET", Path: "https://evil.example.test/repos/{owner}/{repo}"},
			want:     "absolute URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{
				Commands: []connectors.CommandSurfaceCommand{
					{
						Path:         "repo read-file",
						Intent:       "direct_read",
						Availability: "implemented",
						APISurface:   []connectors.CommandSurfaceEndpointRef{tt.endpoint},
						OutputPolicy: "github_contents_file_metadata",
						Flags: []connectors.CommandSurfaceFlag{
							{Name: "path", Type: "string", MapsTo: "path.path"},
						},
					},
				},
			}}

			_, err := Run(context.Background(), connector, Request{
				Path:  []string{"repo", "read-file"},
				Flags: map[string][]string{"path": []string{"README.md"}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for rejected direct-read command")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want rejection")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunDirectReadRequiresOutputPolicy(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"repo", "read-file"},
		Flags: map[string][]string{"path": []string{"README.md"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for rejected direct-read command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want output policy rejection")
	}
	if !strings.Contains(err.Error(), "output_policy") {
		t.Fatalf("Run error = %q, want output_policy", err.Error())
	}
}
