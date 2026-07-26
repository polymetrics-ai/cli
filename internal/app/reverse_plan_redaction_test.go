package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestPlanReverseETLRedactsWhatsAppGenericSendSamples(t *testing.T) {
	ctx := context.Background()
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
		Config:    map[string]string{"mode": "cloud", "phone_number_id": "phone_123"},
	}); err != nil {
		t.Fatalf("AddCredential(whatsapp) error = %v", err)
	}
	if err := writeWarehouseRows(t, a, "whatsapp_outbound", []connectors.Record{{
		"phone":        "+15551234567",
		"product":      "whatsapp",
		"message_type": "text",
		"message":      map[string]any{"body": "your lab result is ready"},
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
	assertWhatsAppTextSampleRedacted(t, plan)

	rawState, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	stateText := string(rawState)
	for _, leak := range []string{"+15551234567", "your lab result is ready"} {
		if strings.Contains(stateText, leak) {
			t.Fatalf("state leaked %q: %s", leak, stateText)
		}
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() after plan error = %v", err)
	}
	stored, err := reopened.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatalf("GetReversePlan() error = %v", err)
	}
	assertWhatsAppTextSampleRedacted(t, stored)
	listed := reopened.ListReversePlans()
	if len(listed) != 1 {
		t.Fatalf("ListReversePlans() length = %d, want 1", len(listed))
	}
	assertWhatsAppTextSampleRedacted(t, listed[0])
}

func TestPlanReverseETLRedactsWhatsAppStatusSamples(t *testing.T) {
	ctx := context.Background()
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
		Config:    map[string]string{"mode": "cloud", "phone_number_id": "phone_123"},
	}); err != nil {
		t.Fatalf("AddCredential(whatsapp) error = %v", err)
	}
	if err := writeWarehouseRows(t, a, "whatsapp_status", []connectors.Record{{
		"product":          "whatsapp",
		"status":           "read",
		"inbound_id":       "wamid.SECRET",
		"typing_indicator": map[string]any{"type": "text"},
	}}); err != nil {
		t.Fatalf("write whatsapp status rows: %v", err)
	}

	markRead, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "whatsapp_mark_read",
		SourceTable:           "whatsapp_status",
		DestinationConnector:  "whatsapp",
		DestinationCredential: "whatsapp-local",
		Action:                "mark_message_read",
		Mappings: map[string]string{
			"product":    "messaging_product",
			"status":     "status",
			"inbound_id": "message_id",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(mark_message_read) error = %v", err)
	}
	assertWhatsAppStatusSampleRedacted(t, markRead, false)

	typing, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "whatsapp_typing",
		SourceTable:           "whatsapp_status",
		DestinationConnector:  "whatsapp",
		DestinationCredential: "whatsapp-local",
		Action:                "send_typing_indicator",
		Mappings: map[string]string{
			"product":          "messaging_product",
			"status":           "status",
			"inbound_id":       "message_id",
			"typing_indicator": "typing_indicator",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(send_typing_indicator) error = %v", err)
	}
	assertWhatsAppStatusSampleRedacted(t, typing, true)

	rawState, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if strings.Contains(string(rawState), "wamid.SECRET") {
		t.Fatalf("state leaked inbound message id: %s", rawState)
	}
}

func assertWhatsAppTextSampleRedacted(t *testing.T, plan app.ReversePlan) {
	t.Helper()
	if len(plan.Sample) != 1 {
		t.Fatalf("sample length = %d, want 1", len(plan.Sample))
	}
	sample := plan.Sample[0]
	if sample["to"] != "***" || sample["text"] != "***" {
		t.Fatalf("sample = %+v, want to/text redacted", sample)
	}
	if sample["messaging_product"] != "whatsapp" || sample["type"] != "text" {
		t.Fatalf("sample = %+v, want provider fields visible", sample)
	}
}

func assertWhatsAppStatusSampleRedacted(t *testing.T, plan app.ReversePlan, wantTypingRedacted bool) {
	t.Helper()
	if len(plan.Sample) != 1 {
		t.Fatalf("sample length = %d, want 1", len(plan.Sample))
	}
	sample := plan.Sample[0]
	if sample["message_id"] != "***" {
		t.Fatalf("sample = %+v, want message_id redacted", sample)
	}
	if wantTypingRedacted && sample["typing_indicator"] != "***" {
		t.Fatalf("sample = %+v, want typing_indicator redacted", sample)
	}
	if sample["messaging_product"] != "whatsapp" || sample["status"] != "read" {
		t.Fatalf("sample = %+v, want provider fields visible", sample)
	}
}
