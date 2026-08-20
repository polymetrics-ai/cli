package connsdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/credential"
	"polymetrics.ai/internal/safety"
)

const (
	// defaultRefreshExpirySafetyMargin is how far before a provider's stated
	// deadline the access token is renewed, so it cannot expire in flight.
	// Matches OAuth2ClientCredentials' own 60s window.
	defaultRefreshExpirySafetyMargin = 60 * time.Second

	// defaultRefreshFallbackTTL is the lifetime assumed when a token response
	// carries no usable expires_in. It is deliberately SHORT.
	//
	// OAuth2ClientCredentials assumes 3600s in the same situation. For an
	// app-only token that is a reasonable guess; for a user-context token it is
	// not — one hour is precisely the interval at which providers such as
	// Reddit expire theirs, so guessing an hour turns "the provider did not
	// tell us" into "the connector fails an hour in", which is the exact
	// failure this mode exists to remove. Treating an absent expires_in as a
	// short life costs one extra exchange every few minutes in the worst case
	// and cannot produce a silently-dead sync.
	defaultRefreshFallbackTTL = 5 * time.Minute

	// maxTokenResponseBytes bounds the token response body. A token response is
	// a handful of JSON fields; anything larger is a broken or hostile endpoint
	// and must not be buffered unboundedly.
	maxTokenResponseBytes = 1 << 20

	// maxTokenLifetimeSeconds caps a provider-declared expires_in so an absurd
	// value cannot overflow time.Duration. One day is far beyond any real
	// access-token lifetime.
	maxTokenLifetimeSeconds = 24 * 60 * 60
)

// AuthRefresher is implemented by Authenticators whose applied credential can
// be invalidated by the provider before its stated expiry — a revoked grant, a
// password change, a scope change. Expiry-based renewal cannot see any of
// those: the cached credential still looks valid, and every request 401s until
// the process dies.
//
// Requester calls RefreshAuth on a 401 AT MOST ONCE PER REQUEST (see
// Requester.doWithBody), so an endpoint that keeps returning 401 terminates
// with that 401 rather than being hammered.
//
// req is the request that was just rejected, with the credential still applied
// to it. Implementations may read it to tell which credential failed — that is
// what lets concurrent refreshes for the same stale credential collapse into
// one exchange — but must not mutate or resend it.
//
// The interface is optional. An Authenticator that does not implement it (every
// mode that predates the refresh-token grant) sees no behavioural change
// whatsoever on a 401.
type AuthRefresher interface {
	RefreshAuth(ctx context.Context, req *http.Request) error
}

// OAuth2RefreshToken authenticates with a user-context access token obtained
// through the OAuth2 refresh-token grant (RFC 6749 §6) and renewed
// automatically, so a connector whose provider issues short-lived user tokens
// can run unattended.
//
// It is not a variant of OAuth2ClientCredentials. That grant obtains an
// APP-ONLY token: it authenticates the application, not a user, and therefore
// cannot reach any endpoint that acts on behalf of an end user. A user-context
// token comes from the authorization-code flow and is renewed with a refresh
// token, which is what this type does.
//
// Behaviour:
//
//   - The access token is fetched once and reused until shortly before expiry
//     (see ExpirySafetyMargin), not exchanged per request.
//   - A missing/zero/negative/unparseable expires_in is a conservative short
//     life (see FallbackTTL), never "never expires".
//   - A rotated refresh token in the token response replaces the in-memory one
//     and is handed to OnRefreshTokenRotated for persistence. Many providers
//     rotate on every exchange and invalidate the old token; losing that value
//     silently breaks the connector on its NEXT run.
//   - Every exchange happens under one mutex, so concurrent streams sharing one
//     authenticator produce ONE exchange whose result they all share.
//
// The only request this type can emit is a fixed POST of a form-encoded
// refresh-token grant to TokenURL. Method, path and body structure are literals
// here; nothing a caller supplies can change them.
//
// RefreshToken, ClientSecret and the obtained access token are all secrets.
// They are never logged and never reach an error string — see the error
// construction in exchangeLocked, and TestOAuth2RefreshTokenErrorsNeverLeakCredentials.
type OAuth2RefreshToken struct {
	// TokenURL is the provider's token endpoint. Required.
	TokenURL string
	// ClientID and ClientSecret identify the OAuth2 client. Some providers
	// (public clients) omit ClientSecret.
	ClientID     string
	ClientSecret string
	// ClientSecretRequired distinguishes an absent public-client secret from a
	// declared but empty secret, which must be rejected before token emission.
	ClientSecretRequired bool
	// RefreshToken is the initial refresh token. Required. On a provider that
	// rotates, the value actually presented is the most recent one, not this.
	RefreshToken string
	// Scopes, when set, are sent as a space-joined scope parameter.
	Scopes []string
	// ExtraParams are added to the token request form (e.g. audience).
	ExtraParams url.Values

	// Client is used for the token request. Defaults to a 30s client.
	Client *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	// ExpirySafetyMargin overrides how far before expiry renewal happens.
	// Defaults to defaultRefreshExpirySafetyMargin. The effective margin is
	// clamped to half the token's lifetime, so a provider whose lifetime is
	// shorter than the margin still gets caching instead of an exchange per
	// request.
	ExpirySafetyMargin time.Duration
	// FallbackTTL overrides the lifetime assumed when the token response
	// carries no usable expires_in. Defaults to defaultRefreshFallbackTTL.
	FallbackTTL time.Duration

	// OnRefreshTokenRotated, when set, is invoked with a provider-rotated
	// refresh token so the caller can persist it. It is called while the
	// authenticator's lock is held, so the persisted value can never be older
	// than the in-memory one. An error from it fails the exchange: a rotated
	// token that was not persisted breaks the next run, so the run that caused
	// it must fail loudly rather than drift silently.
	OnRefreshTokenRotated func(ctx context.Context, refreshToken string) error

	mu sync.Mutex
	// refreshToken is the current grant. Empty until the first use, when it is
	// seeded from RefreshToken.
	refreshToken string
	seeded       bool
	accessToken  string
	// renewAt is when the cached access token stops being reused. It is
	// strictly before expiresAt by the clamped safety margin.
	renewAt   time.Time
	expiresAt time.Time
}

