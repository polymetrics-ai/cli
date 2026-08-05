package store_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth"
	"polymetrics.ai/internal/browserauth/store"
	"polymetrics.ai/internal/vault"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	v, err := vault.InitWithProtector(t.TempDir(), fakeProtector{})
	if err != nil {
		t.Fatalf("vault.InitWithProtector() error = %v", err)
	}
	return store.New(v)
}

// fakeProtector is an in-memory vault.KeyProtector — store tests never touch
// a real OS keychain.
type fakeProtector struct{}

func (fakeProtector) Name() string { return "fake" }
func (fakeProtector) LoadOrCreateKey(string) ([]byte, error) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key, nil
}

func TestStoreOAuthRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)
	cred := browserauth.OAuthCredential{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt,
		Scopes:       []string{"read", "write"},
		TokenURL:     "https://provider.example/token",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}

	if err := s.SaveOAuth(ctx, "reddit", "default", cred); err != nil {
		t.Fatalf("SaveOAuth() error = %v", err)
	}
	got, err := s.LoadOAuth(ctx, "reddit", "default")
	if err != nil {
		t.Fatalf("LoadOAuth() error = %v", err)
	}
	if got.AccessToken != cred.AccessToken || got.RefreshToken != cred.RefreshToken {
		t.Fatalf("LoadOAuth() = %+v, want tokens matching %+v", got, cred)
	}
	if !got.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", got.ExpiresAt, expiresAt)
	}
	if len(got.Scopes) != 2 || got.Scopes[0] != "read" || got.Scopes[1] != "write" {
		t.Fatalf("Scopes = %v", got.Scopes)
	}

	if err := s.DeleteOAuth(ctx, "reddit", "default"); err != nil {
		t.Fatalf("DeleteOAuth() error = %v", err)
	}
	if _, err := s.LoadOAuth(ctx, "reddit", "default"); err == nil {
		t.Fatalf("LoadOAuth() after delete: want error, got nil")
	}
}

func TestStoreSessionRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	captured := time.Now().Truncate(time.Second)
	cred := browserauth.SessionCredential{
		Cookies: []browserauth.Cookie{
			{Name: "li_at", Value: "cookie-value", Domain: ".linkedin.com", Path: "/", Secure: true, HTTPOnly: true},
		},
		CSRFHeader:     "csrf-token",
		CSRFValue:      "csrf-value",
		Origin:         "https://www.linkedin.com",
		FingerprintRef: "chromium-1321438",
		CapturedAt:     captured,
	}

	if err := s.SaveSession(ctx, "linkedin-web", "default", cred); err != nil {
		t.Fatalf("SaveSession() error = %v", err)
	}
	got, err := s.LoadSession(ctx, "linkedin-web", "default")
	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
	if len(got.Cookies) != 1 || got.Cookies[0].Value != "cookie-value" {
		t.Fatalf("LoadSession() cookies = %+v", got.Cookies)
	}
	if got.CSRFValue != "csrf-value" || got.Origin != "https://www.linkedin.com" {
		t.Fatalf("LoadSession() = %+v", got)
	}

	if err := s.DeleteSession(ctx, "linkedin-web", "default"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := s.LoadSession(ctx, "linkedin-web", "default"); err == nil {
		t.Fatalf("LoadSession() after delete: want error, got nil")
	}
}

