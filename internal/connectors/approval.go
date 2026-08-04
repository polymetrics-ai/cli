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

	"polymetrics.ai/internal/vault"
)

const (
	writeApprovalGrantVersion = 2
	writePlanSealVersion      = 1
	writePlanLifetime         = 24 * time.Hour
	writeApprovalLifetime     = 15 * time.Minute

	WriteApprovalScopeProject = "project"
	WriteApprovalScopeFixture = "fixture_loopback"
)

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
	writeApprovalAuthorityProcess
	writeApprovalAuthorityFixture
)

type WriteApprovalAuthority struct {
	key  [sha256.Size]byte
	kind writeApprovalAuthorityKind
	root vault.WriteApprovalRoot
}

type WriteApprovalEvidence struct {
	target               WriteApprovalTarget
	previewDigest        string
	expiresAt            time.Time
	authorityKind        writeApprovalAuthorityKind
	persistentlyConsumed bool
	use                  *writeApprovalUse
}

type writeApprovalUse struct {
	consumed atomic.Bool
}

func NewUntrustedWriteApprovalAuthority(key []byte) (*WriteApprovalAuthority, error) {
	if len(key) < sha256.Size {
		return nil, fmt.Errorf("write approval key must be at least %d bytes", sha256.Size)
	}
	digest := sha256.Sum256(append([]byte("polymetrics/write-approval-untrusted/v2\x00"), key...))
	return &WriteApprovalAuthority{key: digest, kind: writeApprovalAuthorityUntrusted}, nil
}

func NewProcessWriteApprovalAuthority(root vault.WriteApprovalRoot) (*WriteApprovalAuthority, error) {
	if !root.Valid() {
		return nil, errors.New("trusted write approval root is required")
	}
	return &WriteApprovalAuthority{kind: writeApprovalAuthorityProcess, root: root}, nil
}

