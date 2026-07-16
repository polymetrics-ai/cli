package outbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	authoritystore "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
)

type ControllerFacts struct {
	DeliveryID  string
	Repository  string
	Issue       int
	PullRequest int
	Generation  int64
	HeadSHA     string
	Owner       string
	Epoch       int64
}

type Controller struct {
	facts     ControllerFacts
	authority *authoritystore.Store
	lease     authoritystore.Lease
}

func NewController(ctx context.Context, authority *authoritystore.Store, lease authoritystore.Lease,
	facts ControllerFacts, now time.Time) (*Controller, error) {
	if authority == nil || !safeDeliveryID(facts.DeliveryID) || validateTarget(Target{
		Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest,
	}) != nil || facts.Generation <= 0 || !validGitSHA(facts.HeadSHA) || !safeID(facts.Owner) || facts.Epoch <= 0 ||
		lease.RunID != facts.DeliveryID || lease.Owner != facts.Owner || lease.Epoch != facts.Epoch || now.IsZero() {
		return nil, errors.New("complete store-issued fenced controller facts are required")
	}
	if err := authority.CheckLease(ctx, lease, now); err != nil {
		return nil, fmt.Errorf("verify controller authority lease: %w", err)
	}
	controller := &Controller{facts: facts, authority: authority, lease: lease}
	if err := controller.revalidateFacts(ctx); err != nil {
		return nil, err
	}
	return controller, nil
}

type Grant struct {
	id         string
	effectID   string
	capability Capability
	owner      string
	epoch      int64
	issuedAt   time.Time
}

type Authorization struct {
	intent Intent
	grant  Grant
}

func (c *Controller) Authorize(ctx context.Context, intent Intent, now time.Time) (Authorization, error) {
	return c.authorize(ctx, intent, now, false)
}

func (c *Controller) authorize(ctx context.Context, intent Intent, now time.Time,
	allowExpiredReconciliation bool) (Authorization, error) {
	if c == nil || now.IsZero() {
		return Authorization{}, errors.New("controller and authorization time are required")
	}
	if err := c.Revalidate(ctx, now); err != nil {
		return Authorization{}, err
	}
	if err := c.matches(intent); err != nil {
		return Authorization{}, err
	}
	capability, err := capabilityForKind(intent.kind)
	if err != nil {
		return Authorization{}, err
	}
	if intent.kind == KindDecisionQuestion {
		var payload QuestionPayload
		if err := decodeStrict(intent.payload, &payload); err != nil {
			return Authorization{}, err
		}
		expiresAt, err := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
		if err != nil || (!allowExpiredReconciliation && !expiresAt.After(now)) {
			return Authorization{}, errors.New("expired question effect cannot be authorized")
		}
	}
	grantIdentity := strings.Join([]string{
		intent.effectID, string(capability), c.facts.DeliveryID, c.facts.Repository,
		fmt.Sprintf("%d", c.facts.Issue), fmt.Sprintf("%d", c.facts.PullRequest),
		fmt.Sprintf("%d", c.facts.Generation), c.facts.HeadSHA, c.facts.Owner, fmt.Sprintf("%d", c.facts.Epoch),
	}, "\x00")
	digest := sha256.Sum256([]byte(grantIdentity))
	return Authorization{intent: intent, grant: Grant{
		id: "grant-" + hex.EncodeToString(digest[:16]), effectID: intent.effectID, capability: capability,
		owner: c.facts.Owner, epoch: c.facts.Epoch, issuedAt: now.UTC(),
	}}, nil
}

