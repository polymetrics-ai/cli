package outbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/sensitive"
)

const (
	payloadSchemaVersion = 1
	maxPayloadBytes      = 56 * 1024
	maxSummaryBytes      = 48 * 1024
	maxEvidenceBytes     = 1000
	maxOptions           = 8
)

type Kind string

const (
	KindDecisionSummary  Kind = "github.comment.decision_summary.v1"
	KindDecisionQuestion Kind = "github.comment.decision_question.v1"
)

type Capability string

const CapabilityGitHubCommentWrite Capability = "github.comment.write"

func IsGrantableCapability(capability Capability) bool {
	return capability == CapabilityGitHubCommentWrite
}

type State string

const (
	StatePending   State = "pending"
	StateClaimed   State = "claimed"
	StateSent      State = "sent"
	StateFailed    State = "failed"
	StateUncertain State = "uncertain"
	StateBlocked   State = "blocked"
	StateCancelled State = "cancelled"
)

type ResultCode string

const (
	ResultSent       ResultCode = "sent"
	ResultReconciled ResultCode = "reconciled"
)

type ErrorCode string

const (
	ErrorNone              ErrorCode = ""
	ErrorPreSend           ErrorCode = "pre_send_failure"
	ErrorPostSendAmbiguous ErrorCode = "post_send_ambiguous"
	ErrorMarkerConflict    ErrorCode = "marker_conflict"
	ErrorStaleRevision     ErrorCode = "stale_revision"
	ErrorChangedTarget     ErrorCode = "changed_target"
	ErrorUnsupported       ErrorCode = "unsupported_effect"
	ErrorClaimExpired      ErrorCode = "claim_expired_after_execution"
	ErrorRetryExhausted    ErrorCode = "pre_send_retry_exhausted"
	ErrorEffectExpired     ErrorCode = "effect_expired"
)

type EventKind string

const (
	EventRequested        EventKind = "effect_requested"
	EventAuthorized       EventKind = "effect_authorized"
	EventEnqueued         EventKind = "effect_enqueued"
	EventClaimed          EventKind = "effect_claimed"
	EventExecutionStarted EventKind = "effect_execution_started"
	EventSent             EventKind = "effect_sent"
	EventFailed           EventKind = "effect_failed"
	EventUncertain        EventKind = "effect_uncertain"
	EventReconciled       EventKind = "effect_reconciled"
	EventBlocked          EventKind = "effect_blocked"
	EventClaimRecovered   EventKind = "effect_claim_recovered"
	EventCancelled        EventKind = "effect_cancelled"
)

type Target struct {
	Repository  string `json:"repository"`
	Issue       int    `json:"issue"`
	PullRequest int    `json:"pull_request"`
}

type SummaryPayload struct {
	SchemaVersion int    `json:"schema_version"`
	Repository    string `json:"repository"`
	Issue         int    `json:"issue"`
	PullRequest   int    `json:"pull_request"`
	DeliveryID    string `json:"delivery_id"`
	Generation    int64  `json:"generation"`
	HeadSHA       string `json:"head_sha"`
	Revision      int64  `json:"revision"`
	LedgerHash    string `json:"ledger_hash"`
	Summary       string `json:"summary"`
}

type QuestionPayload struct {
	SchemaVersion int      `json:"schema_version"`
	Repository    string   `json:"repository"`
	Issue         int      `json:"issue"`
	PullRequest   int      `json:"pull_request"`
	RequestID     string   `json:"request_id"`
	DeliveryID    string   `json:"delivery_id"`
	UnitID        string   `json:"unit_id"`
	Generation    int64    `json:"generation"`
	HeadSHA       string   `json:"head_sha"`
	Evidence      string   `json:"evidence"`
	Options       []string `json:"options"`
	Recommended   string   `json:"recommended_option,omitempty"`
	SafeDefault   string   `json:"safe_default,omitempty"`
	ExpiresAt     string   `json:"expires_at"`
	Mention       string   `json:"mention"`
}

type Question struct {
	RequestID         string
	DeliveryID        string
	UnitID            string
	Generation        int64
	HeadSHA           string
	Evidence          string
	Options           []string
	RecommendedOption string
	SafeDefault       string
	ExpiresAt         time.Time
	Mention           string
}

type Intent struct {
	kind           Kind
	target         Target
	deliveryID     string
	generation     int64
	headSHA        string
	sourceID       string
	revision       int64
	payload        []byte
	payloadHash    string
	idempotencyKey string
	effectID       string
}

