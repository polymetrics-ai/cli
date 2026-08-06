package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/webhook"
)

// ConfigureWebhookReceiverRequest configures a provider-neutral ingress
// subscription. CallbackURL is input-only because a callback may contain a
// route secret and is never persisted in ordinary project state.
type ConfigureWebhookReceiverRequest struct {
	Name            string                 `json:"name"`
	Credential      string                 `json:"-"`
	Exposure        webhook.ExposureConfig `json:"-"`
	ReceiptCapacity int                    `json:"receipt_capacity"`
}

// WebhookReceiverStatus is safe to render in CLI text or JSON. It excludes
// callback URLs, event bodies, credentials, signing headers, and secrets.
type WebhookReceiverStatus struct {
	Name                   string                       `json:"name"`
	Mode                   webhook.ExposureMode         `json:"mode"`
	TunnelTool             webhook.TunnelTool           `json:"tunnel_tool,omitempty"`
	AdapterReference       string                       `json:"adapter_reference,omitempty"`
	ListenerScope          webhook.ListenerScope        `json:"listener_scope"`
	EndpointGeneration     string                       `json:"endpoint_generation,omitempty"`
	Status                 webhook.SubscriptionStatus   `json:"status"`
	LastHeartbeatAt        time.Time                    `json:"last_heartbeat_at,omitempty"`
	RecoveryOutcome        synccontract.RecoveryOutcome `json:"recovery_outcome,omitempty"`
	ReregistrationRequired bool                         `json:"reregistration_required"`
	ReconciliationRequired bool                         `json:"reconciliation_required"`
	ReceiptCapacity        int                          `json:"receipt_capacity"`
}

type webhookSubscriptionState struct {
	Subscription     webhook.Subscription     `json:"subscription"`
	CredentialCohort connectors.AuthCohortKey `json:"credential_cohort,omitempty"`
	ReceiptCapacity  int                      `json:"receipt_capacity"`
}

type webhookReceiptState struct {
	Subscription       string    `json:"subscription"`
	EventFingerprint   string    `json:"event_fingerprint"`
	EncryptedPayloadID string    `json:"encrypted_payload_id"`
	ReceivedAt         time.Time `json:"received_at"`
}

// ConfigureWebhookReceiver records the selected ingress mode. It intentionally
// does not call a provider registration endpoint or start a tunnel process.
func (a *App) ConfigureWebhookReceiver(ctx context.Context, req ConfigureWebhookReceiverRequest) (WebhookReceiverStatus, error) {
	if err := ctx.Err(); err != nil {
		return WebhookReceiverStatus{}, err
	}
	if err := safety.ValidateIdentifier(req.Name, "webhook receiver"); err != nil || req.ReceiptCapacity <= 0 {
		return WebhookReceiverStatus{}, errors.New("webhook receiver configuration is invalid")
	}

	var cohort connectors.AuthCohortKey
	if req.Credential != "" {
		credential, ok := a.findCredential(req.Credential)
		if !ok {
			return WebhookReceiverStatus{}, errors.New("webhook receiver credential is unavailable")
		}
		identity, err := a.coordinationIdentityForCredential(credential)
		if err != nil {
			return WebhookReceiverStatus{}, errors.New("webhook receiver credential coordination is unavailable")
		}
		cohort = identity.AuthCohortKey()
	}

	now := time.Now().UTC()
	updated, err := a.updateState(func(current state) (state, error) {
		if current.WebhookSubscriptions == nil {
			current.WebhookSubscriptions = map[string]webhookSubscriptionState{}
		}
		exposure, configErr := webhook.ConfigureExposure(req.Exposure, []byte(current.CoordinationSalt))
		if configErr != nil {
			return current, errors.New("webhook receiver configuration is invalid")
		}
		entry, exists := current.WebhookSubscriptions[req.Name]
		if !exists {
			entry.Subscription = webhook.NewSubscription(req.Name, exposure, now)
		} else {
			entry.Subscription.ApplyExposure(exposure, now)
		}
		entry.CredentialCohort = cohort
		entry.ReceiptCapacity = req.ReceiptCapacity
		current.WebhookSubscriptions[req.Name] = entry
		return current, nil
	})
	if err != nil {
		return WebhookReceiverStatus{}, err
	}
	entry, ok := updated.WebhookSubscriptions[req.Name]
	if !ok {
		return WebhookReceiverStatus{}, errors.New("webhook receiver state is unavailable")
	}
	return webhookStatus(req.Name, entry), nil
}

// WebhookReceiverStatus loads safe receiver state and persists external tunnel
// heartbeat expiry as a visible degraded subscription state.
func (a *App) WebhookReceiverStatus(name string, now time.Time) (WebhookReceiverStatus, error) {
	if err := safety.ValidateIdentifier(name, "webhook receiver"); err != nil {
		return WebhookReceiverStatus{}, errors.New("webhook receiver is invalid")
	}
	updated, err := a.updateState(func(current state) (state, error) {
		entry, ok := current.WebhookSubscriptions[name]
		if !ok {
			return current, errors.New("webhook receiver not found")
		}
		entry.Subscription.DegradeIfHeartbeatExpired(now)
		current.WebhookSubscriptions[name] = entry
		return current, nil
	})
	if err != nil {
		return WebhookReceiverStatus{}, err
	}
	entry, ok := updated.WebhookSubscriptions[name]
	if !ok {
		return WebhookReceiverStatus{}, errors.New("webhook receiver state is unavailable")
	}
	return webhookStatus(name, entry), nil
}

