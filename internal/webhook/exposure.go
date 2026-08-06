// Package webhook provides the provider-neutral ingress contract for webhook
// receivers. Provider lanes supply signature verification, event identity, and
// subscription registration; this package deliberately supplies none of them.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// ExposureMode is the closed operator-visible ingress configuration.
type ExposureMode string

const (
	ExposureModeOperatorEndpoint     ExposureMode = "operator_endpoint"
	ExposureModeExternalTunnel       ExposureMode = "external_tunnel"
	ExposureModeProviderPullOrStream ExposureMode = "provider_pull_or_stream"
)

// TunnelTool identifies the external tool the operator has already started.
// It is not a command path and is never executed by pm.
type TunnelTool string

const (
	TunnelToolTailscaleFunnel TunnelTool = "tailscale_funnel"
)

// ListenerScope states whether pm owns a local HTTP listener for an exposure.
type ListenerScope string

const (
	ListenerScopeNone     ListenerScope = "none"
	ListenerScopeLoopback ListenerScope = "loopback"
)

// ExposureConfig contains one operator-supplied configuration attempt. The
// callback URL is intentionally absent from Exposure, the durable/output-safe
// result of this constructor.
type ExposureConfig struct {
	Mode             ExposureMode
	TunnelTool       TunnelTool
	CallbackURL      string
	AdapterReference string
	HeartbeatTTL     time.Duration
}

// Exposure is the output-safe representation persisted for a subscription.
// EndpointGeneration is an opaque keyed fingerprint, never a callback URL.
type Exposure struct {
	Mode               ExposureMode
	TunnelTool         TunnelTool
	AdapterReference   string
	ListenerScope      ListenerScope
	EndpointGeneration string
	HeartbeatTTL       time.Duration
}

// AtLeastOnceDelivery is the ingress-side delivery truth that provider
// changefeed declarations must carry when they use this receiver. Webhook
// deliveries may be duplicated and may arrive out of order; a provider lane
// still owns its provider-specific delete semantics and event identity.
func AtLeastOnceDelivery() connectors.ChangefeedDelivery {
	return connectors.ChangefeedDelivery{
		Ordering:   "not_guaranteed",
		Duplicates: "at_least_once",
		Deletes:    "provider_declared",
		DedupeKey:  []string{"provider_event_id"},
	}
}

// SubscriptionStatus states whether an existing provider registration can be
// trusted to reach the receiver. It says nothing about provider-side mutation.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusDegraded SubscriptionStatus = "degraded"
)

// Subscription holds provider-neutral lifecycle state. Its recovery field is
// deliberately the #3810 outcome type rather than a webhook-local duplicate.
type Subscription struct {
	Name                   string
	Exposure               Exposure
	Status                 SubscriptionStatus
	LastHeartbeatAt        time.Time
	RecoveryOutcome        synccontract.RecoveryOutcome
	ReregistrationRequired bool
	ReconciliationRequired bool
}

// ConfigureExposure validates one of the three declared modes and returns
// only safe, durable metadata. endpointKey must be project-stable protected
// material and must not be derived from a provider credential or signing key.
func ConfigureExposure(config ExposureConfig, endpointKey []byte) (Exposure, error) {
	if len(endpointKey) < 16 {
		return Exposure{}, errors.New("endpoint generation key is invalid")
	}
	switch config.Mode {
	case ExposureModeOperatorEndpoint:
		callback, err := parseHTTPSCallback(config.CallbackURL)
		if err != nil {
			return Exposure{}, err
		}
		if config.TunnelTool != "" || config.AdapterReference != "" || config.HeartbeatTTL != 0 {
			return Exposure{}, errors.New("operator endpoint configuration contains unsupported fields")
		}
		return Exposure{
			Mode:               config.Mode,
			ListenerScope:      ListenerScopeNone,
			EndpointGeneration: endpointGeneration(endpointKey, config.Mode, "", callback.String()),
		}, nil
	case ExposureModeExternalTunnel:
		if config.TunnelTool != TunnelToolTailscaleFunnel {
			return Exposure{}, errors.New("external tunnel tool is unsupported")
		}
		if config.AdapterReference != "" || config.HeartbeatTTL <= 0 {
			return Exposure{}, errors.New("external tunnel requires a positive heartbeat timeout and no adapter reference")
		}
		callback, err := parseHTTPSCallback(config.CallbackURL)
		if err != nil {
			return Exposure{}, err
		}
		if !strings.HasSuffix(strings.ToLower(callback.Hostname()), ".ts.net") {
			return Exposure{}, errors.New("tailscale funnel callback host is invalid")
		}
		return Exposure{
			Mode:               config.Mode,
			TunnelTool:         config.TunnelTool,
			ListenerScope:      ListenerScopeLoopback,
			EndpointGeneration: endpointGeneration(endpointKey, config.Mode, string(config.TunnelTool), callback.String()),
			HeartbeatTTL:       config.HeartbeatTTL,
		}, nil
	case ExposureModeProviderPullOrStream:
		if config.CallbackURL != "" || config.TunnelTool != "" || config.HeartbeatTTL != 0 {
			return Exposure{}, errors.New("provider pull or stream does not accept a callback or tunnel")
		}
		if !safeReference(config.AdapterReference) {
			return Exposure{}, errors.New("provider pull or stream adapter reference is invalid")
		}
		return Exposure{
			Mode:             config.Mode,
			AdapterReference: config.AdapterReference,
			ListenerScope:    ListenerScopeNone,
		}, nil
	default:
		return Exposure{}, errors.New("webhook exposure mode is unsupported")
	}
}

