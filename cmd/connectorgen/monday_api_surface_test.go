package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMondayAPISurfaceOperationLedger(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "monday")

	var metadata struct {
		Capabilities struct {
			Write bool `json:"write"`
		} `json:"capabilities"`
	}
	readJSONFile(t, filepath.Join(root, "metadata.json"), &metadata)
	if !metadata.Capabilities.Write {
		t.Fatalf("metadata capabilities.write = false, want true for typed Monday reverse ETL actions")
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			CoveredBy *struct {
				Stream     string `json:"stream"`
				Write      string `json:"write"`
				DirectRead string `json:"direct_read"`
			} `json:"covered_by"`
			Excluded  any `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				Notes            string `json:"notes"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	readJSONFile(t, filepath.Join(root, "api_surface.json"), &surface)
	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != 254 {
		t.Fatalf("endpoints = %d, want 254 official docs GraphQL operations", len(surface.Endpoints))
	}
	seen := map[string]bool{}
	coveredStreams := map[string]bool{}
	coveredWrites := map[string]bool{}
	blocked := 0
	blockedQueries := 0
	for _, ep := range surface.Endpoints {
		if ep.Excluded != nil {
			t.Fatalf("%s %s uses legacy excluded in operation ledger mode", ep.Method, ep.Path)
		}
		if ep.Path != "/" {
			t.Fatalf("%s path = %q, want real Monday GraphQL endpoint /", ep.Method, ep.Path)
		}
		classifier := ""
		if ep.CoveredBy != nil {
			switch {
			case ep.CoveredBy.Stream != "":
				classifier = "stream:" + ep.CoveredBy.Stream
			case ep.CoveredBy.Write != "":
				classifier = "write:" + ep.CoveredBy.Write
			case ep.CoveredBy.DirectRead != "":
				classifier = "direct_read:" + ep.CoveredBy.DirectRead
			}
		}
		if ep.Operation != nil {
			if strings.TrimSpace(ep.Operation.Notes) == "" {
				t.Fatalf("operation row %s %s is missing non-routing operation notes", ep.Method, ep.Path)
			}
			classifier = "operation:" + ep.Operation.Notes
		}
		key := strings.ToUpper(ep.Method) + " " + ep.Path + " " + classifier
		if seen[key] {
			t.Fatalf("duplicate api_surface endpoint classifier %s", key)
		}
		seen[key] = true
		if ep.CoveredBy != nil {
			if ep.CoveredBy.Stream != "" {
				coveredStreams[ep.CoveredBy.Stream] = true
			}
			if ep.CoveredBy.Write != "" {
				coveredWrites[ep.CoveredBy.Write] = true
			}
			if ep.CoveredBy.DirectRead != "" {
				t.Fatalf("api_surface direct_read coverage %q must stay planned until shared duplicate POST / validation lands", ep.CoveredBy.DirectRead)
			}
		}
		if ep.Operation != nil {
			blocked++
			if ep.Operation.Status != "blocked" || !ep.Operation.BlockedByDefault {
				t.Fatalf("operation row %s must remain blocked_by_default", key)
			}
			if strings.HasPrefix(ep.Operation.Notes, "Monday GraphQL query ") {
				blockedQueries++
				if !strings.Contains(ep.Operation.Reason, "duplicate Monday POST / query classifiers") || !strings.Contains(ep.Operation.Reason, "errors[]") {
					t.Fatalf("query operation row %s missing shared blocker evidence", key)
				}
			}
		}
	}
	for _, stream := range []string{"boards", "items", "users", "teams", "tags"} {
		if !coveredStreams[stream] {
			t.Fatalf("stream %q is not covered in api_surface.json", stream)
		}
	}
	if blockedQueries != 61 {
		t.Fatalf("blocked query operation rows = %d, want 61", blockedQueries)
	}
	if blocked != 147 {
		t.Fatalf("blocked operation rows = %d, want 147", blocked)
	}

	type mondaySensitivePolicy struct {
		InputMode    string   `json:"input_mode"`
		RedactFields []string `json:"redact_fields"`
	}
	type mondayOperation struct {
		ID              string                 `json:"id"`
		Kind            string                 `json:"kind"`
		Description     string                 `json:"description"`
		Approval        string                 `json:"approval"`
		Risk            string                 `json:"risk"`
		MutationClass   string                 `json:"mutation_class"`
		Destructive     bool                   `json:"destructive"`
		SensitivePolicy *mondaySensitivePolicy `json:"sensitive_policy"`
		REST            *struct {
			Path string `json:"path"`
		} `json:"rest"`
		GraphQL *struct {
			Document      string `json:"document"`
			OperationName string `json:"operation_name"`
		} `json:"graphql"`
	}
	var operations struct {
		Operations []mondayOperation `json:"operations"`
	}
	readJSONFile(t, filepath.Join(root, "operations.json"), &operations)
	if len(operations.Operations) != 254 {
		t.Fatalf("operations = %d, want 254", len(operations.Operations))
	}
	counts := map[string]int{}
	fileUploadOps := map[string]bool{}
	operationByID := map[string]mondayOperation{}
	sensitiveWriteRedactions := map[string][]string{}
	for _, op := range operations.Operations {
		counts[op.Kind]++
		operationByID[op.ID] = op
		if op.SensitivePolicy != nil && len(op.SensitivePolicy.RedactFields) > 0 {
			if op.SensitivePolicy.InputMode != "stdin" {
				t.Fatalf("sensitive operation %q input_mode = %q, want stdin", op.ID, op.SensitivePolicy.InputMode)
			}
			writeName := strings.TrimPrefix(op.ID, "monday.mutation.")
			if coveredWrites[writeName] {
				sensitiveWriteRedactions[writeName] = append([]string(nil), op.SensitivePolicy.RedactFields...)
			}
		}
		if op.ID == "monday.mutation.add_file_to_column" || op.ID == "monday.mutation.add_file_to_update" {
			fileUploadOps[op.ID] = true
			if !strings.Contains(op.Description, "multipart/binary GraphQL upload handling") || !strings.Contains(op.Description, "planned/blocked") {
				t.Fatalf("file upload operation %q missing multipart planned/blocker evidence", op.ID)
			}
			if !strings.Contains(op.Approval, "blocked until typed multipart/binary GraphQL upload contract exists") {
				t.Fatalf("file upload operation %q approval = %q, want binary blocker", op.ID, op.Approval)
			}
			if op.GraphQL == nil || strings.Contains(op.GraphQL.Document, "YOUR_FILE") || !strings.Contains(op.GraphQL.Document, "$file: File!") {
				t.Fatalf("file upload operation %q must keep fixed file variable metadata without executable placeholders", op.ID)
			}
		}
		if op.Kind == "rest_read" && strings.HasPrefix(op.ID, "monday.query.") {
			t.Fatalf("query operation %q must stay planned as graphql_query, not rest_read", op.ID)
		}
		if op.Kind == "graphql_query" {
			if op.GraphQL == nil || strings.TrimSpace(op.GraphQL.Document) == "" || strings.TrimSpace(op.GraphQL.OperationName) == "" {
				t.Fatalf("graphql_query operation %q is missing fixed GraphQL metadata", op.ID)
			}
			if op.REST != nil {
				t.Fatalf("graphql_query operation %q must not declare executable rest metadata", op.ID)
			}
		}
	}
	if counts["graphql_query"] != 66 {
		t.Fatalf("graphql_query operations = %d, want 66", counts["graphql_query"])
	}
	if counts["rest_read"] != 0 {
		t.Fatalf("rest_read operations = %d, want 0 for Monday query blockers", counts["rest_read"])
	}
	if counts["graphql_mutation"] != 188 {
		t.Fatalf("graphql_mutation operations = %d, want 188", counts["graphql_mutation"])
	}
	for _, id := range []string{"monday.mutation.add_file_to_column", "monday.mutation.add_file_to_update"} {
		if !fileUploadOps[id] {
			t.Fatalf("file upload operation %q missing from operation ledger", id)
		}
	}
	for _, tt := range []struct {
		id            string
		mutationClass string
	}{
		{id: "monday.mutation.create_doc_block", mutationClass: "create"},
		{id: "monday.mutation.update_doc_block", mutationClass: "update"},
	} {
		op, ok := operationByID[tt.id]
		if !ok {
			t.Fatalf("operation %q missing from operation ledger", tt.id)
		}
		if op.MutationClass != tt.mutationClass {
			t.Fatalf("operation %q mutation_class = %q, want %q", tt.id, op.MutationClass, tt.mutationClass)
		}
		if op.Destructive {
			t.Fatalf("operation %q is destructive, want false", tt.id)
		}
		if op.SensitivePolicy != nil {
			t.Fatalf("operation %q has destructive sensitive_policy, want none", tt.id)
		}
	}

	type mondayWriteAction struct {
		Name         string   `json:"name"`
		Kind         string   `json:"kind"`
		Risk         string   `json:"risk"`
		Confirm      string   `json:"confirm"`
		RedactFields []string `json:"redact_fields"`
		GraphQL      *struct {
			Document string `json:"document"`
		} `json:"graphql"`
	}
	var writes struct {
		Actions []mondayWriteAction `json:"actions"`
	}
	readJSONFile(t, filepath.Join(root, "writes.json"), &writes)
	if len(writes.Actions) == 0 {
		t.Fatal("writes.json has zero actions")
	}
	var sawDeleteBoard bool
	writeActionByName := map[string]mondayWriteAction{}
	for _, action := range writes.Actions {
		writeActionByName[action.Name] = action
		if strings.Contains(action.Name, "add_file_to_") {
			t.Fatalf("binary upload action %q must remain planned/blocked, not executable in writes.json", action.Name)
		}
		if !coveredWrites[action.Name] {
			t.Fatalf("write action %q has no api_surface covered_by.write row", action.Name)
		}
		if action.GraphQL == nil || !strings.Contains(action.GraphQL.Document, action.Name) {
			t.Fatalf("write action %q must use a fixed GraphQL document", action.Name)
		}
		if action.Name == "delete_board" {
			sawDeleteBoard = true
			if action.Confirm != "destructive" {
				t.Fatalf("delete_board confirm = %q, want destructive", action.Confirm)
			}
		}
	}
	if !sawDeleteBoard {
		t.Fatal("delete_board write action missing")
	}
	for _, tt := range []struct {
		name string
		kind string
	}{
		{name: "create_doc_block", kind: "create"},
		{name: "update_doc_block", kind: "update"},
	} {
		action, ok := writeActionByName[tt.name]
		if !ok {
			t.Fatalf("write action %q missing", tt.name)
		}
		if action.Kind != tt.kind {
			t.Fatalf("write action %q kind = %q, want %q", tt.name, action.Kind, tt.kind)
		}
		if action.Confirm != "" {
			t.Fatalf("write action %q confirm = %q, want empty", tt.name, action.Confirm)
		}
		if strings.Contains(action.Risk, "critical") {
			t.Fatalf("write action %q risk = %q, want non-critical", tt.name, action.Risk)
		}
	}
	for writeName, fields := range sensitiveWriteRedactions {
		action, ok := writeActionByName[writeName]
		if !ok {
			t.Fatalf("sensitive write action %q missing", writeName)
		}
		for _, field := range fields {
			var found bool
			for _, redacted := range action.RedactFields {
				if redacted == field {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("write action %q redact_fields = %v, want %q", writeName, action.RedactFields, field)
			}
		}
	}

	type mondayCLIFlag struct {
		Name   string `json:"name"`
		MapsTo string `json:"maps_to"`
	}
	type mondayCLICommand struct {
		Path         string          `json:"path"`
		Intent       string          `json:"intent"`
		Availability string          `json:"availability"`
		Write        string          `json:"write"`
		Operation    string          `json:"operation"`
		Notes        string          `json:"notes"`
		Flags        []mondayCLIFlag `json:"flags"`
		APISurface   []struct {
			Path string `json:"path"`
		} `json:"api_surface"`
	}
	var cli struct {
		Commands []mondayCLICommand `json:"commands"`
	}
	readJSONFile(t, filepath.Join(root, "cli_surface.json"), &cli)
	var sawDeleteCommand bool
	var sawPlannedAccountQuery bool
	implementedDirectReads := 0
	plannedQueryCommands := 0
	cliCommandByPath := map[string]mondayCLICommand{}
	for _, cmd := range cli.Commands {
		cliCommandByPath[cmd.Path] = cmd
		for _, ep := range cmd.APISurface {
			if ep.Path != "/" {
				t.Fatalf("cli command %q api_surface path = %q, want real Monday GraphQL endpoint /", cmd.Path, ep.Path)
			}
		}
		if cmd.Intent == "direct_read" && strings.HasPrefix(cmd.Path, "query ") {
			if cmd.Availability == "implemented" {
				implementedDirectReads++
			}
			if cmd.Availability == "planned" {
				plannedQueryCommands++
			}
			if cmd.Operation != "" || len(cmd.APISurface) != 0 {
				t.Fatalf("planned query command %q must not declare executable operation/api_surface metadata", cmd.Path)
			}
			if cmd.Path == "query account" && cmd.Availability == "planned" && strings.Contains(cmd.Notes, "errors[]") {
				sawPlannedAccountQuery = true
			}
		}
		if cmd.Availability == "implemented" && cmd.Write != "" && len(sensitiveWriteRedactions[cmd.Write]) > 0 {
			t.Fatalf("implemented CLI command %q exposes sensitive write %q", cmd.Path, cmd.Write)
		}
		if cmd.Path == "reverse delete-board" && cmd.Intent == "reverse_etl" && cmd.Availability == "implemented" && cmd.Write == "delete_board" {
			sawDeleteCommand = true
		}
	}
	for writeName, fields := range sensitiveWriteRedactions {
		path := "reverse " + strings.ReplaceAll(writeName, "_", "-")
		cmd, ok := cliCommandByPath[path]
		if !ok {
			t.Fatalf("CLI command %q missing", path)
		}
		if cmd.Availability != "planned" || cmd.Operation != "monday.mutation."+writeName || cmd.Write != "" {
			t.Fatalf("CLI command %q availability/write/operation = %q/%q/%q, want planned operation-only", path, cmd.Availability, cmd.Write, cmd.Operation)
		}
		for _, flag := range cmd.Flags {
			for _, field := range fields {
				if flag.MapsTo == "record."+field {
					t.Fatalf("planned CLI command %q still exposes sensitive flag %q", path, flag.Name)
				}
			}
		}
	}
	if implementedDirectReads != 0 {
		t.Fatalf("implemented direct query commands = %d, want 0", implementedDirectReads)
	}
	if plannedQueryCommands != 61 {
		t.Fatalf("planned query commands = %d, want 61", plannedQueryCommands)
	}
	if !sawPlannedAccountQuery {
		t.Fatal("planned query account metadata missing shared GraphQL blocker evidence")
	}
	if !sawDeleteCommand {
		t.Fatal("implemented CLI metadata for reverse delete-board missing")
	}
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}
