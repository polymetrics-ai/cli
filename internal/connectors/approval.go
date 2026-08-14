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
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	writeApprovalGrantVersion = 2
	writeApprovalLifetime     = 15 * time.Minute

	WriteApprovalScopeProject = "project"
	WriteApprovalScopeFixture = "fixture_loopback"
)

var ErrWriteApprovalConsumed = errors.New("write approval has already been consumed")

type WriteApprovalTarget struct {
	Connector           string            `json:"connector"`
	Operation           string            `json:"operation"`
	Method              string            `json:"method"`
	MutationClass       string            `json:"mutation_class"`
	TargetDigest        string            `json:"target_digest"`
	CredentialRevision  string            `json:"credential_revision"`
	ConfigurationDigest string            `json:"configuration_digest"`
	Batchable           bool              `json:"batchable"`
	Scope               string            `json:"scope"`
	Confirmation        WriteConfirmation `json:"confirmation"`
}

type WritePlanSeal struct {
	Version             int               `json:"version"`
	AuthorityID         string            `json:"authority_id"`
	PlanID              string            `json:"plan_id"`
	PlanHash            string            `json:"plan_hash"`
	Mode                string            `json:"mode,omitempty"`
	Connector           string            `json:"connector"`
	Operation           string            `json:"operation"`
	CredentialRevision  string            `json:"credential_revision"`
	ConfigurationDigest string            `json:"configuration_digest"`
	Batchable           bool              `json:"batchable"`
	Scope               string            `json:"scope"`
	Confirmation        WriteConfirmation `json:"confirmation"`
	IssuedAt            time.Time         `json:"issued_at"`
	ExpiresAt           time.Time         `json:"expires_at"`
	MAC                 string            `json:"mac"`
}

type WritePlanSealRequest struct {
	PlanID              string
	PlanHash            string
	Mode                string
	Connector           string
	Operation           string
	CredentialRevision  string
	ConfigurationDigest string
	Batchable           bool
	Scope               string
	Confirmation        WriteConfirmation
}

type WritePlanSealExpectation = WritePlanSealRequest

type WriteApprovalGrant struct {
	Version           int                 `json:"version"`
	AuthorityID       string              `json:"authority_id"`
	PlanID            string              `json:"plan_id"`
	PlanHash          string              `json:"plan_hash"`
	Mode              string              `json:"mode,omitempty"`
	PlanSealMAC       string              `json:"plan_seal_mac,omitempty"`
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
	Mode          string
	PlanSeal      *WritePlanSeal
	PreviewDigest string
	ApprovalToken string
	Target        WriteApprovalTarget
	Confirmation  WriteConfirmation
}

type WriteApprovalExpectation struct {
	PlanID        string
	PlanHash      string
	Mode          string
	PreviewDigest string
	ApprovalToken string
	Target        WriteApprovalTarget
	Confirmation  WriteConfirmation
}

type writeApprovalAuthorityKind uint8

const (
	writeApprovalAuthorityUntrusted writeApprovalAuthorityKind = iota
	writeApprovalAuthorityFixture
)

type WriteApprovalAuthority struct {
	key      [sha256.Size]byte
	kind     writeApprovalAuthorityKind
	consumed *fixtureWriteApprovalConsumptions
}

type fixtureWriteApprovalConsumptions struct {
	grants sync.Map
}

type projectWriteApprovalEvidence interface {
	AuthorizeProjectWrite(WriteApprovalTarget, string, time.Time) error
}

type WriteApprovalEvidence struct {
	target        WriteApprovalTarget
	previewDigest string
	expiresAt     time.Time
	authorityKind writeApprovalAuthorityKind
	use           *writeApprovalUse
	project       projectWriteApprovalEvidence
}

type writeApprovalUse struct {
	consumed atomic.Bool
}

func NewUntrustedWriteApprovalAuthority(key []byte) (*WriteApprovalAuthority, error) {
	if len(key) < sha256.Size {
		return nil, fmt.Errorf("write approval key must be at least %d bytes", sha256.Size)
	}
	var authorityKey [sha256.Size]byte
	if len(key) == sha256.Size {
		copy(authorityKey[:], key)
	} else {
		authorityKey = sha256.Sum256(key)
	}
	return &WriteApprovalAuthority{key: authorityKey, kind: writeApprovalAuthorityUntrusted}, nil
}

func NewFixtureWriteApprovalAuthority() (*WriteApprovalAuthority, error) {
	var key [sha256.Size]byte
	if _, err := rand.Read(key[:]); err != nil {
		return nil, fmt.Errorf("generate fixture write approval authority: %w", err)
	}
	return &WriteApprovalAuthority{key: key, kind: writeApprovalAuthorityFixture, consumed: &fixtureWriteApprovalConsumptions{}}, nil
}

