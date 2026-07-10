package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestGorgiasAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gorgias/api_surface.json")
	if err != nil {
		t.Fatalf("read gorgias api_surface.json: %v", err)
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
		t.Fatalf("unmarshal gorgias api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
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
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 114 {
		t.Fatalf("endpoints = %d, want 114", len(surface.Endpoints))
	}
	if covered != 24 {
		t.Fatalf("covered endpoints = %d, want 24", covered)
	}
	if operations != 90 {
		t.Fatalf("operation endpoints = %d, want 90", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 18,
		"GET":    46,
		"POST":   23,
		"PUT":    27,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"GET": 24,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 18,
		"GET":    22,
		"POST":   23,
		"PUT":    27,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":     27,
		"binary_read":           5,
		"destructive_action":    20,
		"direct_read":           22,
		"disallowed":            1,
		"sensitive_reverse_etl": 15,
	})
}

func TestGorgiasStreamRunnerReadSweep(t *testing.T) {
	streamsRaw, err := os.ReadFile("../../internal/connectors/defs/gorgias/streams.json")
	if err != nil {
		t.Fatalf("read gorgias streams.json: %v", err)
	}
	var streams struct {
		Streams []struct {
			Name string `json:"name"`
			Path string `json:"path"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(streamsRaw, &streams); err != nil {
		t.Fatalf("unmarshal gorgias streams.json: %v", err)
	}

	gotStreams := map[string]string{}
	for _, stream := range streams.Streams {
		gotStreams[stream.Name] = stream.Path
	}
	wantStreams := map[string]string{
		"account_settings":       "/account/settings",
		"custom_fields":          "/custom-fields",
		"customer_custom_fields": "/customers/{{ fanout.id }}/custom-fields",
		"customers":              "/customers",
		"events":                 "/events",
		"integrations":           "/integrations",
		"jobs":                   "/jobs",
		"macros":                 "/macros",
		"messages":               "/messages",
		"metric_cards":           "/metric-cards",
		"rules":                  "/rules",
		"satisfaction_surveys":   "/satisfaction-surveys",
		"tags":                   "/tags",
		"teams":                  "/teams",
		"ticket_custom_fields":   "/tickets/{{ fanout.id }}/custom-fields",
		"ticket_messages":        "/tickets/{{ fanout.id }}/messages",
		"ticket_tags":            "/tickets/{{ fanout.id }}/tags",
		"tickets":                "/tickets",
		"users":                  "/users",
		"view_items":             "/views/{{ fanout.id }}/items",
		"views":                  "/views",
		"voice_call_events":      "/phone/voice-call-events",
		"voice_calls":            "/phone/voice-calls",
		"widgets":                "/widgets",
	}
	assertStringMap(t, "streams", gotStreams, wantStreams)

	surfaceRaw, err := os.ReadFile("../../internal/connectors/defs/gorgias/api_surface.json")
	if err != nil {
		t.Fatalf("read gorgias api_surface.json: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("unmarshal gorgias api_surface.json: %v", err)
	}
	covered := map[string]string{}
	for _, ep := range surface.Endpoints {
		stream, ok := ep.CoveredBy["stream"].(string)
		if !ok || stream == "" {
			continue
		}
		covered[stream] = ep.Method + " " + ep.Path
	}
	wantCovered := map[string]string{
		"account_settings":       "GET /account/settings",
		"custom_fields":          "GET /custom-fields",
		"customer_custom_fields": "GET /customers/{customer_id}/custom-fields",
		"customers":              "GET /customers",
		"events":                 "GET /events",
		"integrations":           "GET /integrations",
		"jobs":                   "GET /jobs",
		"macros":                 "GET /macros",
		"messages":               "GET /messages",
		"metric_cards":           "GET /metric-cards",
		"rules":                  "GET /rules",
		"satisfaction_surveys":   "GET /satisfaction-surveys",
		"tags":                   "GET /tags",
		"teams":                  "GET /teams",
		"ticket_custom_fields":   "GET /tickets/{ticket_id}/custom-fields",
		"ticket_messages":        "GET /tickets/{ticket_id}/messages",
		"ticket_tags":            "GET /tickets/{ticket_id}/tags",
		"tickets":                "GET /tickets",
		"users":                  "GET /users",
		"view_items":             "GET /views/{view_id}/items",
		"views":                  "GET /views",
		"voice_call_events":      "GET /phone/voice-call-events",
		"voice_calls":            "GET /phone/voice-calls",
		"widgets":                "GET /widgets",
	}
	assertStringMap(t, "covered streams", covered, wantCovered)
}

func assertStringMap(t *testing.T, name string, got, want map[string]string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
