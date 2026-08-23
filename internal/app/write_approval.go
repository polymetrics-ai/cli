package app

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	projectWritePlanSealVersion = 1
	projectWritePlanLifetime    = 24 * time.Hour
)

type projectWriteApprovalAuthority struct {
	dir    string
	key    [sha256.Size]byte
	signer *connectors.WriteApprovalAuthority
}

type projectWriteApprovalEvidence struct {
	target        connectors.WriteApprovalTarget
	previewDigest string
	expiresAt     time.Time
	authorization *AuthorizationScope
	use           *projectWriteApprovalUse
}

type projectWriteApprovalUse struct {
	consumed atomic.Bool
}

func newProjectWriteApprovalAuthority(projectDir string) (*projectWriteApprovalAuthority, error) {
	vaultDir := filepath.Join(projectDir, "vault")
	vaultKey, err := os.ReadFile(filepath.Join(vaultDir, "key"))
	if err != nil {
		return nil, fmt.Errorf("open project write approval authority: %w", err)
	}
	return newProjectWriteApprovalAuthorityFromKey(vaultDir, vaultKey)
}

// newEphemeralProjectWriteApprovalAuthority supplies certification's
// process-local credentials with the revision and configuration fingerprinting
// required by RuntimeConfig. Its root key is random and never written to the
// project; write-approval consumption remains rooted in the ephemeral project
// only if a certification scenario explicitly exercises a write.
func newEphemeralProjectWriteApprovalAuthority(projectDir string) (*projectWriteApprovalAuthority, error) {
	var rootKey [sha256.Size]byte
	if _, err := rand.Read(rootKey[:]); err != nil {
		return nil, fmt.Errorf("generate ephemeral write approval key: %w", err)
	}
	return newProjectWriteApprovalAuthorityFromKey(projectDir, rootKey[:])
}

func newProjectWriteApprovalAuthorityFromKey(dir string, vaultKey []byte) (*projectWriteApprovalAuthority, error) {
	if len(vaultKey) != sha256.Size {
		return nil, fmt.Errorf("project write approval key must be %d bytes", sha256.Size)
	}
	rootMAC := hmac.New(sha256.New, vaultKey)
	_, _ = rootMAC.Write([]byte("polymetrics/write-approval-root/v2"))
	var rootKey [sha256.Size]byte
	copy(rootKey[:], rootMAC.Sum(nil))
	signer, err := connectors.NewUntrustedWriteApprovalAuthority(rootKey[:])
	if err != nil {
		return nil, err
	}
	return &projectWriteApprovalAuthority{dir: dir, key: rootKey, signer: signer}, nil
}

func (a *projectWriteApprovalAuthority) CredentialRevision(credentialID string, secrets map[string]string) (string, error) {
	if a == nil || a.signer == nil {
		return "", errors.New("project write approval authority is required")
	}
	return a.signer.CredentialRevision(credentialID, secrets)
}

func (a *projectWriteApprovalAuthority) ConfigurationDigest(credentialID string, config map[string]string) (string, error) {
	if a == nil || a.signer == nil {
		return "", errors.New("project write approval authority is required")
	}
	return a.signer.ConfigurationDigest(credentialID, config)
}

func (a *projectWriteApprovalAuthority) IssueWritePlanSeal(req connectors.WritePlanSealRequest) (connectors.WritePlanSeal, error) {
	if a == nil {
		return connectors.WritePlanSeal{}, errors.New("project write approval authority is required")
	}
	if err := validateProjectPlanSealRequest(req); err != nil {
		return connectors.WritePlanSeal{}, err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return connectors.WritePlanSeal{}, err
	}
	now := time.Now().UTC()
	seal := connectors.WritePlanSeal{
		Version:             projectWritePlanSealVersion,
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
		ExpiresAt:           now.Add(projectWritePlanLifetime),
	}
	seal.MAC, err = a.planSealMAC(seal)
	if err != nil {
		return connectors.WritePlanSeal{}, err
	}
	return seal, nil
}