func BindProjectWriteApprovalEvidence(evidence projectWriteApprovalEvidence) (*WriteApprovalEvidence, error) {
	if evidence == nil {
		return nil, errors.New("project write approval evidence is required")
	}
	typeOf := reflect.TypeOf(evidence)
	valueOf := reflect.ValueOf(evidence)
	if typeOf.Kind() != reflect.Pointer || valueOf.IsNil() || typeOf.Elem().PkgPath() != "polymetrics.ai/internal/app" || typeOf.Elem().Name() != "projectWriteApprovalEvidence" {
		return nil, errors.New("project write approval evidence must originate from App approval consumption")
	}
	return &WriteApprovalEvidence{project: evidence}, nil
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
	}{Domain: "credential-revision-v2", CredentialID: credentialID, Secrets: values})
	if err != nil {
		return "", fmt.Errorf("encode credential revision: %w", err)
	}
	return a.authenticate(payload)
}

func (a *WriteApprovalAuthority) ConfigurationDigest(credentialID string, config map[string]string) (string, error) {
	if a == nil {
		return "", errors.New("write approval authority is required")
	}
	if strings.TrimSpace(credentialID) == "" {
		return "", errors.New("credential identity is required")
	}
	payload, err := json.Marshal(struct {
		Domain       string            `json:"domain"`
		CredentialID string            `json:"credential_id"`
		Config       map[string]string `json:"config"`
	}{Domain: "configuration-digest-v1", CredentialID: credentialID, Config: config})
	if err != nil {
		return "", fmt.Errorf("encode configuration digest: %w", err)
	}
	return a.authenticate(payload)
}