// WebhookReceiptStore returns the encrypted, durable receipt store used by a
// provider verifier/receiver pair. It has no provider-specific parser or
// dispatcher and never exposes stored payload bytes through status APIs.
func (a *App) WebhookReceiptStore(name string) (webhook.ReceiptStore, error) {
	if err := safety.ValidateIdentifier(name, "webhook receiver"); err != nil {
		return nil, errors.New("webhook receiver is invalid")
	}
	if _, ok := a.state.WebhookSubscriptions[name]; !ok {
		return nil, errors.New("webhook receiver not found")
	}
	return &appWebhookReceiptStore{app: a, subscription: name}, nil
}

func webhookStatus(name string, entry webhookSubscriptionState) WebhookReceiverStatus {
	return WebhookReceiverStatus{
		Name:                   name,
		Mode:                   entry.Subscription.Exposure.Mode,
		TunnelTool:             entry.Subscription.Exposure.TunnelTool,
		AdapterReference:       entry.Subscription.Exposure.AdapterReference,
		ListenerScope:          entry.Subscription.Exposure.ListenerScope,
		EndpointGeneration:     entry.Subscription.Exposure.EndpointGeneration,
		Status:                 entry.Subscription.Status,
		LastHeartbeatAt:        entry.Subscription.LastHeartbeatAt,
		RecoveryOutcome:        entry.Subscription.RecoveryOutcome,
		ReregistrationRequired: entry.Subscription.ReregistrationRequired,
		ReconciliationRequired: entry.Subscription.ReconciliationRequired,
		ReceiptCapacity:        entry.ReceiptCapacity,
	}
}

type appWebhookReceiptStore struct {
	app          *App
	subscription string
	mu           sync.Mutex
}

func (s *appWebhookReceiptStore) Insert(ctx context.Context, receipt webhook.Receipt) (webhook.ReceiptInsertResult, error) {
	if s == nil || s.app == nil {
		return webhook.ReceiptInsertRejected, errors.New("webhook receipt store is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	fingerprint, err := s.app.webhookEventFingerprint(receipt.Event.ID)
	if err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	key := s.subscription + ":" + fingerprint
	current, err := s.app.store.Load()
	if err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	if current.WebhookSubscriptions == nil {
		return webhook.ReceiptInsertRejected, errors.New("webhook receiver state is unavailable")
	}
	if _, exists := current.WebhookReceipts[key]; exists {
		return webhook.ReceiptInsertDuplicate, nil
	}
	entry, ok := current.WebhookSubscriptions[s.subscription]
	if !ok {
		return webhook.ReceiptInsertRejected, errors.New("webhook receiver not found")
	}
	pending := 0
	for _, existing := range current.WebhookReceipts {
		if existing.Subscription == s.subscription {
			pending++
		}
	}
	if pending >= entry.ReceiptCapacity {
		return webhook.ReceiptInsertRejected, webhook.ErrReceiptBackpressure
	}

	payloadID := "webhook-" + s.subscription + "-" + fingerprint
	if err := s.app.vault.Put(ctx, payloadID, map[string]string{
		"body_base64": base64.RawStdEncoding.EncodeToString(receipt.RawBody),
	}); err != nil {
		return webhook.ReceiptInsertRejected, fmt.Errorf("persist encrypted webhook receipt: %w", err)
	}

	result := webhook.ReceiptInsertRejected
	_, err = s.app.updateState(func(current state) (state, error) {
		if current.WebhookSubscriptions == nil {
			return current, errors.New("webhook receiver state is unavailable")
		}
		if current.WebhookReceipts == nil {
			current.WebhookReceipts = map[string]webhookReceiptState{}
		}
		if _, exists := current.WebhookReceipts[key]; exists {
			result = webhook.ReceiptInsertDuplicate
			return current, nil
		}
		entry, ok := current.WebhookSubscriptions[s.subscription]
		if !ok {
			return current, errors.New("webhook receiver not found")
		}
		pending := 0
		for _, existing := range current.WebhookReceipts {
			if existing.Subscription == s.subscription {
				pending++
			}
		}
		if pending >= entry.ReceiptCapacity {
			return current, webhook.ErrReceiptBackpressure
		}
		current.WebhookReceipts[key] = webhookReceiptState{
			Subscription:       s.subscription,
			EventFingerprint:   fingerprint,
			EncryptedPayloadID: payloadID,
			ReceivedAt:         receipt.ReceivedAt,
		}
		result = webhook.ReceiptInsertNew
		return current, nil
	})
	if err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	return result, nil
}

func (a *App) webhookEventFingerprint(eventID string) (string, error) {
	if a == nil || a.state.CoordinationSalt == "" {
		return "", errors.New("webhook receipt identity is unavailable")
	}
	mac := hmac.New(sha256.New, []byte(a.state.CoordinationSalt))
	_, _ = mac.Write([]byte("webhook-event-receipt-v1\x00"))
	_, _ = mac.Write([]byte(eventID))
	return "evt_" + hex.EncodeToString(mac.Sum(nil)[:16]), nil
}
