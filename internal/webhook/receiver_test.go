package webhook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/synccontract"
)

func TestStartLoopbackServesOnlyExternalTunnelMode(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	exposure, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	receiver, err := NewReceiver(ReceiverConfig{
		Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: 1024, MaxInFlight: 1, RequestTimeout: time.Second,
		Verifier: &fixtureVerifier{eventID: "evt-loopback"}, Store: &memoryReceiptStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	running, err := receiver.StartLoopback(exposure, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := running.Shutdown(ctx); err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
	})
	if !strings.HasPrefix(running.Address(), "127.0.0.1:") {
		t.Fatalf("listener address = %q, want loopback", running.Address())
	}
	request, err := http.NewRequest(http.MethodPost, "http://"+running.Address()+"/receiver", strings.NewReader(`{"event":"loopback"}`))
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("loopback response status = %d", response.StatusCode)
	}
	operator, err := ConfigureExposure(ExposureConfig{Mode: ExposureModeOperatorEndpoint, CallbackURL: "https://operator.example.test/receiver"}, key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := receiver.StartLoopback(operator, 0); err == nil {
		t.Fatal("operator endpoint started a local receiver")
	}
}

func TestLoopbackServerShutdownIsIdempotent(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	exposure, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	receiver, err := NewReceiver(ReceiverConfig{
		Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: 1024, MaxInFlight: 1, RequestTimeout: time.Second,
		Verifier: &fixtureVerifier{eventID: "evt-loopback"}, Store: &memoryReceiptStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	running, err := receiver.StartLoopback(exposure, 0)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := running.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown() error = %v", err)
	}
	retryCtx, retryCancel := context.WithTimeout(context.Background(), time.Second)
	defer retryCancel()
	retry := make(chan error, 1)
	go func() {
		retry <- running.Shutdown(retryCtx)
	}()
	select {
	case err := <-retry:
		if err != nil {
			t.Fatalf("second Shutdown() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("second Shutdown() did not return")
	}
}

func TestConfigureExposureModesRemainDistinctAndOpaque(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	tests := []struct {
		name     string
		config   ExposureConfig
		listener ListenerScope
	}{
		{
			name: "operator endpoint does not start a listener",
			config: ExposureConfig{
				Mode:        ExposureModeOperatorEndpoint,
				CallbackURL: "https://operator.example.test/receiver",
			},
			listener: ListenerScopeNone,
		},
		{
			name: "tailscale funnel starts a loopback receiver",
			config: ExposureConfig{
				Mode:         ExposureModeExternalTunnel,
				TunnelTool:   TunnelToolTailscaleFunnel,
				CallbackURL:  "https://node.tailnet.ts.net/receiver",
				HeartbeatTTL: time.Minute,
			},
			listener: ListenerScopeLoopback,
		},
		{
			name: "pull or stream has no callback or listener",
			config: ExposureConfig{
				Mode:             ExposureModeProviderPullOrStream,
				AdapterReference: "provider-event-stream-v1",
			},
			listener: ListenerScopeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exposure, err := ConfigureExposure(tt.config, key)
			if err != nil {
				t.Fatalf("ConfigureExposure() error = %v", err)
			}
			if exposure.ListenerScope != tt.listener {
				t.Fatalf("ListenerScope = %q, want %q", exposure.ListenerScope, tt.listener)
			}
			if tt.config.Mode != ExposureModeProviderPullOrStream && exposure.EndpointGeneration == "" {
				t.Fatal("EndpointGeneration is empty")
			}
			if tt.config.Mode == ExposureModeProviderPullOrStream && exposure.EndpointGeneration != "" {
				t.Fatalf("non-callback mode has endpoint generation %q", exposure.EndpointGeneration)
			}
			if strings.Contains(exposure.EndpointGeneration, "operator.example.test") || strings.Contains(exposure.EndpointGeneration, "node.tailnet.ts.net") {
				t.Fatalf("EndpointGeneration leaked callback metadata: %q", exposure.EndpointGeneration)
			}
		})
	}
}

func TestAtLeastOnceDeliveryStatesWebhookOrderingTruthfully(t *testing.T) {
	t.Parallel()

	delivery := AtLeastOnceDelivery()
	if delivery.Duplicates != "at_least_once" {
		t.Fatalf("duplicates = %q, want at_least_once", delivery.Duplicates)
	}
	if delivery.Ordering != "not_guaranteed" || len(delivery.DedupeKey) != 1 || delivery.DedupeKey[0] != "provider_event_id" {
		t.Fatalf("delivery = %+v, want unordered provider event dedupe", delivery)
	}
}

func TestConfigureExposureRejectsModeBoundaryViolations(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	tests := []ExposureConfig{
		{Mode: ExposureModeOperatorEndpoint, CallbackURL: "http://operator.example.test/receiver"},
		{Mode: ExposureModeExternalTunnel, TunnelTool: "another-tunnel", CallbackURL: "https://node.tailnet.ts.net/receiver", HeartbeatTTL: time.Minute},
		{Mode: ExposureModeExternalTunnel, TunnelTool: TunnelToolTailscaleFunnel, CallbackURL: "https://not-tailscale.example.test/receiver", HeartbeatTTL: time.Minute},
		{Mode: ExposureModeProviderPullOrStream, CallbackURL: "https://operator.example.test/receiver", AdapterReference: "adapter"},
		{Mode: ExposureModeProviderPullOrStream, AdapterReference: ""},
	}
	for _, config := range tests {
		if _, err := ConfigureExposure(config, key); err == nil {
			t.Fatalf("ConfigureExposure(%+v) succeeded", config)
		}
	}
}

func TestExternalTunnelValidatesDeclaredPublicPorts(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	base := ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}
	defaultExposure, err := ConfigureExposure(base, key)
	if err != nil {
		t.Fatal(err)
	}
	if got := formatPorts(defaultExposure.ExternalTunnel.AllowedPublicPorts); got != "443, 8443, 10000" {
		t.Fatalf("default allowed public ports = %q", got)
	}

	base.CallbackURL = "https://node.tailnet.ts.net:8443/receiver"
	if _, err := ConfigureExposure(base, key); err != nil {
		t.Fatalf("ConfigureExposure(default allowed port) error = %v", err)
	}
	base.CallbackURL = "https://node.tailnet.ts.net:4444/receiver"
	if _, err := ConfigureExposure(base, key); err == nil || !strings.Contains(err.Error(), "443, 8443, 10000") {
		t.Fatalf("ConfigureExposure(disallowed default port) error = %v", err)
	}

	custom := base
	custom.ExternalTunnel = ExternalTunnelConfig{AllowedPublicPorts: []int{4444}}
	customExposure, err := ConfigureExposure(custom, key)
	if err != nil {
		t.Fatalf("ConfigureExposure(custom allowed port) error = %v", err)
	}
	now := time.Now().UTC()
	subscription := NewSubscription("fixture", customExposure, now)
	if changed, err := subscription.Heartbeat(custom.CallbackURL, now, key, subscription.RecoveryEpoch); err != nil || changed {
		t.Fatalf("Heartbeat(custom allowed port) changed=%t err=%v", changed, err)
	}
	custom.CallbackURL = "https://node.tailnet.ts.net/receiver"
	if _, err := ConfigureExposure(custom, key); err == nil || !strings.Contains(err.Error(), "4444") {
		t.Fatalf("ConfigureExposure(disallowed custom port) error = %v", err)
	}
}

func TestConfigureExposureCanonicalizesEquivalentCallbackGenerations(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	tests := []struct {
		name       string
		configured ExposureConfig
		equivalent ExposureConfig
	}{
		{
			name: "operator endpoint",
			configured: ExposureConfig{
				Mode:        ExposureModeOperatorEndpoint,
				CallbackURL: "https://OPERATOR.EXAMPLE.TEST:443/receiver",
			},
			equivalent: ExposureConfig{
				Mode:        ExposureModeOperatorEndpoint,
				CallbackURL: "https://operator.example.test/receiver",
			},
		},
		{
			name: "external tunnel",
			configured: ExposureConfig{
				Mode:         ExposureModeExternalTunnel,
				TunnelTool:   TunnelToolTailscaleFunnel,
				CallbackURL:  "https://NODE.TAILNET.TS.NET:443/receiver",
				HeartbeatTTL: time.Minute,
			},
			equivalent: ExposureConfig{
				Mode:         ExposureModeExternalTunnel,
				TunnelTool:   TunnelToolTailscaleFunnel,
				CallbackURL:  "https://node.tailnet.ts.net/receiver",
				HeartbeatTTL: time.Minute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configured, err := ConfigureExposure(tt.configured, key)
			if err != nil {
				t.Fatal(err)
			}
			equivalent, err := ConfigureExposure(tt.equivalent, key)
			if err != nil {
				t.Fatal(err)
			}
			if configured.EndpointGeneration != equivalent.EndpointGeneration {
				t.Fatalf("equivalent endpoint generations = %q and %q", configured.EndpointGeneration, equivalent.EndpointGeneration)
			}
		})
	}
}

func TestSubscriptionLifecycleDegradesOnGenerationRotationAndHeartbeatLoss(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	initial, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC)
	subscription := NewSubscription("fixture", initial, now)

	if changed, err := subscription.Heartbeat("https://node.tailnet.ts.net/receiver", now.Add(30*time.Second), key, subscription.RecoveryEpoch); err != nil || changed {
		t.Fatalf("same endpoint Heartbeat() changed=%t, err=%v", changed, err)
	}
	if subscription.DegradeIfHeartbeatExpired(now.Add(2*time.Minute)) != true {
		t.Fatal("heartbeat expiry did not degrade subscription")
	}
	if subscription.RecoveryOutcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		t.Fatalf("heartbeat RecoveryOutcome = %q", subscription.RecoveryOutcome)
	}
	if subscription.Status != SubscriptionStatusDegraded || !subscription.ReconciliationRequired || !subscription.ReregistrationRequired {
		t.Fatalf("heartbeat expiry state = %+v", subscription)
	}

	rotated, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/new-receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	subscription = NewSubscription("fixture", initial, now)
	if changed := subscription.ApplyExposure(rotated, now.Add(time.Second)); !changed {
		t.Fatal("changed callback URL did not create a new endpoint generation")
	}
	if subscription.Status != SubscriptionStatusDegraded || subscription.RecoveryOutcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		t.Fatalf("rotated endpoint state = %+v", subscription)
	}
	if subscription.CompleteReregistration(subscription.RecoveryEpoch) != nil {
		t.Fatal("explicit re-registration completion failed")
	}
	if subscription.Status != SubscriptionStatusDegraded || subscription.ReregistrationRequired || !subscription.ReconciliationRequired {
		t.Fatalf("partial recovery state = %+v", subscription)
	}
	if subscription.CompleteReconciliation(subscription.RecoveryEpoch) != nil {
		t.Fatal("explicit reconciliation completion failed")
	}
	if subscription.Status != SubscriptionStatusActive || subscription.ReregistrationRequired || subscription.ReconciliationRequired {
		t.Fatalf("explicit recovery state = %+v", subscription)
	}
}