func TestStoreRiskAcceptanceGate(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	warningText := "Using linkedin-web may get your personal LinkedIn account restricted or shut down."
	hash := store.HashWarning(warningText)

	accepted, err := s.HasAcceptedCurrentRisk(ctx, "linkedin-web", "default", hash)
	if err != nil {
		t.Fatalf("HasAcceptedCurrentRisk() error = %v", err)
	}
	if accepted {
		t.Fatalf("HasAcceptedCurrentRisk() = true before any acceptance was saved")
	}

	err = s.SaveRiskAcceptance(ctx, store.RiskAcceptance{
		Connector:     "linkedin-web",
		Profile:       "default",
		MechanismKind: "web_session",
		WarningSHA256: hash,
		CLIVersion:    "test-version",
		AcceptedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("SaveRiskAcceptance() error = %v", err)
	}

	accepted, err = s.HasAcceptedCurrentRisk(ctx, "linkedin-web", "default", hash)
	if err != nil {
		t.Fatalf("HasAcceptedCurrentRisk() error = %v", err)
	}
	if !accepted {
		t.Fatalf("HasAcceptedCurrentRisk() = false after saving a matching acceptance")
	}

	// A material risk-text update changes the hash — re-required, not
	// silently still-accepted.
	newHash := store.HashWarning(warningText + " Updated wording.")
	accepted, err = s.HasAcceptedCurrentRisk(ctx, "linkedin-web", "default", newHash)
	if err != nil {
		t.Fatalf("HasAcceptedCurrentRisk() error = %v", err)
	}
	if accepted {
		t.Fatalf("HasAcceptedCurrentRisk() = true against a changed warning hash, want false")
	}
}

func TestStoreRiskAcceptanceRejectsEmptyHash(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	err := s.SaveRiskAcceptance(ctx, store.RiskAcceptance{
		Connector:  "linkedin-web",
		Profile:    "default",
		AcceptedAt: time.Now(),
	})
	if err == nil {
		t.Fatalf("SaveRiskAcceptance() with empty warning hash: want error, got nil")
	}
}

func TestLogoutDeletesEverythingAndBestEffortRevokes(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)

	if err := s.SaveOAuth(ctx, "reddit", "default", browserauth.OAuthCredential{AccessToken: "tok"}); err != nil {
		t.Fatalf("SaveOAuth() error = %v", err)
	}
	if err := s.SaveRiskAcceptance(ctx, store.RiskAcceptance{
		Connector: "reddit", Profile: "default", WarningSHA256: "x", AcceptedAt: time.Now(),
	}); err != nil {
		t.Fatalf("SaveRiskAcceptance() error = %v", err)
	}

	var revokedWith *browserauth.Credential
	revokeErr := errors.New("provider revoke endpoint down")
	err := s.Logout(ctx, "reddit", "default", func(_ context.Context, cred browserauth.Credential) error {
		revokedWith = &cred
		return revokeErr
	})
	if err == nil || !errors.Is(err, revokeErr) {
		t.Fatalf("Logout() error = %v, want it to wrap the revoke error", err)
	}
	if revokedWith == nil || revokedWith.OAuth == nil || revokedWith.OAuth.AccessToken != "tok" {
		t.Fatalf("Logout() did not call revoke with the stored credential: %+v", revokedWith)
	}

	// Local state must be gone even though revoke failed.
	if _, err := s.LoadOAuth(ctx, "reddit", "default"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadOAuth() after Logout() error = %v, want os.ErrNotExist", err)
	}
	if _, err := s.LoadRiskAcceptance(ctx, "reddit", "default"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadRiskAcceptance() after Logout() error = %v, want os.ErrNotExist", err)
	}
}

func TestLogoutWithNoRevokeFunc(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	if err := s.SaveSession(ctx, "twitter-web", "default", browserauth.SessionCredential{Origin: "https://x.com"}); err != nil {
		t.Fatalf("SaveSession() error = %v", err)
	}
	if err := s.Logout(ctx, "twitter-web", "default", nil); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := s.LoadSession(ctx, "twitter-web", "default"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadSession() after Logout() error = %v, want os.ErrNotExist", err)
	}
}

func TestLogoutReportsUnreadableCredentialAfterLocalDeletion(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	v, err := vault.InitWithProtector(root, fakeProtector{})
	if err != nil {
		t.Fatalf("InitWithProtector() error = %v", err)
	}
	s := store.New(v)
	if err := s.SaveOAuth(ctx, "reddit", "default", browserauth.OAuthCredential{AccessToken: "token"}); err != nil {
		t.Fatalf("SaveOAuth() error = %v", err)
	}
	path := filepath.Join(root, "vault", "ns", "reddit", "default", "oauth.enc")
	if err := os.WriteFile(path, []byte("corrupt"), 0o600); err != nil {
		t.Fatalf("corrupt OAuth credential: %v", err)
	}

	called := false
	err = s.Logout(ctx, "reddit", "default", func(context.Context, browserauth.Credential) error {
		called = true
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "provider revoke could not be attempted") {
		t.Fatalf("Logout() error = %v, want failed credential-load report", err)
	}
	if called {
		t.Fatal("Logout() called revoke with an unreadable credential")
	}
	if _, err := s.LoadOAuth(ctx, "reddit", "default"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadOAuth() after Logout() error = %v, want os.ErrNotExist", err)
	}
}
