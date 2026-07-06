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
	readReq       connectors.ReadRequest
	directReadReq connectors.DirectReadRequest
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
func (f *fakeConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, errors.New("write should not be called")
}
func (f *fakeConnector) CommandSurface() *connectors.CommandSurface { return f.surface }

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
				Path:         "issue create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "create_issue",
				Risk:         "creates a visible issue",
			},
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
		{name: "reverse_etl", path: []string{"issue", "create"}, want: "reverse_etl"},
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
