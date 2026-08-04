package connectors

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const writeApprovalGrantVersion = 1

type WriteApprovalTarget struct {
	Connector          string `json:"connector"`
	Operation          string `json:"operation"`
	Method             string `json:"method"`
	MutationClass      string `json:"mutation_class"`
	TargetDigest       string `json:"target_digest"`
	CredentialRevision string `json:"credential_revision"`
}

type WriteApprovalGrant struct {
	Version           int                 `json:"version"`
	PlanID            string              `json:"plan_id"`
	PlanHash          string              `json:"plan_hash"`
	PreviewDigest     string              `json:"preview_digest"`
	ApprovalTokenHash string              `json:"approval_token_hash"`
	Nonce             string              `json:"nonce"`
	Target            WriteApprovalTarget `json:"target"`
	IssuedAt          time.Time           `json:"issued_at"`
	ExpiresAt         time.Time           `json:"expires_at"`
	Confirmation      WriteConfirmation   `json:"confirmation"`
	MAC               string              `json:"mac"`
}

type WriteApprovalGrantRequest struct {
	PlanID        string
	PlanHash      string
	PreviewDigest string
	ApprovalToken string
	Target        WriteApprovalTarget
	IssuedAt      time.Time
	ExpiresAt     time.Time
	Confirmation  WriteConfirmation
}

type WriteApprovalExpectation struct {
	PlanID        string
	PlanHash      string
	PreviewDigest string
	ApprovalToken string
	Target        WriteApprovalTarget
	ExpiresAt     time.Time
	Confirmation  WriteConfirmation
	Now           time.Time
}

type WriteApprovalAuthority struct {
	key [sha256.Size]byte
}

type WriteApprovalEvidence struct {
	target        WriteApprovalTarget
	previewDigest string
	expiresAt     time.Time
	authenticated bool
	use           *writeApprovalUse
}

type writeApprovalUse struct {
	consumed atomic.Bool
}

func NewWriteApprovalAuthority(key []byte) (*WriteApprovalAuthority, error) {
	if len(key) < sha256.Size {
		return nil, fmt.Errorf("write approval key must be at least %d bytes", sha256.Size)
	}
	digest := sha256.Sum256(append([]byte("polymetrics/write-approval-authority/v1\x00"), key...))
	return &WriteApprovalAuthority{key: digest}, nil
}

func (a *WriteApprovalAuthority) CredentialRevision(credentialID string, secrets map[string]string) (string, error) {
	if a == nil {
		return "", errors.New("write approval authority is required")
	}
	if strings.TrimSpace(credentialID) == "" {
		return "", errors.New("credential identity is required")
	}
	keys := make([]string, 0, len(secrets))
	for key := range secrets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([][2]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, [2]string{key, secrets[key]})
	}
	payload, err := json.Marshal(struct {
		Domain       string      `json:"domain"`
		CredentialID string      `json:"credential_id"`
		Secrets      [][2]string `json:"secrets"`
	}{Domain: "credential-revision-v1", CredentialID: credentialID, Secrets: values})
	if err != nil {
		return "", fmt.Errorf("encode credential revision: %w", err)
	}
	return a.authenticate(payload), nil
}

func (a *WriteApprovalAuthority) IssueWriteGrant(req WriteApprovalGrantRequest) (WriteApprovalGrant, error) {
	if a == nil {
		return WriteApprovalGrant{}, errors.New("write approval authority is required")
	}
	if err := validateGrantRequest(req); err != nil {
		return WriteApprovalGrant{}, err
	}
	nonceBytes := make([]byte, 18)
	if _, err := rand.Read(nonceBytes); err != nil {
		return WriteApprovalGrant{}, fmt.Errorf("generate write approval nonce: %w", err)
	}
	grant := WriteApprovalGrant{
		Version:           writeApprovalGrantVersion,
		PlanID:            req.PlanID,
		PlanHash:          req.PlanHash,
		PreviewDigest:     req.PreviewDigest,
		ApprovalTokenHash: hashApprovalToken(req.ApprovalToken),
		Nonce:             hex.EncodeToString(nonceBytes),
		Target:            req.Target,
		IssuedAt:          req.IssuedAt.UTC(),
		ExpiresAt:         req.ExpiresAt.UTC(),
		Confirmation:      req.Confirmation,
	}
	mac, err := a.grantMAC(grant)
	if err != nil {
		return WriteApprovalGrant{}, err
	}
	grant.MAC = mac
	return grant, nil
}

