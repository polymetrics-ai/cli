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
	handoffFailure := errors.New("fixture handoff failure")
	consumed, err := store.Consume(ctx, "event-one", func(context.Context, webhook.Receipt) error {
		return handoffFailure
	})
	if consumed != webhook.ReceiptConsumeRejected || !errors.Is(err, handoffFailure) {
		t.Fatalf("failed Consume() result=%q err=%v", consumed, err)
	}
	if _, err := store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-two"},
		RawBody:    []byte(`{"event":"second"}`),
		ReceivedAt: time.Now().UTC(),
	}); !errors.Is(err, webhook.ErrReceiptBackpressure) {
		t.Fatalf("receipt capacity released before handoff completion: %v", err)
	}
	handoffCalls := 0
	consumed, err = store.Consume(ctx, "event-one", func(_ context.Context, receipt webhook.Receipt) error {
		handoffCalls++
		if receipt.Event.ID != "event-one" || string(receipt.RawBody) != `{"event":"first"}` {
			return errors.New("unexpected handoff receipt")
		}
		return nil
	})
	if err != nil || consumed != webhook.ReceiptConsumeCompleted || handoffCalls != 1 {
		t.Fatalf("completed Consume() result=%q calls=%d err=%v", consumed, handoffCalls, err)
	}
	consumed, err = store.Consume(ctx, "event-one", func(context.Context, webhook.Receipt) error {
		handoffCalls++
		return nil
	})
	if err != nil || consumed != webhook.ReceiptConsumeDuplicate || handoffCalls != 1 {
		t.Fatalf("duplicate Consume() result=%q calls=%d err=%v", consumed, handoffCalls, err)
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
	reopened, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	reopenedStore, err := reopened.WebhookReceiptStore("fixture-receiver")
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err = reopenedStore.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-one"},
		RawBody:    []byte(`{"event":"first"}`),
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil || duplicate != webhook.ReceiptInsertDuplicate {
		t.Fatalf("retained duplicate Insert() result=%q err=%v", duplicate, err)
	}
	second, err := reopenedStore.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-two"},
		RawBody:    []byte(`{"event":"second"}`),
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil || second != webhook.ReceiptInsertNew {
		t.Fatalf("released-capacity Insert() result=%q err=%v", second, err)
	}
	vaultFiles, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if len(vaultFiles) < 3 {
		t.Fatalf("vault entries = %d, want retained encrypted receipts", len(vaultFiles))
	}
}