func (a *WriteApprovalAuthority) IssueWriteGrant(req WriteApprovalGrantRequest) (WriteApprovalGrant, error) {
	if a == nil {
		return WriteApprovalGrant{}, errors.New("write approval authority is required")
	}
	if err := validateGrantRequest(req); err != nil {
		return WriteApprovalGrant{}, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(writeApprovalLifetime)
	planSealMAC := ""
	switch a.kind {
	case writeApprovalAuthorityFixture:
		if req.Target.Scope != WriteApprovalScopeFixture {
			return WriteApprovalGrant{}, errors.New("fixture write approval requires fixture scope")
		}
	case writeApprovalAuthorityUntrusted:
		if req.PlanSeal != nil {
			planSealMAC = req.PlanSeal.MAC
			if req.PlanSeal.ExpiresAt.Before(expiresAt) {
				expiresAt = req.PlanSeal.ExpiresAt
			}
		}
	default:
		return WriteApprovalGrant{}, errors.New("write approval authority is invalid")
	}
	nonceBytes := make([]byte, 18)
	if _, err := rand.Read(nonceBytes); err != nil {
		return WriteApprovalGrant{}, fmt.Errorf("generate write approval nonce: %w", err)
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return WriteApprovalGrant{}, err
	}
	grant := WriteApprovalGrant{
		Version:           writeApprovalGrantVersion,
		AuthorityID:       authorityID,
		PlanID:            req.PlanID,
		PlanHash:          req.PlanHash,
		Mode:              req.Mode,
		PlanSealMAC:       planSealMAC,
		PreviewDigest:     req.PreviewDigest,
		ApprovalTokenHash: hashApprovalToken(req.ApprovalToken),
		Nonce:             hex.EncodeToString(nonceBytes),
		Target:            req.Target,
		IssuedAt:          now,
		ExpiresAt:         expiresAt,
		Confirmation:      req.Confirmation,
	}
	grant.MAC, err = a.grantMAC(grant)
	if err != nil {
		return WriteApprovalGrant{}, err
	}
	return grant, nil
}

func (a *WriteApprovalAuthority) ValidateWriteGrant(grant WriteApprovalGrant, expected WriteApprovalExpectation) error {
	if a == nil {
		return errors.New("write approval authority is required")
	}
	mac, err := a.grantMAC(grant)
	if err != nil {
		return err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return err
	}
	if !constantStringEqual(mac, grant.MAC) || !constantStringEqual(authorityID, grant.AuthorityID) {
		return errors.New("write approval grant authentication failed")
	}
	if grant.Version != writeApprovalGrantVersion || strings.TrimSpace(grant.Nonce) == "" {
		return errors.New("write approval grant is invalid")
	}
	now := time.Now().UTC()
	if grant.IssuedAt.IsZero() || grant.ExpiresAt.IsZero() || !grant.ExpiresAt.After(grant.IssuedAt) || now.Before(grant.IssuedAt) || !now.Before(grant.ExpiresAt) {
		return errors.New("write approval grant has expired or is not active")
	}
	if !sameGrantBinding(grant, expected) {
		return errors.New("write approval grant does not match the approved plan")
	}
	if !constantStringEqual(grant.ApprovalTokenHash, hashApprovalToken(expected.ApprovalToken)) {
		return errors.New("approval token is invalid")
	}
	switch a.kind {
	case writeApprovalAuthorityFixture:
		if grant.Target.Scope != WriteApprovalScopeFixture {
			return errors.New("fixture write approval scope is invalid")
		}
		if a.consumed == nil {
			return errors.New("fixture write approval consumption registry is unavailable")
		}
		if _, consumed := a.consumed.grants.Load(writeApprovalConsumptionID(grant)); consumed {
			return ErrWriteApprovalConsumed
		}
	case writeApprovalAuthorityUntrusted:
	default:
		return errors.New("write approval authority is invalid")
	}
	return nil
}

func (a *WriteApprovalAuthority) VerifyWriteGrant(grant WriteApprovalGrant, expected WriteApprovalExpectation) (*WriteApprovalEvidence, error) {
	if err := a.ValidateWriteGrant(grant, expected); err != nil {
		return nil, err
	}
	if a.kind == writeApprovalAuthorityFixture {
		if _, consumed := a.consumed.grants.LoadOrStore(writeApprovalConsumptionID(grant), struct{}{}); consumed {
			return nil, ErrWriteApprovalConsumed
		}
	}
	return &WriteApprovalEvidence{
		target:        grant.Target,
		previewDigest: grant.PreviewDigest,
		expiresAt:     grant.ExpiresAt,
		authorityKind: a.kind,
		use:           &writeApprovalUse{},
	}, nil
}

func writeApprovalConsumptionID(grant WriteApprovalGrant) string {
	return grant.AuthorityID + "\x00" + grant.Nonce + "\x00" + grant.MAC
}

func (e *WriteApprovalEvidence) Authorize(target WriteApprovalTarget, previewDigest string, now time.Time) error {
	if e == nil {
		return errors.New("authenticated write approval evidence is required")
	}
	if e.project != nil {
		return e.project.AuthorizeProjectWrite(target, previewDigest, now)
	}
	if e.use == nil {
		return errors.New("authenticated write approval evidence is required")
	}
	if e.authorityKind != writeApprovalAuthorityFixture || target.Scope != WriteApprovalScopeFixture {
		return errors.New("caller-selected write approval authority is not trusted")
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
	if req.Confirmation.Kind != ConfirmationKindDestructive || req.Target.Confirmation.Kind != req.Confirmation.Kind {
		return errors.New("write approval grant requires destructive confirmation")
	}
	return nil
}

func validateApprovalTarget(target WriteApprovalTarget) error {
	if strings.TrimSpace(target.Connector) == "" || strings.TrimSpace(target.Operation) == "" || strings.TrimSpace(target.Method) == "" {
		return errors.New("write approval target is incomplete")
	}
	if strings.TrimSpace(target.TargetDigest) == "" || strings.TrimSpace(target.CredentialRevision) == "" || strings.TrimSpace(target.ConfigurationDigest) == "" {
		return errors.New("write approval target requires resource, credential, and configuration revisions")
	}
	if target.Scope != WriteApprovalScopeProject && target.Scope != WriteApprovalScopeFixture {
		return errors.New("write approval target scope is invalid")
	}
	if target.Confirmation.Kind != ConfirmationKindDestructive {
		return errors.New("write approval target requires destructive confirmation")
	}
	return nil
}

func (a *WriteApprovalAuthority) grantMAC(grant WriteApprovalGrant) (string, error) {
	grant.MAC = ""
	payload, err := json.Marshal(grant)
	if err != nil {
		return "", fmt.Errorf("encode write approval grant: %w", err)
	}
	return a.authenticate(append([]byte("write-grant-v2\x00"), payload...))
}

func (a *WriteApprovalAuthority) authorityID() (string, error) {
	return a.authenticate([]byte("write-approval-authority-id-v1"))
}

func (a *WriteApprovalAuthority) authenticate(payload []byte) (string, error) {
	if a == nil || a.key == [sha256.Size]byte{} {
		return "", errors.New("write approval authority is invalid")
	}
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func sameGrantBinding(grant WriteApprovalGrant, expected WriteApprovalExpectation) bool {
	return constantStringEqual(grant.PlanID, expected.PlanID) &&
		constantStringEqual(grant.PlanHash, expected.PlanHash) &&
		constantStringEqual(grant.Mode, expected.Mode) &&
		constantStringEqual(grant.PreviewDigest, expected.PreviewDigest) &&
		sameApprovalTarget(grant.Target, expected.Target) &&
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
