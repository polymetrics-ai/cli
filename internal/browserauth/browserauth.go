// Package browserauth is the one place Polymetrics authenticates a human
// through their own real browser. It never fetches data — only a
// credential. Two mechanisms hand back two different credential shapes: an
// official OAuth token (loopback authorization-code+PKCE, or the device
// flow for headless hosts) for the official connectors, or a captured
// browser session (driver) for the -web connectors. Acquisition happens
// once, interactively, outside a sync — it is deliberately not an
// internal/connectors/engine auth mode (see engine/auth.go's doc comment on
// bearer/oauth2_refresh for the request-time half of this split).
//
// The password rule is enforced mechanically here, not by policy: this
// package's driver subpackage exposes no typing, keyboard, or form-fill
// API, and a guard test (driver/guard_test.go) fails the build if one ever
// appears. The tool cannot type a password because the code to do so does
// not exist.
package browserauth

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	reauthenticationSafetyMargin = 30 * time.Second
	providerOAuthErrorCode       = "provider_error"
)

// SafeOAuthErrorCode returns a recognized OAuth error code or a generic code.
func SafeOAuthErrorCode(raw string) string {
	switch raw {
	case "invalid_request",
		"invalid_client",
		"invalid_grant",
		"invalid_scope",
		"unauthorized_client",
		"unsupported_grant_type",
		"unsupported_response_type",
		"authorization_pending",
		"slow_down",
		"access_denied",
		"expired_token",
		"server_error",
		"temporarily_unavailable",
		"interaction_required",
		"login_required",
		"consent_required",
		"account_selection_required":
		return raw
	default:
		return providerOAuthErrorCode
	}
}

// Credential is the result of a successful Login: exactly one of OAuth or
// Session is set, matching which Flow produced it.
type Credential struct {
	OAuth   *OAuthCredential
	Session *SessionCredential
}

// Validate enforces the public credential contract: each login has one
// mechanism outcome, never both and never neither.
func (c Credential) Validate() error {
	if (c.OAuth == nil) == (c.Session == nil) {
		return errors.New("browserauth: credential must contain exactly one of OAuth or Session")
	}
	return nil
}

// NeedsReauthentication reports whether this credential should be replaced by
// another interactive Login. It does not try to refresh or silently replay a
// browser session: a caller uses true as the signal to send the human through
// the same approved browser-auth flow again.
func (c Credential) NeedsReauthentication(now time.Time) bool {
	if c.OAuth != nil {
		return c.OAuth.Expired(now)
	}
	if c.Session != nil {
		return c.Session.Expired(now)
	}
	// An invalid/no-outcome credential cannot authenticate a request safely.
	return true
}

// OAuthCredential is what the loopback (authorization-code + PKCE) and
// device flows yield: an official provider token. It is consumed by the
// engine's existing "bearer" auth mode
// (Authorization: Bearer {{ secrets.access_token }}) directly, or by the
// engine's oauth2_refresh mode (report §3.7) once a scheduled sync needs to
// refresh a short-lived token without re-running this interactive flow.
type OAuthCredential struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scopes       []string

	// TokenURL/ClientID/ClientSecret are recorded alongside the token so a
	// later refresh (engine oauth2_refresh, or this package re-run) does not
	// need the original request re-supplied out of band.
	TokenURL     string
	ClientID     string
	ClientSecret string
}

// Expired reports whether the access token is at or past its expiry, with a
// 30s safety margin so a caller never starts a request with a token that
// expires mid-flight. A zero ExpiresAt (provider did not report one) is
// treated as never-expiring — callers that need a hard guarantee should
// prefer a provider that reports expiry.
func (c OAuthCredential) Expired(now time.Time) bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return !now.Before(c.ExpiresAt.Add(-reauthenticationSafetyMargin))
}

// SessionCredential is what the driver (browser session capture) flow
// yields for the -web connectors: the named minimum of cookies the
// mechanism needs — never "everything on the domain".
type SessionCredential struct {
	Cookies []Cookie
	// CSRFHeader/CSRFValue are set when the mechanism needs a CSRF token
	// derived from (or captured alongside) the session cookies, stamped as
	// a header on every write (e.g. LinkedIn's Voyager csrf-token header
	// derived from JSESSIONID).
	CSRFHeader string
	CSRFValue  string
	// Origin is the provider origin this session belongs to
	// (e.g. "https://www.linkedin.com"). A guard test in the native
	// transport layer that consumes this (added alongside the -web
	// connectors, not here) asserts every request stays on Origin's host.
	Origin string
	// FingerprintRef names the resolved browser build that captured this
	// session (driver.Resolution.Version) — recorded because captured
	// cookies are only coherent with the browser build that produced them.
	// It is the browser's reported version string where the binary reports
	// one, the pinned revision for a downloaded Chromium, and otherwise a
	// path/size/mtime fingerprint of the binary. The guarantee it supports
	// is therefore "did the browser build change since capture?", not a
	// semantic version comparison.
	FingerprintRef string
	CapturedAt     time.Time
	ExpiresHint    *time.Time
}

// Expired reports whether the browser session should be reacquired. A missing
// expiry hint means the provider did not expose a reliable session deadline;
// callers may still reauthenticate after an upstream 401/403, but this helper
// does not force an unnecessary browser login preemptively.
func (c SessionCredential) Expired(now time.Time) bool {
	if c.ExpiresHint == nil {
		return false
	}
	return !now.Before(c.ExpiresHint.Add(-reauthenticationSafetyMargin))
}

// Cookie is one captured session cookie.
type Cookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	Expires  *time.Time
}

// Flow is implemented by each acquisition mechanism (loopback, device,
// driver). A Flow is used exactly once per Login call — it is not a
// long-lived session object.
type Flow interface {
	// Name identifies the flow for logs/errors ("loopback_pkce",
	// "device_code", "browser_session_capture").
	Name() string
	Login(ctx context.Context) (Credential, error)
}

// Login runs flow and returns its credential. It is the one documented
// entry point — callers depend on this stable import regardless of which
// concrete flow package backs a given connector's mechanism, matching the
// "one command, two credential outcomes" design (report §3.2): the bundle
// declares which flow it needs; the user never picks one.
func Login(ctx context.Context, flow Flow) (Credential, error) {
	if flow == nil {
		return Credential{}, errors.New("browserauth: login flow is required")
	}
	credential, err := flow.Login(ctx)
	if err != nil {
		return Credential{}, err
	}
	if err := credential.Validate(); err != nil {
		return Credential{}, fmt.Errorf("browserauth: %s returned invalid credential: %w", flow.Name(), err)
	}
	return credential, nil
}
