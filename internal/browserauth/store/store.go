// Package store persists browserauth credentials in the vault, isolated per
// connector/profile via vault.Namespace (F15 "session credential +
// transport + pins": the store half). It also owns the risk-acceptance
// record that gates a -web connector's first login and first write (report
// §5.1), and a generic Logout that deletes every stored entry for one
// (connector, profile) after an optional best-effort provider-side revoke.
package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"polymetrics.ai/internal/browserauth"
	"polymetrics.ai/internal/vault"
)

const (
	kindOAuth          = "oauth"
	kindSession        = "session"
	kindRiskAcceptance = "risk_acceptance"
)

// Store adapts vault.Vault to browserauth's two credential kinds.
type Store struct {
	vault *vault.Vault
}

func New(v *vault.Vault) *Store {
	return &Store{vault: v}
}

func namespaceFor(connector, profile, kind string) vault.Namespace {
	return vault.Namespace{Connector: connector, Profile: profile, Kind: kind}
}

// SaveOAuth persists an official OAuth token for (connector, profile).
func (s *Store) SaveOAuth(ctx context.Context, connector, profile string, cred browserauth.OAuthCredential) error {
	m := map[string]string{
		"access_token":  cred.AccessToken,
		"refresh_token": cred.RefreshToken,
		"token_type":    cred.TokenType,
		"token_url":     cred.TokenURL,
		"client_id":     cred.ClientID,
		"client_secret": cred.ClientSecret,
	}
	if len(cred.Scopes) > 0 {
		m["scopes"] = strings.Join(cred.Scopes, " ")
	}
	if !cred.ExpiresAt.IsZero() {
		m["expires_at"] = cred.ExpiresAt.Format(time.RFC3339)
	}
	return s.vault.PutNamespaced(ctx, namespaceFor(connector, profile, kindOAuth), m)
}

// LoadOAuth returns the OAuth token stored for (connector, profile).
func (s *Store) LoadOAuth(ctx context.Context, connector, profile string) (browserauth.OAuthCredential, error) {
	m, err := s.vault.GetNamespaced(ctx, namespaceFor(connector, profile, kindOAuth))
	if err != nil {
		return browserauth.OAuthCredential{}, err
	}
	cred := browserauth.OAuthCredential{
		AccessToken:  m["access_token"],
		RefreshToken: m["refresh_token"],
		TokenType:    m["token_type"],
		TokenURL:     m["token_url"],
		ClientID:     m["client_id"],
		ClientSecret: m["client_secret"],
	}
	if scopes := m["scopes"]; scopes != "" {
		cred.Scopes = strings.Fields(scopes)
	}
	if raw := m["expires_at"]; raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return browserauth.OAuthCredential{}, fmt.Errorf("decode expires_at: %w", err)
		}
		cred.ExpiresAt = t
	}
	return cred, nil
}

// DeleteOAuth removes the OAuth token stored for (connector, profile).
func (s *Store) DeleteOAuth(ctx context.Context, connector, profile string) error {
	return s.vault.DeleteNamespaced(ctx, namespaceFor(connector, profile, kindOAuth))
}

// SaveSession persists a captured browser session for (connector, profile).
// It is stored as an opaque blob (vault.PutBlob), not the flat credential
// map, because a cookie list does not fit map[string]string.
func (s *Store) SaveSession(ctx context.Context, connector, profile string, cred browserauth.SessionCredential) error {
	raw, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("marshal session credential: %w", err)
	}
	return s.vault.PutBlob(ctx, namespaceFor(connector, profile, kindSession), raw)
}

// LoadSession returns the captured browser session for (connector, profile).
func (s *Store) LoadSession(ctx context.Context, connector, profile string) (browserauth.SessionCredential, error) {
	raw, err := s.vault.GetBlob(ctx, namespaceFor(connector, profile, kindSession))
	if err != nil {
		return browserauth.SessionCredential{}, err
	}
	var cred browserauth.SessionCredential
	if err := json.Unmarshal(raw, &cred); err != nil {
		return browserauth.SessionCredential{}, fmt.Errorf("decode session credential: %w", err)
	}
	return cred, nil
}

// DeleteSession removes the captured browser session for (connector, profile).
func (s *Store) DeleteSession(ctx context.Context, connector, profile string) error {
	return s.vault.DeleteBlob(ctx, namespaceFor(connector, profile, kindSession))
}

// Load returns whichever credential kind is stored for (connector, profile)
// — OAuth if present, else the captured session. Used by Logout to attempt
// provider-side revocation without the caller needing to know which kind a
// given connector uses.
func (s *Store) Load(ctx context.Context, connector, profile string) (browserauth.Credential, error) {
	if oauth, err := s.LoadOAuth(ctx, connector, profile); err == nil {
		return browserauth.Credential{OAuth: &oauth}, nil
	}
	if session, err := s.LoadSession(ctx, connector, profile); err == nil {
		return browserauth.Credential{Session: &session}, nil
	}
	return browserauth.Credential{}, fmt.Errorf("no credential stored for %s/%s", connector, profile)
}