func (a *WriteApprovalAuthority) VerifyWriteGrant(grant WriteApprovalGrant, expected WriteApprovalExpectation) (*WriteApprovalEvidence, error) {
	if a == nil {
		return nil, errors.New("write approval authority is required")
	}
	mac, err := a.grantMAC(grant)
	if err != nil {
		return nil, err
	}
	if !constantStringEqual(mac, grant.MAC) {
		return nil, errors.New("write approval grant authentication failed")
	}
	if grant.Version != writeApprovalGrantVersion || strings.TrimSpace(grant.Nonce) == "" {
		return nil, errors.New("write approval grant is invalid")
	}
	now := expected.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if grant.IssuedAt.IsZero() || grant.ExpiresAt.IsZero() || !grant.ExpiresAt.After(grant.IssuedAt) || now.Before(grant.IssuedAt) || !now.Before(grant.ExpiresAt) {
		return nil, errors.New("write approval grant has expired or is not active")
	}
	if !sameGrantBinding(grant, expected) {
		return nil, errors.New("write approval grant does not match the approved plan")
	}
	if !constantStringEqual(grant.ApprovalTokenHash, hashApprovalToken(expected.ApprovalToken)) {
		return nil, errors.New("approval token is invalid")
	}
	return &WriteApprovalEvidence{
		target:        grant.Target,
		previewDigest: grant.PreviewDigest,
		expiresAt:     grant.ExpiresAt,
		authenticated: true,
		use:           &writeApprovalUse{},
	}, nil
}

func (e *WriteApprovalEvidence) Authorize(target WriteApprovalTarget, previewDigest string, now time.Time) error {
	if e == nil || !e.authenticated || e.use == nil {
		return errors.New("authenticated write approval evidence is required")
	}
	if !sameApprovalTarget(e.target, target) || !constantStringEqual(e.previewDigest, previewDigest) {
		return errors.New("write approval evidence does not match the prepared write")
	}
	if now.UTC().IsZero() {
		now = time.Now().UTC()
	}
	if !now.UTC().Before(e.expiresAt) {
		return errors.New("write approval evidence has expired")
	}
	if !e.use.consumed.CompareAndSwap(false, true) {
		return errors.New("write approval evidence has already been consumed")
	}
	return nil
}

func validateGrantRequest(req WriteApprovalGrantRequest) error {
	if strings.TrimSpace(req.PlanID) == "" || strings.TrimSpace(req.PlanHash) == "" {
		return errors.New("write approval grant requires a plan identity")
	}
	if strings.TrimSpace(req.PreviewDigest) == "" || strings.TrimSpace(req.ApprovalToken) == "" {
		return errors.New("write approval grant requires preview and token identity")
	}
	if err := validateApprovalTarget(req.Target); err != nil {
		return err
	}
	if req.IssuedAt.IsZero() || req.ExpiresAt.IsZero() || !req.ExpiresAt.After(req.IssuedAt) {
		return errors.New("write approval grant requires a bounded validity window")
	}
	if req.Confirmation.Kind != ConfirmationKindDestructive {
		return errors.New("write approval grant requires destructive confirmation")
	}
	return nil
}

func validateApprovalTarget(target WriteApprovalTarget) error {
	if strings.TrimSpace(target.Connector) == "" || strings.TrimSpace(target.Operation) == "" || strings.TrimSpace(target.Method) == "" {
		return errors.New("write approval target is incomplete")
	}
	if strings.TrimSpace(target.TargetDigest) == "" || strings.TrimSpace(target.CredentialRevision) == "" {
		return errors.New("write approval target requires resource and credential revisions")
	}
	return nil
}

func (a *WriteApprovalAuthority) grantMAC(grant WriteApprovalGrant) (string, error) {
	grant.MAC = ""
	payload, err := json.Marshal(grant)
	if err != nil {
		return "", fmt.Errorf("encode write approval grant: %w", err)
	}
	return a.authenticate(append([]byte("write-grant-v1\x00"), payload...)), nil
}

func (a *WriteApprovalAuthority) authenticate(payload []byte) string {
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func sameGrantBinding(grant WriteApprovalGrant, expected WriteApprovalExpectation) bool {
	return constantStringEqual(grant.PlanID, expected.PlanID) &&
		constantStringEqual(grant.PlanHash, expected.PlanHash) &&
		constantStringEqual(grant.PreviewDigest, expected.PreviewDigest) &&
		sameApprovalTarget(grant.Target, expected.Target) &&
		grant.ExpiresAt.Equal(expected.ExpiresAt.UTC()) &&
		grant.Confirmation.Kind == expected.Confirmation.Kind
}

func sameApprovalTarget(left, right WriteApprovalTarget) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	leftDigest := sha256.Sum256(leftJSON)
	rightDigest := sha256.Sum256(rightJSON)
	return subtle.ConstantTimeCompare(leftDigest[:], rightDigest[:]) == 1
}

func hashApprovalToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func constantStringEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