// NewSubscription creates active lifecycle state after an operator has made
// the configuration choice. It does not register a provider subscription.
func NewSubscription(name string, exposure Exposure, now time.Time) Subscription {
	return Subscription{
		Name:            name,
		Exposure:        exposure,
		Status:          SubscriptionStatusActive,
		LastHeartbeatAt: now,
	}
}

// Heartbeat records an operator-observed external-tunnel heartbeat. It never
// executes or probes a tunnel. A changed callback URL is a new generation and
// therefore degrades the subscription until the provider lane confirms both
// re-registration and reconciliation.
func (s *Subscription) Heartbeat(callbackURL string, now time.Time, endpointKey []byte) (bool, error) {
	if s == nil || s.Exposure.Mode != ExposureModeExternalTunnel {
		return false, errors.New("subscription does not use an external tunnel")
	}
	next, err := ConfigureExposure(ExposureConfig{
		Mode:         ExposureModeExternalTunnel,
		TunnelTool:   s.Exposure.TunnelTool,
		CallbackURL:  callbackURL,
		HeartbeatTTL: s.Exposure.HeartbeatTTL,
	}, endpointKey)
	if err != nil {
		return false, err
	}
	changed := s.ApplyExposure(next, now)
	if !changed {
		s.LastHeartbeatAt = now
	}
	return changed, nil
}

// ApplyExposure records a changed generation as degraded. It never calls a
// provider API; provider lanes own the subsequent registration mutation.
func (s *Subscription) ApplyExposure(next Exposure, now time.Time) bool {
	if s == nil {
		return false
	}
	changed := s.Exposure.Mode != next.Mode || s.Exposure.EndpointGeneration != next.EndpointGeneration
	s.Exposure = next
	s.LastHeartbeatAt = now
	if changed {
		s.degradeForEndpointChange()
	}
	return changed
}

// DegradeIfHeartbeatExpired makes an unobserved tunnel outage visible. It is
// meaningful only for external_tunnel and is intentionally idempotent.
func (s *Subscription) DegradeIfHeartbeatExpired(now time.Time) bool {
	if s == nil || s.Exposure.Mode != ExposureModeExternalTunnel || s.Status == SubscriptionStatusDegraded {
		return false
	}
	if now.Before(s.LastHeartbeatAt.Add(s.Exposure.HeartbeatTTL)) {
		return false
	}
	s.degradeForEndpointChange()
	return true
}

// CompleteReregistrationAndReconciliation records that a provider lane has
// explicitly completed both required actions. Calling it cannot itself perform
// provider registration, replay, or polling.
func (s *Subscription) CompleteReregistrationAndReconciliation() error {
	if s == nil {
		return errors.New("subscription is required")
	}
	if !s.ReregistrationRequired || !s.ReconciliationRequired {
		return errors.New("subscription does not require recovery")
	}
	s.Status = SubscriptionStatusActive
	s.RecoveryOutcome = ""
	s.ReregistrationRequired = false
	s.ReconciliationRequired = false
	return nil
}

func (s *Subscription) degradeForEndpointChange() {
	s.Status = SubscriptionStatusDegraded
	s.RecoveryOutcome = synccontract.RecoveryOutcomeSourceGenerationChanged
	s.ReregistrationRequired = true
	s.ReconciliationRequired = true
}

func parseHTTPSCallback(raw string) (*url.URL, error) {
	callback, err := url.Parse(raw)
	if err != nil || callback == nil || callback.Scheme != "https" || callback.Host == "" || callback.User != nil || callback.Fragment != "" {
		return nil, errors.New("callback endpoint must be an absolute HTTPS URL")
	}
	return callback, nil
}

func endpointGeneration(key []byte, mode ExposureMode, tool, callback string) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("webhook-endpoint-generation-v1\x00"))
	_, _ = mac.Write([]byte(mode))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(tool))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(callback))
	return "epg_" + hex.EncodeToString(mac.Sum(nil)[:16])
}

func safeReference(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if unicode.IsLower(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return !strings.Contains(value, "..")
}

func (m ExposureMode) String() string { return string(m) }

func (s ListenerScope) String() string { return string(s) }

func (s SubscriptionStatus) String() string { return string(s) }

var _ fmt.Stringer = ExposureMode("")
var _ fmt.Stringer = ListenerScope("")
var _ fmt.Stringer = SubscriptionStatus("")