func (a *OAuth2RefreshToken) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *OAuth2RefreshToken) safetyMargin() time.Duration {
	if a.ExpirySafetyMargin > 0 {
		return a.ExpirySafetyMargin
	}
	return defaultRefreshExpirySafetyMargin
}

func (a *OAuth2RefreshToken) fallbackTTL() time.Duration {
	if a.FallbackTTL > 0 {
		return a.FallbackTTL
	}
	return defaultRefreshFallbackTTL
}

func (a *OAuth2RefreshToken) httpClient() *http.Client {
	if a.Client != nil {
		return a.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a usable access token and sets the Authorization header.
func (a *OAuth2RefreshToken) Apply(ctx context.Context, req *http.Request) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	token, err := a.accessTokenLocked(ctx)
	if err != nil {
		return err
	}
	if err := credential.ValidateHTTPHeaderValue("OAuth2 access token", token); err != nil {
		return fmt.Errorf("oauth2 refresh: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// RefreshAuth implements AuthRefresher: it discards the credential the rejected
// request carried and obtains a replacement.
//
// If the cached access token has already moved past the one req carried,
// another in-flight request refreshed it first — this caller returns
// immediately and its retry picks up the fresher token, so N concurrent 401s
// still cost one exchange.
func (a *OAuth2RefreshToken) RefreshAuth(ctx context.Context, req *http.Request) error {
	stale := bearerCredential(req)

	a.mu.Lock()
	defer a.mu.Unlock()
	if stale != "" && a.accessToken != "" && a.accessToken != stale {
		return nil
	}
	a.invalidateLocked()
	_, err := a.accessTokenLocked(ctx)
	return err
}

// invalidateLocked drops the cached access token, forcing the next
// accessTokenLocked to exchange. The refresh token is deliberately kept: it is
// the credential that survives an access-token revocation.
func (a *OAuth2RefreshToken) invalidateLocked() {
	a.accessToken = ""
	a.renewAt = time.Time{}
	a.expiresAt = time.Time{}
}

// bearerCredential returns the bearer credential applied to req, or "" when it
// carries no bearer Authorization header.
func bearerCredential(req *http.Request) string {
	if req == nil {
		return ""
	}
	const prefix = "Bearer "
	value := req.Header.Get("Authorization")
	if len(value) < len(prefix) || !strings.EqualFold(value[:len(prefix)], prefix) {
		return ""
	}
	return value[len(prefix):]
}

// accessTokenLocked returns a usable access token, exchanging only when the
// cached one is absent or due for renewal. The caller must hold a.mu — holding
// it across the exchange is what makes concurrent callers share one refresh
// rather than each starting their own.
func (a *OAuth2RefreshToken) accessTokenLocked(ctx context.Context) (string, error) {
	if a.accessToken != "" && a.now().Before(a.renewAt) {
		return a.accessToken, nil
	}
	return a.exchangeLocked(ctx)
}

// exchangeLocked performs the refresh-token grant. The caller must hold a.mu.
//
// No error it returns carries the refresh token, the client secret, the access
// token, the request URL, or any provider-supplied body text.
func (a *OAuth2RefreshToken) exchangeLocked(ctx context.Context) (string, error) {
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("oauth2 refresh: token_url is required")
	}
	if !a.seeded {
		a.refreshToken = a.RefreshToken
		a.seeded = true
	}
	if err := credential.RequireAuthenticationValue("OAuth2 refresh token", a.refreshToken); err != nil {
		return "", fmt.Errorf("oauth2 refresh: %w", err)
	}
	if a.ClientID != "" {
		if err := credential.RequireAuthenticationValue("OAuth2 client ID", a.ClientID); err != nil {
			return "", fmt.Errorf("oauth2 refresh: %w", err)
		}
	}
	if a.ClientSecret != "" || a.ClientSecretRequired {
		if err := credential.RequireAuthenticationValue("OAuth2 client secret", a.ClientSecret); err != nil {
			return "", fmt.Errorf("oauth2 refresh: %w", err)
		}
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	if a.ClientID != "" {
		form.Set("client_id", a.ClientID)
	}
	if a.ClientSecret != "" {
		form.Set("client_secret", a.ClientSecret)
	}
	if len(a.Scopes) > 0 {
		form.Set("scope", strings.Join(a.Scopes, " "))
	}
	for k, vs := range a.ExtraParams {
		for _, v := range vs {
			form.Add(k, v)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", redact("oauth2 refresh: build token request", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(req)
	if err != nil {
		// url.Error embeds the request URL, which may itself carry a query.
		return "", redact("oauth2 refresh: token request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Status code only. A token endpoint's error body routinely echoes the
		// grant back (RFC 6749 §5.2 permits an arbitrary error_description),
		// so it is never included.
		return "", fmt.Errorf("oauth2 refresh: token endpoint returned %d %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		// Deliberately RawMessage rather than json.Number: providers return
		// expires_in as a bare number, as a quoted number, and occasionally as
		// something unparseable. Decoding it into a typed field would turn any
		// of those into a whole-response decode failure, when the correct
		// answer is to fall back to a conservative short life.
		ExpiresIn    json.RawMessage `json:"expires_in"`
		RefreshToken string          `json:"refresh_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxTokenResponseBytes)).Decode(&out); err != nil {
		// The decoder's message can quote provider bytes; report the shape of
		// the failure, not its content.
		return "", errors.New("oauth2 refresh: token endpoint returned a body that is not a valid token response")
	}
	if err := credential.RequireAuthenticationValue("OAuth2 access token", out.AccessToken); err != nil {
		return "", fmt.Errorf("oauth2 refresh: %w", err)
	}

	if rotated := out.RefreshToken; rotated != "" && rotated != a.refreshToken {
		if err := credential.RequireAuthenticationValue("OAuth2 refresh token", rotated); err != nil {
			return "", fmt.Errorf("oauth2 refresh: %w", err)
		}
		a.refreshToken = rotated
		if a.OnRefreshTokenRotated != nil {
			if err := a.OnRefreshTokenRotated(ctx, rotated); err != nil {
				return "", fmt.Errorf("oauth2 refresh: persist rotated refresh token: %w", err)
			}
		}
	}

	ttl := a.fallbackTTL()
	if parsed, ok := parseExpiresIn(out.ExpiresIn); ok {
		ttl = parsed
	}
	margin := a.safetyMargin()
	if half := ttl / 2; margin > half {
		margin = half
	}

	now := a.now()
	a.accessToken = out.AccessToken
	a.expiresAt = now.Add(ttl)
	a.renewAt = now.Add(ttl - margin)
	return a.accessToken, nil
}

// parseExpiresIn reads a token response's expires_in as a duration. It reports
// false — meaning "use the conservative fallback" — for an absent, null,
// non-numeric, zero or negative value. It never reports "never expires".
//
// Both a bare number (3600) and a quoted number ("3600") are accepted; the
// latter is common enough in the wild that rejecting it would push working
// providers onto the fallback path for no reason.
func parseExpiresIn(raw json.RawMessage) (time.Duration, bool) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return 0, false
	}
	var quoted string
	if err := json.Unmarshal(raw, &quoted); err == nil {
		trimmed = strings.TrimSpace(quoted)
	}
	secs, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || math.IsNaN(secs) || secs <= 0 {
		return 0, false
	}
	if secs > maxTokenLifetimeSeconds {
		// Clamp rather than overflow time.Duration. Costs at most one extra
		// exchange a day on an unusually long-lived token.
		secs = maxTokenLifetimeSeconds
	}
	return time.Duration(secs * float64(time.Second)), true
}

// redact renders err with URLs and secret assignments stripped while keeping
// the chain intact for errors.Is/errors.As (so a cancelled context still
// matches context.Canceled). Mirrors HTTPError.Error, which redacts at render
// time rather than at capture.
func redact(prefix string, err error) error {
	return &redactedError{
		msg: safety.RedactErrorText(fmt.Sprintf("%s: %v", prefix, err)),
		err: err,
	}
}

type redactedError struct {
	msg string
	err error
}

func (e *redactedError) Error() string { return e.msg }
func (e *redactedError) Unwrap() error { return e.err }
