package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const asanaWriteApprovalPolicy = "Requires plan, preview, explicit approval, then execute."

var expectedAsanaCommandTargets = map[string]string{
	"workspaces list":            "stream:workspaces",
	"projects list":              "stream:projects",
	"projects create":            "write:create_project",
	"projects update":            "write:update_project",
	"projects delete":            "write:delete_project",
	"tasks list":                 "stream:tasks",
	"tasks create":               "write:create_task",
	"tasks update":               "write:update_task",
	"tasks delete":               "write:delete_task",
	"tasks comment":              "write:add_comment",
	"users list":                 "stream:users",
	"teams list":                 "stream:teams",
	"tags list":                  "stream:tags",
	"tags create":                "write:create_tag",
	"tags update":                "write:update_tag",
	"tags delete":                "write:delete_tag",
	"sections list":              "stream:sections",
	"sections create":            "write:create_section",
	"sections update":            "write:update_section",
	"sections delete":            "write:delete_section",
	"stories list":               "stream:stories",
	"custom-fields list":         "stream:custom_fields",
	"project-statuses list":      "stream:project_statuses",
	"team-memberships list":      "stream:team_memberships",
	"workspace-memberships list": "stream:workspace_memberships",
}

var expectedAsanaStreamFlagTargets = map[string][]string{
	"workspaces list":            {},
	"projects list":              {"config.workspace_id"},
	"tasks list":                 {"config.assignee", "config.project_id", "config.workspace_id"},
	"users list":                 {"config.workspace_id"},
	"teams list":                 {"config.workspace_id"},
	"tags list":                  {"config.workspace_id"},
	"sections list":              {},
	"stories list":               {},
	"custom-fields list":         {"config.workspace_id"},
	"project-statuses list":      {},
	"team-memberships list":      {"config.team_id", "config.workspace_id"},
	"workspace-memberships list": {"config.workspace_id"},
}

func TestAsanaCLISurfaceTargetsAndCommandTree(t *testing.T) {
	bundle := loadAsanaBundleForCLISurface(t)
	surface := bundle.CLISurface
	if surface == nil {
		t.Fatal("Asana cli_surface.json was not loaded")
	}
	if surface.Usage != "pm asana <topic> <leaf> [flags]" {
		t.Fatalf("usage = %q", surface.Usage)
	}
	if surface.CoveragePolicy == nil || !surface.CoveragePolicy.RequireAllStreams || !surface.CoveragePolicy.RequireAllWrites {
		t.Fatalf("coverage_policy = %+v, want complete stream and write coverage", surface.CoveragePolicy)
	}
	if got, want := len(surface.Commands), 25; got != want {
		t.Fatalf("command count = %d, want %d", got, want)
	}

	gotTargets := make(map[string]string, len(surface.Commands))
	topics := map[string]bool{}
	for _, command := range surface.Commands {
		parts := strings.Fields(command.Path)
		if len(parts) != 2 || strings.Join(parts, " ") != command.Path {
			t.Errorf("command path %q must parse as exactly one topic and one leaf", command.Path)
			continue
		}
		topics[parts[0]] = true
		if command.Availability != "implemented" {
			t.Errorf("command %q availability = %q, want implemented", command.Path, command.Availability)
		}
		if command.Operation != "" {
			t.Errorf("command %q exposes operation target %q", command.Path, command.Operation)
		}
		if command.Intent == "raw_api" || command.Intent == "direct_write" {
			t.Errorf("command %q uses forbidden intent %q", command.Path, command.Intent)
		}
		target, err := asanaCommandTarget(command)
		if err != nil {
			t.Errorf("command %q: %v", command.Path, err)
			continue
		}
		if _, duplicate := gotTargets[command.Path]; duplicate {
			t.Errorf("duplicate command path %q", command.Path)
		}
		gotTargets[command.Path] = target
	}
	if !reflect.DeepEqual(gotTargets, expectedAsanaCommandTargets) {
		t.Errorf("command targets differ\n got: %v\nwant: %v", gotTargets, expectedAsanaCommandTargets)
	}

	groupedTopics := map[string]int{}
	for _, group := range surface.Groups {
		if strings.TrimSpace(group.ID) == "" || strings.TrimSpace(group.Title) == "" {
			t.Errorf("group must have an id and title: %+v", group)
		}
		for _, topic := range group.Commands {
			groupedTopics[topic]++
		}
	}
	for topic := range topics {
		if groupedTopics[topic] != 1 {
			t.Errorf("topic %q occurs in %d groups, want exactly one", topic, groupedTopics[topic])
		}
	}
	for topic, count := range groupedTopics {
		if !topics[topic] {
			t.Errorf("group references unknown topic %q", topic)
		}
		if count != 1 {
			t.Errorf("group topic %q occurs %d times", topic, count)
		}
	}
}

func TestAsanaCLISurfaceAPIRowCorrespondence(t *testing.T) {
	bundle := loadAsanaBundleForCLISurface(t)
	wantByTarget := map[string][]string{}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		target := ""
		if endpoint.CoveredBy.Stream != "" {
			target = "stream:" + endpoint.CoveredBy.Stream
		}
		if endpoint.CoveredBy.Write != "" {
			target = "write:" + endpoint.CoveredBy.Write
		}
		if target != "" {
			wantByTarget[target] = append(wantByTarget[target], strings.ToUpper(endpoint.Method)+" "+endpoint.Path)
		}
	}
	for target := range wantByTarget {
		sort.Strings(wantByTarget[target])
	}

	for _, command := range bundle.CLISurface.Commands {
		target, err := asanaCommandTarget(command)
		if err != nil {
			t.Fatalf("command %q: %v", command.Path, err)
		}
		got := make([]string, 0, len(command.APISurface))
		for _, endpoint := range command.APISurface {
			got = append(got, strings.ToUpper(endpoint.Method)+" "+endpoint.Path)
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, wantByTarget[target]) {
			t.Errorf("command %q API rows = %v, want %v for %s", command.Path, got, wantByTarget[target], target)
		}
	}
}

