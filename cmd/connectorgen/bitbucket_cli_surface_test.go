package main

import (
	"encoding/json"
	"os"
	"testing"
)

const bitbucketDefsDir = "../../internal/connectors/defs/bitbucket"

func TestBitbucketCLISurfaceMetadata(t *testing.T) {
	raw, err := os.ReadFile(bitbucketDefsDir + "/cli_surface.json")
	if err != nil {
		t.Fatalf("read bitbucket cli_surface.json: %v", err)
	}

	var surface struct {
		Tagline string `json:"tagline"`
		Usage   string `json:"usage"`
		Groups  []struct {
			ID       string   `json:"id"`
			Commands []string `json:"commands"`
		} `json:"groups"`
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Risk         string `json:"risk"`
			Approval     string `json:"approval"`
			Notes        string `json:"notes"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal bitbucket cli_surface.json: %v", err)
	}

	if surface.Tagline == "" {
		t.Fatal("tagline is empty")
	}
	if surface.Usage != "pm bitbucket <command> <subcommand> [flags]" {
		t.Fatalf("usage = %q, want Bitbucket command usage", surface.Usage)
	}
	if len(surface.Commands) < 20 {
		t.Fatalf("commands = %d, want at least 20 provider-like commands", len(surface.Commands))
	}

	groups := map[string]bool{}
	for _, group := range surface.Groups {
		groups[group.ID] = len(group.Commands) > 0
	}
	for _, id := range []string{"repositories", "pull_requests", "issues", "pipelines", "admin", "local"} {
		if !groups[id] {
			t.Fatalf("group %q missing or empty", id)
		}
	}

	commands := map[string]struct {
		Intent       string
		Availability string
		Risk         string
		Approval     string
		Notes        string
	}{}
	implemented := 0
	for _, cmd := range surface.Commands {
		commands[cmd.Path] = struct {
			Intent       string
			Availability string
			Risk         string
			Approval     string
			Notes        string
		}{cmd.Intent, cmd.Availability, cmd.Risk, cmd.Approval, cmd.Notes}
		if cmd.Availability == "implemented" {
			implemented++
		}
		if cmd.Intent == "reverse_etl" && (cmd.Risk == "" || cmd.Approval == "") {
			t.Fatalf("reverse_etl command %q missing risk/approval", cmd.Path)
		}
		if cmd.Intent == "raw_api" && cmd.Availability != "unsafe_or_disallowed" {
			t.Fatalf("raw API command %q availability = %q, want unsafe_or_disallowed", cmd.Path, cmd.Availability)
		}
		if cmd.Intent == "direct_write" && cmd.Availability != "unsafe_or_disallowed" {
			t.Fatalf("direct write command %q availability = %q, want unsafe_or_disallowed", cmd.Path, cmd.Availability)
		}
	}
	if implemented < 10 {
		t.Fatalf("implemented commands = %d, want at least 10 executable Bitbucket commands after parity implementation", implemented)
	}

	wantIntents := map[string]string{
		"repo list":                 "etl",
		"repo view":                 "direct_read",
		"repo clone":                "local_workflow",
		"pull-request list":         "etl",
		"pull-request create":       "reverse_etl",
		"issue list":                "etl",
		"issue create":              "reverse_etl",
		"pipeline list":             "etl",
		"pipeline run":              "reverse_etl",
		"deployment list":           "etl",
		"download list":             "etl",
		"download get":              "local_workflow",
		"webhook list":              "etl",
		"webhook create":            "reverse_etl",
		"branch-restriction list":   "etl",
		"branch-restriction create": "reverse_etl",
		"workspace list":            "direct_read",
		"project list":              "direct_read",
		"snippet list":              "direct_read",
		"api":                       "raw_api",
	}
	for path, intent := range wantIntents {
		cmd, ok := commands[path]
		if !ok {
			t.Fatalf("command %q missing", path)
		}
		if cmd.Intent != intent {
			t.Fatalf("command %q intent = %q, want %q", path, cmd.Intent, intent)
		}
		if cmd.Availability == "unknown" {
			t.Fatalf("command %q availability must be classified", path)
		}
		if cmd.Intent == "raw_api" && cmd.Availability != "unsafe_or_disallowed" {
			t.Fatalf("raw API command %q availability = %q, want unsafe_or_disallowed", path, cmd.Availability)
		}
	}
}

func TestBitbucketFullParityMetadata(t *testing.T) {
	apiRaw, err := os.ReadFile(bitbucketDefsDir + "/api_surface.json")
	if err != nil {
		t.Fatalf("read bitbucket api_surface.json: %v", err)
	}
	var apiSurface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			CoveredBy *struct {
				Stream      string   `json:"stream"`
				Write       string   `json:"write"`
				DirectRead  string   `json:"direct_read"`
				DirectReads []string `json:"direct_reads"`
			} `json:"covered_by"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				Risk             string `json:"risk"`
				BlockedByDefault bool   `json:"blocked_by_default"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(apiRaw, &apiSurface); err != nil {
		t.Fatalf("unmarshal bitbucket api_surface.json: %v", err)
	}
	if apiSurface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", apiSurface.OperationLedgerVersion)
	}
	if len(apiSurface.Endpoints) != 331 {
		t.Fatalf("api_surface endpoints = %d, want official Bitbucket Swagger count 331", len(apiSurface.Endpoints))
	}

	covered := map[string]bool{}
	blocked := 0
	for _, ep := range apiSurface.Endpoints {
		if ep.CoveredBy != nil {
			if ep.CoveredBy.Stream != "" {
				covered["stream:"+ep.CoveredBy.Stream] = true
			}
			if ep.CoveredBy.Write != "" {
				covered["write:"+ep.CoveredBy.Write] = true
			}
			if ep.CoveredBy.DirectRead != "" {
				covered["direct:"+ep.CoveredBy.DirectRead] = true
			}
			for _, name := range ep.CoveredBy.DirectReads {
				covered["direct:"+name] = true
			}
		}
		if ep.Operation != nil {
			blocked++
			if ep.Operation.Status != "blocked" || !ep.Operation.BlockedByDefault {
				t.Fatalf("operation %s %s status/default = %s/%t, want blocked/default", ep.Method, ep.Path, ep.Operation.Status, ep.Operation.BlockedByDefault)
			}
		}
	}
	for _, want := range []string{
		"stream:repositories",
		"stream:pull_requests",
		"stream:issues",
		"stream:pipelines",
		"write:create_issue",
		"write:create_pull_request",
		"direct:repo view",
		"direct:pull-request view",
	} {
		if !covered[want] {
			t.Fatalf("api_surface missing coverage %q", want)
		}
	}
	if blocked == 0 {
		t.Fatal("api_surface has no blocked operation rows")
	}

	opsRaw, err := os.ReadFile(bitbucketDefsDir + "/operations.json")
	if err != nil {
		t.Fatalf("read bitbucket operations.json: %v", err)
	}
	var ops struct {
		Operations []struct {
			ID              string `json:"id"`
			Kind            string `json:"kind"`
			MutationClass   string `json:"mutation_class"`
			Destructive     bool   `json:"destructive"`
			SecretSensitive bool   `json:"secret_sensitive"`
			SensitivePolicy *struct {
				ApprovalMode string `json:"approval_mode"`
			} `json:"sensitive_policy"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(opsRaw, &ops); err != nil {
		t.Fatalf("unmarshal bitbucket operations.json: %v", err)
	}
	if len(ops.Operations) != 331 {
		t.Fatalf("operations = %d, want official Bitbucket Swagger count 331", len(ops.Operations))
	}
	seenSensitive := false
	seenDestructive := false
	for _, op := range ops.Operations {
		if op.Kind == "graphql_query" || op.Kind == "graphql_mutation" {
			t.Fatalf("Bitbucket operation %q uses GraphQL kind; want REST-only ledger", op.ID)
		}
		if op.Destructive {
			seenDestructive = true
			if op.SensitivePolicy == nil || op.SensitivePolicy.ApprovalMode != "typed_confirmation" {
				t.Fatalf("destructive operation %q missing typed confirmation policy", op.ID)
			}
		}
		if op.SecretSensitive {
			seenSensitive = true
			if op.SensitivePolicy == nil {
				t.Fatalf("secret-sensitive operation %q missing sensitive_policy", op.ID)
			}
		}
	}
	if !seenDestructive || !seenSensitive {
		t.Fatalf("operations ledger destructive=%t secret_sensitive=%t, want both classifications", seenDestructive, seenSensitive)
	}
}