func (c *Controller) Reauthorize(ctx context.Context, record EffectRecord, now time.Time) (Authorization, error) {
	intent, err := DecodeIntent(record.Kind, record.Payload)
	if err != nil {
		return Authorization{}, err
	}
	if record.EffectID != intent.effectID || record.IdempotencyKey != intent.idempotencyKey ||
		record.PayloadHash != intent.payloadHash || record.DeliveryID != intent.deliveryID ||
		record.Target != intent.target || record.Generation != intent.generation || record.HeadSHA != intent.headSHA ||
		record.SourceID != intent.sourceID || record.Revision != intent.revision {
		return Authorization{}, errors.New("persisted effect identity does not match its immutable payload")
	}
	return c.authorize(ctx, intent, now, record.State == StateUncertain)
}

func (c *Controller) Revalidate(ctx context.Context, now time.Time) error {
	if c == nil || c.authority == nil || now.IsZero() {
		return errors.New("controller authority and validation time are required")
	}
	if err := c.authority.CheckLease(ctx, c.lease, now); err != nil {
		return fmt.Errorf("controller authority lease is stale: %w", err)
	}
	if err := c.revalidateFacts(ctx); err != nil {
		return fmt.Errorf("controller store authority is stale: %w", err)
	}
	return nil
}

func (c *Controller) revalidateFacts(ctx context.Context) error {
	delivery, err := c.authority.GetDelivery(ctx, c.facts.DeliveryID)
	if err != nil {
		return fmt.Errorf("read controller delivery authority: %w", err)
	}
	run, err := c.authority.GetDeliveryRun(ctx, c.facts.DeliveryID)
	if err != nil {
		return fmt.Errorf("read controller run authority: %w", err)
	}
	if delivery.Repository != c.facts.Repository || delivery.Issue != c.facts.Issue ||
		delivery.PullRequest != c.facts.PullRequest || run.Generation != c.facts.Generation {
		return errors.New("controller facts do not match immutable store-issued delivery authority")
	}
	return nil
}

func (c *Controller) matches(intent Intent) error {
	facts := c.facts
	if intent.effectID == "" || intent.deliveryID != facts.DeliveryID || intent.target.Repository != facts.Repository ||
		intent.target.Issue != facts.Issue || intent.target.PullRequest != facts.PullRequest ||
		intent.generation != facts.Generation || intent.headSHA != facts.HeadSHA {
		return errors.New("effect intent does not match protected controller facts")
	}
	return nil
}

func capabilityForKind(kind Kind) (Capability, error) {
	switch kind {
	case KindDecisionSummary, KindDecisionQuestion:
		return CapabilityGitHubCommentWrite, nil
	default:
		return "", fmt.Errorf("effect kind %q has no grantable capability", kind)
	}
}

func (a Authorization) EffectID() string       { return a.intent.effectID }
func (a Authorization) GrantID() string        { return a.grant.id }
func (a Authorization) Capability() Capability { return a.grant.capability }
func (a Authorization) Owner() string          { return a.grant.owner }
func (a Authorization) Epoch() int64           { return a.grant.epoch }
func (a Authorization) Intent() Intent         { return cloneIntent(a.intent) }

func (a Authorization) Record() EffectRecord {
	return recordFromIntent(a.intent)
}

func cloneIntent(intent Intent) Intent {
	intent.payload = append([]byte(nil), intent.payload...)
	return intent
}

func (a Authorization) validate() error {
	if a.grant.id == "" || a.grant.effectID != a.intent.effectID || !IsGrantableCapability(a.grant.capability) ||
		!safeID(a.grant.owner) || a.grant.epoch <= 0 || a.grant.issuedAt.IsZero() {
		return errors.New("authorization is incomplete or forged")
	}
	capability, err := capabilityForKind(a.intent.kind)
	if err != nil || capability != a.grant.capability {
		return errors.New("authorization capability does not match effect kind")
	}
	redecoded, err := DecodeIntent(a.intent.kind, a.intent.payload)
	if err != nil || redecoded.effectID != a.intent.effectID || redecoded.payloadHash != a.intent.payloadHash ||
		redecoded.idempotencyKey != a.intent.idempotencyKey {
		return errors.New("authorization payload identity is invalid")
	}
	return nil
}