func TestAsanaCLISurfaceFlagsExamplesAndApproval(t *testing.T) {
	bundle := loadAsanaBundleForCLISurface(t)
	writes := map[string]engine.WriteAction{}
	for _, action := range bundle.Writes {
		writes[action.Name] = action
	}

	for _, command := range bundle.CLISurface.Commands {
		if len(command.Examples) == 0 {
			t.Errorf("command %q has no credential-free example", command.Path)
		}
		for _, example := range command.Examples {
			if !strings.HasPrefix(example, "pm asana "+command.Path) {
				t.Errorf("command %q example %q does not use its path", command.Path, example)
			}
			lower := strings.ToLower(example)
			for _, forbidden := range []string{"access-token", "credential", "password", "secret"} {
				if strings.Contains(lower, forbidden) {
					t.Errorf("command %q example contains credential-like text %q", command.Path, forbidden)
				}
			}
			if command.Stream != "" && !strings.Contains(example, "--json") {
				t.Errorf("stream command %q example must use deterministic --json output", command.Path)
			}
		}

		mapped := make([]string, 0, len(command.Flags))
		seenNames := map[string]bool{}
		for _, flag := range command.Flags {
			if seenNames[flag.Name] {
				t.Errorf("command %q duplicates flag %q", command.Path, flag.Name)
			}
			seenNames[flag.Name] = true
			for _, forbidden := range []string{"method", "url", "path", "headers", "raw-body"} {
				if flag.Name == forbidden {
					t.Errorf("command %q exposes raw transport flag %q", command.Path, flag.Name)
				}
			}
			if strings.HasPrefix(flag.MapsTo, "transport.") || strings.HasPrefix(flag.MapsTo, "request.") {
				t.Errorf("command %q maps flag %q to raw transport target %q", command.Path, flag.Name, flag.MapsTo)
			}
			mapped = append(mapped, flag.MapsTo)
		}
		sort.Strings(mapped)

		if command.Stream != "" {
			if !reflect.DeepEqual(mapped, expectedAsanaStreamFlagTargets[command.Path]) {
				t.Errorf("stream command %q flag targets = %v, want %v", command.Path, mapped, expectedAsanaStreamFlagTargets[command.Path])
			}
			continue
		}

		action := writes[command.Write]
		if command.Risk != action.Risk {
			t.Errorf("write command %q risk does not match action %q", command.Path, action.Name)
		}
		if !strings.Contains(strings.ToLower(action.Risk), "approval required") {
			t.Errorf("write action %q risk does not require approval: %q", action.Name, action.Risk)
		}
		if command.Approval != asanaWriteApprovalPolicy {
			t.Errorf("write command %q approval = %q, want %q", command.Path, command.Approval, asanaWriteApprovalPolicy)
		}

		valid, required := asanaWriteSchemaPaths(t, action)
		mappedSet := map[string]bool{}
		for _, target := range mapped {
			mappedSet[target] = true
			if !valid[target] {
				t.Errorf("write command %q flag maps to unknown record schema path %q", command.Path, target)
			}
		}
		for _, field := range action.PathFields {
			required["record."+field] = true
		}
		for target := range required {
			if !mappedSet[target] {
				t.Errorf("write command %q lacks required field mapping %q", command.Path, target)
			}
		}
	}
}

func loadAsanaBundleForCLISurface(t *testing.T) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "asana")
	if err != nil {
		t.Fatalf("load production Asana bundle: %v", err)
	}
	if bundle.CLISurface == nil {
		t.Fatal("Asana cli_surface.json was not loaded")
	}
	return bundle
}

func asanaCommandTarget(command engine.CLICommand) (string, error) {
	targets := make([]string, 0, 3)
	if command.Stream != "" {
		targets = append(targets, "stream:"+command.Stream)
	}
	if command.Write != "" {
		targets = append(targets, "write:"+command.Write)
	}
	if command.Operation != "" {
		targets = append(targets, "operation:"+command.Operation)
	}
	if len(targets) != 1 {
		return "", fmt.Errorf("has %d declarative targets, want exactly one", len(targets))
	}
	return targets[0], nil
}

type asanaCLIRecordSchema struct {
	Required   []string                        `json:"required"`
	Properties map[string]asanaCLIRecordSchema `json:"properties"`
}

func asanaWriteSchemaPaths(t *testing.T, action engine.WriteAction) (map[string]bool, map[string]bool) {
	t.Helper()
	var schema asanaCLIRecordSchema
	if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
		t.Fatalf("decode record schema for %q: %v", action.Name, err)
	}
	valid := map[string]bool{}
	required := map[string]bool{}
	var walk func(asanaCLIRecordSchema, string)
	walk = func(node asanaCLIRecordSchema, prefix string) {
		requiredNames := map[string]bool{}
		for _, name := range node.Required {
			requiredNames[name] = true
		}
		for name, child := range node.Properties {
			path := prefix + "." + name
			valid[path] = true
			if requiredNames[name] {
				required[path] = true
			}
			walk(child, path)
		}
	}
	walk(schema, "record")
	return valid, required
}