func (a *projectWriteApprovalAuthority) VerifyWritePlanSeal(seal connectors.WritePlanSeal, expected connectors.WritePlanSealExpectation) error {
	if a == nil {
		return errors.New("project write approval authority is required")
	}
	mac, err := a.planSealMAC(seal)
	if err != nil {
		return err
	}
	authorityID, err := a.authorityID()
	if err != nil {
		return err
	}
	if !projectApprovalStringEqual(mac, seal.MAC) || !projectApprovalStringEqual(authorityID, seal.AuthorityID) {
		return errors.New("write plan seal authentication failed")
	}
	now := time.Now().UTC()
	if seal.Version != projectWritePlanSealVersion || seal.IssuedAt.IsZero() || seal.ExpiresAt.IsZero() || !seal.ExpiresAt.After(seal.IssuedAt) || now.Before(seal.IssuedAt) || !now.Before(seal.ExpiresAt) {
		return errors.New("write plan seal has expired or is not active")
	}
	if !sameProjectPlanSealBinding(seal, expected) {
		return errors.New("write plan seal does not match the stored plan")
	}
	return nil
}

func (a *projectWriteApprovalAuthority) IssueWriteGrant(req connectors.WriteApprovalGrantRequest) (connectors.WriteApprovalGrant, error) {
	if a == nil || a.signer == nil {
		return connectors.WriteApprovalGrant{}, errors.New("project write approval authority is required")
	}
	if req.PlanSeal == nil {
		return connectors.WriteApprovalGrant{}, errors.New("authenticated write plan seal is required")
	}
	expected := connectors.WritePlanSealExpectation{
		PlanID: req.PlanID, PlanHash: req.PlanHash, Mode: req.Mode,
		Connector: req.Target.Connector, Operation: req.Target.Operation,
		CredentialRevision: req.Target.CredentialRevision, ConfigurationDigest: req.Target.ConfigurationDigest,
		Batchable: req.Target.Batchable, Scope: req.Target.Scope, Confirmation: req.Confirmation,
	}
	if err := a.VerifyWritePlanSeal(*req.PlanSeal, expected); err != nil {
		return connectors.WriteApprovalGrant{}, err
	}
	if req.Target.Scope != connectors.WriteApprovalScopeProject {
		return connectors.WriteApprovalGrant{}, errors.New("project write approval requires project scope")
	}
	consumed, err := a.consumed(projectWriteApprovalConsumptionID(req.PlanID, req.PlanHash, req.Mode))
	if err != nil {
		return connectors.WriteApprovalGrant{}, err
	}
	if consumed {
		return connectors.WriteApprovalGrant{}, connectors.ErrWriteApprovalConsumed
	}
	return a.signer.IssueWriteGrant(req)
}

func (a *projectWriteApprovalAuthority) ValidateWriteGrant(grant connectors.WriteApprovalGrant, expected connectors.WriteApprovalExpectation, seal *connectors.WritePlanSeal) error {
	if a == nil || a.signer == nil {
		return errors.New("project write approval authority is required")
	}
	if seal == nil {
		return errors.New("authenticated write plan seal is required")
	}
	expectedSeal := connectors.WritePlanSealExpectation{
		PlanID: expected.PlanID, PlanHash: expected.PlanHash, Mode: expected.Mode,
		Connector: expected.Target.Connector, Operation: expected.Target.Operation,
		CredentialRevision: expected.Target.CredentialRevision, ConfigurationDigest: expected.Target.ConfigurationDigest,
		Batchable: expected.Target.Batchable, Scope: expected.Target.Scope, Confirmation: expected.Confirmation,
	}
	if err := a.VerifyWritePlanSeal(*seal, expectedSeal); err != nil {
		return err
	}
	if !projectApprovalStringEqual(grant.PlanSealMAC, seal.MAC) {
		return errors.New("write approval grant does not match the authenticated plan seal")
	}
	if err := a.signer.ValidateWriteGrant(grant, expected); err != nil {
		return err
	}
	if grant.Target.Scope != connectors.WriteApprovalScopeProject {
		return errors.New("write approval grant is not project-owned")
	}
	consumed, err := a.consumed(projectWriteApprovalConsumptionID(grant.PlanID, grant.PlanHash, grant.Mode))
	if err != nil {
		return err
	}
	if consumed {
		return connectors.ErrWriteApprovalConsumed
	}
	return nil
}

func (a *projectWriteApprovalAuthority) VerifyWriteGrant(grant connectors.WriteApprovalGrant, expected connectors.WriteApprovalExpectation, seal *connectors.WritePlanSeal) (*connectors.WriteApprovalEvidence, error) {
	if err := a.ValidateWriteGrant(grant, expected, seal); err != nil {
		return nil, err
	}
	if err := a.consume(projectWriteApprovalConsumptionID(grant.PlanID, grant.PlanHash, grant.Mode), grant.Nonce, grant.MAC, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("consume write approval grant: %w", err)
	}
	return connectors.BindProjectWriteApprovalEvidence(&projectWriteApprovalEvidence{
		target:        grant.Target,
		previewDigest: grant.PreviewDigest,
		expiresAt:     grant.ExpiresAt,
		use:           &projectWriteApprovalUse{},
	})
}

