package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
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

// WebhookReceiverHeartbeatRequest records an operator-observed tunnel heartbeat.
type WebhookReceiverHeartbeatRequest struct {
	Name          string    `json:"name"`
	CallbackURL   string    `json:"-"`
	ObservedAt    time.Time `json:"observed_at"`
	RecoveryEpoch uint64    `json:"recovery_epoch"`
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
	AllowedPublicPorts     []int                        `json:"allowed_public_ports,omitempty"`
	Status                 webhook.SubscriptionStatus   `json:"status"`
	LastHeartbeatAt        time.Time                    `json:"last_heartbeat_at,omitempty"`
	RecoveryEpoch          uint64                       `json:"recovery_epoch"`
	RecoveryOutcome        synccontract.RecoveryOutcome `json:"recovery_outcome,omitempty"`
	ReregistrationRequired bool                         `json:"reregistration_required"`
	ReconciliationRequired bool                         `json:"reconciliation_required"`
	ReceiptCapacity        int                          `json:"receipt_capacity"`
}

// IngressAdapterStatus is the safe non-webhook status projection for a
// provider-owned pull or stream adapter.
type IngressAdapterStatus struct {
	Name             string                     `json:"name"`
	Mode             webhook.ExposureMode       `json:"mode"`
	AdapterReference string                     `json:"adapter_reference"`
	Status           webhook.SubscriptionStatus `json:"status"`
}

// IngressAdapterStatus returns the safe adapter-only projection of a status.
func (status WebhookReceiverStatus) IngressAdapterStatus() IngressAdapterStatus {
	return IngressAdapterStatus{
		Name:             status.Name,
		Mode:             status.Mode,
		AdapterReference: status.AdapterReference,
		Status:           status.Status,
	}
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
	HandoffLeaseID     string    `json:"handoff_lease_id,omitempty"`
	HandoffLeaseUntil  time.Time `json:"handoff_lease_until,omitempty"`
	ConsumedAt         time.Time `json:"consumed_at,omitempty"`
}

const receiptHandoffLeaseDuration = time.Minute

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
		current := a.snapshotState()
		credential, ok := findCredentialInState(current, req.Credential)
		if !ok {
			return WebhookReceiverStatus{}, errors.New("webhook receiver credential is unavailable")
		}
		identity, err := coordinationIdentityForCredential(current, credential)
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
			return current, fmt.Errorf("webhook receiver configuration: %w", configErr)
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

// RecordWebhookReceiverHeartbeat persists an observed external-tunnel heartbeat
// without probing or changing the external tunnel.
func (a *App) RecordWebhookReceiverHeartbeat(ctx context.Context, req WebhookReceiverHeartbeatRequest) (WebhookReceiverStatus, error) {
	if req.ObservedAt.IsZero() {
		return WebhookReceiverStatus{}, errors.New("webhook receiver heartbeat time is required")
	}
	return a.updateWebhookSubscription(ctx, req.Name, func(subscription *webhook.Subscription, current state) error {
		if _, err := subscription.Heartbeat(req.CallbackURL, req.ObservedAt.UTC(), []byte(current.CoordinationSalt), req.RecoveryEpoch); err != nil {
			return fmt.Errorf("webhook receiver heartbeat: %w", err)
		}
		return nil
	})
}

// CompleteWebhookReceiverReregistration persists a provider lane's completed
// re-registration declaration without invoking a provider API.
func (a *App) CompleteWebhookReceiverReregistration(ctx context.Context, name string, recoveryEpoch uint64) (WebhookReceiverStatus, error) {
	return a.updateWebhookSubscription(ctx, name, func(subscription *webhook.Subscription, _ state) error {
		return subscription.CompleteReregistration(recoveryEpoch)
	})
}

// CompleteWebhookReceiverReconciliation persists a provider lane's completed
// reconciliation declaration without invoking a provider API.
func (a *App) CompleteWebhookReceiverReconciliation(ctx context.Context, name string, recoveryEpoch uint64) (WebhookReceiverStatus, error) {
	return a.updateWebhookSubscription(ctx, name, func(subscription *webhook.Subscription, _ state) error {
		return subscription.CompleteReconciliation(recoveryEpoch)
	})
}

