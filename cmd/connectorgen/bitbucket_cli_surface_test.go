package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBitbucketCLISurfaceMetadata(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/bitbucket/cli_surface.json")
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
	if implemented != 0 {
		t.Fatalf("implemented commands = %d, want 0 in metadata-only seed slice", implemented)
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