func NewSummaryIntent(target Target, deliveryID string, generation int64, headSHA string, revision int64, ledgerHash, summary string) (Intent, error) {
	payload := SummaryPayload{
		SchemaVersion: payloadSchemaVersion, Repository: target.Repository, Issue: target.Issue,
		PullRequest: target.PullRequest, DeliveryID: deliveryID, Generation: generation,
		HeadSHA: headSHA, Revision: revision, LedgerHash: ledgerHash, Summary: summary,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Intent{}, err
	}
	return decodeSummaryIntent(raw)
}

func NewQuestionIntent(target Target, question Question) (Intent, error) {
	payload := QuestionPayload{
		SchemaVersion: payloadSchemaVersion, Repository: target.Repository, Issue: target.Issue,
		PullRequest: target.PullRequest, RequestID: question.RequestID, DeliveryID: question.DeliveryID,
		UnitID: question.UnitID, Generation: question.Generation, HeadSHA: question.HeadSHA,
		Evidence: question.Evidence, Options: append([]string(nil), question.Options...),
		Recommended: question.RecommendedOption, SafeDefault: question.SafeDefault,
		ExpiresAt: question.ExpiresAt.UTC().Format(time.RFC3339Nano), Mention: strings.TrimPrefix(question.Mention, "@"),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Intent{}, err
	}
	return decodeQuestionIntent(raw)
}

func DecodeIntent(kind Kind, raw []byte) (Intent, error) {
	if len(raw) == 0 || len(raw) > maxPayloadBytes {
		return Intent{}, errors.New("effect payload is empty or oversized")
	}
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return Intent{}, err
	}
	switch kind {
	case KindDecisionSummary:
		return decodeSummaryIntent(raw)
	case KindDecisionQuestion:
		return decodeQuestionIntent(raw)
	default:
		return Intent{}, fmt.Errorf("effect kind %q is unsupported", kind)
	}
}

func decodeSummaryIntent(raw []byte) (Intent, error) {
	var payload SummaryPayload
	if err := decodeStrict(raw, &payload); err != nil {
		return Intent{}, fmt.Errorf("decode decision summary payload: %w", err)
	}
	target := Target{Repository: payload.Repository, Issue: payload.Issue, PullRequest: payload.PullRequest}
	summaryDigest := sha256.Sum256([]byte(payload.Summary))
	expectedLedgerHash := "sha256:" + hex.EncodeToString(summaryDigest[:])
	if payload.SchemaVersion != payloadSchemaVersion || validateTarget(target) != nil || !safeDeliveryID(payload.DeliveryID) ||
		payload.Generation <= 0 || !validGitSHA(payload.HeadSHA) || payload.Revision <= 0 ||
		payload.LedgerHash != expectedLedgerHash || strings.TrimSpace(payload.Summary) == "" ||
		len(payload.Summary) > maxSummaryBytes || sensitive.ValidatePublicDocument(payload.Summary) != nil {
		return Intent{}, errors.New("decision summary payload is incomplete, unsafe, or unsupported")
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return Intent{}, err
	}
	return newIntent(KindDecisionSummary, target, payload.DeliveryID, payload.Generation, payload.HeadSHA,
		payload.DeliveryID, payload.Revision, canonical), nil
}

