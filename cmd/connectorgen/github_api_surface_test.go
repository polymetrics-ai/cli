package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestGitHubAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal github api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	risks := map[string]int{}
	statuses := map[string]int{}
	covered, excluded, operations := 0, 0, 0

	for i, ep := range surface.Endpoints {
		totalByMethod[ep.Method]++
		if len(ep.CoveredBy) > 0 {
			covered++
			coveredByMethod[ep.Method]++
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			risks[ep.Operation.Risk]++
			statuses[ep.Operation.Status]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
			if ep.Operation.Model == "duplicate" && ep.Operation.DuplicateOf == "" {
				t.Fatalf("endpoint %d duplicate operation is missing duplicate_of", i)
			}
		}
	}

	// These are a snapshot of derived truth, not a budget. They moved when the
	// bundle stopped enumerating only /repos/{owner}/{repo}/… and recorded the
	// whole documented surface: 1220 REST operations plus the 4 fixed GraphQL
	// rows, which are counted separately and never folded into the REST total.
	// Every structural assertion below is unchanged; only the counts the surface
	// now actually holds were re-derived.
	if len(surface.Endpoints) != 1224 {
		t.Fatalf("endpoints = %d, want 1224", len(surface.Endpoints))
	}
	if covered != 1126 {
		t.Fatalf("covered endpoints = %d, want 1126", covered)
	}
	if operations != 98 {
		t.Fatalf("operation endpoints = %d, want 98 (duplicate/deprecated, plus the operations no runtime component can execute)", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE":  187,
		"GET":     636,
		"GRAPHQL": 4,
		"PATCH":   70,
		"POST":    193,
		"PUT":     134,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE":  179,
		"GET":     571,
		"GRAPHQL": 4,
		"PATCH":   65,
		"POST":    175,
		"PUT":     132,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 8,
		"GET":    65,
		"PATCH":  5,
		"POST":   18,
		"PUT":    2,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"disallowed": 19,
		"duplicate":  67,
		"deprecated": 1,
		// GETs GitHub documents with no JSON success body: 9 boolean 204 status
		// checks plus /zen and /octocat, which return text/plain and
		// application/octocat-stream. engine.decodeDirectReadBody json-decodes
		// every direct-read body, so these are blocked with that dependency
		// named rather than shipped as commands that always fail.
		"direct_read": 11,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"low": 98,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 98,
	})
}