func TestApplyExposureEvaluatesPriorHeartbeatPolicy(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	now := time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC)
	exposure, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	updatedPolicy, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Hour,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	subscription := NewSubscription("fixture", exposure, now)
	initialEpoch := subscription.RecoveryEpoch
	if changed := subscription.ApplyExposure(updatedPolicy, now.Add(2*time.Minute)); changed {
		t.Fatal("same endpoint reconfiguration changed the generation")
	}
	if subscription.Exposure.HeartbeatTTL != time.Hour || !subscription.LastHeartbeatAt.Equal(now) || subscription.RecoveryEpoch != initialEpoch+1 || subscription.Status != SubscriptionStatusDegraded || !subscription.ReregistrationRequired || !subscription.ReconciliationRequired {
		t.Fatalf("same endpoint reconfiguration did not apply the prior liveness policy: %+v", subscription)
	}
	if subscription.DegradeIfHeartbeatExpired(now.Add(3 * time.Minute)) {
		t.Fatal("already degraded reconfiguration started another recovery epoch")
	}
}

func TestHeartbeatExpiryStartsFencedRecoveryEpoch(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	now := time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC)
	exposure, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	subscription := NewSubscription("fixture", exposure, now)
	if !subscription.DegradeIfHeartbeatExpired(now.Add(2 * time.Minute)) {
		t.Fatal("first heartbeat expiry did not start recovery")
	}
	firstRecoveryEpoch := subscription.RecoveryEpoch
	if err := subscription.CompleteReregistration(firstRecoveryEpoch); err != nil {
		t.Fatalf("first recovery re-registration error = %v", err)
	}
	if err := subscription.CompleteReconciliation(firstRecoveryEpoch); err != nil {
		t.Fatalf("first recovery reconciliation error = %v", err)
	}
	if changed, err := subscription.Heartbeat("https://node.tailnet.ts.net/receiver", now.Add(3*time.Minute), key, firstRecoveryEpoch); err != nil || changed {
		t.Fatalf("recovery heartbeat changed=%t err=%v", changed, err)
	}
	secondRecoveryEpoch := subscription.RecoveryEpoch
	if secondRecoveryEpoch != firstRecoveryEpoch+1 {
		t.Fatalf("heartbeat recovery epoch = %d, want %d", secondRecoveryEpoch, firstRecoveryEpoch+1)
	}
	if !subscription.LastHeartbeatAt.Equal(now.Add(3*time.Minute)) || subscription.Status != SubscriptionStatusDegraded || !subscription.ReregistrationRequired || !subscription.ReconciliationRequired {
		t.Fatalf("late heartbeat did not start fenced recovery: %+v", subscription)
	}
	if subscription.DegradeIfHeartbeatExpired(now.Add(5 * time.Minute)) {
		t.Fatal("already degraded subscription started another recovery epoch")
	}
	if subscription.RecoveryEpoch != secondRecoveryEpoch {
		t.Fatalf("degraded recovery epoch = %d, want %d", subscription.RecoveryEpoch, secondRecoveryEpoch)
	}
	if err := subscription.CompleteReregistration(firstRecoveryEpoch); !errors.Is(err, ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first recovery re-registration error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if err := subscription.CompleteReconciliation(firstRecoveryEpoch); !errors.Is(err, ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first recovery reconciliation error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if !subscription.ReregistrationRequired || !subscription.ReconciliationRequired || subscription.Status != SubscriptionStatusDegraded {
		t.Fatalf("delayed first recovery completion changed state: %+v", subscription)
	}
}

func TestSubscriptionFencesDelayedRecoveryGenerations(t *testing.T) {
	t.Parallel()

	key := []byte("test-project-key-material")
	now := time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC)
	initial, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/receiver",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	subscription := NewSubscription("fixture", initial, now)
	first, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/first-rotation",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	if !subscription.ApplyExposure(first, now.Add(time.Second)) {
		t.Fatal("first rotation did not change the exposure")
	}
	firstEpoch := subscription.RecoveryEpoch
	second, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   TunnelToolTailscaleFunnel,
		CallbackURL:  "https://node.tailnet.ts.net/second-rotation",
		HeartbeatTTL: time.Minute,
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	if !subscription.ApplyExposure(second, now.Add(2*time.Second)) {
		t.Fatal("second rotation did not change the exposure")
	}
	secondEpoch := subscription.RecoveryEpoch
	if secondEpoch != firstEpoch+1 {
		t.Fatalf("second recovery epoch = %d, want %d", secondEpoch, firstEpoch+1)
	}
	secondGeneration := subscription.Exposure.EndpointGeneration

	if _, err := subscription.Heartbeat("https://node.tailnet.ts.net/first-rotation", now.Add(3*time.Second), key, firstEpoch); !errors.Is(err, ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first-generation heartbeat error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if subscription.Exposure.EndpointGeneration != secondGeneration || subscription.RecoveryEpoch != secondEpoch {
		t.Fatalf("delayed heartbeat changed state: %+v", subscription)
	}
	if err := subscription.CompleteReregistration(firstEpoch); !errors.Is(err, ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first-generation re-registration error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if err := subscription.CompleteReconciliation(firstEpoch); !errors.Is(err, ErrStaleRecoveryEpoch) {
		t.Fatalf("delayed first-generation reconciliation error = %v, want ErrStaleRecoveryEpoch", err)
	}
	if !subscription.ReregistrationRequired || !subscription.ReconciliationRequired {
		t.Fatalf("delayed completions changed recovery state: %+v", subscription)
	}
	if err := subscription.CompleteReregistration(secondEpoch); err != nil {
		t.Fatalf("current-generation re-registration error = %v", err)
	}
	if err := subscription.CompleteReconciliation(secondEpoch); err != nil {
		t.Fatalf("current-generation reconciliation error = %v", err)
	}
	if subscription.Status != SubscriptionStatusActive {
		t.Fatalf("current-generation completion state = %+v", subscription)
	}
	if _, err := subscription.Heartbeat("https://node.tailnet.ts.net/second-rotation", now.Add(time.Second), key, secondEpoch); !errors.Is(err, ErrStaleHeartbeat) {
		t.Fatalf("out-of-order heartbeat error = %v, want ErrStaleHeartbeat", err)
	}
}

func TestReceiverVerifiesRawBodyAndPersistsBeforeAcknowledging(t *testing.T) {
	t.Parallel()

	store := &memoryReceiptStore{}
	verifier := &fixtureVerifier{eventID: "evt-1"}
	receiver, err := NewReceiver(ReceiverConfig{
		Method:         http.MethodPost,
		Path:           "/receiver",
		MaxBodyBytes:   1024,
		MaxInFlight:    1,
		RequestTimeout: time.Second,
		Verifier:       verifier,
		Store:          store,
	})
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{"z":1,"a":2}`))
	response := &acknowledgementWriter{store: store}
	receiver.ServeHTTP(response, request)
	if response.status != http.StatusOK {
		t.Fatalf("response status = %d, want %d", response.status, http.StatusOK)
	}
	if !verifier.called || verifier.rawBody != `{"z":1,"a":2}` {
		t.Fatalf("verifier called=%t body=%q", verifier.called, verifier.rawBody)
	}
	if len(store.receipts) != 1 || string(store.receipts[0].RawBody) != `{"z":1,"a":2}` {
		t.Fatalf("durable receipt = %+v", store.receipts)
	}
	if !store.persistedBeforeAcknowledgement || !response.acknowledgedAfterPersist {
		t.Fatal("receiver acknowledged before durable receipt")
	}

	duplicateResponse := httptest.NewRecorder()
	receiver.ServeHTTP(duplicateResponse, httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{"z":1,"a":2}`)))
	if duplicateResponse.Code != http.StatusOK || len(store.receipts) != 1 {
		t.Fatalf("duplicate response=%d receipts=%d", duplicateResponse.Code, len(store.receipts))
	}
}