func (a *App) updateWebhookSubscription(ctx context.Context, name string, update func(*webhook.Subscription, state) error) (WebhookReceiverStatus, error) {
	if err := ctx.Err(); err != nil {
		return WebhookReceiverStatus{}, err
	}
	if err := safety.ValidateIdentifier(name, "webhook receiver"); err != nil {
		return WebhookReceiverStatus{}, errors.New("webhook receiver is invalid")
	}
	updated, err := a.updateState(func(current state) (state, error) {
		entry, ok := current.WebhookSubscriptions[name]
		if !ok {
			return current, errors.New("webhook receiver not found")
		}
		if err := update(&entry.Subscription, current); err != nil {
			return current, err
		}
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
func (a *App) WebhookReceiptStore(name string) (webhook.DurableReceiptStore, error) {
	if err := safety.ValidateIdentifier(name, "webhook receiver"); err != nil {
		return nil, errors.New("webhook receiver is invalid")
	}
	a.stateMu.RLock()
	_, ok := a.state.WebhookSubscriptions[name]
	a.stateMu.RUnlock()
	if !ok {
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
		AllowedPublicPorts:     append([]int(nil), entry.Subscription.Exposure.ExternalTunnel.AllowedPublicPorts...),
		Status:                 entry.Subscription.Status,
		LastHeartbeatAt:        entry.Subscription.LastHeartbeatAt,
		RecoveryEpoch:          entry.Subscription.RecoveryEpoch,
		RecoveryOutcome:        entry.Subscription.RecoveryOutcome,
		ReregistrationRequired: entry.Subscription.ReregistrationRequired,
		ReconciliationRequired: entry.Subscription.ReconciliationRequired,
		ReceiptCapacity:        entry.ReceiptCapacity,
	}
}

type appWebhookReceiptStore struct {
	app          *App
	subscription string
}

func (s *appWebhookReceiptStore) Insert(ctx context.Context, receipt webhook.Receipt) (webhook.ReceiptInsertResult, error) {
	if s == nil || s.app == nil {
		return webhook.ReceiptInsertRejected, errors.New("webhook receipt store is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	if !webhook.ValidEventIdentity(receipt.Event.ID) {
		return webhook.ReceiptInsertRejected, errors.New("webhook receipt identity is invalid")
	}
	result := webhook.ReceiptInsertRejected
	_, err := s.app.updateWebhookReceiptState(ctx, func(current state) (state, error) {
		if current.WebhookSubscriptions == nil {
			return current, errors.New("webhook receiver state is unavailable")
		}
		if current.WebhookReceipts == nil {
			current.WebhookReceipts = map[string]webhookReceiptState{}
		}
		fingerprint, err := webhookEventFingerprint(current.CoordinationSalt, receipt.Event.ID)
		if err != nil {
			return current, err
		}
		key := s.subscription + ":" + fingerprint
		if existing, exists := current.WebhookReceipts[key]; exists {
			if err := s.app.validateWebhookReceiptPayload(ctx, existing.EncryptedPayloadID); err != nil {
				return current, fmt.Errorf("recover encrypted webhook receipt: %w", err)
			}
			result = webhook.ReceiptInsertDuplicate
			return current, nil
		}
		entry, ok := current.WebhookSubscriptions[s.subscription]
		if !ok {
			return current, errors.New("webhook receiver not found")
		}
		pending := 0
		for _, existing := range current.WebhookReceipts {
			if existing.Subscription == s.subscription && existing.ConsumedAt.IsZero() {
				pending++
			}
		}
		if pending >= entry.ReceiptCapacity {
			return current, webhook.ErrReceiptBackpressure
		}
		payloadID := "webhook-" + s.subscription + "-" + fingerprint
		created, err := s.app.vault.PutDurableIfAbsent(ctx, payloadID, map[string]string{
			"body_base64": base64.RawStdEncoding.EncodeToString(receipt.RawBody),
			"event_id":    receipt.Event.ID,
		})
		if err != nil {
			return current, fmt.Errorf("persist encrypted webhook receipt: %w", err)
		}
		if !created {
			if err := s.app.validateWebhookReceiptPayload(ctx, payloadID); err != nil {
				return current, fmt.Errorf("recover encrypted webhook receipt: %w", err)
			}
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
	if err := ctx.Err(); err != nil {
		return webhook.ReceiptInsertRejected, err
	}
	return result, nil
}

func (s *appWebhookReceiptStore) Consume(ctx context.Context, eventID string, handoff webhook.DurableReceiptHandoff) (webhook.ReceiptConsumeResult, error) {
	if s == nil || s.app == nil {
		return webhook.ReceiptConsumeRejected, errors.New("webhook receipt store is unavailable")
	}
	if handoff == nil {
		return webhook.ReceiptConsumeRejected, errors.New("webhook receipt handoff is required")
	}
	if !webhook.ValidEventIdentity(eventID) {
		return webhook.ReceiptConsumeRejected, errors.New("webhook receipt identity is invalid")
	}
	if err := ctx.Err(); err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	leaseID, err := newWebhookReceiptHandoffLeaseID()
	if err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	receiptState, claimed, err := s.claimReceiptHandoff(ctx, eventID, leaseID)
	if err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	if claimed == webhook.ReceiptConsumeDuplicate {
		return claimed, nil
	}
	return s.deliverClaimedReceipt(ctx, eventID, receiptState, leaseID, handoff)
}

func (s *appWebhookReceiptStore) ConsumeNext(ctx context.Context, handoff webhook.DurableReceiptHandoff) (webhook.ReceiptConsumeResult, error) {
	if s == nil || s.app == nil {
		return webhook.ReceiptConsumeRejected, errors.New("webhook receipt store is unavailable")
	}
	if handoff == nil {
		return webhook.ReceiptConsumeRejected, errors.New("webhook receipt handoff is required")
	}
	if err := ctx.Err(); err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	leaseID, err := newWebhookReceiptHandoffLeaseID()
	if err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	receiptState, err := s.claimNextReceiptHandoff(ctx, leaseID)
	if err != nil {
		return webhook.ReceiptConsumeRejected, err
	}
	body, eventID, err := s.app.webhookReceiptPayload(ctx, receiptState.EncryptedPayloadID)
	if err != nil {
		return s.rejectClaimedReceipt(ctx, receiptState, leaseID, fmt.Errorf("recover encrypted webhook receipt: %w", err))
	}
	if !webhook.ValidEventIdentity(eventID) {
		return s.rejectClaimedReceipt(ctx, receiptState, leaseID, errors.New("encrypted webhook receipt identity is invalid"))
	}
	if err := s.validateClaimedReceiptEvent(eventID, receiptState); err != nil {
		return s.rejectClaimedReceipt(ctx, receiptState, leaseID, err)
	}
	return s.handoffClaimedReceipt(ctx, eventID, body, receiptState, leaseID, handoff)
}

func (s *appWebhookReceiptStore) deliverClaimedReceipt(ctx context.Context, eventID string, receiptState webhookReceiptState, leaseID string, handoff webhook.DurableReceiptHandoff) (webhook.ReceiptConsumeResult, error) {
	body, err := s.app.webhookReceiptBody(ctx, receiptState.EncryptedPayloadID)
	if err != nil {
		return s.rejectClaimedReceipt(ctx, receiptState, leaseID, fmt.Errorf("recover encrypted webhook receipt: %w", err))
	}
	return s.handoffClaimedReceipt(ctx, eventID, body, receiptState, leaseID, handoff)
}

func (s *appWebhookReceiptStore) handoffClaimedReceipt(ctx context.Context, eventID string, body []byte, receiptState webhookReceiptState, leaseID string, handoff webhook.DurableReceiptHandoff) (webhook.ReceiptConsumeResult, error) {
	if err := handoff(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: eventID},
		RawBody:    body,
		ReceivedAt: receiptState.ReceivedAt,
	}); err != nil {
		return s.rejectClaimedReceipt(ctx, receiptState, leaseID, fmt.Errorf("handoff webhook receipt: %w", err))
	}
	if err := s.completeReceiptHandoff(ctx, receiptState, leaseID); err != nil {
		return webhook.ReceiptConsumeRejected, fmt.Errorf("complete webhook receipt handoff: %w", err)
	}
	return webhook.ReceiptConsumeCompleted, nil
}

func (s *appWebhookReceiptStore) rejectClaimedReceipt(ctx context.Context, receiptState webhookReceiptState, leaseID string, cause error) (webhook.ReceiptConsumeResult, error) {
	if releaseErr := s.releaseReceiptHandoff(ctx, receiptState, leaseID); releaseErr != nil {
		return webhook.ReceiptConsumeRejected, fmt.Errorf("%w (release handoff: %v)", cause, releaseErr)
	}
	return webhook.ReceiptConsumeRejected, cause
}

func (s *appWebhookReceiptStore) claimReceiptHandoff(ctx context.Context, eventID, leaseID string) (webhookReceiptState, webhook.ReceiptConsumeResult, error) {
	now := time.Now().UTC()
	leaseUntil := now.Add(receiptHandoffLeaseDuration)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(leaseUntil) {
		leaseUntil = deadline.UTC()
	}
	claimed := webhookReceiptState{}
	result := webhook.ReceiptConsumeRejected
	_, err := s.app.updateWebhookReceiptState(ctx, func(current state) (state, error) {
		fingerprint, err := webhookEventFingerprint(current.CoordinationSalt, eventID)
		if err != nil {
			return current, err
		}
		key := s.receiptKey(fingerprint)
		existing, ok := current.WebhookReceipts[key]
		if !ok {
			return current, webhook.ErrReceiptNotFound
		}
		if !existing.ConsumedAt.IsZero() {
			result = webhook.ReceiptConsumeDuplicate
			return current, nil
		}
		if !existing.HandoffLeaseUntil.IsZero() && existing.HandoffLeaseUntil.After(now) {
			return current, webhook.ErrReceiptHandoffInProgress
		}
		existing.HandoffLeaseID = leaseID
		existing.HandoffLeaseUntil = leaseUntil
		current.WebhookReceipts[key] = existing
		claimed = existing
		result = webhook.ReceiptConsumeCompleted
		return current, nil
	})
	if err != nil {
		return webhookReceiptState{}, webhook.ReceiptConsumeRejected, err
	}
	return claimed, result, nil
}

func (s *appWebhookReceiptStore) claimNextReceiptHandoff(ctx context.Context, leaseID string) (webhookReceiptState, error) {
	now := time.Now().UTC()
	leaseUntil := now.Add(receiptHandoffLeaseDuration)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(leaseUntil) {
		leaseUntil = deadline.UTC()
	}
	claimed := webhookReceiptState{}
	_, err := s.app.updateWebhookReceiptState(ctx, func(current state) (state, error) {
		candidateKey := ""
		candidate := webhookReceiptState{}
		pending := false
		for key, existing := range current.WebhookReceipts {
			if existing.Subscription != s.subscription || !existing.ConsumedAt.IsZero() {
				continue
			}
			pending = true
			if !existing.HandoffLeaseUntil.IsZero() && existing.HandoffLeaseUntil.After(now) {
				continue
			}
			if candidateKey == "" || existing.ReceivedAt.Before(candidate.ReceivedAt) || (existing.ReceivedAt.Equal(candidate.ReceivedAt) && key < candidateKey) {
				candidateKey = key
				candidate = existing
			}
		}
		if candidateKey == "" {
			if pending {
				return current, webhook.ErrReceiptHandoffInProgress
			}
			return current, webhook.ErrReceiptNotFound
		}
		candidate.HandoffLeaseID = leaseID
		candidate.HandoffLeaseUntil = leaseUntil
		current.WebhookReceipts[candidateKey] = candidate
		claimed = candidate
		return current, nil
	})
	if err != nil {
		return webhookReceiptState{}, err
	}
	return claimed, nil
}

func (s *appWebhookReceiptStore) receiptKey(fingerprint string) string {
	return s.subscription + ":" + fingerprint
}

func (s *appWebhookReceiptStore) releaseReceiptHandoff(ctx context.Context, receiptState webhookReceiptState, leaseID string) error {
	_, err := s.app.updateWebhookReceiptState(ctx, func(current state) (state, error) {
		key := s.receiptKey(receiptState.EventFingerprint)
		existing, ok := current.WebhookReceipts[key]
		if !ok {
			return current, webhook.ErrReceiptNotFound
		}
		if existing.ConsumedAt.IsZero() && existing.HandoffLeaseID == leaseID {
			existing.HandoffLeaseID = ""
			existing.HandoffLeaseUntil = time.Time{}
			current.WebhookReceipts[key] = existing
		}
		return current, nil
	})
	return err
}

func (s *appWebhookReceiptStore) completeReceiptHandoff(ctx context.Context, receiptState webhookReceiptState, leaseID string) error {
	now := time.Now().UTC()
	_, err := s.app.updateWebhookReceiptState(ctx, func(current state) (state, error) {
		key := s.receiptKey(receiptState.EventFingerprint)
		existing, ok := current.WebhookReceipts[key]
		if !ok {
			return current, webhook.ErrReceiptNotFound
		}
		if !existing.ConsumedAt.IsZero() {
			return current, nil
		}
		if existing.HandoffLeaseID != leaseID {
			return current, webhook.ErrReceiptHandoffInProgress
		}
		existing.ConsumedAt = now
		existing.HandoffLeaseID = ""
		existing.HandoffLeaseUntil = time.Time{}
		current.WebhookReceipts[key] = existing
		return current, nil
	})
	return err
}

func (s *appWebhookReceiptStore) validateClaimedReceiptEvent(eventID string, receiptState webhookReceiptState) error {
	current := s.app.snapshotState()
	fingerprint, err := webhookEventFingerprint(current.CoordinationSalt, eventID)
	if err != nil {
		return err
	}
	if fingerprint != receiptState.EventFingerprint {
		return errors.New("encrypted webhook receipt identity is invalid")
	}
	return nil
}

func (a *App) updateWebhookReceiptState(ctx context.Context, update func(state) (state, error)) (state, error) {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(time.Second)
	}
	for {
		if err := ctx.Err(); err != nil {
			return state{}, err
		}
		updated, err := a.updateState(update)
		if !errors.Is(err, os.ErrExist) {
			return updated, err
		}
		if !time.Now().Before(deadline) {
			return updated, err
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return state{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func (a *App) validateWebhookReceiptPayload(ctx context.Context, payloadID string) error {
	_, err := a.webhookReceiptBody(ctx, payloadID)
	return err
}

func (a *App) webhookReceiptBody(ctx context.Context, payloadID string) ([]byte, error) {
	body, _, err := a.webhookReceiptPayload(ctx, payloadID)
	return body, err
}

func (a *App) webhookReceiptPayload(ctx context.Context, payloadID string) ([]byte, string, error) {
	payload, err := a.vault.Get(ctx, payloadID)
	if err != nil {
		return nil, "", err
	}
	body, ok := payload["body_base64"]
	if !ok {
		return nil, "", errors.New("encrypted webhook receipt payload is invalid")
	}
	decoded, err := base64.RawStdEncoding.DecodeString(body)
	if err != nil {
		return nil, "", errors.New("encrypted webhook receipt payload is invalid")
	}
	return decoded, payload["event_id"], nil
}

func newWebhookReceiptHandoffLeaseID() (string, error) {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate webhook receipt handoff lease: %w", err)
	}
	return hex.EncodeToString(data), nil
}

func webhookEventFingerprint(coordinationSalt, eventID string) (string, error) {
	if coordinationSalt == "" {
		return "", errors.New("webhook receipt identity is unavailable")
	}
	mac := hmac.New(sha256.New, []byte(coordinationSalt))
	_, _ = mac.Write([]byte("webhook-event-receipt-v1\x00"))
	_, _ = mac.Write([]byte(eventID))
	return "evt_" + hex.EncodeToString(mac.Sum(nil)[:16]), nil
}
