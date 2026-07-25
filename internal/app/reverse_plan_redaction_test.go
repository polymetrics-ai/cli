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