func NewFixtureWriteApprovalAuthority() (*WriteApprovalAuthority, error) {
	var key [sha256.Size]byte
	if _, err := rand.Read(key[:]); err != nil {
		return nil, fmt.Errorf("generate fixture write approval authority: %w", err)
	}
	return &WriteApprovalAuthority{key: key, kind: writeApprovalAuthorityFixture}, nil
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

func (a *WriteApprovalAuthority) IssueWritePlanSeal(req WritePlanSealRequest) (WritePlanSeal, error) {
	if a == nil || a.kind != writeApprovalAuthorityProcess {
		return WritePlanSeal{}, errors.New("process-owned write approval authority is required")
	}
	if err := validatePlanSealRequest(req); err != nil {
		return WritePlanSeal{}, err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return WritePlanSeal{}, err
	}
	now := time.Now().UTC()
	seal := WritePlanSeal{
		Version:             writePlanSealVersion,
		AuthorityID:         authorityID,
		PlanID:              req.PlanID,
		PlanHash:            req.PlanHash,
		Mode:                req.Mode,
		Connector:           req.Connector,
		Operation:           req.Operation,
		CredentialRevision:  req.CredentialRevision,
		ConfigurationDigest: req.ConfigurationDigest,
		Batchable:           req.Batchable,
		Scope:               req.Scope,
		Confirmation:        req.Confirmation,
		IssuedAt:            now,
		ExpiresAt:           now.Add(writePlanLifetime),
	}
	seal.MAC, err = a.planSealMAC(seal)
	if err != nil {
		return WritePlanSeal{}, err
	}
	return seal, nil
}

func (a *WriteApprovalAuthority) VerifyWritePlanSeal(seal WritePlanSeal, expected WritePlanSealExpectation) error {
	if a == nil || a.kind != writeApprovalAuthorityProcess {
		return errors.New("process-owned write approval authority is required")
	}
	mac, err := a.planSealMAC(seal)
	if err != nil {
		return err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return err
	}
	if !constantStringEqual(mac, seal.MAC) || !constantStringEqual(authorityID, seal.AuthorityID) {
		return errors.New("write plan seal authentication failed")
	}
	now := time.Now().UTC()
	if seal.Version != writePlanSealVersion || seal.IssuedAt.IsZero() || seal.ExpiresAt.IsZero() || !seal.ExpiresAt.After(seal.IssuedAt) || now.Before(seal.IssuedAt) || !now.Before(seal.ExpiresAt) {
		return errors.New("write plan seal has expired or is not active")
	}
	if !samePlanSealBinding(seal, expected) {
		return errors.New("write plan seal does not match the stored plan")
	}
	return nil
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
	case writeApprovalAuthorityProcess:
		if req.PlanSeal == nil {
			return WriteApprovalGrant{}, errors.New("authenticated write plan seal is required")
		}
		expected := WritePlanSealExpectation{
			PlanID: req.PlanID, PlanHash: req.PlanHash, Mode: req.Mode,
			Connector: req.Target.Connector, Operation: req.Target.Operation,
			CredentialRevision: req.Target.CredentialRevision, ConfigurationDigest: req.Target.ConfigurationDigest,
			Batchable: req.Target.Batchable, Scope: req.Target.Scope, Confirmation: req.Confirmation,
		}
		if err := a.VerifyWritePlanSeal(*req.PlanSeal, expected); err != nil {
			return WriteApprovalGrant{}, err
		}
		if req.PlanSeal.ExpiresAt.Before(expiresAt) {
			expiresAt = req.PlanSeal.ExpiresAt
		}
		planSealMAC = req.PlanSeal.MAC
		if req.Target.Scope != WriteApprovalScopeProject {
			return WriteApprovalGrant{}, errors.New("process write approval requires project scope")
		}
		consumed, err := a.root.Consumed(writeApprovalConsumptionID(req.PlanID, req.PlanHash, req.Mode))
		if err != nil {
			return WriteApprovalGrant{}, err
		}
		if consumed {
			return WriteApprovalGrant{}, vault.ErrWriteApprovalConsumed
		}
	case writeApprovalAuthorityFixture:
		if req.Target.Scope != WriteApprovalScopeFixture {
			return WriteApprovalGrant{}, errors.New("fixture write approval requires fixture scope")
		}
	case writeApprovalAuthorityUntrusted:
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

func (a *WriteApprovalAuthority) VerifyWriteGrant(grant WriteApprovalGrant, expected WriteApprovalExpectation) (*WriteApprovalEvidence, error) {
	if a == nil {
		return nil, errors.New("write approval authority is required")
	}
	mac, err := a.grantMAC(grant)
	if err != nil {
		return nil, err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return nil, err
	}
	if !constantStringEqual(mac, grant.MAC) || !constantStringEqual(authorityID, grant.AuthorityID) {
		return nil, errors.New("write approval grant authentication failed")
	}
	if grant.Version != writeApprovalGrantVersion || strings.TrimSpace(grant.Nonce) == "" {
		return nil, errors.New("write approval grant is invalid")
	}
	now := time.Now().UTC()
	if grant.IssuedAt.IsZero() || grant.ExpiresAt.IsZero() || !grant.ExpiresAt.After(grant.IssuedAt) || now.Before(grant.IssuedAt) || !now.Before(grant.ExpiresAt) {
		return nil, errors.New("write approval grant has expired or is not active")
	}
	if !sameGrantBinding(grant, expected) {
		return nil, errors.New("write approval grant does not match the approved plan")
	}
	if !constantStringEqual(grant.ApprovalTokenHash, hashApprovalToken(expected.ApprovalToken)) {
		return nil, errors.New("approval token is invalid")
	}
	persistentlyConsumed := false
	switch a.kind {
	case writeApprovalAuthorityProcess:
		if grant.Target.Scope != WriteApprovalScopeProject || strings.TrimSpace(grant.PlanSealMAC) == "" {
			return nil, errors.New("write approval grant is not process-owned")
		}
		if err := a.root.Consume(writeApprovalConsumptionID(grant.PlanID, grant.PlanHash, grant.Mode), grant.Nonce, grant.MAC, now); err != nil {
			return nil, fmt.Errorf("consume write approval grant: %w", err)
		}
		persistentlyConsumed = true
	case writeApprovalAuthorityFixture:
		if grant.Target.Scope != WriteApprovalScopeFixture {
			return nil, errors.New("fixture write approval scope is invalid")
		}
	case writeApprovalAuthorityUntrusted:
	default:
		return nil, errors.New("write approval authority is invalid")
	}
	return &WriteApprovalEvidence{
		target:               grant.Target,
		previewDigest:        grant.PreviewDigest,
		expiresAt:            grant.ExpiresAt,
		authorityKind:        a.kind,
		persistentlyConsumed: persistentlyConsumed,
		use:                  &writeApprovalUse{},
	}, nil
}

func (e *WriteApprovalEvidence) Authorize(target WriteApprovalTarget, previewDigest string, now time.Time) error {
	if e == nil || e.use == nil {
		return errors.New("authenticated write approval evidence is required")
	}
	switch e.authorityKind {
	case writeApprovalAuthorityProcess:
		if !e.persistentlyConsumed || target.Scope != WriteApprovalScopeProject {
			return errors.New("process-owned consumed write approval evidence is required")
		}
	case writeApprovalAuthorityFixture:
		if target.Scope != WriteApprovalScopeFixture {
			return errors.New("fixture write approval cannot authorize a project target")
		}
	default:
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

func validatePlanSealRequest(req WritePlanSealRequest) error {
	if strings.TrimSpace(req.PlanID) == "" || strings.TrimSpace(req.PlanHash) == "" {
		return errors.New("write plan seal requires a plan identity")
	}
	if strings.TrimSpace(req.Connector) == "" || strings.TrimSpace(req.Operation) == "" {
		return errors.New("write plan seal requires connector and operation identity")
	}
	if strings.TrimSpace(req.CredentialRevision) == "" || strings.TrimSpace(req.ConfigurationDigest) == "" {
		return errors.New("write plan seal requires credential and configuration identity")
	}
	if req.Scope != WriteApprovalScopeProject {
		return errors.New("write plan seal requires project scope")
	}
	if req.Confirmation.Kind != ConfirmationKindDestructive {
		return errors.New("write plan seal requires destructive confirmation")
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

func (a *WriteApprovalAuthority) planSealMAC(seal WritePlanSeal) (string, error) {
	seal.MAC = ""
	payload, err := json.Marshal(seal)
	if err != nil {
		return "", fmt.Errorf("encode write plan seal: %w", err)
	}
	return a.authenticate(append([]byte("write-plan-seal-v1\x00"), payload...))
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
	if a == nil {
		return "", errors.New("write approval authority is required")
	}
	if a.kind == writeApprovalAuthorityProcess {
		return a.root.Authenticate(payload)
	}
	if a.key == [sha256.Size]byte{} {
		return "", errors.New("write approval authority is invalid")
	}
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func samePlanSealBinding(seal WritePlanSeal, expected WritePlanSealExpectation) bool {
	return constantStringEqual(seal.PlanID, expected.PlanID) &&
		constantStringEqual(seal.PlanHash, expected.PlanHash) &&
		constantStringEqual(seal.Mode, expected.Mode) &&
		constantStringEqual(seal.Connector, expected.Connector) &&
		constantStringEqual(seal.Operation, expected.Operation) &&
		constantStringEqual(seal.CredentialRevision, expected.CredentialRevision) &&
		constantStringEqual(seal.ConfigurationDigest, expected.ConfigurationDigest) &&
		seal.Batchable == expected.Batchable &&
		constantStringEqual(seal.Scope, expected.Scope) &&
		seal.Confirmation.Kind == expected.Confirmation.Kind
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

func writeApprovalConsumptionID(planID, planHash, mode string) string {
	return planID + "\x00" + planHash + "\x00" + mode
}

func constantStringEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