func TestReceiverAcceptsOutOfOrderAndDuplicateDeliveries(t *testing.T) {
	t.Parallel()

	store := &memoryReceiptStore{}
	receiver, err := NewReceiver(ReceiverConfig{
		Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: 1024, MaxInFlight: 2, RequestTimeout: time.Second,
		Verifier: eventIDHeaderVerifier{}, Store: store,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, eventID := range []string{"evt-later", "evt-earlier", "evt-later"} {
		request := httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{"event":"fixture"}`))
		request.Header.Set("X-Provider-Event-ID", eventID)
		response := httptest.NewRecorder()
		receiver.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("event %q status = %d, want %d", eventID, response.Code, http.StatusOK)
		}
	}
	if len(store.receipts) != 2 || store.receipts[0].Event.ID != "evt-later" || store.receipts[1].Event.ID != "evt-earlier" {
		t.Fatalf("out-of-order receipts = %+v", store.receipts)
	}
}

func TestReceiverRejectsRequestsBeyondInFlightLimit(t *testing.T) {
	t.Parallel()

	verifier := &blockingVerifier{started: make(chan struct{}), release: make(chan struct{})}
	receiver, err := NewReceiver(ReceiverConfig{
		Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: 1024, MaxInFlight: 1, RequestTimeout: time.Second,
		Verifier: verifier, Store: &memoryReceiptStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response := httptest.NewRecorder()
		receiver.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{}`)))
		firstDone <- response
	}()
	select {
	case <-verifier.started:
	case <-time.After(time.Second):
		t.Fatal("first request did not enter verifier")
	}
	second := httptest.NewRecorder()
	receiver.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{}`)))
	if second.Code != http.StatusServiceUnavailable {
		t.Fatalf("bounded request status = %d, want %d", second.Code, http.StatusServiceUnavailable)
	}
	close(verifier.release)
	select {
	case first := <-firstDone:
		if first.Code != http.StatusOK {
			t.Fatalf("first request status = %d, want %d", first.Code, http.StatusOK)
		}
	case <-time.After(time.Second):
		t.Fatal("first request did not complete")
	}
}

func TestReceiverRejectsVerificationThatExceedsRequestTimeout(t *testing.T) {
	t.Parallel()

	receiver, err := NewReceiver(ReceiverConfig{
		Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: 1024, MaxInFlight: 1, RequestTimeout: 10 * time.Millisecond,
		Verifier: timeoutVerifier{}, Store: &memoryReceiptStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	receiver.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(`{}`)))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("timeout status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestReceiverRejectsUnverifiedOversizedAndBackpressuredRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		verifier   *fixtureVerifier
		store      *memoryReceiptStore
		body       string
		maxBody    int64
		wantStatus int
	}{
		{name: "unverified", verifier: &fixtureVerifier{err: errors.New("fixture rejection")}, store: &memoryReceiptStore{}, body: `{}`, maxBody: 32, wantStatus: http.StatusUnauthorized},
		{name: "oversized", verifier: &fixtureVerifier{eventID: "evt-2"}, store: &memoryReceiptStore{}, body: strings.Repeat("x", 33), maxBody: 32, wantStatus: http.StatusRequestEntityTooLarge},
		{name: "backpressured", verifier: &fixtureVerifier{eventID: "evt-3"}, store: &memoryReceiptStore{capacityErr: true}, body: `{}`, maxBody: 32, wantStatus: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver, err := NewReceiver(ReceiverConfig{
				Method: http.MethodPost, Path: "/receiver", MaxBodyBytes: tt.maxBody, MaxInFlight: 1, RequestTimeout: time.Second, Verifier: tt.verifier, Store: tt.store,
			})
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			receiver.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/receiver", strings.NewReader(tt.body)))
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if len(tt.store.receipts) != 0 {
				t.Fatalf("rejected request persisted %d receipts", len(tt.store.receipts))
			}
		})
	}
}

type fixtureVerifier struct {
	eventID string
	err     error
	called  bool
	rawBody string
}

type eventIDHeaderVerifier struct{}

func (eventIDHeaderVerifier) Verify(_ context.Context, _ []byte, headers http.Header) (VerifiedEvent, error) {
	return VerifiedEvent{ID: headers.Get("X-Provider-Event-ID")}, nil
}

type blockingVerifier struct {
	started chan struct{}
	release chan struct{}
}

type timeoutVerifier struct{}

func (timeoutVerifier) Verify(ctx context.Context, _ []byte, _ http.Header) (VerifiedEvent, error) {
	<-ctx.Done()
	return VerifiedEvent{}, ctx.Err()
}

func (v *blockingVerifier) Verify(ctx context.Context, _ []byte, _ http.Header) (VerifiedEvent, error) {
	close(v.started)
	select {
	case <-v.release:
		return VerifiedEvent{ID: "evt-blocked"}, nil
	case <-ctx.Done():
		return VerifiedEvent{}, ctx.Err()
	}
}

type acknowledgementWriter struct {
	header                   http.Header
	status                   int
	store                    *memoryReceiptStore
	acknowledgedAfterPersist bool
}

func (w *acknowledgementWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *acknowledgementWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return len(body), nil
}

func (w *acknowledgementWriter) WriteHeader(status int) {
	w.status = status
	w.acknowledgedAfterPersist = w.store.persistedBeforeAcknowledgement
}

func (v *fixtureVerifier) Verify(_ context.Context, rawBody []byte, _ http.Header) (VerifiedEvent, error) {
	v.called = true
	v.rawBody = string(rawBody)
	if v.err != nil {
		return VerifiedEvent{}, v.err
	}
	return VerifiedEvent{ID: v.eventID}, nil
}

type memoryReceiptStore struct {
	mu                             sync.Mutex
	receipts                       []Receipt
	ids                            map[string]struct{}
	capacityErr                    bool
	persistedBeforeAcknowledgement bool
}

func (s *memoryReceiptStore) Insert(_ context.Context, receipt Receipt) (ReceiptInsertResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ids == nil {
		s.ids = make(map[string]struct{})
	}
	if _, exists := s.ids[receipt.Event.ID]; exists {
		return ReceiptInsertDuplicate, nil
	}
	if s.capacityErr {
		return ReceiptInsertRejected, ErrReceiptBackpressure
	}
	s.ids[receipt.Event.ID] = struct{}{}
	s.receipts = append(s.receipts, receipt)
	s.persistedBeforeAcknowledgement = true
	return ReceiptInsertNew, nil
}
