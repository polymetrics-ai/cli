package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFreshchatAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/freshchat/api_surface.json")
	if err != nil {
		t.Fatalf("read freshchat api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			Path      string           `json:"path"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal freshchat api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	streams, writes, directReads, excluded, operations := 0, 0, 0, 0, 0
	models := map[string]int{}
	for i, ep := range surface.Endpoints {
		if _, ok := ep.CoveredBy["stream"]; ok {
			streams++
		}
		if _, ok := ep.CoveredBy["write"]; ok {
			writes++
		}
		if _, ok := ep.CoveredBy["direct_read"]; ok {
			directReads++
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			models[ep.Operation.Model]++
			if ep.Operation.Status != "blocked" {
				t.Fatalf("endpoint %d (%s %s) operation status = %q, want blocked", i, ep.Method, ep.Path, ep.Operation.Status)
			}
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d (%s %s) operation is not blocked by default", i, ep.Method, ep.Path)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d (%s %s) operation is missing reason", i, ep.Method, ep.Path)
			}
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d (%s %s) operation %q is missing source_url or notes", i, ep.Method, ep.Path, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 34 {
		t.Fatalf("endpoints = %d, want 34", len(surface.Endpoints))
	}
	if streams != 18 {
		t.Fatalf("stream endpoints = %d, want 18", streams)
	}
	if writes != 13 {
		t.Fatalf("write endpoints = %d, want 13", writes)
	}
	if directReads != 1 {
		t.Fatalf("direct-read endpoints = %d, want 1", directReads)
	}
	if operations != 2 {
		t.Fatalf("blocked operation endpoints = %d, want 2", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "models", models, map[string]int{
		"disallowed": 2,
	})
}

func TestFreshchatSensitiveAdminWritesRequireTypedConfirmation(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/freshchat/writes.json")
	if err != nil {
		t.Fatalf("read freshchat writes.json: %v", err)
	}
	var writes struct {
		Actions []struct {
			Name    string `json:"name"`
			Confirm string `json:"confirm"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(raw, &writes); err != nil {
		t.Fatalf("unmarshal freshchat writes.json: %v", err)
	}

	confirmByAction := map[string]string{}
	for _, action := range writes.Actions {
		confirmByAction[action.Name] = action.Confirm
	}
	want := map[string]string{
		"delete_user":                    "destructive",
		"create_agent":                   "admin",
		"update_agent":                   "admin",
		"update_agent_status":            "admin",
		"delete_agent":                   "destructive",
		"send_conversation_message":      "sensitive",
		"send_outbound_whatsapp_message": "sensitive",
		"extract_report":                 "sensitive",
	}
	for action, confirm := range want {
		if got := confirmByAction[action]; got != confirm {
			t.Fatalf("write action %q confirm = %q, want %q", action, got, confirm)
		}
	}
}

func TestFreshchatBinaryUploadCommandsUseTypedOperations(t *testing.T) {
	cliRaw, err := os.ReadFile("../../internal/connectors/defs/freshchat/cli_surface.json")
	if err != nil {
		t.Fatalf("read freshchat cli_surface.json: %v", err)
	}
	opRaw, err := os.ReadFile("../../internal/connectors/defs/freshchat/operations.json")
	if err != nil {
		t.Fatalf("read freshchat operations.json: %v", err)
	}

	var cliSurface struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Operation    string `json:"operation"`
			Notes        string `json:"notes"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cliSurface); err != nil {
		t.Fatalf("unmarshal freshchat cli_surface.json: %v", err)
	}
	var operations struct {
		Operations []struct {
			ID       string `json:"id"`
			Kind     string `json:"kind"`
			Approval string `json:"approval"`
			File     *struct {
				Direction string `json:"direction"`
				Path      string `json:"path"`
				MaxBytes  int    `json:"max_bytes"`
			} `json:"file"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(opRaw, &operations); err != nil {
		t.Fatalf("unmarshal freshchat operations.json: %v", err)
	}

	ops := map[string]struct {
		kind      string
		path      string
		maxBytes  int
		approval  string
		direction string
	}{}
	for _, op := range operations.Operations {
		entry := struct {
			kind      string
			path      string
			maxBytes  int
			approval  string
			direction string
		}{kind: op.Kind, approval: op.Approval}
		if op.File != nil {
			entry.path = op.File.Path
			entry.maxBytes = op.File.MaxBytes
			entry.direction = op.File.Direction
		}
		ops[op.ID] = entry
	}

	want := map[string]struct {
		operation string
		path      string
	}{
		"file upload":  {operation: "freshchat.files.upload", path: "/files/upload"},
		"image upload": {operation: "freshchat.images.upload", path: "/images/upload"},
	}
	for _, cmd := range cliSurface.Commands {
		wantCmd, ok := want[cmd.Path]
		if !ok {
			continue
		}
		delete(want, cmd.Path)
		if cmd.Intent != "direct_write" {
			t.Fatalf("command %q intent = %q, want direct_write", cmd.Path, cmd.Intent)
		}
		if cmd.Availability != "unsupported_local" {
			t.Fatalf("command %q availability = %q, want unsupported_local", cmd.Path, cmd.Availability)
		}
		if cmd.Operation != wantCmd.operation {
			t.Fatalf("command %q operation = %q, want %q", cmd.Path, cmd.Operation, wantCmd.operation)
		}
		if cmd.Notes == "" {
			t.Fatalf("command %q missing notes", cmd.Path)
		}
		op, ok := ops[cmd.Operation]
		if !ok {
			t.Fatalf("command %q references missing operation %q", cmd.Path, cmd.Operation)
		}
		if op.kind != "file_upload" || op.direction != "upload" || op.path != wantCmd.path || op.maxBytes <= 0 || op.approval == "" || op.approval == "none" {
			t.Fatalf("operation %q policy = %+v, want upload path %s with approval", cmd.Operation, op, wantCmd.path)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing Freshchat upload commands: %+v", want)
	}
}