func decodeQuestionIntent(raw []byte) (Intent, error) {
	var payload QuestionPayload
	if err := decodeStrict(raw, &payload); err != nil {
		return Intent{}, fmt.Errorf("decode decision question payload: %w", err)
	}
	target := Target{Repository: payload.Repository, Issue: payload.Issue, PullRequest: payload.PullRequest}
	expiresAt, timeErr := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
	if payload.SchemaVersion != payloadSchemaVersion || validateTarget(target) != nil || !safeID(payload.RequestID) ||
		!safeDeliveryID(payload.DeliveryID) || !safeUnitID(payload.UnitID) || payload.Generation <= 0 ||
		!validGitSHA(payload.HeadSHA) || strings.TrimSpace(payload.Evidence) == "" || len(payload.Evidence) > maxEvidenceBytes ||
		sensitive.ValidatePublicText(payload.Evidence) != nil || len(payload.Options) == 0 || len(payload.Options) > maxOptions ||
		timeErr != nil || expiresAt.IsZero() || !safeID(payload.Mention) ||
		sensitive.ValidatePublicIdentifier(payload.Mention) != nil {
		return Intent{}, errors.New("decision question payload is incomplete, unsafe, or unsupported")
	}
	allowed := make(map[string]struct{}, len(payload.Options))
	for _, option := range payload.Options {
		if !safeID(option) || sensitive.ValidatePublicIdentifier(option) != nil {
			return Intent{}, errors.New("decision question option is unsafe")
		}
		if _, duplicate := allowed[strings.ToLower(option)]; duplicate {
			return Intent{}, errors.New("decision question options contain a duplicate")
		}
		allowed[strings.ToLower(option)] = struct{}{}
	}
	for _, option := range []string{payload.Recommended, payload.SafeDefault} {
		if option != "" {
			if _, ok := allowed[strings.ToLower(option)]; !ok {
				return Intent{}, errors.New("decision question policy option is not allowlisted")
			}
		}
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return Intent{}, err
	}
	return newIntent(KindDecisionQuestion, target, payload.DeliveryID, payload.Generation, payload.HeadSHA,
		payload.RequestID, 0, canonical), nil
}

func newIntent(kind Kind, target Target, deliveryID string, generation int64, headSHA, sourceID string, revision int64, canonical []byte) Intent {
	payloadDigest := sha256.Sum256(canonical)
	payloadHash := "sha256:" + hex.EncodeToString(payloadDigest[:])
	identity := strings.Join([]string{
		string(kind), deliveryID, target.Repository, strconv.Itoa(target.Issue), strconv.Itoa(target.PullRequest),
		strconv.FormatInt(generation, 10), headSHA, sourceID, strconv.FormatInt(revision, 10),
	}, "\x00")
	identityDigest := sha256.Sum256([]byte(identity))
	idempotencyKey := "sha256:" + hex.EncodeToString(identityDigest[:])
	return Intent{
		kind: kind, target: target, deliveryID: deliveryID, generation: generation, headSHA: headSHA,
		sourceID: sourceID, revision: revision, payload: append([]byte(nil), canonical...),
		payloadHash: payloadHash, idempotencyKey: idempotencyKey, effectID: "effect-" + hex.EncodeToString(identityDigest[:16]),
	}
}

func (i Intent) Kind() Kind             { return i.kind }
func (i Intent) Target() Target         { return i.target }
func (i Intent) DeliveryID() string     { return i.deliveryID }
func (i Intent) Generation() int64      { return i.generation }
func (i Intent) HeadSHA() string        { return i.headSHA }
func (i Intent) SourceID() string       { return i.sourceID }
func (i Intent) Revision() int64        { return i.revision }
func (i Intent) Payload() []byte        { return append([]byte(nil), i.payload...) }
func (i Intent) PayloadHash() string    { return i.payloadHash }
func (i Intent) IdempotencyKey() string { return i.idempotencyKey }
func (i Intent) EffectID() string       { return i.effectID }

func decodeStrict(raw []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are forbidden")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := consumeJSONValue(decoder); err != nil {
		return fmt.Errorf("inspect effect payload: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("effect payload has trailing JSON")
		}
		return err
	}
	return nil
}

func consumeJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			normalized := strings.ToLower(key)
			if _, duplicate := seen[normalized]; duplicate {
				return fmt.Errorf("duplicate JSON field %q", key)
			}
			seen[normalized] = struct{}{}
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func validateTarget(target Target) error {
	parts := strings.Split(target.Repository, "/")
	if len(parts) != 2 || !safeID(parts[0]) || !safeID(parts[1]) || target.Issue <= 0 || target.PullRequest <= 0 {
		return errors.New("complete bounded GitHub target is required")
	}
	return nil
}

func safeDeliveryID(value string) bool { return strings.HasPrefix(value, "issue-") && safeID(value) }

func safeID(value string) bool {
	if value == "" || value == "." || value == ".." || len(value) > 256 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("._-", character) {
			continue
		}
		return false
	}
	return true
}

func safeUnitID(value string) bool {
	return sensitive.ValidatePublicIdentifier(value) == nil
}

func validGitSHA(value string) bool { return len(value) == 40 && validLowerHex(value) }

func validSHA256(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == 71 && validLowerHex(strings.TrimPrefix(value, "sha256:"))
}

func validLowerHex(value string) bool {
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return value != ""
}