func TestWebhookReceiptStoreResumesPendingHandoffAfterRestart(t *testing.T) {
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
		Name: "resumable-receiver",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: 1,
	}); err != nil {
		t.Fatal(err)
	}
	store, err := instance.WebhookReceiptStore("resumable-receiver")
	if err != nil {
		t.Fatal(err)
	}
	if result, err := store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-one"},
		RawBody:    []byte(`{"event":"first"}`),
		ReceivedAt: time.Now().UTC(),
	}); err != nil || result != webhook.ReceiptInsertNew {
		t.Fatalf("Insert() result=%q err=%v", result, err)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	resumedStore, err := reopened.WebhookReceiptStore("resumable-receiver")
	if err != nil {
		t.Fatal(err)
	}
	handoffCalls := 0
	result, err := resumedStore.ConsumeNext(ctx, func(_ context.Context, receipt webhook.Receipt) error {
		handoffCalls++
		if receipt.Event.ID != "event-one" || string(receipt.RawBody) != `{"event":"first"}` {
			return errors.New("unexpected resumed receipt")
		}
		return nil
	})
	if err != nil || result != webhook.ReceiptConsumeCompleted || handoffCalls != 1 {
		t.Fatalf("ConsumeNext() result=%q calls=%d err=%v", result, handoffCalls, err)
	}
	if result, err := resumedStore.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-one"},
		RawBody:    []byte(`{"event":"first"}`),
		ReceivedAt: time.Now().UTC(),
	}); err != nil || result != webhook.ReceiptInsertDuplicate {
		t.Fatalf("retained Insert() result=%q err=%v", result, err)
	}
	if result, err := resumedStore.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "event-two"},
		RawBody:    []byte(`{"event":"second"}`),
		ReceivedAt: time.Now().UTC(),
	}); err != nil || result != webhook.ReceiptInsertNew {
		t.Fatalf("released-capacity Insert() result=%q err=%v", result, err)
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
	configured, err := instance.ConfigureWebhookReceiver(ctx, app.ConfigureWebhookReceiverRequest{
		Name: "durable-funnel",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net/receiver",
			HeartbeatTTL: time.Hour,
		},
		ReceiptCapacity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	observedAt := time.Now().UTC().Add(time.Minute)
	heartbeat, err := instance.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "durable-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/receiver",
		ObservedAt:    observedAt,
		RecoveryEpoch: configured.RecoveryEpoch,
	})
	if err != nil {
		t.Fatalf("RecordWebhookReceiverHeartbeat() error = %v", err)
	}
	if !heartbeat.LastHeartbeatAt.Equal(observedAt) || heartbeat.RecoveryEpoch != configured.RecoveryEpoch || len(heartbeat.AllowedPublicPorts) != 3 || heartbeat.AllowedPublicPorts[0] != 443 || heartbeat.AllowedPublicPorts[1] != 8443 || heartbeat.AllowedPublicPorts[2] != 10000 {
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
	if !persisted.LastHeartbeatAt.Equal(observedAt) || persisted.RecoveryEpoch != heartbeat.RecoveryEpoch {
		t.Fatalf("persisted status = %+v, want heartbeat epoch %d", persisted, heartbeat.RecoveryEpoch)
	}
	firstRotation, err := reopened.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "durable-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/rotated",
		ObservedAt:    observedAt.Add(time.Minute),
		RecoveryEpoch: persisted.RecoveryEpoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstRotation.Status != webhook.SubscriptionStatusDegraded || !firstRotation.ReregistrationRequired || !firstRotation.ReconciliationRequired {
		t.Fatalf("first rotation status = %+v", firstRotation)
	}
	secondRotation, err := reopened.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "durable-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/rotated-again",
		ObservedAt:    observedAt.Add(2 * time.Minute),
		RecoveryEpoch: firstRotation.RecoveryEpoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	if secondRotation.RecoveryEpoch != firstRotation.RecoveryEpoch+1 || secondRotation.Status != webhook.SubscriptionStatusDegraded || !secondRotation.ReregistrationRequired || !secondRotation.ReconciliationRequired {
		t.Fatalf("second rotation status = %+v", secondRotation)
	}
	if _, err := reopened.CompleteWebhookReceiverReregistration(ctx, "durable-funnel", firstRotation.RecoveryEpoch); !errors.Is(err, webhook.ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed re-registration error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if _, err := reopened.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "durable-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/rotated",
		ObservedAt:    observedAt.Add(3 * time.Minute),
		RecoveryEpoch: firstRotation.RecoveryEpoch,
	}); !errors.Is(err, webhook.ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed heartbeat error = %v, want ErrStaleRecoveryEpoch", err)
	}
	afterDelayed, err := reopened.WebhookReceiverStatus("durable-funnel", observedAt.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if afterDelayed.RecoveryEpoch != secondRotation.RecoveryEpoch || afterDelayed.EndpointGeneration != secondRotation.EndpointGeneration || afterDelayed.Status != webhook.SubscriptionStatusDegraded || !afterDelayed.ReregistrationRequired || !afterDelayed.ReconciliationRequired {
		t.Fatalf("delayed generation changed persisted state: %+v", afterDelayed)
	}
	partial, err := reopened.CompleteWebhookReceiverReregistration(ctx, "durable-funnel", secondRotation.RecoveryEpoch)
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
	completed, err := afterPartial.CompleteWebhookReceiverReconciliation(ctx, "durable-funnel", secondRotation.RecoveryEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != webhook.SubscriptionStatusActive || completed.ReregistrationRequired || completed.ReconciliationRequired || completed.RecoveryOutcome != "" {
		t.Fatalf("completed recovery status = %+v", completed)
	}
}

func TestWebhookReceiverHeartbeatExpiryPersistsFencedRecovery(t *testing.T) {
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
	configured, err := instance.ConfigureWebhookReceiver(ctx, app.ConfigureWebhookReceiverRequest{
		Name: "expiry-funnel",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net/receiver",
			HeartbeatTTL: time.Minute,
		},
		ReceiptCapacity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	observedAt := time.Now().UTC().Add(time.Second)
	heartbeat, err := instance.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "expiry-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/receiver",
		ObservedAt:    observedAt,
		RecoveryEpoch: configured.RecoveryEpoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	firstExpiry, err := instance.WebhookReceiverStatus("expiry-funnel", observedAt.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if firstExpiry.RecoveryEpoch != heartbeat.RecoveryEpoch+1 || firstExpiry.Status != webhook.SubscriptionStatusDegraded || !firstExpiry.ReregistrationRequired || !firstExpiry.ReconciliationRequired {
		t.Fatalf("first expiry status = %+v", firstExpiry)
	}
	if _, err := instance.CompleteWebhookReceiverReregistration(ctx, "expiry-funnel", firstExpiry.RecoveryEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := instance.CompleteWebhookReceiverReconciliation(ctx, "expiry-funnel", firstExpiry.RecoveryEpoch); err != nil {
		t.Fatal(err)
	}
	refreshed, err := instance.RecordWebhookReceiverHeartbeat(ctx, app.WebhookReceiverHeartbeatRequest{
		Name:          "expiry-funnel",
		CallbackURL:   "https://node.tailnet.ts.net/receiver",
		ObservedAt:    observedAt.Add(3 * time.Minute),
		RecoveryEpoch: firstExpiry.RecoveryEpoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.RecoveryEpoch != firstExpiry.RecoveryEpoch+1 || refreshed.Status != webhook.SubscriptionStatusDegraded || !refreshed.ReregistrationRequired || !refreshed.ReconciliationRequired {
		t.Fatalf("late heartbeat status = %+v", refreshed)
	}
	secondExpiry, err := instance.WebhookReceiverStatus("expiry-funnel", observedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if secondExpiry.RecoveryEpoch != refreshed.RecoveryEpoch || secondExpiry.Status != webhook.SubscriptionStatusDegraded || !secondExpiry.ReregistrationRequired || !secondExpiry.ReconciliationRequired {
		t.Fatalf("second expiry status = %+v", secondExpiry)
	}
	reopened, err := app.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.WebhookReceiverStatus("expiry-funnel", observedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if persisted.RecoveryEpoch != secondExpiry.RecoveryEpoch || persisted.Status != webhook.SubscriptionStatusDegraded || !persisted.ReregistrationRequired || !persisted.ReconciliationRequired {
		t.Fatalf("persisted second expiry status = %+v", persisted)
	}
	if _, err := reopened.CompleteWebhookReceiverReregistration(ctx, "expiry-funnel", firstExpiry.RecoveryEpoch); !errors.Is(err, webhook.ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first expiry re-registration error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if _, err := reopened.CompleteWebhookReceiverReconciliation(ctx, "expiry-funnel", firstExpiry.RecoveryEpoch); !errors.Is(err, webhook.ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first expiry reconciliation error = %v, want ErrStaleRecoveryEpoch", err)
	}
	stillDegraded, err := reopened.WebhookReceiverStatus("expiry-funnel", observedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if stillDegraded.RecoveryEpoch != secondExpiry.RecoveryEpoch || stillDegraded.Status != webhook.SubscriptionStatusDegraded || !stillDegraded.ReregistrationRequired || !stillDegraded.ReconciliationRequired {
		t.Fatalf("delayed completion changed expiry recovery state: %+v", stillDegraded)
	}
}

func TestWebhookReceiverReconfigureDoesNotMaskHeartbeatExpiry(t *testing.T) {
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
		Name: "reconfigure-funnel",
		Exposure: webhook.ExposureConfig{
			Mode:         webhook.ExposureModeExternalTunnel,
			TunnelTool:   webhook.TunnelToolTailscaleFunnel,
			CallbackURL:  "https://node.tailnet.ts.net/receiver",
			HeartbeatTTL: time.Nanosecond,
		},
		ReceiptCapacity: 2,
	}
	configured, err := instance.ConfigureWebhookReceiver(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	updatedRequest := request
	updatedRequest.Exposure.HeartbeatTTL = time.Hour
	reconfigured, err := instance.ConfigureWebhookReceiver(ctx, updatedRequest)
	if err != nil {
		t.Fatal(err)
	}
	if !reconfigured.LastHeartbeatAt.Equal(configured.LastHeartbeatAt) || reconfigured.RecoveryEpoch != configured.RecoveryEpoch+1 || reconfigured.Status != webhook.SubscriptionStatusDegraded || !reconfigured.ReregistrationRequired || !reconfigured.ReconciliationRequired {
		t.Fatalf("same endpoint reconfiguration status = %+v", reconfigured)
	}
}