func (e *projectWriteApprovalEvidence) ValidateProjectWrite(target connectors.WriteApprovalTarget, previewDigest string, now time.Time) error {
	if e == nil || e.use == nil {
		return errors.New("consumed project write approval evidence is required")
	}
	if e.authorization != nil {
		scope := e.authorization
		if target.Scope != connectors.WriteApprovalScopeProject ||
			!authorizationScopeAllowsWriteAction(*scope, target.Operation) ||
			!projectApprovalStringEqual(target.CredentialRevision, scope.DestinationCredentialRevision) ||
			!projectApprovalStringEqual(target.ConfigurationDigest, scope.DestinationConfigurationDigest) ||
			target.Confirmation.Kind != scope.ConfirmationPolicy.Kind {
			return errors.New("durable authorization evidence does not match the prepared write")
		}
		if now.UTC().IsZero() {
			now = time.Now().UTC()
		}
		if !now.UTC().Before(scope.ExpiresAt.UTC()) {
			return errors.New("durable authorization evidence has expired")
		}
		return nil
	}
	if target.Scope != connectors.WriteApprovalScopeProject || !sameProjectApprovalTarget(e.target, target) || !projectApprovalStringEqual(e.previewDigest, previewDigest) {
		return errors.New("write approval evidence does not match the prepared write")
	}
	if now.UTC().IsZero() {
		now = time.Now().UTC()
	}
	if !now.UTC().Before(e.expiresAt) {
		return errors.New("write approval evidence has expired")
	}
	return nil
}

func authorizationScopeAllowsWriteAction(scope AuthorizationScope, action string) bool {
	if action == scope.WriteAction {
		return true
	}
	for _, allowed := range scope.AllowedWriteActions {
		if action == allowed {
			return true
		}
	}
	return false
}

func (e *projectWriteApprovalEvidence) AuthorizeProjectWrite(target connectors.WriteApprovalTarget, previewDigest string, now time.Time) error {
	if err := e.ValidateProjectWrite(target, previewDigest, now); err != nil {
		return err
	}
	if !e.use.consumed.CompareAndSwap(false, true) {
		if e.authorization != nil {
			return errors.New("durable authorization evidence has already been consumed")
		}
		return errors.New("write approval evidence has already been consumed")
	}
	return nil
}

// durableAuthorizationEvidence adapts a pre-validated standing App scope to
// the closed connector write gate. It carries no token or raw credential and
// intentionally leaves the per-run preview digest unbound: payload is excluded
// from durable authorization by contract.
func durableAuthorizationEvidence(scope AuthorizationScope) (*connectors.WriteApprovalEvidence, error) {
	canonical := canonicalAuthorizationScope(scope)
	return connectors.BindProjectWriteApprovalEvidence(&projectWriteApprovalEvidence{
		authorization: &canonical,
		use:           &projectWriteApprovalUse{},
	})
}

func validateProjectPlanSealRequest(req connectors.WritePlanSealRequest) error {
	if strings.TrimSpace(req.PlanID) == "" || strings.TrimSpace(req.PlanHash) == "" {
		return errors.New("write plan seal requires a plan identity")
	}
	if strings.TrimSpace(req.Connector) == "" || strings.TrimSpace(req.Operation) == "" {
		return errors.New("write plan seal requires connector and operation identity")
	}
	if strings.TrimSpace(req.CredentialRevision) == "" || strings.TrimSpace(req.ConfigurationDigest) == "" {
		return errors.New("write plan seal requires credential and configuration identity")
	}
	if req.Scope != connectors.WriteApprovalScopeProject {
		return errors.New("write plan seal requires project scope")
	}
	if req.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		return errors.New("write plan seal requires destructive confirmation")
	}
	return nil
}

func (a *projectWriteApprovalAuthority) planSealMAC(seal connectors.WritePlanSeal) (string, error) {
	seal.MAC = ""
	payload, err := json.Marshal(seal)
	if err != nil {
		return "", fmt.Errorf("encode write plan seal: %w", err)
	}
	return a.authenticate(append([]byte("write-plan-seal-v1\x00"), payload...))
}

func (a *projectWriteApprovalAuthority) authorityID() (string, error) {
	return a.authenticate([]byte("write-approval-authority-id-v1"))
}

