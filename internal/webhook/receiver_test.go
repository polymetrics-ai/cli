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

	if changed, err := subscription.Heartbeat("https://node.tailnet.ts.net/receiver", now.Add(30*time.Second), key); err != nil || changed {
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
	if subscription.CompleteReregistrationAndReconciliation() != nil {
		t.Fatal("explicit recovery did not restore active status")
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
