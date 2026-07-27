package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestRunReverseETLExecutesWhatsAppStrictWriteSchema(t *testing.T) {
	ctx := context.Background()
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "unexpected method", http.StatusBadRequest)
			return
		}
		if r.URL.Path != "/phone_123/messages" {
			http.Error(w, "unexpected path", http.StatusBadRequest)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "decode body", http.StatusBadRequest)
			return
		}
		requests <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.fixture"}]}`))
	}))
	defer server.Close()

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "whatsapp-local",
		Connector: "whatsapp",
		Config: map[string]string{
			"mode":            "cloud",
			"base_url":        server.URL,
			"phone_number_id": "phone_123",
		},
	}); err != nil {
		t.Fatalf("AddCredential(whatsapp) error = %v", err)
	}
	if err := writeWarehouseRows(t, a, "whatsapp_outbound", []connectors.Record{{
		"phone":        "+12025550123",
		"product":      "whatsapp",
		"message_type": "text",
		"message":      map[string]any{"body": "fixture message body"},
	}}); err != nil {
		t.Fatalf("write whatsapp source rows: %v", err)
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "whatsapp_text_send",
		SourceTable:           "whatsapp_outbound",
		DestinationConnector:  "whatsapp",
		DestinationCredential: "whatsapp-local",
		Action:                "send_text_message",
		Mappings: map[string]string{
			"phone":        "to",
			"product":      "messaging_product",
			"message_type": "type",
			"message":      "text",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  "destructive",
	})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 1 || run.RecordsFailed != 0 {
		t.Fatalf("RunReverseETL() result = %+v, want one completed write", run)
	}

	body := <-requests
	if _, ok := body["_polymetrics_reverse_plan_id"]; ok {
		t.Fatalf("write body included internal reverse plan id: %#v", body)
	}
	if body["messaging_product"] != "whatsapp" || body["to"] != "+12025550123" || body["type"] != "text" {
		t.Fatalf("write body = %#v, want whatsapp text payload", body)
	}
	text, ok := body["text"].(map[string]any)
	if !ok || text["body"] != "fixture message body" {
		t.Fatalf("write text body = %#v, want mapped text body", body["text"])
	}
}