// RiskAcceptance records that a user read and accepted a -web connector's
// risk warning. Its presence — with WarningSHA256 matching the CURRENT
// warning text's hash — is the enable-gate: a web_session mechanism must
// not authenticate or perform its first write without one (report §5.1).
type RiskAcceptance struct {
	Connector     string
	Profile       string
	MechanismKind string
	WarningSHA256 string
	CLIVersion    string
	AcceptedAt    time.Time
}

// HashWarning returns the SHA-256 hex digest of the exact warning text — the
// value both the interactive confirmation path and the non-interactive
// `--accept-risk <sha256>` flag record, so a non-interactive caller can only
// pass a hash of text they actually read (never a bare boolean flag that
// could be copy-pasted into CI by someone who never read anything).
func HashWarning(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// SaveRiskAcceptance records ra, overwriting any prior acceptance for the
// same (connector, profile) — accepting a NEW warning text (a new
// WarningSHA256, after a material risk update) supersedes the old
// acceptance rather than stacking.
func (s *Store) SaveRiskAcceptance(ctx context.Context, ra RiskAcceptance) error {
	if strings.TrimSpace(ra.WarningSHA256) == "" {
		return errors.New("risk acceptance requires a non-empty warning_sha256")
	}
	if ra.AcceptedAt.IsZero() {
		return errors.New("risk acceptance requires a non-zero accepted_at")
	}
	m := map[string]string{
		"connector":      ra.Connector,
		"profile":        ra.Profile,
		"mechanism_kind": ra.MechanismKind,
		"warning_sha256": ra.WarningSHA256,
		"cli_version":    ra.CLIVersion,
		"accepted_at":    ra.AcceptedAt.Format(time.RFC3339),
	}
	return s.vault.PutNamespaced(ctx, namespaceFor(ra.Connector, ra.Profile, kindRiskAcceptance), m)
}

// LoadRiskAcceptance returns the risk acceptance recorded for
// (connector, profile). Its error wraps os.ErrNotExist (via the vault's own
// %w-wrapped os.ReadFile error) when none has been recorded, so callers can
// use errors.Is(err, os.ErrNotExist) — HasAcceptedCurrentRisk does this for
// callers that just want a bool.
func (s *Store) LoadRiskAcceptance(ctx context.Context, connector, profile string) (RiskAcceptance, error) {
	m, err := s.vault.GetNamespaced(ctx, namespaceFor(connector, profile, kindRiskAcceptance))
	if err != nil {
		return RiskAcceptance{}, err
	}
	ra := RiskAcceptance{
		Connector:     m["connector"],
		Profile:       m["profile"],
		MechanismKind: m["mechanism_kind"],
		WarningSHA256: m["warning_sha256"],
		CLIVersion:    m["cli_version"],
	}
	if raw := m["accepted_at"]; raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return RiskAcceptance{}, fmt.Errorf("decode accepted_at: %w", err)
		}
		ra.AcceptedAt = t
	}
	return ra, nil
}

// HasAcceptedCurrentRisk reports whether (connector, profile) has a recorded
// acceptance whose WarningSHA256 matches currentWarningSHA256 — false both
// when nothing was ever accepted and when the warning text has since
// changed (re-required whenever the warning text hash changes, per §5.1).
func (s *Store) HasAcceptedCurrentRisk(ctx context.Context, connector, profile, currentWarningSHA256 string) (bool, error) {
	ra, err := s.LoadRiskAcceptance(ctx, connector, profile)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return ra.WarningSHA256 == currentWarningSHA256, nil
}

// RevokeFunc is an optional provider-side revocation call, invoked by
// Logout before local state is deleted. A connector with no revoke
// endpoint (most -web sessions have none) passes nil.
type RevokeFunc func(ctx context.Context, cred browserauth.Credential) error

// Logout deletes every stored entry for (connector, profile) — OAuth
// credential, session credential, risk acceptance, and any native blobs
// saved under the same namespace — after a best-effort call to revoke, if
// given. A revoke failure does not stop local deletion: a provider
// revocation endpoint being down must never leave local state that
// `pm auth status` would still report as live (the local half of "working
// revocation/logout" is unconditional; the remote half is best-effort and
// its failure is reported, not swallowed).
func (s *Store) Logout(ctx context.Context, connector, profile string, revoke RevokeFunc) error {
	var revokeErr error
	if revoke != nil {
		if cred, err := s.Load(ctx, connector, profile); err == nil {
			revokeErr = revoke(ctx, cred)
		}
	}
	if _, err := s.vault.DeleteAll(ctx, connector, profile); err != nil {
		return fmt.Errorf("logout %s/%s: %w", connector, profile, err)
	}
	if revokeErr != nil {
		return fmt.Errorf("logout %s/%s: local state deleted, but provider revoke failed: %w", connector, profile, revokeErr)
	}
	return nil
}
