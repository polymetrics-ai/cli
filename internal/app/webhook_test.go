package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/webhook"
)

func TestWebhookReceiverPersistsOnlyOpaqueExposureAndEncryptedReceipts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}

	status, err := instance.ConfigureWebhookReceiver(ctx, app.ConfigureWebhookReceiverRequest{
		Name: "fixture-receiver",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: 1,
	})
	if err != nil {
		t.Fatalf("ConfigureWebhookReceiver() error = %v", err)
	}
	if status.Mode != webhook.ExposureModeOperatorEndpoint || status.ListenerScope != webhook.ListenerScopeNone || status.EndpointGeneration == "" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if strings.Contains(status.EndpointGeneration, "operator.example.test") {
		t.Fatalf("endpoint generation leaked callback URL: %q", status.EndpointGeneration)
	}

	store, err := instance.WebhookReceiptStore("fixture-receiver")
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-one"},
		RawBody:    []byte(`{"event":"first"}`),
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil || first != webhook.ReceiptInsertNew {
		t.Fatalf("first Insert() result=%q err=%v", first, err)
	}
	duplicate, err := store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-one"},
		RawBody:    []byte(`{"event":"first"}`),
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil || duplicate != webhook.ReceiptInsertDuplicate {
		t.Fatalf("duplicate Insert() result=%q err=%v", duplicate, err)
	}
	vaultFilesBeforeReject, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-two"},
		RawBody:    []byte(`{"event":"second"}`),
		ReceivedAt: time.Now().UTC(),
	})
	if !errors.Is(err, webhook.ErrReceiptBackpressure) {
		t.Fatalf("full receipt store error = %v, want ErrReceiptBackpressure", err)
	}
	vaultFilesAfterReject, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if len(vaultFilesAfterReject) != len(vaultFilesBeforeReject) {
		t.Fatalf("rejected receipt added encrypted payload: before=%d after=%d", len(vaultFilesBeforeReject), len(vaultFilesAfterReject))
	}

	stateBytes, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	state := string(stateBytes)
	for _, forbidden := range []string{"operator.example.test", `{"event":"first"}`, "event-one"} {
		if strings.Contains(state, forbidden) {
			t.Fatalf("ordinary state leaked protected ingress value %q", forbidden)
		}
	}
	vaultFiles, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if len(vaultFiles) < 2 { // key plus one encrypted receipt
		t.Fatalf("vault entries = %d, want encrypted receipt", len(vaultFiles))
	}
}

func TestWebhookReceiverExternalTunnelRequiresReconciliationAfterRotation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	request := app.ConfigureWebhookReceiverRequest{
		Name: "funnel-receiver",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net/receiver",
			HeartbeatTTL: time.Minute,
		},
		ReceiptCapacity: 2,
	}
	if _, err := instance.ConfigureWebhookReceiver(ctx, request); err != nil {
		t.Fatal(err)
	}
	request.Exposure.CallbackURL = "https://node.tailnet.ts.net/rotated"
	rotated, err := instance.ConfigureWebhookReceiver(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.Status != webhook.SubscriptionStatusDegraded || !rotated.ReregistrationRequired || !rotated.ReconciliationRequired {
		t.Fatalf("rotated receiver status = %+v", rotated)
	}
	if rotated.RecoveryOutcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		t.Fatalf("rotated recovery outcome = %q", rotated.RecoveryOutcome)
	}
}

func TestWebhookReceiverPersistsHeartbeatAndSeparateRecoveryCompletion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := instance.ConfigureWebhookReceiver(ctx, app.ConfigureWebhookReceiverRequest{
		Name: "durable-funnel",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net/receiver",
			HeartbeatTTL: time.Hour,
		},
		ReceiptCapacity: 2,
	}); err != nil {
		t.Fatal(err)
	}
	observedAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	heartbeat, err := instance.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:        "durable-funnel",
		CallbackURL: "https://node.tailnet.ts.net/receiver",
		ObservedAt:  observedAt,
	})
	if err != nil {
		t.Fatalf("RecordWebhookReceiverHeartbeat() error = %v", err)
	}
	if !heartbeat.LastHeartbeatAt.Equal(observedAt) || len(heartbeat.AllowedPublicPorts) != 3 || heartbeat.AllowedPublicPorts[0] != 443 || heartbeat.AllowedPublicPorts[1] != 8443 || heartbeat.AllowedPublicPorts[2] != 10000 {
		t.Fatalf("heartbeat status = %+v", heartbeat)
	}
	if _, err := instance.ConfigureWebhookReceiver(ctx, app.ConfigureWebhookReceiverRequest{
		Name: "invalid-port",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net:4444/receiver",
			HeartbeatTTL: time.Hour,
		},
		ReceiptCapacity: 2,
	}); err == nil || !strings.Contains(err.Error(), "443, 8443, 10000") {
		t.Fatalf("ConfigureWebhookReceiver(disallowed port) error = %v", err)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.WebhookReceiverStatus("durable-funnel", observedAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !persisted.LastHeartbeatAt.Equal(observedAt) {
		t.Fatalf("persisted heartbeat = %s, want %s", persisted.LastHeartbeatAt, observedAt)
	}
	rotated, err := reopened.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:        "durable-funnel",
		CallbackURL: "https://node.tailnet.ts.net/rotated",
		ObservedAt:  observedAt.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if rotated.Status != webhook.SubscriptionStatusDegraded || !rotated.ReregistrationRequired || !rotated.ReconciliationRequired {
		t.Fatalf("rotated heartbeat status = %+v", rotated)
	}
	partial, err := reopened.CompleteWebhookReceiverReregistration(ctx, "durable-funnel")
	if err != nil {
		t.Fatal(err)
	}
	if partial.Status != webhook.SubscriptionStatusDegraded || partial.ReregistrationRequired || !partial.ReconciliationRequired {
		t.Fatalf("partial recovery status = %+v", partial)
	}

	afterPartial, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := afterPartial.CompleteWebhookReceiverReconciliation(ctx, "durable-funnel")
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != webhook.SubscriptionStatusActive || completed.ReregistrationRequired || completed.ReconciliationRequired || completed.RecoveryOutcome != "" {
		t.Fatalf("completed recovery status = %+v", completed)
	}
}