// TestGitHubStatusAndTextOperationContracts pins the classified endpoints
// which become executable only after the closed none/text/raw-body engine
// contracts exist. This is intentionally a concrete table rather than a
// promotion of every blocked direct_read row: each entry asserts its documented
// method, bounded response policy, input shape, and the real operation
// preflight the command uses.
func TestGitHubStatusAndTextOperationContracts(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load embedded github bundle: %v", err)
	}
	commands := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}
	operations := map[string]engine.OperationSpec{}
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}

	type flag struct {
		name, kind, mapsTo string
		required           bool
		values             []string
	}
	tests := []struct {
		command, method, endpoint, outputPolicy, contentType, bodyType string
		maxBytes                                                       int
		flags                                                          []flag
	}{
		{"gists star check", "GET", "/gists/{gist_id}/star", "none", "", "", 1024, []flag{{"gist-id", "string", "path.gist_id", true, nil}}},
		{"orgs blocks check", "GET", "/orgs/{org}/blocks/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"orgs members check", "GET", "/orgs/{org}/members/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"orgs public-members check", "GET", "/orgs/{org}/public_members/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"teams members check", "GET", "/teams/{team_id}/members/{username}", "none", "", "", 1024, []flag{{"team-id", "integer", "path.team_id", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"user blocks check", "GET", "/user/blocks/{username}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}}},
		{"user following check", "GET", "/user/following/{username}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}}},
		{"user starred check", "GET", "/user/starred/{owner}/{repo}", "none", "", "", 1024, []flag{{"owner", "string", "path.owner", true, nil}, {"repo", "string", "path.repo", true, nil}}},
		{"users following check", "GET", "/users/{username}/following/{target_user}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}, {"target-user", "string", "path.target_user", true, nil}}},
		{"meta zen view", "GET", "/zen", "text", "", "", 1024, nil},
		{"meta octocat view", "GET", "/octocat", "text", "", "", 16384, []flag{{"s", "string", "query.s", false, nil}}},
		{"markdown render", "POST", "/markdown", "text", "application/json", "object", 1048576, []flag{{"text", "string", "body.text", true, nil}, {"mode", "enum", "body.mode", false, []string{"markdown", "gfm"}}, {"context", "string", "body.context", false, nil}}},
		{"markdown raw render", "POST", "/markdown/raw", "text", "text/plain", "string", 400 * 1024, []flag{{"text", "string", "body", true, nil}}},
	}

	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.command, " ", "_"), func(t *testing.T) {
			command, ok := commands[tt.command]
			if !ok {
				t.Fatalf("generated command %q is missing", tt.command)
			}
			if command.Intent != "direct_read" || command.Availability != "implemented" {
				t.Fatalf("command = intent %q availability %q, want implemented direct_read", command.Intent, command.Availability)
			}
			if command.OutputPolicy != tt.outputPolicy {
				t.Fatalf("output_policy = %q, want %q", command.OutputPolicy, tt.outputPolicy)
			}
			op, ok := operations[command.Operation]
			if !ok || op.REST == nil {
				t.Fatalf("command operation %q = %#v, want REST operation", command.Operation, op)
			}
			if op.Kind != "rest_read" || op.REST.Method != tt.method || op.REST.Path != tt.endpoint || op.REST.MaxBytes != tt.maxBytes {
				t.Fatalf("operation = kind=%q method=%q path=%q max_bytes=%d, want rest_read %s %s %d", op.Kind, op.REST.Method, op.REST.Path, op.REST.MaxBytes, tt.method, tt.endpoint, tt.maxBytes)
			}
			if op.REST.ContentType != tt.contentType {
				t.Fatalf("content_type = %q, want %q", op.REST.ContentType, tt.contentType)
			}
			if tt.bodyType != "" {
				var schema map[string]any
				if err := json.Unmarshal(op.REST.BodySchema, &schema); err != nil {
					t.Fatalf("unmarshal body_schema: %v", err)
				}
				if schema["type"] != tt.bodyType {
					t.Fatalf("body_schema type = %#v, want %q", schema["type"], tt.bodyType)
				}
			}
			if err := engine.PreflightOperationDirectRead(bundle, command.Operation, tt.method, tt.endpoint, tt.maxBytes, command.OutputPolicy); err != nil {
				t.Fatalf("operation preflight: %v", err)
			}
			for _, want := range tt.flags {
				got, ok := githubCommandFlag(command, want.name)
				if !ok {
					t.Fatalf("flag --%s is missing from %#v", want.name, command.Flags)
				}
				if got.Type != want.kind || got.MapsTo != want.mapsTo || got.Required != want.required || !reflect.DeepEqual(got.Values, want.values) {
					t.Fatalf("flag --%s = type=%q maps_to=%q required=%t values=%#v, want type=%q maps_to=%q required=%t values=%#v", want.name, got.Type, got.MapsTo, got.Required, got.Values, want.kind, want.mapsTo, want.required, want.values)
				}
			}
		})
	}
}

func githubCommandFlag(command engine.CLICommand, name string) (engine.CLIFlag, bool) {
	for _, flag := range command.Flags {
		if flag.Name == name {
			return flag, true
		}
	}
	return engine.CLIFlag{}, false
}

type githubOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
	DuplicateOf      string `json:"duplicate_of"`
}

func requiresSourceOrNotes(model string) bool {
	switch model {
	case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action", "disallowed":
		return true
	default:
		return false
	}
}

func assertStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
