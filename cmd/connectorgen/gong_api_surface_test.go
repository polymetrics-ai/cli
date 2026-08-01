package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGongAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gong/api_surface.json")
	if err != nil {
		t.Fatalf("read gong api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string                `json:"method"`
			Path      string                `json:"path"`
			CoveredBy map[string]any        `json:"covered_by"`
			Excluded  map[string]any        `json:"excluded"`
			Operation *gongSurfaceOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gong api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	operationByKey := map[string]*gongSurfaceOperation{}
	models := map[string]int{}
	covered, excluded, operations := 0, 0, 0
	seen := map[string]bool{}

	for i, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Fatalf("duplicate endpoint %q", key)
		}
		seen[key] = true
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
			operationByKey[key] = ep.Operation
			models[ep.Operation.Model]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if gongRequiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 69 {
		t.Fatalf("endpoints = %d, want 69", len(surface.Endpoints))
	}
	if covered != 68 {
		t.Fatalf("covered endpoints = %d, want 68", covered)
	}
	if covered+operations != 69 {
		t.Fatalf("classified endpoints = %d, want 69", covered+operations)
	}
	if operations != 1 {
		t.Fatalf("operation endpoints = %d, want 1", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertGongStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 3,
		"GET":    29,
		"PATCH":  1,
		"POST":   28,
		"PUT":    8,
	})
	assertGongStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 3,
		"GET":    28,
		"PATCH":  1,
		"POST":   28,
		"PUT":    8,
	})
	assertGongStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{"GET": 1})
	assertGongStringIntMap(t, "models", models, map[string]int{"direct_read": 1})
	crmEntitiesOp, ok := operationByKey["GET /v2/crm/entities"]
	if !ok {
		t.Fatalf("GET /v2/crm/entities should be an explicit blocked operation")
	}
	if !strings.Contains(crmEntitiesOp.Reason, "FORM/explode") || !strings.Contains(crmEntitiesOp.Reason, "objectsCrmIds") {
		t.Fatalf("GET /v2/crm/entities reason = %q, want FORM/explode objectsCrmIds dependency", crmEntitiesOp.Reason)
	}

	for _, key := range []string{
		"POST /v2/calls/extensive",
		"POST /v2/calls/transcript",
		"POST /v2/stats/interaction",
		"GET /v2/targets",
		"POST /v2/targets/{targetId}/assignments",
	} {
		if !seen[key] {
			t.Fatalf("expected official Gong endpoint %q", key)
		}
	}
	for _, key := range []string{
		"GET /v2/calls/extensive",
		"GET /v2/calls/transcript",
		"GET /v2/stats/interaction",
		"GET /v2/stats/activity/trackers",
		"GET /v2/settings/webhooks",
	} {
		if seen[key] {
			t.Fatalf("stale or wrong-method endpoint %q should not be present", key)
		}
	}

	rawOps, err := os.ReadFile("../../internal/connectors/defs/gong/operations.json")
	if err != nil {
		t.Fatalf("read gong operations.json: %v", err)
	}
	var ledger struct {
		Operations []gongOperationSpec `json:"operations"`
	}
	if err := json.Unmarshal(rawOps, &ledger); err != nil {
		t.Fatalf("unmarshal gong operations.json: %v", err)
	}
	if len(ledger.Operations) != 69 {
		t.Fatalf("operations = %d, want 69", len(ledger.Operations))
	}

	operationKinds := map[string]int{}
	operationIDs := map[string]gongOperationSpec{}
	for _, op := range ledger.Operations {
		if _, ok := operationIDs[op.ID]; ok {
			t.Fatalf("duplicate operation id %q", op.ID)
		}
		operationIDs[op.ID] = op
		operationKinds[op.Kind]++
	}
	assertGongStringIntMap(t, "operationKinds", operationKinds, map[string]int{
		"rest_read":  30,
		"rest_write": 27,
		"stream_etl": 12,
	})

	targetsList, ok := operationIDs["gong.list_target_definitions"]
	if !ok {
		t.Fatalf("missing gong.list_target_definitions operation")
	}
	if targetsList.Kind != "rest_read" || targetsList.REST == nil || targetsList.REST.Method != "GET" || targetsList.REST.Path != "/v2/targets" || targetsList.REST.MaxBytes != 1048576 {
		t.Fatalf("target definitions operation = %+v", targetsList)
	}

	targetUpload, ok := operationIDs["gong.upload_assignments"]
	if !ok {
		t.Fatalf("missing gong.upload_assignments operation")
	}
	if targetUpload.Kind != "rest_write" || targetUpload.REST == nil || targetUpload.REST.Method != "POST" || targetUpload.REST.Path != "/v2/targets/{targetId}/assignments" {
		t.Fatalf("target assignments operation = %+v", targetUpload)
	}
	if !targetUpload.Destructive || targetUpload.SensitivePolicy == nil || targetUpload.SensitivePolicy.ApprovalMode != "typed_confirmation" {
		t.Fatalf("target assignments destructive policy = %+v", targetUpload)
	}
	if !containsGongString(targetUpload.SensitivePolicy.RedactFields, "assignments_file_path") || !containsGongString(targetUpload.SensitivePolicy.RedactFields, "assignments_file_content") {
		t.Fatalf("target assignments sensitive redaction = %+v", targetUpload.SensitivePolicy)
	}
}

type gongSurfaceOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
}

type gongOperationSpec struct {
	ID              string                   `json:"id"`
	Kind            string                   `json:"kind"`
	Destructive     bool                     `json:"destructive"`
	REST            *gongRESTOperation       `json:"rest"`
	SensitivePolicy *gongSensitivePolicySpec `json:"sensitive_policy"`
}

type gongRESTOperation struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	MaxBytes int    `json:"max_bytes"`
}

type gongSensitivePolicySpec struct {
	ApprovalMode string   `json:"approval_mode"`
	RedactFields []string `json:"redact_fields"`
}

func containsGongString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func gongRequiresSourceOrNotes(model string) bool {
	switch model {
	case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action", "disallowed":
		return true
	default:
		return false
	}
}

func assertGongStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