func (a *projectWriteApprovalAuthority) authenticate(payload []byte) (string, error) {
	if a == nil || strings.TrimSpace(a.dir) == "" || a.key == [sha256.Size]byte{} {
		return "", errors.New("project write approval authority is invalid")
	}
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (a *projectWriteApprovalAuthority) consumed(approvalID string) (bool, error) {
	path, err := a.consumptionMarkerPath(approvalID)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("inspect write approval consumption marker: %w", err)
}

func (a *projectWriteApprovalAuthority) consume(approvalID, nonce, grantMAC string, consumedAt time.Time) error {
	if strings.TrimSpace(approvalID) == "" || strings.TrimSpace(nonce) == "" || strings.TrimSpace(grantMAC) == "" {
		return errors.New("write approval consumption identity is incomplete")
	}
	path, err := a.consumptionMarkerPath(approvalID)
	if err != nil {
		return err
	}
	markerID := strings.TrimSuffix(filepath.Base(path), ".used")
	nonceID, err := a.authenticate(append([]byte("consumed-nonce-v1\x00"), []byte(nonce)...))
	if err != nil {
		return err
	}
	dir := filepath.Join(a.dir, "write-approval-consumed")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create write approval consumption directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return connectors.ErrWriteApprovalConsumed
	}
	if err != nil {
		return fmt.Errorf("create write approval consumption marker: %w", err)
	}
	record := struct {
		Version     int       `json:"version"`
		AuthorityID string    `json:"authority_id"`
		MarkerID    string    `json:"marker_id"`
		NonceID     string    `json:"nonce_id"`
		GrantMAC    string    `json:"grant_mac"`
		ConsumedAt  time.Time `json:"consumed_at"`
		MAC         string    `json:"mac"`
	}{Version: 1, MarkerID: markerID, NonceID: nonceID, GrantMAC: grantMAC, ConsumedAt: consumedAt.UTC()}
	record.AuthorityID, err = a.authorityID()
	if err == nil {
		var unsigned []byte
		unsigned, err = json.Marshal(record)
		if err == nil {
			record.MAC, err = a.authenticate(append([]byte("consumed-record-v1\x00"), unsigned...))
		}
	}
	var payload []byte
	if err == nil {
		payload, err = json.Marshal(record)
		payload = append(payload, '\n')
	}
	if err == nil {
		_, err = file.Write(payload)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("write write approval consumption marker: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close write approval consumption marker: %w", closeErr)
	}
	if dirHandle, openErr := os.Open(dir); openErr == nil {
		_ = dirHandle.Sync()
		_ = dirHandle.Close()
	}
	return nil
}

func (a *projectWriteApprovalAuthority) consumptionMarkerPath(approvalID string) (string, error) {
	if a == nil || strings.TrimSpace(a.dir) == "" || a.key == [sha256.Size]byte{} {
		return "", errors.New("project write approval authority is invalid")
	}
	if strings.TrimSpace(approvalID) == "" {
		return "", errors.New("write approval identity is required")
	}
	markerID, err := a.authenticate(append([]byte("consumed-plan-v2\x00"), []byte(approvalID)...))
	if err != nil {
		return "", err
	}
	return filepath.Join(a.dir, "write-approval-consumed", markerID+".used"), nil
}

func sameProjectPlanSealBinding(seal connectors.WritePlanSeal, expected connectors.WritePlanSealExpectation) bool {
	return projectApprovalStringEqual(seal.PlanID, expected.PlanID) &&
		projectApprovalStringEqual(seal.PlanHash, expected.PlanHash) &&
		projectApprovalStringEqual(seal.Mode, expected.Mode) &&
		projectApprovalStringEqual(seal.Connector, expected.Connector) &&
		projectApprovalStringEqual(seal.Operation, expected.Operation) &&
		projectApprovalStringEqual(seal.CredentialRevision, expected.CredentialRevision) &&
		projectApprovalStringEqual(seal.ConfigurationDigest, expected.ConfigurationDigest) &&
		seal.Batchable == expected.Batchable &&
		projectApprovalStringEqual(seal.Scope, expected.Scope) &&
		seal.Confirmation.Kind == expected.Confirmation.Kind
}

func sameProjectApprovalTarget(left, right connectors.WriteApprovalTarget) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	leftDigest := sha256.Sum256(leftJSON)
	rightDigest := sha256.Sum256(rightJSON)
	return subtle.ConstantTimeCompare(leftDigest[:], rightDigest[:]) == 1
}

func projectWriteApprovalConsumptionID(planID, planHash, mode string) string {
	return planID + "\x00" + planHash + "\x00" + mode
}

func projectApprovalStringEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
